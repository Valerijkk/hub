# Edge Expansion Wave3 E01 — Auth / RBAC / Platform

> **Статус:** foresight backlog · 2026-07-10 · владелец: `qa-manual`  
> **Тип:** **только Edge** (границы: refresh-family/reuse, session UA/label, magic resend cooldown, profile TZ/avatar, GDPR session wipe, onboarding fields, owner transfer, cross-project roles, imex pack, pending filters, audit/outbox/presigned/webhook)  
> **ID ranges:** `ECC-AUTH-300+`, `ECC-RBAC-300+`, `ECC-PLT-300+` — **не пересекают** wave1 `*100+`, wave2 `*200+`, wave4 `*400+`, wave5 `*500+`, канон `001–040`, corner `*C*`  
> **Не заменяет** каталог и `docs/TESTING.md` — углубление edge для July Rebuild / QA-HARDENING  
> **Wave1 (не дублировать 1:1):** `.planning/reference/edge-expansions/01-AUTH-RBAC-PLATFORM.md` (`ECC-*-100…`)  
> **Wave2 (не дублировать 1:1):** `.planning/reference/edge-expansions/wave2/edge/E01-AUTH-RBAC-PLATFORM.md` (`ECC-*-200…`)  
> **Фокус wave3:** refresh rotation family · session UA/label · magic resend cooldown · profile TZ/avatar · GDPR erase wipe · onboarding fields · owner transfer · cross-project role matrix · imex role pack · pending queue filters · audit retention · outbox batch · presigned TTL · webhook HMAC window  
> **Связи:** US-M01/M08/M09/M10, US-S03 · NFR-05/06/07/08/09/10/11/12/13/24/25 · ADR-AUTH-01 · TC-AUTH-*, TC-RBAC-*, TC-ADMIN-*, TC-PROJ-*  
> **Канон каталога:** `.planning/archive/july-2026/EDGE-CORNER-CATALOG.md` — **не изменять** этим файлом  
> **Gap:** только `TODO`

### Gap-легенда (для этого файла)

| Gap | Значение |
|-----|----------|
| **TODO** | Нет dedicated edge-теста / TC |

### Отличие от wave1 / wave2

| Тема | Wave1 (`*100+`) | Wave2 (`*200+`) | Wave3 (`*300+`) — новое углубление |
|------|-----------------|-----------------|-------------------------------------|
| Auth | Magic TTL/rate, JWT 15m, refresh 30d, sessions list, email len, pending 7d | Unicode/IDN email, device MaxSessions | **refresh family/reuse**, **UA/label bounds**, **resend cooldown**, **TZ/avatar**, **GDPR wipe**, **onboarding fields** |
| RBAC | cache TTL hit/miss, founder=1, invite max-uses, bitmask bits | invite batch, cache TTL deepen | **owner transfer**, **cross-project roles**, **imex pack size**, **pending filters**, **org settings strings** |
| Platform | pagination min/max, soft-delete day, BodyLimit smoke, rate 100/s | pagination OBO, soft-delete 29/30/31 | **audit retention/filter**, **outbox batch**, **presigned TTL**, **webhook HMAC window**, **backup ZIP bounds**, **metrics label cardinality** |

---

## 1. Auth — `ECC-AUTH-300+`

> Фокус: **refresh rotation family / reuse**, **session UA/label**, **magic resend cooldown**, **profile TZ/avatar**, **GDPR erase wipe**, **onboarding fields**

### 1.1 Refresh rotation family / reuse window

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-AUTH-300 | Edge | Refresh success → old refresh immediately reused | 401; family revoked **или** single-use documented | API | P0 | NFR-10, ECC-AUTH-022 sibling | TODO |
| ECC-AUTH-301 | Edge | Refresh reuse within grace window `0` / `ε` / `grace+1s` (if grace exists) | 0/ε: policy documented; grace+1s → full family kill | API | P0 | refresh reuse | TODO |
| ECC-AUTH-302 | Edge | Rotation chain length = MaxChain−1 → one more refresh | OK; chain grows | API | P1 | rotation chain | TODO |
| ECC-AUTH-303 | Edge | Rotation chain length = MaxChain → next refresh | 401 force re-login **или** trim oldest; documented | API | P1 | chain cap | TODO |
| ECC-AUTH-304 | Edge | Refresh after logout of **same** session only | that family dead; sibling sessions alive | API | P0 | TC-AUTH-12 deepen | TODO |
| ECC-AUTH-305 | Edge | Refresh token byte length = min opaque / max Redis / max+1 | min/max OK; +1 reject; no truncate-match | API | P0 | refresh format | TODO |
| ECC-AUTH-306 | Edge | Refresh with valid access still unexpired (dual credential) | refresh OK; new access; old access may stay until exp | API | P1 | dual token | TODO |

