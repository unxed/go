# Sourced by run_ladder.sh / run_exit.sh inside the Redox VM.
export GOTRACEBACK=all
# Never `wait` on a possibly wedged process (kill -9 does not always free it
# on Redox): run it in a background subshell that records its exit code, poll
# for that, and move on either way. Optional 4th arg: poll limit in seconds.
run_one() {
  bin="$1"; variant="$2"; n="$3"
  echo "=== BEGIN $bin variant=$variant run=$n"
  rm -f /tmp/rc.done
  (
    case "$variant" in
      default) $bin ;;
      preempt) GODEBUG=asyncpreemptoff=0 $bin ;;
      procs1) GOMAXPROCS=1 $bin ;;
      stdin) printf 'alpha\nbeta\ngamma\n' | $bin ;;
      exitarg) $bin exit ;;
    esac
    echo $? > /tmp/rc.done
  ) &
  sp=$!
  t=0
  while [ ! -f /tmp/rc.done ] && [ $t -lt ${4:-12} ]; do sleep 1; t=$((t+1)); done
  if [ -f /tmp/rc.done ]; then
    rc=`cat /tmp/rc.done`
  else
    echo "HANG $bin"
    # kernel view of every thread (STAT: U=user/K=kernel, then R=runnable
    # S=blocked-but-awake B=blocked-asleep Z=dead; + = on a CPU)
    b=`basename $bin .bin`
    echo "--- /scheme/sys/context: header, rows of $b"
    head -1 /scheme/sys/context 2>&1
    grep "$b" /scheme/sys/context 2>&1
    echo "--- full table (first 80 rows)"
    head -80 /scheme/sys/context 2>&1
    echo "---"
    # Go programs: SIGQUIT makes the runtime print every goroutine's stack
    set -- `grep "$b" /scheme/sys/context | head -1`
    if [ -n "$1" ]; then
      echo "--- SIGQUIT to pid $1"
      kill -QUIT $1 2>&1
      sleep 3
    fi
    # kernel-side view (only kernels built with kernel/debug/000[23]-*ring*.patch have it): the
    # first 3 hangs of a boot
    if [ -e /scheme/sys/dbgring ]; then
      nh=`cat /tmp/nhang 2>/dev/null || echo 0`
      if [ $nh -lt 3 ]; then
        echo $((nh+1)) > /tmp/nhang
        echo "--- /scheme/sys/dbgring (hang #$((nh+1)))"
        cat /scheme/sys/dbgring 2>&1
        echo "--- full context table"
        cat /scheme/sys/context 2>&1
        echo "--- end dbgring"
      fi
    fi
    kill -9 $sp 2>/dev/null
    rc=hang
  fi
  echo "=== END $bin variant=$variant run=$n exit=$rc"
}
loop_run() { # loop_run bin variant count [limit]
  i=1
  while [ $i -le $3 ]; do run_one $1 $2 $i $4; i=$((i+1)); done
}
