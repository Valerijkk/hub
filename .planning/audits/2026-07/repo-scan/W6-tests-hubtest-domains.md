# W6-tests-hubtest-domains — Backend hubtest domains

> **Provenance class:** FALLBACK_SUMMARY · **Confidence:** PRELIMINARY · **current_verified:** false (pending code/source recheck)
> Embedded original verdict below = historical self-check only — **not** independent acceptance. See [`PROVENANCE.md`](./PROVENANCE.md).

> **Роль:** qa-auto-ver (readonly) · **Дата:** 2026-07-23
> **Scope:** denial-suite gaps по доменам с CRITICAL IDOR · **Код не менялся**

## Summary

Есть сильный эталон `bases/rbac_test.go`. **Нет** полноценных cross-project/cross-channel denial suites там, где W1/W2 нашли CRITICAL (tms, project, messenger, calls, crm, imex).

**Вердикт:** `verified: false`.

## Findings

| Severity | file:line | Issue | Fix |
|----------|-----------|-------|-----|
| **CRITICAL** | `testframework/tms/` | Нет IDOR denial (case/run/suite cross-project) при W2 CRITICAL | Suite как bases/rbac |
| **CRITICAL** | `testframework/project/` | Нет foreign-project CRUD denial (W2) | Denial suite |
| **CRITICAL** | `testframework/messenger/` | Search/delete/polls IDOR без denial (W1) | Denial suite |
| **HIGH** | `testframework/calls/` | CreateCall/membership gaps (W2) | Negative tests |
| **HIGH** | `testframework/crm/` | Нет B-mutates-A denial (W3) | Denial suite |
| **HIGH** | `testframework/imex/` | Нет RBAC/membership tests | Добавить |
| **PASS** | `testframework/bases/rbac_test.go` | Эталон owner-scope | Копировать паттерн |
| **PASS** | auth hubtest | Magic-link / org tests существуют | Pending-session denial missing (W1) |
