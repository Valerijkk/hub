---
phase: 05-voice-calls
verified: 2026-03-21T00:00:00Z
status: gaps_found
score: 2/4 must-haves verified
gaps:
  - truth: "Пользователь видит кнопку звонка в канале и может начать/присоединиться к звонку"
    status: failed
    reason: "CallButton и CallPanel существуют как компоненты, но не смонтированы нигде в дереве приложения (AppShell, __root.tsx, ни один маршрут их не импортирует). Запустить звонок невозможно."
    artifacts:
      - path: "frontend/src/components/calls/CallButton.tsx"
        issue: "Компонент не импортируется нигде в приложении — ORPHANED"
      - path: "frontend/src/components/calls/CallPanel.tsx"
        issue: "Компонент не импортируется нигде в приложении — ORPHANED"
    missing:
      - "Добавить <CallPanel /> в frontend/src/components/layout/app-shell.tsx (на уровне root для глобального оверлея)"
      - "Добавить <CallButton channelId={...} /> в топбар или хедер канала мессенджера"

  - truth: "Эфемерный чат звонка доступен во время звонка и исчезает после его завершения"
    status: failed
    reason: "Backend ChatRepository, SendChatMessage/ListChatMessages endpoints, NATS bridge и логика удаления при OnRoomFinished — всё реализовано. Фронтенд: компонент CallChat.tsx отсутствует, API-клиент (calls.ts) не содержит sendChatMessage/listChatMessages. Задача была явно отложена в SUMMARY Plan 03 как 'frontend files not modified per execution constraint'."
    artifacts:
      - path: "frontend/src/components/calls/CallChat.tsx"
        issue: "Файл отсутствует — MISSING"
      - path: "frontend/src/api/calls.ts"
        issue: "Отсутствуют функции sendChatMessage и listChatMessages — INCOMPLETE"
    missing:
      - "Создать frontend/src/components/calls/CallChat.tsx с UI эфемерного чата"
      - "Добавить sendChatMessage(callId, text) и listChatMessages(callId, limit?) в frontend/src/api/calls.ts"
      - "Интегрировать CallChat в CallRoom.tsx"

  - truth: "Пользователь может включать/выключать микрофон, поднимать руку"
    status: partial
    reason: "Mic toggle, screen share toggle и raise hand работают корректно через LiveKit хуки. Однако кнопка 'Завершить звонок' (EndCall) никогда не показывается создателю: frontend проверяет activeCall?.creator_id, а backend сериализует поле как created_by. Поле creator_id всегда undefined."
    artifacts:
      - path: "frontend/src/components/calls/CallControls.tsx"
        issue: "Строка 47: activeCall?.creator_id — поле не существует в CallResponse (backend возвращает created_by)"
      - path: "frontend/src/api/calls.ts"
        issue: "CallResponse интерфейс объявляет creator_id, но backend DTO отправляет created_by — несоответствие имён полей"
    missing:
      - "Исправить: либо переименовать json tag в backend/internal/calls/models.go с 'created_by' на 'creator_id', либо обновить frontend CallResponse тип и CallControls.tsx"
human_verification:
  - test: "Запустить LiveKit сервер и создать звонок через API"
    expected: "JWT токен возвращается, WebRTC соединение устанавливается, аудио передаётся"
    why_human: "Требует работающий Docker Compose с LiveKit, WebRTC браузерное подключение"
  - test: "После завершения звонка проверить что сообщения чата удалены из MongoDB"
    expected: "GET /api/calls/:callId/chat возвращает пустой список после room_finished"
    why_human: "Требует интеграционный тест с реальным LiveKit webhook"
---

# Phase 5: Voice Calls — Verification Report

**Phase Goal:** Пользователи могут присоединяться к аудиозвонкам с демонстрацией экрана для стендапов, демо и спонтанных обсуждений
**Verified:** 2026-03-21
**Status:** gaps_found
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths (Success Criteria from ROADMAP.md)

