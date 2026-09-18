#!/bin/sh
# usage: vm_run.sh <script-in-out-dir> <logfile> [kernel-file]
#   out dir: $PKGOUT (default /tmp/pkg/out), copied by redoxer into the VM as /root/mnt/
#   env: SMP (default 4), VM_TIMEOUT (default 240 s), VM_IDLE (default 45 s of silence => VM considered frozen)
# One QEMU boot under redoxer. KVM is used when the host has /dev/kvm (see
# enable_kvm.sh; redoxer auto-detects it inside the container). With a kernel
# file, redoxer's base image (a tar of the installed packages in ~/.redoxer) is
# built first, its usr/lib/boot/kernel is swapped for the given one, then the
# real command runs.
script=$1; log=$2; kernel=$3
docker pull redoxos/redoxer >/dev/null
kmount=""; [ -n "$kernel" ] && kmount="-v $(dirname $kernel):/kdir"
kvmdev=""; [ -e /dev/kvm ] && kvmdev="--device /dev/kvm"
smp=${SMP:-4}
echo "vm_run: kvm=${kvmdev:-none} smp=$smp kernel=${kernel:-stock} host-cpus=$(nproc)" | tee $log
( timeout ${VM_TIMEOUT:-240} docker run --name ladder --rm $kvmdev -e REDOXER_QEMU_ARGS="-smp $smp" \
    -v ${PKGOUT:-/tmp/pkg/out}:/mnt $kmount redoxos/redoxer sh -c '
  D=/root/.redoxer/x86_64-unknown-redox
  redoxer exec 2>&1 | grep "/dev/kvm"
  if [ -n "'"$kernel"'" ]; then
    redoxer exec -- true >/tmp/prep.log 2>&1 &
    p=$!
    i=0
    while [ $i -lt 240 ]; do
      if ls $D/*.tar >/dev/null 2>&1 && ! ls $D/*.partial >/dev/null 2>&1; then break; fi
      sleep 2; i=$((i+2))
    done
    kill $p 2>/dev/null; pkill qemu 2>/dev/null; sleep 2
    echo "== base image after $i s:"; ls -la $D
    B=$(ls $D/*.tar | head -1)
    mkdir /tmp/b && tar -xpf $B -C /tmp/b
    K=$(find /tmp/b -path "*boot/kernel" -type f | head -1)
    echo "== kernel in image: $K"; ls -la $K; sha1sum $K
    cp /kdir/'"$(basename "$kernel")"' $K
    echo "== swapped in:"; ls -la $K; sha1sum $K
    tar -cpf $B.new -C /tmp/b . && mv $B.new $B
  fi
  redoxer exec -f /mnt -- sh /root/mnt/'"$script"'
' >> $log 2>&1 ) &
dpid=$!
last=0; idle=0
while kill -0 $dpid 2>/dev/null; do
  sleep 3
  size=$(stat -c %s $log)
  if [ "$size" = "$last" ]; then idle=$((idle+3)); else idle=0; last=$size; fi
  grep -aq "LADDER DONE" $log && break
  if [ $idle -ge ${VM_IDLE:-45} ]; then
    echo "VM FROZE: no output for ${idle}s; last line: $(tail -c 200 $log | tr -d '\r' | tail -1)" | tee -a $log
    docker kill ladder 2>/dev/null
    break
  fi
done
docker kill ladder 2>/dev/null; wait $dpid 2>/dev/null
tail -5 $log
