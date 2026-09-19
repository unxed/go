# Runs the C repros against the unpatched and the patched relibc (libc.so used as the ELF
# interpreter of each binary, found through LD_LIBRARY_PATH=/root/mnt/<variant>).
run() { # run <variant> <name> <run#> <limit-seconds>
  v=$1; name=$2; n=$3; lim=${4:-30}
  bin=/root/mnt/bins/${name}_$v.bin
  echo "=== BEGIN $bin variant=$v run=$n"
  rm -f /tmp/rc.done
  ( LD_LIBRARY_PATH=/root/mnt/$v $bin; echo $? > /tmp/rc.done ) &
  sp=$!
  t=0
  while [ ! -f /tmp/rc.done ] && [ $t -lt $lim ]; do sleep 1; t=$((t+1)); done
  if [ -f /tmp/rc.done ]; then rc=`cat /tmp/rc.done`; else echo "HANG $bin"; kill -9 $sp 2>/dev/null; rc=hang; fi
  echo "=== END $bin variant=$v run=$n exit=$rc"
}
v=${1:-patched}
if [ "$v" != debug ] && [ "$v" != lockclose ]; then
run $v x26_spawn_cloexec_c 1 30
run $v x32_poll_regular_c 1 30
n=1; while [ $n -le 3 ]; do run $v x33_thread_sigmask_c $n 60; n=$((n+1)); done
fi
# last: a wedged/panicked VM ends the boot
n=1; while [ $n -le 15 ]; do run $v x27_spawnpar_c $n 60; n=$((n+1)); done
echo "=== LADDER DONE"
