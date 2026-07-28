"""Compute per-package statement coverage from go coverprofile (atomic mode)."""
from __future__ import annotations

import re
from collections import defaultdict
from pathlib import Path

PROFILE = Path(r"C:/Users/Valerii/AppData/Local/Temp/hub-cov-out/hub-cov.out")
FUNC = Path(r"C:/Users/Valerii/AppData/Local/Temp/hub-cov-out/hub-cov-func.txt")
OUT = Path(r"c:/Users/Valerii/Desktop/hub/.planning/planning-audit-2026-07")

# Baseline packages from COVERAGE.md 2026-06-22 (statements %)
BASELINE = {
    "overall": 83.0,
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


def top_pkg(file_path: str) -> str:
    # file_path like github.com/valerii/hub/backend/internal/auth/foo.go
    if "/internal/" not in file_path:
        return "OTHER"
    rel = file_path.split("/internal/", 1)[1]
    segs = rel.split("/")
    if segs[0] == "platform" and len(segs) > 2:
        return "platform/" + segs[1]
    if segs[0] == "integrations" and len(segs) > 2:
        return "integrations/" + segs[1]
    return segs[0]


def parse_profile(path: Path) -> dict[str, tuple[int, int]]:
    """Return pkg -> (covered_stmts, total_stmts). Atomic mode: count>0 means covered."""
    text = path.read_text(encoding="utf-8", errors="replace")
    lines = text.splitlines()
    if not lines:
        raise SystemExit("empty profile")
    # mode: set|count|atomic
    pkg_cov: dict[str, list[int]] = defaultdict(lambda: [0, 0])
    for line in lines[1:]:
        # name.go:startLine.startCol,endLine.endCol numStmt count
        m = re.match(r"(.+):(\d+)\.(\d+),(\d+)\.(\d+) (\d+) (\d+)$", line.strip())
        if not m:
            continue
        fpath, _sl, _sc, _el, _ec, num_stmt_s, count_s = m.groups()
        num_stmt = int(num_stmt_s)
        count = int(count_s)
        pkg = top_pkg(fpath.replace("\\", "/"))
        pkg_cov[pkg][1] += num_stmt
        if count > 0:
            pkg_cov[pkg][0] += num_stmt
    return {k: (v[0], v[1]) for k, v in pkg_cov.items()}


def pct(c: int, t: int) -> float:
    return 0.0 if t == 0 else 100.0 * c / t


def main() -> None:
    OUT.mkdir(parents=True, exist_ok=True)
    if not PROFILE.exists():
        raise SystemExit(f"missing {PROFILE}")
    data = parse_profile(PROFILE)
    overall_c = sum(c for c, _ in data.values())
    overall_t = sum(t for _, t in data.values())
    overall = pct(overall_c, overall_t)

    # Prefer go tool cover total if available
    if FUNC.exists():
        ftxt = FUNC.read_text(encoding="utf-8", errors="replace")
        m = re.search(r"total:\s+\(statements\)\s+([\d.]+)%", ftxt)
        if m:
            overall = float(m.group(1))

    rows = []
    for pkg in sorted(data):
        c, t = data[pkg]
        p = pct(c, t)
        base = BASELINE.get(pkg)
        delta = None if base is None else round(p - base, 1)
        rows.append((pkg, p, c, t, base, delta))

    tsv = ["package\tpct\tcovered\ttotal\tbaseline_2026_06_22\tdelta"]
    for pkg, p, c, t, base, delta in rows:
        tsv.append(
            f"{pkg}\t{p:.1f}\t{c}\t{t}\t{'' if base is None else base}\t{'' if delta is None else delta}"
        )
    (OUT / "cov-2026-07-27-packages.tsv").write_text("\n".join(tsv) + "\n", encoding="utf-8")

    declines = [(pkg, p, base, delta) for pkg, p, c, t, base, delta in rows if delta is not None and delta < -0.05]
    declines.sort(key=lambda x: x[3])

    md = []
    md.append("# Coverage gate — measured 2026-07-27")
    md.append("")
    md.append("**Method:** Docker `golang:1.25-bookworm` on `hub_default`,")
    md.append("`go test -tags hubtest -count=1 -timeout=900s -covermode=atomic -coverpkg=./internal/... ./testframework/...`")
    md.append("")
    md.append(f"**Overall `internal/` statements:** **{overall:.1f}%**")
    md.append(f"**Baseline 2026-06-22:** 83.0%")
    md.append(f"**Delta:** **{overall - 83.0:+.1f} pp**")
    md.append("")
    md.append("## Packages that declined vs baseline")
    md.append("")
    if not declines:
        md.append("_None among packages present in baseline table._")
    else:
        md.append("| Package | 2026-07-27 | Baseline 2026-06-22 | Delta |")
        md.append("|---------|------------:|--------------------:|------:|")
        for pkg, p, base, delta in declines:
            md.append(f"| {pkg} | {p:.1f}% | {base}% | {delta:+.1f} |")
    md.append("")
    md.append("## Full package table")
    md.append("")
    md.append("| Package | Coverage 2026-07-27 | Covered/Total stmts |")
    md.append("|---------|--------------------:|--------------------:|")
    for pkg, p, c, t, base, delta in rows:
        md.append(f"| {pkg} | {p:.1f}% | {c}/{t} |")
    md.append("")
    (OUT / "cov-2026-07-27-REPORT.md").write_text("\n".join(md) + "\n", encoding="utf-8")

    print(f"overall={overall:.1f}%")
    print(f"packages={len(rows)} declines={len(declines)}")
    for pkg, p, base, delta in declines[:20]:
        print(f"DOWN {pkg}: {p:.1f} (was {base}) {delta:+.1f}")


if __name__ == "__main__":
    main()
