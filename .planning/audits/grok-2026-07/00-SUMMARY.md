# Hub Audit Summary — grok-2026-07

**Дата синтеза:** 2026-07-26  
**Ветка:** `cursor-grok-4.5`  
**Источники:** `01`–`19` (волна зон + verification + cross-cutting + red-team)  
**Режим:** READ-ONLY product code; артефакт — только этот summary  
**Верификация:** `17-verification.md` применена (CONFIRMED / DOWNGRADED_TO_P2; **FALSE_POSITIVE = 0**)

---

## Итог после дедупа + verification

| Severity | Count (unique) |
|----------|---------------:|
| **P0** | **13** |
| **P1** | **77** |
| **P2** | **119** |
| **P3** | **52** |
| **Всего** | **261** |

| Метрика | Значение |
|---------|----------|
| Raw findings A01–A16 | ~282 строк |
| Исключено как дубли (A9 rehash + cross-zone aliases) | 21 |
| DOWNGRADED_TO_P2 (A17) | 8 (A8-001/002, A11-003, A12-004…007, A16-004) |
| FALSE_POSITIVE | **0** |
| Composite amplifications (A18, не +к unique root) | 20 |
| Red-team chains (A19 → existing IDs) | 12 |

**Вердикт волны:** системный **Broken Access Control (OWASP A01)** внутри org + сломанный Ent codegen (тесты/gate не runnable) + несколько мёртвых FE↔BE контрактов. Периметр (SSRF, LiveKit webhook, Magic Link RL, bases owner-scope, WS room authz) держится.

---

## 1. Единый реестр находок (дедуплицированный)

Правила дедупа:
- Канон = первичный ID зоны домена (A1–A8, A10–A16).
- A9 `= Ax-yyy` → alias, не отдельный дефект.
- Cross-zone aliases (A13-009=A7-017, A13-010=A3-020, A13-014=A10-008, A14-011=A8-007, A9-015→A14-005) свёрнуты.
- Severity после A17: DOWNGRADED отмечены явно.
- A18 = комбинации (ссылки на root IDs); A19 = attack chains на те же roots.

### 1.1 P0 — уникальные (13) · все CONFIRMED (A17)

| Canonical ID | Aliases / amplify | Проблема | Где |
|--------------|-------------------|----------|-----|
| **A1-001** | A9-001, RT-001, A18-001 | Pending-user → JWT (ORG-06 bypass) | `auth/service_login.go:64`, `middleware.go:128` |
| **A1-002** | A9-002, RT-004, A18-002 | Cross-project role Update/Delete IDOR | `rbac/custom_roles.go:73-84` |
| **A1-003** | A9-003, RT-004, A18-002 | Assign чужой роли → escalation | `rbac/custom_roles.go:126-132` |
| **A1-004** | A9-011 (часть) | GDPR hard-delete без защиты founder | `audit/service.go:379-441` |
| **A2-001** | A9-006, RT-002, A18-001 | Messenger search без membership (все каналы) | `messenger/service_search.go:11-29` |
| **A3-001** | A9-005, RT-005, A18-003 | CompleteSprint переносит чужие issues в свой project | `sprint/service.go:107-149` |
| **A5-001** | A18-011 | CSV import: mappings всегда `nil` → все строки skip | `imex/handler.go:80` |
| **A6-001** | A9-004, RT-003, A18-004 | CRM Update/Delete без owner → + won→invoice | `crm/service.go:80-214` |
| **A6-002** | A18-004 | `MarkInvoicePaid` вне tx → orphan journals / double post | `accounting/service_invoice.go:116-168` |
| **A11-001** | A18-019 | FE `POST /projects/join` ≠ BE `/invites/join` (onboarding 404) | `frontend/src/api/projects.ts:133-135` |
| **A11-002** | A18-011 | TMS FormData → `JSON.stringify` → import мёртв | `frontend/src/api/tms.ts:681-691` |
| **A12-001** | A18-005 | `ent.init` panic (`DefaultAttempts` cast) — hubtest/ent pkgs FAIL | `ent/runtime.go:1735-1737` |
| **A15-001** | — | Planning P0: STATE 125/125, ROADMAP без 65–125 / Progress→17 | `STATE.md` vs `ROADMAP.md` |

