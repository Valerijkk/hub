# DISCUSS — Hub Roadmap 2026 (GSD CONTEXT)

> **Режим:** discuss-phase · **Дата:** 2026-06-21  
> **Цель:** зафиксировать решения для planner/researcher без повторных вопросов  
> **Триггер:** обсуждение всех планов + **Magic Link only auth** + microservices/market refresh

---

## 1. Карта планов (что уже есть)

| Документ | Горизонт | Статус | Суть |
|----------|----------|--------|------|
| `PLAN-CLOSE-GAPS.md` | v2.0.1→v2.5 | Waves 1–6 largely done | Security, coverage 50%, e2e, observability |
| `IMPROVEMENT-MASTER-PLAN.md` | v2.0.2→v2.6 | UX/C done, PRODUCT planned | Waves A–C ✅, PRODUCT next |
| `FEATURE-BACKLOG-2026.md` | v2.5→v4 | Active | FB-01..28, MOD-01..08, AI-01..07, NOTIF-01..10 |
| `PROD-01-SPIKE.md` | v2.6 | Design ready | Workflow automation |
| `../specs/AI-PROVIDER-HUB-SPEC.md` | v3.0–v3.5 | Design ready | Multi-model AI, OpenRouter |
| `../specs/NOTIF-TG-EMAIL-SPEC.md` | v2.5–v2.6 | Design ready | TG bind + email alerts + run_failed |
| `../specs/PLATFORM-VISION-v4.md` | v4 2027–28 | Vision | Mighty/Bitrix super-app |
| `MICROSERVICES-ROADMAP.md` | v3–v4.5 | Strategy | Selective split |
| `MILESTONE-3.md` | v3 AI | Requirements | M3-001..030 |
| `EXECUTIVE-REPORT-2026.md` | Summary | Current | Score 93→98 path |
| `AUDIT-SCORECARD.md` | Now | v2.0.2 shipped | 93/100 |

**Единая timeline (рекомендация):**

```
SHIPPED v2.0.2 (93/100)
  │
  ├─ v2.5  Hardening     FB-07..14, NOTIF-02..04, AUTH-01 (magic-only)
  ├─ v2.6  Differentiation PROD-01..04, NOTIF-05..08, flaky
  ├─ v3.0  AI Hub        AI-01..03, M3-009..014, hub-ai-gateway
  ├─ v3.5  AI scale      OpenRouter, dev API, model picker
  └─ v4.0  Corp app      MOD-INTRA, CAL, CRM-lite, HR-lite
```

---

## 2. LOCKED DECISION (pending user confirm): AUTH-01 Magic Link Only

### 2.1 Запрос PO

> Убрать вход через Telegram. Только Magic Link — как в Cloud Code / Cursor.

### 2.2 Рекомендация: **ПРИНЯТЬ** с разделением auth vs notifications

| Было (TZ) | Станет |
|-----------|--------|
| Magic Link + Telegram **login** | **Magic Link only** login |
| Telegram bot для login code | Bot **только** notifications bind |
| `POST /auth/telegram*` | **Deprecated → remove** v2.5 |
| `user.telegram_id` | Остаётся — заполняется через **NOTIF bind**, не login |

### 2.3 Почему это хорошо

- Один понятный flow (email → code → JWT) — меньше support
- Меньше attack surface (Telegram widget HMAC, bot polling login)
- Согласуется с NOTIF-TG-EMAIL: TG становится **каналом**, не identity provider
- Cloud Code / Linear / Notion pattern

### 2.4 Migration & breaking

1. Existing users с `telegram_id` от login — **сохранить** для notify
2. Login page: убрать `TelegramButton`, widget, `/auth/telegram-code`
3. Bot: убрать `/login` handler; оставить `/start bind_*`, `/link`, `/test`
4. Docs: `60-tz-constraints.mdc`, `vision US-M01`, `system_spec`, onboarding copy
5. Tests: удалить handler_telegram login tests; добавить NOTIF bind tests
6. **ADR-AUTH-01** файл в `.planning/decisions/`

### 2.5 Effort

| Task | Effort |
|------|--------|
| Remove login endpoints + FE | M |
| Bot refactor notification-only | M |
| Migration note + docs | S |
| NOTIF bind (already specced) | M |

**Total AUTH wave:** ~2 weeks parallel with v2.5

---

## 3. Gray areas — полный разбор (2026-06-21)

> **Цель:** снять подводные камни **до** кодирования. Рекомендации — default для planner.

---

### 🔴 GA-CRITICAL: «Magic Link only» ≠ «убрать Telegram»

**Факт из кода:**

