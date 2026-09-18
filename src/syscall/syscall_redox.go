// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Redox system calls (through relibc's libc.so.6).
// This file is compiled as ordinary Go code,
// but it is also input to mksyscall,
// which parses the //sys lines and generates system call stubs.
// Note that sometimes we use a lowercase //sys name and wrap
// it in our own nicer implementation, either here or in
// syscall_unix.go.

package syscall

import "unsafe"

func Syscall(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err Errno)
func RawSyscall(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err Errno)

// Implemented in asm_redox_amd64.s.
func rawSysvicall6(trap, nargs, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err Errno)
func sysvicall6(trap, nargs, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err Errno)

// Constant expected by package but not supported
const (
	_ = iota
	TIOCNOTTY
)

const (
	_F_DUP2FD_CLOEXEC = 0
)


/*type SockaddrDatalink struct {
	Family uint16
	Index  uint16
	Type   uint8
	Nlen   uint8
	Alen   uint8
	Slen   uint8
	Data   [244]int8
	raw    RawSockaddrDatalink
}*/


func direntIno(buf []byte) (uint64, bool) {
	return readInt(buf, unsafe.Offsetof(Dirent{}.Ino), unsafe.Sizeof(Dirent{}.Ino))
}

func direntReclen(buf []byte) (uint64, bool) {
	return readInt(buf, unsafe.Offsetof(Dirent{}.Reclen), unsafe.Sizeof(Dirent{}.Reclen))
}

func direntNamlen(buf []byte) (uint64, bool) {
	reclen, ok := direntReclen(buf)
	if !ok {
		return 0, false
	}
	return reclen - uint64(unsafe.Offsetof(Dirent{}.Name)), true
}


func pipe() (r uintptr, w uintptr, err uintptr)

func Pipe(p []int) (err error) {
	if len(p) != 2 {
		return EINVAL
	}
	r0, w0, e1 := pipe()
	if e1 != 0 {
		err = Errno(e1)
	}
	if err == nil {
		p[0], p[1] = int(r0), int(w0)
	}
	return
}

func (sa *SockaddrInet4) sockaddr() (unsafe.Pointer, _Socklen, error) {
	if sa.Port < 0 || sa.Port > 0xFFFF {
		return nil, 0, EINVAL
	}
	sa.raw.Family = AF_INET
	p := (*[2]byte)(unsafe.Pointer(&sa.raw.Port))
	p[0] = byte(sa.Port >> 8)
	p[1] = byte(sa.Port)
	for i := 0; i < len(sa.Addr); i++ {
		sa.raw.Addr[i] = sa.Addr[i]
	}
	return unsafe.Pointer(&sa.raw), SizeofSockaddrInet4, nil
}

func (sa *SockaddrInet6) sockaddr() (unsafe.Pointer, _Socklen, error) {
	if sa.Port < 0 || sa.Port > 0xFFFF {
		return nil, 0, EINVAL
	}
	sa.raw.Family = AF_INET6
	p := (*[2]byte)(unsafe.Pointer(&sa.raw.Port))
	p[0] = byte(sa.Port >> 8)
	p[1] = byte(sa.Port)
	sa.raw.Scope_id = sa.ZoneId
	for i := 0; i < len(sa.Addr); i++ {
		sa.raw.Addr[i] = sa.Addr[i]
	}
	return unsafe.Pointer(&sa.raw), SizeofSockaddrInet6, nil
}

func (sa *SockaddrUnix) sockaddr() (unsafe.Pointer, _Socklen, error) {
	name := sa.Name
	n := len(name)
	if n >= len(sa.raw.Path) {
		return nil, 0, EINVAL
	}
	sa.raw.Family = AF_UNIX
	for i := 0; i < n; i++ {
		sa.raw.Path[i] = int8(name[i])
	}
	// length is family (uint16), name, NUL.
	sl := _Socklen(2)
	if n > 0 {
		sl += _Socklen(n) + 1
	}
	if sa.raw.Path[0] == '@' {
		sa.raw.Path[0] = 0
		// Don't count trailing NUL for abstract address.
		sl--
	}

	return unsafe.Pointer(&sa.raw), sl, nil
}

