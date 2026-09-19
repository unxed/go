/* Deterministic, cheap trigger for the kernel's exit-after-exception path: many short-lived
 * children whose only thread takes an exception immediately (null store or ud2), spawned from a
 * few parallel workers that wait for them. A wedge (kernel panic or silent freeze) ends the boot.
 *   x37 [workers=4] [iters=50] [spawn|fork] [null|ud2]      x37 child <kind>   (the crashing child) */
#include <pthread.h>
#include <spawn.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/wait.h>
#include <unistd.h>

extern char **environ;
static char *self, *kind, *how;
static int iters;
static volatile int done_count, crashed, spawn_err;

static void *worker(void *a) {
    for (int i = 0; i < iters; i++) {
        pid_t p = 0;
        char *av[] = {self, "child", kind, 0};
        if (!strcmp(how, "fork")) {
            p = fork();
            if (p == 0) { execv(self, av); _exit(99); }
            if (p < 0) { __sync_fetch_and_add(&spawn_err, 1); continue; }
        } else if (posix_spawn(&p, self, 0, 0, av, environ) != 0) {
            __sync_fetch_and_add(&spawn_err, 1);
            continue;
        }
        int st = 0;
        waitpid(p, &st, 0);
        if (!WIFEXITED(st) || WEXITSTATUS(st) != 0) __sync_fetch_and_add(&crashed, 1);
        int d = __sync_add_and_fetch(&done_count, 1);
        if (d % 50 == 0) { printf("x37 progress: %d children\n", d); fflush(stdout); }
    }
    return 0;
}

int main(int argc, char **argv) {
    if (argc > 2 && !strcmp(argv[1], "child")) {
        if (!strcmp(argv[2], "ud2")) __builtin_trap();
        *(volatile int *)0 = 1;
        _exit(0);
    }
    self = argv[0];
    int workers = argc > 1 ? atoi(argv[1]) : 4;
    iters = argc > 2 ? atoi(argv[2]) : 50;
    how = argc > 3 ? argv[3] : "spawn";
    kind = argc > 4 ? argv[4] : "null";
    pthread_t t[16];
    for (int i = 0; i < workers; i++) pthread_create(&t[i], 0, worker, 0);
    for (int i = 0; i < workers; i++) pthread_join(t[i], 0);
    printf("x37: %d children done, %d crashed, %d spawn errors (%s, %s, %d workers)\n", done_count, crashed, spawn_err, how, kind, workers);
    printf("OK x37_crash_loop_c\n");
    fflush(stdout);
    _exit(0);
}
