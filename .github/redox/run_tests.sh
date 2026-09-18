# Runs the cross-compiled std test binaries (go test -c) inside the Redox VM.
# Layout: /root/mnt/<pkg>.test and /root/mnt/data/<pkg>/ (testdata, cwd of the run).
run_test() {
  n=$1; limit=${2:-100}
  echo "=== BEGIN /root/mnt/$n.test variant=test run=1"
  rm -f /tmp/rc.done /tmp/$n.out
  (
    cd /root/mnt/data/$n
    /root/mnt/$n.test -test.v -test.short -test.timeout=${limit}s > /tmp/$n.out 2>&1
    echo $? > /tmp/rc.done
  ) &
  t=0; fin=0
  while [ ! -f /tmp/rc.done ] && [ $t -lt $limit ]; do
    sleep 1; t=$((t+1))
    # the run has printed its verdict: allow a few seconds for the process to exit
    if [ $fin -eq 0 ] && grep '^PASS' /tmp/$n.out >/dev/null 2>&1; then fin=$t; fi
    if [ $fin -eq 0 ] && grep '^FAIL' /tmp/$n.out >/dev/null 2>&1; then fin=$t; fi
    if [ $fin -gt 0 ] && [ $t -ge $((fin+4)) ]; then break; fi
  done
  if [ -f /tmp/rc.done ]; then rc=`cat /tmp/rc.done`; else rc=noexit; fi
  echo "--- top-level results of $n"
  grep '^--- ' /tmp/$n.out
  echo "--- failures (any level) of $n"
  grep -- '--- FAIL' /tmp/$n.out
  echo "--- tail of $n"
  tail -12 /tmp/$n.out
  echo "=== END /root/mnt/$n.test variant=test run=1 exit=$rc after=${t}s"
}
for n in strings bytes sort container_list encoding_json sync time; do
  run_test $n 100
done
echo "=== LADDER DONE"
