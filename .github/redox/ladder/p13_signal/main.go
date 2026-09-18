// os/signal: Notify + a self-sent SIGUSR1 (and SIGUSR2), Ignore/Reset, Stop.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ok := true
	ch := make(chan os.Signal, 4)
	signal.Notify(ch, syscall.SIGUSR1, syscall.SIGUSR2)
	for _, sig := range []syscall.Signal{syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGUSR1} {
		if err := syscall.Kill(os.Getpid(), sig); err != nil {
			fmt.Println("FAIL kill:", sig, err)
			return
		}
		select {
		case got := <-ch:
			fmt.Println("got signal:", got)
			if got != sig {
				ok = false
				fmt.Println("FAIL wrong signal, want", sig)
			}
		case <-time.After(3 * time.Second):
			fmt.Println("FAIL timeout waiting for", sig)
			return
		}
	}
	signal.Stop(ch)
	signal.Ignore(syscall.SIGUSR1)
	if err := syscall.Kill(os.Getpid(), syscall.SIGUSR1); err != nil {
		fmt.Println("FAIL kill ignored:", err)
		ok = false
	}
	time.Sleep(50 * time.Millisecond) // process must survive an ignored signal
	fmt.Println("survived ignored SIGUSR1")
	if ok {
		fmt.Println("OK p13_signal")
	}
}
