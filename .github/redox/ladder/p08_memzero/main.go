package main

import (
	"fmt"
	"runtime"
)

// Fresh heap memory must be zero. Allocate lots of big and small objects,
// dirty them, free them (GC), and re-allocate, checking for non-zero bytes.
func main() {
	bad := 0
	for round := 0; round < 2; round++ {
		var keep [][]byte
		for i := 0; i < 12; i++ {
			b := make([]byte, 1<<20)
			for j := range b {
				if b[j] != 0 {
					bad++
					break
				}
			}
			for j := range b {
				b[j] = 0xAA
			}
			if i%6 == 0 {
				keep = append(keep, b)
			}
		}
		type small struct{ a, b, c, d uint64 }
		var ss []*small
		for i := 0; i < 20000; i++ {
			s := &small{}
			if s.a|s.b|s.c|s.d != 0 {
				bad++
			}
			s.a, s.b, s.c, s.d = ^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)
			if i%3 == 0 {
				ss = append(ss, s)
			}
		}
		runtime.GC()
		_ = keep
		_ = ss
	}
	fmt.Println("non-zero fresh allocations:", bad)
	if bad != 0 {
		fmt.Println("FAIL")
		return
	}
	fmt.Println("OK p08_memzero")
}
