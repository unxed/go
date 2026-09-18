//go:build hurd

package unix

import (
	"syscall"
	"unsafe"
)

// Types.
type (
	Errno         = syscall.Errno
	Signal        = syscall.Signal
	Termios       = syscall.Termios
	Timespec      = syscall.Timespec
	Timeval       = syscall.Timeval
	Stat_t        = syscall.Stat_t
	Rlimit        = syscall.Rlimit
	Rusage        = syscall.Rusage
	Flock_t       = syscall.Flock_t
	Dirent        = syscall.Dirent
	FdSet         = syscall.FdSet
	Iovec         = syscall.Iovec
	Msghdr        = syscall.Msghdr
	Cmsghdr       = syscall.Cmsghdr
	WaitStatus    = syscall.WaitStatus
	Sockaddr      = syscall.Sockaddr
	SockaddrInet4 = syscall.SockaddrInet4
	SockaddrInet6 = syscall.SockaddrInet6
	SockaddrUnix  = syscall.SockaddrUnix
	Linger        = syscall.Linger
	IPMreq        = syscall.IPMreq
	IPv6Mreq      = syscall.IPv6Mreq
)

// Winsize is struct winsize.
type Winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

// PollFd is struct pollfd.
type PollFd struct {
	Fd      int32
	Events  int16
	Revents int16
}

// Opaque request codes understood by IoctlGetTermios/IoctlSetTermios (Hurd has
// no TCGETS ioctl numbers; these map to tcgetattr/tcsetattr).
const (
	TCGETS  = 0x5401
	TCSETS  = 0x5402
	TCSETSW = 0x5403
	TCSETSF = 0x5404
	TCGETA  = 0x5405
	TCSETA  = 0x5406
)

const (
	tcsaNow   = 0
	tcsaDrain = 1
	tcsaFlush = 2
)

func sysvicall6(trap, nargs, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err syscall.Errno)
func rawSysvicall6(trap, nargs, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err syscall.Errno)

type libcFunc uintptr

//go:cgo_import_dynamic libc_tcgetattr tcgetattr "libc.so.0.3"
//go:cgo_import_dynamic libc_tcsetattr tcsetattr "libc.so.0.3"
//go:cgo_import_dynamic libc_ioctl ioctl "libc.so.0.3"
//go:cgo_import_dynamic libc_poll poll "libc.so.0.3"
//go:cgo_import_dynamic libc_mprotect mprotect "libc.so.0.3"
//go:cgo_import_dynamic libc_fcntl fcntl "libc.so.0.3"
//go:cgo_import_dynamic libc_fchmodat fchmodat "libc.so.0.3"
//go:cgo_import_dynamic libc_utimensat utimensat "libc.so.0.3"
//go:cgo_import_dynamic libc_uname uname "libc.so.0.3"
//go:cgo_import_dynamic libc_select select "libc.so.0.3"

//go:linkname libc_tcgetattr libc_tcgetattr
//go:linkname libc_tcsetattr libc_tcsetattr
//go:linkname libc_ioctl libc_ioctl
//go:linkname libc_poll libc_poll
//go:linkname libc_mprotect libc_mprotect
//go:linkname libc_fcntl libc_fcntl
//go:linkname libc_fchmodat libc_fchmodat
//go:linkname libc_utimensat libc_utimensat
//go:linkname libc_uname libc_uname
//go:linkname libc_select libc_select

var (
	libc_tcgetattr,
	libc_tcsetattr,
	libc_ioctl,
	libc_poll,
	libc_mprotect,
	libc_fcntl,
	libc_fchmodat,
	libc_utimensat,
	libc_uname,
	libc_select libcFunc
)

func call(fn *libcFunc, n uintptr, a1, a2, a3, a4, a5, a6 uintptr) (uintptr, error) {
	r, _, e := sysvicall6(uintptr(unsafe.Pointer(fn)), n, a1, a2, a3, a4, a5, a6)
	if e != 0 {
		return r, e
	}
	return r, nil
}

func callErr(fn *libcFunc, n uintptr, a1, a2, a3, a4, a5, a6 uintptr) error {
	r, _, e := sysvicall6(uintptr(unsafe.Pointer(fn)), n, a1, a2, a3, a4, a5, a6)
	if int(r) == -1 {
		if e == 0 {
			e = syscall.EINVAL
		}
		return e
	}
	return nil
}

// Termios and window size.

func IoctlGetTermios(fd int, req uint) (*Termios, error) {
	var t Termios
	if err := callErr(&libc_tcgetattr, 2, uintptr(fd), uintptr(unsafe.Pointer(&t)), 0, 0, 0, 0); err != nil {
		return nil, err
	}
	return &t, nil
}

func IoctlSetTermios(fd int, req uint, value *Termios) error {
	how := uintptr(tcsaNow)
	switch req {
	case TCSETSW:
		how = tcsaDrain
	case TCSETSF:
		how = tcsaFlush
	}
	return callErr(&libc_tcsetattr, 3, uintptr(fd), how, uintptr(unsafe.Pointer(value)), 0, 0, 0)
}