| #   | Truth                                                                                                          | Status      | Evidence                                                                                                             |
|-----|----------------------------------------------------------------------------------------------------------------|-------------|----------------------------------------------------------------------------------------------------------------------|
| 1   | Пользователь может начать или присоединиться к аудиозвонку с участием до 50 человек по ссылке или из канала   | ORPHANED    | CallButton/CallPanel существуют и работоспособны, но не смонтированы в AppShell или любом маршруте                 |
| 2   | Пользователь может включать/выключать микрофон, поднимать руку и демонстрировать экран (несколько участников)  | PARTIAL     | Mic, screen share, raise hand работают; EndCall кнопка сломана из-за creator_id vs created_by несоответствия        |
| 3   | Эфемерный чат звонка доступен во время звонка и исчезает после завершения                                     | FAILED      | Backend полностью реализован; frontend CallChat.tsx отсутствует; API клиент не содержит chat-функций               |
| 4   | События начала/окончания звонка отображаются в связанном канале мессенджера                                    | VERIFIED    | bridge.go подписан на calls.channel.> и публикует в channel:{channelID} WS-комнаты; код реален и полон            |

**Score:** 1/4 truths fully verified, 1/4 partial, 1/4 orphaned, 1/4 failed

### Required Artifacts

| Artifact                                                   | Expected                                      | Status     | Details                                                             |
|------------------------------------------------------------|-----------------------------------------------|------------|---------------------------------------------------------------------|
| `backend/ent/schema/call.go`                               | Ent-схема звонка                              | VERIFIED   | type Call struct с полями, рёбрами, индексами                        |
| `backend/ent/schema/callparticipant.go`                    | Ent-схема участника                           | VERIFIED   | type CallParticipant struct, unique index (call_id, user_id)         |
| `backend/internal/calls/service.go`                        | Бизнес-логика звонков                         | VERIFIED   | NewCallService, CreateCall, JoinCall, LeaveCall, EndCall, GenerateToken, SendChatMessage, ListChatMessages |
| `backend/internal/calls/handler.go`                        | HTTP-обработчики                              | VERIFIED   | NewHandler, RegisterRoutes, все эндпоинты включая /chat             |
| `backend/internal/calls/repository.go`                     | Ent репозиторий                               | VERIFIED   | CreateCall, GetCall, GetActiveCallByChannel, AddParticipant, RemoveParticipant |
| `backend/internal/calls/chat.go`                           | MongoDB эфемерный чат                         | VERIFIED   | ChatRepository, SendMessage, ListMessages, DeleteCallChat            |
| `backend/internal/ws/bridge.go`                            | NATS bridge с calls подписками                | VERIFIED   | calls.channel.> и calls.chat.> подписки с handleCallEvent/handleCallChatEvent |
| `deployments/livekit.yaml`                                 | Конфигурация LiveKit                          | VERIFIED   | port 7880, webhook URL, max_participants 50                          |
| `frontend/src/api/calls.ts`                                | API-клиент звонков                            | PARTIAL    | createCall, joinCall, leaveCall, endCall, getActiveCall, raiseHand — есть. sendChatMessage, listChatMessages — отсутствуют |
| `frontend/src/stores/call-store.ts`                        | Zustand store                                 | VERIFIED   | startCall, joinCall, leaveCall, endCall, toggleHand, checkActiveCall |
| `frontend/src/components/calls/CallRoom.tsx`               | Главный контейнер звонка                      | VERIFIED   | LiveKitRoom wrapper, audio=true, video=false, RoomAudioRenderer      |
| `frontend/src/components/calls/CallControls.tsx`           | Панель управления                             | PARTIAL    | Mic/screen share/raise hand работают; EndCall сломан (creator_id bug) |
| `frontend/src/components/calls/ParticipantList.tsx`        | Список участников                             | VERIFIED   | useParticipants, сортировка, mute/hand/screenshare индикаторы        |
| `frontend/src/components/calls/ScreenShareView.tsx`        | Просмотр демонстраций экрана                  | VERIFIED   | useTracks, VideoTrack, grid layout для нескольких демонстраций       |
| `frontend/src/components/calls/CallButton.tsx`             | Кнопка вызова                                 | ORPHANED   | Существует, функционален, но не используется нигде в приложении      |
| `frontend/src/components/calls/CallPanel.tsx`              | Оверлей звонка                                | ORPHANED   | Существует, функционален, но не монтируется нигде                    |
| `frontend/src/components/calls/CallChat.tsx`               | UI эфемерного чата                            | MISSING    | Файл не создан                                                       |

