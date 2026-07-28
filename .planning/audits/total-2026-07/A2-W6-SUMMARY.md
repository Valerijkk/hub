# A2-W6 SUMMARY — auth + rbac + audit

> TASK-008 · Stage A2 · read-only · 2026-07-28  
> Периметр: `W6-ASSIGNMENTS.md` · 92 файла (+ `gdpr_mongo.go` в A4)  
> Отчёты: `A2W6-1-auth-core.md` … `A2W6-5-audit-tests.md`

## Файлы

| Агент | Список | Прочитано | Пропущено |
|-------|-------:|----------:|----------:|
| A1 auth core | 30 | 30 | 0 |
| A2 auth tests | 13 | 13 | 0 |
| A3 rbac | 15 | 15 | 0 |
| A4 audit core | 22 | 22 | 0 |
| A5 audit tests | 12 | 12 | 0 |
| **Σ** | **92** | **92** | **0** |

## Находки по тегам

| Агент | NEW | REGRESS | RESIDUAL | FIXED-OK | Σ open |
|-------|----:|--------:|---------:|---------:|-------:|
| A1 | 0 | 0 | 6 | 4 | 6 |
| A2 | 0 | 0 | 3 | 2 | 3 |
| A3 | 0 | 0 | 5 | 10 | 5 |
| A4 | 2 | 0 | 6 | 9 | 8 |
| A5 | 1 | 0 | 8 | 2 | 9 |
| **Σ** | **3** | **0** | **28** | **27** | **31** |

## Severity (только open = NEW+REGRESS+RESIDUAL)

| Агент | P0 | P1 | P2 | P3 | n |
|-------|---:|---:|---:|---:|--:|
| A1 | 0 | 0 | 4 | 2 | 6 |
| A2 | 0 | 1 | 2 | 0 | 3 |
| A3 | 0 | 0 | 3 | 2 | 5 |
| A4 | 0 | 0 | 7 | 1 | 8 |
| A5 | 0 | 0 | 6 | 3 | 9 |
| **Σ** | **0** | **1** | **22** | **8** | **31** |

Сравнение Stage A W6: было **55** сырых (P1=6 · P2=28 · P3=21) → A2 open **31**; **27 FIXED-OK**; **REGRESS=0**.

## Топ открытых

1. **A2W6-A2-001** (P1 RESIDUAL) — позитивный Telegram `HandleLogin_StoresCode` / `tglogin:` (ADR-AUTH-01).
2. **Telegram login surface** (P2 RESIDUAL) — A1-001/002 + A2-002/003 — мёртвый login API + suite (P1/P2-U unresolved).
3. **PurgeExpired fail-open** (P2 RESIDUAL + NEW cement) — A4-001 + A5-002 `LogsErrorsWithoutFailing`.
4. **CheckLimit / numeric limits** (P2 RESIDUAL) — A4-002; TASK-040 не покрыл audit.
5. **GDPR cascade non-atomic** (P2 RESIDUAL) — A4-003 partial erase без rollback.
6. **AvatarURL / audit err.Error() / middleware recover** — P2 residuals/new.

## Вердикт

- Критичных P0 нет; единственный P1 — Telegram login positive test.
- P1 Stage A (RBAC cache, GDPR HardDelete abort) — **FIXED-OK**, мутации W18 зелёные.
- Unresolved product decisions: Telegram login delete vs quarantine (P*-U*).
- Код **не менялся** (A2 read-only).
