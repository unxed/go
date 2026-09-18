# Runs the cross-compiled std test binaries (go test -c) inside the Redox VM.
# Layout: /root/mnt/<pkg>.test and /root/mnt/data/<pkg>/ (testdata, cwd of the run).
# run_test <name> [limit] [label]: label distinguishes variants of the same binary (log key)
run_test() {
  n=$1; limit=${2:-60}; label=${3:-$1}
  echo "=== BEGIN /root/mnt/$label.test variant=test run=1"
  rm -f /tmp/rc.done /tmp/$n.out
  extra=""; [ "$label" != "$n" ] && extra="1"
  (
    cd /root/mnt/data/$n
    if [ -n "$extra" ]; then export GODEBUG=asyncpreemptoff=0; fi
    /root/mnt/$n.test -test.v -test.short -test.timeout=${limit}s > /tmp/$n.out 2>&1
    echo $? > /tmp/rc.done
  ) &
  t=0; fin=0
  while [ ! -f /tmp/rc.done ] && [ $t -lt $limit ]; do
    sleep 1; t=$((t+1))
    [ $((t % 15)) -eq 0 ] && echo "... $n still running after ${t}s"
    # the run has printed its verdict (last line PASS/FAIL): give the process a few seconds to exit
    last=`tail -1 /tmp/$n.out`
    if [ $fin -eq 0 ] && { [ "$last" = "PASS" ] || [ "$last" = "FAIL" ]; }; then fin=$t; fi
    if [ $fin -gt 0 ] && [ $t -ge $((fin+4)) ]; then break; fi
  done
  if [ -f /tmp/rc.done ]; then rc=`cat /tmp/rc.done`; else rc=noexit; fi
  [ "$rc" = noexit ] && echo "(process did not exit: exit hang)"
  echo "--- top-level results of $n"
  grep -- '--- ' /tmp/$n.out
  echo "--- failures (any level) of $n"
  grep -- '--- FAIL' /tmp/$n.out
  echo "--- tail of $n"
  if [ "$rc" = 0 ]; then tail -3 /tmp/$n.out; else tail -70 /tmp/$n.out; fi
  echo "=== END /root/mnt/$label.test variant=test run=1 exit=$rc after=${t}s"
}
for n in strings bytes sort container_list encoding_json sync time; do
  run_test $n 60
done
# async preemption re-enabled: sync (AllocsPerRun => GOMAXPROCS(1) => stop-the-world) and time
# (TestIssue5745 spins) need it; risk: relibc's sigentry RCX bug on unpatched images
run_test sync 60 sync_preempt
run_test time 60 time_preempt
echo "=== LADDER DONE"
