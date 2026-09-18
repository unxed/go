// os/exec: run a child, capture output, exit status, pipes, and re-exec self.
package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "child" {
		fmt.Println("child says hi, args:", strings.Join(os.Args[2:], ","))
		os.Exit(7)
	}
	ok := true
	check := func(c bool, f string, a ...any) {
		if !c {
			ok = false
			fmt.Printf("FAIL "+f+"\n", a...)
		}
	}
	stage := func(s string) { fmt.Println("stage:", s) }

	stage("self re-exec")
	// self re-exec: exit status + captured stdout, no dependency on the guest's userland
	self := os.Args[0]
	out, err := exec.Command(self, "child", "a", "b").Output()
	ee, isExit := err.(*exec.ExitError)
	check(isExit && ee.ExitCode() == 7, "self exec exit code: %v", err)
	check(strings.TrimSpace(string(out)) == "child says hi, args: a,b", "self exec output %q", out)

	stage("sh -c")
	// guest shell
	out, err = exec.Command("sh", "-c", "echo from-sh; exit 0").CombinedOutput()
	check(err == nil && strings.TrimSpace(string(out)) == "from-sh", "sh: %v %q", err, out)

	stage("cat with pipes")
	// stdin pipe -> child -> stdout pipe
	cmd := exec.Command("cat")
	cmd.Stdin = strings.NewReader("piped through cat\n")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	err = cmd.Run()
	check(err == nil && buf.String() == "piped through cat\n", "cat: %v %q", err, buf.String())

	stage("lookup/start failures")
	// lookup failure and start failure
	_, err = exec.LookPath("definitely-not-a-program")
	check(err != nil, "LookPath should fail")
	err = exec.Command("/nonexistent/prog").Run()
	check(err != nil, "running a nonexistent path should fail")

	stage("parallel children")
	// parallel children
	done := make(chan string, 4)
	for i := 0; i < 4; i++ {
		go func(i int) {
			o, e := exec.Command("sh", "-c", fmt.Sprintf("echo n%d", i)).Output()
			if e != nil {
				done <- "ERR " + e.Error()
				return
			}
			done <- strings.TrimSpace(string(o))
		}(i)
	}
	seen := map[string]bool{}
	for i := 0; i < 4; i++ {
		seen[<-done] = true
	}
	check(len(seen) == 4 && seen["n0"] && seen["n3"], "parallel children: %v", seen)

	if ok {
		fmt.Println("OK p14_exec")
	}
}
