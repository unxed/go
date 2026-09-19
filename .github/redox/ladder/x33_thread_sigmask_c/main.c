/* A new thread inherits its creator's signal mask (POSIX). The creator blocks every
 * signal, then creates many short-lived threads while a helper thread (SIGUSR1
 * unblocked) sends process-directed SIGUSR1s to the process. The handler may only ever
 * run on the helper: any other thread is a signal delivered to a thread that has it
 * blocked (in relibc: during thread start-up, before the inherited mask is applied). */
#define _GNU_SOURCE
#include <pthread.h>
#include <signal.h>
#include <stdatomic.h>
#include <stdio.h>
#include <unistd.h>

static pthread_t helper;
static atomic_int stop, wrong, total, ready;

static void handler(int sig) {
    atomic_fetch_add(&total, 1);
    if (!pthread_equal(pthread_self(), helper)) atomic_fetch_add(&wrong, 1);
}

static void *helper_main(void *p) {
    sigset_t s;
    sigemptyset(&s);
    sigaddset(&s, SIGUSR1);
    pthread_sigmask(SIG_UNBLOCK, &s, 0);
    while (!atomic_load(&ready)) {}
    while (!atomic_load(&stop)) { kill(getpid(), SIGUSR1); }
    return 0;
}

static void *short_lived(void *p) { return 0; }

int main(void) {
    struct sigaction sa = {0};
    sa.sa_handler = handler;
    sigaction(SIGUSR1, &sa, 0);
    sigset_t all;
    sigfillset(&all);
    pthread_sigmask(SIG_SETMASK, &all, 0);      /* creator blocks everything; children inherit */
    pthread_create(&helper, 0, helper_main, 0);
    atomic_store(&ready, 1);
    for (int i = 0; i < 1500; i++) {
        pthread_t t;
        if (pthread_create(&t, 0, short_lived, 0) == 0) pthread_join(t, 0);
    }
    atomic_store(&stop, 1);
    pthread_join(helper, 0);
    printf("handler runs: %d, on a thread that had SIGUSR1 blocked: %d\n", atomic_load(&total), atomic_load(&wrong));
    printf("%s x33_thread_sigmask_c\n", atomic_load(&wrong) == 0 ? "OK" : "FAIL");
    _exit(0);
}
