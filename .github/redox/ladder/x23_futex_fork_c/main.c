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

static int round_(int do_fork, int use_heap) {
    sem_t *s = use_heap ? heap_sem : &bss_sem;
    sem_init(s, 0, 0);
    atomic_store(&woken, 0);
    pthread_t t;
    pthread_create(&t, 0, sleeper, s);
    struct timespec d = {0, 150 * 1000 * 1000};
    nanosleep(&d, 0);                      /* let the helper block in the futex */
    if (do_fork) {
        pid_t c = fork();
        if (c == 0) _exit(0);
        int st; waitpid(c, &st, 0);
    }
    sem_post(s);                           /* writes the page that fork() made CoW */
    for (int i = 0; i < 100; i++) {        /* up to 2 s */
        if (atomic_load(&woken)) break;
        struct timespec w = {0, 20 * 1000 * 1000};
        nanosleep(&w, 0);
    }
    int ok = atomic_load(&woken);
    if (ok) pthread_join(t, 0);            /* on a lost wake-up the helper stays asleep: leak it */
    return ok;
}

int main(void) {
    heap_sem = malloc(sizeof *heap_sem);
    int fails = 0;
    for (int use_heap = 0; use_heap < 2; use_heap++)
        for (int do_fork = 0; do_fork < 2; do_fork++) {
            int okc = 0, n = 3;
            for (int i = 0; i < n; i++) okc += round_(do_fork, use_heap);
            printf("%s sem, %s: %d/%d woke\n", use_heap ? "heap" : "bss", do_fork ? "fork before sem_post" : "no fork         ", okc, n);
            fflush(stdout);
            if (!do_fork && okc != n) fails++;
        }
    printf("%s x23_futex_fork_c\n", fails ? "FAIL" : "OK");
    _exit(0);
}
