# Hub — Brief перестройки (возврат ~конец июля 2026)

> **Статус:** зафиксировано 2026-07-10 · обновлено 2026-07-10 (SFU) · код freeze до возврата конца июля  
> **SFU-решение (пользователь 2026-07-10):** **остаёмся на LiveKit** — миграция на mediasoup **отклонена** (нет драйвера; дорогая перепись calls без user-value). SEED-001 = rejected.  
> **Триггер старта:** конец июля 2026 — пересмотр планов, edge/corner cases, CAL-SYNC (RSVP), редизайн Cloud Code Designer  
> **Связанные seeds:** `../seeds/SEED-001-mediasoup-migration.md` (**REJECTED**), `../seeds/SEED-002-calendar-calls-email-sync.md` (active)  
> **Базовый roadmap:** `ROADMAP-FINAL-2026-07-03.md` (Волна 1 исполнена в ветках; мерж master — за пользователем)

---

## 1. Что поняла AI-команда (снимок на 2026-07-10)

### 1.1 Продукт и позиция

Hub — self-hosted corporate super-app (10–50 чел.): Tracker + Messenger + TMS + Calls + Dashboards + v4 (intranet/calendar/CRM/HR/accounting/bases/AI). Один инстанс = одна org. Magic Link only (SSO — после ADR-AUTH-SSO). Стек: Go 1.24/Fiber/Ent, React 19/TanStack, PG/Mongo/Redis/NATS/LiveKit/SeaweedFS, Compose→K3s/Caddy.

Метрики baseline: COV ~83%, product ~89/100, 286+ REST, 82 Ent schemas, TF в `backend/testframework/` + Playwright. Единственный prod-блокер ops: FA-06 (ручная ротация секретов).

### 1.2 Как устроена AI-команда (30 агентов)

| Слой | Кто | Модель | Суть |
|------|-----|--------|------|
| Оркестрация | `team-lead` / `team-lead-opus` | Fable 5 / Opus 4.8@1M | Волны, приёмка; Opus только эскалация |
| Декомпозиция / продукт | `tech-lead`, `product`+`product-ver` | Sonnet 5 | Планы ≤200–300 LOC, MoSCoW/KPI |
| Аналитика | `ba1/ba2`+`ba-ver`, `sa1/sa2`+`sa-ver` | Sonnet 5 | Потоки A/B доменов; US/AC → FR/ERD/API |
| Архитектура / дизайн | `arch`+`arch-ver`, `designer`+`designer-ver` | Sonnet 5 | ADR/C4; DESIGN-SPEC + `--hub-*` |
| Реализация | `be1/be2`+`be-ver`, `fe1/fe2`+`fe-ver` | Sonnet 5 | Параллель без пересечения файлов; кросс-ревью |
| QA | `qa-manual`+ver, `qa-auto1/2`+ver | Sonnet 5 | TC→verify→автотесты; юниты пишут be/fe |
| Cross-cut | devops, security-auditor, debugger, code-reviewer, docs-writer | Sonnet 5 | Infra/OWASP/RCA/PR/docs |

**Инварианты процесса (уже в правилах):**

- Агенты коммитят/пушат **фича-ветки**; **master — только пользователь**; rebase-merge.
- Регрессия домена у агента; полный `/qa-gate` — на интеграции волны.
- Docs-sync в той же сессии (`55-docs-sync`).
- `verified: true` только с логом green regression.
- TZ: Magic Link, UUID, UTC, optimistic `version`, Fiber WS, NATS (не Redis pub/sub), SeaweedFS, TanStack Router, Shadcn className-only.

**Playbooks:** `.cursor/playbooks/` (10 файлов) · ростер: `.cursor/agents/` · RACI: `team-workflow.md`.

### 1.3 Архитектурный горизонт (уже принято)

- `ADR-ARCH-MICROSERVICES-2026`: strangler-fig; каталог core + hub-ai + realtime-трио (`hub-ws` / `hub-messenger` / `hub-calls`) + пилот `hub-notification`; tracker **не** пилим.
- Calendar v2 (GZ-V4-05): iCal export + внешние iCal-подписки, **без OAuth** Google/Outlook.
- M4 `hub-calls` в roadmap сейчас завязан на **LiveKit-домен** — идея SFU-миграции ниже **пересматривает** этот пункт.

### 1.4 Факт по звонкам / календарю / почте (as-is)

| Область | Сейчас | Дыра относительно идеи пользователя |
|---------|--------|--------------------------------------|
| SFU | LiveKit v1.8 (compose/K3s), JWT room tokens, webhooks, egress recording, FE `@livekit/components-react` | Нет абстракции SFU-адаптера; миграция = переписать room/token/webhook/egress/UI |
| Scheduled calls | `ScheduledCall` + RRULE, channel/project, participants M2M, conflict check | Нет RSVP (accept/decline/tentative); нет персонального «моего» календаря-инвайта |
| Unified calendar | **Read-агрегатор**: scheduled calls + sprints + issue due + approved leave | Нет first-class `CalendarEvent`; нет двусторонней записи из календаря в звонок с инвайтами |
| Messenger → call | Instant call / invite code / channel call button | Планирование из чата не создаёт календарные инвайты с accept-flow |
| Email | Magic Link + hub-mail (transactional/admin) | Нет calendar invites (ICS в письме), нет sync почтового ящика как источника событий |
| Внешний sync | iCal export (частично/план v2) | OAuth Google/Outlook **вне scope** по GZ; идея «синхронизация почты» требует **нового ADR** |

