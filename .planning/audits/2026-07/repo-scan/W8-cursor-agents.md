# W8-cursor-agents — `.cursor/agents` GSD coverage

> **Provenance class:** FALLBACK_SUMMARY · **Confidence:** PRELIMINARY · **current_verified:** false (pending code/source recheck)
> Embedded original verdict below = historical self-check only — **not** independent acceptance. See [`PROVENANCE.md`](./PROVENANCE.md).

> **Роль:** docs-writer / team-lead (readonly) · **Дата:** 2026-07-23  
> **Scope:** 32 agent markdown files · **Код не менялся**

## Summary

На диске **32/32** agent файлов содержат упоминания **GSD/gsd-*** (G2 acceptance). `team-lead-opus` — pointer/deprecated path к team-lead.

**Вердикт:** `verified: false` · **historical_self_check:** PASS (G2 scope) · polish M6 agent patterns — open до 15.08.

## Findings

| Severity | file:line | Issue | Fix |
|----------|-----------|-------|-----|
| **PASS** | `.cursor/agents/*.md` ×32 | GSD mentions **32/32** | Держать при новых агентах |
| **INFO** | `team-lead-opus.md` | Deprecated post-cutover → team-lead | Не удалять до cutover ADR |
| **MEDIUM** | M6 P0 patterns | 15/21 P0 только route/practice в agents (07-ACCEPTANCE PARTIAL) | Углубить до 15.08 |
| **LOW** | model frontmatter | Cutover day-X (W4 tech plan) open | Checklist 02-TECH-PLAN |
| **PASS** | GSD-INTEGRATION | Матрица agent→skills | — |
| **INFO** | verifiers `*-ver` | GSD mentions present | — |
