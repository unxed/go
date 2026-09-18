# GOOS=hurd port — status log (append-only)

Branch: `golang-1.26-hurd` (based on `golang-1.26-haiku` @ go1.26.6).
Test sandbox: `unxed/debian-hurd` (QEMU + Debian GNU/Hurd amd64 image, driven from GitHub Actions).
All builds/tests run in CI (`.github/workflows/hurd-cross-build.yml` here, `run-hurd-poc.yml` there).

## 2026-09-18 — Phase 1 done: `println("hello from hurd")` runs on real GNU Hurd

A trivial program with no stdlib imports (`func main() { println("hello from hurd") }`),
cross-compiled on Linux with `GOOS=hurd GOARCH=amd64 CGO_ENABLED=1`, runs on real
Debian GNU/Hurd (QEMU) and exits 0. Runtime initialization, `pthread_create`-based OS thread
creation (sysmon/GC workers), signal setup and `main()` all work.

### Design

Hurd has no flat raw-syscall ABI: nearly everything is a glibc function that does Mach RPC.
So the port follows the libc-calling model (`asmsysvicall6`/`asmcgocall` + `cgo_import_dynamic`),
same family as Solaris/illumos/Haiku/AIX, not Linux's freestanding syscalls.
Files: `src/runtime/{defs,os,os2,signal,netpoll,security}_hurd*.go`, `rt0_hurd_amd64.s`,
`sys_hurd_amd64.s`; `HeadType` `Hhurd` in `cmd/internal/objabi` + `cmd/link` + `cmd/internal/obj/x86`;
dynamic linker path `/lib/ld-x86-64.so.1`.

Behaves like Solaris/Haiku for the cgo `libc_xxx` symbol scheme, but like Linux/BSD/Solaris
(standard ELF, negative-offset `%fs` TLS) for TLS — NOT like Haiku's non-standard TLS slot.

### Facts measured on real Hurd (never guessed — see unxed/debian-hurd `poc/abi_probe.c`, `RESULTS.md`)

- errno values are Mach error codes (`EINTR=0x40000004`, `EAGAIN=0x40000023`, …), not small ints.
- Signal numbers above 15 differ from Linux (`SIGUSR1=30`, `SIGCHLD=20`, `SIGURG=16`, …).
- `mcontext_t`/`ucontext_t`: glibc's generic `gregs[23]` layout (REG_R8=0 … REG_CR2=22); no fs/gs slots.
- `sigset_t` is one 64-bit word; `pthread_attr_t` is 48 bytes (not pointer-sized); `sem_t` 20 bytes.
- `PROT_READ=4, PROT_WRITE=2, PROT_EXEC=1`, `MAP_PRIVATE=0`, `MAP_ANON=2`, `MAP_FIXED=0x100`.
- `pthread_create`, `pthread_kill`, `pthread_self`, `sem_*` live in **`libpthread.so.0.3`**;
  `pthread_attr_*` and everything else used are in `libc.so.0.3`.
- `issetugid()` is not exported by glibc; `secureMode` uses the AIX-style uid/gid comparison.
- Low-level primitives verified before writing the port: `gsync_wait/gsync_wake` (futex-like),
  signal delivery, `sigaltstack`+`SA_SIGINFO`+exact `si_addr` on SIGSEGV.

### Bugs found bringing it up (each root-caused with real runs, not guesses)

1. Missing `hurd` in `asm_amd64.s` `rt0_go` skip-list for the software `m0.tls` self-check
   (Solaris/illumos/Darwin/Haiku/OpenBSD skip it). Symptom: SIGTRAP/`runtime.abort` before `osinit`.
2. `libc_xxx` dynamic imports are `R_X86_64_GLOB_DAT` slots: the resolved address is *stored in*
   the slot; code must load-then-call (`MOVQ (AX), AX`), not `CALL` the slot address.
   Fixed in `asmsysvicall6`, `miniterrno`, `usleep2`, `osyield1`. Symptom: SIGSEGV with PC inside `.got`.
3. `_rt0_amd64_hurd` is the real ELF entry (no glibc `_start`), so argc/argv are on the stack
   (SysV exec ABI): must `JMP _rt0_amd64` (like Linux), not `JMP rt0_go` (Haiku). Symptom:
   SIGSEGV in `getGodebugEarly` walking a garbage argv.
4. pthread symbols split into `libpthread.so.0.3` (see above). Symptom: `undefined symbol: pthread_create`.

### Not done yet (later phases)

- `syscall`, `internal/syscall/unix`, `os`, `net` (only `runtime` is ported; programs must avoid stdlib imports).
- `stat`/`fstat` layout (deferred; not probed in detail yet).
- Native `gsync`-based `lock_futex_hurd.go` (currently POSIX semaphores via `lock_sema.go`).
- `go test` on the runtime; `crash_test.go`/`export_pipe2_test.go`/`semasleep_test.go` build tags.
- Upstreaming: not planned yet.

## 2026-09-18 — poll/sigmask constants measured (abi_probe.c, run-hurd-poc #35372599009)

`_POLLIN/_POLLOUT/_POLLERR/_POLLHUP`, `_SS_DISABLE`, `_SIG_UNBLOCK`, `_SIG_SETMASK`, `_NSIG`
had been copied from Haiku and never measured. On real Hurd:

- `POLLIN=1 POLLOUT=4 POLLERR=8 POLLHUP=16` — same as Linux, NOT Haiku. `netpoll_hurd.go` had
  `POLLOUT=2 POLLERR=4 POLLHUP=0x80` (all three wrong; would have broken the netpoller); fixed.
- `SS_DISABLE=4 SIG_UNBLOCK=2 SIG_SETMASK=3 _NSIG=33` — already correct in `os_hurd.go`.
