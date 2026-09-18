/* Pure-C counterpart of x20: 4 threads each fork+exec `sh -c "echo x"` with a
 * pipe, serialized by a mutex around fork like Go's ForkLock, then read + waitpid
 * concurrently. Stages are announced first. */
#include <pthread.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/wait.h>
#include <unistd.h>
#include <fcntl.h>

static pthread_mutex_t forklock = PTHREAD_MUTEX_INITIALIZER;
static int usepipe;
static int bad;

static void *worker(void *p) {
    int fds[2];
    if (usepipe) { pthread_mutex_lock(&forklock); if (pipe(fds) || fcntl(fds[0], F_SETFD, FD_CLOEXEC) || fcntl(fds[1], F_SETFD, FD_CLOEXEC)) { bad++; pthread_mutex_unlock(&forklock); return 0; } }
    else pthread_mutex_lock(&forklock);
    pid_t c = fork();
    if (c == 0) {
        if (usepipe) dup2(fds[1], 1);
        char *av[] = {"sh", "-c", "echo x", 0};
        execvp("sh", av);
        _exit(99);
    }
    pthread_mutex_unlock(&forklock);
    if (c < 0) { bad++; return 0; }
    if (usepipe) {
        close(fds[1]);
        char buf[16]; int n = 0, r;
        while ((r = read(fds[0], buf + n, sizeof buf - n)) > 0) n += r;
        close(fds[0]);
        if (n < 1) bad++;
    }
    int st;
    if (waitpid(c, &st, 0) != c || !WIFEXITED(st) || WEXITSTATUS(st) != 0) bad++;
    return 0;
}

static void stage(const char *name, int threads, int pipes) {
    char m[96]; snprintf(m, sizeof m, "stage: %s x%d\n", name, threads); write(1, m, strlen(m));
    usepipe = pipes; bad = 0;
    pthread_t t[8];
    for (int i = 0; i < threads; i++) pthread_create(&t[i], 0, worker, 0);
    for (int i = 0; i < threads; i++) pthread_join(t[i], 0);
    if (bad) { snprintf(m, sizeof m, "FAIL %s: %d bad\n", name, bad); write(1, m, strlen(m)); }
}

int main(void) {
    stage("no pipes", 1, 0);
    stage("no pipes", 4, 0);
    stage("pipes", 1, 1);
    stage("pipes", 2, 1);
    stage("pipes", 4, 1);
    write(1, "OK x22_forkpar_c\n", 17);
    _exit(0);
}