---

## 2. Идея A — миграция LiveKit → mediasoup — **REJECTED (2026-07-10)**

> **Решение пользователя:** не переходим, **остаёмся на LiveKit**.  
> Причина: Hub уже имеет полный LiveKit-стек (rooms/JWT/webhooks/egress/FE components); mediasoup = перепись signaling+UI+recording без JTBD; команда 10–50 — типичный LiveKit fit. Пересмотр только при жёстком драйвере (лицензия/стоимость/SIP/кастомный RTP).  
> Исторически «лайвсуп» интерпретировался как mediasoup — вопрос закрыт.  
> Seed: `seeds/SEED-001-mediasoup-migration.md` = REJECTED. M4 `hub-calls` остаётся LiveKit-доменом.

---

## 3. Идея B — синхронизация почты / созвонов / календаря (RSVP)

### 3.1 Продуктовый сценарий (целевой)

> Человек ставит созвон в **мессенджере** или **календаре** → коллеги получают инвайт → при **Accept** событие появляется в календаре у обоих (и у организатора уже есть) → Decline/Tentative отражаются; конфликт времени предупреждается.

Это **шире**, чем GZ-V4-05 (только iCal feed). Нужен **новый продуктовый эпик** + возможный ADR на email/calendar sync.

### 3.2 Целевая модель (черновик для ba/sa)

```
CalendarEvent (first-class)
  id, title, description, starts_at, ends_at, rrule?, timezone=UTC store
  source: messenger | calendar | call | email | hr_leave | sprint | issue
  source_ref_id (scheduled_call_id / message_id / …)
  organizer_id
  location / call_link / channel_id?
  version (optimistic lock)

CalendarAttendee
  event_id, user_id, role: organizer|required|optional
  rsvp: needs_action|accepted|declined|tentative
  responded_at

Projection rules:
  - В «моём» календаре видны: events где я organizer ИЛИ rsvp=accepted|tentative
  - needs_action — inbox инвайтов / badge, не busy (или soft-busy — решить)
  - declined — скрыто из grid (остаётся в истории инвайта)
```

**Связь со ScheduledCall:** либо ScheduledCall становится specialization `source=call`, либо CalendarEvent — канон, а Call создаётся при наступлении `starts_at` (worker уже есть для due scheduled calls).

### 3.3 Потоки

| Вход | Действие | Выход |
|------|----------|-------|
| Messenger: «Запланировать созвон» | Create event + attendees (channel members или selected) + optional ScheduledCall | NATS `calendar.invite.created`; in-app + email ICS |
| Calendar UI: Create | То же; если «созвон» — channel/call link | То же |
| Accept | attendee.rsvp=accepted; invalidate calendar queries; busy/free update | WS push обоим |
| Decline | rsvp=declined; organizer notification | — |
| External iCal (GZ) | Import as read-only `source=external_ical` | Без RSVP write-back |
| Email sync (новое) | См. §3.5 — отдельный ADR | — |

### 3.4 Edge / corner cases

| # | Кейс | Предлагаемое решение (к утверждению) |
|---|------|--------------------------------------|
| C1 | Accept после того как организатор удалил событие | 404/gone; очистить inbox |
| C2 | Recurring: accept серии vs одного instance | v1: **серия целиком** (как GZ recurring edit); instance-exception — backlog |
| C3 | Два инвайта на одно время | Conflict warning (уже есть `FindConflictingCalls`); не блокировать hard без политики |
| C4 | Участник не в project/channel | Нельзя добавить / или guest-only link без calendar write |
| C5 | Organizer declines сам | Запрет; только cancel event |
| C6 | TZ: хранение UTC, показ local | NFR-12; ICS с VTIMEZONE |
| C7 | Partial update title mid-flight | Attendees с accepted получают update notification; rsvp не сбрасывать |
| C8 | Reschedule | Сброс rsvp→needs_action **или** сохранить accepted + «time changed» banner — **решить в discuss** |
| C9 | Soft-delete / GDPR | Корзина 30 дней; hard delete admin |
| C10 | Race: simultaneous accept/decline | Optimistic version на attendee row |
| C11 | Channel call без явных attendees | Default: все members channel = required; mute notifications opt-in |
| C12 | Leave/busy overlap | Availability API (US-CAL-04) учитывает accepted calls + leave |
| C13 | Email bounce / no mailbox | In-app invite остаётся source of truth |
| C14 | External organizer (email-only person) | v1 out of scope (нет multi-org guests) — только Hub users |
| C15 | Messenger thread delete ≠ cancel call | Явная связь; cancel только через event API |