**Downgraded from P0:** A11-003 → **P2** (TelegramButton нигде не смонтирован).

### 1.2 P1 — уникальные (77) · CONFIRMED; агрегаты A9 не считаются отдельно

Сгруппировано по классам; полный список ID сохранён.

#### Auth / RBAC / Audit / Founder
| IDs | Суть |
|-----|------|
| A1-005, A1-006, A7-002, A18-007 | Founder ≠ admin-bypass (RBAC, RequireAdmin, dashboard admin) |
| A1-007, A18-009 | Resend fail → HTTP 200 «sent» |
| A1-008, A2-010, A3-006, A18-008 | JetStream Ack-on-error (audit / notification / automation) |
| A1-009 | Ban/Deactivate founder |
| A9-012, A18-020 | APIKey пишет `userID`, RequireAuth всё равно требует JWT |

#### Realtime / Calls / Push
| IDs | Суть |
|-----|------|
| A2-002…A2-006, RT-008/009 | CreateCall / chat / recording / RaiseHand без membership; push hijack |
| A2-007, A18-012 | Instant-call WS `call:{id}` всегда deny |

#### Tracker / TMS / Sprint
| IDs | Суть |
|-----|------|
| A3-002, A3-003, A3-007…A3-012, RT-007 | Resource-by-key/UUID без project scope (issues, TC, runs, results, comments, Allure, import) |
| A3-004, A6-003, A5-011→P2 class, A18-006 | Optimistic lock сломан / optional `version:0` |
| A3-005 | LogTime + StopTimer → nil panic |

#### Search / Storage / AI
| IDs | Суть |
|-----|------|
| A4-001…A4-006 | JQL assignee dead; soft-delete в Ctrl+K; saved filters без membership; Ollama vs SafeDial; Confirm без status |

#### Integrations
| IDs | Суть |
|-----|------|
| A5-002, A5-010, RT-006, A18-014 | imex auth-only + PDF unscoping |
| A5-003…A5-009, RT-012 | GH/GL delete/trigger/CreateBranch IDOR; webhook any-secret |

#### Business / Calendar / Dashboard / Backup / Mail
| IDs | Суть |
|-----|------|
| A6-004, A6-005, RT-010 | Duplicate invoice-per-deal; leave org-wide leak |
| A7-001, RT-011, A18-015 | Dashboard project stats IDOR |
| A7-003, A7-004 | Restore без maintenance; failed backup → success HTTP |
| A7-005 | Mail SendAlert stub при живом Resend key |
| A7-006 | Coffee rematch Offset(1) off-by-one |

#### Data layer / CMD (после downgrade индексов)
| IDs | Суть |
|-----|------|
| A8-003, A8-004 | Project / ~80% схем без `version` (TZ) |
| A8-005, A8-006 | Channel/Comment hard-delete vs soft-delete TZ |
| A8-007 (=A14-011), A8-008 | Mongo без Disconnect; Listen `os.Exit` минуя cleanup |

#### Frontend / API contract
| IDs | Суть |
|-----|------|
| A10-001…A10-004, A18-012/013 | Channel store↔Query drift; issues cursor; WS no reconnect; presence mutate |
| A11-004…A11-007, A18-010/011 | Dead telegram client; notify-mode 404; birthday date; TMS blob export |

#### Tests (после downgrade coverage-gaps)
| IDs | Суть |
|-----|------|
| A12-002, A12-003 | Assert-free «named» tests |

#### Perf hot paths
| IDs | Суть |
|-----|------|
| A13-001…A13-006 | Unbounded board/bases/dashboard/sprint; Gantt N+1; notification fan-out |

#### Observability / ops
| IDs | Суть |
|-----|------|
| A14-001…A14-005 | Health=liveness+deps; SeaweedFS committed secrets; NATS no auth; monitoring ports; open `/metrics` |

