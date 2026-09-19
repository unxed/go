/* Failed posix_spawn calls leave half-built children behind (relibc creates the child process
 * first, then fails, e.g. with ENOENT for a missing program). What happens to them, and can the
 * kernel survive them? Loops N failing spawns, optionally from 4 threads, then waits so that
 * whatever kills the leftovers gets a chance, then exits (exit of the parent is another chance). */
#include <errno.h>
#include <pthread.h>
#include <spawn.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

extern char **environ;
static int n_iter = 100;

static void *worker(void *a) {
    int bad = 0;
    for (int i = 0; i < n_iter; i++) {
        pid_t p = 0;
        char *av[] = {"nonexistent", 0};
        int e = posix_spawn(&p, "/nonexistent/prog", 0, 0, av, environ);
        if (e == 0) bad++;
    }
    return (void *)(long)bad;
}

int main(int argc, char **argv) {
    int threads = argc > 1 ? atoi(argv[1]) : 1;
    pthread_t t[8];
    for (int i = 0; i < threads; i++) pthread_create(&t[i], 0, worker, 0);
    long bad = 0;
    for (int i = 0; i < threads; i++) { void *r; pthread_join(t[i], &r); bad += (long)r; }
    printf("%d threads x %d failing spawns done, unexpected successes: %ld\n", threads, n_iter, bad);
    fflush(stdout);
    usleep(500000);
    printf("OK x34_spawn_fail_c\n");
    fflush(stdout);
    _exit(0);
}
