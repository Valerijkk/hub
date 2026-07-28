import re
import collections
from pathlib import Path

raw = Path(r"C:/Users/Valerii/AppData/Local/Temp/hub-cov-run.log").read_bytes()
log = raw.decode("utf-16-le")

# Per-file statement approx: average of function percents in file, then package =
# mean of functions (same as go tool cover doesn't give package totals from -func alone).
# Better approach used here: treat each function equally (documented as proxy).

BASELINE = {
    "ws": 87.7,
    "storage": 86.8,
    "rbac": 84.7,
    "search": 92.2,
    "platform/outbox": 88.6,
    "notification": 82.5,
    "auth": 82.4,
    "tracker": 82.4,
    "messenger": 82.3,
    "audit": 80.5,
    "calls": 80.2,
    "tms": 82.0,
    "project": 82.0,
    "sprint": 81.7,
    "hr": 83.7,
    "crm": 81.9,
    "imex": 84.7,
    "ai": 78.9,
    "github": 80.8,
    "apigateway": 82.2,
    "dashboard": 86.4,
    "accounting": 80.4,
    "calendar": 86.7,
    "intranet": 81.3,
    "bases": 78.6,
    "marketplace": 80.1,
    "backup": 85.8,
    "integrations/onec": 83.4,
}

funcs = collections.defaultdict(list)
for line in log.splitlines():
    if "backend/internal/" not in line or not line.strip().endswith("%"):
        continue
    if line.startswith("total:"):
        continue
    parts = re.split(r"[\t ]+", line.strip())
    if len(parts) < 2:
        continue
    pct_s = parts[-1]
    if not pct_s.endswith("%"):
        continue
    try:
        pct = float(pct_s[:-1])
    except ValueError:
        continue
    loc = parts[0]
    if ".go:" not in loc:
        continue
    fpath = loc.split(".go:")[0] + ".go"
    rel = fpath.split("/internal/", 1)[1]
    segs = rel.rsplit("/", 1)[0].split("/")
    if segs[0] == "platform" and len(segs) > 1:
        top = "platform/" + segs[1]
    elif segs[0] == "integrations" and len(segs) > 1:
        top = "integrations/" + segs[1]
    else:
        top = segs[0]
    funcs[top].append(pct)

tm = re.findall(r"total:\s+\(statements\)\s+([\d.]+)%", log)
overall = float(tm[-1]) if tm else None

out = Path(r"c:/Users/Valerii/Desktop/hub/.planning/planning-audit-2026-07")
rows = []
for pkg, vals in sorted(funcs.items()):
    avg = sum(vals) / len(vals)
    base = BASELINE.get(pkg)
    delta = None if base is None else round(avg - base, 1)
    rows.append((pkg, avg, len(vals), base, delta))

lines = [
    "# Coverage gate — measured 2026-07-27 (Docker, full hubtest)",
    "",
    "**Method:** `docker run --network hub_default golang:1.25-bookworm`",
    "`go test -tags hubtest -count=1 -timeout=900s -covermode=atomic -coverpkg=./internal/... ./testframework/...`",
    "`go tool cover -func` (statement overall from cover tool).",
    "",
    f"**Overall `internal/` statements (go tool cover):** **{overall}%**",
    "**Baseline 2026-06-22:** 83.0%",
    f"**Delta overall:** **{overall - 83.0:+.1f} pp**",
    "",
    "> Per-package numbers below are **mean of per-function statement %** from",
    "> `go tool cover -func` (proxy). Overall line is true statement coverage.",
    "> A second Docker run that would emit a full coverprofile failed mid-suite",
    "> on proxy.golang.org EOF; overall from the successful first run is authoritative.",
    "",
    "## Packages in baseline that declined (func-mean proxy)",
    "",
    "| Package | 2026-07-27 (func-mean) | Baseline 2026-06-22 | Delta |",
    "|---------|------------------------:|--------------------:|------:|",
]
declines = [r for r in rows if r[4] is not None and r[4] < -0.05]
declines.sort(key=lambda x: x[4])
for pkg, avg, n, base, delta in declines:
    lines.append(f"| {pkg} | {avg:.1f}% | {base}% | {delta:+.1f} |")
if not declines:
    lines.append("| _(none)_ | | | |")

lines += [
    "",
    "## Full package table (func-mean proxy)",
    "",
    "| Package | Func-mean % | #funcs | Baseline | Delta |",
    "|---------|------------:|-------:|---------:|------:|",
]
for pkg, avg, n, base, delta in rows:
    b = "" if base is None else f"{base}"
    d = "" if delta is None else f"{delta:+.1f}"
    lines.append(f"| {pkg} | {avg:.1f}% | {n} | {b} | {d} |")
lines.append("")

(out / "cov-2026-07-27-REPORT.md").write_text("\n".join(lines), encoding="utf-8")
(out / "cov-2026-07-27-func-avg.tsv").write_text(
    "package\tfunc_avg\tfuncs\tbaseline\tdelta\n"
    + "\n".join(
        f"{pkg}\t{avg:.1f}\t{n}\t{'' if base is None else base}\t{'' if delta is None else delta}"
        for pkg, avg, n, base, delta in rows
    )
    + "\n",
    encoding="utf-8",
)
(out / "cov-2026-07-27-total.txt").write_text(
    f"measured=2026-07-27\noverall_statements={overall}%\nbaseline_2026_06_22=83.0%\ndelta={overall - 83.0:+.1f}pp\nmethod=docker_golang_1.25_hub_default\n",
    encoding="utf-8",
)
print("overall", overall)
print("packages", len(rows), "declines", len(declines))
for r in declines[:25]:
    print("DOWN", r)
