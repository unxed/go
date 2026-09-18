package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	var wg sync.WaitGroup
	var total atomic.Int64
	var mu sync.Mutex
	seen := map[int]bool{}
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				total.Add(1)
			}
			mu.Lock()
			seen[i] = true
			mu.Unlock()
		}(i)
	}
	wg.Wait()
	ch := make(chan int)
	done := make(chan struct{})
	go func() {
		sum := 0
		for v := range ch {
			sum += v
		}
		fmt.Println("channel sum:", sum)
		close(done)
	}()
	for i := 1; i <= 100; i++ {
		ch <- i
	}
	close(ch)
	<-done
	if total.Load() != 16000 || len(seen) != 16 {
		fmt.Println("FAIL: total", total.Load(), "seen", len(seen))
		return
	}
	fmt.Println("OK p03_goroutines")
}
