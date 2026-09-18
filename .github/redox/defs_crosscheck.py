#!/usr/bin/env python3
"""Mechanically cross-check the hand-derived Redox runtime definitions.

Runs the real `go tool cgo -godefs src/runtime/defs_redox.go` with $CC set to a
wrapper that execs redoxer's cross compiler (so the numbers come from relibc's
actual generated C headers), then compares against the hand-derived
definitions in defs_redox_amd64.go (plus the constants that live in
os_redox.go / os2_redox.go / netpoll_redox.go) using a small Go program:

  * every constant listed in the probe: value equality
  * every struct/union/typedef pair: size, alignment and (for structs) the
    ordered offsets+sizes of the non-padding fields

Any C name that relibc's headers don't provide is dropped from the probe and
reported, so one missing macro doesn't hide all other results.

Exit status 1 if any MISMATCH is found. Run from the repo root.
"""
import os
import re
import shutil
import subprocess
import sys

RT = "src/runtime"
WORK = "/tmp/crosscheck"

TYPE_PAIRS = [
    ("SemT", "semt"),
    ("Sigset", "sigset"),
    ("StackT", "stackt"),
    ("Siginfo", "siginfo"),
    ("SigactionT", "sigactiont"),
    ("Mcontext", "mcontext"),
    ("Ucontext", "ucontext"),
    ("Timespec", "timespec"),
    ("Timeval", "timeval"),
    ("Itimerval", "itimerval"),
    ("Pthread", "pthread"),
    ("PthreadAttr", "pthreadattr"),
]

# Files whose `_NAME = <int>` constants should be visible to the checker as
# "hand" values (they are not in defs_redox_amd64.go).
EXTRA_HAND_FILES = ["os_redox.go", "os2_redox.go", "netpoll_redox.go"]


def sh(cmd, **kw):
    return subprocess.run(cmd, capture_output=True, text=True, **kw)


