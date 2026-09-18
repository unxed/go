// net on loopback: TCP listen/dial echo, UDP packet round trip, resolver-free
// address parsing. Diagnostic (x*) until we know the guest has a working netstack.
package main

import (
	"fmt"
	"io"
	"net"
	"time"
)

func main() {
	ok := true
	check := func(c bool, f string, a ...any) {
		if !c {
			ok = false
			fmt.Printf("FAIL "+f+"\n", a...)
		}
	}
	ip := net.ParseIP("127.0.0.1")
	check(ip != nil && ip.To4() != nil, "ParseIP")

	// TCP
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Println("FAIL tcp Listen:", err)
		return
	}
	fmt.Println("tcp listening on", ln.Addr())
	go func() {
		c, err := ln.Accept()
		if err != nil {
			fmt.Println("FAIL Accept:", err)
			return
		}
		defer c.Close()
		io.Copy(c, c)
	}()
	c, err := net.DialTimeout("tcp", ln.Addr().String(), 5*time.Second)
	if err != nil {
		fmt.Println("FAIL tcp Dial:", err)
		return
	}
	c.SetDeadline(time.Now().Add(5 * time.Second))
	if _, err := c.Write([]byte("ping over tcp")); err != nil {
		fmt.Println("FAIL tcp Write:", err)
		return
	}
	buf := make([]byte, 64)
	n, err := io.ReadAtLeast(c, buf, len("ping over tcp"))
	check(err == nil && string(buf[:n]) == "ping over tcp", "tcp echo: %v %q", err, buf[:n])
	c.Close()
	ln.Close()
	fmt.Println("tcp done")

	// UDP
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		fmt.Println("FAIL udp ListenPacket:", err)
		return
	}
	defer pc.Close()
	uc, err := net.Dial("udp", pc.LocalAddr().String())
	if err != nil {
		fmt.Println("FAIL udp Dial:", err)
		return
	}
	defer uc.Close()
	uc.Write([]byte("ping over udp"))
	pc.SetReadDeadline(time.Now().Add(5 * time.Second))
	n, from, err := pc.ReadFrom(buf)
	check(err == nil && string(buf[:n]) == "ping over udp", "udp read: %v %q", err, buf[:n])
	if err == nil {
		pc.WriteTo([]byte("pong"), from)
		uc.SetReadDeadline(time.Now().Add(5 * time.Second))
		n, err = uc.Read(buf)
		check(err == nil && string(buf[:n]) == "pong", "udp reply: %v %q", err, buf[:n])
	}
	if ok {
		fmt.Println("OK x15_net")
	}
}
