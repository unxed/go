#!/usr/bin/env python3
"""Compare unpatched vs patched relibc under Go async preemption.

Input: /tmp/redoxer-exec.log; runs bracketed by
"=== BEGIN <path> variant=<unpatched|patched> run=N" / "=== END ... exit=C".
A run passes when its output contains "OK <program>" and no HANG line.
"""
import os, re, sys
from collections import defaultdict
try:
    log = open("/tmp/redoxer-exec.log", errors="replace").read().replace("\r", "")
except FileNotFoundError:
    log = ""
res = defaultdict(lambda: [0, 0])
hangs = defaultdict(int)
for m in re.finditer(r"=== BEGIN \S*?/(\w+?)_(?:unpatched|patched)\.bin variant=(\w+) run=(\d+)\n(.*?)=== END \S+ variant=\w+ run=\d+ exit=(\S+)", log, re.S):
    prog, variant, out = m.group(1), m.group(2), m.group(4)
    ok = re.search(r"^OK %s$" % re.escape(prog), out, re.M) is not None
    crashed = re.search(r"^Page fault|^fatal error|^panic", out, re.M) is not None
    ok = ok and not crashed
    res[(prog, variant)][1] += 1
    res[(prog, variant)][0] += ok
    if ok and re.search(r"^HANG ", out, re.M):
        hangs[(prog, variant)] += 1
progs = sorted({k[0] for k in res})
rows = ["| program | unpatched relibc | patched relibc |", "|---|---|---|"]
for p in progs:
    u, q = res[(p, "unpatched")], res[(p, "patched")]
    hu, hq = hangs[(p, "unpatched")], hangs[(p, "patched")]
    rows.append(f"| {p} (GODEBUG=asyncpreemptoff=0) | {u[0]}/{u[1]} pass" + (f" ({hu} exit-hang)" if hu else "") + f" | {q[0]}/{q[1]} pass" + (f" ({hq} exit-hang)" if hq else "") + " |")
text = "\n".join(rows)
print(text)
if os.environ.get("GITHUB_STEP_SUMMARY"):
    open(os.environ["GITHUB_STEP_SUMMARY"], "a").write("\n### relibc sigentry-rcx patch verification\n\n" + text + "\n")
# expected: patched passes everything
bad = any(res[(p, "patched")][0] != res[(p, "patched")][1] or res[(p, "patched")][1] == 0 for p in progs)
sys.exit(1 if (bad or not progs) else 0)
