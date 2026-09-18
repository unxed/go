// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build redox

package syscall

import "unsafe"

// Process creation on Redox goes through relibc's posix_spawn, not fork+exec.
//
// Redox's kernel keys futexes by *physical* address. fork() makes every page
// of the (multi-threaded) parent copy-on-write; when the parent then writes to
// a page that the child still shares, the page moves to a new physical frame
// and a FUTEX_WAKE for the new address no longer finds the threads that went
// to sleep on the old one. For Go that means parked Ms (runtime.semasleep) are
// never woken again and the whole program stalls after the first os/exec.
// relibc's posix_spawn builds the child directly from the executable and does
// not clone the address space, so it does not have this problem (and is much
// cheaper than fork).

//go:cgo_import_dynamic libc_posix_spawn posix_spawn "libc.so.6"
//go:cgo_import_dynamic libc_posix_spawn_file_actions_init posix_spawn_file_actions_init "libc.so.6"
//go:cgo_import_dynamic libc_posix_spawn_file_actions_destroy posix_spawn_file_actions_destroy "libc.so.6"
//go:cgo_import_dynamic libc_posix_spawn_file_actions_adddup2 posix_spawn_file_actions_adddup2 "libc.so.6"
//go:cgo_import_dynamic libc_posix_spawn_file_actions_addclose posix_spawn_file_actions_addclose "libc.so.6"
//go:cgo_import_dynamic libc_posix_spawn_file_actions_addchdir posix_spawn_file_actions_addchdir "libc.so.6"
//go:cgo_import_dynamic libc_posix_spawnattr_init posix_spawnattr_init "libc.so.6"
//go:cgo_import_dynamic libc_posix_spawnattr_destroy posix_spawnattr_destroy "libc.so.6"
//go:cgo_import_dynamic libc_posix_spawnattr_setflags posix_spawnattr_setflags "libc.so.6"
//go:cgo_import_dynamic libc_posix_spawnattr_setpgroup posix_spawnattr_setpgroup "libc.so.6"

//go:linkname libc_posix_spawn libc_posix_spawn
//go:linkname libc_posix_spawn_file_actions_init libc_posix_spawn_file_actions_init
//go:linkname libc_posix_spawn_file_actions_destroy libc_posix_spawn_file_actions_destroy
//go:linkname libc_posix_spawn_file_actions_adddup2 libc_posix_spawn_file_actions_adddup2
//go:linkname libc_posix_spawn_file_actions_addclose libc_posix_spawn_file_actions_addclose
//go:linkname libc_posix_spawn_file_actions_addchdir libc_posix_spawn_file_actions_addchdir
//go:linkname libc_posix_spawnattr_init libc_posix_spawnattr_init
//go:linkname libc_posix_spawnattr_destroy libc_posix_spawnattr_destroy
//go:linkname libc_posix_spawnattr_setflags libc_posix_spawnattr_setflags
//go:linkname libc_posix_spawnattr_setpgroup libc_posix_spawnattr_setpgroup

var (
	libc_posix_spawn                       libcFunc
	libc_posix_spawn_file_actions_init     libcFunc
	libc_posix_spawn_file_actions_destroy  libcFunc
	libc_posix_spawn_file_actions_adddup2  libcFunc
	libc_posix_spawn_file_actions_addclose libcFunc
	libc_posix_spawn_file_actions_addchdir libcFunc
	libc_posix_spawnattr_init              libcFunc
	libc_posix_spawnattr_destroy           libcFunc
	libc_posix_spawnattr_setflags          libcFunc
	libc_posix_spawnattr_setpgroup         libcFunc
)

// relibc's posix_spawn_file_actions_t (a Vec, 24 bytes) and posix_spawnattr_t
// (< 64 bytes); both are initialised by their _init functions.
type (
	spawnFileActions [3]uintptr
	spawnAttr        [8]uintptr
)

const (
	spawnSetPgroup = 2    // relibc's POSIX_SPAWN_SETPGROUP
	spawnScanLimit = 1024 // highest descriptor (exclusive) examined for close-on-exec
)

func spawnCall(fn *libcFunc, nargs, a1, a2, a3, a4, a5, a6 uintptr) Errno {
	r, _, _ := sysvicall6(uintptr(unsafe.Pointer(fn)), nargs, a1, a2, a3, a4, a5, a6)
	return Errno(r)
}