func IoctlGetWinsize(fd int, req uint) (*Winsize, error) {
	var ws Winsize
	if err := callErr(&libc_ioctl, 3, uintptr(fd), uintptr(req), uintptr(unsafe.Pointer(&ws)), 0, 0, 0); err != nil {
		return nil, err
	}
	return &ws, nil
}

func IoctlSetWinsize(fd int, req uint, value *Winsize) error {
	return callErr(&libc_ioctl, 3, uintptr(fd), uintptr(req), uintptr(unsafe.Pointer(value)), 0, 0, 0)
}

func IoctlGetInt(fd int, req uint) (int, error) {
	var v int32
	if err := callErr(&libc_ioctl, 3, uintptr(fd), uintptr(req), uintptr(unsafe.Pointer(&v)), 0, 0, 0); err != nil {
		return 0, err
	}
	return int(v), nil
}

func IoctlSetInt(fd int, req uint, value int) error {
	v := int32(value)
	return callErr(&libc_ioctl, 3, uintptr(fd), uintptr(req), uintptr(unsafe.Pointer(&v)), 0, 0, 0)
}

// Poll.

func Poll(fds []PollFd, timeout int) (int, error) {
	var p unsafe.Pointer
	if len(fds) > 0 {
		p = unsafe.Pointer(&fds[0])
	}
	r, _, e := sysvicall6(uintptr(unsafe.Pointer(&libc_poll)), 3, uintptr(p), uintptr(len(fds)), uintptr(timeout), 0, 0, 0)
	if int(r) == -1 {
		return -1, e
	}
	return int(r), nil
}

// Memory.

func Getpagesize() int { return syscall.Getpagesize() }

func Mmap(fd int, offset int64, length int, prot int, flags int) ([]byte, error) {
	return syscall.Mmap(fd, offset, length, prot, flags)
}

func Munmap(b []byte) error { return syscall.Munmap(b) }

func Mprotect(b []byte, prot int) error {
	if len(b) == 0 {
		return nil
	}
	return callErr(&libc_mprotect, 3, uintptr(unsafe.Pointer(&b[0])), uintptr(len(b)), uintptr(prot), 0, 0, 0)
}

// Files.

func Close(fd int) error                                { return syscall.Close(fd) }
func Open(path string, mode int, perm uint32) (int, error) { return syscall.Open(path, mode, perm) }
func Read(fd int, p []byte) (int, error)                { return syscall.Read(fd, p) }
func Write(fd int, p []byte) (int, error)               { return syscall.Write(fd, p) }
func Pread(fd int, p []byte, off int64) (int, error)    { return syscall.Pread(fd, p, off) }
func Pwrite(fd int, p []byte, off int64) (int, error)   { return syscall.Pwrite(fd, p, off) }
func Seek(fd int, off int64, whence int) (int64, error) { return syscall.Seek(fd, off, whence) }
func Dup(fd int) (int, error)                           { return syscall.Dup(fd) }
func Dup2(oldfd, newfd int) error                       { return syscall.Dup2(oldfd, newfd) }
func Pipe(p []int) error                                { return syscall.Pipe(p) }
func Pipe2(p []int, flags int) error                    { return syscall.Pipe2(p, flags) }
func Fstat(fd int, st *Stat_t) error                    { return syscall.Fstat(fd, st) }
func Stat(path string, st *Stat_t) error                { return syscall.Stat(path, st) }
func Lstat(path string, st *Stat_t) error               { return syscall.Lstat(path, st) }
func Chdir(path string) error                           { return syscall.Chdir(path) }
func Fchdir(fd int) error                               { return syscall.Fchdir(fd) }
func Getcwd(buf []byte) (int, error)                    { return syscall.Getcwd(buf) }
func Mkdir(path string, mode uint32) error              { return syscall.Mkdir(path, mode) }
func Rmdir(path string) error                           { return syscall.Rmdir(path) }
func Unlink(path string) error                          { return syscall.Unlink(path) }
func Rename(from, to string) error                      { return syscall.Rename(from, to) }
func Symlink(path, link string) error                   { return syscall.Symlink(path, link) }
func Link(path, link string) error                      { return syscall.Link(path, link) }
func Readlink(path string, buf []byte) (int, error)     { return syscall.Readlink(path, buf) }
func Chmod(path string, mode uint32) error              { return syscall.Chmod(path, mode) }
func Fchmod(fd int, mode uint32) error                  { return syscall.Fchmod(fd, mode) }
func Chown(path string, uid, gid int) error             { return syscall.Chown(path, uid, gid) }
func Fchown(fd int, uid, gid int) error                 { return syscall.Fchown(fd, uid, gid) }
func Lchown(path string, uid, gid int) error            { return syscall.Lchown(path, uid, gid) }
func Ftruncate(fd int, length int64) error              { return syscall.Ftruncate(fd, length) }
func Truncate(path string, length int64) error          { return syscall.Truncate(path, length) }
func Fsync(fd int) error                                { return syscall.Fsync(fd) }
func Sync()                                             { syscall.Sync() }
func Access(path string, mode uint32) error             { return syscall.Access(path, mode) }
func Umask(mask int) int                                { return syscall.Umask(mask) }
func Getpid() int                                       { return syscall.Getpid() }
func Getppid() int                                      { return syscall.Getppid() }
func Getuid() int                                       { return syscall.Getuid() }
func Geteuid() int                                      { return syscall.Geteuid() }
func Getgid() int                                       { return syscall.Getgid() }
func Getegid() int                                      { return syscall.Getegid() }
func Getgroups() ([]int, error)                         { return syscall.Getgroups() }
func Kill(pid int, sig syscall.Signal) error            { return syscall.Kill(pid, sig) }
func Setsid() (int, error)                              { return syscall.Setsid() }
func Getrlimit(which int, lim *Rlimit) error            { return syscall.Getrlimit(which, lim) }
func Setrlimit(which int, lim *Rlimit) error            { return syscall.Setrlimit(which, lim) }
func Getrusage(who int, ru *Rusage) error               { return syscall.Getrusage(who, ru) }
func Gettimeofday(tv *Timeval) error                    { return syscall.Gettimeofday(tv) }
func Nanosleep(t, rem *Timespec) error                  { return syscall.Nanosleep(t, rem) }
func Wait4(pid int, ws *WaitStatus, options int, ru *Rusage) (int, error) {
	return syscall.Wait4(pid, ws, options, ru)
}
func SetNonblock(fd int, nonblocking bool) error { return syscall.SetNonblock(fd, nonblocking) }
func CloseOnExec(fd int)                         { syscall.CloseOnExec(fd) }
func UtimesNano(path string, ts []Timespec) error { return syscall.UtimesNano(path, ts) }
func FcntlFlock(fd uintptr, cmd int, lk *Flock_t) error { return syscall.FcntlFlock(fd, cmd, lk) }

