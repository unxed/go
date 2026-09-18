#!/usr/bin/env python3
"""Summarise the validation ladder: which programs built, which ran on Redox.

Reads /tmp/ladder-build.txt ("<name> BUILD_OK|BUILD_FAIL" per line) and
/tmp/redoxer-exec.log (output of run_ladder.sh inside the Redox VM). Every
program is run several times under three variants:

    default     plain
    nopreempt   GODEBUG=asyncpreemptoff=1   (isolates the signal/preemption path)
    procs1      GOMAXPROCS=1                (isolates multi-thread effects)

Each run is bracketed by "=== BEGIN <path> variant=V run=N" and
"=== END <path> variant=V run=N exit=C". A run passes when its output has the
line "OK <name>" and no watchdog kill. The job passes only when every
`default` run of every program passed; the other variants are diagnostics.

Writes a markdown table to $GITHUB_STEP_SUMMARY (if set) and stdout.
"""
import os
import re
import sys
from collections import defaultdict

build = {}
for line in open("/tmp/ladder-build.txt"):
    parts = line.split()
    if len(parts) == 2:
        build[parts[0]] = parts[1]

try:
    log = open("/tmp/redoxer-exec.log", errors="replace").read().replace("\r", "")
except FileNotFoundError:
    log = ""

runs = defaultdict(list)  # (name, variant) -> [(ok, code, out)]
for m in re.finditer(
    r"=== BEGIN \S*?/(\w+)\.bin variant=(\w+) run=(\d+)\n(.*?)=== END \S+ variant=\w+ run=\d+ exit=(\S+)",
    log,
    re.S,
):
    name, variant, out, code = m.group(1), m.group(2), m.group(4), m.group(5)
    ok = (
        re.search(r"^OK %s$" % re.escape(name), out, re.M) is not None
        and not re.search(r"^HANG ", out, re.M)
    )
    runs[(name, variant)].append((ok, code, out))

VARIANTS = ["default", "nopreempt", "procs1"]
rows, failed = [], False
for name in sorted(build):
    if build[name] != "BUILD_OK":
        rows.append([name, "build FAILED", "-", "-", "-", ""])
        failed = True
        continue
    cells, tail = [], ""
    for v in VARIANTS:
        rs = runs.get((name, v), [])
        cells.append(f"{sum(1 for r in rs if r[0])}/{len(rs)}" if rs else "no run")
        for ok, code, out in rs:
            if not ok and not tail:
                tail = f"[{v}] " + " / ".join(
                    [l for l in out.strip().split("\n") if l and "getrlimit" not in l][:3]
                )
        if v == "default":
            if not rs or not all(r[0] for r in rs):
                failed = True
    if not tail:
        rs = runs.get((name, "default"), [])
        if rs:
            tail = " / ".join([l for l in rs[-1][2].strip().split("\n") if l and "getrlimit" not in l][-2:])
    rows.append([name, "built"] + cells + [tail])

md = [
    "| program | build | default | nopreempt | procs1 | first failing output / last output |",
    "|---|---|---|---|---|---|",
]
for r in rows:
    md.append("| " + " | ".join(str(x).replace("|", "\\|")[:300] for x in r) + " |")
text = "\n".join(md)
print(text)
summ = os.environ.get("GITHUB_STEP_SUMMARY")
if summ:
    with open(summ, "a") as f:
        f.write("\n### Validation ladder (passes/runs per variant)\n\n" + text + "\n")
sys.exit(1 if failed else 0)
