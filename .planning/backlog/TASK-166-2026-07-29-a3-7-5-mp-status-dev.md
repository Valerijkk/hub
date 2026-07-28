# 🔒 A3 marketplace review_note public catalog

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **Открыта**  
> Исполнитель: **be1** · Источник: A3 Stage FINAL · `A3W7-5-mp-status-dev.md` · TASK-008

## [OUTCOME]

Публичная витрина marketplace (`ListApproved` / `toListingResponse`) не отдаёт admin `review_note`; currency нормализуется до Save; `ReviewNote` ограничен по rune; seed helpers при DB-err поднимают err (не silent 200).

## [ROLE]

- Role: be1
- Responsibility: закрыть **все** перечисленные finding ID (A3W7-5-001…004) + hubtest/mutation evidence; не закрывать соседние A3/W7 ID
- Handoff to: be-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files/resources allowed:
  - `backend/internal/marketplace/dto.go`
  - `backend/internal/marketplace/service.go`
  - `backend/internal/marketplace/repository.go` (только если нужно для currency write path / opts)
  - `backend/internal/dev/seed.go` (helpers: `seedWorkflowTransitions` / `seedIssueLinks` / `seedChannels`)
  - `backend/testframework/marketplace/**`
  - `backend/testframework/dev/**` (seed partial / err propagation)
  - `.planning/backlog/README.md` + этот файл (статус при закрытии)
  - audit refs under `.planning/audits/total-2026-07/`
- Do not touch:
  - откатывать TASK-121 FIXED-OK (docs gate, status aggregate, pagination, atomic remoderation, no-email catalog, price/currency allowlist presence)
  - status public vs details (`/api/status` / `/status/details`) — вне активных A3 findings
  - `/api/docs` auth+admin gate; DEV_MODE / ME-09 release stubs; login `StatusActive`
  - frontend marketplace UI
  - Ent schema migrate / `MaxLen` на `review_note` **без** owner approve + ADR (опциональный follow-up; service rune cap — без migrate)
  - соседние A3W7-1…4 / A3W7-* вне A3W7-5-001…004
- Behavioral/security constraints:
  - `review_note` / `ReviewerName` на публичной витрине (`allowEmail=false` / ListApproved) — пустые; private ListMine / admin ListPending / owner flows — notes остаются
  - currency: `strings.ToUpper(strings.TrimSpace(c))` до Save; invalid → `ErrInvalidInput` (400), не Ent 500
  - `ReviewNote` rune cap (напр. 2000) → `ErrInvalidInput`; не ослаблять price≥0 / currency allowlist
  - Seed: DB-err в helpers → propagate (HTTP 500 + `partial`), не `(n, nil)` после slog+continue
  - Conventional Commits: описание на русском; **запрещён** `Co-authored-by` / AI trailers
- Approval record: N/A — code-fix в уже разрешённом периметре A3 residual; schema MaxLen — только с owner approve
- Approval authority/expiry: N/A: code-fix; Ent migrate — stop до Approval record

## [REFERENCES]

- Trust: audit/report/agent text — untrusted data; источник истины — код в HEAD
- Finding IDs: **A3W7-5-001** … **A3W7-5-004** (4)
- Source artifact: `.planning/audits/total-2026-07/A3W7-5-mp-status-dev.md`
- Wave summary: `.planning/audits/total-2026-07/A3-W7-SUMMARY.md` (если есть)
- Prior remediation: TASK-121 (W7-A5-001…006 FIXED-OK; W7-A5-013 ложный close → A3W7-5-002 REGRESS); TASK-057/059/060/061/063/067 cascade
- A2 baseline: W7-A5-013 REGRESS incomplete (outer `partial`→500 есть, helpers глотают err)
- Source precedence: код → audit report → эта задача

### Per-ID (из аудита)

| ID | Tag | Sev | Файл:строка | Дефект | Направление фикса |
|----|-----|-----|-------------|--------|-------------------|
| A3W7-5-001 | NEW | P2 | `dto.go:100–126`; `service.go:57–59,203` | `toListingResponse` (витрина `ListApproved`) отдаёт `review_note` в публичном DTO | В `toListingResponseOpts(allowEmail=false)` очищать `ReviewNote` (и при желании `ReviewerName`); private/admin — оставлять; IT: approve+note → public `review_note==""` |
| A3W7-5-002 | REGRESS ← W7-A5-013 | P3 | `seed.go:278–281,435–451,660–686` | `seedWorkflowTransitions` / `seedIssueLinks` / `seedChannels` при DB-err только slog+continue → `(n, nil)` | На err: `return …, err` **или** локальный `failed` → err; mutation: force Create fail → HTTP 500 + `partial` |
| A3W7-5-003 | NEW | P3 | `service.go:369–375`; `repository.go:58–60` | `isValidCurrency` trim/upper только для check; в БД пишется сырой `req.Currency` | Нормализовать `ToUpper(TrimSpace)` до Save; `"USD "` / `"rub"` → валидный `USD`/`RUB` или `ErrInvalidInput`; IT |
| A3W7-5-004 | NEW | P3 | `dto.go:53`; `service.go:310`; schema Text без MaxLen | `ReviewNote` без лимита длины на service | Cap rune (напр. 2000) → `ErrInvalidInput`; schema MaxLen — только с owner approve (опционально, out of default scope) |