func SetsockoptInt(fd, level, opt int, value int) error { return syscall.SetsockoptInt(fd, level, opt, value) }
func GetsockoptInt(fd, level, opt int) (int, error)     { return syscall.GetsockoptInt(fd, level, opt) }

func Fcntl(fd int, cmd int, arg uintptr) (int, error) {
	r, _, e := sysvicall6(uintptr(unsafe.Pointer(&libc_fcntl)), 3, uintptr(fd), uintptr(cmd), arg, 0, 0, 0)
	if int(r) == -1 {
		return -1, e
	}
	return int(r), nil
}

func FcntlInt(fd uintptr, cmd, arg int) (int, error) { return Fcntl(int(fd), cmd, uintptr(arg)) }

func bytePtr(s string) (*byte, error) { return syscall.BytePtrFromString(s) }

func Fchmodat(dirfd int, path string, mode uint32, flags int) error {
	p, err := bytePtr(path)
	if err != nil {
		return err
	}
	return callErr(&libc_fchmodat, 4, uintptr(dirfd), uintptr(unsafe.Pointer(p)), uintptr(mode), uintptr(flags), 0, 0)
}

func UtimesNanoAt(dirfd int, path string, ts []Timespec, flags int) error {
	if len(ts) != 2 {
		return syscall.EINVAL
	}
	p, err := bytePtr(path)
	if err != nil {
		return err
	}
	return callErr(&libc_utimensat, 4, uintptr(dirfd), uintptr(unsafe.Pointer(p)), uintptr(unsafe.Pointer(&ts[0])), uintptr(flags), 0, 0)
}

// Lutimes sets the access and modification times of a symlink itself.
func Lutimes(path string, tv []Timeval) error {
	if len(tv) != 2 {
		return syscall.EINVAL
	}
	ts := []Timespec{NsecToTimespec(TimevalToNsec(tv[0])), NsecToTimespec(TimevalToNsec(tv[1]))}
	return UtimesNanoAt(AT_FDCWD, path, ts, AT_SYMLINK_NOFOLLOW)
}

// Time conversions.

func TimespecToNsec(ts Timespec) int64 { return ts.Nano() }
func TimevalToNsec(tv Timeval) int64   { return tv.Nano() }

func NsecToTimespec(nsec int64) Timespec { return syscall.NsecToTimespec(nsec) }
func NsecToTimeval(nsec int64) Timeval   { return syscall.NsecToTimeval(nsec) }

// Device numbers (glibc's generic encoding).

func Major(dev uint64) uint32 { return uint32(((dev >> 8) & 0xfff) | ((dev >> 32) &^ 0xfff)) }
func Minor(dev uint64) uint32 { return uint32((dev & 0xff) | ((dev >> 12) &^ 0xff)) }
func Mkdev(major, minor uint32) uint64 {
	maj, min := uint64(major), uint64(minor)
	return ((maj &^ 0xfff) << 32) | ((maj & 0xfff) << 8) | ((min &^ 0xff) << 12) | (min & 0xff)
}

// Utsname is struct utsname with fixed-size fields as in glibc.
type Utsname struct {
	Sysname    [1024]byte
	Nodename   [1024]byte
	Release    [1024]byte
	Version    [1024]byte
	Machine    [1024]byte
}
