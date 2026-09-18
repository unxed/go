// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import "unsafe"

// Hurd's mcontext_t/ucontext_t (usr/include/x86_64-gnu/sys/ucontext.h) is
// glibc's generic x86_64 gregset_t/REG_* layout -- the same one Linux and
// Solaris use -- not Haiku's flat named-field mcontext. Register offsets
// (REG_R8=0 .. REG_CR2=22) were confirmed with abi_probe.c against the
// real header; there is no REG_FS/REG_GS slot in Hurd's 23-entry gregset,
// so fs()/gs() are stubbed like Haiku already does for the same reason.

const (
	_REG_R8      = 0
	_REG_R9      = 1
	_REG_R10     = 2
	_REG_R11     = 3
	_REG_R12     = 4
	_REG_R13     = 5
	_REG_R14     = 6
	_REG_R15     = 7
	_REG_RDI     = 8
	_REG_RSI     = 9
	_REG_RBP     = 10
	_REG_RSP     = 11
	_REG_RBX     = 12
	_REG_RDX     = 13
	_REG_RCX     = 14
	_REG_RAX     = 15
	_REG_RIP     = 16
	_REG_CS      = 17
	_REG_RFL     = 18
	_REG_ERR     = 19
	_REG_TRAPNO  = 20
	_REG_OLDMASK = 21
	_REG_CR2     = 22
)

type sigctxt struct {
	info *siginfo
	ctxt unsafe.Pointer
}

//go:nosplit
//go:nowritebarrierrec
func (c *sigctxt) regs() *mcontext {
	return (*mcontext)(unsafe.Pointer(&(*ucontext)(c.ctxt).uc_mcontext))
}

func (c *sigctxt) rax() uint64 { return c.regs().gregs[_REG_RAX] }
func (c *sigctxt) rbx() uint64 { return c.regs().gregs[_REG_RBX] }
func (c *sigctxt) rcx() uint64 { return c.regs().gregs[_REG_RCX] }
func (c *sigctxt) rdx() uint64 { return c.regs().gregs[_REG_RDX] }
func (c *sigctxt) rdi() uint64 { return c.regs().gregs[_REG_RDI] }
func (c *sigctxt) rsi() uint64 { return c.regs().gregs[_REG_RSI] }
func (c *sigctxt) rbp() uint64 { return c.regs().gregs[_REG_RBP] }
func (c *sigctxt) rsp() uint64 { return c.regs().gregs[_REG_RSP] }
func (c *sigctxt) r8() uint64  { return c.regs().gregs[_REG_R8] }
func (c *sigctxt) r9() uint64  { return c.regs().gregs[_REG_R9] }
func (c *sigctxt) r10() uint64 { return c.regs().gregs[_REG_R10] }
func (c *sigctxt) r11() uint64 { return c.regs().gregs[_REG_R11] }
func (c *sigctxt) r12() uint64 { return c.regs().gregs[_REG_R12] }
func (c *sigctxt) r13() uint64 { return c.regs().gregs[_REG_R13] }
func (c *sigctxt) r14() uint64 { return c.regs().gregs[_REG_R14] }
func (c *sigctxt) r15() uint64 { return c.regs().gregs[_REG_R15] }

//go:nosplit
//go:nowritebarrierrec
func (c *sigctxt) rip() uint64 { return c.regs().gregs[_REG_RIP] }

func (c *sigctxt) rflags() uint64 { return c.regs().gregs[_REG_RFL] }
func (c *sigctxt) cs() uint64     { return c.regs().gregs[_REG_CS] }

// Hurd's gregset_t has no fs/gs slots (unlike Linux's packed CSGSFS).
func (c *sigctxt) fs() uint64 { return 0 }
func (c *sigctxt) gs() uint64 { return 0 }

func (c *sigctxt) sigcode() uint64 { return uint64(c.info.si_code) }
func (c *sigctxt) sigaddr() uint64 { return uint64(c.info.si_addr) }

