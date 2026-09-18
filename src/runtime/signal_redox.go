// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

// Indexed by Redox's signal numbers (relibc's src/header/signal/constants.rs,
// which states outright "These constants have the same values on Redox and
// Linux") -- so, unlike Hurd, this table's positions match Linux's generic
// sigtable (sigtab_linux_generic.go) exactly; per-signal flags are still
// taken by signal *name* rather than assumed from position, in case that
// ever changes.
var sigtable = [...]sigTabT{
	/*  0 */ {0, "SIGNONE: no trap"},
	/*  1 */ {_SigNotify + _SigKill, "SIGHUP: terminal line hangup"},
	/*  2 */ {_SigNotify + _SigKill, "SIGINT: interrupt"},
	/*  3 */ {_SigNotify + _SigThrow, "SIGQUIT: quit"},
	/*  4 */ {_SigThrow + _SigUnblock, "SIGILL: illegal instruction"},
	/*  5 */ {_SigThrow + _SigUnblock, "SIGTRAP: trace trap"},
	/*  6 */ {_SigNotify + _SigThrow, "SIGABRT: abort"},
	/*  7 */ {_SigPanic + _SigUnblock, "SIGBUS: bus error"},
	/*  8 */ {_SigPanic + _SigUnblock, "SIGFPE: floating-point exception"},
	/*  9 */ {0, "SIGKILL: kill"},
	/* 10 */ {_SigNotify, "SIGUSR1: user-defined signal 1"},
	/* 11 */ {_SigPanic + _SigUnblock, "SIGSEGV: segmentation violation"},
	/* 12 */ {_SigNotify, "SIGUSR2: user-defined signal 2"},
	/* 13 */ {_SigNotify, "SIGPIPE: write to broken pipe"},
	/* 14 */ {_SigNotify, "SIGALRM: alarm clock"},
	/* 15 */ {_SigNotify + _SigKill, "SIGTERM: termination"},
	/* 16 */ {_SigThrow, "SIGSTKFLT: stack fault (unused)"},
	/* 17 */ {_SigNotify + _SigUnblock + _SigIgn, "SIGCHLD: child status has changed"},
	/* 18 */ {_SigNotify + _SigDefault + _SigIgn, "SIGCONT: continue after stop"},
	/* 19 */ {0, "SIGSTOP: stop, unblockable"},
	/* 20 */ {_SigNotify + _SigDefault + _SigIgn, "SIGTSTP: keyboard stop"},
	/* 21 */ {_SigNotify + _SigDefault + _SigIgn, "SIGTTIN: background read from tty"},
	/* 22 */ {_SigNotify + _SigDefault + _SigIgn, "SIGTTOU: background write to tty"},
	/* 23 */ {_SigNotify + _SigIgn, "SIGURG: urgent condition on socket"},
	/* 24 */ {_SigNotify, "SIGXCPU: cpu limit exceeded"},
	/* 25 */ {_SigNotify, "SIGXFSZ: file size limit exceeded"},
	/* 26 */ {_SigNotify, "SIGVTALRM: virtual alarm clock"},
	/* 27 */ {_SigNotify + _SigUnblock, "SIGPROF: profiling alarm clock"},
	/* 28 */ {_SigNotify + _SigIgn, "SIGWINCH: window size change"},
	/* 29 */ {_SigNotify, "SIGIO: I/O now possible"},
	/* 30 */ {_SigNotify, "SIGPWR: power failure"},
	/* 31 */ {_SigThrow, "SIGSYS: bad system call"},
}