def main():
    shutil.rmtree(WORK, ignore_errors=True)
    os.makedirs(WORK)

    src = open(f"{RT}/defs_redox.go").read().split("\n")
    removed = []
    env = dict(os.environ, GOARCH="amd64", CGO_ENABLED="1")
    r = None
    for attempt in range(10):
        open(f"{WORK}/defs_redox.go", "w").write("\n".join(src))
        r = sh(["go", "tool", "cgo", "-godefs", "defs_redox.go"], cwd=WORK, env=env)
        if r.returncode == 0:
            break
        print(f"--- cgo attempt {attempt + 1} failed ---\n{r.stderr}")
        names = set(
            re.findall(r"could not determine (?:what|kind of name for) C\.(\w+)", r.stderr)
        )
        if not names:
            print("cgo failed for a reason other than unresolved C names; giving up")
            return 2
        removed += sorted(names)
        pat = re.compile(r"\bC\.(%s)\b" % "|".join(map(re.escape, names)))
        src = [l for l in src if not pat.search(l)]
    else:
        print("too many cgo attempts")
        return 2

    print("=== C names not provided by relibc's headers (dropped from probe) ===")
    print("\n".join(removed) if removed else "(none)")

    cg = r.stdout.replace("package runtime", "package main", 1)
    open(f"{WORK}/cg.go", "w").write(cg)

    hand = open(f"{RT}/defs_redox_amd64.go").read().replace("package runtime", "package main", 1)
    open(f"{WORK}/hand.go", "w").write(hand)

    extra = ["package main", ""]
    seen = set(re.findall(r"^\s*(_\w+)\s*=", hand, re.M))
    for f in EXTRA_HAND_FILES:
        text = open(f"{RT}/{f}").read()
        for name, val in re.findall(
            r"^\s*(?:const\s+)?(_\w+)\s*=\s*(0x[0-9a-fA-F_]+|\d+)\b", text, re.M
        ):
            if name not in seen:
                seen.add(name)
                extra.append(f"const {name} = {val}")
    open(f"{WORK}/handextra.go", "w").write("\n".join(extra) + "\n")

    hand_names = seen
    consts, not_in_hand = [], []
    for line in src:
        m = re.match(r"^\s+(\w+)\s+= C\.(\w+)\s*$", line)
        if not m:
            continue
        n = m.group(1)
        if "_" + n in hand_names:
            consts.append(n)
        else:
            not_in_hand.append(n)

    types = []
    for c, h in TYPE_PAIRS:
        if any(re.search(r"^type %s C\." % c, l) for l in src):
            types.append((c, h))

    rows = ",\n\t".join(f'{{"{n}", int64({n}), int64(_{n})}}' for n in consts)
    trows = "\n\t".join(
        f'cmpType("{c}", reflect.TypeOf((*{c})(nil)).Elem(), reflect.TypeOf((*{h})(nil)).Elem())'
        for c, h in types
    )
    main_go = f'''package main

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

var bad int

type fld struct {{
	n         string
	off, size uintptr
}}

func skip(n string) bool {{
	l := strings.ToLower(n)
	return strings.HasPrefix(l, "pad") || strings.HasPrefix(l, "_pad") || strings.HasPrefix(l, "x_pad") || n == "_"
}}

func fields(t reflect.Type) []fld {{
	var out []fld
	for i := 0; i < t.NumField(); i++ {{
		f := t.Field(i)
		if skip(f.Name) {{
			continue
		}}
		out = append(out, fld{{f.Name, f.Offset, f.Type.Size()}})
	}}
	return out
}}

func cmpType(name string, c, h reflect.Type) {{
	fmt.Printf("type %s: cgo size=%d align=%d | hand size=%d align=%d\\n", name, c.Size(), c.Align(), h.Size(), h.Align())
	if c.Size() != h.Size() {{
		fmt.Printf("MISMATCH %s: size cgo=%d hand=%d\\n", name, c.Size(), h.Size())
		bad++
	}}
	// cgo renders unions as byte arrays (align 1), so only compare alignment
	// when both sides are real structs.
	if c.Kind() == reflect.Struct && h.Kind() == reflect.Struct && c.Align() != h.Align() {{
		fmt.Printf("MISMATCH %s: align cgo=%d hand=%d\\n", name, c.Align(), h.Align())
		bad++
	}}
	if c.Kind() != reflect.Struct || h.Kind() != reflect.Struct {{
		return
	}}
	cf, hf := fields(c), fields(h)
	for _, f := range cf {{
		fmt.Printf("    cgo  %-16s off=%-4d size=%d\\n", f.n, f.off, f.size)
	}}
	for _, f := range hf {{
		fmt.Printf("    hand %-16s off=%-4d size=%d\\n", f.n, f.off, f.size)
	}}
	if len(cf) != len(hf) {{
		fmt.Printf("MISMATCH %s: %d non-padding fields in cgo vs %d in hand\\n", name, len(cf), len(hf))
		bad++
		return
	}}
	for i := range cf {{
		if cf[i].off != hf[i].off || cf[i].size != hf[i].size {{
			fmt.Printf("MISMATCH %s field #%d (%s vs %s): cgo off=%d size=%d, hand off=%d size=%d\\n",
				name, i, cf[i].n, hf[i].n, cf[i].off, cf[i].size, hf[i].off, hf[i].size)
			bad++
		}}
	}}
}}

var consts = []struct {{
	n    string
	c, h int64
}}{{
	{rows},
}}

func main() {{
	{trows}
	for _, k := range consts {{
		if k.c != k.h {{
			fmt.Printf("MISMATCH const %s: cgo=%#x hand=%#x\\n", k.n, k.c, k.h)
			bad++
		}} else {{
			fmt.Printf("ok const %s = %#x\\n", k.n, k.c)
		}}
	}}
	if bad != 0 {{
		fmt.Printf("\\n%d MISMATCH(ES)\\n", bad)
		os.Exit(1)
	}}
	fmt.Println("\\nall checked definitions match")
}}
'''
    open(f"{WORK}/main.go", "w").write(main_go)
    open(f"{WORK}/go.mod", "w").write("module crosscheck\n\ngo 1.24\n")

    print("=== constants in the probe with no hand-derived counterpart (informational) ===")
    print("\n".join(not_in_hand) if not_in_hand else "(none)")

    print("=== running comparison ===")
    r = sh(["go", "run", "."], cwd=WORK, env=dict(os.environ, GOOS="linux", GOARCH="amd64", CGO_ENABLED="0"))
    sys.stdout.write(r.stdout)
    sys.stderr.write(r.stderr)
    return r.returncode


if __name__ == "__main__":
    sys.exit(main())
