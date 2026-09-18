/* Concurrent posix_spawn from 4 threads, the way the Go port does it: under a
 * mutex, scan for close-on-exec descriptors, add a close action for each, dup2
 * a pipe onto the child's stdout, spawn `sh -c "echo x"`. Closes of the parent's
 * ends happen under the same mutex. Reports every spawn error (errno) and whether
 * the parent's read side saw EOF. Rounds: 1 thread, then 4 threads. */
#include <errno.h>
#include <fcntl.h>
#include <pthread.h>
#include <spawn.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/wait.h>
#include <unistd.h>

extern char **environ;
static pthread_mutex_t big = PTHREAD_MUTEX_INITIALIZER;
static int errs, hangs, spawned;
static int noscan;

static void *worker(void *arg) {
    int fds[2];
    pthread_mutex_lock(&big);
    if (pipe(fds) || fcntl(fds[0], F_SETFD, FD_CLOEXEC) || fcntl(fds[1], F_SETFD, FD_CLOEXEC)) { pthread_mutex_unlock(&big); return 0; }
    posix_spawn_file_actions_t fa;
    posix_spawn_file_actions_init(&fa);
    posix_spawn_file_actions_adddup2(&fa, fds[1], 1);
    if (!noscan) {
        for (int fd = 3; fd < 256; fd++) {
            int v = fcntl(fd, F_GETFD);
            if (v >= 0 && (v & FD_CLOEXEC)) posix_spawn_file_actions_addclose(&fa, fd);
        }
    }
    pid_t p = 0;
    char *av[] = {"sh", "-c", "echo x", 0};
    int e = posix_spawn(&p, "/usr/bin/sh", &fa, 0, av, environ);
    posix_spawn_file_actions_destroy(&fa);
    if (e) {
        char m[80]; snprintf(m, sizeof m, "spawn error %d (%s)\n", e, strerror(e)); write(1, m, strlen(m));
        __sync_fetch_and_add(&errs, 1);
        close(fds[0]); close(fds[1]);
        pthread_mutex_unlock(&big);
        return 0;
    }
    __sync_fetch_and_add(&spawned, 1);
    close(fds[1]);
    pthread_mutex_unlock(&big);
    char buf[16]; int n = 0, r;
    /* read until EOF, but not forever */
    fcntl(fds[0], F_SETFL, O_NONBLOCK);
    for (int i = 0; i < 200; i++) {
        r = read(fds[0], buf, sizeof buf);
        if (r == 0) break;
        if (r > 0) n += r;
        else usleep(20000);
        if (i == 199) { __sync_fetch_and_add(&hangs, 1); }
    }
    close(fds[0]);
    int st; waitpid(p, &st, 0);
    return 0;
}

static void round_(int threads, int ns) {
    noscan = ns; errs = hangs = spawned = 0;
    pthread_t t[8];
    for (int i = 0; i < threads; i++) pthread_create(&t[i], 0, worker, 0);
    for (int i = 0; i < threads; i++) pthread_join(t[i], 0);
    printf("threads=%d scan=%s: spawned %d, spawn errors %d, pipe never EOF %d\n", threads, ns ? "no " : "yes", spawned, errs, hangs);
    fflush(stdout);
}

int main(void) {
    round_(1, 0); round_(4, 0); round_(4, 0); round_(4, 1);
    printf("OK x27_spawnpar_c\n");
    _exit(0);
}
