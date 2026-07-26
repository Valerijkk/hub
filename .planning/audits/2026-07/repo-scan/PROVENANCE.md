# Repo Scan 2026-07 — Provenance & Confidence Policy

> **Создано:** 2026-07-26 (provenance correction pass)
> **Цель:** прозрачная история происхождения отчётов без удаления артефактов и без ложного впечатления «40 независимых Grok-прогонов».

## Что было неверно в первоначальной метке

| Утверждение (устарело) | Факт |
|------------------------|------|
| Заголовок «40× Grok 4.5» | **Ложно** как описание 40 независимых agent-run |
| «Ask-recovery complete» как единый метод | Смешаны **Ask-mode recovery**, **прерванные** прогоны и **один fallback** |
| SCORECARD totals ≈20 CRITICAL как «verified repo» | **Непроверенный агрегат**; dedup приблизительный; репозиторий **не verified** |

## Фактическая модель: 40 report slots

На диске **40 файлов** в слотах W1–W8 + SCORECARD. Это **не** 40 независимых readonly-аудитов одной модели.

```
40 slots =
  19 W*.md до fallback (разное происхождение)
+ 20 W*.md от одного fallback generalPurpose-агента (оркестратор)
+  1 SCORECARD.md (агрегат fallback-сессии)
```

**До fallback** на диске было **19** `W*.md`. Один fallback-агент (`generalPurpose`) создал **оставшиеся 20** domain-отчётов и **SCORECARD**.

Per-file класс зафиксирован ниже (19 + 20 = **39** domain-отчётов). Модель/агент per-file **не** выводить без evidence в transcript или git.

## Классы происхождения (provenance class)

| Class | Описание | Доверие к severity |
|-------|----------|-------------------|
| **PRE_EXISTING** | Файл на диске до fallback-сессии (19 шт.) | Mixed — часть полная, часть Ask-recovery/interrupted |
| **FALLBACK_SUMMARY** | Создан fallback generalPurpose в одной сессии (20 W*) | **PRELIMINARY / UNVERIFIED** до перечитывания кода |
| **AGGREGATE** | SCORECARD.md | **Non-authoritative** rollup; не независимый audit |
### Политика confidence

1. **Не удалять** отчёты — история сохраняется; исправляем метаданные и интерпретацию.
2. **FALLBACK_SUMMARY** и **UNKNOWN** — findings **не считаются независимыми** до code recheck (handler/service/test grep или targeted verifier).
3. **Echo severity** (FE «CRITICAL» = echo BE IDOR; test «missing suite» = process echo) — **не увеличивают** unique product CRITICAL count.
4. **PASS** в отдельном W*-файле (например W7-tz-grep) — valid **только для scope этого файла**, не для всего репо.
5. **`current_verified: false`** на каждом W*-файле — до code/source recheck; embedded PASS = **historical_self_check**, не independent acceptance.
6. **`verified: false`** для репозитория как целого — до закрытия P0 **и** выборочной reverification fallback-отчётов.

## Per-file registry (39 domain reports)

Каждый `W*.md` несёт [стандартный banner](./W1-auth.md) с `current_verified: false`. Таблица — канон класса и confidence.

