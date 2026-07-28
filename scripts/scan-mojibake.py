# -*- coding: utf-8 -*-
"""Scan tracked files for UTF-8 mojibake (Russian text corrupted via CP1251 mis-decode)."""
from __future__ import annotations

import re
import subprocess
from pathlib import Path

SKIP_EXT = {
    ".png", ".jpg", ".jpeg", ".gif", ".webp", ".ico", ".pdf", ".zip", ".gz",
    ".woff", ".woff2", ".ttf", ".eot", ".mp4", ".webm", ".exe", ".dll", ".so",
    ".sum", ".mod", ".bin", ".wasm",
}

# Mojibake of UTF-8 Cyrillic decoded as CP1251: each Cyrillic letter becomes
# two chars starting with Р (U+0420) or С (U+0421). Real Russian almost never
# has 3+ consecutive such digraphs like Р°РЅРё.
DIGRAPH = (
    "Р°|Р±|РІ|Рі|Рґ|Рµ|Р¶|Р·|Рё|Рй|Рє|Р»|Рј|РЅ|Рѕ|Рї|"
    "СЂ|СЃ|С‚|Сѓ|С„|С…|С†|С‡|С€|С‰|СЉ|С‹|СЊ|СЭ|СЮ|СЯ|"
    "Рђ|Р‘|Р’|Р“|Р”|Р•|Р–|Р—|Р˜|Р™|Рљ|Р›|Рњ|Рќ|Рћ|Рџ|"
    "СЂ|СЃ|С‚|Сѓ|С„|С…|С†|С‡|С€|С‰|"
    "РЎ|Рў|РЈ|Р¤|РҐ|Р¦|Р§|РЁ|Р©|РЄ|Р«|Р¬|Р­|Р®|РЇ|"
    # degree/ellipsis variants when continuation byte mapped to Latin-1
    "Р°|Р±|РІ|Р°|С…"  # keep С…
)
# At least 3 consecutive digraphs (6+ chars) — strong signal
CHAIN = re.compile(
    r"(?:Р[А-Яа-яЁё°±І]|С[А-Яа-яЁёЂ-џ…]){3,}"
)

# Also: invalid UTF-8
# And classic latin1 mojibake
LATIN1 = re.compile(r"Ã[\x80-\xBF]|â€[™œ]|â€™|â€œ")


def main() -> None:
    files = subprocess.check_output(["git", "ls-files"], text=True, encoding="utf-8").splitlines()
    hits: list[tuple[str, str, str, int]] = []
    for f in files:
        p = Path(f)
        if p.suffix.lower() in SKIP_EXT:
            continue
        try:
            data = p.read_bytes()
        except OSError:
            continue
        if b"\x00" in data[:2048]:
            continue
        try:
            text = data.decode("utf-8")
        except UnicodeDecodeError as e:
            hits.append((f, "INVALID_UTF8", str(e)[:100], 1))
            continue
        found = list(CHAIN.finditer(text))
        if not found:
            found = list(LATIN1.finditer(text))
            kind = "LATIN1_MOJIBAKE"
        else:
            kind = "CP1251_MOJIBAKE"
        if found:
            i = found[0].start()
            sample = text[max(0, i - 10) : i + 50].replace("\n", " ").replace("\r", "")
            hits.append((f, kind, sample, len(found)))

    out = Path(".planning/audits/2026-07/ENCODING-MOJIBAKE-SCAN.md")
    out.parent.mkdir(parents=True, exist_ok=True)
    lines = [
        "# Encoding / mojibake scan",
        "",
        f"**Date:** 2026-07-28",
        f"**Tracked files scanned:** {len(files)}",
        f"**Hits:** {len(hits)}",
        "",
        "## Method",
        "",
        "Heuristic: ≥3 consecutive CP1251-style digraphs (`Рx`/`Сx`) that appear when",
        "UTF-8 Cyrillic is mis-decoded as Windows-1251 and re-saved as UTF-8.",
        "Legitimate Russian does **not** match (avoids false positives on `Регламент`).",
        "",
        "## Why prior audits missed this",
        "",
        "- Planning / security / FR audits check requirements and structure, not UTF-8 validity.",
        "- `code-reviewer` playbook has no encoding checklist item.",
        "- Mojibake compiles fine in Go (comments only) — CI never fails.",
        "- Console CP1251 on Windows can hide or garble evidence during agent sessions.",
        "",
        "## Hits",
        "",
        "| Kind | n | Path | Sample |",
        "|------|--:|------|--------|",
    ]
    for path, kind, sample, n in hits:
        safe = sample.replace("|", "\\|")[:80]
        lines.append(f"| {kind} | {n} | `{path}` | `{safe}` |")
    if not hits:
        lines.append("| — | 0 | — | clean |")
    lines.append("")
    out.write_text("\n".join(lines), encoding="utf-8")
    # machine summary to stdout (ascii-safe)
    print(f"TRACKED={len(files)} HITS={len(hits)} REPORT={out}")
    for path, kind, sample, n in hits[:40]:
        print(f"{kind}\t{n}\t{path}")
    if len(hits) > 40:
        print(f"... +{len(hits)-40} more")


if __name__ == "__main__":
    main()
