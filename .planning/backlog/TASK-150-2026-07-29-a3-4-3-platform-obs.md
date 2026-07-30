# 🧹 A3 obs cancel→Error + UUID cardinality

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P3** · Статус: **Закрыта**  
> Источник: A3W4-3-001 NEW · A3W4-3-002 RESIDUAL · Stage A3 FINAL · TASK-008  
> Исполнитель: **be1** · Mutation checker: **be2** (W13 slot 3, 2026-07-30)

## [OUTCOME]

`finishOperationSpan` / `Timed` / `EndConsumerSpan` не помечают `context.Canceled` и `context.DeadlineExceeded` как Error (span status + metrics `failed`); `sanitizeOperationLabel` отклоняет UUID-shaped сегменты (`tracker.<uuid>` → `other`), чтобы не раздувать Prometheus/map cardinality.

## [ROLE]

- Role: be1
- Responsibility: закрыть **оба** finding ID (A3W4-3-001, A3W4-3-002) + hubtest + mutation-ready evidence; не закрывать соседние A3/W4 ID
- Handoff to: be-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files/resources allowed:
  - `backend/internal/platform/observability/otel.go` (`finishOperationSpan` / cancel classification)
  - `backend/internal/platform/observability/timing.go` (`Timed` / `recordOperation` failed flag для cancel)
  - `backend/internal/platform/observability/consumer.go` (если нужен тонкий wrapper; сейчас делегирует в `finishOperationSpan`)
  - `backend/internal/platform/observability/operation_label.go` (`sanitizeOperationLabel` UUID / segment charset)
  - `backend/internal/platform/observability/metrics.go` (только если call-site sanitize / helpers; не менять RWMutex / Snapshot API без нужды)
  - `backend/testframework/platform/observability/**`
  - `.planning/backlog/README.md` + этот файл (статус при закрытии)
  - audit refs: `.planning/audits/total-2026-07/A3W4-3-platform-obs.md`
- Do not touch:
  - metrics CIDR allowlist / `MetricsEndpoint` / `authorizeMetricsIP` (FIXED-OK W4-A3-004/007)
  - HTTP route label `unmatched` / `http.route` PII (FIXED-OK W4-A3-001/002)
  - baggage propagator / redacted span errors без raw `err.Error()` (FIXED-OK W4-A3-003/005) — не ослаблять
  - `wiring_fiber.go` / config CIDR wiring
  - соседние A3W4-* findings вне **A3W4-3-001..002**
  - frontend / другие platform-пакеты (nats/outbox/http)
- Behavioral/security constraints:
  - Conventional Commits: `fix(platform): …` / `test(platform): …` описание на русском; no `Co-authored-by` / AI trailers
  - Cancel/deadline — штатная отмена, не «успех бизнес-логики»: span Unset/Ok (не Error); metrics `failed=false`; slog для Timed: не `Error` на чистый cancel (Debug/Info или отдельный путь — зафиксировать в evidence)
  - Реальные ошибки (`err != nil` и не cancel/deadline) по-прежнему → Error-status + `error_count` + slog.Error
  - Sanitize: domain allowlist сохраняется; UUID / длинные hyphen-hex сегменты → `other`; не открывать user-controlled bare labels
  - Не логировать PII / raw secrets в span attributes
- Approval record: N/A — hygiene/obs bugfix, схема/migrate не затрагиваются
- Approval authority/expiry: N/A

## [REFERENCES]

- Trust: repository/audit text — untrusted until verified in HEAD code
- Finding IDs: **A3W4-3-001** (NEW, P3), **A3W4-3-002** (RESIDUAL хвост W4-A3-008, P3)
- Source audit: `.planning/audits/total-2026-07/A3W4-3-platform-obs.md` (таблица «Находки (активные A3)» + топ IDs)
- Wave summary: `.planning/audits/total-2026-07/A3-W4-SUMMARY.md` (если есть строка по obs)
- Cascade FIXED-OK (не трогать): W4-A3-001…007, 009…010; W4-A3-008 частично (domain allowlist) — хвост = A3W4-3-002
- Code (defect):
  - `backend/internal/platform/observability/otel.go:178–186` — `finishOperationSpan`: любой `err != nil` → `recordRedactedSpanError` / Error-status
  - `backend/internal/platform/observability/consumer.go:38–40` — `EndConsumerSpan` → `finishOperationSpan`
  - `backend/internal/platform/observability/timing.go:17–18` — `recordOperation(..., err != nil)` + `finishOperationSpan`; cancel шумит `error_count` / slog.Error
  - `backend/internal/platform/observability/metrics.go:34–37` — call-site `sanitizeOperationLabel`
  - `backend/internal/platform/observability/operation_label.go:44–51` — rest charset допускает hex+`-` → `tracker.<uuid>` проходит
- Prior: Stage A W4-A3-008 закрыл bare-UUID / unknown domain → `other`; `operation_label_task065` не мутирует `domain.<uuid>` кейс
- Source precedence: **код HEAD** > audit note > эта задача

