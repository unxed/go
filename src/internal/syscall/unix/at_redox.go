// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unix

import "syscall"

// Implemented as sysvicall6 in runtime/syscall_redox.go.
func syscall6(trap, nargs, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err syscall.Errno)

//go:cgo_import_dynamic libc_faccessat faccessat "libc.so.6"
//go:cgo_import_dynamic libc_fchmodat fchmodat "libc.so.6"
//go:cgo_import_dynamic libc_fchownat fchownat "libc.so.6"
//go:cgo_import_dynamic libc_fstatat fstatat "libc.so.6"
//go:cgo_import_dynamic libc_linkat linkat "libc.so.6"
//go:cgo_import_dynamic libc_openat openat "libc.so.6"
//go:cgo_import_dynamic libc_renameat renameat "libc.so.6"
//go:cgo_import_dynamic libc_symlinkat symlinkat "libc.so.6"
//go:cgo_import_dynamic libc_unlinkat unlinkat "libc.so.6"
//go:cgo_import_dynamic libc_readlinkat readlinkat "libc.so.6"
//go:cgo_import_dynamic libc_mkdirat mkdirat "libc.so.6"

// Values from relibc's src/header/fcntl/redox.rs and sys_stat/mod.rs; these
// differ from Linux's (e.g. AT_SYMLINK_NOFOLLOW is 0x200, not 0x100).
const (
	AT_EACCESS          = 0x400
	AT_FDCWD            = -0x64
	AT_REMOVEDIR        = 0x200
	AT_SYMLINK_NOFOLLOW = 0x200

	UTIME_OMIT = 0x3ffffffe
)
