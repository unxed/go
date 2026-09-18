#!/usr/bin/env python3
"""Per-package pass/fail counts from the VM log of run_tests.sh (go test -c binaries, -test.v -test.short)."""
import os, re, sys

log = open(sys.argv[1], errors="replace").read().replace("\r", "") if os.path.exists(sys.argv[1]) else ""
rows = []
for m in re.finditer(r"=== BEGIN /root/mnt/(\w+)\.test variant=test run=1\n(.*?)=== END \S+ variant=test run=1 exit=(\S+) after=(\d+)s", log, re.S):
    name, body, rc, secs = m.groups()
    top = re.search(r"--- top-level results of \w+\n(.*?)--- failures", body, re.S)
    top = top.group(1) if top else ""
    p = len(re.findall(r"^--- PASS", top, re.M))
    f = len(re.findall(r"^--- FAIL", top, re.M))
    s = len(re.findall(r"^--- SKIP", top, re.M))
    fails = re.search(r"--- failures \(any level\) of \w+\n(.*?)--- tail", body, re.S)
    fails = [l.strip() for l in fails.group(1).strip().split("\n") if l.strip()] if fails else []
    tail = re.search(r"--- tail of \w+\n(.*)", body, re.S)
    tail = [l for l in (tail.group(1).strip().split("\n") if tail else []) if l][-2:]
    rows.append((name, p, f, s, rc, secs, fails[:6], tail))
md = ["| package | pass | fail | skip | exit | secs | failing tests / tail |", "|---|---|---|---|---|---|---|"]
for name, p, f, s, rc, secs, fails, tail in rows:
    md.append(f"| {name} | {p} | {f} | {s} | {rc} | {secs} | {'; '.join(fails) or ' / '.join(tail)} |".replace("\n", " "))
if "VM FROZE" in log:
    md.append("\nVM FROZE during the run")
text = "\n".join(md) if rows else "no test results in the log"
print(text)
if os.environ.get("GITHUB_STEP_SUMMARY"):
    open(os.environ["GITHUB_STEP_SUMMARY"], "a").write("\n### std package tests on Redox (go test -c, -short)\n\n" + text + "\n")
