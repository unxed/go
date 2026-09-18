// Diagnostic rung: which open flags / openat use make fdopendir+readdir fail
// with EBADF on Redox? (os.RemoveAll's openat-based path hit that.) Always
// prints its OK line; the table above it is the payload.
package main

import (
	"fmt"
	"os"
	"syscall"
)

func listVia(label string, fd int, dup bool) {
	d := fd
	if dup {
		var err error
		d, err = syscall.Dup(fd)
		if err != nil {
			fmt.Printf("%-44s dup error: %v\n", label, err)
			return
		}
	}
	dir, err := syscall.Fdopendir(d)
	if err != nil {
		fmt.Printf("%-44s fdopendir error: %v\n", label, err)
		return
	}
	n := 0
	for {
		ent, err := syscall.Readdir(dir)
		if err != nil {
			fmt.Printf("%-44s readdir error after %d entries: %v\n", label, n, err)
			syscall.Closedir(dir)
			return
		}
		if ent == nil {
			break
		}
		n++
	}
	syscall.Closedir(dir)
	fmt.Printf("%-44s ok, %d entries\n", label, n)
}

func main() {
	os.RemoveAll("/tmp/p07")
	os.MkdirAll("/tmp/p07/sub", 0o755)
	os.WriteFile("/tmp/p07/f", []byte("x"), 0o644)
	rd := syscall.O_RDONLY | syscall.O_CLOEXEC
	combos := []struct {
		name  string
		flags int
	}{
		{"O_RDONLY|CLOEXEC", rd},
		{"+O_DIRECTORY", rd | syscall.O_DIRECTORY},
		{"+O_NOFOLLOW", rd | syscall.O_NOFOLLOW},
		{"+O_DIRECTORY|O_NOFOLLOW", rd | syscall.O_DIRECTORY | syscall.O_NOFOLLOW},
	}
	for _, c := range combos {
		fd, err := syscall.Open("/tmp/p07", c.flags, 0)
		if err != nil {
			fmt.Printf("open abs %-36s error: %v\n", c.name, err)
			continue
		}
		listVia("open abs "+c.name+" (dup)", fd, true)
		syscall.Close(fd)
		fd, _ = syscall.Open("/tmp/p07", c.flags, 0)
		listVia("open abs "+c.name+" (no dup)", fd, false)
	}
	parent, err := syscall.Open("/tmp", rd, 0)
	if err != nil {
		fmt.Println("open /tmp:", err)
		return
	}
	for _, c := range combos {
		fd, err := syscall.Openat(parent, "p07", c.flags, 0)
		if err != nil {
			fmt.Printf("openat %-39s error: %v\n", c.name, err)
			continue
		}
		listVia("openat "+c.name+" (dup)", fd, true)
		syscall.Close(fd)
		fd, _ = syscall.Openat(parent, "p07", c.flags, 0)
		listVia("openat "+c.name+" (no dup)", fd, false)
	}
	// The same through os.File, as RemoveAll's dirFile does.
	fd, err := syscall.Openat(parent, "p07", rd|syscall.O_DIRECTORY|syscall.O_NOFOLLOW, 0)
	if err == nil {
		f := os.NewFile(uintptr(fd), "p07")
		names, err := f.Readdirnames(-1)
		fmt.Println("os.NewFile(openat dir).Readdirnames:", names, err)
		f.Close()
	}
	os.RemoveAll("/tmp/p07")
	fmt.Println("OK p07_openat_diag")
}
