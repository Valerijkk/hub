# -*- coding: utf-8 -*-
"""Recover full CP1251-mojibake spans (UTF-8 Cyrillic mis-decoded as CP1251)."""
from __future__ import annotations

import re
import subprocess
from pathlib import Path

# Start of a typical digraph
START = re.compile(r"[РС]")


def can_recover(s: str) -> str | None:
    if len(s) < 4:
        return None
    try:
        out = s.encode("cp1251").decode("utf-8")
    except (UnicodeEncodeError, UnicodeDecodeError):
        return None
    # recovered text should contain real Cyrillic letters
    if not re.search(r"[А-Яа-яЁё]", out):
        return None
    # and should NOT still look like digraph soup
    if re.search(r"(?:Р[А-Яа-я]|С[А-Яа-я]){3,}", out):
        return None
    return out


def fix_text(text: str) -> tuple[str, int]:
    """Greedy: from each digraph start, take longest recoverable suffix-prefix."""
    chars = list(text)
    i = 0
    nfix = 0
    out: list[str] = []
    while i < len(chars):
        if chars[i] not in ("Р", "С"):
            out.append(chars[i])
            i += 1
            continue
        # try longest recoverable span starting at i
        best = None
        best_j = i
        # limit span length
        max_j = min(len(chars), i + 400)
        for j in range(max_j, i + 3, -1):
            cand = "".join(chars[i:j])
            # prefer ending on word boundary-ish
            recovered = can_recover(cand)
            if recovered is not None:
                best = recovered
                best_j = j
                break
        if best is not None:
            out.append(best)
            nfix += 1
            i = best_j
        else:
            out.append(chars[i])
            i += 1
    return "".join(out), nfix


def main() -> None:
    files = subprocess.check_output(
        ["git", "ls-files", "backend/ent/schema/*.go"], text=True, encoding="utf-8"
    ).splitlines()
    total = 0
    for f in files:
        p = Path(f)
        original = p.read_text(encoding="utf-8")
        fixed, n = fix_text(original)
        if n and fixed != original:
            # Preserve newlines as in original (read_text normalizes; write with \n)
            p.write_text(fixed, encoding="utf-8", newline="\n")
            print(f"FIXED\t{n}\t{f}")
            total += 1
    print(f"FILES_FIXED={total}")


if __name__ == "__main__":
    main()