### Key Link Verification

| From                                              | To                           | Via                                        | Status      | Details                                                                          |
|---------------------------------------------------|------------------------------|--------------------------------------------|-------------|----------------------------------------------------------------------------------|
| `backend/internal/calls/service.go`               | LiveKit RoomServiceClient    | `lksdk.NewRoomServiceClient`               | VERIFIED    | Строка 34: `lksdk.NewRoomServiceClient(cfg.URL, cfg.APIKey, cfg.APISecret)`      |
| `backend/internal/calls/handler.go`               | `backend/internal/calls/service.go` | `h.service.*` method calls          | VERIFIED    | Все handlers вызывают h.service методы                                           |
| `backend/cmd/server/main.go`                      | `backend/internal/calls/handler.go` | `callsHandler.RegisterRoutes`        | VERIFIED    | Строки 169-176: полная цепочка инициализации и регистрации                       |
| `backend/internal/ws/bridge.go`                   | `calls.channel.>`            | NATS subscription                          | VERIFIED    | Строки 60-66: callSub подписка с handleCallEvent                                 |
| `backend/internal/ws/bridge.go`                   | `calls.chat.>`               | NATS subscription                          | VERIFIED    | Строки 70-76: callChatSub подписка с handleCallChatEvent                         |
| `backend/internal/calls/service.go`               | `ChatRepository.DeleteCallChat` | `OnRoomFinished` cleanup               | VERIFIED    | Строки 217-221: вызов `s.chatRepo.DeleteCallChat(ctx, callEnt.ID.String())`      |
| `frontend/src/components/calls/CallRoom.tsx`      | LiveKitRoom                  | `@livekit/components-react`                | VERIFIED    | Строка 30: `<LiveKitRoom ... />`                                                  |
| `frontend/src/components/calls/CallControls.tsx`  | `useCallStore`               | zustand store                              | VERIFIED    | Строка 31: `useCallStore()` деструктурирован                                     |
| `frontend/src/api/calls.ts`                       | `/api/calls`                 | apiClient fetch                            | VERIFIED    | Все 6 функций используют `api.post`/`api.get`                                    |
| `frontend/src/components/calls/CallPanel.tsx`     | App tree                     | import в AppShell или root layout          | NOT_WIRED   | Не импортируется нигде вне собственного файла                                    |
| `frontend/src/components/calls/CallButton.tsx`    | App tree                     | import в channel UI                        | NOT_WIRED   | Не импортируется нигде вне собственного файла                                    |

### Requirements Coverage

| Requirement | Description                                                                                           | Status          | Evidence                                                                                       |
|-------------|-------------------------------------------------------------------------------------------------------|-----------------|-----------------------------------------------------------------------------------------------|
| CALL-01     | Аудиозвонки и демонстрация экрана (без веб-камер), до 50 участников                                  | PARTIAL         | Backend: CreateRoom с MaxParticipants=50, LiveKit в docker-compose. Frontend: CallButton/CallPanel orphaned — пользователь не может начать звонок через UI |
| CALL-03     | Включение/выключение микрофона, поднятие руки, выбор экрана/окна, одновременная демонстрация нескольких экранов | PARTIAL  | Mic/screen share/raise hand реализованы. EndCall (вторичный для CALL-03) сломан из-за creator_id bug |
| CALL-06     | Эфемерный чат звонка (исчезает после завершения звонка)                                               | PARTIAL         | Backend: ChatRepository, endpoints, NATS bridge, DeleteCallChat при OnRoomFinished — VERIFIED. Frontend: нет CallChat.tsx и chat API-функций |

