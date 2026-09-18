#!/bin/sh
# GitHub-hosted ubuntu runners have /dev/kvm but only root-accessible; open it up.
echo 'KERNEL=="kvm", GROUP="kvm", MODE="0666", OPTIONS+="static_node=kvm"' | sudo tee /etc/udev/rules.d/99-kvm4all.rules >/dev/null
sudo udevadm control --reload-rules
sudo udevadm trigger --name-match=kvm
ls -l /dev/kvm 2>&1 || echo "no /dev/kvm on this runner (TCG fallback)"
nproc
