# W2 — calls + LiveKit (be-ver + security)

> **Provenance class:** PRE_EXISTING · **Confidence:** MIXED · **current_verified:** false (pending code/source recheck)
> Embedded original verdict below = historical self-check only — **not** independent acceptance. See [`PROVENANCE.md`](./PROVENANCE.md).

> Дата: 2026-07-23 · Агент: be-ver/security (Grok 4.5) · Режим: **readonly** (код не менялся)
> Scope: `backend/internal/calls/**`, `deployments/livekit.yaml`, `deployments/livekit-prod.yaml.example`, refs в compose/gitignore
> Фокус: no mediasoup · secrets not in git · room auth

## Вердикт (сжатый)

| Инвариант | Статус |
|-----------|--------|
| SFU = LiveKit, нет mediasoup в коде calls/deploy | ✅ OK |
| Prod secrets не в tree (`livekit-prod.yaml` отсутствует; только `.example`) | ✅ OK (history/gitignore — см. findings) |
| Webhook LiveKit: JWT + обязательный `sha256` (SEC-02) | ✅ OK |
| Room auth (Join/End) — membership / creator | ✅ частично OK |
| AuthZ на Create / chat / recording / scheduled / GetActive | ❌ gaps (CRITICAL/HIGH) |
| Регрессия `go test -tags hubtest -race ./testframework/calls/...` | ⚠️ **не запускалась** (Ask/readonly scan; нет лога) |

**verified: false** — 2026-07-23 00:58 (UTC+3) — блокируют CRITICAL/HIGH room-auth IDOR; полный hubtest не прогонялся в этой сессии.

---

## Findings

| Sev | ID | Файл:строка | Evidence | Fix |
|-----|-----|-------------|----------|-----|
| **CRITICAL** | C-01 | `service.go:48-66` · `handler_calls.go:32-54` | `CreateCall` создаёт LiveKit room + БД + токен создателю **без** `IsChannelMember`. Любой auth-user с UUID чужого канала получает `RoomJoin` JWT. | Перед `provisionRoom`: `IsChannelMember(ctx, channelID, userID)`; иначе `ErrCallForbidden`. Негативный тест. |
| **HIGH** | H-01 | `handler_calls.go:137-155` · `service.go:184-194` | `GetActiveCall` / `GetActiveCallByChannel` без membership: утечка `id`, `room_name`, `created_by`, `invite_code`, status по любому `channelId`. | `IsChannelMember` до query; non-member → 403/404. |
| **HIGH** | H-02 | `service_chat.go:33-79` · `handler_calls.go:311-377` | `SendChatMessage` / `ListChatMessages`: только «call active», **нет** participant/channel membership. Любой auth-user R/W чата по `callId`. | `authorizeCallAccess` (member канала **или** участник / invite path) до send/list. Denial-suite. |
| **HIGH** | H-03 | `recording.go:89-168` · `handler_recording.go:32-48` | `StartRecording`: нет проверки членства канала / RBAC; достаточно знать `callId` активного канального звонка → Egress + S3. | AuthZ до Egress (канал/project perm). 403-тест. |
| **HIGH** | H-04 | `recording.go:173-201` · `handler_recording.go:63-80` | `StopRecording(recordingID)` **без** `userID`/project RBAC — любой auth-user останавливает чужой egress. | Передавать userID; RBAC/`RecordingProjectID` + perm (как download). |
| **HIGH** | H-05 | `scheduled.go:34-102` · `handler_scheduled.go:34-142` | CRUD scheduled (`Create`/`List`/`Update`/`Delete`) без project/channel membership. Чужой `projectId`/`channelId`/`id` → create/list/mutate/delete. | RBAC project member (+ owner для mutate); List — только member канала. |
| **HIGH** | H-06 | `handler_scheduled.go:158-199` · `scheduled.go:172+` | `CheckConflicts`: любой auth-user передаёт чужие `participant_ids` → утечка расписания (privacy). | Ограничить IDs org/project peers или self-only. |
| **HIGH** | H-07 | `handler_calls.go:284` · `GuestJoin` `401-423` · `service.go:74` | Layering: handler лезет в `h.service.repo` (RaiseHand, GuestJoin). Invite = `uuid.New().String()[:8]` (~32 bit hex) + **unauth** `POST /calls/guest-join` → полный `RoomJoin` guest JWT (4h). | Сервисный `authorize*`; invite ≥ 128-bit / rate-limit guest-join; guest grant ужесточить (см. M-03). |
| **MEDIUM** | M-01 | `.gitignore` (нет `livekit-prod.yaml`) | В tree только `livekit.yaml` + `livekit-prod.yaml.example`; `livekit-prod.yaml` **ABSENT**. Файл **не** в `.gitignore` → риск повторного commit plaintext (SEC-03 класс). History: PG-03 / FA-06. | Добавить `deployments/livekit-prod.yaml` в `.gitignore`; не монтировать в prod (heredoc остаётся каноном). |
| **MEDIUM** | M-02 | `config.go:186-187` vs `livekit.yaml:10-12` | Backend default `devkey`/`devsecret`; LiveKit local keys `devkey-hub-local-development-2026` / `devsecret-hub-local-…`. Dev compose не прокидывает `LIVEKIT_*` → рассинхрон keys без `.env`. Prod: `ValidateProductionConfig` режет `devkey`/`devsecret` — OK. | Синхронизировать defaults/`deployments/.env.example` с `livekit.yaml`; document pair. |
| **MEDIUM** | M-03 | `service_room.go:125-136` · `handler_calls.go:426-438` | `VideoGrant{RoomJoin:true}` без явных `CanPublish`/`CanSubscribe`/hidden; токен user 24h, guest 4h — широкие права/TTL. | Явные grant flags; короткий TTL; guest: ограничить publish при политике «listen-only». |
| **MEDIUM** | M-04 | `service.go:25-30` · `repository.go:18-19` | Нет `Repository` interface: `CallService` держит `*EntRepository` (отклонение от канона layering; `*ent.Client` только в repo — OK). | Ввести interface + mock для unit без Ent. |
| **MEDIUM** | M-05 | `handler_recording.go:109-121,172-189` | RBAC download/list только при `h.rbacSvc != nil`. Wiring передаёт rbac (`wiring_integrations.go:54`) — prod OK; паттерн fail-open при nil — риск тестов/miswire. | Fail-closed: `rbacSvc == nil` → 500/forbid вне test build tag. |
| **MEDIUM** | M-06 | `handler_calls.go:268-294` | `RaiseHand`: нет Hub-authZ; server SDK `UpdateParticipant` с API key может трогать metadata identity=userID в любом `room` по `callId`. | Требовать active participant / channel member до UpdateParticipant. |
| **LOW** | L-01 | `service.go:143-154` | `LeaveCall` без проверки участия — идемпотентный no-op в repo, низкий риск. | Опционально: 404 если не участник. |
| **LOW** | L-02 | `recording.go:161` subjects | NATS `calls.recording.*` / `calls.scheduled.due` напрямую `PublishJSON`, не outbox — вне TZ pub/sub JetStream-канона для доменных событий. | Align с outbox+JetStream subjects `hub.calls.*` при следующей волне. |

