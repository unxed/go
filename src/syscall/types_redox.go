// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore

/*
Input to cgo -godefs. Run by CI (.github/workflows/redox-gen-syscall.yml)
with redoxer's cross compiler, so the layouts come from relibc's real
headers rather than being written by hand:

	GOARCH=amd64 CC=redox-cc go tool cgo -godefs types_redox.go
*/

// +godefs map struct_in_addr [4]byte /* in_addr */
// +godefs map struct_in6_addr [16]byte /* in6_addr */

package syscall

/*
#include <dirent.h>
#include <fcntl.h>
#include <signal.h>
#include <termios.h>
#include <stdio.h>
#include <unistd.h>
#include <limits.h>
#include <sys/mman.h>
#include <sys/resource.h>
#include <sys/select.h>
#include <sys/socket.h>
#include <sys/stat.h>
#include <sys/time.h>
#include <sys/types.h>
#include <sys/un.h>
#include <sys/wait.h>
#include <net/if.h>
#include <netinet/in.h>
#include <netinet/tcp.h>

enum {
	sizeofPtr = sizeof(void*),
};

union sockaddr_all {
	struct sockaddr s1;	// this one gets used for fields
	struct sockaddr_in s2;	// these pad it out
	struct sockaddr_in6 s3;
	struct sockaddr_un s4;
};

struct sockaddr_any {
	struct sockaddr addr;
	char pad[sizeof(union sockaddr_all) - sizeof(struct sockaddr)];
};

*/
import "C"

// Machine characteristics; for internal use.

const (
	sizeofPtr      = C.sizeofPtr
	sizeofShort    = C.sizeof_short
	sizeofInt      = C.sizeof_int
	sizeofLong     = C.sizeof_long
	sizeofLongLong = C.sizeof_longlong
	PathMax        = C.PATH_MAX
)

// Basic types

type (
	_C_short     C.short
	_C_int       C.int
	_C_long      C.long
	_C_long_long C.longlong
)

// Time

type Timespec C.struct_timespec

type Timeval C.struct_timeval

// Processes

type Rusage C.struct_rusage

type Rlimit C.struct_rlimit

type _Pid_t C.pid_t

type _Gid_t C.gid_t

// Files

type Stat_t C.struct_stat

type Flock_t C.struct_flock

type Dirent C.struct_dirent

// Sockets

type RawSockaddrInet4 C.struct_sockaddr_in

type RawSockaddrInet6 C.struct_sockaddr_in6

type RawSockaddrUnix C.struct_sockaddr_un

type RawSockaddr C.struct_sockaddr

type RawSockaddrAny C.struct_sockaddr_any

type _Socklen C.socklen_t

type Linger C.struct_linger

type Iovec C.struct_iovec

type IPMreq C.struct_ip_mreq

type IPv6Mreq C.struct_ipv6_mreq

type Msghdr C.struct_msghdr

type Cmsghdr C.struct_cmsghdr

const (
	SizeofSockaddrInet4 = C.sizeof_struct_sockaddr_in
	SizeofSockaddrInet6 = C.sizeof_struct_sockaddr_in6
	SizeofSockaddrAny   = C.sizeof_struct_sockaddr_any
	SizeofSockaddrUnix  = C.sizeof_struct_sockaddr_un
	SizeofLinger        = C.sizeof_struct_linger
	SizeofIPMreq        = C.sizeof_struct_ip_mreq
	SizeofIPv6Mreq      = C.sizeof_struct_ipv6_mreq
	SizeofMsghdr        = C.sizeof_struct_msghdr
	SizeofCmsghdr       = C.sizeof_struct_cmsghdr
)

// Select

type FdSet C.fd_set

// Misc

const (
	_AT_FDCWD = C.AT_FDCWD
)

// Terminal handling

type Termios C.struct_termios
