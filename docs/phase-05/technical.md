# Фаза 5: Голосовые звонки -- Техническая документация

## Архитектура

### Диаграмма взаимодействия

```
+------------------+       REST API        +------------------+      LiveKit SDK     +------------------+
|                  |  (токены, комнаты)     |                  |  (управление rooms)  |                  |
|   React Client   | --------------------> |   Hub Backend    | --------------------> |  LiveKit Server  |
|   (Browser)      |                       |   (Go/Fiber)     |                       |  (WebRTC SFU)    |
|                  |                       |                  | <-------------------- |                  |
|                  |                       |                  |     Webhooks (JWT)     |                  |
+------------------+                       +------------------+                       +------------------+
        |                                         |                                          |
        |        WebSocket (сигнализация)          |                                          |
        | ---------------------------------------->|  NATS pub/sub                            |
        |                                         |  (события звонков)                        |
        |        UDP (медиа: аудио, screen)        |                                          |
        | ------->----->----->----->----->----->----|----------------------------------------->|
        |                                                                                     |
        +-------------------------------------------------------------------------------------+
                           WebRTC (UDP порты 50000-50100)
```

**Потоки данных:**
1. **REST API** -- клиент обращается к Hub Backend для создания/присоединения/выхода/завершения звонка и получения JWT-токена LiveKit
2. **LiveKit SDK** -- бэкенд управляет комнатами через `livekit/server-sdk-go/v2` (CreateRoom, DeleteRoom, UpdateParticipant)
3. **WebRTC** -- клиент подключается к LiveKit напрямую через WebSocket (сигнализация, порт 7880) и UDP (медиа, порты 50000-50100)
4. **Webhooks** -- LiveKit отправляет события жизненного цикла (room_started, room_finished, participant_joined, participant_left) на POST /api/calls/webhook
5. **NATS** -- бэкенд публикует события звонков для доставки через WebSocket bridge в браузеры

### Caddy прокси

Caddy проксирует **только** сигнальный WebSocket LiveKit:

```
/livekit/* -> localhost:7880
```

UDP-порты 50000-50100 экспонируются напрямую в Docker Compose (не через Caddy), так как WebRTC медиа идёт по UDP и не требует HTTP-проксирования.

## Структура модуля: Backend

### Пакет `backend/internal/calls/`

| Файл | Назначение |
|------|-----------|
| `handler.go` | HTTP-обработчики (Fiber) для всех эндпоинтов звонков + webhook handler |
| `service.go` | Бизнес-логика: управление комнатами LiveKit, генерация токенов, обработка webhook-событий, NATS-публикация |
| `repository.go` | `EntRepository` -- доступ к данным звонков через Ent ORM (PostgreSQL) |
| `chat.go` | `ChatRepository` -- эфемерный чат через MongoDB (коллекция `call_chat_messages`) |
| `models.go` | DTO: CreateCallRequest, JoinCallRequest, RaiseHandRequest, CallResponse, CallParticipantResponse |
| `service_test.go` | Unit-тесты для генерации токенов, nil-safe NATS, верификации webhook |
| `chat_test.go` | Unit-тесты для ChatRepository (MongoDB операции) |

### Ent-схемы

**Call** (`backend/ent/schema/call.go`):

| Поле | Тип | Описание |
|------|-----|----------|
| id | uuid.UUID | Первичный ключ |
| room_name | string | Имя комнаты LiveKit (`call-{UUID}`) |
| channel_id | uuid.UUID | FK на канал мессенджера |
| created_by | uuid.UUID | FK на пользователя-создателя |
| status | enum(active, ended) | Статус звонка |
| started_at | *time.Time | Время начала (от webhook room_started) |
| ended_at | *time.Time | Время завершения |
| participant_count | int | Текущее количество участников |

Edges: Channel (M2O), Creator/User (M2O), Participants (O2M -> CallParticipant)

**CallParticipant** (`backend/ent/schema/callparticipant.go`):

| Поле | Тип | Описание |
|------|-----|----------|
| id | uuid.UUID | Первичный ключ |
| call_id | uuid.UUID | FK на звонок |
| user_id | uuid.UUID | FK на пользователя |
| joined_at | time.Time | Время присоединения (default: now) |
| left_at | *time.Time | Время выхода (nil = ещё в звонке) |

