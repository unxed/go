#!/bin/sh
# usage: vm_run.sh <script-in-out-dir> <logfile>
# One QEMU boot under redoxer (no KVM on hosted runners => TCG, ~1 min boot).
# `-f /mnt` copies the host folder into the VM as /root/mnt/. A host-side idle
# timer kills a frozen VM (no output for 75 s).
script=$1; log=$2
docker pull redoxos/redoxer >/dev/null
: > $log
( timeout 300 docker run --name ladder --rm -e REDOXER_QEMU_ARGS='-smp 2' -v /tmp/pkg/out:/mnt redoxos/redoxer \
    redoxer exec -f /mnt -- sh /root/mnt/$script > $log 2>&1 ) &
dpid=$!
last=0; idle=0
while kill -0 $dpid 2>/dev/null; do
  sleep 5
  size=$(stat -c %s $log)
  if [ "$size" = "$last" ]; then idle=$((idle+5)); else idle=0; last=$size; fi
  grep -aq "LADDER DONE" $log && break
  if [ $idle -ge 75 ]; then
    echo "VM FROZE: no output for ${idle}s; last line: $(tail -c 200 $log | tr -d '\r' | tail -1)" | tee -a $log
    docker kill ladder 2>/dev/null
    break
  fi
done
docker kill ladder 2>/dev/null; wait $dpid 2>/dev/null
tail -5 $log
