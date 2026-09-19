/* Background heartbeat: proves the kernel and this process are alive while something else is
 * wedged, and periodically dumps the kernel's view of every context (/scheme/sys/context),
 * which needs no help from procmgr. Runs until killed. */
#include <fcntl.h>
#include <stdio.h>
#include <string.h>
#include <time.h>
#include <unistd.h>

int main(void) {
    for (int n = 1;; n++) {
        struct timespec t = {1, 0};
        nanosleep(&t, 0);
        char m[32];
        int l = snprintf(m, sizeof m, "HB %d\n", n);
        write(1, m, l);
        if (n % 20 == 0) {
            int fd = open("/scheme/sys/context", O_RDONLY);
            if (fd >= 0) {
                char buf[4096];
                int r;
                write(1, "HBDUMP begin\n", 13);
                while ((r = read(fd, buf, sizeof buf)) > 0) write(1, buf, r);
                write(1, "HBDUMP end\n", 11);
                close(fd);
            }
        }
    }
}
