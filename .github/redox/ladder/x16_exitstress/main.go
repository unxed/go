// Go equivalent of x17: goroutine pairs ping-pong over unbuffered channels
// (Ms park on indefinite futexes), plus sleepers; main exits at a random moment.
package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	for i := 0; i < 6; i++ {
		a, b := make(chan int), make(chan int)
		go func() {
			for v := range a {
				b <- v + 1
			}
		}()
		go func() {
			for v := range b {
				a <- v + 1
			}
		}()
		a <- 0
	}
	for i := 0; i < 4; i++ {
		go func() {
			for {
				time.Sleep(time.Millisecond)
			}
		}()
	}
	time.Sleep(time.Duration(time.Now().UnixNano()>>7%15000) * time.Microsecond)
	fmt.Println("OK x16_exitstress")
	os.Exit(0)
}
