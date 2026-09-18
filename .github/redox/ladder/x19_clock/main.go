// Diagnostic for p11's "Sleep(20ms) returned after 17.7ms": is it the runtime's
// timers or a clock that disagrees between CPUs/threads?
package main

import (
	"fmt"
	"runtime"
	"sync/atomic"
	"time"
)

func main() {
	// 1. sleeps: elapsed must never be shorter than requested
	for _, procs := range []int{0, 1} {
		if procs > 0 {
			runtime.GOMAXPROCS(procs)
		}
		short, n := 0, 100
		var minEl, maxEl time.Duration = time.Hour, 0
		for i := 0; i < n; i++ {
			t0 := time.Now()
			time.Sleep(5 * time.Millisecond)
			el := time.Since(t0)
			if el < 5*time.Millisecond {
				short++
			}
			if el < minEl {
				minEl = el
			}
			if el > maxEl {
				maxEl = el
			}
		}
		fmt.Printf("sleep(5ms) x%d GOMAXPROCS=%d: %d too short; min %v max %v\n", n, runtime.GOMAXPROCS(0), short, minEl, maxEl)
	}
	// 2. does the monotonic clock go backwards across threads? two locked threads ping-pong timestamps
	var shared atomic.Int64
	done := make(chan struct{})
	neg, maxNeg := 0, int64(0)
	go func() {
		runtime.LockOSThread()
		last := int64(0)
		for i := 0; i < 200000; i++ {
			for shared.Load() == last {
			}
			last = shared.Load()
			now := int64(time.Since(start))
			if now < last {
				neg++
				if last-now > maxNeg {
					maxNeg = last - now
				}
			}
		}
		close(done)
	}()
	runtime.LockOSThread()
	for i := 1; i <= 200000; i++ {
		shared.Store(int64(time.Since(start)) + int64(i&0)) // ping
		select {
		case <-done:
			i = 1 << 30
		default:
		}
		runtime.Gosched()
	}
	fmt.Printf("cross-thread monotonic: %d backwards steps, worst %dns\n", neg, maxNeg)
	fmt.Println("OK x19_clock")
}

var start = time.Now()
