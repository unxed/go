# Redox issues found while porting Go (kernel / relibc / procmgr)

Everything below was found with the Go port and reduced to a small program in
`.github/redox/ladder/` (C wherever possible, so it does not need Go). "Verified" means
measured in CI (`redox-cross-build`, `redox-kernel-verify`) on QEMU with KVM and 4 CPUs,
unless stated otherwise.

| # | component | problem | repro | status |
|---|---|---|---|---|
| 1 | kernel | ForceKill of a thread that is about to sleep is lost; the process never finishes exiting | `x17_exit_c` | **patch verified**: `kernel/0001-*.patch`, text in `kernel/ISSUE.md`, results in `kernel/VERIFICATION.md` |
| 2 | kernel | futexes are keyed by physical address; `fork()` makes pages copy-on-write, the parent's next write moves the page, sleepers on the old frame are never woken | `x23_futex_fork_c` | worked around in Go (posix_spawn); kernel fix not written |
| 3 | kernel + procmgr | CPU exceptions (page fault, ...) are never delivered to a user signal handler (`excp_handler` is a TODO); the faulting thread is killed and a multi-threaded process then lingers, because procmgr starts the exit only when *all* threads are dead | `x24_nilpanic` (before the compiler emitted explicit nil checks), `x10_sigstorm` | worked around in the Go compiler (explicit nil checks); kernel/procmgr not fixed |
| 4 | relibc | a process created by `posix_spawn` starts with MXCSR = 0 and x87 CW = 0, so the first inexact SSE result raises #UD ("Invalid opcode fault"); `rlct_clone_ret` sets them for threads, `spawn` does not | `x25_spawn_fpu_c` | worked around in Go (`_rt0_amd64_redox`) |
| 5 | relibc | `posix_spawn` does not close, in the child, the parent's close-on-exec descriptors (a `cat` child keeps the write end of its own stdin pipe open and never sees EOF); a spawn that fails halfway (e.g. EBADF because a listed fd vanished) leaves a half-built child process holding copies of every fd | `x26_spawn_cloexec_c` | worked around in Go (explicit `addclose`, closes serialised with ForkLock) |
| 6 | relibc | a new thread starts with an empty signal mask and applies the creator's mask a little later (pthread/mod.rs); a process-directed signal (SIGCHLD of an exited child) can hit it before TLS/`g` is set up | seen as "fatal: bad g in signal handler" in `x20_execpar` | worked around in Go (drop benign async signals on a g-less thread) |
| 7 | relibc | `sigentry` clobbers RCX of the interrupted code (async signals corrupt user state) | `x10_sigstorm` | fixed upstream: unxed/relibc `fix-sigentry-rcx-clobber` (8022058e) |
| 8 | QEMU TCG (not KVM) | CLOCK_MONOTONIC / CLOCK_REALTIME read on different vCPUs go backwards by up to ~4 ms | `x21_clock_c` | informational; gone with KVM |

## Reproducing with the images redoxer ships

`.github/redox/vm_run.sh <script> <log> [kernel]` boots one VM (`redoxer exec`), optionally
after swapping the kernel inside redoxer's base image (`usr/lib/boot/kernel` in the tar
under `~/.redoxer/x86_64-unknown-redox/`). The kernel is built with
`make ARCH=x86_64 OBJCOPY=objcopy` from the pinned source (nightly-2026-05-24, `nasm`).