| Слой | Реальность |
|------|------------|
| Backend | `/auth/magic-link` + `/verify` ✅ полностью |
| FE API | `sendMagicLink` / `verifyMagicLink` ✅ |
| FE UI | `MagicLinkForm` **существует, но НЕ подключён** к `login.tsx` |
| Login page | **Только** `TelegramButton` + dev panel |
| Docs | `TECHNICAL.md` / `TESTING.md` **устарели** (пишут «только Telegram») |

**Подводный камень:** AUTH-01 без wiring `MagicLinkForm` = **prod lockout** (нет входа без DEV_MODE).

**Решение (LOCK):**
1. PR#1-A: подключить `MagicLinkForm` → primary login
2. PR#1-B: удалить TG login endpoints + `TelegramButton`
3. PR#1-C: sync docs + e2e login.spec.ts
4. Prod **блокер:** `RESEND_API_KEY` обязателен (уже в config validation при `DEV_MODE=false`)

---

### GA-AUTH-01: Legacy users `telegram_*@hub.local`

**Код:** `findOrCreateUserByTelegram` создаёт placeholder email.

| Сценарий | Риск | Mitigation |
|----------|------|------------|
| TG-only user, убираем TG login | Не может войти | **Blocking banner** + admin bulk invite email |
| `telegram_id` unique | Bind конфликт если два аккаунта | Bind проверяет 409; legacy id сохраняем |
| Placeholder email в Resend alerts | Bounce / spam score | NOTIF: block email channel if `@hub.local` or `telegram_*@` |
| Onboarding не требует real email | Остаются phantom users | Onboarding step 0: **real email required** if placeholder |

**Решение (LOCK):** soft migration 90 days — login only via magic link to **verified** email; placeholder users forced to add email on first magic link attempt (merge by admin if duplicate).

---

### GA-AUTH-02: Bot — один token, два lifecycle

**Сейчас:** `TelegramBotAuth` — `/login`, Redis prefix `tglogin:`  
**Цель:** bind `/start bind_*`, `/link`, prefix `tg:bind:` / `tg:linkcode:`

| Вопрос | Рекомендация |
|--------|--------------|
| Один bot или два? | **Один token** — меньше ops; handlers разделить пакетами |
| Удалить `/login` сразу? | **Да** v2.5 — иначе обход AUTH-01 |
| Polling vs webhook | Polling OK dev; prod webhook + secret (SEC) |
| Bot без token | Login OK (Resend); TG notify disabled gracefully |

---

### GA-AUTH-03: invite_only + Magic Link

**Риск:** Magic link на неизвестный email → auto-register в open mode.

**Решение (LOCK):** `CheckRegistrationAllowed` уже есть — magic link **must** respect `invite_only`; unknown email → 403 «нужен invite»; invite flow unchanged.

---

### GA-AUTH-04: Rate limit shared?

Magic link: 3/10min per email. TG bind: 5/10min per user.

**Риск:** attacker enumerates emails via timing.

**Решение (LOCK):** generic error «если email есть — код отправлен»; одинаковый response time; audit failed attempts.

---

### GA-NOTIF-01: Email verify — поля нет в schema

**Факт:** `user` schema **нет** `email_verified_at`.

**Решение (LOCK):** migration add `email_verified_at`; magic link login **sets verified**; profile re-verify on email change; separate micro-flow for alert opt-in optional (can reuse same code path).

---

### GA-NOTIF-02: `ChannelEmail` + defaults

**Факт:** preferences default **enabled=true** for all channels; only in_app/telegram/push exist.

| Риск | Mitigation |
|------|------------|
| Email spam on deploy | **Default OFF** for `channel=email` all types |
| Migration seeds | No auto-create email prefs |
| Resend 3000/mo free | Rate 20/user/day + project cap |

---

### GA-NOTIF-03: `run_failed` recipients — spam bomb

**Код сейчас:** `handleRunCompleted` → **все** project members except initiator, type `run_completed`.

**Подводный камень:** 50 members × every CI run = notification hell.

**Решение (LOCK):**
- New type `run_failed` only if `failed_count > 0`
- Recipients: **QA role** + run creator + explicit watchers (NOT all members)
- `run_completed` pass → in-app only, TG/email **off** by default
- Project setting override (NOTIF spec §5.4) — v2.6 not v2.5

---

### GA-NOTIF-04: Group TG chat (#qa-alerts)

**Риск:** `telegram_chat_id` per project — bot must be admin in group; privacy (run titles leak).

**Решение (LOCK):** **v2.6**, not v2.5; optional per project; RBAC `PermProjectManageSettings`; audit.

---

### GA-NOTIF-05: `RESEND_FROM` vs `RESEND_FROM_ALERTS`

**Риск:** magic link from `noreply@`, alerts from `alerts@` — DMARC/SPF misconfig.

**Решение (LOCK):** two env vars; document DNS checklist in deploy wiki; same domain recommended.

---

