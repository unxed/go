// pty pair through package os: reads with a pending Read, with deadlines, and
// alongside timers must work (ttys are not reliably pollable on Redox, see
// runtime/netpoll_redox.go). Master = /scheme/pty/ptmx, slave = /scheme/pty/N.
package main

import (
	"fmt"
	"os"
	"time"
)

func readWithin(f *os.File, d time.Duration) (string, error) {
	type res struct {
		s string
		e error
	}
	ch := make(chan res, 1)
	go func() {
		b := make([]byte, 64)
		n, err := f.Read(b)
		ch <- res{string(b[:n]), err}
	}()
	select {
	case r := <-ch:
		return r.s, r.e
	case <-time.After(d):
		return "", fmt.Errorf("timeout after %v", d)
	}
}

func main() {
	m, err := os.OpenFile("/scheme/pty/ptmx", os.O_RDWR, 0)
	if err != nil {
		fmt.Println("FAIL open ptmx:", err)
		return
	}
	// Find the slave: the stat carries no id on Redox, so probe ids from the top;
	// only the real slave echoes what the master writes.
	var s *os.File
	for id := 40; id >= 0 && s == nil; id-- {
		f, err := os.OpenFile(fmt.Sprintf("/scheme/pty/%d", id), os.O_RDWR, 0)
		if err != nil {
			continue
		}
		m.WriteString("probe\n")
		if got, err := readWithin(f, 400*time.Millisecond); err == nil && got == "probe\n" {
			s = f
			break
		}
		f.Close()
	}
	if s == nil {
		fmt.Println("FAIL: cannot find the slave of the master")
		return
	}
	fmt.Println("slave:", s.Name())
	ok := true
	check := func(what, want string) {
		got, err := readWithin(s, 3*time.Second)
		if err != nil || got != want {
			ok = false
			fmt.Printf("FAIL %s: got %q err %v want %q\n", what, got, err, want)
		} else {
			fmt.Printf("ok: %s\n", what)
		}
	}
	m.WriteString("hello\n")
	check("slave read after master write (data already there)", "hello\n")
	go func() { time.Sleep(300 * time.Millisecond); m.WriteString("later\n") }()
	check("slave read pending, data arrives 300 ms later", "later\n")
	go func() { time.Sleep(300 * time.Millisecond); s.WriteString("back\n") }()
	got, err := readWithin(m, 3*time.Second)
	if err != nil || got == "" {
		ok = false
		fmt.Printf("FAIL master read pending: got %q err %v\n", got, err)
	} else {
		fmt.Printf("ok: master read pending: %q\n", got)
	}
	// a blocked reader must not stall timers / other goroutines
	go func() { b := make([]byte, 8); s.Read(b) }()
	t0 := time.Now()
	time.Sleep(200 * time.Millisecond)
	if el := time.Since(t0); el > time.Second {
		ok = false
		fmt.Println("FAIL: timer starved by a blocked tty read:", el)
	}
	// deadline on a tty
	s.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
	b := make([]byte, 8)
	_, err = s.Read(b)
	fmt.Printf("read with deadline: err=%v\n", err)
	if ok {
		fmt.Println("OK x30_pty")
	}
}
