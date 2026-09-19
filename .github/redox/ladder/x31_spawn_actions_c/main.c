/* Which posix_spawn file-action layouts does relibc reject with EBADF? Single thread,
 * three CLOEXEC files (/dev/null x1, pipes) dup2'ed onto 0,1,2 like os/exec does, with and
 * without close actions for the parent's close-on-exec descriptors. */
#include <errno.h>
#include <fcntl.h>
#include <spawn.h>
#include <stdio.h>
#include <string.h>
#include <sys/wait.h>
#include <unistd.h>

extern char **environ;

/* mode bit0: close all CLOEXEC fds >= 3; bit1: but not the dup2 sources; bit2: dup2 only fds 0 and 1 */
static void run(int mode) {
    int in = open("/dev/null", O_RDONLY | O_CLOEXEC);
    int p1[2], p2[2];
    pipe(p1); fcntl(p1[0], F_SETFD, FD_CLOEXEC); fcntl(p1[1], F_SETFD, FD_CLOEXEC);
    pipe(p2); fcntl(p2[0], F_SETFD, FD_CLOEXEC); fcntl(p2[1], F_SETFD, FD_CLOEXEC);
    int src[3] = {in, p1[1], p2[1]};
    posix_spawn_file_actions_t fa;
    posix_spawn_file_actions_init(&fa);
    int n = (mode & 4) ? 2 : 3;
    for (int i = 0; i < n; i++) posix_spawn_file_actions_adddup2(&fa, src[i], i);
    if (mode & 1) {
        for (int fd = 3; fd < 64; fd++) {
            int v = fcntl(fd, F_GETFD);
            if (v < 0 || !(v & FD_CLOEXEC)) continue;
            int is_src = 0;
            for (int i = 0; i < n; i++) if (src[i] == fd) is_src = 1;
            if ((mode & 2) && is_src) continue;
            posix_spawn_file_actions_addclose(&fa, fd);
        }
    }
    pid_t p = 0;
    char *av[] = {"sh", "-c", "echo hi", 0};
    int e = posix_spawn(&p, "/usr/bin/sh", &fa, 0, av, environ);
    printf("mode %d (closes=%d, srcs-kept=%d, dup2 %d fds): srcs %d,%d,%d -> %s\n", mode, mode & 1, (mode >> 1) & 1, n, src[0], src[1], src[2], e ? strerror(e) : "ok");
    fflush(stdout);
    posix_spawn_file_actions_destroy(&fa);
    if (!e) { int st; waitpid(p, &st, 0); }
    close(in); close(p1[0]); close(p1[1]); close(p2[0]); close(p2[1]);
}

int main(void) {
    int modes[] = {0, 1, 3, 4, 5, 7};
    for (int i = 0; i < 6; i++) run(modes[i]);
    printf("OK x31_spawn_actions_c\n");
    _exit(0);
}