### Anti-Patterns Found

| File                                                              | Line | Pattern                                      | Severity   | Impact                                                                                      |
|-------------------------------------------------------------------|------|----------------------------------------------|------------|---------------------------------------------------------------------------------------------|
| `frontend/src/components/calls/CallControls.tsx`                  | 47   | `activeCall?.creator_id` — несуществующее поле | Blocker  | EndCall кнопка никогда не отображается создателю звонка; creator не может завершить звонок  |
| `frontend/src/api/calls.ts`                                       | 27   | `CallResponse.creator_id` — несоответствие с backend | Blocker | Backend возвращает `created_by`, frontend ожидает `creator_id` |
| `frontend/src/components/calls/CallPanel.tsx`                     | —    | Orphaned component (not mounted in any route) | Blocker  | Пользователь не видит активный звонок                                                       |
| `frontend/src/components/calls/CallButton.tsx`                    | —    | Orphaned component (not mounted in any route) | Blocker  | Пользователь не может начать или присоединиться к звонку                                   |

### Human Verification Required

#### 1. WebRTC Audio Connection Test

**Test:** Запустить Docker Compose, создать двух пользователей, использовать API напрямую (POST /api/calls, POST /api/calls/:id/join) для получения токенов, подключиться через LiveKit client SDK
**Expected:** Оба пользователя слышат друг друга; mic toggle работает в реальном времени
**Why human:** Требует работающий Docker Compose с LiveKit, браузерное WebRTC соединение

#### 2. Screen Share Picker Test

**Test:** После исправления orphaned компонентов — нажать кнопку демонстрации экрана
**Expected:** Браузер показывает системный пикер для выбора экрана/окна; демонстрация отображается в ScreenShareView у других участников
**Why human:** Требует живой браузер и реальный WebRTC

#### 3. Ephemeral Chat Disappears Test

**Test:** Отправить сообщения в чат во время звонка, завершить звонок через DeleteRoom/EndCall, проверить GET /api/calls/:callId/chat
**Expected:** Список сообщений пуст после завершения звонка
**Why human:** Требует интеграционный тест с реальным LiveKit webhook room_finished

### Gaps Summary

Фаза 5 реализовала полноценный backend модуль голосовых звонков (LiveKit интеграция, Ent-схемы, REST API, MongoDB эфемерный чат, NATS bridge) — этот слой верифицирован и работоспособен.

Однако три проблемы блокируют достижение цели фазы:

**1. Orphaned UI компоненты (критично):** CallButton и CallPanel созданы и функциональны, но не встроены ни в один маршрут или layout. Пользователь физически не может инициировать звонок через интерфейс. Исправление: добавить `<CallPanel />` в `app-shell.tsx` и `<CallButton channelId={...} />` в UI канала.

**2. Отсутствующий CallChat фронтенд (CALL-06 частично не выполнен):** Backend эфемерного чата реализован полностью, но frontend компонент CallChat.tsx и chat-функции в API-клиенте не созданы. Это было явно задокументировано в SUMMARY Plan 03 как отложенная задача ("frontend files not modified per execution constraint"), но фаза закрыта без их добавления.

**3. Поле `creator_id` vs `created_by` (мелкий баг, но блокирует функционал):** Backend сериализует поле как `created_by` (json tag в models.go), frontend ожидает `creator_id`. Кнопка "Завершить звонок" никогда не показывается создателю. Простое однострочное исправление в одном из двух мест.

---

_Verified: 2026-03-21_
_Verifier: Claude (gsd-verifier)_
