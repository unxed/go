// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

TEXT _rt0_amd64_hurd(SB),NOSPLIT,$-8
	// _rt0_amd64_hurd is the real ELF entry point (Hurd's /hurd/exec
	// loads us directly, there is no glibc _start in front of it), so
	// argc/argv/envp/auxv are on the stack per the standard SysV ABI
	// exec() convention, same as Linux -- use the generic extractor
	// rather than jumping straight to rt0_go with garbage in DI/SI.
	JMP	_rt0_amd64(SB)

TEXT _rt0_amd64_hurd_lib(SB),NOSPLIT,$0
	JMP	_rt0_amd64_lib(SB)
