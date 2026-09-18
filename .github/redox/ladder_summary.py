#!/usr/bin/env python3
"""Summarise the validation ladder: which programs built, which ran on Redox.

Reads /tmp/ladder-build.txt ("<name> BUILD_OK|BUILD_FAIL" per line) and
/tmp/redoxer-exec.log (output of run_ladder.sh inside the Redox VM: each
program's output is bracketed by "=== BEGIN <path>" / "=== END <path> exit=N").
A program passes when its output contains the line "OK <name>".

Writes a markdown table to $GITHUB_STEP_SUMMARY (if set) and stdout; exits 1
if any rung failed.
"""
import os
import re
import sys

build = {}
for line in open("/tmp/ladder-build.txt"):
    parts = line.split()
    if len(parts) == 2:
        build[parts[0]] = parts[1]

try:
    log = open("/tmp/redoxer-exec.log", errors="replace").read().replace("\r", "")
except FileNotFoundError:
    log = ""

blocks = {}
for m in re.finditer(r"=== BEGIN \S*?/(\w+)\.bin\n(.*?)=== END \S+ exit=(\d+)", log, re.S):
    blocks[m.group(1)] = (m.group(2), int(m.group(3)))

rows, failed = [], False
for name in sorted(build):
    b = build[name]
    if b != "BUILD_OK":
        rows.append((name, "build FAILED", "-", ""))
        failed = True
        continue
    if name not in blocks:
        rows.append((name, "built", "did not run / no output", ""))
        failed = True
        continue
    out, code = blocks[name]
    ok = re.search(r"^OK %s$" % re.escape(name), out, re.M) is not None
    tail = " / ".join([l for l in out.strip().split("\n") if l][-3:])
    rows.append((name, "built", ("PASS" if ok else "FAIL") + f" (exit {code})", tail))
    failed |= not ok

md = ["| program | build | run on Redox | last output |", "|---|---|---|---|"]
for r in rows:
    md.append("| " + " | ".join(x.replace("|", "\\|") for x in r) + " |")
text = "\n".join(md)
print(text)
summ = os.environ.get("GITHUB_STEP_SUMMARY")
if summ:
    with open(summ, "a") as f:
        f.write("\n### Validation ladder\n\n" + text + "\n")
sys.exit(1 if failed else 0)
