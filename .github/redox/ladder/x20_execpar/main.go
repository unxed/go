// os/exec concurrency bisect (p14 hangs in "parallel children"). Each stage is
// announced first, so the last line before HANG names the culprit.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

func par(n int, name string, f func(i int) error) {
	fmt.Printf("stage: %s (x%d parallel)\n", name, n)
	var wg sync.WaitGroup
	errs := make([]error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) { defer wg.Done(); errs[i] = f(i) }(i)
	}
	wg.Wait()
	for i, e := range errs {
		if e != nil {
			fmt.Printf("FAIL %s[%d]: %v\n", name, i, e)
		}
	}
}

func main() {
	fmt.Println("stage: 4 sequential Output()")
	for i := 0; i < 4; i++ {
		o, err := exec.Command("sh", "-c", fmt.Sprintf("echo s%d", i)).Output()
		if err != nil || strings.TrimSpace(string(o)) != fmt.Sprintf("s%d", i) {
			fmt.Println("FAIL sequential", i, err, string(o))
		}
	}
	par(2, "Run() no stdio", func(i int) error { return exec.Command("sh", "-c", "exit 0").Run() })
	par(4, "Run() no stdio", func(i int) error { return exec.Command("sh", "-c", "exit 0").Run() })
	par(2, "Output() pipes", func(i int) error {
		_, err := exec.Command("sh", "-c", "echo x").Output()
		return err
	})
	par(4, "Output() pipes", func(i int) error {
		_, err := exec.Command("sh", "-c", "echo x").Output()
		return err
	})
	fmt.Println("OK x20_execpar")
	os.Exit(0)
}
