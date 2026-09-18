// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build aix || haiku || solaris

package syscall

// Only redox starts children with posix_spawn (see exec_redox_spawn.go).
func spawnInChild(argv0 *byte, argv, envv []*byte, chroot, dir *byte, attr *ProcAttr, sys *SysProcAttr) (pid int, err Errno, ok bool) {
	return 0, 0, false
}
