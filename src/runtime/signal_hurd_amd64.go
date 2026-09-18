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

// The ucontext_t glibc passes to an SA_SIGINFO handler is a copy made by
// fill_ucontext(); __sigreturn() restores the thread from the struct sigcontext
// in the same stack frame, so changes a handler makes to the ucontext (Go's
// async preemption and sigpanic injection rewrite RIP/RSP) are lost unless they
// are written back. gregs[REG_R8..REG_RFL] mirror the sigcontext from sc_r8
// on. Both numbers were measured on real Hurd (unxed/debian-hurd poc/ctx_poc.c).
const (
	_SC_R8_OFFSET  = 32
	_NGREGS_MIRROR = 19
)

// sighandlerhurd runs sighandler and then propagates register changes to the
// sigcontext scp (see sys_hurd_amd64.s: sigtramp). It first checks that the
// sigcontext still mirrors the registers we were handed, so an unexpected
// layout degrades to "changes ignored" instead of corrupting the frame.
//
//go:nosplit
//go:nowritebarrierrec
func sighandlerhurd(sig uint32, info *siginfo, ctx unsafe.Pointer, gp *g, scp unsafe.Pointer) {
	mc := &(*ucontext)(ctx).uc_mcontext
	orig := mc.gregs
	sighandler(sig, info, ctx, gp)
	if scp == nil || mc.gregs == orig {
		return
	}
	sc := (*[_NGREGS_MIRROR]uint64)(add(scp, _SC_R8_OFFSET))
	if *sc == *(*[_NGREGS_MIRROR]uint64)(unsafe.Pointer(&orig)) {
		*sc = *(*[_NGREGS_MIRROR]uint64)(unsafe.Pointer(&mc.gregs))
	}
}
