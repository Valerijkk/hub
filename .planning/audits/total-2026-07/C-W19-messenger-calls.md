# C-W19 — Red-team: messenger / ws / calls / notification / mail

> TASK-008 · Stage C · Wave 19 · READ-ONLY · 2026-07-29  
> Агент: независимый red-team (`cursor-grok-4.5`) · **не автор** фиксов TASK-143…147 / 162  
> HEAD проверен: `617af223` · Audit baseline A3: `ee353287`  
> Scope: `messenger` · `ws` · `calls` · `notification` · `mail`  
> Выход: **только этот файл** · коммит **не** делать · **TASK-008 остаётся открытым**

---

## Вердикт

| Вопрос | Ответ |
|--------|--------|
| Gaps A3W3*/A3W7-1 **не** в backlog? | **0** — все 17 open ID из CSV → TASK-143…147, TASK-162 |
| P0/P1 orphans в периметре? | **0** (A3 open были только P2/P3) |
| Комбо 1–3 на HEAD | **MITIGATED** в prod-коде; residual — process/ops, не silent IDOR |
| Attack chains member/non-member | Классический IDOR NF-02/HI-01 **закрыт**; остались by-design public join + ops DoS |

**Итог для Stage C:** постановка бэклога по A3W3 + A3W7-1 **полная**. Код на HEAD уже содержит remediation (часто с маркерами `A3W3-*` / `A3W7-1-*`), но README ещё держит TASK-143…147 как «Открыта» при статусе файлов «На проверке» / evidence у 162 — **нельзя закрывать TASK-008** без независимой мутации (W18-класс) и синхронизации статусов.

---

## 1. Gaps: A3W3* / A3W7* → backlog

### 1.1 Периметр findings (CSV + отчёты)

| Источник | Open ID (A3) | Sev | Backlog | Coverage |
|----------|--------------|-----|---------|----------|
| A3W3-1 | A3W3-1-001 | P2 | [TASK-143](../../backlog/archive/2026-07-29/TASK-143-2026-07-29-a3-3-1-messenger-handlers.md) | ✅ |
| A3W3-2 | A3W3-2-001…006 | P2×3 / P3×3 | [TASK-144](../../backlog/archive/2026-07-29/TASK-144-2026-07-29-a3-3-2-messenger-services.md) | ✅ 6/6 |
| A3W3-3 | A3W3-3-001…003 | P2 / P3×2 | [TASK-145](../../backlog/archive/2026-07-29/TASK-145-2026-07-29-a3-3-3-messenger-repos.md) | ✅ 3/3 |
| A3W3-4 | A3W3-4-001 | P3 | [TASK-146](../../backlog/archive/2026-07-29/TASK-146-2026-07-29-a3-3-4-ws.md) | ✅ |
| A3W3-5 | A3W3-A5-001…004 | P2 / P3×3 | [TASK-147](../../backlog/archive/2026-07-29/TASK-147-2026-07-29-a3-3-5-calls.md) | ✅ 4/4 |
| A3W7-1 | A3W7-1-001…002 | P3×2 | [TASK-162](../../backlog/archive/2026-07-29/TASK-162-2026-07-29-a3-7-1-notification-mail.md) | ✅ 2/2 |
| **Σ** | **17** | | **6 TASK** | **missed = 0** |

Дедуп-канон: [`A3-DEDUP-2026-07-29.md`](A3-DEDUP-2026-07-29.md) строки TASK-143…147, 162.

### 1.2 A3W7-2…5 (вне product-scope этого слота)

`A3W7-2` (v4) · `A3W7-3` (integrations) · `A3W7-4` (storage/dash/backup) · `A3W7-5` (mp/status/dev) — **не** messenger/ws/calls/notification/mail. В бэклоге как TASK-163…166. Для C-W19 **не** orphans слота; отмечено, чтобы не путать с «дырой A3W7*».

### 1.3 Процессный gap (не finding-orphan)

| Симптом | Evidence |
|---------|----------|
| README: TASK-143…147 = **Открыта** | `.planning/backlog/README.md` |
| Файлы задач: **На проверке** | TASK-143…147 frontmatter |
| Prod-код уже с A3-маркерами | см. §2 (Join/ListMembers/DeleteRoom/project_id/Redis) |
| TASK-162: evidence FIXED в файле, статус «На проверке» | checklist A3W7-1-001/002 |

→ Red-team: **пробела «не завели задачу» нет**; есть риск ложного «ещё открыто» без W18-мутации независимым агентом.

---

## 2. Комбо (X + Y тяжелее суммы)

### C-W19-C1 — Join fail-closed × ListMembers soft-delete