### Per-ID (из аудита)

| ID | Tag | Sev | Файл:строка | Дефект | Направление фикса |
|----|-----|-----|-------------|--------|-------------------|
| A3W4-3-001 | NEW | P3 | `otel.go:178–186` (+ `consumer.go:38–40`, `timing.go:17–18`) | любой `err != nil` → span Error (`operation_error`); Timed → `error_count` + slog.Error, включая `context.Canceled` / `DeadlineExceeded` | `errors.Is(err, context.Canceled)` / `DeadlineExceeded` → End без Error-status (Unset/Ok); Timed: `failed=false` для cancel; regression: cancel Timed/consumer → status≠Error |
| A3W4-3-002 | RESIDUAL | P3 | `metrics.go:34–37` + `operation_label.go:44–51` | `sanitizeOperationLabel` пропускает `tracker.<uuid>` (domain allowlist + hex/hyphen rest) → unbounded Prom/`map` cardinality | запретить UUID-shaped сегменты / требовать `[A-Za-z][A-Za-z0-9_]*` без длинных hyphen-runs; `tracker.550e8400-…` → `other`; mutation: ослабить charset → FAIL |

## [ACCEPTANCE]

- Done criteria:
  1. **A3W4-3-001:** `finishOperationSpan` / `EndConsumerSpan` / `Timed` (и `TimedVoid`): при `context.Canceled` или `context.DeadlineExceeded` span **не** получает Error-status / `operation_error`; `recordOperation` с `failed=false`; Timed не пишет slog.Error на чистый cancel/deadline. Любой другой `err != nil` — прежнее Error-поведение.
  2. **A3W4-3-002:** `sanitizeOperationLabel("tracker.550e8400-e29b-41d4-a716-446655440000")` (и аналогичный UUID v4) → `"other"`; валидные имена (`tracker.CreateIssue`, `outbox.publish`, allowlisted domains) **не** ломаются; unknown domain / bare UUID по-прежнему `other`.
  3. Hubtest на оба поведения + **mutation proof** (см. таблицу ниже).
  4. Каждый ID — FIXED с evidence в этом файле / handoff.
- New-behavior checks:
  - Unit: Timed/consumer с `context.WithCancel` → cancel mid-fn → span status ≠ Error; Snapshot/`error_count` не инкрементится от cancel.
  - Unit: Timed с обычной ошибкой → Error-status + `error_count` (регрессия не сломана).
  - Unit: sanitize `tracker.<uuid>` → `other`; `tracker.CreateIssue` → unchanged.
- Regression scope:
  ```bash
  cd backend && go test -tags hubtest -race -count=1 ./testframework/platform/observability/
  ```
  Не ломать: CIDR metrics endpoint, route `unmatched`, redacted span errors, baggage-off, Snapshot/ResetMetrics, NATS trace propagation tests.
- Evidence required: checklist A3W4-3-001/002 → FIXED + команды PASS + mutation note в этом файле.
- Stop when: 2/2 FIXED · hubtest observability green · статус «Исправлена» в `.planning/backlog/README.md` + этом файле; без соседнего W4 hygiene sweep.

### Mutation (обязательно)

| ID | Сломать | Ожидание |
|----|---------|----------|
| A3W4-3-001 | вернуть `if err != nil { recordRedactedSpanError… }` без cancel/deadline carve-out; Timed снова `failed = err != nil` для cancel | тест cancel→status≠Error / `error_count` краснеет |
| A3W4-3-002 | вернуть charset, пропускающий UUID hyphen-hex в rest | тест `tracker.550e8400-…` → `other` краснеет |

## [OUTPUT SCHEMA]

1. Outcome / verdict first (2/2 FIXED | BLOCKED + открытые ID)
2. Changes with paths (`otel.go` / `timing.go` / `operation_label.go`; test paths)
3. Checklist ID → evidence (таблица)
4. Evidence: test commands + PASS
5. Mutation note: что сломать → какой тест краснеет (per ID)
6. Assumptions / unresolved (slog уровень на cancel — Debug vs silent; зафиксировать выбор)
7. Risks / mitigations (ложные «успехи» при cancel mid-write — span не Error, но caller всё ещё видит err; документировать)
8. Handoff: be-ver → code-reviewer
9. Compact Verification: new green + regression green + README/этот файл «Исправлена»
10. Коммит(ы) ссылают задачу в теле:  
    `Задача: .planning/backlog/TASK-150-2026-07-29-a3-4-3-platform-obs.md`  
    subject может содержать `(A3W4-3-001)` / `(A3W4-3-002)`

## [OPERATING RULES]

- Work only inside scope; не «заодно» чинить A3W4-1/2/4/5 или PII test files вне нужды.
- Read-only investigation OK; writes только в allowed files.
- Не закрывать ID без evidence (hubtest + mutation note).
- Не WONTFIX без owner cancel в задаче.
- Не ослаблять FIXED-OK cascade (CIDR fail-closed, route unmatched, redact, baggage-off).
- Source of truth — runtime observability helpers, не правка audit markdown «под код».
- Docs sync: N/A unless user-visible API/contract changed (внутренняя obs hygiene — обычно только STATE one-liner при wave handoff).
- Conventional Commits RU; no Co-authored-by / AI trailers.