Unique index: (call_id, user_id, left_at IS NULL) -- один активный участник на звонок.

### MongoDB: коллекция `call_chat_messages`

| Поле | Тип | Описание |
|------|-----|----------|
| _id | ObjectID | Первичный ключ MongoDB |
| call_id | string | ID звонка (UUID в строковом формате) |
| user_id | string | ID отправителя |
| user_name | string | Имя отправителя |
| text | string | Текст сообщения (макс. 2000 символов) |
| created_at | DateTime | Время отправки |

**Составной индекс:** `{call_id: 1, created_at: 1}` -- для эффективной выборки сообщений звонка в хронологическом порядке.

**Эфемерность:** при получении webhook `room_finished` вызывается `ChatRepository.DeleteCallChat(callID)`, который выполняет `DeleteMany({call_id: callID})`.

## Структура модуля: Frontend

### Компоненты `frontend/src/components/calls/`

| Компонент | Назначение |
|-----------|-----------|
| `CallRoom.tsx` | Главный контейнер звонка. Оборачивает `LiveKitRoom` с audio=true, video=false. Содержит RoomAudioRenderer, ParticipantList, ScreenShareView, CallControls. |
| `CallControls.tsx` | Кастомная панель управления (вместо встроенного ControlBar). Кнопки: микрофон, демонстрация экрана, поднятие руки, покинуть звонок, завершить звонок. |
| `ParticipantList.tsx` | Список участников звонка с индикаторами: mute, поднятая рука, демонстрация экрана. Парсит metadata участника для hand_raised. |
| `ScreenShareView.tsx` | Отображение демонстраций экрана в адаптивной сетке. Поддерживает несколько одновременных screen share через массив VideoTrack. |
| `CallButton.tsx` | Кнопка в заголовке канала для начала/присоединения к звонку. Проверяет наличие активного звонка через getActiveCall. |
| `CallPanel.tsx` | Фиксированная оверлейная панель (420x500px) в правом нижнем углу экрана для отображения активного звонка. |

### API и стор

| Файл | Назначение |
|------|-----------|
| `frontend/src/api/calls.ts` | API-клиент: createCall, joinCall, leaveCall, endCall, getActiveCall, raiseHand, sendChatMessage, listChatMessages |
| `frontend/src/stores/call-store.ts` | Zustand store: activeCall, token, isInCall, handRaised + методы startCall, joinCall, leaveCall, endCall, raiseHand |

## API Reference

### POST /api/calls

Создание нового звонка в канале.

**Auth:** JWT (auth middleware)
**Body:** `{ "channel_id": "uuid" }`
**Response (201):**
```json
{
  "id": "uuid",
  "room_name": "call-uuid",
  "channel_id": "uuid",
  "status": "active",
  "created_by": "uuid",
  "participant_count": 1,
  "token": "livekit-jwt-token"
}
```

### POST /api/calls/:callId/join

Присоединение к активному звонку.

**Auth:** JWT
**Response (200):** CallResponse с token

### POST /api/calls/:callId/leave

Выход из звонка.

**Auth:** JWT
**Response (200):** пустой

### POST /api/calls/:callId/end

Завершение звонка (удаление комнаты LiveKit).

**Auth:** JWT
**Response (200):** пустой

### POST /api/calls/:callId/raise-hand

Поднять/опустить руку.

**Auth:** JWT
**Body:** `{ "raised": true }`
**Response (200):** `{ "raised": true }`

### GET /api/calls/channel/:channelId

Получение активного звонка канала.

**Auth:** JWT
**Response (200):** CallResponse (без token) или `{ "call": null }`

### POST /api/calls/:callId/chat

Отправка сообщения в эфемерный чат звонка.

**Auth:** JWT
**Body:** `{ "text": "сообщение" }`
**Response (201):**
```json
{
  "id": "mongo-objectid",
  "call_id": "uuid",
  "user_id": "uuid",
  "user_name": "Имя",
  "text": "сообщение",
  "created_at": "2026-03-21T12:00:00Z"
}
```

### GET /api/calls/:callId/chat