func Getsockname(fd int) (sa Sockaddr, err error) {
	var rsa RawSockaddrAny
	var len _Socklen = SizeofSockaddrAny
	if err = getsockname(fd, &rsa, &len); err != nil {
		return
	}
	return anyToSockaddr(&rsa)
}

const ImplementsGetwd = true

//sys	Getcwd(buf []byte) (n int, err error)

func Getwd() (wd string, err error) {
	var buf [PathMax]byte
	// Getcwd will return an error if it failed for any reason.
	_, err = Getcwd(buf[0:])
	if err != nil {
		return "", err
	}
	n := clen(buf[:])
	if n < 1 {
		return "", EINVAL
	}
	return string(buf[:n]), nil
}

/*
 * Wrapped
 */

//sysnb	getgroups(ngid int, gid *_Gid_t) (n int, err error)
//sysnb	setgroups(ngid int, gid *_Gid_t) (err error)

func Getgroups() (gids []int, err error) {
	n, err := getgroups(0, nil)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, nil
	}

	// Sanity check group count.  Max is 16 on BSD.
	if n < 0 || n > 1000 {
		return nil, EINVAL
	}

	a := make([]_Gid_t, n)
	n, err = getgroups(n, &a[0])
	if err != nil {
		return nil, err
	}
	gids = make([]int, n)
	for i, v := range a[0:n] {
		gids[i] = int(v)
	}
	return
}

func Setgroups(gids []int) (err error) {
	if len(gids) == 0 {
		return setgroups(0, nil)
	}

	a := make([]_Gid_t, len(gids))
	for i, v := range gids {
		a[i] = _Gid_t(v)
	}
	return setgroups(len(a), &a[0])
}

// ReadDirent is not supported on Redox: relibc's readdir_r is unimplemented
// and there is no getdents wrapper, so directories are read through
// fdopendir/readdir instead (see Readdir below and os/dir_redox.go).
func ReadDirent(fd int, buf []byte) (n int, err error) {
	return 0, ENOSYS
}

// Getdirentries is not supported; see ReadDirent.
func Getdirentries(fd int, buf []byte, basep *uintptr) (n int, err error) {
	return 0, ENOSYS
}

// Readdir returns the next entry of the directory stream dir, or nil at the
// end of the stream. The returned entry is only valid until the next call.
func Readdir(dir uintptr) (*Dirent, error) {
	p, err := readdir(dir)
	if p == 0 {
		// errno is cleared before each libc call, so err is nil at EOF.
		return nil, err
	}
	return (*Dirent)(unsafe.Pointer(p)), nil
}

// relibc uses the Linux wait status encoding (see sys_wait/cbindgen.toml).
type WaitStatus uint32

// Wait status is 7 bits at bottom, either 0 (exited),
// 0x7F (stopped), or a signal number that caused an exit.
// The 0x80 bit is whether there was a core dump.
// An extra number (exit code, signal causing a stop)
// is in the high bits. At least that's the idea.
// There are various irregularities. For example, the
// "continued" status is 0xFFFF, distinguishing itself
// from stopped via the core dump bit.

const (
	mask    = 0x7F
	core    = 0x80
	exited  = 0x00
	stopped = 0x7F
	shift   = 8
)

func (w WaitStatus) Exited() bool { return w&mask == exited }

func (w WaitStatus) Signaled() bool { return w&mask != stopped && w&mask != exited }

func (w WaitStatus) Stopped() bool { return w&0xFF == stopped }

func (w WaitStatus) Continued() bool { return w == 0xFFFF }

func (w WaitStatus) CoreDump() bool { return w.Signaled() && w&core != 0 }

func (w WaitStatus) ExitStatus() int {
	if !w.Exited() {
		return -1
	}
	return int(w>>shift) & 0xFF
}

func (w WaitStatus) Signal() Signal {
	if !w.Signaled() {
		return -1
	}
	return Signal(w & mask)
}

func (w WaitStatus) StopSignal() Signal {
	if !w.Stopped() {
		return -1
	}
	return Signal(w>>shift) & 0xFF
}

func (w WaitStatus) TrapCause() int {
	if w.StopSignal() != SIGTRAP {
		return -1
	}
	return int(w>>shift) >> 8
}

