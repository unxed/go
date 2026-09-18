// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

// relibc's ld.so jumps to the executable's entry point with rsp restored to
// the original process stack (argc at 0(SP), argv at 8(SP)) and rdi/rsi
// zeroed, so unlike Haiku's loader we must load argc/argv from the stack
// ourselves -- _rt0_amd64 does that before falling into rt0_go.
TEXT _rt0_amd64_redox(SB),NOSPLIT,$-8
	JMP	_rt0_amd64(SB)

TEXT _rt0_amd64_redox_lib(SB),NOSPLIT,$0
	JMP	_rt0_amd64_lib(SB)
