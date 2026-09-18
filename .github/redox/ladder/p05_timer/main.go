package main

import (
	"fmt"
	"time"
)

func main() {
	start := time.Now()
	time.Sleep(100 * time.Millisecond)
	el := time.Since(start)
	fmt.Println("slept", el)
	if el < 90*time.Millisecond || el > 5*time.Second {
		fmt.Println("FAIL: sleep elapsed", el)
		return
	}
	select {
	case <-time.After(50 * time.Millisecond):
		fmt.Println("time.After fired")
	case <-time.After(5 * time.Second):
		fmt.Println("FAIL: time.After too slow")
		return
	}
	tk := time.NewTicker(20 * time.Millisecond)
	n := 0
	for range tk.C {
		n++
		if n == 3 {
			tk.Stop()
			break
		}
	}
	fmt.Println("ticks:", n, "now:", time.Now().Year() >= 2024)
	fmt.Println("OK p05_timer")
}
