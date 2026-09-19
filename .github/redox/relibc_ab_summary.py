#!/usr/bin/env python3
"""Summarise run_relibc_ab.sh: per repro and variant, how many runs printed OK / FAIL / hung, plus
the interesting counters (x27 spawn errors, x33 wrong-thread handler runs)."""
import os, re, sys, collections
log = open(sys.argv[1], errors="replace").read().replace("\r", "") if os.path.exists(sys.argv[1]) else ""
rows = collections.OrderedDict()
for m in re.finditer(r"=== BEGIN \S*?/(\w+?)_(unpatched|patched)\.bin variant=\w+ run=(\d+)\n(.*?)=== END \S+ variant=\w+ run=\d+ exit=(\S+)", log, re.S):
    name, v, n, out, rc = m.groups()
    r = rows.setdefault((name, v), dict(runs=0, ok=0, fail=0, hang=0, extra=0))
    r["runs"] += 1
    if "HANG" in out: r["hang"] += 1
    elif re.search(r"^OK %s" % name, out, re.M): r["ok"] += 1
    else: r["fail"] += 1
    if name == "x27_spawnpar_c":
        r["extra"] += sum(int(x) for x in re.findall(r"spawn errors (\d+)", out))
    if name == "x33_thread_sigmask_c":
        r["extra"] += sum(int(x) for x in re.findall(r"had SIGUSR1 blocked: (\d+)", out))
md = ["| repro | variant | runs | OK | FAIL | HANG | counter (x27: spawn errors, x33: wrong-thread handler runs) |", "|---|---|---|---|---|---|---|"]
for (name, v), r in rows.items():
    md.append(f"| {name} | {v} | {r['runs']} | {r['ok']} | {r['fail']} | {r['hang']} | {r['extra'] if name in ('x27_spawnpar_c','x33_thread_sigmask_c') else ''} |")
faults = len(re.findall(r"^Invalid opcode fault", log, re.M))
zero = len(re.findall(r"^Page fault: 0000000000000000 US \| ID", log, re.M))
md.append(f"zero-register thread faults (Page fault at RIP 0): {zero}")
md.append(f"\nInvalid opcode faults (a spawned child killed in the loader before main): {faults}; VM froze: {'yes' if 'VM FROZE' in log else 'no'}; kernel panic: {'yes' if 'KERNEL PANIC' in log else 'no'}")
if "VM FROZE" in log or "KERNEL PANIC" in log:
    md.append("\nVM froze / kernel panic during the run (later runs missing)")
text = "\n".join(md) if rows else "no results in the log"
print(text)
if os.environ.get("GITHUB_STEP_SUMMARY"):
    open(os.environ["GITHUB_STEP_SUMMARY"], "a").write("\n### relibc patches A/B\n\n" + text + "\n")
