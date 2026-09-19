. /root/mnt/common.sh
# Concurrent posix_spawn stress FIRST (a kernel panic there ends the boot): does the patch matter?
loop_run /root/mnt/x27_spawnpar_c.bin default 40 60
# A/B probe for the exit hang, pure C (x17): exit(0) hung ~40% on the stock kernel.
loop_run /root/mnt/x17_exit_c.bin exitarg 60 5
loop_run /root/mnt/x17_exit_c.bin default 30 5
echo "=== LADDER DONE"
