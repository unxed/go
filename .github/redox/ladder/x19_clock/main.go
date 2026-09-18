// Diagnostic for p11's one-off "Sleep(20ms) returned after 17.7ms": timers vs a
// clock that disagrees between threads.
package main

import (
	"fmt"
	"runtime"
	"sync/atomic"
	"time"
)

var start = time.Now()

func now() int64 { return int64(time.Since(start)) }

func main() {
	for _, procs := range []int{2, 1} {
		runtime.GOMAXPROCS(procs)
		short, n := 0, 100
		var minEl, maxEl time.Duration = time.Hour, 0
		for i := 0; i < n; i++ {
			t0 := time.Now()
			time.Sleep(5 * time.Millisecond)
			el := time.Since(t0)
			if el < 5*time.Millisecond {
				short++
			}
			minEl, maxEl = min(minEl, el), max(maxEl, el)
		}
		fmt.Printf("sleep(5ms) x%d GOMAXPROCS=%d: %d too short; min %v max %v\n", n, procs, short, minEl, maxEl)
	}
	runtime.GOMAXPROCS(2)
	// Two OS-thread-locked goroutines alternate strictly via a handshake; every
	// reading must be >= the previous one although they are taken on different threads.
	const N = 4000
	var seq, ack atomic.Int64
	var tsA, tsB atomic.Int64
	var negAB, negBA int
	var worstAB, worstBA int64
	done := make(chan struct{})
	go func() {
		runtime.LockOSThread()
		for i := int64(1); i <= N; i++ {
			for seq.Load() != i {
				runtime.Gosched()
			}
			t := now()
			if a := tsA.Load(); t < a {
				negAB++
				worstAB = max(worstAB, a-t)
			}
			tsB.Store(t)
			ack.Store(i)
		}
		close(done)
	}()
	runtime.LockOSThread()
	for i := int64(1); i <= N; i++ {
		tsA.Store(now())
		seq.Store(i)
		for ack.Load() != i {
			runtime.Gosched()
		}
		t := now()
		if b := tsB.Load(); t < b {
			negBA++
			worstBA = max(worstBA, b-t)
		}
	}
	<-done
	fmt.Printf("cross-thread monotonic (%d handshakes): A->B backwards %d (worst %dns), B->A backwards %d (worst %dns)\n", N, negAB, worstAB, negBA, worstBA)
	fmt.Println("OK x19_clock")
}