func wait4(pid uintptr, wstatus *WaitStatus, options uintptr, rusage *Rusage) (wpid uintptr, err uintptr)

func Wait4(pid int, wstatus *WaitStatus, options int, rusage *Rusage) (wpid int, err error) {
	r0, e1 := wait4(uintptr(pid), wstatus, uintptr(options), rusage)
	if e1 != 0 {
		err = Errno(e1)
	}
	return int(r0), err
}

func gethostname() (name string, err uintptr)

func Gethostname() (name string, err error) {
	name, e1 := gethostname()
	if e1 != 0 {
		err = Errno(e1)
	}
	return name, err
}

func UtimesNano(path string, ts []Timespec) error {
	if len(ts) != 2 {
		return EINVAL
	}
	return utimensat(_AT_FDCWD, path, (*[2]Timespec)(unsafe.Pointer(&ts[0])), 0)
}

//sys	fcntl(fd int, cmd int, arg int) (val int, err error)

// FcntlFlock performs a fcntl syscall for the F_GETLK, F_SETLK or F_SETLKW command.
func FcntlFlock(fd uintptr, cmd int, lk *Flock_t) error {
	_, _, e1 := sysvicall6(uintptr(unsafe.Pointer(&libc_Fcntl)), 3, uintptr(fd), uintptr(cmd), uintptr(unsafe.Pointer(lk)), 0, 0, 0)
	if e1 != 0 {
		return e1
	}
	return nil
}

func anyToSockaddr(rsa *RawSockaddrAny) (Sockaddr, error) {
	switch rsa.Addr.Family {
	case AF_UNIX:
		pp := (*RawSockaddrUnix)(unsafe.Pointer(rsa))
		sa := new(SockaddrUnix)
		// Assume path ends at NUL.
		// This is not technically the Solaris semantics for
		// abstract Unix domain sockets -- they are supposed
		// to be uninterpreted fixed-size binary blobs -- but
		// everyone uses this convention.
		n := 0
		for n < len(pp.Path) && pp.Path[n] != 0 {
			n++
		}
		bytes := (*[10000]byte)(unsafe.Pointer(&pp.Path[0]))[0:n]
		sa.Name = string(bytes)
		return sa, nil

	case AF_INET:
		pp := (*RawSockaddrInet4)(unsafe.Pointer(rsa))
		sa := new(SockaddrInet4)
		p := (*[2]byte)(unsafe.Pointer(&pp.Port))
		sa.Port = int(p[0])<<8 + int(p[1])
		for i := 0; i < len(sa.Addr); i++ {
			sa.Addr[i] = pp.Addr[i]
		}
		return sa, nil

	case AF_INET6:
		pp := (*RawSockaddrInet6)(unsafe.Pointer(rsa))
		sa := new(SockaddrInet6)
		p := (*[2]byte)(unsafe.Pointer(&pp.Port))
		sa.Port = int(p[0])<<8 + int(p[1])
		sa.ZoneId = pp.Scope_id
		for i := 0; i < len(sa.Addr); i++ {
			sa.Addr[i] = pp.Addr[i]
		}
		return sa, nil
	}
	return nil, EAFNOSUPPORT
}

//sys	accept(s int, rsa *RawSockaddrAny, addrlen *_Socklen) (fd int, err error)
func Accept(fd int) (nfd int, sa Sockaddr, err error) {
	var rsa RawSockaddrAny
	var len _Socklen = SizeofSockaddrAny
	nfd, err = accept(fd, &rsa, &len)
	if err != nil {
		return
	}
	sa, err = anyToSockaddr(&rsa)
	if err != nil {
		Close(nfd)
		nfd = 0
	}
	return
}

