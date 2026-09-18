// Exit-hang probe: several Ms exist (sleepers, a blocked goroutine, timers),
// main prints and returns. Run many times under a watchdog; a run that printed
// its OK line and never exited is the "exit hang".
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			time.Sleep(time.Duration(i) * time.Millisecond)
			_ = fmt.Sprint(i)
		}(i)
	}
	wg.Wait()
	go func() { time.Sleep(time.Hour) }()
	go func() { select {} }()
	t := time.NewTicker(time.Millisecond)
	defer t.Stop()
	<-t.C
	fmt.Println("OK x11_exitloop")
}