### GA-EXEC-01: Parallel tracks — ops trap

**PO:** всё параллельно потихоньку.

| Камень | Что ломается | Guardrail |
|--------|--------------|-----------|
| 5 tracks × 1 dev | Context switch, half-done PRs | **WIP limit 2** active tracks |
| CI 3000 min/mo | Parallel PRs burn minutes | Path filters in CI (already partial) |
| Schema conflicts | ent migrations clash | **Track owner** per schema domain |
| Feature flags absent | Half-shipped UX | Use env flags: `AI_GATEWAY_URL`, `HUB_INTRANET_URL` optional |

**Решение (LOCK):** parallel **design**, serial **merge** to main — max 2 feature PRs/week unless hotfix.

---

### GA-MS-01: Microservices too early

**PO:** v4 modules as separate services.

| Anti-pattern | Symptom | Rule |
|--------------|---------|------|
| Split before metrics | 8 deploys, 1 user | Extract only with **trigger** (MS roadmap §1) |
| Shared DB writes | Cross-service corruption | **No cross-service SQL** — NATS + API |
| JWT in 6 services | Key rotation nightmare | Shared `platform/auth` lib, single issuer **hub-core** |
| Distributed transactions | Partial automation rule | **Outbox** in core; worker idempotent |
| hub-comms split | WS room desync | **Defer** until >2k WS |

**Решение (LOCK):** v4 modules start as `internal/intranet/` **in monolith**; extract to `cmd/intranet` when >3k LOC or independent release cycle.

---

### GA-MS-02: hub-notify extract

**Риск:** duplicate TelegramSender in auth + notification; circular deps.

**Решение (LOCK):** before extract — move bot to `internal/platform/telegram/bot.go`; auth package **drops** bot entirely; notification owns bind + dispatch.

---

### GA-AI-01: Prompt/content leakage

**Риск:** @hub-ai in project channel → secrets in prompts; audit «metadata only» but messenger **stores** messages in Mongo.

**Решение (LOCK):**
- AI gateway **never** reads Mongo directly; receives prompt via API
- Admin toggle: «allow issue context» strips attachments/secrets regex
- `AI_OFFLINE_ONLY` for air-gap
- Usage quota per user/day **before** v3.5

---

### GA-AI-02: Provider keys in DB

**Риск:** AES key in env single point; backup exposes keys.

**Решение (LOCK):** `platform/crypto` master key from K8s secret; keys rotatable; backup excludes `ai_provider.api_key_enc` or encrypted at rest backup ADR.

---

### GA-AI-03: OpenRouter / VPN / 152-FZ

**Риск:** RU corp compliance — data leaves jurisdiction.

**Решение (LOCK):** admin labels providers `ru_resident: true/false`; UI warning on non-RU model; default preset DeepSeek+GigaChat+Ollama for RU.

---

### GA-PROD-01: Automation loops

**Риск:** status→Done→post channel→bot creates issue→status change→∞

**Решение (LOCK):**
- Max **1 action per event** per rule (spec)
- Idempotency key: `rule_id+issue_id+transition_id`
- **No** automation triggers on system messages (actor=system)
- Async via NATS consumer, not sync in MoveIssue (avoid txn timeout)

---

### GA-PROD-02: Sync vs async automation

**Spike says:** MoveIssue hook OR NATS consumer.

**Решение (LOCK):** **NATS consumer** on `issue.moved` — same as notifications; 5s SLA enough; rollback safe.

---

### GA-V4-01: CRM/HR vs dev suite identity

**Риск:** duplicate `contact` vs `user`; CRM deals linked to wrong person.

**Решение (LOCK):** CRM **contacts** separate entity; optional `linked_user_id`; HR leave uses existing `user` + PROD-01 workflow.

---

### GA-V4-02: Calendar iCal + calls

**Риск:** double bookings; timezone bugs (UTC mandate).

**Решение (LOCK):** store UTC; iCal export read-only v4.0; two-way sync **v4.5** only.

---

### GA-V4-03: Intranet + messenger overlap

**Риск:** news feed vs channel announcements — duplicate features.

**Решение (LOCK):** news = company-wide moderated; channels = team; cross-post as optional action PROD-01 v2.6.1.

---

### GA-DOC-01: TZ constraints file stale

**Still says:** «Magic Link + Telegram login» + «no email notif».

**Решение (LOCK):** update `60-tz-constraints.mdc` in same PR as AUTH-01:
- Login: Magic Link only
- Telegram: notification channel
- Email: magic link + transactional alerts (ADR-NOTIF-01)

---

### GA-SEC-01: E2E prod login

**Сейчас:** e2e uses dev login panel only.

**Решение (LOCK):** CI keeps dev login; add optional `@resend` e2e with `MAGIC_LINK_TEST_CODE` env in smoke job; prod smoke manual checklist.

