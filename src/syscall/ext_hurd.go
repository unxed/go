// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syscall

import "unsafe"

// Additional libc entry points for code outside the standard library that has to
// talk to libc on Hurd (the golang.org/x/sys/unix stand-in used to build f4).
// Dynamic imports only link reliably from this package, so they live here and are
// reached through the single pushed linkname extCall.

//go:cgo_import_dynamic libc_Tcgetattr tcgetattr "libc.so.0.3"
//go:cgo_import_dynamic libc_Tcsetattr tcsetattr "libc.so.0.3"
//go:cgo_import_dynamic libc_Ioctl ioctl "libc.so.0.3"
//go:cgo_import_dynamic libc_Poll poll "libc.so.0.3"
//go:cgo_import_dynamic libc_Mprotect mprotect "libc.so.0.3"
//go:cgo_import_dynamic libc_Fchmodat fchmodat "libc.so.0.3"
//go:cgo_import_dynamic libc_Flock flock "libc.so.0.3"
//go:cgo_import_dynamic libc_Getpgid getpgid "libc.so.0.3"
//go:cgo_import_dynamic libc_PosixOpenpt posix_openpt "libc.so.0.3"
//go:cgo_import_dynamic libc_Grantpt grantpt "libc.so.0.3"
//go:cgo_import_dynamic libc_Unlockpt unlockpt "libc.so.0.3"
//go:cgo_import_dynamic libc_PtsnameR ptsname_r "libc.so.0.3"

//go:linkname libc_Tcgetattr libc_Tcgetattr
//go:linkname libc_Tcsetattr libc_Tcsetattr
//go:linkname libc_Ioctl libc_Ioctl
//go:linkname libc_Poll libc_Poll
//go:linkname libc_Mprotect libc_Mprotect
//go:linkname libc_Fchmodat libc_Fchmodat
//go:linkname libc_Flock libc_Flock
//go:linkname libc_Getpgid libc_Getpgid
//go:linkname libc_PosixOpenpt libc_PosixOpenpt
//go:linkname libc_Grantpt libc_Grantpt
//go:linkname libc_Unlockpt libc_Unlockpt
//go:linkname libc_PtsnameR libc_PtsnameR

var (
	libc_Tcgetattr,
	libc_Tcsetattr,
	libc_Ioctl,
	libc_Poll,
	libc_Mprotect,
	libc_Fchmodat,
	libc_Flock,
	libc_Getpgid,
	libc_PosixOpenpt,
	libc_Grantpt,
	libc_Unlockpt,
	libc_PtsnameR libcFunc
)

// Function indexes for extCall.
const (
	extTcgetattr = iota
	extTcsetattr
	extIoctl
	extPoll
	extMprotect
	extFcntl
	extFchmodat
	extUtimensat
	extFlock
	extMmap
	extGetpgid
	extPosixOpenpt
	extGrantpt
	extUnlockpt
	extPtsnameR
)

var extFns = [...]*libcFunc{
	extTcgetattr:   &libc_Tcgetattr,
	extTcsetattr:   &libc_Tcsetattr,
	extIoctl:       &libc_Ioctl,
	extPoll:        &libc_Poll,
	extMprotect:    &libc_Mprotect,
	extFcntl:       &libc_Fcntl,
	extFchmodat:    &libc_Fchmodat,
	extUtimensat:   &libc_Utimensat,
	extFlock:       &libc_Flock,
	extMmap:        &libc_Mmap,
	extGetpgid:     &libc_Getpgid,
	extPosixOpenpt: &libc_PosixOpenpt,
	extGrantpt:     &libc_Grantpt,
	extUnlockpt:    &libc_Unlockpt,
	extPtsnameR:    &libc_PtsnameR,
}

// extCall calls the libc function selected by idx (an ext* constant) and
// returns its raw results and errno. Pushed to golang.org/x/sys/unix.
//
//go:linkname extCall
func extCall(idx, nargs, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err Errno) {
	return sysvicall6(uintptr(unsafe.Pointer(extFns[idx])), nargs, a1, a2, a3, a4, a5, a6)
}
