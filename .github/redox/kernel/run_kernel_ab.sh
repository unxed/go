. /root/mnt/common.sh
# A/B probe for the exit hang, pure C (x17): exit(0) hung ~40% on the stock kernel.
loop_run /root/mnt/x17_exit_c.bin exitarg 60 5
loop_run /root/mnt/x17_exit_c.bin default 30 5
echo "=== LADDER DONE"
