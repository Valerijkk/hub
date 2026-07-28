import re
import collections
from pathlib import Path

raw = Path(r"C:/Users/Valerii/AppData/Local/Temp/hub-cov-run.log").read_bytes()
log = raw.decode("utf-16-le")
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
    if "/internal/" not in fpath:
        continue
    rel = fpath.split("/internal/", 1)[1]
    pkg = rel.rsplit("/", 1)[0]
    segs = pkg.split("/")
    top = ("platform/" + segs[1]) if segs[0] == "platform" and len(segs) > 1 else segs[0]
    funcs[top].append(pct)

rows = sorted(((k, sum(v) / len(v), len(v)) for k, v in funcs.items()), key=lambda x: x[0])
print("packages", len(rows))
for k, avg, n in rows:
    print(f"{k}\t{avg:.1f}\t{n}")

tm = re.findall(r"total:\s+\(statements\)\s+([\d.]+)%", log)
total = tm[-1] if tm else "?"
print("TOTAL_STATEMENTS", total)

out = Path(r"c:/Users/Valerii/Desktop/hub/.planning/planning-audit-2026-07")
out.mkdir(parents=True, exist_ok=True)
(out / "cov-2026-07-27-func-avg.tsv").write_text(
    "package\tfunc_avg_pct\tfunc_count\n"
    + "\n".join(f"{k}\t{avg:.1f}\t{n}" for k, avg, n in rows)
    + "\n",
    encoding="utf-8",
)
# Save full func dump as utf-8 for archive
(out / "cov-2026-07-27-func.txt").write_text(log, encoding="utf-8")
meta = "\n".join(
    [
        "measured=2026-07-27",
        "method=docker golang:1.25-bookworm network=hub_default",
        "command=go test -tags hubtest -count=1 -timeout=900s -covermode=atomic -coverpkg=./internal/... ./testframework/...",
        f"overall_statements={total}%",
        "note=per-package TSV is mean of go tool cover -func percentages (proxy); overall is statement total",
        "",
    ]
)
(out / "cov-2026-07-27-total.txt").write_text(meta, encoding="utf-8")
print("written")
