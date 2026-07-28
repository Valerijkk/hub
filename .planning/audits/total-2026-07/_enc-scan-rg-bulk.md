# Encoding / mojibake scan (rg + Python bulk)

Date: 2026-07-29
Working dir: c:\Users\Valerii\Desktop\hub
Do not fix — scan only.

## Scope

- .planning/backlog/TASK-1{2,3,4,5,6,7}*-2026-07-29*.md
- .planning/audits/total-2026-07/A3*
- .planning/audits/total-2026-07/00-*
- .planning/backlog/README.md

Target files scanned: **137**

## Tool notes

- ripgrep: Cursor build does **not** support --files-from; used -g globs on .planning/backlog + .planning/audits/total-2026-07 with post-filter to scope.
- Patterns passed via -f UTF-8 pattern files to avoid PowerShell console mangling.
- Python 3.13: authoritative Unicode match + UTF-8 sequence validation.
- Ð.|Ñ. / Ã. may false-positive on rare Latin text; treat hits as candidates.

## Patterns (rg counts)

| Pattern | Regex | rg matches |
|---------|-------|------------|
| cp866-box-emoji | ╨\|╤\|тЖТ\|тАФ\|тЬЕ\|┬╖\|ЁЯ | 0 |
| replacement-char | U+FFFD | 0 |
| latin1-cyrillic | Ð.\|Ñ. | 0 |
| utf8-as-latin1 | â€\|Ã. | 0 |

## Summary

- Python total matches: **0**
- Python unique FAIL files: **0**
- rg total match lines (sum patterns): **0**
- Unique FAIL files (union rg+py+invalid-utf8): **0**
- Match lines listed below (cap 200): **0**

### Unique FAIL files

(none)

## UTF-8 validity (Python)

- UTF8_OK: **137**
- UTF8_BAD: **0**

No invalid UTF-8 sequences.

## Matching lines (cap 200)

(no hits)
