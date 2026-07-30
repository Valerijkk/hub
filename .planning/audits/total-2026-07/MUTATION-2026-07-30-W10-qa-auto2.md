# MUTATION W10 slot 3 — TASK-120 (qa-auto2) · 2026-07-30

> Checker: **qa-auto2** (!= be2 fixer) · Env: localhost hubtest (mock + Ent sqlite; POSTGRES_TEST_URL set)  
> Criteria: `.planning/backlog/archive/2026-07-28/TASK-120-2026-07-28-integrations-secrets-imex-trash.md` `[MUTATION CRITERIA]`  
> Runner: `_run_mutations_w10_task120.py` · Evidence JSON: `_mut-batch-w10-task120.json`

## Verdict

**4/4 finding IDs PROVEN** · Break→RED→restore→GREEN · статус TASK-120 → **Закрыта**

| ID | Verdict | RED test | Mutation (BREAK) | Layer |
|----|---------|----------|------------------|-------|
| W7-A3-004 | **PROVEN** | `TestHandler_HandleWebhook_RejectsForeignSecret_A5_008` + `TestHandler_HandleWebhook_RejectsForeignToken_A5_009` | restore `ListAllWebhookSecrets` any-of in Verify* (GH+GL) | service (verify) + mock |
| W7-A3-006 | **PROVEN** | `TestTriggerPipeline_UpstreamBodyNotLeaked_W7A3006` + `TestCreateBranch_UpstreamBodyNotLeaked_W7A3006` | `NewInternal(... string(respBody/body))` | service (httptest) |
| W7-A3-007 | **PROVEN** | `TestService_BearerTokenEncryptedAtRest_W7A3007` | `encryptToken` return plaintext (skip EncryptString) | service |
| W7-A3-008 | **PROVEN** | `TestListProjectIssuesPlain_ExcludesTrash_W7A3008` | убрать `issue.DeletedAtIsNil()` | Ent repository |

Production restored (`git checkout --` после каждого BREAK). Weak: `TestEntRepository_WebhookSecretBoundToRepo_W7A3004` **остался PASS** при any-of — **N/A as PROVEN gate** (как в criteria Gaps).

## Per-ID evidence

### W7-A3-004 — webhook secret binding (no ListAll any-of)

```text
BASELINE A5_008 / A5_009 → PASS

BREAK github: temporary ListAllWebhookSecrets on Repository + mock;
  VerifyWebhookSignature → any-of secrets (ignore FindByRepo bind)
→ FAIL A5_008: expected 403 actual 200 (foreign secret accepted)
WEAK W7A3004 → still PASS (FindByRepo only — not a PROVEN gate)
RESTORE → PASS

BREAK gitlab: same any-of in VerifyWebhookToken
→ FAIL A5_009: expected 403 actual 200
RESTORE → PASS
```

### W7-A3-006 — GitLab upstream body redact

```text
BASELINE W7A3006 (pipeline + branch) → PASS

BREAK gitlab/service.go TriggerPipeline + CreateBranch:
  NewInternal includes string(respBody) / string(body)
→ FAIL both: err.Error() contains SECRET_* markers
RESTORE → PASS
```

### W7-A3-007 — 1C BearerToken at-rest encrypt

```text
BASELINE W7A3007 → PASS

BREAK onec/service.go encryptToken: return token, nil (skip cipher)
→ FAIL: system_setting JSON contains plaintext-1c-bearer-secret-W7A3007
RESTORE → PASS
```

### W7-A3-008 — imex ZIP path trash filter

```text
BASELINE W7A3008 → PASS

BREAK imex/repository.go ListProjectIssuesPlain: no DeletedAtIsNil
→ FAIL: Len==2 (trash issue included) expected 1
RESTORE → PASS
```

## Commands

```text
$env:POSTGRES_TEST_URL = "postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:PYTHONIOENCODING = "utf-8"
python .planning/audits/total-2026-07/_run_mutations_w10_task120.py
→ BASELINE OK · 4/4 PROVEN · wall ~141s
→ git diff prod empty after restore
```

## Assumptions / unresolved

- W7-A3-004 BREAK временно возвращал `ListAllWebhookSecrets` в interface + Ent + mocks (метод удалён в TASK-120) — только для prove; restore через `git checkout`.
- `TestEntRepository_WebhookSecretBoundToRepo_W7A3004` — weak, не PROVEN gate (documented Gaps).
- Коммит не создавался (инструкция слота).

## Handoff

TASK-120 **Закрыта** · evidence этот файл · README синхронизирован · mutation gate закрыт · next: qa-auto-ver / wave control.
