# DISCUSS LOCKED — Hub Roadmap PO Decisions

> **Status:** ✅ LOCKED · **Date:** 2026-06-21  
> **Next:** `gsd-plan-phase` → Phase 15 PLAN.md · Branch: `feat/v2.5-trust`  
> **Source:** discuss-phase forms + PO confirmations

---

## Process

| Decision | Value |
|----------|-------|
| Discuss | **CLOSED** |
| UAT | Full automated regression + qa-ver (not manual PO sign-off) |
| Verifiers | Quality profile; inline qa-gate fallback on API limit |
| Git push | Only when PO asks |
| CI minutes | Economize (path filters) |
| Swagger | Wave at end of each phase |
| Security audit | `gsd-secure-phase` after **each** phase |
| Solo dev | Planner sets milestone order |
| Coverage exit Phase 15 | **50% total backend** (blocking) |
| Dev until end of summer | No staging; Prometheus when needed |

---

## Phase map (ROADMAP)

| Phase | Theme | Exit |
|-------|-------|------|
| **15 Trust** | AUTH magic-only, org, NOTIF bind, E2E, WCAG, run_failed, TZ docs, Biome fix, LiveKit CI | Full row below |
| **15.1** | Auth + org + NOTIF + E2E + TZ rewrite | First code slice |
| **15.2** | `hub-mail` μsvc | After 15.1 |
| **16 Automate** | PROD max, flaky, PROD-03/04, Gantt, webcam, marketing email, triage inbox | Gantt→v4 if overload |
| **17 AI Native** | AI gateway max, full JQL automation | BYOK + Ollama |
| **v4 Super-app** | Intranet max, Calendar, CRM, HR, full accounting, Lark-style no-code DB | microservices when ready |
| **v5** | Multi-company / multi-tenant, native apps optional | — |

**Phase 15 exit (full):** AUTH + organization + NOTIF TG bind + E2E FB-07..09 + WCAG + RHF patterns + run_failed + Biome wave + LiveKit in CI + quick wins FB-01..06.

---

## AUTH & Organization

| ID | Decision |
|----|----------|
| AUTH-01 | **Magic Link only** login; wire `MagicLinkForm` before removing TG login |
| AUTH-02 | Telegram **notify bind only**; TG optional (`TELEGRAM_TOKEN` empty OK) |
| AUTH-03 | Dev magic link code in API response: **DEV_MODE only** |
| AUTH-04 | Legacy TG users: **N/A** (not in prod) |
| ORG-01 | Table **`organization`** in v2.5 (1 org per instance) |
| ORG-02 | Model like **Jira Cloud sites**: User → Organization → Projects |
| ORG-03 | URL: **`{slug}.hub.local`** subdomain (prod: planner picks best) |
| ORG-04 | **`system_role founder`**: one founder, then admin only |
| ORG-05 | Registration: **invite link only**; any email domain |
| ORG-06 | **Admin queue** Pending users; **7 days** → auto-delete |
| ORG-07 | Admin email verify: **soft UX** (work with limits until approved) |
| AUTH-05 | **No SSO** (no Google/Microsoft login) |

---

## NOTIF & Mail

| ID | Decision |
|----|----------|
| NOTIF-01 | Transactional email via Resend → **hub-mail** μsvc |
| NOTIF-02 | Email alerts **v2.5**; hub-mail **Phase 15.2** |
| NOTIF-03 | `new_login` email: **opt-in** |
| NOTIF-04 | `run_failed` → **QA/tester + run author**; all TMS alert types |
| NOTIF-05 | Group TG chat: **v2.6** optional |
| NOTIF-06 | SMTP self-host: **v2.6** optional |
| NOTIF-07 | TG bots: **1 per project + system bot**; Hub **auto-creates** via BotFather flow |
| NOTIF-08 | TG bind UI shown even when token optional |
| NOTIF-09 | Email **marketing** v2.6: **Mailchimp-lite**; **admin-only** send |
| MAIL-01 | `cmd/hub-mail` monorepo; admin panel for mail service |

---

## Automation & Tracker (Phase 16–17)

| ID | Decision |
|----|----------|
| PROD-01 | Automation **maximum** (Phase 16) |
| PROD-02 | Flaky detection **maximum** |
| PROD-03/04 | Issue↔message + webhooks dashboard **maximum** |
| JQL-01 | **Full JQL** automation — **Phase 17** |
| GANTT-01 | **Gantt Phase 16**; defer to v4 if overload |
| TRIAGE-01 | Linear-style triage inbox **Phase 16** |
| WEBHOOK-01 | Delivery log if **free/OSS** — Phase 16 |

---

## AI (Phase 17)

| ID | Decision |
|----|----------|
| AI-01 | **Ollama local** + **any API** via user/admin keys |
| AI-02 | Model list from provider subscription/capabilities check |
| AI-03 | **@hub-ai** weak Ollama bot + **dialog** (model picker, **photo+text**, no video) |
| AI-04 | Keys: **admin + user BYOK** |
| AI-05 | Storage: **maximum** (separate `ai_conversations` + audit metadata) |
| AI-06 | Non-RU models: **UI warning** + admin `ru_resident` label |

---

## Calls & v4

| ID | Decision |
|----|----------|
| CALL-01 | **Webcam Phase 16**; recording stays **audio+screen** |
| LIVEKIT-01 | LiveKit **in CI** compose |
| V4-01 | Module order: **Intranet → Calendar → CRM → HR** |
| V4-02 | Intranet **maximum** (news + kudos + directory + …) |
| V4-03 | **Full accounting** v4; **Hub only, no 1C** |
| V4-04 | **Lark Base clone** (no-code DB) v4 |
| V4-05 | Random Coffee / kudos matching **v4 intranet** |
| V4-06 | Public **status page** v4 |
| SEARCH-01 | Meilisearch indexer **v2.6+** |

---

## TZ scope rewrite (PO 2026-06-21)

| Was FORBIDDEN | Now |
|---------------|-----|
| SSO/LDAP | ❌ Still **out** |
| Gantt | ✅ **In** Phase 16 |
| Billing/SaaS | ⏳ **Later** |
| Multi-tenant | ⏳ **v5** (multi-company) |
| Native iOS/Android | **PWA enough** |
| Webcam calls | ✅ **In** Phase 16 |
| Email marketing | ✅ **In** v2.6 (Mailchimp-lite) |
| Email transactional | ✅ Magic link + alerts + marketing (hub-mail) |
| No-code DB | ✅ **In** v4 (Lark clone) |
| Built-in accounting | ✅ **In** v4 (full; no 1C) |
| Magic Link + TG login | ❌ **Magic Link only** |

See: `.planning/decisions/ADR-TZ-REWRITE-2026.md`

---

## ADRs

| File | Topic |
|------|-------|
| `decisions/ADR-AUTH-01-magic-link-only.md` | Login |
| `decisions/ADR-NOTIF-01-transactional-email.md` | Email alerts |
| `decisions/ADR-TZ-REWRITE-2026.md` | Scope expansion |

---

*Parent: DISCUSS-ROADMAP-2026-CONTEXT.md · REGULATION §9 gsd-plan-phase next*