#### Planning / REQ
| IDs | Суть |
|-----|------|
| A15-002, A15-003, A15-004 | Phase 39 gap; PROJECT.md stale; LOGN-05 vs TZ Telegram |

**Downgraded P1→P2:** A8-001, A8-002 (indexes), A12-004…007 (test gaps/strategy), A16-004 (Fiber beta pin).

### 1.3 P2 — уникальные (119) · выборочный реестр классов + ID

| Класс | IDs (неполный перечень всех 119 — см. zone reports) |
|-------|-----------------------------------------------------|
| Auth/RBAC residual | A1-010…A1-013 |
| Realtime residual | A2-008…A2-012 |
| Tracker/TMS/Sprint | A3-013…A3-020 (=A13-010) |
| AI/Search/Storage/Bases | A4-007…A4-018; A4-014↔A13-002 amplify |
| Integrations | A5-011…A5-015 |
| CRM/HR/Accounting | A6-006…A6-010 |
| Infra domains | A7-007…A7-012, A7-017(=A13-009) |
| Schema/wiring | A8-001↓, A8-002↓, A8-009…A8-018 |
| Security unique | A9-014, A9-016, A9-017(=A4-008) |
| Frontend | A10-005…A10-012, A10-008(=A13-014) |
| API contract | A11-003↓, A11-008…A11-013 |
| Tests | A12-004↓…A12-007↓, A12-008…A12-014 |
| Perf | A13-007…A13-019 (минус aliases) |
| Observability | A14-006…A14-015 |
| REQ drift | A15-005…A15-007 |
| Hygiene | A16-001…A16-003, A16-004↓, A16-005…A16-012, A16-017 |
| Cross-cutting P2 | A18-017…A18-020 |

### 1.4 P3 — уникальные (52)

Остаточные: rune/byte truncate, layering, N+1 cosmetic, public docs, livekit local keys, `:latest` tags, thin assert-free smokes, naming drift, ECC TODO noise — IDs **A1-014, A2-013(=A9-020), A2-014, A3-021…023, A4-019…022, A5-016…018, A6-011…014, A7-013…017, A8-019…021, A9-019(=A11-004 surface), A9-021…023, A10-013/014, A11-014, A12-015/016, A13-020…025, A14-016…019, A15-008, A16-011, A16-013…016, A16-018…020**.

### 1.5 Composite (A18) — не увеличивают unique count

A18-001…005 **P0 amplify**; A18-006…016 **P1**; A18-017…020 **P2**. Все ссылаются на canonical IDs выше.

### 1.6 Red-team (A19) → map

| RT | Severity | Maps to |
|----|----------|---------|
| RT-001…005 | CRITICAL | A1-001, A2-001, A6-001, A1-002/003, A3-001 |
| RT-006…012 | HIGH | A5-002/010, A3-002, A2-002…004, A2-005, A6-005, A7-001, A5-006 |

Exploitability score A19: **8.5/10** (insider member).

---

## 2. Сводная таблица оценок по зонам

