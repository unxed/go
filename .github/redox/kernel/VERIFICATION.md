# Verification of the kernel patch

Workflow: `.github/workflows/redox-kernel-verify.yml` (matrix, one QEMU boot each, redoxer):

| variant | kernel | `exit(0)` hangs (60 runs) | `_exit(0)` hangs (30 runs) |
|---|---|---|---|
| stock | the kernel shipped in `redoxos/redoxer` (packages of 2026-09-18) | 19 | 1 |
| baseline | Redox kernel master `2d2eef7`, built in CI, unpatched | 16 | 0 |
| patched | same source + `0001-futex-don-t-sleep-on-an-untimed-wait-when-the-contex.patch` | **0** | **0** |

How a custom kernel is booted under redoxer: redoxer builds its base image as a tar of the
installed packages (`~/.redoxer/x86_64-unknown-redox/*.tar`) and turns it into a disk on every
`redoxer exec`. `vm_run.sh` runs `redoxer exec -- true` once so that the tar exists, unpacks it,
replaces `usr/lib/boot/kernel` with the kernel built by `make ARCH=x86_64 OBJCOPY=objcopy`
(nightly-2026-05-24, `nasm` required), re-packs it, and then runs the real command.
The log of every run prints the sha1 of the kernel that was swapped in.

Earlier evidence for the diagnosis (Go): in a hung `p07_openat_diag` run (async preemption off, so
no signals involved) `/scheme/sys/context` listed exactly one `UB` thread of the process and
procmgr `UB`.

## v2 -> v3

v2 (`Context::block()` refusing sigkilled contexts) verified the same hang counts but panicked the kernel
in the concurrent-spawn stress `x27_spawnpar_c` (run 35409694501: patched 2/2 panics at run 18, stock and
master 0/4). v3 is futex-only; see the table below once run for v3 completes.

### v3 results (run 35410254114, KVM, 4 CPUs, x27 x40 first, then x17)

| kernel | rep | x27 (40 x 4 spawns) | x17 `exit()` hangs /60 | x17 `_exit` hangs /30 |
|---|---|---|---|---|
| stock | 1 | ok | 3 | 0 |
| stock | 2 | froze after an UNHANDLED EXCEPTION of a spawned child (run 13) | - | - |
| master, unpatched | 1 | ok | 22 | 8 |
| master, unpatched | 2 | ok | 6 | 4 |
| patched v3 | 1 | KERNEL PANIC process.rs:83 (run 21) | - | - |
| patched v3 | 2 | ok | **0** | **0** |
