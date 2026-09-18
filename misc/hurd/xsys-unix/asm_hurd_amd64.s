// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build hurd

#include "textflag.h"

// libc calls go through the syscall package's sysvicall6 (see runtime/syscall_hurd.go),
// the same way golang.org/x/sys/unix does it on Solaris.

TEXT ·sysvicall6(SB),NOSPLIT,$0-88
	JMP	syscall·sysvicall6(SB)

TEXT ·rawSysvicall6(SB),NOSPLIT,$0-88
	JMP	syscall·rawSysvicall6(SB)
