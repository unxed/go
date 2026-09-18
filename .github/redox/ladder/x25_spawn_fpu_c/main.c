/* relibc posix_spawn: does the new process start with sane FPU control state?
 * The child (this binary with "child") does inexact SSE and x87 arithmetic and
 * prints its MXCSR / x87 control word. With MXCSR == 0 every SSE exception is
 * unmasked and (CR4.OSXMMEXCPT clear) the first inexact result kills the child
 * with "Invalid opcode fault". A fork()ed child inherits sane values. */
#include <spawn.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/wait.h>
#include <unistd.h>
#include <xmmintrin.h>

extern char **environ;

static int run(const char *self, int use_spawn) {
    pid_t p = 0;
    char *av[] = {(char *)self, "child", 0};
    if (use_spawn) {
        if (posix_spawn(&p, self, 0, 0, av, environ) != 0) return -1;
    } else {
        p = fork();
        if (p == 0) { execv(self, av); _exit(99); }
    }
    int st = 0;
    for (int i = 0; i < 300; i++) {
        pid_t r = waitpid(p, &st, WNOHANG);
        if (r == p) return WIFEXITED(st) ? WEXITSTATUS(st) : 100 + (st & 0x7f);
        usleep(20000);
    }
    kill(p, 9);
    return 98;
}

int main(int argc, char **argv) {
    if (argc > 1 && strcmp(argv[1], "child") == 0) {
        unsigned short cw;
        __asm__ volatile("fnstcw %0" : "=m"(cw));
        printf("child: MXCSR=%#x x87CW=%#x\n", _mm_getcsr(), cw);
        fflush(stdout);
        volatile double a = 1.0, b = 3.0;
        volatile double c = a / b;              /* inexact SSE result */
        volatile long double d = (long double)a / (long double)b;   /* inexact x87 result */
        printf("child: fp ok %f %Lf\n", c, d);
        _exit(0);
    }
    int rf = run(argv[0], 0), rs = run(argv[0], 1);
    printf("fork+exec child exit=%d, posix_spawn child exit=%d (0 = ok)\n", rf, rs);
    printf("%s x25_spawn_fpu_c\n", rs == 0 ? "OK" : "FAIL");
    _exit(0);
}