Получение сообщений чата звонка.

**Auth:** JWT
**Query:** `limit` (default: 50, max: 200)
**Response (200):** `{ "messages": [...] }`

### POST /api/calls/webhook

Приём webhook-событий от LiveKit.

**Auth:** Без auth middleware (верификация через JWT в теле запроса)
**Body:** JSON с полями event, room, participant
**События:** room_started, room_finished, participant_joined, participant_left

## NATS Subjects

| Subject | Паттерн | Описание | WS Room |
|---------|---------|----------|---------|
| `calls.channel.{channelID}.started` | fire-and-forget | Звонок начат в канале | `channel:{channelID}` |
| `calls.channel.{channelID}.joined` | fire-and-forget | Участник присоединился | `channel:{channelID}` |
| `calls.channel.{channelID}.left` | fire-and-forget | Участник покинул звонок | `channel:{channelID}` |
| `calls.channel.{channelID}.ended` | fire-and-forget | Звонок завершён | `channel:{channelID}` |
| `calls.chat.{callID}` | fire-and-forget | Сообщение чата звонка | `call:{callID}` |

**WS Bridge маршрутизация:**
- `calls.channel.>` -- события маршрутизируются в WS room `channel:{channelID}` с префиксом `call.` (call.started, call.ended, call.joined, call.left)
- `calls.chat.>` -- сообщения маршрутизируются в WS room `call:{callID}` с типом `chat.message`

## Конфигурация

### Переменные окружения (Backend)

| Переменная | Описание | Пример |
|-----------|----------|--------|
| `LIVEKIT_URL` | URL сервера LiveKit | `http://localhost:7880` |
| `LIVEKIT_API_KEY` | API ключ LiveKit | `devkey` |
| `LIVEKIT_API_SECRET` | API секрет LiveKit | `devsecret` |

### Переменные окружения (Frontend)

| Переменная | Описание | Пример |
|-----------|----------|--------|
| `VITE_LIVEKIT_URL` | WebSocket URL LiveKit | `ws://localhost:7880` |

### Структура конфигурации (Go)

```go
type LiveKitConfig struct {
    URL       string `env:"LIVEKIT_URL" envDefault:"http://localhost:7880"`
    APIKey    string `env:"LIVEKIT_API_KEY" envDefault:"devkey"`
    APISecret string `env:"LIVEKIT_API_SECRET" envDefault:"devsecret"`
}
```

## Docker Compose

### LiveKit сервис

```yaml
livekit:
  image: livekit/livekit-server:latest
  ports:
    - "7880:7880"       # HTTP + WebSocket (сигнализация)
    - "7881:7881"       # Управление (admin API)
    - "50000-50100:50000-50100/udp"  # WebRTC медиа (UDP)
  volumes:
    - ./livekit.yaml:/etc/livekit.yaml
  command: --config /etc/livekit.yaml
```

### Файл `deployments/livekit.yaml`

```yaml
port: 7880
rtc:
  port_range_start: 50000
  port_range_end: 50100
  use_external_ip: false
keys:
  devkey: devsecret
webhook:
  urls:
    - "http://backend:3000/api/calls/webhook"
room:
  empty_timeout: 60      # Удалить комнату через 60 сек после ухода последнего участника
  max_participants: 50
logging:
  level: info
```

### Caddyfile (дополнение)

```
/livekit/* {
    reverse_proxy localhost:7880
}
```

## Безопасность

### JWT-токены LiveKit

- Бэкенд генерирует JWT-токены для подключения клиентов к LiveKit
- Токены подписаны shared secret (`LIVEKIT_API_SECRET`) алгоритмом HS256
- Каждый токен содержит `VideoGrant` с правом `RoomJoin` для конкретной комнаты
- Время жизни токена: 24 часа
- Identity = user ID (UUID), Name = имя пользователя

### Верификация Webhook

LiveKit подписывает webhook-запросы JWT-токеном:

1. **JWT подпись** -- Authorization header содержит JWT, подписанный API Secret (HMAC/HS256)
2. **SHA256 hash** -- claims содержат `sha256` = base64(SHA256(body)), подтверждающий целостность тела запроса
3. **API Key проверка** -- issuer (iss) в JWT должен совпадать с `LIVEKIT_API_KEY`

