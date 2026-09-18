package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func main() {
	dir := "/tmp/goredox-p06"
	os.RemoveAll(dir)
	if err := os.MkdirAll(filepath.Join(dir, "sub"), 0o755); err != nil {
		fmt.Println("FAIL MkdirAll:", err)
		return
	}
	want := []byte("hello from redox file I/O\n")
	p := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(p, want, 0o644); err != nil {
		fmt.Println("FAIL WriteFile:", err)
		return
	}
	got, err := os.ReadFile(p)
	if err != nil || !bytes.Equal(got, want) {
		fmt.Println("FAIL ReadFile:", err, string(got))
		return
	}
	st, err := os.Stat(p)
	if err != nil || st.Size() != int64(len(want)) || st.IsDir() {
		fmt.Println("FAIL Stat:", err, st)
		return
	}
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b"), 0o644)
	ents, err := os.ReadDir(dir)
	if err != nil {
		fmt.Println("FAIL ReadDir:", err)
		return
	}
	var names []string
	for _, e := range ents {
		names = append(names, e.Name())
	}
	sort.Strings(names)
	fmt.Println("dir entries:", names)
	if fmt.Sprint(names) != "[a.txt b.txt sub]" {
		fmt.Println("FAIL ReadDir contents")
		return
	}
	if err := os.Rename(p, filepath.Join(dir, "c.txt")); err != nil {
		fmt.Println("FAIL Rename:", err)
		return
	}
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		fmt.Println("FAIL old name still exists:", err)
		return
	}
	if err := os.RemoveAll(dir); err != nil {
		fmt.Println("FAIL RemoveAll:", err)
		return
	}
	fmt.Println("OK p06_fileio")
}
