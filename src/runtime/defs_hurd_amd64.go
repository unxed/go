// Hand-written from sizeof/offsetof values measured with a small C probe
// (poc/abi_probe.c in unxed/debian-hurd) compiled and run on real Debian
// GNU/Hurd amd64 (gcc-15, glibc). See defs_hurd.go.
//
// Hurd errno values are Mach error codes (error system 0x10, subsystem 0),
// not small integers like Linux/BSD -- do not "fix" these to look smaller.

package runtime

const (
	_EINTR     = 0x40000004
	_EFAULT    = 0x4000000e
	_EAGAIN    = 0x40000023
	_ENOMEM    = 0x4000000c
	_ETIMEDOUT = 0x4000003c
	_EACCES    = 0x4000000d

	_PROT_NONE  = 0x0
	_PROT_READ  = 0x4
	_PROT_WRITE = 0x2
	_PROT_EXEC  = 0x1

	_MAP_ANON    = 0x2
	_MAP_PRIVATE = 0x0
	_MAP_FIXED   = 0x100

	_SA_SIGINFO = 0x40
	_SA_RESTART = 0x2
	_SA_ONSTACK = 0x1

	_SIGHUP    = 0x1
	_SIGINT    = 0x2
	_SIGQUIT   = 0x3
	_SIGILL    = 0x4
	_SIGTRAP   = 0x5
	_SIGABRT   = 0x6
	_SIGFPE    = 0x8
	_SIGKILL   = 0x9
	_SIGBUS    = 0xa
	_SIGSEGV   = 0xb
	_SIGSYS    = 0xc
	_SIGPIPE   = 0xd
	_SIGALRM   = 0xe
	_SIGTERM   = 0xf
	_SIGURG    = 0x10
	_SIGSTOP   = 0x11
	_SIGTSTP   = 0x12
	_SIGCONT   = 0x13
	_SIGCHLD   = 0x14
	_SIGTTIN   = 0x15
	_SIGTTOU   = 0x16
	_SIGXCPU   = 0x18
	_SIGXFSZ   = 0x19
	_SIGVTALRM = 0x1a
	_SIGPROF   = 0x1b
	_SIGWINCH  = 0x1c
	_SIGUSR1   = 0x1e
	_SIGUSR2   = 0x1f

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

	_PTHREAD_CREATE_DETACHED = 0x1

	_HOST_NAME_MAX = 0x100 // not defined by Hurd's libc; conservative fallback

	_O_NONBLOCK = 0x8
	_FD_CLOEXEC = 0x1
	_F_GETFL    = 0x3
	_F_SETFL    = 0x4
	_F_SETFD    = 0x2

	__SC_PAGESIZE         = 0x1e // _SC_PAGESIZE
	__SC_NPROCESSORS_ONLN = 0x54 // _SC_NPROCESSORS_ONLN

	_MAXHOSTNAMELEN = 0x100

	_O_CREAT  = 0x10
	_O_WRONLY = 0x2
	_O_TRUNC  = 0x10000
)

// sem_t is a 20-byte opaque glibc union (long-aligned); represented here
// with an 8-byte-aligned field to reproduce that alignment.
type semt struct {
	_align int64
	_pad   [16]byte
}

// sigset_t on Hurd is a single 64-bit word (unlike Linux's 128-byte mask).
type sigset uint64

// stack_t / sigaltstack: real member order on Hurd is sp, size, flags
// (not the POSIX declaration order sp/flags/size).
type stackt struct {
	ss_sp     *byte
	ss_size   uintptr
	ss_flags  int32
	pad_cgo_0 [4]byte
}

type sigaltstackt stackt

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
	si_band   int64
	si_value  uintptr
}

type sigactiont struct {
	_funcptr  [8]byte
	sa_mask   sigset
	sa_flags  int32
	pad_cgo_0 [4]byte
}

// gregset_t is greg_t[23] (greg_t is a plain 8-byte integer); Hurd's
// register numbering (REG_R8=0 .. REG_CR2=22) matches glibc's generic
// x86_64 <sys/ucontext.h> exactly, same as Linux/Solaris.
type gregset [23]uint64

// mcontext_t = { gregset_t gregs; fpregset_t fpregs; unsigned long long
// __reserved1[8]; }. fpregs is a pointer we don't dereference for now.
type mcontext struct {
	gregs     gregset
	fpregs    uintptr
	reserved1 [8]uint64
}

// ucontext_t is much larger than what we model here (it also embeds an
// inline struct _libc_fpstate and a 4-word __ssp[4]); we only need
// uc_link/uc_stack/uc_mcontext/uc_sigmask at their real offsets, so the
// FP/shadow-stack tail is left as opaque padding sized to match the real
// 848-byte struct.
type ucontext struct {
	uc_flags    uint64
	uc_link     *ucontext
	uc_stack    stackt
	uc_mcontext mcontext
	uc_sigmask  sigset
	pad_cgo_0   [544]byte
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

type timeval struct {
	tv_sec  int64
	tv_usec int64
}

func (tv *timeval) set_usec(x int32) {
	tv.tv_usec = int64(x)
}

type itimerval struct {
	it_interval timeval
	it_value    timeval
}

// pthread_t is a plain pointer-sized opaque handle on Hurd.
type pthread uintptr

// pthread_attr_t is a 48-byte opaque glibc struct (not pointer-sized like
// on Haiku) -- must be allocated inline with 8-byte alignment, never as a
// bare uintptr, or pthread_attr_init will corrupt adjacent memory.
type pthreadattr struct {
	_align int64
	_pad   [40]byte
}
