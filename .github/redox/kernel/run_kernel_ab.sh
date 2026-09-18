. /root/mnt/common.sh
# A/B probe for the exit hang, pure C (x17): exit(0) hung ~40% on the stock kernel.
loop_run /root/mnt/x17_exit_c.bin exitarg 25 6
loop_run /root/mnt/x17_exit_c.bin default 15 6
echo "=== LADDER DONE"
