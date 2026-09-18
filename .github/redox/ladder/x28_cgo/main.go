// cgo: C code compiled by redoxer's gcc, linked with relibc; goroutines call
// into C concurrently (threads are created through runtime/cgo).
package main

/*
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

static int add(int a, int b) { return a + b; }
static long slow_sum(int n) { long s = 0; for (int i = 0; i < n; i++) s += i; return s; }
*/
import "C"

import (
	"fmt"
	"sync"
	"unsafe"
)

func main() {
	if got := C.add(2, 3); got != 5 {
		fmt.Println("FAIL add:", got)
		return
	}
	s := C.CString("redox")
	n := C.strlen(s)
	C.free(unsafe.Pointer(s))
	if n != 5 {
		fmt.Println("FAIL strlen:", n)
		return
	}
	fmt.Println("getpid from C:", C.getpid())
	var wg sync.WaitGroup
	res := make([]int64, 8)
	for i := range res {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res[i] = int64(C.slow_sum(C.int(1000 * (i + 1))))
		}()
	}
	wg.Wait()
	for i, r := range res {
		n := int64(1000 * (i + 1))
		if want := n * (n - 1) / 2; r != want {
			fmt.Println("FAIL slow_sum", i, r, want)
			return
		}
	}
	fmt.Println("OK x28_cgo")
}
