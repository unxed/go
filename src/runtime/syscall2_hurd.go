// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import _ "unsafe" // for go:linkname

// libc entry points used by package syscall (via syscall_hurd.go). Those needed
// by the runtime itself (close, fcntl, getpid, write, ...) are in os2_hurd.go.

//go:cgo_import_dynamic libc_chdir chdir "libc.so.0.3"
//go:cgo_import_dynamic libc_chroot chroot "libc.so.0.3"
//go:cgo_import_dynamic libc_dup2 dup2 "libc.so.0.3"
//go:cgo_import_dynamic libc_execve execve "libc.so.0.3"
//go:cgo_import_dynamic libc_fork fork "libc.so.0.3"
//go:cgo_import_dynamic libc_gethostname gethostname "libc.so.0.3"
//go:cgo_import_dynamic libc_ioctl ioctl "libc.so.0.3"
//go:cgo_import_dynamic libc_setgid setgid "libc.so.0.3"
//go:cgo_import_dynamic libc_setgroups setgroups "libc.so.0.3"
//go:cgo_import_dynamic libc_setrlimit setrlimit "libc.so.0.3"
//go:cgo_import_dynamic libc_setsid setsid "libc.so.0.3"
//go:cgo_import_dynamic libc_setuid setuid "libc.so.0.3"
//go:cgo_import_dynamic libc_setpgid setpgid "libc.so.0.3"
//go:cgo_import_dynamic libc_wait4 wait4 "libc.so.0.3"

//go:linkname libc_chdir libc_chdir
//go:linkname libc_chroot libc_chroot
//go:linkname libc_dup2 libc_dup2
//go:linkname libc_execve libc_execve
//go:linkname libc_fork libc_fork
//go:linkname libc_gethostname libc_gethostname
//go:linkname libc_ioctl libc_ioctl
//go:linkname libc_setgid libc_setgid
//go:linkname libc_setgroups libc_setgroups
//go:linkname libc_setrlimit libc_setrlimit
//go:linkname libc_setsid libc_setsid
//go:linkname libc_setuid libc_setuid
//go:linkname libc_setpgid libc_setpgid
//go:linkname libc_wait4 libc_wait4