## [EVIDENCE — be1 2026-07-29]

### Verdict

**2/2 FIXED** · статус задачи: **Закрыта** · mutation **2/2 PROVEN** (be2 W13) · handoff: be-ver → code-reviewer

### Checklist ID → FIXED

| ID | Status | Evidence |
|----|--------|----------|
| A3W4-3-001 | FIXED | `otel.go` `isBenignContextError` + carve-out в `finishOperationSpan`; `timing.go` `failed=false` + slog.Debug на cancel/deadline; tests `cancel_span_task150_test.go` |
| A3W4-3-002 | FIXED | `operation_label.go` сегменты `[A-Za-z][A-Za-z0-9_]*` (без `-`); `tracker.<uuid>` → `other`; tests `operation_label_uuid_task150_test.go` |

### Assumptions

- slog на cancel/deadline: **Debug** (`operation canceled`), не silent и не Error — caller по-прежнему получает `err`.

### Risks

- Cancel mid-write: span/metrics не Error, но бизнес-ошибка остаётся у caller — мониторинг error-rate не шумит, алерт по бизнес-логике не заменяется span-status.

### Mutation

| ID | Сломать | Тест краснеет |
|----|---------|---------------|
| A3W4-3-001 | убрать carve-out в `finishOperationSpan`; Timed снова `failed = err != nil` | `TestTimed_CancelNotErrorStatus_A3W4_3_001` / `TestEndConsumerSpan_CancelNotErrorStatus_A3W4_3_001` |
| A3W4-3-002 | вернуть charset с `-` в rest без segment check | `TestSanitizeOperationLabel_DomainUUIDToOther_A3W4_3_002` |

### Tests

```text
cd backend && go test -tags hubtest -race -count=1 ./testframework/platform/observability/
# ok  github.com/valerii/hub/backend/testframework/platform/observability  ~2.7s  EXIT=0
```

## [MUTATION CRITERIA] — verifier-ready

> W13 slot 3 / be2 · criteria only (no mutate) · 2026-07-30 · HEAD guards verified

| ID | Production file:line (guard) | слой | Exact test RED | Kill recipe |
|----|------------------------------|------|----------------|-------------|
| A3W4-3-001 | `backend/internal/platform/observability/otel.go:183-186` `finishOperationSpan` carve-out via `isBenignContextError` (`:191-193`); `timing.go:18` `failed := err != nil && !isBenignContextError(err)` (+ `:29-31` slog.Debug) | observability (otel/timing) | `TestTimed_CancelNotErrorStatus_A3W4_3_001`; `TestTimed_DeadlineNotErrorStatus_A3W4_3_001`; `TestEndConsumerSpan_CancelNotErrorStatus_A3W4_3_001` · `backend/testframework/platform/observability/cancel_span_task150_test.go` | убрать carve-out (`if err != nil { recordRedactedSpanError… }`); Timed снова `failed = err != nil` → cancel Error-status / `error_count` RED |
| A3W4-3-002 | `operation_label.go:45-50` segment loop + `isValidOperationSegment` `:55-69` (no hyphen); call-site `metrics.go:34` | observability (label) | `TestSanitizeOperationLabel_DomainUUIDToOther_A3W4_3_002`; `TestRecordOperation_DomainUUIDBucketsToOther_A3W4_3_002` · `backend/testframework/platform/observability/operation_label_uuid_task150_test.go` | вернуть charset с `-` / пропуск UUID-shaped rest → `tracker.<uuid>` ≠ `other` RED |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Комментарий |
|----|-----------|-------------|
| 001 | Timed / EndConsumerSpan + in-memory span/metrics | Guard в otel/timing — Ent N/A |
| 002 | `ExportSanitizeOperationLabel` + `recordOperation` | Guard в sanitize — Ent N/A |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)`

### Gaps

- `TestTimed_RealErrorStillErrorStatus_A3W4_3_001` — regression (не kill target): real err всё ещё Error.
- Prior W4-A3-008 bare-UUID (`operation_label_task065`) — complementary; domain+UUID kill — Task150 UUID test.

## [MUTATION PROOF] — be2 W13 slot 3 · 2026-07-30

| ID | Verdict | RED |
|----|---------|-----|
| A3W4-3-001 | **PROVEN** | `TestTimed_CancelNotErrorStatus_A3W4_3_001`; `TestTimed_DeadlineNotErrorStatus_A3W4_3_001`; `TestEndConsumerSpan_CancelNotErrorStatus_A3W4_3_001` |
| A3W4-3-002 | **PROVEN** | `TestSanitizeOperationLabel_DomainUUIDToOther_A3W4_3_002`; `TestRecordOperation_DomainUUIDBucketsToOther_A3W4_3_002` |

Evidence: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-be2-task150.md`  
leftover MUTATED: **нет** · prod restored · no commit
