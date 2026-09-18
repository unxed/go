// Pure-Go stdlib coverage: strings, sort, reflect, encoding/json, sync/atomic,
// math/rand, runtime.NumCPU/GOMAXPROCS, sync primitives.
package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

type rec struct {
	Name string             `json:"name"`
	N    int                `json:"n"`
	Tags []string           `json:"tags,omitempty"`
	M    map[string]float64 `json:"m"`
}

func fail(f string, a ...any) { fmt.Printf("FAIL "+f+"\n", a...) }

func main() {
	ok := true
	check := func(c bool, f string, a ...any) {
		if !c {
			ok = false
			fail(f, a...)
		}
	}

	// strings / sort
	s := strings.Fields("  the quick  brown fox jumps over the lazy dog ")
	sort.Strings(s)
	check(strings.Join(s, ",") == "brown,dog,fox,jumps,lazy,over,quick,the,the", "sort/join: %v", s)
	check(strings.ToUpper("redox") == "REDOX" && strings.Repeat("ab", 3) == "ababab", "strings")
	var sb strings.Builder
	for i := 0; i < 1000; i++ {
		fmt.Fprintf(&sb, "%d,", i)
	}
	check(strings.Count(sb.String(), ",") == 1000, "builder")
	xs := rand.New(rand.NewSource(1)).Perm(50)
	sort.Ints(xs)
	for i, x := range xs {
		check(i == x, "perm %d != %d", i, x)
	}

	// reflect
	r := rec{Name: "a", N: 3, Tags: []string{"x", "y"}, M: map[string]float64{"pi": 3.25}}
	rt := reflect.TypeOf(r)
	check(rt.NumField() == 4 && rt.Field(1).Tag.Get("json") == "n", "reflect type")
	rv := reflect.ValueOf(&r).Elem()
	rv.Field(1).SetInt(42)
	check(r.N == 42, "reflect set")
	check(reflect.DeepEqual(r, r) && !reflect.DeepEqual(r, rec{}), "deepequal")
	fn := reflect.ValueOf(strings.ToUpper)
	check(fn.Call([]reflect.Value{reflect.ValueOf("x")})[0].String() == "X", "reflect call")

	// encoding/json
	b, err := json.Marshal(r)
	check(err == nil && string(b) == `{"name":"a","n":42,"tags":["x","y"],"m":{"pi":3.25}}`, "marshal: %v %s", err, b)
	var back rec
	check(json.Unmarshal(b, &back) == nil && reflect.DeepEqual(back, r), "unmarshal: %+v", back)
	var anyv map[string]any
	check(json.Unmarshal([]byte(`{"a":[1,2,{"b":null}],"c":"d"}`), &anyv) == nil && len(anyv) == 2, "json any")

	// sync/atomic + goroutines + mutex
	var cnt atomic.Int64
	var mu sync.Mutex
	total := 0
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 5000; i++ {
				cnt.Add(1)
				mu.Lock()
				total++
				mu.Unlock()
				if i%1000 == 0 {
					runtime.Gosched()
				}
			}
		}()
	}
	wg.Wait()
	check(cnt.Load() == 40000 && total == 40000, "atomic %d mutex %d", cnt.Load(), total)
	var once sync.Once
	n := 0
	for i := 0; i < 3; i++ {
		once.Do(func() { n++ })
	}
	check(n == 1, "once")
	var cv int32
	check(atomic.CompareAndSwapInt32(&cv, 0, 7) && atomic.LoadInt32(&cv) == 7, "cas")

	// math/rand
	rr := rand.New(rand.NewSource(42))
	a1, a2 := rr.Int63(), rand.New(rand.NewSource(42)).Int63()
	check(a1 == a2, "seeded rand")
	check(rand.Intn(10) < 10, "global rand")

	// runtime
	ncpu, procs := runtime.NumCPU(), runtime.GOMAXPROCS(0)
	fmt.Println("NumCPU:", ncpu, "GOMAXPROCS:", procs, "GOOS/GOARCH:", runtime.GOOS, runtime.GOARCH, runtime.Version())
	check(ncpu >= 1 && procs >= 1, "ncpu/procs")
	prev := runtime.GOMAXPROCS(2)
	check(runtime.GOMAXPROCS(0) == 2, "set procs (prev %d)", prev)
	runtime.GOMAXPROCS(prev)

	if ok {
		fmt.Println("OK p10_stdlib")
	}
}
