# GOOS=redox port: status notes

Everything here is exercised only in GitHub Actions (`redox-cross-build`,
`redox-defs-crosscheck`); nothing is built locally.

## Pipeline
* `gen_syscall.py` + `mksyscall_redox.pl`: `syscall/z*_redox_amd64.go` are generated
  from relibc's real headers with redoxer's compiler (cgo -godefs), never by hand.
  The committed copies are re-generated on every CI run (see the drift line in the log).
* `defs_crosscheck.py`: `cgo -godefs` vs. the hand-derived runtime definitions.
* `ladder/`: programs of increasing difficulty, each built separately and run in one
  Redox VM (`ladder_summary.py` renders the table). `x*` programs are diagnostics.

## Known issues / decisions
* **relibc bug: `__relibc_internal_sigentry` clobbers RCX** (`mov ecx, eax` before
  `push rcx`), so every asynchronously delivered signal corrupts the interrupted
  thread's RCX. Go therefore defaults `asyncpreemptoff=1` on redox
  (`runtime/runtime1.go`); `GODEBUG=asyncpreemptoff=0` re-enables it (the `preempt`
  variant in the ladder shows the crashes). Fix upstream: use an already-saved scratch
  register (r8/r10/r12) instead of rcx there.
* `syscall.ReadDirent`/`Getdirentries` return ENOSYS: relibc's `readdir_r` is
  `unimplemented!()` and there is no getdents wrapper; `os` uses fdopendir/readdir.
* Redox's `O_RDONLY` is 0x10000 (not 0): std code that relied on an implicit access
  mode (os.Root opens) needed an explicit `O_RDONLY`.
* `syscall.Syscall*` (numbered raw syscalls) return ENOSYS; `wait4` is built on `waitpid`.
* relibc logs `TODO: getrlimit(7, ...): not implemented` to stderr at every start.
* Occasionally the whole QEMU VM stalls under GC-heavy programs (TCG, 4 vCPUs); the
  ladder runs those rungs last.
* Processes hang / a zero-context thread faults at *exit* in the signal stress test (x10).
