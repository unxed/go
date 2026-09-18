/* fork/exec probe in pure C (no Go): is fork() from a multithreaded process the
 * problem behind the os/exec rung freezing the VM? Every stage is announced
 * before it runs so the last line in the log names the stage that hangs.
 *   ./x18 [threads]   threads = number of spinning/sleeping helper threads (default 0..2 staged) */
#include <pthread.h>
#include <semaphore.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/wait.h>
#include <unistd.h>

static void say(const char *s) { write(1, s, strlen(s)); write(1, "\n", 1); }

static sem_t forever;
static void *sleeper(void *p) { for (;;) { while (sem_wait(&forever) != 0) {} } return 0; }
static volatile int stop;
static void *spinner(void *p) { volatile unsigned long x = 0; while (!stop) x++; return 0; }

static int fork_exit(void) {
    pid_t p = fork();
    if (p == 0) _exit(7);
    if (p < 0) return -1;
    int st = 0;
    if (waitpid(p, &st, 0) != p) return -2;
    return WIFEXITED(st) ? WEXITSTATUS(st) : -3;
}
static int fork_exec(char *self) {
    pid_t p = fork();
    if (p == 0) {
        char *av[] = {self, "child", 0};
        execv(self, av);
        _exit(99);
    }
    if (p < 0) return -1;
    int st = 0;
    if (waitpid(p, &st, 0) != p) return -2;
    return WIFEXITED(st) ? WEXITSTATUS(st) : -3;
}

int main(int argc, char **argv) {
    if (argc > 1 && strcmp(argv[1], "child") == 0) { say("child exec'd ok"); _exit(5); }
    char msg[128];
    int r;
    say("stage 1: fork+_exit, single thread");
    r = fork_exit(); snprintf(msg, sizeof msg, "  -> %d (want 7)", r); say(msg);
    say("stage 2: fork+execv self, single thread");
    r = fork_exec(argv[0]); snprintf(msg, sizeof msg, "  -> %d (want 5)", r); say(msg);
    sem_init(&forever, 0, 0);
    pthread_t t;
    for (int i = 0; i < 2; i++) pthread_create(&t, 0, sleeper, 0);
    say("stage 3: fork+_exit with 2 sleeping threads");
    r = fork_exit(); snprintf(msg, sizeof msg, "  -> %d (want 7)", r); say(msg);
    say("stage 4: fork+execv with 2 sleeping threads");
    r = fork_exec(argv[0]); snprintf(msg, sizeof msg, "  -> %d (want 5)", r); say(msg);
    pthread_create(&t, 0, spinner, 0);
    say("stage 5: fork+_exit with a spinning thread too");
    r = fork_exit(); snprintf(msg, sizeof msg, "  -> %d (want 7)", r); say(msg);
    say("stage 6: fork+execv with a spinning thread too");
    r = fork_exec(argv[0]); snprintf(msg, sizeof msg, "  -> %d (want 5)", r); say(msg);
    stop = 1;
    say("OK x18_fork_c");
    _exit(0);
}
