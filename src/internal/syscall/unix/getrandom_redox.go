// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unix

import (
	"unsafe"
)

//go:cgo_import_dynamic libc_getrandom getrandom "libc.so.6"
//go:linkname libc_getrandom libc_getrandom

var libc_getrandom uintptr

type GetRandomFlag uintptr

const (
	// GRND_NONBLOCK and GRND_RANDOM from relibc's sys_random/mod.rs.
	GRND_NONBLOCK GetRandomFlag = 0x0001
	GRND_RANDOM   GetRandomFlag = 0x0002
)

// GetRandom calls relibc's getrandom, which reads from the rand scheme.
func GetRandom(p []byte, flags GetRandomFlag) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	r1, _, errno := syscall6(uintptr(unsafe.Pointer(&libc_getrandom)),
		3,
		uintptr(unsafe.Pointer(&p[0])),
		uintptr(len(p)),
		uintptr(flags),
		0, 0, 0)
	if errno != 0 {
		return 0, errno
	}
	return int(r1), nil
}
