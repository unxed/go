// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

// relibc's ld.so jumps to the executable's entry point with rsp restored to
// the original process stack (argc at 0(SP), argv at 8(SP)) and rdi/rsi
// zeroed, so unlike Haiku's loader we must load argc/argv from the stack
// ourselves -- _rt0_amd64 does that before falling into rt0_go.
//
// A process created by relibc's posix_spawn starts with all-zero FPU control
// state: MXCSR = 0 leaves every SSE exception unmasked and, since the kernel
// does not enable CR4.OSXMMEXCPT, the first inexact result raises #UD
// ("Invalid opcode fault" in runtime.fastexprand during mallocinit); the x87
// control word is 0 likewise. fork()ed processes inherit sane values, and
// relibc's crt0 sets them for C programs -- but Go's ELF entry does not go
// through crt0. Set the ABI defaults (MXCSR 0x1f80, x87 CW 0x37f) before
// anything computes.
TEXT _rt0_amd64_redox(SB),NOSPLIT,$-8
	MOVQ	$0x1f80, AX
	PUSHQ	AX
	LDMXCSR	0(SP)
	MOVQ	$0x37f, AX
	MOVQ	AX, 0(SP)
	FLDCW	0(SP)
	POPQ	AX
	JMP	_rt0_amd64(SB)

TEXT _rt0_amd64_redox_lib(SB),NOSPLIT,$0
	JMP	_rt0_amd64_lib(SB)
