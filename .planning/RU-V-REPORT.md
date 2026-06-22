# RU-V — Сводный отчёт верификации русификации

**Date:** 2026-06-22  
**Scope:** Wave RU-1 (30 исполнителей) → Wave RU-V (10 верификаторов)  
**Plan:** `.planning/TASKS-RU-COMMENTS-DOCS.md`

**Pass criterion:** 0 английских комментариев/docs-прозы (исключения: `@swagger`, URL, имена полей API, `@example` JSON, `// biome-ignore`, generated files, product-name section headers).

---

## Executive Summary

Wave RU-1 помечена ✅, post-verification выявила системные gaps: EN-заголовки в core docs, устаревший API ref, незавершённая русификация prod-кода (tms/project/messenger/calls/platform) и ~135 gaps в testframework.

| Область | Verdict | Gap estimate |
|---------|---------|--------------|
| docs/ (RU-V01–V02) | **FAIL / PARTIAL** | ~40+ EN headers + stale API ref |
| backend auth/ws/rbac/audit (RU-V03) | **PASS** | minor (export_hooks godoc EN — OK) |
| backend tracker/tms/project (RU-V04) | **FAIL** | EN comments + mojibake |
| backend messenger/calls/platform (RU-V05) | **FAIL** | EN comments + mojibake |
| backend остальные (RU-V06) | **FAIL** | ~173 (~81 dev/seed) |
| ent + cmd (RU-V07) | **PASS** (near-clean) | ~4–6 |
| frontend src (RU-V08) | **PASS** | ~5 |
| testframework + e2e (RU-V09) | **FAIL** | ~135 |
| **RU-V10 gate** | **FAIL** | см. remediation |

---

## Verdict Table — RU-V01 … RU-V09

| ID | Проверка | Verdict | Gap count | Notes |
|----|----------|---------|-----------|-------|
| RU-V01 | docs/ полнота русского | **FAIL** | ~40+ | EN headers в TECHNICAL.md, architecture.md, security_guidelines.md, code_style_guide.md |
| RU-V02 | docs/ актуальность vs код | **PARTIAL/FAIL** | API drift | 82 ent schemas в architecture.md vs «~30» в rules; TECHNICAL.md без v4 modules |
| RU-V03 | backend auth/ws/rbac/audit | **PASS** | ~0 prod | export_hooks EN godoc допустимо |
| RU-V04 | backend tracker/tms/project | **FAIL** | EN + mojibake | tms/service_import.go, testrail_import.go, project/repository.go |
| RU-V05 | backend messenger/calls/platform | **FAIL** | EN + mojibake | messenger services, calls handlers, platform crypto/cron |
| RU-V06 | backend остальные | **FAIL** | ~173 | ~81 в dev/seed |
| RU-V07 | ent + cmd | **PASS** | ~4–6 | 3 ent godoc + main.go prose |
| RU-V08 | frontend src | **PASS** | ~5 | 2 prose + 3 JSX labels |
| RU-V09 | testframework + e2e | **FAIL** | ~135 | backend/testframework bulk |

---

## Wave RU-FIX — Prioritized Backlog

### P0 — Docs

| ID | Task | Owner |
|----|------|-------|
| RU-FIX-01 | RU headers в 4 core docs | docs-writer |
| RU-FIX-02 | TECHNICAL.md sync v4 modules | sa + docs-writer |
| RU-FIX-03 | «~30 Ent schemas» → 82 в rules | docs-writer |

### P1 — Backend prod

| ID | Task | Owner |
|----|------|-------|
| RU-FIX-04 | tms + project + sprint | be |
| RU-FIX-05 | messenger + calls + platform | be |
| RU-FIX-06a | остальные prod (excl. dev) | be |
| RU-FIX-06b | dev/seed (~81) | be |

### P2 — Test infrastructure

| ID | Task | Owner |
|----|------|-------|
| RU-FIX-07 | backend/testframework | be + qa |
| RU-FIX-08 | e2e steps | qa + fe |

### P3 — Residual

| ID | Task | Owner |
|----|------|-------|
| RU-FIX-09 | ent + cmd | be |
| RU-FIX-10 | frontend src | fe |

**Execution:** P0 → P1 → P2 → P3 → RU-V10 re-verify

---

## RU-V10 — Final Gate: **PASS** (2026-06-22 re-verify after RU-FIX-11–13)

**Genuine EN-only comment gaps:** 0 (prod + testframework + e2e + frontend after RU-FIX-11–13).

**Mojibake:** fixed in 6 source files (RU-FIX-13); regenerate swagger with `swag init` optional.

**Next:** optional formal grep gate in CI.

---

## Sign-off

| Role | RU-V10 | Date |
|------|--------|------|
| verifier | **PASS** (post RU-FIX-11–13) | 2026-06-22 |
