/* Kernel-level clock check without Go: (1) does nanosleep/usleep ever return
 * early; (2) do CLOCK_MONOTONIC readings taken on different threads go backwards
 * (per-CPU clock skew)? Also prints CLOCK_REALTIME skew the same way. */
#include <pthread.h>
#include <stdio.h>
#include <stdatomic.h>
#include <time.h>
#include <unistd.h>

static long long ns(clockid_t c) { struct timespec t; clock_gettime(c, &t); return t.tv_sec * 1000000000LL + t.tv_nsec; }
static _Atomic long long shared[2];
static long long neg[2], worst[2];
#define N 300000

static void *other(void *p) {
    clockid_t c = (clockid_t)(long)p; int k = c == CLOCK_MONOTONIC ? 0 : 1;
    long long last = 0;
    for (int i = 0; i < N; i++) {
        while (atomic_load(&shared[k]) == last) {}
        last = atomic_load(&shared[k]);
        long long now = ns(c);
        if (now < last) { neg[k]++; if (last - now > worst[k]) worst[k] = last - now; }
    }
    return 0;
}

int main(void) {
    int early = 0; long long minel = 1LL << 60;
    for (int i = 0; i < 100; i++) {
        long long t0 = ns(CLOCK_MONOTONIC);
        struct timespec r = {0, 5000000};
        nanosleep(&r, 0);
        long long el = ns(CLOCK_MONOTONIC) - t0;
        if (el < 5000000) early++;
        if (el < minel) minel = el;
    }
    printf("nanosleep(5ms) x100: %d early, min %lld ns\n", early, minel);
    for (int k = 0; k < 2; k++) {
        clockid_t c = k == 0 ? CLOCK_MONOTONIC : CLOCK_REALTIME;
        pthread_t t; pthread_create(&t, 0, other, (void *)(long)c);
        for (int i = 1; i <= N; i++) {
            atomic_store(&shared[k], ns(c));
            usleep(0); sched_yield();
            /* wait until the reader consumed: it reads then waits for the next change */
            long long v = atomic_load(&shared[k]); (void)v;
        }
        pthread_join(t, 0);
        printf("%s cross-thread: %lld backwards steps, worst %lld ns\n", k == 0 ? "MONOTONIC" : "REALTIME", neg[k], worst[k]);
    }
    printf("OK x21_clock_c\n");
    _exit(0);
}
