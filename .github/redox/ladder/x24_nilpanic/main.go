// Synchronous faults become Go panics: nil dereference (Redox never delivers page faults as signals, so the compiler emits explicit nil checks) must be recoverable, repeatedly, on any thread.
// (time.TestIssue5745 hangs otherwise; a spin with "goroutine running on other
// thread" points at the fault being retried forever.)
package main

import (
	"fmt"
	"runtime"
	"sync"
)

func try(f func()) (r any) {
	defer func() { r = recover() }()
	f()
	return nil
}

type big struct {
	pad [1 << 16]byte
	x   int
}

func main() {
	fmt.Println("stage: nil pointer dereference")
	for i := 0; i < 3; i++ {
		r := try(func() { var p *int; _ = *p })
		e, ok := r.(runtime.Error)
		if !ok {
			fmt.Printf("FAIL round %d: recovered %v (%T), want runtime.Error\n", i, r, r)
			return
		}
		if i == 0 {
			fmt.Println("recovered:", e)
		}
	}
	fmt.Println("stage: nil deref with a large offset")
	if r := try(func() { var p *big; _ = p.x }); r == nil {
		fmt.Println("FAIL: no panic for large-offset nil deref")
		return
	}
	fmt.Println("stage: concurrent nil derefs on several threads")
	var wg sync.WaitGroup
	bad := make(chan string, 8)
	for g := 0; g < 4; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			runtime.LockOSThread()
			for i := 0; i < 20; i++ {
				if try(func() { var m *map[string]int; _ = (*m)["a"] }) == nil {
					bad <- "no panic"
					return
				}
			}
		}()
	}
	wg.Wait()
	close(bad)
	for b := range bad {
		fmt.Println("FAIL:", b)
		return
	}
	fmt.Println("OK x24_nilpanic")
}