### 1.2 Session UA / device label / fingerprint bounds

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-AUTH-307 | Edge | User-Agent length = 0 / 1 / max / max+1 at login | empty→default; max OK; +1 truncate **или** 400 | API | P1 | session meta | TODO |
| ECC-AUTH-308 | Edge | Device label PATCH: empty / max grapheme / max+1 / emoji+ZWJ | empty clear or 400; max OK; +1 400; emoji intact | API | P1 | device label | TODO |
| ECC-AUTH-309 | Edge | Device fingerprint string: null / 32 / 128 / 129 chars | policy bounds; no crash; stable list key | API | P2 | fingerprint | TODO |
| ECC-AUTH-310 | Edge | ListSessions: UA with control chars / CRLF / null byte | sanitized in response; no header inject in logs | API | P0 | log/header inject | TODO |
| ECC-AUTH-311 | Edge | Session created_at vs last_active: refresh updates last_active only | created_at immutable; last_active ≥ created_at | API | P1 | ListSessions | TODO |

### 1.3 Magic-link resend cooldown (≠ rate 3/10m)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-AUTH-312 | Edge | Resend at cooldown−1s after successful send | 429/400 cooldown; old code still valid until TTL | API | P0 | ECC-AUTH-032 deepen | TODO |
| ECC-AUTH-313 | Edge | Resend at cooldown exact / cooldown+1s | exact policy; +1s → new code; old invalidated **или** both TTL | API | P0 | resend policy | TODO |
| ECC-AUTH-314 | Edge | Resend when no prior code (cold email) | treat as first send 200; no phantom cooldown | API | P1 | cold resend | TODO |
| ECC-AUTH-315 | Edge | Resend while verify lockout active | still cooldown/rate rules; no unlock via resend | API | P0 | lockout∩resend | TODO |
| ECC-AUTH-316 | Edge | Resend count in window = MaxResend / Max+1 | Max OK; +1 → 429 until window end | API | P1 | resend cap | TODO |

### 1.4 Profile timezone / avatar / onboarding fields

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-AUTH-317 | Edge | Timezone: obscure IANA `Pacific/Kiritimati` / `Etc/GMT+12` / `UTC` | accept valid; API timestamps remain UTC Z | API | P1 | NFR-12, ECC-AUTH-117 deepen | TODO |
| ECC-AUTH-318 | Edge | Timezone: `Europe/Kyiv` vs legacy `Europe/Kiev` alias | one canonical store **или** both map; FE consistent | API | P1 | IANA alias | TODO |
| ECC-AUTH-319 | Edge | Timezone: `GMT+3` / `UTC+03:00` / numeric offset only | 400 prefer IANA; no silent accept ambiguous | API | P1 | TZ format | TODO |
| ECC-AUTH-320 | Edge | Avatar URL length = max−1 / max / max+1 | max OK; +1 → 400 | API | P1 | profile | TODO |
| ECC-AUTH-321 | Edge | Avatar URL scheme `https` OK; `http` / `data:` / `javascript:` / `file:` | only allowlist; SSRF schemes 400 | API | P0 | NF-07 sibling | TODO |
| ECC-AUTH-322 | Edge | Onboarding: required fields all empty vs all min-valid | empty→400; min-valid → onboarding=true | API | P1 | TC-AUTH-11 deepen | TODO |
| ECC-AUTH-323 | Edge | Onboarding: display_name at max while timezone invalid | 400; onboarding stays false; no partial commit | API | P0 | atomic onboarding | TODO |
| ECC-AUTH-324 | Edge | Onboarding complete when already true (idempotent boundary) | 200; no second audit storm | API | P2 | idempotent | TODO |

