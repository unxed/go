// time zones (embedded tzdata + system lookup), time formatting, crypto/rand.
package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"time"
	_ "time/tzdata"
)

func main() {
	ok := true
	check := func(c bool, f string, a ...any) {
		if !c {
			ok = false
			fmt.Printf("FAIL "+f+"\n", a...)
		}
	}

	// crypto/rand
	a, b := make([]byte, 32), make([]byte, 32)
	n1, e1 := rand.Read(a)
	n2, e2 := rand.Read(b)
	check(e1 == nil && e2 == nil && n1 == 32 && n2 == 32, "rand.Read: %v %v", e1, e2)
	check(!bytes.Equal(a, b) && !bytes.Equal(a, make([]byte, 32)), "rand output not random: %x %x", a, b)
	fmt.Println("crypto/rand sample:", hex.EncodeToString(a[:8]))
	sum := sha256.Sum256([]byte("redox"))
	check(hex.EncodeToString(sum[:4]) != "", "sha256")

	// time zones: embedded database (time/tzdata) must always work
	loc, err := time.LoadLocation("America/New_York")
	check(err == nil, "LoadLocation(America/New_York): %v", err)
	if err == nil {
		t := time.Date(2024, 7, 1, 12, 0, 0, 0, time.UTC).In(loc)
		name, off := t.Zone()
		check(name == "EDT" && off == -4*3600, "NY summer zone: %s %d", name, off)
		w := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC).In(loc)
		name, off = w.Zone()
		check(name == "EST" && off == -5*3600, "NY winter zone: %s %d", name, off)
	}
	tk, err := time.LoadLocation("Asia/Tokyo")
	check(err == nil, "LoadLocation(Asia/Tokyo): %v", err)
	if err == nil {
		check(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).In(tk).Format("15:04 MST") == "09:00 JST", "Tokyo format")
	}
	u, err := time.LoadLocation("UTC")
	check(err == nil && u == time.UTC, "LoadLocation(UTC)")

	// what does the system itself provide? (informational)
	_, e := os.Stat("/usr/share/zoneinfo/UTC")
	fmt.Println("system zoneinfo present:", e == nil, "| time.Local =", time.Local, "| TZ =", os.Getenv("TZ"))
	now := time.Now()
	check(now.Year() >= 2024, "time.Now year %d", now.Year())
	fmt.Println("now:", now.Format(time.RFC3339), "monotonic delta ok:", time.Since(now) >= 0)

	// timers/format/parse round trip
	p, err := time.Parse(time.RFC1123Z, "Mon, 02 Jan 2006 15:04:05 -0700")
	check(err == nil && p.Unix() == 1136239445, "parse: %v %d", err, p.Unix())
	d := 1500 * time.Millisecond
	check(d.String() == "1.5s", "duration string")
	t0 := time.Now()
	time.Sleep(20 * time.Millisecond)
	el := time.Since(t0)
	check(el >= 20*time.Millisecond && el < 2*time.Second, "sleep elapsed %v", el)

	if ok {
		fmt.Println("OK p11_time_rand")
	}
}
