/* Hypothesis test (pure C): a fork() in a multithreaded process makes the parent's
 * pages copy-on-write. The kernel keys futexes by *physical* address, so when the
 * parent writes to the semaphore word (sem_post) its page is copied to a new
 * physical page, and FUTEX_WAKE no longer finds the thread that went to sleep on
 * the old page => the sleeper is never woken.
 *
 * For each round: a helper thread sleeps in sem_wait(); the main thread (a)
 * optionally fork()s and reaps a child, (b) sem_post()s. The helper must wake.
 * Also tried: the helper is created *before* fork, and the semaphore lives on the
 * heap vs. in .bss. */
#include <pthread.h>
#include <spawn.h>
#include <semaphore.h>
#include <stdatomic.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/wait.h>
#include <time.h>
#include <unistd.h>

static sem_t bss_sem;
static sem_t *heap_sem;
static _Atomic int woken;

static void *sleeper(void *p) {
    sem_t *s = p;
    while (sem_wait(s) != 0) {}
    atomic_store(&woken, 1);
    return 0;
}

static char *self;
extern char **environ;

/* mode 0: no fork; 1: fork, child exits at once and is reaped (page is exclusive again);
 * 2: fork, child stays alive for 400 ms across the sem_post (page stays shared => CoW copy);
 * 3: posix_spawn(self "child"), child stays alive for 400 ms across the sem_post */
static int round_(int mode, int use_heap) {
    sem_t *s = use_heap ? heap_sem : &bss_sem;
    sem_init(s, 0, 0);
    atomic_store(&woken, 0);
    pthread_t t;
    pthread_create(&t, 0, sleeper, s);
    struct timespec d = {0, 150 * 1000 * 1000};
    nanosleep(&d, 0);                      /* let the helper block in the futex */
    pid_t c = 0;
    if (mode == 1 || mode == 2) {
        c = fork();
        if (c == 0) {
            if (mode == 2) { struct timespec k = {0, 400 * 1000 * 1000}; nanosleep(&k, 0); }
            _exit(0);
        }
        if (mode == 1) { int st; waitpid(c, &st, 0); c = 0; }
    } else if (mode == 3) {
        char *av[] = {self, "child", 0};
        if (posix_spawn(&c, self, 0, 0, av, environ) != 0) { printf("posix_spawn failed\n"); c = 0; }
    }
    sem_post(s);                           /* writes the page that fork() made CoW */
    for (int i = 0; i < 100; i++) {        /* up to 2 s */
        if (atomic_load(&woken)) break;
        struct timespec w = {0, 20 * 1000 * 1000};
        nanosleep(&w, 0);
    }
    int ok = atomic_load(&woken);
    if (ok) pthread_join(t, 0);            /* on a lost wake-up the helper stays asleep: leak it */
    if (c > 0) { int st; waitpid(c, &st, 0); }
    return ok;
}

int main(int argc, char **argv) {
    self = argv[0];
    if (argc > 1 && strcmp(argv[1], "child") == 0) {
        struct timespec k = {0, 400 * 1000 * 1000};
        nanosleep(&k, 0);
        _exit(0);
    }
    heap_sem = malloc(sizeof *heap_sem);
    int fails = 0;
    static const char *names[] = {"no fork", "fork, child gone", "fork, child alive", "posix_spawn, child alive"};
    for (int use_heap = 0; use_heap < 2; use_heap++)
        for (int mode = 0; mode < 4; mode++) {
            int okc = 0, n = 4;
            for (int i = 0; i < n; i++) okc += round_(mode, use_heap);
            printf("%s sem, %-24s: %d/%d woke\n", use_heap ? "heap" : "bss ", names[mode], okc, n);
            fflush(stdout);
            if ((mode == 0 || mode == 3) && okc != n) fails++;
        }
    printf("%s x23_futex_fork_c\n", fails ? "FAIL" : "OK");
    _exit(0);
}