---

## Positive (не findings)

| Проверка | Результат |
|----------|-----------|
| mediasoup | Нет импортов/refs в `internal/calls/` и `deployments/livekit*`. SEED-001 REJECTED соблюдён. |
| Secrets in git (working tree) | Tracked: `livekit.yaml` (явные **local** keys), `livekit-prod.yaml.example` (`${LIVEKIT_API_*}`). `livekit-prod.yaml` не tracked, файл отсутствует. |
| Prod LiveKit | `docker-compose.prod.yml:129-152` — heredoc из env + `departure_timeout: 120` (INF-01 закрыт в compose). |
| Webhook | `handler.go:156-273`: пустой Auth → false; HMAC JWT; `iss` via `constantTimeStringEqual`; **sha256 claim обязателен** (SEC-02). |
| Join/End room auth | `JoinCall` channel → `IsChannelMember`; instant → только creator (`service.go:111-131`). `EndCall` creator или channel owner/admin (`156-172`). Тесты: `TestJoinCall_NonMemberForbidden`, `TestHandler_JoinCall_Forbidden`, `TestHandler_EndCall_StrangerForbidden`. |
| Recording download/list RBAC | CR-02 в handler + `handler_rbac_recording_test.go` ForbiddenNonMember. |
| WS call rooms | `ws/authorizer.go:67-71` — `call:` fail-closed при nil checker (вне модуля, смежно OK). |
| UUID parse / CopyString | `parseCallID` + `utils.CopyString` на params — ME-12/fasthttp OK. |
| SeaweedFS egress | `aws-sdk-go-v2` + path-style — TZ storage OK. |

---

## Layering / security checklist (be-ver §3–§5)

| Пункт | OK/FAIL |
|-------|---------|
| `*ent.Client` только в EntRepository | OK (`repository.go:19`) |
| Handler без прямого Ent | FAIL — `h.service.repo` в RaiseHand/GuestJoin |
| AuthZ до мутации (Create/StartRecording/Stop/Scheduled) | FAIL — C-01, H-03..H-05 |
| Owner/member на Get (active call, chat list) | FAIL — H-01, H-02 |
| Webhook подпись без skip | OK |
| Constant-time secret compare (iss) | OK |
| SafeHTTP / user URL | N/A (нет user-controlled egress URL) |
| panic в prod paths | OK (не найдено) |

---

## Регрессия

```
не запускалась в этой сессии (readonly repo-scan / Ask).
Рекомендация team-lead / be2:
  cd backend && go test -tags hubtest -race -count=1 -timeout=600s ./testframework/calls/...
```

Denial-покрытие сейчас: Join/End + recording download/list. **Нет** denial на CreateCall, chat, Start/StopRecording, scheduled CRUD, GetActiveCall.

---

## Handoff

→ `be2` (поток B / calls): закрыть C-01 + H-01..H-07 до merge любых calls-изменений.
→ `security-auditor`: deep-dive guest-join brute-force + grant flags (эскалация, не закрыто здесь).
→ `devops-engineer`: `.gitignore` для `livekit-prod.yaml` + sync local LIVEKIT env ↔ `livekit.yaml` (M-01, M-02).
→ `code-reviewer`: после фиксов authZ.

**Вердикт: verified: false — 2026-07-23 00:58**