| Zone | File | Score | P0 | P1† | P2 | P3 | Notes |
|------|------|------:|---:|----:|---:|---:|-------|
| A01 Auth/RBAC/Audit/Platform | `01-auth-rbac.md` | **5/10** | 4 | 5 | 4 | 1 | Pending JWT + role IDOR |
| A02 Messenger/WS/Calls/Notif | `02-realtime.md` | **5/10** | 1 | 6 | 5 | 2 | Search P0; calls authz |
| A03 TMS/Tracker/Project/Sprint | `03-tracking.md` | **5.5/10** | 1 | 11 | 8 | 3 | Sprint theft + key IDOR |
| A04 AI/Search/Bases/Storage | `04-ai-search-storage.md` | **6.5/10** | 0 | 6 | 12 | 4 | JQL/filters/confirm |
| A05 GH/GL/Imex/Marketplace | `05-integrations.md` | **5.5/10** | 1 | 9 | 5 | 3 | CSV dead + imex RBAC |
| A06 CRM/HR/Acct/Calendar | `06-business.md` | **4.5/10** | 2 | 3 | 5 | 4 | CRM IDOR + pay race |
| A07 Gateway/Mail/Intranet/… | `07-infra-domains.md` | **5/10** | 0 | 6 | 6 | 5 | Dashboard IDOR; backup |
| A08 Ent/CMD/Data | `08-data-layer.md` | **6/10** | 0 | 6‡ | 12‡ | 3 | version/soft-delete gaps |
| A09 Security cross-cut | `09-security.md` | **5/10** | 0* | 1* | 3* | 4* | *unique only; rest = aliases |
| A10 Frontend | `10-frontend.md` | **6.5/10** | 0 | 4 | 8 | 2 | WS/reconnect/pagination |
| A11 API contract | `11-api-contract.md` | **5/10** | 2‡ | 4 | 7‡ | 1 | Join+TMS P0; TG↓P2 |
| A12 Tests/Coverage | `12-tests.md` | **4/10** | 1 | 2‡ | 11‡ | 2 | ent panic; many↓P2 |
| A13 Performance | `13-perf.md` | **5/10** | 0 | 6 | 10‡ | 6 | Hot unbounded paths |
| A14 Observability/Ops | `14-observability.md` | **6/10** | 0 | 5 | 9‡ | 4 | Prod compose holes |
| A15 Requirements vs code | `15-requirements-vs-code.md` | **7.2/10** | 1 | 3 | 3 | 1 | Planning drift |
| A16 Hygiene | `16-hygiene.md` | **7/10** | 0 | 0‡ | 12‡ | 8 | Docs/rules drift |
| A17 Verification | `17-verification.md` | **8/10** | — | — | — | — | 0 FP; 8 downgrades |
| A18 Cross-cutting | `18-cross-cutting.md` | **3.5/10** | 5c | 11c | 4c | 0 | c=composite |
| A19 Red team | `19-red-team.md` | **8.5/10** | 5ch | 7ch | — | — | Exploitability |

† Counts in zone columns = as originally reported (до глобального дедупа).  
‡ После A17 downgrades внутри зоны.  
\* A09 unique after alias strip.  
**Среднее A01–A16 (арифм.): ~5.6/10.** Cross-cutting (A18) тянет системную оценку ниже.

---

## 3. Топ-15 проблем (влияние × вероятность / стоимость починки)

Формула: **Score = Impact(1–10) × Probability(1–10) / FixCost(1–10)**; выше = чинить раньше.

| # | ID | Score | I×P/C | Почему в топе |
|---|-----|------:|-------|---------------|
| 1 | **A11-001** | **90** | 9×10/1 | Onboarding join всегда 404; one-line path fix |
| 2 | **A11-002** | **45** | 9×10/2 | TMS import из UI полностью мёртв; multipart helper уже есть в tracker |
| 3 | **A12-001** | **40** | 10×10/2.5 | Блокирует почти все hubtest + risk server init; `go generate` |
| 4 | **A1-001** | **40** | 10×8/2 | Pending→JWT открывает ВСЕ IDOR (A18-001 / RT-001) |
| 5 | **A5-001** | **35** | 7×10/2 | CSV import HTTP path мёртв; mappings/`auto-map` |
| 6 | **A2-001** | **30** | 10×9/3 | Org-wide message exfil любым JWT (RT-002) |
| 7 | **A5-002** | **32** | 8×8/2 | imex без RBAC — полный export чужого проекта |
| 8 | **A6-001** | **27** | 9×9/3 | CRM PII mutate + invoice side-effect (RT-003) |
| 9 | **A1-002+A1-003** | **23** | 10×7/3 | Role IDOR → privilege escalation (RT-004) |
| 10 | **A3-001** | **21** | 9×7/3 | Кража задач между проектами (RT-005) |
| 11 | **A3-002** | **19** | 8×7/3 | Issue key IDOR (класс A3-00x / RT-007) |
| 12 | **A11-006** | **18** | 6×9/3 | Birthday profile 400 — частый user-facing fail |
| 13 | **A7-001** | **18** | 6×9/2 | Dashboard stats IDOR — разведка (RT-011) |
| 14 | **A6-002** | **15** | 9×5/3 | Двойные проводки при concurrent pay |
| 15 | **A1-004** | **15** | 10×3/2 | GDPR delete founder — редкий, но необратим |