func recvmsgRaw(fd int, p, oob []byte, flags int, rsa *RawSockaddrAny) (n, oobn int, recvflags int, err error) {
	var msg Msghdr
	msg.Name = (*byte)(unsafe.Pointer(rsa))
	msg.Namelen = uint32(SizeofSockaddrAny)
	var iov Iovec
	if len(p) > 0 {
		iov.Base = (*byte)(unsafe.Pointer(&p[0]))
		iov.SetLen(len(p))
	}
	var dummy byte
	if len(oob) > 0 {
		// receive at least one normal byte
		if len(p) == 0 {
			iov.Base = &dummy
			iov.SetLen(1)
		}
		msg.Control = (*byte)(unsafe.Pointer(&oob[0]))
		msg.Controllen = uint64(len(oob))
	}
	msg.Iov = &iov
	msg.Iovlen = 1
	if n, err = recvmsg(fd, &msg, flags); err != nil {
		return
	}
	oobn = int(msg.Controllen)
	return
}

//sys	sendmsg(s int, msg *Msghdr, flags int) (n int, err error)
func sendmsgN(fd int, p, oob []byte, ptr unsafe.Pointer, salen _Socklen, flags int) (n int, err error) {
	var msg Msghdr
	msg.Name = (*byte)(unsafe.Pointer(ptr))
	msg.Namelen = uint32(salen)
	var iov Iovec
	if len(p) > 0 {
		iov.Base = (*byte)(unsafe.Pointer(&p[0]))
		iov.SetLen(len(p))
	}
	var dummy byte
	if len(oob) > 0 {
		// send at least one normal byte
		if len(p) == 0 {
			iov.Base = &dummy
			iov.SetLen(1)
		}
		msg.Control = (*byte)(unsafe.Pointer(&oob[0]))
		msg.Controllen = uint64(len(oob))
	}
	msg.Iov = &iov
	msg.Iovlen = 1
	if n, err = sendmsg(fd, &msg, flags); err != nil {
		return 0, err
	}
	if len(oob) > 0 && len(p) == 0 {
		n = 0
	}
	return n, nil
}

/*
 * Exposed directly
 */
