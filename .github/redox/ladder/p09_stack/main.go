package main

import (
	"fmt"
	"sync"
)

//go:noinline
func rec(n int) int {
	var pad [64]byte
	pad[n%64] = byte(n)
	if n == 0 {
		return int(pad[0])
	}
	return rec(n-1) + int(pad[n%64]&1)
}

func main() {
	// deep recursion forces repeated stack growth (copy) on the main goroutine
	r := rec(60000)
	fmt.Println("rec(60000) =", r)
	var wg sync.WaitGroup
	res := make([]int, 16)
	for i := range res {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			res[i] = rec(5000 + i)
		}(i)
	}
	wg.Wait()
	sum := 0
	for _, v := range res {
		sum += v
	}
	fmt.Println("goroutine recursion sum:", sum)
	fmt.Println("OK p09_stack")
}
