// Signal-return integrity probe. Goroutines run a deterministic register-,
// FP- and stack-heavy computation while another goroutine bombards the
// process with SIGURG (the runtime's preemption signal; when no preemption
// is requested its handler just returns, which exercises handler entry and
// sigreturn without asyncPreempt injection). Every worker must produce
// exactly the expected value; any corruption/crash means the signal
// delivery/return path on Redox loses or clobbers state.
package main

import (
	"fmt"
	"math"
	"sync"
	"syscall"
	"time"
)

//go:noinline
func work(seed uint64) uint64 {
	x := seed
	f := float64(seed) + 0.5
	var buf [16]uint64
	for i := 0; i < 400000; i++ {
		x = x*6364136223846793005 + 1442695040888963407
		f = f*1.0000001 + float64(x&0xff)
		buf[i&15] ^= x
		if i&0xfff == 0 {
			f = math.Sqrt(f * f)
		}
	}
	var s uint64
	for _, v := range buf {
		s ^= v
	}
	return x ^ s ^ math.Float64bits(f)
}

func main() {
	want := work(1)
	fmt.Println("expected:", want)
	stop := make(chan struct{})
	var sent int
	go func() {
		pid := syscall.Getpid()
		for {
			select {
			case <-stop:
				return
			default:
			}
			syscall.Kill(pid, syscall.SIGURG)
			sent++
			time.Sleep(300 * time.Microsecond)
		}
	}()
	var wg sync.WaitGroup
	bad := 0
	var mu sync.Mutex
	for w := 0; w < 4; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for r := 0; r < 5; r++ {
				if got := work(1); got != want {
					mu.Lock()
					bad++
					mu.Unlock()
				}
			}
		}()
	}
	wg.Wait()
	close(stop)
	fmt.Println("signals sent:", sent, "mismatches:", bad)
	if bad != 0 {
		fmt.Println("FAIL")
		return
	}
	fmt.Println("OK x10_sigstorm")
}
