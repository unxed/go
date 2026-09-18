// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import _ "unsafe" // for go:linkname

// Imports used by package syscall (os2_redox.go already imports close,
// fcntl, getpid, write, ... for the runtime's own use). All of these come
// from relibc's single libc.so.6.

//go:cgo_import_dynamic libc_chdir chdir "libc.so.6"
//go:cgo_import_dynamic libc_chroot chroot "libc.so.6"
//go:cgo_import_dynamic libc_dup2 dup2 "libc.so.6"
//go:cgo_import_dynamic libc_execve execve "libc.so.6"
//go:cgo_import_dynamic libc_fork fork "libc.so.6"
//go:cgo_import_dynamic libc_gethostname gethostname "libc.so.6"
//go:cgo_import_dynamic libc_ioctl ioctl "libc.so.6"
//go:cgo_import_dynamic libc_setgid setgid "libc.so.6"
//go:cgo_import_dynamic libc_setgroups setgroups "libc.so.6"
//go:cgo_import_dynamic libc_setrlimit setrlimit "libc.so.6"
//go:cgo_import_dynamic libc_setsid setsid "libc.so.6"
//go:cgo_import_dynamic libc_setuid setuid "libc.so.6"
//go:cgo_import_dynamic libc_setpgid setpgid "libc.so.6"
//go:cgo_import_dynamic libc_waitpid waitpid "libc.so.6"

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
//go:linkname libc_waitpid libc_waitpid
