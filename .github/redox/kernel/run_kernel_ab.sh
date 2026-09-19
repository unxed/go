. /root/mnt/common.sh
# futex vs fork (x23): "fork, child alive" must wake 4/4 (0002-*.patch); the child-gone case is racy
run_one /root/mnt/x23_futex_fork_c.bin default 1 90
run_one /root/mnt/x23_futex_fork_c.bin default 2 90
# Concurrent posix_spawn stress (a kernel panic there ends the boot): does the patch matter?
loop_run /root/mnt/x27_spawnpar_c.bin default 40 60
# A/B probe for the exit hang, pure C (x17): exit(0) hung ~40% on the stock kernel.
loop_run /root/mnt/x17_exit_c.bin exitarg 60 5
loop_run /root/mnt/x17_exit_c.bin default 30 5
echo "=== LADDER DONE"