**Honorable (чуть ниже порога топ-15):** A2-002…004 (calls), A5-008/009 (webhook any-secret), A14-002 (SeaweedFS secrets), A10-003 (WS reconnect), A8-004 (version mixin), A18-008 (Ack triad).

---

## 4. План работ

### 4.1 Чинить немедленно (blocker / P0 runtime + gate)

1. **A12-001** — regenerate Ent; зелёный `go test -tags hubtest` smoke.  
2. **A1-001** — нет сессии при `status ≠ active`; RequireAuth/Refresh fail-closed.  
3. **A1-002 / A1-003** — scope role ↔ `:projectId`; deny system on update/assign.  
4. **A2-001** — messenger search user-scope / `IsMember`.  
5. **A3-001** — sprint ops scoped to `projectID`.  
6. **A6-001** — `authorizeContact` / `authorizeDeal`.  
7. **A6-002** — atomic `MarkInvoicePaid` + unique deal invoice.  
8. **A5-001 + A5-002** — CSV mappings + RBAC на imex.  
9. **A11-001 / A11-002** — join path + TMS multipart.  
10. **A1-004** — deny GDPR/ban на founder.

*Критерий выхода:* RT-001…005 «неудавшиеся»; pending verify → 403; hubtest auth/rbac/messenger/crm/sprint green.

### 4.2 В этот спринт (P1 security/reliability + сломанные UX)

| Batch | IDs | Owner hint |
|-------|-----|------------|
| IDOR sweep | A3-002…012, A5-003…010, A7-001, A4-003/004 | be1/be2 |
| Calls/push | A2-002…007, A2-005 | be2 |
| Founder unify | A1-005/006, A7-002, A16-005 | be1 |
| Ack-on-error | A1-008, A2-010, A3-006 | be1 |
| Optimistic lock | A3-004, A6-003, A3-005 (timer panic) | be2 |
| Storage/search | A4-001/002/006, A4-007/008 | be2 |
| FE contract/UX | A10-001…004, A11-004…007, A18-012 | fe1/fe2 |
| Ops prod | A14-001…005, A7-003/004, A8-007/008 | devops |
| Mail/auth delivery | A1-007, A7-005, A14-006 | be1 + devops |
| Calendar leave | A6-005 | be2 |
| Perf caps | A13-001…006, A8-001/002 (indexes) | be2 |

### 4.3 В бэклог (P2/P3 + planning hygiene)

- Soft-delete Channel/Comment; VersionMixin rollout (A8-003…006, A8-004).  
- Presign size / NF-09 accepted-risk ADR (A4-008).  
- FE QueryStateView rollout, virtualize Kanban, lazy shell/i18n (A10-*, A13-014…017).  
- Test coverage gaps settings/calls (A12-004↓…007↓), assert-free cleanup.  
- Docs sync: ROADMAP 65–125 / Phase 39 / PROJECT.md / LOGN-05 / README routes / architecture Top-5 (A15-*, A16-007…010).  
- Fiber stable track (A16-004↓), pin `"latest"` deps, remove jsdom.  
- Monitoring ports / NATS auth / templated seaweedfs (остаток A14).  
- A11-003 dead TelegramButton — удалить файл+DTO+LoginWithTelegram*.

---

## 5. Чему верить нельзя

Где **docs / `.planning` / KPI расходятся с кодом** (не использовать как источник правды без сверки):