### 3.5 «Синхронизация почты» — уровни (выбрать при возврате)

| Уровень | Что даёт | Сложность | Конфликт с GZ-V4-05 |
|---------|----------|-----------|---------------------|
| **L0** | ICS attachment в transactional email (hub-mail) при invite/update/cancel | S | Нет — дополняет |
| **L1** | iCal publish URL + subscribe (уже в плане Calendar v2) | M | Совпадает с GZ |
| **L2** | IMAP/JMAP ingest «письма-приглашения» → Hub events | L | Новый scope |
| **L3** | OAuth Google/Microsoft Calendar two-way | XL | **Противоречит GZ-V4-05** → нужен явный override ADR |

**Рекомендация brief:** старт с **L0+L1** + RSVP внутри Hub; L2/L3 — только после product-решения и ADR.

### 3.6 Порядок работ (когда вернёшься)

1. `product` + `ba2`: US/AC (Gherkin) для RSVP + messenger/calendar create; MoSCoW; KPI.
2. `sa2`: ERD CalendarEvent/Attendee, API, NATS subjects, ICS schema.
3. `arch`: ADR-CAL-UNIFIED (канон события) + опционально ADR-MAIL-CAL-SYNC (уровень L*).
4. `designer`: инвайт-inbox, accept UI, conflict badge — **по макетам Cloud Code Designer**.
5. be/fe параллельно; qa-manual TC на все C1–C15; security на email SSRF (ValidateOutboundURL для iCal fetch уже в духе NF-07).

---

## 4. План возврата конца июля — «перестройка»

Не продолжать Волну 2 вслепую. Рекомендуемый kickoff:

### Неделя 0 — Align (1–3 дня)

1. Ревью этого brief + seeds; подтвердить: mediasoup? уровень email L0–L3?
2. Смержить/закрыть хвосты Волны 1 (`integration/wave1` и др.) — решение пользователя.
3. FA-06 если ещё не сделан.
4. Забрать макеты Cloud Code Designer → `designer` inventory + gap vs текущий UI.

### Неделя 1 — Planning expansion

1. `tech-lead`: обновить `ROADMAP-FINAL` — вставить эпики:
   - **EPIC-SFU** (mediasoup migration)
   - **EPIC-CAL-SYNC** (unified calendar + RSVP + L0/L1 mail)
   - **EPIC-REDESIGN** (Cloud Code Designer → DESIGN-SPEC → fe waves)
2. Проработка **всех** edge/corner из §2.4 и §3.4 → TC-заготовки `qa-manual`.
3. Пересмотр M4: `hub-calls` после или вместе с SFU-adapter (зависимость!).
4. Gray-zone батч: reschedule-RSVP policy, L2/L3 mail, guest attendees.

### Дальше — Execute по обычному конвейеру

`ba → sa → arch → designer → be/fe → qa → *-ver → /qa-gate`  
Параллельные потоки A/B; team-lead оркестрирует; master только пользователь.

### Что сознательно не делать до align

- Не пилить LiveKit в коде «на всякий случай».
- Не внедрять OAuth calendar без ADR (ломает GZ-V4-05).
- Не смешивать SFU-миграцию и microservices M4 в одном PR.

---

## 5. Открытые вопросы к пользователю (чек-лист возврата)

1. «Лайвсуп» = **mediasoup**? Если нет — что именно?
2. Мотив миграции: cost / control / ops pain / другое?
3. Email sync: L0 / L1 / L2 / L3?
4. При reschedule: сбрасывать RSVP или оставлять accepted?
5. Редизайн: полный product shell или точечно calendar/calls/messenger?
6. Приоритет июля: SFU first vs Calendar-RSVP first vs Redesign first?
7. Волны 2–11 ROADMAP-FINAL — продолжаем каркас или freeze до нового milestone?

---

## 6. Индекс артефактов

| Артефакт | Путь |
|----------|------|
| Этот brief | `.planning/archive/july-2026/JULY-2026-REBUILD-BRIEF.md` |
| Seed SFU | `.planning/reference/seeds/SEED-001-mediasoup-migration.md` |
| Seed calendar sync | `.planning/reference/seeds/SEED-002-calendar-calls-email-sync.md` |
| Roadmap pointer | `ROADMAP-FINAL-2026-07-03.md` § «Отложено: July Rebuild» |
| GZ calendar | `../gray-zones/GRAY-ZONES-DECISIONS-2026-07-03.md` GZ-V4-05 |
| AI team | `.cursor/playbooks/README.md`, `team-workflow.md` |
| Calls as-is | `backend/internal/calls/`, `frontend/src/components/calls/` |
| Calendar as-is | `backend/internal/calendar/`, `frontend/src/modules/calendar/` |

---

*Зафиксировано агентом 2026-07-10 по запросу пользователя: изучить тонкости AI-команды, проработать идеи, задокументировать до возврата конца июля.*