**Примечание:** Используется ручная верификация вместо `webhook.ReceiveWebhookEvent` из LiveKit SDK, так как `ReceiveWebhookEvent` ожидает `*http.Request`, несовместимый с Fiber/fasthttp RequestCtx.

```go
func (h *Handler) verifyWebhookSignature(authHeader string, body []byte) bool {
    // 1. Парсим JWT с проверкой HMAC подписи
    token, err := jwt.Parse(authHeader, func(t *jwt.Token) (interface{}, error) {
        if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("неожиданный метод подписи")
        }
        return []byte(h.cfg.APISecret), nil
    })
    // 2. Проверяем issuer == API Key
    // 3. Проверяем SHA256(body) == claims["sha256"]
}
```

### Fiber/fasthttp безопасность

- `utils.CopyString()` используется для всех строк из RequestCtx (параметры, заголовки) из-за переиспользования буферов fasthttp
- Тело webhook копируется в отдельный slice перед обработкой
- Webhook-эндпоинт регистрируется **без** auth middleware (LiveKit вызывает напрямую)

## Известные ограничения и решения

### 1. Fiber/fasthttp RequestCtx reuse

**Проблема:** fasthttp переиспользует RequestCtx после обработки запроса. Строки из параметров и заголовков становятся невалидными.

**Решение:** Все строки из RequestCtx копируются через `utils.CopyString()` перед использованием. Тело webhook копируется в отдельный `[]byte`.

### 2. empty_timeout для cleanup

**Проблема:** Если все участники покидают звонок без явного завершения, комната должна быть удалена.

**Решение:** LiveKit конфигурирован с `empty_timeout: 60` -- комната автоматически удаляется через 60 секунд после ухода последнего участника. Событие `room_finished` запускает cleanup (удаление чата, обновление статуса).

### 3. Nil-safe NATS

**Проблема:** В тестах и при отсутствии NATS-подключения публикация событий не должна вызывать ошибок.

**Решение:** `publishEvent()` проверяет `s.nats == nil` и пропускает публикацию с логированием на уровне Debug. Паттерн повторно использован из модуля мессенджера.

### 4. Webhook несовместимость с LiveKit SDK

**Проблема:** `webhook.ReceiveWebhookEvent()` из LiveKit SDK ожидает `*http.Request` (net/http), но Fiber использует fasthttp с несовместимым RequestCtx.

**Решение:** Реализована ручная верификация JWT + SHA256:
- JWT парсится через `golang-jwt/jwt/v5` с проверкой HMAC
- SHA256 hash тела вычисляется и сравнивается с claims
- API Key проверяется через issuer claim

### 5. Метаданные участника для поднятия руки

**Проблема:** LiveKit не имеет встроенной функции "поднять руку". Нужен механизм передачи этого состояния всем участникам.

**Решение:** Используются метаданные участника LiveKit (participant metadata). При поднятии руки бэкенд вызывает `UpdateParticipant` с JSON-метаданными `{hand_raised: true, raised_at: "..."}`. Фронтенд парсит metadata каждого участника для отображения индикатора.

## Зависимости

### Backend (Go modules)

| Пакет | Версия | Назначение |
|-------|--------|-----------|
| `github.com/livekit/server-sdk-go/v2` | latest | RoomServiceClient для управления комнатами |
| `github.com/livekit/protocol` | latest | Типы (CreateRoomRequest, VideoGrant, AccessToken) |
| `github.com/golang-jwt/jwt/v5` | v5 | Верификация JWT webhook от LiveKit |
| `go.mongodb.org/mongo-driver/v2` | v2 | MongoDB клиент для эфемерного чата |

### Frontend (npm)

| Пакет | Назначение |
|-------|-----------|
| `@livekit/components-react` | React-компоненты LiveKit (LiveKitRoom, RoomAudioRenderer) |
| `livekit-client` | LiveKit JS SDK (Room, Track, Participant API) |
| `@livekit/components-styles` | CSS-стили для LiveKit компонентов |

---
*Фаза: 05-voice-calls*
*Последнее обновление: 2026-03-21*
