/* Pure-C exit-hang reproducer (no Go). NTHR pthreads run sem ping-pong pairs, so
 * at any instant several threads are entering/leaving *indefinite* futex waits.
 * main sleeps a pseudo-random 0..15 ms, prints OK, then _exit(0) (or exit(0) with
 * argv[1]=="exit"). A run that printed OK and never terminated is the exit hang. */
#include <pthread.h>
#include <semaphore.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <unistd.h>

#define PAIRS 3
static sem_t sa[PAIRS], sb[PAIRS];

static void w(sem_t *s) { while (sem_wait(s) != 0) {} }

static void *ping(void *p) {
    long i = (long)p;
    for (;;) { w(&sa[i]); sem_post(&sb[i]); }
    return 0;
}
static void *pong(void *p) {
    long i = (long)p;
    for (;;) { w(&sb[i]); sem_post(&sa[i]); }
    return 0;
}

int main(int argc, char **argv) {
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    unsigned long seed = ts.tv_nsec ^ (unsigned long)getpid() * 2654435761UL;
    for (long i = 0; i < PAIRS; i++) {
        pthread_t t;
        sem_init(&sa[i], 0, 0);
        sem_init(&sb[i], 0, 0);
        pthread_create(&t, 0, ping, (void *)i);
        pthread_create(&t, 0, pong, (void *)i);
        sem_post(&sa[i]);
    }
    usleep((seed >> 7) % 15000);
    printf("OK x17_exit_c\n");
    fflush(stdout);
    if (argc > 1 && strcmp(argv[1], "exit") == 0) exit(0);
    _exit(0);
}