| | A3W3-2-001 | A3W3-3-002 |
|--|------------|------------|
| Было | `isMember, _ := IsMember` → DB err трактуется как non-member → AddMembers / шум | `ListMembers` без `HasChannelWith(DeletedAtIsNil)` → roster soft-deleted канала |
| Комбо (исторический) | Во время DB blip Join мог писать membership; после soft-delete канала row остаётся → ListMembers отдавал состав **корзины** | Severity **> P2+P3**: ghost join + PII roster из trash |
| HEAD | `service_channels.go:215-218` — err → Internal | `ent_repository.go:251-255` — soft-delete predicate |
| Service ListMembers | `IsMember` fail-closed + Forbidden non-member (`:361-368`) | GetChannel soft-deleted → NotFound до Join |

**Вердикт комбо:** **MITIGATED**.  
**Residual:** строки `channel_member` после soft-delete **не** hard-purge — любой будущий query без `HasChannelWith(DeletedAtIsNil)` регрессирует C1. Рекомендация: regression IT «DeleteChannel → ListMembers empty» (уже в ACCEPTANCE TASK-145) обязателен на мутации.

### C-W19-C2 — LiveKit orphan × recording `project_id`

| | A3W3-A5-001 | A3W3-A5-003 |
|--|-------------|-------------|
| Было | `provisionRoom` OK + DB CreateCall fail → **без** `DeleteRoom` | `GetProjectIDByChannel` → `(Nil, nil)` → CreateRecording FK **500** |
| Комбо (исторический) | Orphan room без call row (нет EndCall cleanup) **плюс** member канала без project получает 500/шум при StartRecording; давление на LiveKit + путаница ops | Не даёт JWT на orphan (нет call), но DoS SFU + ложные 500 на «живых» call |
| HEAD | `service.go:105-107`, `:140-141` — `deleteRoomBestEffort` | `repository.go:337-338` NotFound; `recording.go:131-138` → `ErrInvalidInput` |
| Instant | recording явно отказан (`ChannelID == nil` → «не поддерживают запись») | — |

**Вердикт комбо:** **MITIGATED** для IDOR/500.  
**Residual (ops/DoS, MEDIUM):**

1. `deleteRoomBestEffort` — при fail LiveKit DeleteRoom только `slog.Error` → orphan **всё ещё возможен** (best-effort). Attacker с правом CreateCall может триггерить DB fail (constraints/quota) и копить `call-<uuid>` rooms, если DeleteRoom нестабилен.
2. Канал без `project_id`: CreateCall **разрешён**, StartRecording → 400 — не orphan, но мёртвый feature path (product smell, не authz hole).

### C-W19-C3 — DailyLimit Redis fail-closed

| | A3W7-1-001 |
|--|------------|
| Было | process-local `perEmail` → N реплик × limit; restart reset |
| HEAD | `RedisDailySlotStore` Lua INCR+EXPIREAT; Redis err → **fail-closed** (`daily_slot.go:93-106`) |
| Wiring Hub | `wiring_audit_ai.go:50-51` — всегда `NewRedisDailySlotStore(d.rdb)` |
| Wiring hub-mail | `REDIS_URL` set → Redis + exit(1) on connect fail; **unset** → Memory + `slog.Warn` |

**Вердикт комбо:** **MITIGATED** на пути Hub server.  
**Residual (HIGH ops / LOW auth):**

1. **Fail-closed = DoS alerts:** Redis outage → все `SendAlert` fail (намеренно; не обход лимита).
2. **Split-brain:** Hub на Redis + hub-mail без `REDIS_URL` → **два независимых счётчика** → эффективный лимит ×2. Deploy must pin `REDIS_URL` на hub-mail (уже warn в коде).
3. Memory fallback в `NewService` — безопасен только для unit; prod без `SetDailySlotStore` = regress A3W7-1-001.

---

## 3. Attack chains — member / non-member

Threat actor: обычный authenticated user (не founder/admin), знает UUID чужих ресурсов.

### AC-1 Non-member → private channel (HI-01)

| Шаг | Ожидание HEAD | Статус |
|-----|---------------|--------|
| `GET /channels/{id}` | Forbidden (IsMember) | ✅ blocked |
| `ListMembers` | Forbidden | ✅ blocked |
| Soft-deleted channel ListMembers | IsMember false (DeletedAtIsNil) → Forbidden; repo empty | ✅ blocked (C1) |
| WS `subscribe channel:{id}` | `CanSubscribe` → deny frame (NF-02) | ✅ blocked |
| Join private | Forbidden «только invite» | ✅ blocked |
| Join public | **Allowed by design** → затем member | ⚠️ by-design (org-wide public) |

**Цепочка:** non-member **не** читает private roster/messages через API/WS. Единственный «вход» — public Join или invite (TASK-108).

### AC-2 Non-member → calls / recording

| Шаг | Ожидание HEAD | Статус |
|-----|---------------|--------|
| CreateCall чужого канала | `IsChannelMember` → `ErrCallForbidden` | ✅ |
| StartRecording чужого call | `authorizeCallRecordingAccess` → Forbidden | ✅ |
| StartRecording канал без project | `ErrInvalidInput` / 400 (не 500) | ✅ (C2) |
| Instant recording | отказ «не поддерживают» | ✅ |
| GuestJoin | RL 10/min + invite invite; guest_name cap | ✅ (A3W3-A5-004) |
| WS `call:{id}` | participant OR channel member | ✅ (A2-007) |

