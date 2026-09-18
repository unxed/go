/* relibc posix_spawn: are the parent's close-on-exec descriptors closed in the
 * child? The parent makes a pipe (both ends FD_CLOEXEC), spawns `cat` with the
 * read end dup2'ed onto its stdin, writes a line, closes its own write end, and
 * expects `cat` to see EOF and exit. If the child kept a copy of the write end
 * (open in its own table), cat never sees EOF and hangs. Same test with fork+exec
 * as the control. */
#include <fcntl.h>
#include <spawn.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/wait.h>
#include <unistd.h>

extern char **environ;

static int run(int use_spawn) {
    int fds[2];
    if (pipe(fds)) return -1;
    fcntl(fds[0], F_SETFD, FD_CLOEXEC);
    fcntl(fds[1], F_SETFD, FD_CLOEXEC);
    pid_t p = 0;
    char *av[] = {"cat", 0};
    if (use_spawn) {
        posix_spawn_file_actions_t fa;
        posix_spawn_file_actions_init(&fa);
        posix_spawn_file_actions_adddup2(&fa, fds[0], 0);
        int e = posix_spawn(&p, "/usr/bin/cat", &fa, 0, av, environ);
        posix_spawn_file_actions_destroy(&fa);
        if (e) { printf("posix_spawn error %d\n", e); return -1; }
    } else {
        p = fork();
        if (p == 0) { dup2(fds[0], 0); execv("/usr/bin/cat", av); _exit(99); }
    }
    const char *msg = "hello through cat\n";
    write(fds[1], msg, strlen(msg));
    close(fds[1]);
    close(fds[0]);
    int st = 0;
    for (int i = 0; i < 150; i++) {
        pid_t r = waitpid(p, &st, WNOHANG);
        if (r == p) return 0;
        usleep(20000);
    }
    kill(p, 9);
    waitpid(p, &st, 0);
    return 1;   /* cat never saw EOF */
}

int main(void) {
    int rf = run(0), rs = run(1);
    printf("fork+exec: %s; posix_spawn: %s\n", rf == 0 ? "cat exited" : "cat HUNG", rs == 0 ? "cat exited" : "cat HUNG");
    printf("%s x26_spawn_cloexec_c\n", rf == 0 && rs == 0 ? "OK" : "FAIL");
    _exit(0);
}