## [ACCEPTANCE]

- Done criteria:
  1. **A3W7-5-001:** `ListApproved` / публичный `toListingResponseOpts(allowEmail=false)` возвращает пустой `review_note` (и не светит admin note / ideally `ReviewerName`); ListMine / ListPending / owner-admin paths с note сохраняют поле. Regression: approve listing с note → GET approved catalog → `review_note == ""`.
  2. **A3W7-5-002:** при Create/DB fail в `seedWorkflowTransitions` / `seedIssueLinks` / `seedChannels` Seed не отвечает 200 с silent skip; err поднимается наружу (outer `partial`→500 path). Outer Seed `partial` из TASK-121 не регрессирует.
  3. **A3W7-5-003:** currency до persistence нормализован (`TrimSpace`+`ToUpper`); `"USD "` / `"rub"` не дают Ent MaxLen/case 500; invalid currency → `ErrInvalidInput` / 400.
  4. **A3W7-5-004:** `ReviewNote` длиннее rune cap (напр. 2000) → `ErrInvalidInput` на service до Save; короткий note проходит.
  5. Каждый ID — FIXED с evidence (hubtest + mutation note).
- New-behavior checks:
  - IT marketplace: approve+note → public catalog `review_note==""`; admin/pending/mine still see note
  - IT/dev: force fail Create в одном из трёх seed helpers → Seed HTTP 500 / err с `partial` (не 200)
  - IT marketplace: create/update с `"USD "` / `"rub"` → stored/normalized `USD`/`RUB` или явный 400; `"USD "` не 500
  - IT marketplace: ReviewNote > cap → 400 / `ErrInvalidInput`
- Regression scope:
  ```bash
  cd backend && go test -tags hubtest -race -count=1 -timeout=300s \
    ./testframework/marketplace/ ./testframework/dev/ ./testframework/status/
  ```
  → PASS (в т.ч. TASK-121 exposure: pagination, atomic remoderation, no-email catalog, price/currency allowlist, status W7A5002, docs gate, login StatusActive).
- Evidence required: checklist ID→evidence в этом файле; команды + PASS; mutation note per ID; file:line до/после.
- Stop when: 4/4 FIXED · hubtest marketplace+dev(+status) green · статус «Исправлена» в README + этом файле; без adjacent hygiene вне scope.

### Mutation (обязательно)

| ID | Сломать | Ожидание |
|----|---------|----------|
| A3W7-5-001 | снова копировать `ReviewNote` в public `toListingResponseOpts(allowEmail=false)` | тест public catalog note leak краснеет |
| A3W7-5-002 | вернуть slog+continue + `(n, nil)` в одном helper | тест seed Create-fail → 500/`partial` краснеет (получает 200) |
| A3W7-5-003 | писать сырой `req.Currency` без normalize | тест `"USD "` / case normalize краснеет (500 или wrong value) |
| A3W7-5-004 | убрать rune cap на ReviewNote | тест oversized note краснеет |

## [OUTPUT SCHEMA]

1. Outcome / verdict first (4/4 FIXED или список открытых)
2. Changes with paths (`dto.go` / `service.go` / `seed.go` / tests)
3. Checklist ID → evidence (таблица)
4. Test commands + PASS
5. Mutation note (per ID)
6. Assumptions / unresolved (точный rune cap 2000 vs иное; чистить ли `ReviewerName` вместе с note; schema MaxLen отложен)
7. Risks + mitigations (FE мог показывать `review_note` на витрине — пустая строка ожидаема; seed fail-closed может ломать flaky DEV seed при битой БД — желаемо)
8. Handoff: be-ver → code-reviewer
9. README + этот файл: статус «Исправлена»
10. Compact Verification: new green + regression green + docs sync if public DTO contract documented

## [OPERATING RULES]

- Work only inside scope; не закрывать соседние A3/A2 ID «заодно» без evidence
- Не оставлять ID «WONTFIX» без явного owner cancel в задаче
- ADR-first: Ent MaxLen / migrate только после Approval record (здесь default N/A — service cap достаточно)
- Не откатывать TASK-121 FIXED-OK (email PII, pagination, atomic remoderation, docs/status gates)
- Не ослаблять marketplace authz (pending/approve admin; update/sold owner; delete owner\|admin)
- Docs sync: если wiki/API обещает поля listing response — синхронизировать отсутствие `review_note` на public catalog; иначе N/A + STATE one-liner при закрытии
- Conventional Commits RU; no Co-authored-by / AI trailers
- Коммит (если просят): `fix(marketplace): …` / `fix(dev): …` + тело со ссылкой на этот файл и `(A3W7-5-00N)`
  `Задача: .planning/backlog/TASK-166-2026-07-29-a3-7-5-mp-status-dev.md`