**Цепочка:** non-member не стартует egress и не получает call events. Residual C2 — только orphan DoS при fail DeleteRoom.

### AC-3 Member злоупотребления (внутри канала)

| Шаг | Риск | HEAD |
|-----|------|------|
| Cross-channel bookmark (`channelID=B`, msg from A) | мусор / смешение | ✅ `msg.ChannelID == channelID` → 404 (A3W3-2-003) |
| Poll results VoterIDs | утечка non-member | ✅ GetPollResults IsMember |
| ClosePoll после leave | closed+403 | ✅ membership **до** Close (A3W3-2-005) |
| Presence GetPresence чужих UUID | org-wide leak | ✅ FilterUsersSharingChannel (TASK-108) |

### AC-4 Notification / mail

| Шаг | Риск | HEAD |
|-----|------|------|
| WebPush SSRF | NF-07 | ✅ ValidateOutboundURL save+send (FIXED-OK A3; residual To4 — вне слота, W19-A4) |
| Telegram login | ADR-AUTH-01 | ✅ bind-only |
| Mail daily bypass multi-replica | ×N limit | ✅ Redis shared на Hub; ⚠️ hub-mail без REDIS_URL |
| Mail Redis down | open limit? | ✅ fail-closed (не send) |

### AC-5 WS ticket / query JWT

| Шаг | Риск | HEAD |
|-----|------|------|
| `?token=` + valid ticket | JWT в URI/proxy logs | prod deny `handler.go` query token; тест query+ticket — TASK-146 (A3 отчёт FIXED) |
| Query-only JWT | handshake JWT | ✅ 401 (TASK-103) |

---

## 4. Сводка severity (только residual после remediation)

| ID | Sev | Тип | Суть |
|----|-----|-----|------|
| C-W19-R1 | P2 ops | residual C2 | DeleteRoom best-effort → LiveKit orphan DoS под нагрузкой DB fail |
| C-W19-R2 | P2 ops | residual C3 | hub-mail без `REDIS_URL` → split-brain daily_limit |
| C-W19-R3 | P3 | process | README «Открыта» vs код FIXED / задачи «На проверке» — ложный backlog signal |
| C-W19-R4 | P3 | design | Public channel Join = org-wide member (не bug; осознанный threat model) |

Новых **P0/P1 product IDOR** в периметре **не** найдено. Orphan A3 findings → backlog: **0**.

---

## 5. Evidence (file:line, HEAD `617af223`)

| Проверка | Путь |
|----------|------|
| Join IsMember fail-closed | `backend/internal/messenger/service_channels.go:215-218` |
| ListMembers soft-delete | `backend/internal/messenger/ent_repository.go:251-255` |
| ListMembers service HI-01 | `backend/internal/messenger/service_channels.go:361-368` |
| LiveKit DB-fail DeleteRoom | `backend/internal/calls/service.go:105-107`, `:140-141`, `:168-172` |
| project_id nil → NotFound/400 | `backend/internal/calls/repository.go:337-338`; `recording.go:129-138` |
| DailyLimit Redis fail-closed | `backend/internal/mail/daily_slot.go:93-106` |
| Hub Redis wiring | `backend/cmd/server/wiring_audit_ai.go:50-51` |
| hub-mail REDIS_URL warn | `backend/cmd/hub-mail/main.go:23-36` |
| WS NF-02 | `backend/internal/ws/handler.go:244-269`; `authorizer.go` fail-closed |
| Poll test URL | `handler_wave114_test.go` → `GET /api/messenger/polls/not-a-uuid` |
| Dedup map | `A3-DEDUP-2026-07-29.md` TASK-143…147, 162 |

Мутацию prod **не** выполнял (READ-ONLY Stage C red-team). Для закрытия задач нужна отдельная W18-style мутация **не** автором фикса.

---

## 6. Handoff

1. **Не закрывать TASK-008.**
2. Orphans A3W3/A3W7-1 → backlog: **нет**; новые TASK не требуются для gaps.
3. Приоритет residual: **C-W19-R2** (deploy `REDIS_URL` на hub-mail) → **C-W19-R1** (наблюдаемость/алерт orphan LiveKit rooms / retry DeleteRoom) → **C-W19-R3** (синхрон README статусов после независимой мутации).
4. Перед «Исправлена» на TASK-143…147/162: независимый агент должен сломать ровно один guard на ID и показать красный hubtest (правило Stage C).
5. Public Join (R4) — не чинить без product/ADR; зафиксировать в threat model при желании.

---

## Verification block

```
verified: false
reason: Stage C red-team READ-ONLY; mutation not run; TASK-008 must stay open
coverage: A3W3-1..5 + A3W7-1 findings → TASK map 17/17; combos C1–C3 assessed on HEAD
orphans_p0_p1: 0
new_task_required: false (unless owner wants TASK for R1/R2 ops)
```