| File | Provenance class | Confidence | current_verified | Notes |
|------|------------------|------------|------------------|-------|
| [W1-auth](./W1-auth.md) | PRE_EXISTING | MIXED | false | pending→session HIGH |
| [W1-rbac-ws](./W1-rbac-ws.md) | PRE_EXISTING | MIXED | false | custom role IDOR ×3 |
| [W1-tracker-sprint](./W1-tracker-sprint.md) | PRE_EXISTING | MIXED | false | JQL layering + IDOR |
| [W1-messenger](./W1-messenger.md) | PRE_EXISTING | MIXED | false | search NF-03; delete; polls |
| [W1-security-crosscut](./W1-security-crosscut.md) | PRE_EXISTING | MIXED | false | historical_self_check: PASS (cross-cut scope) |
| [W2-tms](./W2-tms.md) | PRE_EXISTING | MIXED | false | case/run/suite/env IDOR cluster |
| [W2-calls](./W2-calls.md) | PRE_EXISTING | MIXED | false | CreateCall + room-auth gaps |
| [W2-notification-search](./W2-notification-search.md) | PRE_EXISTING | MIXED | false | MEDIUM only in original scan |
| [W2-storage-apigateway](./W2-storage-apigateway.md) | PRE_EXISTING | MIXED | false | CRITICAL/HIGH = 0; MEDIUM open |
| [W2-project-dashboard](./W2-project-dashboard.md) | PRE_EXISTING | MIXED | false | project+dashboard IDOR cluster |
| [W3-v4-bases-crm-hr](./W3-v4-bases-crm-hr.md) | PRE_EXISTING | MIXED | false | crm mutate IDOR |
| [W3-platform](./W3-platform.md) | FALLBACK_SUMMARY | PRELIMINARY | false | historical_self_check: PASS (platform-core scope) |
| [W3-ai](./W3-ai.md) | FALLBACK_SUMMARY | PRELIMINARY | false | SSRF PASS claim — recheck |
| [W3-v4-intranet-calendar](./W3-v4-intranet-calendar.md) | FALLBACK_SUMMARY | PRELIMINARY | false | leave org-wide HIGH |
| [W3-audit-backup-imex](./W3-audit-backup-imex.md) | FALLBACK_SUMMARY | PRELIMINARY | false | imex auth-only CRITICAL |
| [W4-fe-tracker](./W4-fe-tracker.md) | FALLBACK_SUMMARY | PRELIMINARY | false | echo BE tracker IDOR |
| [W4-fe-messenger](./W4-fe-messenger.md) | FALLBACK_SUMMARY | PRELIMINARY | false | echo BE + vitest gap |
| [W4-fe-tms](./W4-fe-tms.md) | FALLBACK_SUMMARY | PRELIMINARY | false | echo BE tms IDOR |
| [W4-fe-calls](./W4-fe-calls.md) | FALLBACK_SUMMARY | PRELIMINARY | false | echo BE + no vitest |
| [W4-fe-auth-shell](./W4-fe-auth-shell.md) | FALLBACK_SUMMARY | PRELIMINARY | false | echo pending session |
| [W5-fe-v4](./W5-fe-v4.md) | FALLBACK_SUMMARY | PRELIMINARY | false | echo crm + smoke |
| [W5-fe-admin](./W5-fe-admin.md) | FALLBACK_SUMMARY | PRELIMINARY | false | historical_self_check: PASS (file scope) |
| [W5-fe-design-system](./W5-fe-design-system.md) | FALLBACK_SUMMARY | PRELIMINARY | false | historical_self_check: PASS (file scope) |
| [W5-fe-a11y](./W5-fe-a11y.md) | FALLBACK_SUMMARY | PRELIMINARY | false | NFR-16 full open |
| [W5-fe-stores-api](./W5-fe-stores-api.md) | FALLBACK_SUMMARY | PRELIMINARY | false | historical_self_check: PASS (file scope) |
| [W6-tests-coverage](./W6-tests-coverage.md) | PRE_EXISTING | MIXED | false | 6×80% CI; inventory |
| [W6-tests-vitest](./W6-tests-vitest.md) | PRE_EXISTING | MIXED | false | TC↔Vitest; calls/messenger holes |
| [W6-tests-hubtest-core](./W6-tests-hubtest-core.md) | FALLBACK_SUMMARY | PRELIMINARY | false | inventory/CI echo |
| [W6-tests-hubtest-domains](./W6-tests-hubtest-domains.md) | FALLBACK_SUMMARY | PRELIMINARY | false | missing denial suites (process echo) |
| [W6-tests-e2e](./W6-tests-e2e.md) | FALLBACK_SUMMARY | PRELIMINARY | false | waitForTimeout ai-hub |
| [W7-ci](./W7-ci.md) | PRE_EXISTING | MIXED | false | MEDIUM minutes economy |
| [W7-deploy-compose](./W7-deploy-compose.md) | PRE_EXISTING | MIXED | false | deploy/compose drift |
| [W7-docs-drift](./W7-docs-drift.md) | PRE_EXISTING | MIXED | false | wiki≠code HIGH omissions |
| [W7-tz-grep](./W7-tz-grep.md) | PRE_EXISTING | MIXED | false | historical_self_check: PASS (grep scope only) |
| [W7-fa06-secrets](./W7-fa06-secrets.md) | FALLBACK_SUMMARY | PRELIMINARY | false | Phase 0 only — prod blocker |
| [W8-gsd-ru](./W8-gsd-ru.md) | PRE_EXISTING | MIXED | false | historical_self_check: PASS (pack depth) |
| [W8-prompt-engineering](./W8-prompt-engineering.md) | PRE_EXISTING | MIXED | false | historical_self_check: PASS (readonly audit scope) |
| [W8-cursor-agents](./W8-cursor-agents.md) | FALLBACK_SUMMARY | PRELIMINARY | false | historical_self_check: PASS (G2 scope) |
| [W8-ai-expansion](./W8-ai-expansion.md) | FALLBACK_SUMMARY | PRELIMINARY | false | historical_self_check: PASS (Must structural); M6/M7 open |

**Итого:** PRE_EXISTING **19** · FALLBACK_SUMMARY **20** · все `current_verified: false`.

## SCORECARD — ограничения использования

[`SCORECARD.md`](./SCORECARD.md) — **вспомогательный dedup-черновик**, не финальный security sign-off.

- Totals (**≈20 CRITICAL**, **≈45+ HIGH**) — **approximate**, с эхом и без полного code recheck.
- Top-10 fix order — **рекомендация для triage**, не verified backlog.
- Строки «AI-expansion Must PASS» / «Overall merge-ready» — **не независимая верификация** planning-артефактов; см. [`07-ACCEPTANCE.md`](../../july-2026/ai-expansion/07-ACCEPTANCE.md) (qualified).

## Связанные документы

| Документ | Роль |
|----------|------|
| [`INDEX.md`](./INDEX.md) | Навигация по слотам + ссылка сюда |
| [`SCORECARD.md`](./SCORECARD.md) | Non-authoritative aggregate |
| [`../../july-2026/ai-expansion/07-ACCEPTANCE.md`](../../july-2026/ai-expansion/07-ACCEPTANCE.md) | Qualified Must PASS (structural gaps) |

## История (сохранена)

- **2026-07-23:** первичная волна repo-scan; INDEX помечен «40× Grok 4.5»; fallback закрыл missing slots.
- **2026-07-26:** provenance correction — исправлена метка происхождения, добавлена confidence policy; отчёты **не удалялись**.
- **2026-07-26 (banner pass):** per-file registry + banners на всех 39 W*; `verified: true` → `false` + `historical_self_check` где был PASS.