---

### GA-SEC-02: `telegram_id` hijack via bind

**Риск:** stolen bind token links victim TG to attacker account.

**Решение (LOCK):** token one-time JWT-bound; 10m TTL; HTTPS deep link only; confirm in UI with «pending bind» state; rate limit.

---

## 3.1 Decision matrix — рекомендуем lock (defaults)

| ID | Topic | Recommended default |
|----|-------|---------------------|
| D-01 | AUTH scope | Wire MagicLinkForm + remove TG login (not remove-only) |
| D-02 | Legacy TG users | Forced email migration + admin merge tool |
| D-03 | Bot | Single bot, notification handlers only |
| D-04 | Email verify | New `email_verified_at` column |
| D-05 | Email defaults | Opt-in OFF all types |
| D-06 | run_failed scope | QA roles only, not all members |
| D-07 | Group TG | v2.6 defer |
| D-08 | Parallel WIP | Max 2 active tracks |
| D-09 | μsvc v4 | Monolith module first, extract on trigger |
| D-10 | PROD-01 | NATS async + idempotency |
| D-11 | AI compliance | RU provider labels + offline mode |
| D-12 | TZ file | Update in AUTH PR |

---

## 3.2 Open — нужен PO только если disagree

| # | Question | Default if silence |
|---|----------|-------------------|
| 1 | Legacy TG user without email — block login or admin-only recovery? | **Grace 30d** then block |
| 2 | Magic link code in API response dev-only — keep for local? | **Yes** (existing `IsDevMode`) |
| 3 | First provider AI: DeepSeek or Ollama? | **DeepSeek** prod, Ollama compose profile |
| 4 | Intranet MVP: news only or news+directory? | **News + directory read** |

---

## 4. Proposed milestone re-pack (для обсуждения)

### Milestone v2.5 «Trust» (6–8 weeks)

- AUTH-01 Magic Link only
- NOTIF-02..04 Telegram bind
- FB-07..13 E2E + WCAG + RHF
- NOTIF-08 run_failed events
- PG-03 secrets rotation

### Milestone v2.6 «Automate» (8 weeks)

- PROD-01 workflow automation
- PROD-02 flaky
- PROD-03 issue↔message
- PROD-04 webhooks
- NOTIF-05..07 email alerts (if ADR yes)
- Optional: `hub-automation-worker` extract

### Milestone v3.0 «AI Native» (12 weeks)

- hub-ai-gateway service
- AI-01..03, M3 core
- DeepSeek + GigaChat + Ollama

### Milestone v4.0 «Hub One» (6+ months)

- MOD-INTRA, CAL, CRM, HR
- hub-intranet optional split

---

## 5. Risks if we change auth now

| Risk | Mitigation |
|------|------------|
| RU users без email | Corporate email required; admin invite |
| Telegram-only legacy users | One-time «add email» banner |
| TZ rules conflict | ADR-AUTH-01 + update `60-tz-constraints.mdc` |
| Bot token still needed | Yes — for notifications only |

---

## 6. Decisions (PO confirmed 2026-06-21)

| ID | Decision | Status |
|----|----------|--------|
| **AUTH-01** | Magic Link **only** login; Telegram **notifications bind only** | ✅ LOCKED |
| **NOTIF-01** | Transactional email via Resend (opt-in, run_failed etc.) | ✅ LOCKED |
| **EXEC-01** | **Parallel tracks** — всё потихоньку, не strict waterfall | ✅ LOCKED |
| **V4-01** | v4 modules **parallel** as **separate microservices** when ready | ✅ LOCKED |

### Parallel workstreams (EXEC-01)

```
Track A  Trust      AUTH-01, NOTIF bind, E2E, WCAG          (ongoing)
Track B  Automate   PROD-01..04, automation-worker         (ongoing)
Track C  Notify     NOTIF email, run_failed                 (ongoing)
Track D  AI         hub-ai-gateway POC → v3                 (ongoing)
Track E  Corp μsvc  hub-intranet, hub-calendar, hub-crm…    (ongoing, separate deploys)
```

**Rule:** каждый track = отдельный K3s Deployment / compose profile; monolith остаётся default до cutover.

---

## 7. Next GSD steps

1. ~~User confirms gray areas~~ ✅
2. Create `v2.5-PLAN.md` per track (or update ROADMAP)
3. **ADR-AUTH-01** + **ADR-NOTIF-01** files
4. Update `60-tz-constraints.mdc`: login = Magic Link only; TG = notify channel
5. First slice: AUTH-01 remove TG login (1 PR) + NOTIF-02 bind (1 PR)

---

*Status: **LOCKED** 2026-06-21 · See `DISCUSS-ROADMAP-LOCKED.md` · Next: gsd-plan-phase Phase 15*
