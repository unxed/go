package main

import (
	"fmt"
	"runtime"
)

type node struct {
	next *node
	data [64]byte
}

func main() {
	var keep *node
	for round := 0; round < 200; round++ {
		var head *node
		for i := 0; i < 5000; i++ {
			head = &node{next: head}
		}
		if round%50 == 0 {
			keep = head
		}
	}
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	n := 0
	for p := keep; p != nil; p = p.next {
		n++
	}
	fmt.Println("NumGC:", m.NumGC, "HeapAlloc:", m.HeapAlloc, "kept nodes:", n)
	if m.NumGC == 0 || n != 5000 {
		fmt.Println("FAIL")
		return
	}
	fmt.Println("OK p04_gc")
}
