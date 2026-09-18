// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

// Indexed by Hurd's real glibc signal numbers (bits/signum-generic.h and
// bits/signum-arch.h, measured on real Debian GNU/Hurd) -- these do NOT
// match Linux's or any BSD's numbering above 15, so this table cannot be
// copied from another port; per-signal flags are taken from the generic
// Linux sigtable (sigtab_linux_generic.go) by signal *name*, not position.
var sigtable = [...]sigTabT{
	/*  0 */ {0, "SIGNONE: no trap"},
	/*  1 */ {_SigNotify + _SigKill, "SIGHUP: terminal line hangup"},
	/*  2 */ {_SigNotify + _SigKill, "SIGINT: interrupt"},
	/*  3 */ {_SigNotify + _SigThrow, "SIGQUIT: quit"},
	/*  4 */ {_SigThrow + _SigUnblock, "SIGILL: illegal instruction"},
	/*  5 */ {_SigThrow + _SigUnblock, "SIGTRAP: trace trap"},
	/*  6 */ {_SigNotify + _SigThrow, "SIGABRT: abort"},
	/*  7 */ {_SigThrow, "SIGEMT: emulator trap"},
	/*  8 */ {_SigPanic + _SigUnblock, "SIGFPE: floating-point exception"},
	/*  9 */ {0, "SIGKILL: kill"},
	/* 10 */ {_SigPanic + _SigUnblock, "SIGBUS: bus error"},
	/* 11 */ {_SigPanic + _SigUnblock, "SIGSEGV: segmentation violation"},
	/* 12 */ {_SigThrow, "SIGSYS: bad system call"},
	/* 13 */ {_SigNotify, "SIGPIPE: write to broken pipe"},
	/* 14 */ {_SigNotify, "SIGALRM: alarm clock"},
	/* 15 */ {_SigNotify + _SigKill, "SIGTERM: termination"},
	/* 16 */ {_SigNotify + _SigIgn, "SIGURG: urgent condition on socket"},
	/* 17 */ {0, "SIGSTOP: stop, unblockable"},
	/* 18 */ {_SigNotify + _SigDefault + _SigIgn, "SIGTSTP: keyboard stop"},
	/* 19 */ {_SigNotify + _SigDefault + _SigIgn, "SIGCONT: continue after stop"},
	/* 20 */ {_SigNotify + _SigUnblock + _SigIgn, "SIGCHLD: child status has changed"},
	/* 21 */ {_SigNotify + _SigDefault + _SigIgn, "SIGTTIN: background read from tty"},
	/* 22 */ {_SigNotify + _SigDefault + _SigIgn, "SIGTTOU: background write to tty"},
	/* 23 */ {_SigNotify, "SIGPOLL/SIGIO: pollable event occurred"},
	/* 24 */ {_SigNotify, "SIGXCPU: cpu limit exceeded"},
	/* 25 */ {_SigNotify, "SIGXFSZ: file size limit exceeded"},
	/* 26 */ {_SigNotify, "SIGVTALRM: virtual alarm clock"},
	/* 27 */ {_SigNotify + _SigUnblock, "SIGPROF: profiling alarm clock"},
	/* 28 */ {_SigNotify + _SigIgn, "SIGWINCH: window size change"},
	/* 29 */ {_SigNotify, "SIGINFO: information request"},
	/* 30 */ {_SigNotify, "SIGUSR1: user-defined signal 1"},
	/* 31 */ {_SigNotify, "SIGUSR2: user-defined signal 2"},
	/* 32 */ {_SigNotify, "SIGLOST: resource lost"},
}
