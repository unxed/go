. /root/mnt/common.sh
if [ "$1" = dbg ]; then
  # debug kernel: only the spawn stress, as long as it takes to wedge (DBGDEAD lines say why)
  /root/mnt/x36_heartbeat_c.bin &
  loop_run /root/mnt/x27_spawnpar_c.bin default 80 60
  echo "=== LADDER DONE"
  exit 0
fi
# futex vs fork (x23): "fork, child alive" must wake 4/4 (0002-*.patch)
run_one /root/mnt/x23_futex_fork_c.bin default 1 90
run_one /root/mnt/x23_futex_fork_c.bin default 2 90
# exit hang repro (x17) BEFORE the spawn stress: the spawn stress wedges the VM on every kernel
# variant sooner or later (upstream issue #9), which would cut the boot short
loop_run /root/mnt/x17_exit_c.bin exitarg 60 5
loop_run /root/mnt/x17_exit_c.bin default 30 5
# failing spawns (half-built children) then exit
loop_run /root/mnt/x34_spawn_fail_c.bin default 5 60
# concurrent posix_spawn stress last
loop_run /root/mnt/x27_spawnpar_c.bin default 40 60
echo "=== LADDER DONE"
