# Encoding / mojibake scan

**Date:** 2026-07-28  
**Task:** [TASK-007](../../backlog/archive/legacy/TASK-007-utf8-mojibake-audit-fix.md)  
**Tracked files scanned:** 2684  
**Hits before fix:** 33–39 (`backend/ent/schema/*.go` comments)  
**Hits after fix:** **0**

## Method

Heuristic: ≥3 consecutive CP1251-style digraphs (`Рx` / `Сx`) that appear when
UTF-8 Cyrillic is mis-decoded as Windows-1251 and re-saved as UTF-8.
Legitimate Russian does **not** match (avoids false positives on `Регламент`).

Recovery: greedy longest span starting at digraph → `encode('cp1251').decode('utf-8')`.

Scripts (repo):

- `scripts/scan-mojibake.py` — scan + write this report
- `scripts/fix-mojibake.py` — apply recovery to `backend/ent/schema/*.go`

## Why prior audits missed this

1. Planning / security / FR audits check requirements and structure, **not** UTF-8 validity or mojibake heuristics.
2. `code-reviewer` / verification playbooks had **no encoding checklist** item.
3. Mojibake lived in **comments only** — Go compiles, CI stays green.
4. Windows console CP1251 can hide or garble evidence in agent transcripts.
5. No automated scan in `/qa-gate` or pre-commit for encoding.

### Remediation for audits (checklist add)

Add to code-reviewer / docs-writer / qa-gate hygiene:

- [ ] Run `python scripts/scan-mojibake.py` → HITS must be 0
- [ ] New Russian comments must be valid UTF-8 (editor encoding UTF-8)

## Result

All corrupted schema comments restored to readable Russian UTF-8.
`deleted_at` / new comments from TASK-002 were already correct and unchanged in meaning.
