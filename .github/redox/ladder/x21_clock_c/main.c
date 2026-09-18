/* Kernel-level clock check without Go: (1) does nanosleep ever return early;
 * (2) do clock readings taken on different threads, strictly ordered by a
 * handshake, ever go backwards (per-CPU clock skew)? Checked for both
 * CLOCK_MONOTONIC and CLOCK_REALTIME. */
#include <pthread.h>
#include <sched.h>
#include <stdatomic.h>
#include <stdio.h>
#include <time.h>
#include <unistd.h>

static long long ns(clockid_t c) { struct timespec t; clock_gettime(c, &t); return t.tv_sec * 1000000000LL + t.tv_nsec; }

static clockid_t clk;
static _Atomic long long tsA, tsB;
static _Atomic int seq, ack;
static long long neg, worst;
#define N 4000

static void *reader(void *p) {
    for (int i = 1; i <= N; i++) {
        while (atomic_load(&seq) != i) sched_yield();
        long long t = ns(clk), a = atomic_load(&tsA);
        if (t < a) { neg++; if (a - t > worst) worst = a - t; }
        atomic_store(&tsB, t);
        atomic_store(&ack, i);
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
        clk = k == 0 ? CLOCK_MONOTONIC : CLOCK_REALTIME;
        neg = worst = 0; atomic_store(&seq, 0); atomic_store(&ack, 0);
        pthread_t t; pthread_create(&t, 0, reader, 0);
        long long neg2 = 0, worst2 = 0;
        for (int i = 1; i <= N; i++) {
            atomic_store(&tsA, ns(clk));
            atomic_store(&seq, i);
            while (atomic_load(&ack) != i) sched_yield();
            long long t2 = ns(clk), b = atomic_load(&tsB);
            if (t2 < b) { neg2++; if (b - t2 > worst2) worst2 = b - t2; }
        }
        pthread_join(t, 0);
        printf("%s across threads (%d handshakes): A->B backwards %lld (worst %lld ns), B->A backwards %lld (worst %lld ns)\n",
               k == 0 ? "MONOTONIC" : "REALTIME", N, neg, worst, neg2, worst2);
    }
    printf("OK x21_clock_c\n");
    _exit(0);
}
