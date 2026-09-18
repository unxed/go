/* poll() on a pty pair (pure C): does the slave / master report POLLIN when data
 * arrives while poll() is already waiting? Canonical and raw mode (cfmakeraw).
 * A helper thread writes to the other side after 300 ms; poll waits up to 2 s. */
#define _GNU_SOURCE
#include <fcntl.h>
#include <poll.h>
#include <pthread.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <termios.h>
#include <time.h>
#include <unistd.h>

static int m, s;
static int wfd;
static const char *wmsg;

static void *writer(void *p) {
    usleep(300000);
    write(wfd, wmsg, strlen(wmsg));
    return 0;
}

static long long ms(void) { struct timespec t; clock_gettime(CLOCK_MONOTONIC, &t); return t.tv_sec * 1000LL + t.tv_nsec / 1000000; }

static void try_poll(const char *label, int pfd, int wr_to, const char *msg) {
    pthread_t t;
    wfd = wr_to; wmsg = msg;
    pthread_create(&t, 0, writer, 0);
    struct pollfd p = {pfd, POLLIN, 0};
    long long t0 = ms();
    int n = poll(&p, 1, 2000);
    printf("%-34s poll -> n=%d revents=%#x after %lld ms\n", label, n, p.revents, ms() - t0);
    fflush(stdout);
    pthread_join(t, 0);
    char b[64];
    fcntl(pfd, F_SETFL, O_NONBLOCK);
    int r = read(pfd, b, sizeof b);
    printf("%-34s read after poll -> %d\n", "", r);
    fcntl(pfd, F_SETFL, 0);
}

int main(void) {
    m = posix_openpt(O_RDWR | O_NOCTTY);
    if (m < 0 || grantpt(m) || unlockpt(m)) { printf("FAIL openpt\n"); return 0; }
    char name[64];
    if (ptsname_r(m, name, sizeof name)) { printf("FAIL ptsname\n"); return 0; }
    s = open(name, O_RDWR | O_NOCTTY);
    printf("master fd %d, slave %s fd %d\n", m, name, s);
    if (s < 0) { printf("FAIL open slave\n"); return 0; }
    try_poll("canonical: slave, master writes line", s, m, "hello\n");
    try_poll("canonical: master, slave writes", m, s, "world\n");
    struct termios t;
    if (tcgetattr(s, &t) == 0) {
        cfmakeraw(&t);
        printf("tcsetattr(raw) -> %d\n", tcsetattr(s, TCSANOW, &t));
    } else printf("tcgetattr failed\n");
    try_poll("raw: slave, master writes 1 byte", s, m, "x");
    try_poll("raw: master, slave writes 1 byte", m, s, "y");
    printf("OK x29_pty_c\n");
    _exit(0);
}