| Claim / документ | Реальность в коде | ID |
|------------------|-------------------|-----|
| COVERAGE / AGENTS: **COV 83%**, gate PASS, security 6/6 | На ветке `ent.init` panic → hubtest с Ent **не runnable**; % **stale** (baseline 2026-06-22) | A12-001, A18-005 |
| STATE: **125/125 · 100%** как «единый ROADMAP» | Каталоги фаз есть; корневой `ROADMAP.md` обрывается на Phase 17 / checklist без 65–125 и без Phase 39 | A15-001, A15-002 |
| `PROJECT.md` почти весь `[ ]` | Модули v1–v4 **уже в коде** — ложный backlog gap | A15-003 |
| REQ **LOGN-05** / ROADMAP Phase 1: Magic Link **+ Telegram login** | TZ + routes: Telegram **не** login; UI login без TG; dead service/API | A15-004, A11-004, A9-019 |
| TZ error shape `{"error","request_id"}` | Факт BE↔FE: `code` + `message` + `request_id` (`AppError`) | A11-014 |
| TZ: `version` на **каждой** изменяемой схеме | ~14/82 схем с `version`; Project/User/Channel/Sprint/… без | A8-003, A8-004 |
| TZ soft-delete / корзина 30д | Channel/Comment (и др.) — **hard delete** | A8-005, A8-006 |
| `mail` / `ResendReady` / Broadcast `Sent++` | SendAlert — **stub** при непустом ключе | A7-005 |
| Magic Link «код в логах при сбое Resend» | Логируется email/error, **не** код; HTTP 200 | A1-007 |
| Notification consumer «ACK только после успеха» | ACK всегда (handlers глотают err) | A2-010 |
| Maintenance mode на restore | Только log; live `pg_restore --clean` | A7-003 |
| `/api/health` как K8s **liveness** | 503 при деградации deps → restart storm semantics | A14-001 |
| Prod SeaweedFS + `validate()` non-dev | Committed `hub`/`hubsecret` в `seaweedfs-s3.json` **запрещены** validate | A14-002 |
| OTEL в prod (`.env.example` обещает) | Prod compose **без** `OTEL_EXPORTER_OTLP_ENDPOINT` | A14-009 |
| README: **39** routes, **495** tests; rules: Go **1.24**, **11** stores | Факт: ≫39 route files, **499** tests, Go **1.25**, **13** stores | A16-007…009 |
| `docs/architecture.md` Top-5 / TD на `main.go:540`, rbac `*ent.Client` | `main.go` ≈34 строки + wiring_*; rbac без Ent в middleware | A16-010 |
| Module Summary calls E2E «⚠️» как покрытие | Functional e2e/vitest **нет** (только a11y) | A12-005↓ |
| «API keys работают для CI» | X-API-Key не проходит RequireAuth без JWT | A9-012 |
| Offline/Ollama AI | SafeDial блокирует loopback | A4-005 |

**Можно относительно доверять (сверка A17/«Чисто» зон):** Magic Link-only routes, JWT HS256+refresh rotation, CORS fail-start, SafeHTTP/SSRF на основных outbound, LiveKit webhook sha256, bases owner-scope, storage path sanitize+CR-04, WS channel/project subscribe authz, Fiber v3 RBAC argument order на tracker/tms, inventory 499 TF / 0 internal tests / Vitest 99/99.

---

## Приложение A — Verification snapshot (A17)

| Verdict | Count |
|---------|------:|
| CONFIRMED (same severity) | 101 / 110 P0·P1 rows |
| DOWNGRADED_TO_P2 | 8 listed (report header said 9; explicit table has 8) |
| FALSE_POSITIVE | **0** |
| UPGRADED / NEEDS_INFO | 0 |

Качество волны 1–2 (A17): **8/10**.

## Приложение B — Рекомендуемый порядок фикс-волн

```
W0  A12-001 (ent generate)           → разблокировать тесты
W1  A1-001 + A1-002/003 + A2-001     → срезать RT-001/002/004
W2  A3-001 + A6-001/002 + A5-001/002 → срезать RT-003/005/006
W3  A11-001/002 + IDOR sweep A3/A5   → UX + lateral movement
W4  Ack triad + founder unify + ops  → надёжность
W5  FE WS/pagination + docs sync     → hygiene
```

---

*Синтез A20 · Grok · 2026-07-26 · product code не изменялся · path: `.planning/audits/grok-2026-07/00-SUMMARY.md`*