### 1.5 GDPR / deactivate session wipe

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-AUTH-325 | Edge | Admin deactivate: refresh count before = N → after = 0 | all Redis refresh gone; access expire naturally or deny | API | P0 | US-S03, ECC-AUTH-033 sibling | TODO |
| ECC-AUTH-326 | Edge | GDPR hard-erase: sessions list for erased user | 404/401; no PII in error; audit redacted | API | P0 | TC-ADMIN-03 | TODO |
| ECC-AUTH-327 | Edge | Soft-delete self (if allowed) then Magic Link send | reject; no new code for trashed user | API | P0 | auth×trash | TODO |
| ECC-AUTH-328 | Edge | Deactivate at access TTL−1s: next API call | 401/403 immediately (not wait access exp) | API | P0 | deactivate gate | TODO |
| ECC-AUTH-329 | Edge | Invite accept handoff: join → first Magic Link within 1s | pending/active per mode; code keyed to canonical email | API | P1 | invite→magic | TODO |

**Auth Edge в этом файле: 30**

---

## 2. RBAC / Org — `ECC-RBAC-300+`

> Фокус: **owner transfer**, **cross-project role matrix**, **imex role pack**, **pending queue filters**, **org settings string bounds**

### 2.1 Owner transfer / last-owner bounds

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-RBAC-300 | Edge | Transfer Owner A→B when B is Viewer | B→Owner; A→Developer/Viewer per policy; Owners≥1 | API | P0 | US-M10 | TODO |
| ECC-RBAC-301 | Edge | Transfer Owner A→B when B not a member | 400/404; A remains sole Owner | API | P0 | transfer | TODO |
| ECC-RBAC-302 | Edge | Transfer when project has exactly 1 Owner (A) and B exists | A demoted; B Owner; never 0 Owners | API | P0 | TC-PROJ-10 deepen | TODO |
| ECC-RBAC-303 | Edge | Transfer to self (A→A) | 400 no-op **или** idempotent 200; no audit spam | API | P2 | transfer self | TODO |
| ECC-RBAC-304 | Edge | Transfer with stale `version` on membership | 409; roles unchanged | API | P0 | NFR-24 | TODO |
| ECC-RBAC-305 | Edge | Org founder transfer attempt (system_role) | reject; founder immutable except documented break-glass | API | P0 | TZ founder | TODO |

### 2.2 Cross-project role matrix

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-RBAC-306 | Edge | Same user: Owner in P1, Viewer in P2 | Has(P1)=full; Has(P2)=read-only; no bleed | API | P0 | project scope | TODO |
| ECC-RBAC-307 | Edge | Same user: custom role in P1, default Dev in P2 | masks independent; cache keys per project | API | P0 | cache key | TODO |
| ECC-RBAC-308 | Edge | Remove membership from P1 only; P2 intact | P1 403; P2 200 | API | P0 | remove scope | TODO |
| ECC-RBAC-309 | Edge | User in MaxProjects−1 / Max / Max+1 projects | Max OK; Max+1 invite/join reject | API | P1 | membership cap | TODO |
| ECC-RBAC-310 | Edge | List my projects at exactly MaxProjects | all returned; stable order; no truncate silent | API | P1 | list bound | TODO |

### 2.3 Imex role pack / custom-role export size

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-RBAC-311 | Edge | Export custom roles: count = 0 / 1 / Max | empty file OK; Max OK; schema version present | API | P1 | imex | TODO |
| ECC-RBAC-312 | Edge | Import role pack size = max bytes / max+1 | max OK; +1 → 413/400; no partial apply | API | P0 | BodyLimit∩imex | TODO |
| ECC-RBAC-313 | Edge | Import pack with role count = MaxCustomRoles / Max+1 | Max OK; +1 reject all-or-nothing | API | P0 | custom role cap | TODO |
| ECC-RBAC-314 | Edge | Import pack: bitmask unknown high bit at edge | reject or strip documented; no open-perm | Unit | P0 | ECC-RBAC-017 deepen | TODO |
| ECC-RBAC-315 | Edge | Export while org has only default 4 roles | defaults exported or omitted per schema; import idempotent | API | P2 | defaults | TODO |

