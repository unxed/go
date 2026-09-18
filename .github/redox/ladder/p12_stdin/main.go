// bufio over os.Stdin: the runner pipes "alpha\nbeta\ngamma\n" in.
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	sc := bufio.NewScanner(os.Stdin)
	var lines []string
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	if err := sc.Err(); err != nil {
		fmt.Println("FAIL scan:", err)
		return
	}
	got := strings.Join(lines, "|")
	fmt.Println("stdin lines:", got)
	if got != "alpha|beta|gamma" {
		fmt.Println("FAIL unexpected stdin content")
		return
	}
	w := bufio.NewWriter(os.Stdout)
	fmt.Fprintf(w, "buffered write %d\n", 42)
	w.Flush()
	fmt.Fprintln(os.Stderr, "stderr write works")
	fmt.Println("OK p12_stdin")
}
