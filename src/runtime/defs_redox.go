// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore

/*
Input to cgo -cdefs.

Unlike GOOS=hurd (see defs_hurd.go), Redox has a real, working C cross
compiler (redoxer's x86_64-unknown-redox-gcc, wrapping relibc), so this
file is not just documentation -- the CI workflow at
.github/workflows/redox-cross-build.yml can actually run it. In practice
defs_redox_amd64.go below was hand-derived directly from relibc's Rust
source (github.com/redox-os/relibc), which is authoritative and easier
to audit than reverse-engineering cgo -cdefs' semi-obscure .h output;
this file documents the intended cgo -cdefs source for anyone who wants
to double check or regenerate it against a real toolchain.

GOARCH=amd64 go tool cgo -cdefs defs_redox.go >defs_redox_amd64.h
*/

package runtime

/*
#include <sys/types.h>
#include <sys/time.h>
#include <signal.h>
#include <errno.h>
#include <sys/mman.h>
#include <semaphore.h>
#include <pthread.h>
#include <sys/resource.h>
#include <sys/stat.h>
#include <poll.h>
#include <time.h>
#include <fcntl.h>
#include <unistd.h>
*/
import "C"

const (
	EINTR     = C.EINTR
	EFAULT    = C.EFAULT
	EAGAIN    = C.EAGAIN
	ENOMEM    = C.ENOMEM
	ETIMEDOUT = C.ETIMEDOUT
	EACCES    = C.EACCES

	PROT_NONE  = C.PROT_NONE
	PROT_READ  = C.PROT_READ
	PROT_WRITE = C.PROT_WRITE
	PROT_EXEC  = C.PROT_EXEC

	MAP_ANON    = C.MAP_ANON
	MAP_PRIVATE = C.MAP_PRIVATE
	MAP_FIXED   = C.MAP_FIXED

	SA_SIGINFO = C.SA_SIGINFO
	SA_RESTART = C.SA_RESTART
	SA_ONSTACK = C.SA_ONSTACK

	SIGHUP    = C.SIGHUP
	SIGINT    = C.SIGINT
	SIGQUIT   = C.SIGQUIT
	SIGILL    = C.SIGILL
	SIGTRAP   = C.SIGTRAP
	SIGABRT   = C.SIGABRT
	SIGFPE    = C.SIGFPE
	SIGKILL   = C.SIGKILL
	SIGBUS    = C.SIGBUS
	SIGSEGV   = C.SIGSEGV
	SIGSYS    = C.SIGSYS
	SIGPIPE   = C.SIGPIPE
	SIGALRM   = C.SIGALRM
	SIGTERM   = C.SIGTERM
	SIGURG    = C.SIGURG
	SIGSTOP   = C.SIGSTOP
	SIGTSTP   = C.SIGTSTP
	SIGCONT   = C.SIGCONT
	SIGCHLD   = C.SIGCHLD
	SIGTTIN   = C.SIGTTIN
	SIGTTOU   = C.SIGTTOU
	SIGXCPU   = C.SIGXCPU
	SIGXFSZ   = C.SIGXFSZ
	SIGVTALRM = C.SIGVTALRM
	SIGPROF   = C.SIGPROF
	SIGWINCH  = C.SIGWINCH
	SIGUSR1   = C.SIGUSR1
	SIGUSR2   = C.SIGUSR2

	FPE_INTDIV = C.FPE_INTDIV
	FPE_INTOVF = C.FPE_INTOVF
	FPE_FLTDIV = C.FPE_FLTDIV
	FPE_FLTOVF = C.FPE_FLTOVF
	FPE_FLTUND = C.FPE_FLTUND
	FPE_FLTRES = C.FPE_FLTRES
	FPE_FLTINV = C.FPE_FLTINV
	FPE_FLTSUB = C.FPE_FLTSUB

	BUS_ADRALN = C.BUS_ADRALN
	BUS_ADRERR = C.BUS_ADRERR
	BUS_OBJERR = C.BUS_OBJERR

	SEGV_MAPERR = C.SEGV_MAPERR
	SEGV_ACCERR = C.SEGV_ACCERR

	ITIMER_REAL    = C.ITIMER_REAL
	ITIMER_VIRTUAL = C.ITIMER_VIRTUAL
	ITIMER_PROF    = C.ITIMER_PROF

	PTHREAD_CREATE_DETACHED = C.PTHREAD_CREATE_DETACHED

	O_WRONLY   = C.O_WRONLY
	O_NONBLOCK = C.O_NONBLOCK
	O_CREAT    = C.O_CREAT
	O_TRUNC    = C.O_TRUNC
	O_CLOEXEC  = C.O_CLOEXEC
	FD_CLOEXEC = C.FD_CLOEXEC
	F_GETFL    = C.F_GETFL
	F_SETFL    = C.F_SETFL
	F_SETFD    = C.F_SETFD

	_SC_NPROCESSORS_ONLN = C._SC_NPROCESSORS_ONLN
	_SC_PAGESIZE         = C._SC_PAGESIZE

	// Constants that live outside defs_redox_amd64.go (os_redox.go,
	// os2_redox.go, netpoll_redox.go); listed here so the CI cross-check
	// (.github/redox/defs_crosscheck.py) can verify them against the headers
	// too -- several of those were initially copied from Hurd/Haiku and wrong.
	O_RDONLY        = C.O_RDONLY
	SS_DISABLE      = C.SS_DISABLE
	SIG_UNBLOCK     = C.SIG_UNBLOCK
	SIG_SETMASK     = C.SIG_SETMASK
	NSIG            = C.NSIG
	CLOCK_REALTIME  = C.CLOCK_REALTIME
	CLOCK_MONOTONIC = C.CLOCK_MONOTONIC
	POLLIN          = C.POLLIN
	POLLOUT         = C.POLLOUT
	POLLHUP         = C.POLLHUP
	POLLERR         = C.POLLERR
	RLIMIT_AS       = C.RLIMIT_AS
)

type SemT C.sem_t

type Sigset C.sigset_t
type StackT C.stack_t

type Siginfo C.siginfo_t

type SigactionT C.struct_sigaction

type Mcontext C.mcontext_t
type Ucontext C.ucontext_t

type Timespec C.struct_timespec
type Timeval C.struct_timeval
type Itimerval C.struct_itimerval

type Pthread C.pthread_t
type PthreadAttr C.pthread_attr_t
