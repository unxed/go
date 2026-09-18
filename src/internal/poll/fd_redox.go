// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package poll

import "syscall"

func init() {
	// os/exec starts children with relibc's posix_spawn on Redox, which is
	// told (see syscall/exec_redox_spawn.go) to close the parent's
	// close-on-exec descriptors *by number* in the child. A descriptor that
	// another goroutine closes in the meantime makes the whole spawn fail
	// with EBADF and leaves a half-built child process behind, holding copies
	// of every pipe. So no close may happen while a child is being prepared:
	// spawn holds ForkLock for writing, closes take it for reading.
	CloseFunc = func(fd int) error {
		syscall.ForkLock.RLock()
		defer syscall.ForkLock.RUnlock()
		return syscall.Close(fd)
	}
}
