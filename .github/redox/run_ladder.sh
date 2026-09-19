. /root/mnt/common.sh
# file I/O and the basics first: they must always get to run
run_one /root/mnt/p01_println.bin default 1
run_one /root/mnt/p06_fileio.bin default 1
run_one /root/mnt/p07_openat_diag.bin default 1
run_one /root/mnt/p03_goroutines.bin default 1
run_one /root/mnt/p05_timer.bin default 1
# "default" = runtime defaults (async preemption OFF on redox, see
# runtime1.go); "preempt" = GODEBUG=asyncpreemptoff=0 re-enables the signal
# path that trips over relibc's RCX-clobbering sigentry
loop_run /root/mnt/p02_fmt.bin default 3
loop_run /root/mnt/p02_fmt.bin procs1 2
loop_run /root/mnt/p02_fmt.bin preempt 2
# std coverage
run_one /root/mnt/p10_stdlib.bin default 1
run_one /root/mnt/p11_time_rand.bin default 1
run_one /root/mnt/p12_stdin.bin stdin 1
run_one /root/mnt/p13_signal.bin default 1
run_one /root/mnt/x15_net.bin default 1
run_one /root/mnt/x24_nilpanic.bin default 1
run_one /root/mnt/x29_pty_c.bin default 1 30
run_one /root/mnt/x30_pty.bin default 1 30
run_one /root/mnt/x28_cgo.bin default 1 30
run_one /root/mnt/x10_sigstorm.bin default 1
run_one /root/mnt/x10_sigstorm.bin preempt 1
# memory/GC-heavy rungs last: they have occasionally frozen the whole VM
run_one /root/mnt/p04_gc.bin default 1
run_one /root/mnt/p09_stack.bin default 1
run_one /root/mnt/p08_memzero.bin default 1
# clock/timing diagnostics (p11 saw Sleep(20ms) return after 17.7ms)
run_one /root/mnt/x21_clock_c.bin default 1 60
run_one /root/mnt/x19_clock.bin default 1 60
# fork/exec last: they can freeze the whole VM, so nothing else may come after them
run_one /root/mnt/x23_futex_fork_c.bin default 1 60
run_one /root/mnt/x27_spawnpar_c.bin default 1 60
run_one /root/mnt/x26_spawn_cloexec_c.bin default 1 30
run_one /root/mnt/x31_spawn_actions_c.bin default 1 30
run_one /root/mnt/x18_fork_c.bin default 1
run_one /root/mnt/x22_forkpar_c.bin default 1
run_one /root/mnt/x20_execpar.bin default 1
run_one /root/mnt/p14_exec.bin default 1
echo "=== LADDER DONE"
