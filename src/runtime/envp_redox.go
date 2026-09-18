// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"internal/goarch"
	"unsafe"
)

// redoxEnvp is the third argument of main(argc, argv, envp), saved by main
// (asm_amd64.s) when a C runtime starts the program, i.e. when cgo makes the
// link external. relibc builds argv separately from the environment block, so
// unlike on the initial process stack, envp does not follow argv there. It is
// 0 when the program was entered through _rt0_amd64_redox (internal linking),
// where the environment does follow argv.
var redoxEnvp uintptr

// envIndex returns the i'th environment string pointer (nil at the end).
func envIndex(i int32) *byte {
	if redoxEnvp != 0 {
		return *(**byte)(add(unsafe.Pointer(redoxEnvp), uintptr(i)*goarch.PtrSize))
	}
	return argv_index(argv, argc+1+i)
}