// spawnInChild starts the child with posix_spawn when the request can be
// expressed that way; ok == false means the caller must fall back to
// fork+exec. Unlike fork+exec, the child's argv[0] is always the program path
// (relibc's spawn overwrites it).
func spawnInChild(argv0 *byte, argv, envv []*byte, chroot, dir *byte, attr *ProcAttr, sys *SysProcAttr) (pid int, err Errno, ok bool) {
	if chroot != nil || sys.Credential != nil || sys.Setsid || sys.Foreground || sys.Setctty || sys.Noctty ||
		len(attr.Files) > 3 || len(argv) == 0 || len(envv) == 0 {
		// Fds >= 3 would need care: relibc closes, in the child, every fd number that is
		// close-on-exec in the parent, even one just dup2'ed onto.
		return 0, 0, false
	}

	var fa spawnFileActions
	if e := spawnCall(&libc_posix_spawn_file_actions_init, 1, uintptr(unsafe.Pointer(&fa)), 0, 0, 0, 0, 0); e != 0 {
		return 0, e, true
	}
	defer spawnCall(&libc_posix_spawn_file_actions_destroy, 1, uintptr(unsafe.Pointer(&fa)), 0, 0, 0, 0, 0)

	if dir != nil {
		if e := spawnCall(&libc_posix_spawn_file_actions_addchdir, 2, uintptr(unsafe.Pointer(&fa)), uintptr(unsafe.Pointer(dir)), 0, 0, 0, 0); e != 0 {
			return 0, e, true
		}
	}

	// Same fd shuffling as forkAndExecInChild: first move any source fd that a
	// lower-numbered target would overwrite out of the way, then dup2 into place.
	fds := make([]int, len(attr.Files))
	next := len(attr.Files)
	for i, ufd := range attr.Files {
		if next < int(ufd) {
			next = int(ufd)
		}
		fds[i] = int(ufd)
	}
	next++
	var temps []int
	for i := range fds {
		if fds[i] >= 0 && fds[i] < i {
			if e := spawnCall(&libc_posix_spawn_file_actions_adddup2, 3, uintptr(unsafe.Pointer(&fa)), uintptr(fds[i]), uintptr(next), 0, 0, 0); e != 0 {
				return 0, e, true
			}
			fds[i] = next
			temps = append(temps, next)
			next++
		}
	}
	for i, fd := range fds {
		var e Errno
		switch {
		case fd == -1:
			e = spawnCall(&libc_posix_spawn_file_actions_addclose, 2, uintptr(unsafe.Pointer(&fa)), uintptr(i), 0, 0, 0, 0)
		case fd != i:
			e = spawnCall(&libc_posix_spawn_file_actions_adddup2, 3, uintptr(unsafe.Pointer(&fa)), uintptr(fd), uintptr(i), 0, 0, 0)
		}
		if e != 0 {
			return 0, e, true
		}
	}
	for _, t := range temps {
		if e := spawnCall(&libc_posix_spawn_file_actions_addclose, 2, uintptr(unsafe.Pointer(&fa)), uintptr(t), 0, 0, 0, 0); e != 0 {
			return 0, e, true
		}
	}

	// relibc's posix_spawn is meant to close, in the child, every descriptor
	// that is close-on-exec in the parent, but in practice does not (a `cat`
	// child keeps the write end of its own stdin pipe open and never sees EOF;
	// the exec status pipe stays open too). Do it explicitly.
	for fd := 3; fd < spawnScanLimit; fd++ {
		if v, e := fcntl(fd, F_GETFD, 0); e == nil && v&FD_CLOEXEC != 0 {
			if e := spawnCall(&libc_posix_spawn_file_actions_addclose, 2, uintptr(unsafe.Pointer(&fa)), uintptr(fd), 0, 0, 0, 0); e != 0 {
				return 0, e, true
			}
		}
	}

	var sa spawnAttr
	var saptr uintptr
	if sys.Setpgid {
		if e := spawnCall(&libc_posix_spawnattr_init, 1, uintptr(unsafe.Pointer(&sa)), 0, 0, 0, 0, 0); e != 0 {
			return 0, e, true
		}
		defer spawnCall(&libc_posix_spawnattr_destroy, 1, uintptr(unsafe.Pointer(&sa)), 0, 0, 0, 0, 0)
		spawnCall(&libc_posix_spawnattr_setflags, 2, uintptr(unsafe.Pointer(&sa)), spawnSetPgroup, 0, 0, 0, 0)
		spawnCall(&libc_posix_spawnattr_setpgroup, 2, uintptr(unsafe.Pointer(&sa)), uintptr(sys.Pgid), 0, 0, 0, 0)
		saptr = uintptr(unsafe.Pointer(&sa))
	}

	var cpid _Pid_t
	e := spawnCall(&libc_posix_spawn, 6, uintptr(unsafe.Pointer(&cpid)), uintptr(unsafe.Pointer(argv0)),
		uintptr(unsafe.Pointer(&fa)), saptr, uintptr(unsafe.Pointer(&argv[0])), uintptr(unsafe.Pointer(&envv[0])))
	if e != 0 {
		return 0, e, true
	}
	forkExecSpawned = true // read by forkExec while it still holds ForkLock
	return int(cpid), 0, true
}