### 2.4 Pending queue filters / org settings strings

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-RBAC-316 | Edge | Pending list: filter age = 0d / 6d / 7d−ε / ≥7d | ≥7d empty (purged); others visible | API | P0 | ORG-06 | TODO |
| ECC-RBAC-317 | Edge | Pending list: offset/limit at exact total pending | last page partial; has_more false | API | P1 | NFR-14 | TODO |
| ECC-RBAC-318 | Edge | Pending search email: prefix 1 char / max / empty | empty=all or 400; max OK; no ReDoS | API | P1 | admin search | TODO |
| ECC-RBAC-319 | Edge | Org display_name: empty / 1 / max / max+1 | empty 400; max OK; +1 400 | API | P1 | org settings | TODO |
| ECC-RBAC-320 | Edge | Org slug: reserved / max / unicode | reserved 400; unicode reject or slugify | API | P1 | org slug | TODO |
| ECC-RBAC-321 | Edge | registration_mode set to same value twice | idempotent 200; single audit or none | API | P2 | mode idem | TODO |
| ECC-RBAC-322 | Edge | Invite list filter: status=open count = 0 / MaxOpen | empty []; Max listed; create Max+1 → 400 | API | P1 | invite list | TODO |
| ECC-RBAC-323 | Edge | Member role PATCH: unknown role slug at boundary | 400; cache untouched | API | P0 | role enum | TODO |

**RBAC Edge в этом файле: 24**

---

## 3. Platform cross-cutting — `ECC-PLT-300+`

> Фокус: **audit retention/filter**, **outbox batch**, **presigned TTL**, **webhook HMAC window**, **backup ZIP bounds**, **metrics cardinality**

### 3.1 Audit retention / filter bounds

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-PLT-300 | Edge | Audit list: `from`/`to` window = 0s / 1s / max range / max+1 | max OK; +1 → 400; empty window `[]` | API | P1 | TC-ADMIN-01 | TODO |
| ECC-PLT-301 | Edge | Audit retention day = retention−1 / exact / +1 | −1 queryable; exact policy; +1 purged or 404 | API | P0 | audit retention | TODO |
| ECC-PLT-302 | Edge | Audit filter actor_id = nil UUID / valid / unknown | nil 400; unknown `[]` 200 | API | P1 | NFR-25 | TODO |
| ECC-PLT-303 | Edge | Audit page: limit=max at last event offset | partial page; no dup events | API | P1 | NFR-14 | TODO |
| ECC-PLT-304 | Edge | Audit action string length = max / max+1 in filter | max OK; +1 400 | API | P2 | filter len | TODO |

### 3.2 Outbox / NATS batch bounds

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-PLT-305 | Edge | Outbox flush batch size = 1 / MaxBatch / Max+1 queued | flush ≤MaxBatch per tick; remainder next tick | Unit | P1 | outbox | TODO |
| ECC-PLT-306 | Edge | Outbox payload size = max / max+1 | max publish OK; +1 reject at enqueue; no poison infinite | Unit | P0 | outbox size | TODO |
| ECC-PLT-307 | Edge | NATS subject length = max / max+1 | max OK; +1 fail-closed at publish | Unit | P2 | TC-NATS-01 deepen | TODO |
| ECC-PLT-308 | Edge | Outbox empty queue tick | no-op; no error; metrics zero | Unit | P2 | empty flush | TODO |

### 3.3 Presigned URL TTL / storage key bounds

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-PLT-309 | Edge | Presigned PUT TTL = min / default / max / max+1 | min–max OK; max+1 clamp or 400 | API | P0 | SeaweedFS S3 | TODO |
| ECC-PLT-310 | Edge | Presigned GET at TTL−1s / TTL+1s | −1s 200; +1s 403/expired | API | P0 | presign expiry | TODO |
| ECC-PLT-311 | Edge | Object key length = max / max+1 / empty | empty 400; max OK; +1 400 | API | P1 | S3 key | TODO |
| ECC-PLT-312 | Edge | Presigned content-length hint = 25MB−1 / 25MB / 25MB+1 (Fiber path) | −1/exact per BodyLimit; +1 413 on non-presigned | API | P0 | NFR-04 / NF-09 | TODO |