func (c *sigctxt) set_rip(x uint64)     { c.regs().gregs[_REG_RIP] = x }
func (c *sigctxt) set_rsp(x uint64)     { c.regs().gregs[_REG_RSP] = x }
func (c *sigctxt) set_sigcode(x uint64) { c.info.si_code = int32(x) }
func (c *sigctxt) set_sigaddr(x uint64) { c.info.si_addr = uintptr(x) }

// glibc's Hurd signal delivery has two properties Go must work around.
//
// 1. Register changes are lost. The ucontext_t passed to an SA_SIGINFO handler is
// a copy made by fill_ucontext(); __sigreturn() restores the thread from the
// struct sigcontext that sits in the same stack frame, below the ucontext_t.
// Go's async preemption and sigpanic injection rewrite RIP/RSP in the ucontext,
// so the changes must be written back into the sigcontext. gregs[REG_R8..REG_RFL]
// mirror the sigcontext from sc_r8 on (measured on real Hurd, unxed/debian-hurd
// poc/ctx_poc.c).
//
// 2. The alternate signal stack is selected by the SS_ONSTACK *flag*, not by the
// stack pointer. While the flag is set (a signal handler is active, or
// __sigreturn is replaying a signal that was pending), the next handler runs on
// the interrupted user stack below the red zone: for us, on a goroutine stack
// instead of the gsignal stack.
const (
	_NGREGS_MIRROR = 19 // R8 .. RFL
	_SC_SEARCH_MAX = 2048
)

// sigtrampgoIndirect breaks the linker's nosplit stack-depth chain between
// sigtrampgohurd and sigtrampgo (the same trick as adjustSignalStack2Indirect).
var sigtrampgoIndirect = sigtrampgo

// sigtrampgohurd is called by sigtramp (sys_hurd_amd64.s) and wraps sigtrampgo.
//
//go:nosplit
//go:nowritebarrierrec
func sigtrampgohurd(sig uint32, info *siginfo, ctx unsafe.Pointer) {
	mc := &(*ucontext)(ctx).uc_mcontext
	orig := mc.gregs

	// Delivered on the interrupted goroutine's stack (see 2 above): declare that
	// stack the signal stack for the duration of the handler, as
	// adjustSignalStack does for the g0 stack.
	var saved gsignalStack
	adjusted := false
	if gp := getg(); gp != nil && gp.m != nil && gp.m.gsignal != nil && gp != gp.m.g0 {
		sp := uintptr(unsafe.Pointer(&sig))
		if (sp < gp.m.gsignal.stack.lo || sp >= gp.m.gsignal.stack.hi) && sp >= gp.stack.lo && sp < gp.stack.hi {
			st := stackt{ss_size: gp.stack.hi - gp.stack.lo}
			setSignalstackSP(&st, gp.stack.lo)
			setGsignalStack(&st, &saved)
			adjusted = true
		}
	}

	sigtrampgoIndirect(sig, info, ctx)

	if adjusted {
		restoreGsignalStack(&saved)
	}

	// Propagate register changes to the sigcontext (see 1 above). The block is
	// found by matching the original registers, so this does not depend on
	// glibc's exact frame layout; if it cannot be found die loudly instead of
	// re-faulting forever.
	if mc.gregs == orig {
		return
	}
	want := (*[_NGREGS_MIRROR]uint64)(unsafe.Pointer(&orig))
	for a := uintptr(ctx) - _NGREGS_MIRROR*8; a >= uintptr(ctx)-_SC_SEARCH_MAX; a -= 8 {
		sc := (*[_NGREGS_MIRROR]uint64)(unsafe.Pointer(a))
		if *sc == *want {
			*sc = *(*[_NGREGS_MIRROR]uint64)(unsafe.Pointer(&mc.gregs))
			return
		}
	}
	println("sigtrampgohurd: sig", sig, "changed registers but the sigcontext was not found")
	throw("sigtrampgohurd: cannot propagate register changes")
}
