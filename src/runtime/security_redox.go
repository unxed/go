// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

// relibc does not export issetugid() -- it's a BSD-only libc symbol, and
// importing it here would fail to resolve at dynamic-link time. Use the
// same uid/gid comparison AIX and Hurd use instead.

// secureMode is only ever mutated in schedinit, so we don't need to worry about
// synchronization primitives.
var secureMode bool

func initSecureMode() {
	secureMode = !(getuid() == geteuid() && getgid() == getegid())
}

func isSecureMode() bool {
	return secureMode
}
