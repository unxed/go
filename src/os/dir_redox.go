// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os

import (
	"io"
	"runtime"
	"syscall"
	"unsafe"
)

// Auxiliary information if the File describes a directory
type dirInfo struct {
	dir uintptr // Pointer to DIR structure from dirent.h
}

func (d *dirInfo) close() {
	if d.dir == 0 {
		return
	}
	sysClosedir(d.dir)
	d.dir = 0
}

func (f *File) readdir(n int, mode readdirMode) (names []string, dirents []DirEntry, infos []FileInfo, err error) {
	// If this file has no dirinfo, create one.
	var d *dirInfo
	for {
		d = f.dirinfo.Load()
		if d != nil {
			break
		}
		dir, call, errno := f.pfd.OpenDir()
		if errno != nil {
			return nil, nil, nil, &PathError{Op: call, Path: f.name, Err: errno}
		}
		d = &dirInfo{dir: dir}
		if f.dirinfo.CompareAndSwap(nil, d) {
			break
		}
		// We lost the race: try again.
		d.close()
	}

	size := n
	if size <= 0 {
		size = 100
		n = -1
	}

	for len(names)+len(dirents)+len(infos) < size || n == -1 {
		dirent, err := sysReaddir(d.dir)
		if err != nil {
			if err == syscall.EINTR {
				continue
			}
			return names, dirents, infos, &PathError{Op: "readdir", Path: f.name, Err: err}
		}
		if dirent == nil { // EOF
			break
		}
		if dirent.Ino == 0 {
			continue
		}
		// d_reclen is not a reliable name length on Redox; d_name is NUL
		// terminated within its fixed-size array.
		nameb := (*[len(dirent.Name)]byte)(unsafe.Pointer(&dirent.Name[0]))[:]
		name := nameb
		for i, c := range nameb {
			if c == 0 {
				name = nameb[:i]
				break
			}
		}
		// Check for useless names before allocating a string.
		if string(name) == "." || string(name) == ".." {
			continue
		}
		if mode == readdirName {
			names = append(names, string(name))
		} else if mode == readdirDirEntry {
			de, err := newUnixDirent(f, string(name), ^FileMode(0))
			if IsNotExist(err) {
				// File disappeared between readdir and stat.
				// Treat as if it didn't exist.
				continue
			}
			if err != nil {
				return nil, dirents, nil, err
			}
			dirents = append(dirents, de)
		} else {
			info, err := f.lstatat(string(name))
			if IsNotExist(err) {
				// File disappeared between readdir + stat.
				// Treat as if it didn't exist.
				continue
			}
			if err != nil {
				return nil, nil, infos, err
			}
			infos = append(infos, info)
		}
		runtime.KeepAlive(f)
	}

	if n > 0 && len(names)+len(dirents)+len(infos) == 0 {
		return nil, nil, nil, io.EOF
	}
	return names, dirents, infos, nil
}

// Implemented in syscall/zsyscall_redox_amd64.go.

//go:linkname sysClosedir syscall.Closedir
func sysClosedir(dir uintptr) (err error)

//go:linkname sysReaddir syscall.Readdir
func sysReaddir(dir uintptr) (*syscall.Dirent, error)
