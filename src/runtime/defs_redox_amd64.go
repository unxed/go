// Hand-derived from relibc (github.com/redox-os/relibc) Rust source, which
// is the authoritative definition of Redox's libc ABI (relibc's structs are
// #[repr(C)], so they map 1:1 onto the C/Go layout below). See defs_redox.go
// for how these could alternatively be regenerated via a real cgo -cdefs run
// against redoxer's cross toolchain, which -- unlike GOOS=hurd -- actually
// exists and works from a Linux CI host.
//
// Sources consulted (all in redox-os/relibc):
//   src/header/errno/mod.rs            -- errno values (same as Linux)
//   src/header/signal/constants.rs     -- signal numbers (same as Linux),
//                                          FPE_*/BUS_*/SEGV_* codes
//   src/header/signal/redox.rs         -- SA_* flags (Redox-specific, NOT
//                                          the same bit values as Linux),
//                                          native ucontext/mcontext structs
//   src/header/signal/mod.rs           -- struct sigaction/sigaltstack/siginfo
//   src/header/bits_sigset-t/mod.rs    -- sigset_t = c_ulonglong (not a
//                                          128-byte mask like Linux/glibc)
//   src/header/sys_mman/redox.rs       -- PROT_*/MAP_* (Redox-specific values)
//   src/header/bits_open-flags/redox.rs, src/header/fcntl/{mod,redox}.rs -- O_*/F_*
//   src/header/sys_time/mod.rs         -- ITIMER_*, itimerval
//   src/header/bits_timespec, bits_timeval/mod.rs -- timespec/timeval (both
//                                          64-bit fields, like Linux amd64)
//   src/header/semaphore/mod.rs        -- sem_t (8-byte union)
//   src/header/bits_pthreadattr-t/mod.rs -- pthread_attr_t (32-byte union)
//   src/header/pthread/mod.rs          -- pthread_t = *mut c_void;
//                                          PTHREAD_CREATE_DETACHED = 0 (NOT 1)
//   src/header/unistd/sysconf/constants.rs -- _SC_NPROCESSORS_ONLN=84, _SC_PAGESIZE=30

package runtime

const (
	_EINTR     = 0x4
	_EFAULT    = 0xe
	_EAGAIN    = 0xb
	_ENOMEM    = 0xc
	_ETIMEDOUT = 0x6e
	_EACCES    = 0xd
	_ENOSYS    = 0x26

	_PROT_NONE  = 0x0
	_PROT_READ  = 0x4
	_PROT_WRITE = 0x2
	_PROT_EXEC  = 0x1

	_MAP_ANON    = 0x20
	_MAP_PRIVATE = 0x2
	_MAP_FIXED   = 0x4

	// Non-POSIX, Redox-specific values (do NOT match Linux's SA_* bits).
	_SA_SIGINFO = 0x02000000
	_SA_RESTART = 0x08000000
	_SA_ONSTACK = 0x04000000

	_SIGHUP    = 0x1
	_SIGINT    = 0x2
	_SIGQUIT   = 0x3
	_SIGILL    = 0x4
	_SIGTRAP   = 0x5
	_SIGABRT   = 0x6
	_SIGBUS    = 0x7
	_SIGFPE    = 0x8
	_SIGKILL   = 0x9
	_SIGUSR1   = 0xa
	_SIGSEGV   = 0xb
	_SIGUSR2   = 0xc
	_SIGPIPE   = 0xd
	_SIGALRM   = 0xe
	_SIGTERM   = 0xf
	_SIGCHLD   = 0x11
	_SIGCONT   = 0x12
	_SIGSTOP   = 0x13
	_SIGTSTP   = 0x14
	_SIGTTIN   = 0x15
	_SIGTTOU   = 0x16
	_SIGURG    = 0x17
	_SIGXCPU   = 0x18
	_SIGXFSZ   = 0x19
	_SIGVTALRM = 0x1a
	_SIGPROF   = 0x1b
	_SIGWINCH  = 0x1c
	_SIGSYS    = 0x1f

	_FPE_INTDIV = 0x1
	_FPE_INTOVF = 0x2
	_FPE_FLTDIV = 0x3
	_FPE_FLTOVF = 0x4
	_FPE_FLTUND = 0x5
	_FPE_FLTRES = 0x6
	_FPE_FLTINV = 0x7
	_FPE_FLTSUB = 0x8

	_BUS_ADRALN = 0x1
	_BUS_ADRERR = 0x2
	_BUS_OBJERR = 0x3

	_SEGV_MAPERR = 0x1
	_SEGV_ACCERR = 0x2

	_ITIMER_REAL    = 0x0
	_ITIMER_VIRTUAL = 0x1
	_ITIMER_PROF    = 0x2

	// Redox's PTHREAD_CREATE_DETACHED is 0 (PTHREAD_CREATE_JOINABLE is 1) --
	// the reverse of Linux/glibc and Hurd. Confirmed from
	// src/header/pthread/mod.rs; do not "fix" this to look like other ports.
	_PTHREAD_CREATE_DETACHED = 0x0

	_HOST_NAME_MAX  = 0x100 // not probed; conservative POSIX-typical fallback
	_MAXHOSTNAMELEN = 0x100

	_O_RDONLY   = 0x00010000
	_O_NONBLOCK = 0x00040000
	_FD_CLOEXEC = 0x01000000
	_F_GETFL    = 0x3
	_F_SETFL    = 0x4
	_F_SETFD    = 0x2

	__SC_PAGESIZE         = 0x1e // _SC_PAGESIZE = 30
	__SC_NPROCESSORS_ONLN = 0x54 // _SC_NPROCESSORS_ONLN = 84

	_O_CREAT  = 0x02000000
	_O_WRONLY = 0x00020000
	_O_TRUNC  = 0x04000000
)