//sys	Access(path string, mode uint32) (err error)
//sys	Chdir(path string) (err error)
//sys	Chmod(path string, mode uint32) (err error)
//sys	Chown(path string, uid int, gid int) (err error)
//sys	Chroot(path string) (err error)
//sys	Close(fd int) (err error)
//sys	Closedir(dir uintptr) (err error)
//sys	Dup(fd int) (nfd int, err error)
//sys	Dup2(old int, new int) (err error)
//sys	Fchdir(fd int) (err error)
//sys	Fchmod(fd int, mode uint32) (err error)
//sys	Fchown(fd int, uid int, gid int) (err error)
//sys	Fdopendir(fd int) (dir uintptr, err error)
//sys	Fpathconf(fd int, name int) (val int, err error)
//sys	Fstat(fd int, stat *Stat_t) (err error)
//sysnb	Getgid() (gid int)
//sysnb	Getpid() (pid int)
//sys	Geteuid() (euid int)
//sys	Getegid() (egid int)
//sys	Getppid() (ppid int)
//sys	Getpriority(which int, who int) (n int, err error)
//sysnb	Getrlimit(which int, lim *Rlimit) (err error)
//sysnb	Getrusage(who int, rusage *Rusage) (err error)
//sysnb	Gettimeofday(tv *Timeval) (err error)
//sysnb	Getuid() (uid int)
//sys	Kill(pid int, signum Signal) (err error)
//sys	Lchown(path string, uid int, gid int) (err error)
//sys	Link(path string, link string) (err error)
//sys	Listen(s int, backlog int) (err error)
//sys	Lstat(path string, stat *Stat_t) (err error)
//sys	Mkdir(path string, mode uint32) (err error)
//sys	Mknod(path string, mode uint32, dev int) (err error)
//sys	Nanosleep(time *Timespec, leftover *Timespec) (err error)
//sys	Open(path string, mode int, perm uint32) (fd int, err error)
//sys	Openat(fd int, path string, flags int, perm uint32) (fdret int, err error)
//sys	Pathconf(path string, name int) (val int, err error)
//sys	pread(fd int, p []byte, offset int64) (n int, err error)
//sys	pwrite(fd int, p []byte, offset int64) (n int, err error)
//sys	read(fd int, p []byte) (n int, err error)
//sys	readdir(dir uintptr) (ent uintptr, err error)
//sys	Readlink(path string, buf []byte) (n int, err error)
//sys	Rename(from string, to string) (err error)
//sys	Rmdir(path string) (err error)
//sys	Seek(fd int, offset int64, whence int) (newoffset int64, err error) = lseek
//sysnb	Setegid(egid int) (err error)
//sysnb	Seteuid(euid int) (err error)
//sysnb	Setgid(gid int) (err error)
//sysnb	Setpgid(pid int, pgid int) (err error)
//sys	Setpriority(which int, who int, prio int) (err error)
//sysnb	Setregid(rgid int, egid int) (err error)
//sysnb	Setreuid(ruid int, euid int) (err error)
//sysnb	setrlimit(which int, lim *Rlimit) (err error)
//sysnb	Setsid() (pid int, err error)
//sysnb	Setuid(uid int) (err error)
//sys	Shutdown(s int, how int) (err error)
//sys	Stat(path string, stat *Stat_t) (err error)
//sys	Symlink(path string, link string) (err error)
//sys	Sync() (err error)
//sys	Truncate(path string, length int64) (err error)
//sys	Fsync(fd int) (err error)
//sys	Ftruncate(fd int, length int64) (err error)
//sys	Umask(newmask int) (oldmask int)
//sys	Unlink(path string) (err error)
//sys	Utimes(path string, times *[2]Timeval) (err error)
//sys	bind(s int, addr unsafe.Pointer, addrlen _Socklen) (err error)
//sys	connect(s int, addr unsafe.Pointer, addrlen _Socklen) (err error)
//sys	mmap(addr uintptr, length uintptr, prot int, flag int, fd int, pos int64) (ret uintptr, err error)
//sys	munmap(addr uintptr, length uintptr) (err error)
//sys	sendto(s int, buf []byte, flags int, to unsafe.Pointer, addrlen _Socklen) (err error)
//sys	socket(domain int, typ int, proto int) (fd int, err error)
//sysnb	socketpair(domain int, typ int, proto int, fd *[2]int32) (err error)
//sys	write(fd int, p []byte) (n int, err error)
//sys	writev(fd int, iovecs []Iovec) (n uintptr, err error)
//sys	getsockopt(s int, level int, name int, val unsafe.Pointer, vallen *_Socklen) (err error)
//sysnb	getpeername(fd int, rsa *RawSockaddrAny, addrlen *_Socklen) (err error)
//sys	getsockname(fd int, rsa *RawSockaddrAny, addrlen *_Socklen) (err error)
//sys	setsockopt(s int, level int, name int, val unsafe.Pointer, vallen uintptr) (err error)
//sys	recvfrom(fd int, p []byte, flags int, from *RawSockaddrAny, fromlen *_Socklen) (n int, err error)
//sys	recvmsg(s int, msg *Msghdr, flags int) (n int, err error)
//sys	utimensat(dirfd int, path string, times *[2]Timespec, flag int) (err error)

func readlen(fd int, buf *byte, nbuf int) (n int, err error) {
	r0, _, e1 := sysvicall6(uintptr(unsafe.Pointer(&libc_Read)), 3, uintptr(fd), uintptr(unsafe.Pointer(buf)), uintptr(nbuf), 0, 0, 0)
	n = int(r0)
	if e1 != 0 {
		err = e1
	}
	return
}

func writelen(fd int, buf *byte, nbuf int) (n int, err error) {
	r0, _, e1 := sysvicall6(uintptr(unsafe.Pointer(&libc_Write)), 3, uintptr(fd), uintptr(unsafe.Pointer(buf)), uintptr(nbuf), 0, 0, 0)
	n = int(r0)
	if e1 != 0 {
		err = e1
	}
	return
}

var mapper = &mmapper{
	active: make(map[*byte][]byte),
	mmap:   mmap,
	munmap: munmap,
}

func Mmap(fd int, offset int64, length int, prot int, flags int) (data []byte, err error) {
	return mapper.Mmap(fd, offset, length, prot, flags)
}

func Munmap(b []byte) (err error) {
	return mapper.Munmap(b)
}

// relibc has no netinet/icmp6.h, so ICMPv6Filter (Linux/BSD layout: eight
// 32-bit words) is declared by hand rather than generated.
type ICMPv6Filter struct {
	Data [8]uint32
}

const SizeofICMPv6Filter = 0x20
