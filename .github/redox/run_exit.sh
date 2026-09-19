. /root/mnt/common.sh
# Exit-hang probes. Poll limit 10s: a run that does not finish is a hang.
# x17: pure C (no Go): pthread/sem ping-pong, main `_exit`s at a random moment
loop_run /root/mnt/x17_exit_c.bin default 25 10
loop_run /root/mnt/x17_exit_c.bin exitarg 10 10
# x16: Go equivalent (channel ping-pong => Ms parking on indefinite futexes)
loop_run /root/mnt/x16_exitstress.bin default 25 10
# x11: the original small probe
loop_run /root/mnt/x11_exitloop.bin default 10 10
# concurrent posix_spawn stress LAST: on the patched kernel it once panicked (process.rs:83)
# after a spawned child faulted at RIP=0; repeated to see whether stock does it too
loop_run /root/mnt/x27_spawnpar_c.bin default 8 60
echo "=== LADDER DONE"