### 3.4 Webhook HMAC window / backup ZIP / metrics

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-PLT-313 | Edge | Webhook timestamp skew = −window / 0 / +window / window+1s | within OK; outside reject; **never skip** sig check | API | P0 | SEC-02 | TODO |
| ECC-PLT-314 | Edge | Webhook body size = max / max+1 | max verify OK; +1 413 before verify **or** verify then reject — no skip | API | P0 | SEC-02 | TODO |
| ECC-PLT-315 | Edge | Backup ZIP: 0 entities / 1 / MaxEntities export | empty ZIP OK; Max OK; schema present | API | P1 | TC-ADMIN-04 | TODO |
| ECC-PLT-316 | Edge | Backup ZIP uncompressed size = limit / limit+1 | limit OK; +1 reject; no OOM | API | P0 | backup DoS | TODO |
| ECC-PLT-317 | Edge | Metrics label cardinality: same label = MaxSeries / Max+1 | Max OK; +1 drop/aggregate; process alive | Manual | P1 | OBS | TODO |
| ECC-PLT-318 | Edge | Health `/ready` when Postgres OK / Redis down | 503 ready; `/health` may still 200 — documented | API | P1 | ops | TODO |
| ECC-PLT-319 | Edge | Error `request_id`: two parallel requests | two distinct UUID v4; both in body+header | API | P0 | NFR-13 | TODO |
| ECC-PLT-320 | Edge | Optimistic `version` after imex import of entity | imported version = 1 or preserved; next update uses returned | API | P1 | NFR-24∩imex | TODO |

**Platform Edge в этом файле: 21**

---

## Счётчик

| Домен | ID range | Кол-во Edge | P0 | P1 | P2 |
|-------|----------|-------------|----|----|-----|
| Auth | ECC-AUTH-300…329 | **30** | 14 | 14 | 2 |
| RBAC/Org | ECC-RBAC-300…323 | **24** | 13 | 8 | 3 |
| Platform | ECC-PLT-300…320 | **21** | 9 | 9 | 3 |
| **Итого** | | **75** | **36** | **31** | **8** |

### Фокус-покрытие (wave3)

| Фокус | ID (примеры) | Count |
|-------|--------------|------:|
| Refresh family / reuse / chain | ECC-AUTH-300…306 | 7 |
| Session UA / label / fingerprint | ECC-AUTH-307…311 | 5 |
| Magic resend cooldown | ECC-AUTH-312…316 | 5 |
| Profile TZ / avatar / onboarding | ECC-AUTH-317…324 | 8 |
| GDPR / deactivate wipe / invite handoff | ECC-AUTH-325…329 | 5 |
| Owner transfer | ECC-RBAC-300…305 | 6 |
| Cross-project role matrix | ECC-RBAC-306…310 | 5 |
| Imex role pack | ECC-RBAC-311…315 | 5 |
| Pending filters / org settings | ECC-RBAC-316…323 | 8 |
| Audit retention / filter | ECC-PLT-300…304 | 5 |
| Outbox / NATS batch | ECC-PLT-305…308 | 4 |
| Presigned TTL / key / 25MB | ECC-PLT-309…312 | 4 |
| Webhook HMAC / backup / metrics / ready | ECC-PLT-313…320 | 8 |

### Метрики задачи

| Метрика | Значение |
|---------|----------|
| Тип сценариев | **100% Edge** (Happy/Neg/Corner = 0) |
| Минимум задачи | ≥50 |
| Факт | **75** |
| Gap | **все TODO** |
| Пересечение с wave1 `*100+` | **нет** (новые темы, не 1:1) |
| Пересечение с wave2 `*200+` | **нет** (новые темы, не 1:1) |
| Пересечение с каталогом 001–040 | **нет** |
| Изменения `EDGE-CORNER-CATALOG.md` | **нет** |
| Коммит | **не создавался** |

### Рекомендуемый next step

1. `qa-manual-ver` — верификация wave3 E01 (границы, нет дублей wave1/wave2, трассируемость).  
2. P0: AUTH refresh-reuse/GDPR wipe, RBAC owner transfer/cross-project, PLT webhook HMAC/presigned TTL → полные `TC-*` в `docs/TESTING.md`.  
3. После verified → `qa-auto1` (API refresh family, transfer, webhook skew) + `qa-auto2` (onboarding/TZ E2E).  
4. Каталог не трогать; импорт в канон — отдельным PR позже.

---

*Создано: 2026-07-10 · qa-manual · wave3 · только Edge · domains Auth + RBAC/Org + Platform*