// sem_t is an 8-byte opaque union (an 8-byte c_long dominates a 4-byte
// c_char[4] member for size/alignment purposes); represented here as a
// single 8-byte-aligned field.
type semt struct {
	_align int64
}

// sigset_t on Redox is a single 64-bit word (bits_sigset-t/mod.rs:
// pub type sigset_t = c_ulonglong), unlike Linux's 128-byte mask.
type sigset uint64

// stack_t / sigaltstack: real declared member order on Redox is
// sp, flags, size (signal/mod.rs: pub struct sigaltstack), matching neither
// the traditional POSIX order nor Hurd's sp/size/flags.
type stackt struct {
	ss_sp     *byte
	ss_flags  int32
	pad_cgo_0 [4]byte
	ss_size   uintptr
}

type sigaltstackt stackt

// siginfo (signal/mod.rs: pub struct siginfo) has no si_band field, unlike
// Hurd/glibc's siginfo_t -- Redox doesn't model realtime signal queueing
// metadata at this layer.
type siginfo struct {
	si_signo  int32
	si_errno  int32
	si_code   int32
	si_pid    int32
	si_uid    uint32
	pad_cgo_0 [4]byte
	si_addr   uintptr
	si_status int32
	pad_cgo_1 [4]byte
	si_value  uintptr // sigval union; only the pointer member matters here
}

// struct sigaction (signal/mod.rs): sa_handler, sa_flags, sa_restorer,
// sa_mask -- note the sa_restorer field (glibc/Linux-style signal
// trampoline slot) sitting between sa_flags and sa_mask; we never set
// SA_RESTORER so it is always left nil.
type sigactiont struct {
	_funcptr    [8]byte
	sa_flags    int32
	pad_cgo_0   [4]byte
	sa_restorer uintptr
	sa_mask     sigset
}

// mcontext (signal/redox.rs, target_arch = x86_64) is a named-field struct
// -- like Haiku's, unlike Linux/Solaris/Hurd's indexed gregset_t -- fronted
// by the FPU/vector save area. No cs/fs/gs slots exist at all, so, as on
// Haiku, fs()/gs()/cs() are stubbed in signal_redox_amd64.go.
type mcontext struct {
	ymmUpper [16][2]uint64
	fxsave   [29][2]uint64
	r15      uint64
	r14      uint64
	r13      uint64
	r12      uint64
	rbp      uint64
	rbx      uint64
	r11      uint64
	r10      uint64
	r9       uint64
	r8       uint64
	rax      uint64
	rcx      uint64
	rdx      uint64
	rsi      uint64
	rdi      uint64
	rflags   uint64
	rip      uint64
	rsp      uint64
}

// ucontext (signal/redox.rs: pub struct ucontext). The leading pad_cgo_0
// and the _sival/_sigcode/_signum tail are relibc-internal bookkeeping we
// never read.
type ucontext struct {
	pad_cgo_0   uint64
	uc_link     *ucontext
	uc_stack    stackt
	uc_sigmask  sigset
	_sival      uint64
	_sigcode    uint32
	_signum     uint32
	uc_mcontext mcontext
}

type timespec struct {
	tv_sec  int64
	tv_nsec int64
}

func (ts *timespec) set_sec(x int64) {
	ts.tv_sec = x
}

func (ts *timespec) set_nsec(x int32) {
	ts.tv_nsec = int64(x)
}

// suseconds_t is a 4-byte c_int on this target (bits_suseconds-t/mod.rs),
// not the 8-byte c_long I had initially assumed -- caught by the CI cgo
// cross-check (.github/workflows/redox-defs-crosscheck.yml).
type timeval struct {
	tv_sec    int64
	tv_usec   int32
	pad_cgo_0 [4]byte
}

func (tv *timeval) set_usec(x int32) {
	tv.tv_usec = x
}

type itimerval struct {
	it_interval timeval
	it_value    timeval
}

// pthread_t is *mut c_void on Redox (pthread/mod.rs), a plain opaque
// pointer-sized handle, same shape as Hurd/Haiku.
type pthread uintptr

// pthread_attr_t is a 32-byte opaque union (bits_pthreadattr-t/mod.rs:
// union of [c_uchar; 32] and size_t) -- must be allocated inline with
// 8-byte alignment, never as a bare uintptr.
type pthreadattr struct {
	_align int64
	_pad   [24]byte
}
