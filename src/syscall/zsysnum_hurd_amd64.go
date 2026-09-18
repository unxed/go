// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syscall

// Hurd has no system call numbers (everything is a glibc function doing Mach
// RPC). exec_unix.go names SYS_EXECVE in a branch that is never taken on libc
// systems, so it only needs to compile.
const (
	SYS_EXECVE = 0
	SYS_FCNTL  = 0
)
