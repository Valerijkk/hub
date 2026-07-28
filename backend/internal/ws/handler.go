// Package ws реализует WebSocket-шлюз для real-time событий через NATS bridge.
package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
	"github.com/valerii/hub/backend/internal/auth"
)

// Константы таймаутов и лимитов WebSocket-соединения.
const (
	// writeWait — максимальное время ожидания записи сообщения.
	writeWait = 10 * time.Second
	// pongWait — максимальное время ожидания pong-ответа от клиента.
	pongWait = 60 * time.Second
	// pingPeriod — интервал отправки ping-сообщений (90% от pongWait).
	pingPeriod = (pongWait * 9) / 10
	// maxMessageSize — максимальный размер входящего WebSocket-сообщения в байтах.
	maxMessageSize = 4096
)

// SEC-H1: CheckOrigin защищает от cross-origin WS-soup (CSWSH).
// Источники задаются той же переменной окружения CORS_ORIGINS, что и
// обычные HTTP CORS-настройки — это исключает рассинхрон между REST и WS.
// В DEV_MODE=true допускается allow-all, но с warn-логом.
//
// Парсинг выполняется один раз через sync.Once для предотвращения накладных
// расходов на каждый upgrade. В тестах можно сбросить через resetOriginCheck.
var (
	originOnce      sync.Once
	allowedOrigins  map[string]struct{}
	originDevMode   bool
	originDevWarned bool
)

// resetOriginCheck используется в тестах для повторной инициализации с другим
// окружением. В production вызывается ровно один раз через sync.Once.
func resetOriginCheck() {
	originOnce = sync.Once{}
	allowedOrigins = nil
	originDevMode = false
	originDevWarned = false
}

func initOriginCheck() {
	originOnce.Do(func() {
		env := strings.TrimSpace(os.Getenv("CORS_ORIGINS"))
		set := make(map[string]struct{})
		for _, raw := range strings.Split(env, ",") {
			o := strings.TrimSpace(raw)
			if o == "" || o == "*" {
				continue
			}
			set[o] = struct{}{}
		}
		allowedOrigins = set

		dev := strings.TrimSpace(os.Getenv("DEV_MODE"))
		switch strings.ToLower(dev) {
		case "true", "1", "yes", "on":
			originDevMode = true
		}

		if len(allowedOrigins) == 0 && !originDevMode {
			slog.Error("ws CheckOrigin: CORS_ORIGINS не задан в production — все WS-апгрейды будут отвергнуты",
				"hint", "set CORS_ORIGINS=https://yourdomain.example")
		}
	})
}

// checkOrigin возвращает true, если Origin запроса разрешён.
// Пустой Origin (например, native-клиенты или curl без браузера) разрешается:
// CSWSH применим только к браузерам, которые ВСЕГДА посылают Origin.
func checkOrigin(ctx *fasthttp.RequestCtx) bool {
	initOriginCheck()

	origin := strings.TrimSpace(string(ctx.Request.Header.Peek("Origin")))
	if origin == "" {
		// Небраузерные клиенты (CLI, тесты) — CSWSH не применим без Origin.
		return true
	}

	if originDevMode {
		if !originDevWarned {
			slog.Warn("ws CheckOrigin: DEV_MODE=true — все Origin разрешены", "origin", origin)
			originDevWarned = true
		}
		return true
	}

	_, ok := allowedOrigins[origin]
	if !ok {
		slog.Warn("ws CheckOrigin: отклонён запрещённый Origin", "origin", origin)
	}
	return ok
}

// upgrader — апгрейдер FastHTTP для преобразования HTTP-соединения в WebSocket.
var upgrader = websocket.FastHTTPUpgrader{
	CheckOrigin: checkOrigin,
}

// Authorizer проверяет, имеет ли пользователь право подписаться на
// конкретную WS-комнату. SEC-H1: без этой проверки Hub.Subscribe был
// broadcast-IDOR — любой авторизованный пользователь мог слушать события
// чужого канала/звонка/проекта.
//
// Реализация должна понимать формат room (см. ws/bridge.go):
//   - "channel:{channelID}" — членство в канале мессенджера
//   - "call:{callID}"       — участник звонка (или член канала звонка)
//   - "project:{projectID}" — членство в проекте
//
// Контекст передаётся для возможности отмены и трассировки. userID —
// строковое представление UUID (как в JWT claims).
type Authorizer interface {
	CanSubscribe(ctx context.Context, userID, room string) (bool, error)
}

// allowAllAuthorizer разрешает любую подписку. Используется только в тестах,
// где Hub конструируется без реальной БД. В production Handler ОБЯЗАН
// получить настоящий Authorizer.
type allowAllAuthorizer struct{}

func (allowAllAuthorizer) CanSubscribe(_ context.Context, _, _ string) (bool, error) {
	return true, nil
}

// Handler возвращает Fiber-обработчик для апгрейда HTTP-соединения в WebSocket.
// JWT-токен передаётся через query-параметр ?token=, так как WebSocket
// не поддерживает кастомные заголовки при рукопожатии.
//
// ВАЖНО (Pitfall 1): Все значения извлекаются из fiber.Ctx ДО апгрейда
// через utils.CopyString для предотвращения проблем переиспользования fasthttp RequestCtx.
//
// SEC-H1 / W3-A4-003: authz обязателен. nil → отказ upgrade (fail-closed).
// Для тестов без БД передавайте ExportAllowAllAuthorizer() явно (hubtest).
func Handler(hub *Hub, authSvc *auth.AuthService, authz Authorizer) fiber.Handler {
	return func(c fiber.Ctx) error {
		if authz == nil {
			slog.Error("ws Handler: Authorizer is nil — refusing upgrade (fail-closed)")
			return fiber.NewError(fiber.StatusServiceUnavailable, "websocket authorizer not configured")
		}
		// JWT из query до upgrade — после hijack fiber.Ctx недоступен (Pitfall 1).
		tokenStr := utils.CopyString(c.Query("token"))
		if tokenStr == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing token query parameter")
		}

		claims, err := authSvc.ParseAccessToken(tokenStr)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid or expired token")
		}

		// Копируем userID/sessionID до upgrade (Pitfall 1).
		userID := utils.CopyString(claims.UserID)
		sessionID := utils.CopyString(claims.SessionID)

		// ME-13: та же проверка deactivated, что и RequireAuth — до WS upgrade.
		uid, parseErr := uuid.Parse(userID)
		if parseErr != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid token claims")
		}
		deactivated, statusErr := authSvc.IsUserDeactivated(c.Context(), uid)
		if statusErr != nil {
			return fiber.NewError(fiber.StatusServiceUnavailable, "unable to verify account status")
		}
		if deactivated {
			return fiber.NewError(fiber.StatusForbidden, "user account is deactivated")
		}

		return upgrader.Upgrade(c.Context(), func(conn *websocket.Conn) {
			client := NewClient(userID, sessionID)
			hub.Register(client)

			defer func() {
				hub.Unregister(client)
				conn.Close()
			}()

			// writePump в отдельной goroutine — readPump блокирует conn на чтении.
			go writePump(conn, client)

			// readPump на текущей goroutine — conn закрывается при выходе из readPump.
			readPump(conn, client, hub, authz)
		})
	}
}

// readPump читает сообщения из WebSocket-соединения и диспетчеризирует их.
//
// SEC-H1: TypeSubscribe проходит через Authorizer.CanSubscribe — без этой
// проверки любой авторизованный пользователь мог подписаться на чужую
// комнату (channel:/call:/project:) и получать все события (broadcast IDOR).
func readPump(conn *websocket.Conn, client *Client, hub *Hub, authz Authorizer) {
	conn.SetReadLimit(maxMessageSize)
	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				slog.Warn("ws read error", "userID", client.UserID, "error", err)
			}
			return
		}

		var msg ClientMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			slog.Warn("ws invalid message", "userID", client.UserID, "error", err)
			continue
		}

		switch msg.Type {
		case TypeSubscribe:
			if msg.Room == "" {
				continue
			}
			// SEC-H1: проверяем право подписки до фактического Subscribe.
			authCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			allowed, authErr := authz.CanSubscribe(authCtx, client.UserID, msg.Room)
			cancel()
			if authErr != nil {
				slog.Warn("ws subscribe authz error",
					"userID", client.UserID, "room", msg.Room, "error", authErr)
				continue
			}
			if !allowed {
				slog.Warn("ws subscribe denied",
					"userID", client.UserID, "room", msg.Room)
				continue
			}
			hub.Subscribe(client, msg.Room)
		case TypeUnsubscribe:
			if msg.Room != "" {
				hub.Unsubscribe(client, msg.Room)
			}
		case TypePing:
			pong := ServerMessage{Type: TypePong}
			data, _ := json.Marshal(pong)
			// SEC-05: client.send никогда не закрывается, но проверяем done,
			// чтобы не писать в send после Unregister.
			select {
			case <-client.done:
			case client.send <- data:
			default:
			}
		}
	}
}

// writePump записывает сообщения из канала отправки клиента в WebSocket-соединение.
// Завершается при закрытии client.done (до закрытия conn), предотвращая
// nil pointer dereference при записи в закрытый fasthttp hijackConn.
//
// SEC-05: канал client.send больше НЕ закрывается со стороны Hub — сигналом
// завершения служит исключительно client.done. Это устраняет панику
// "send on closed channel" в параллельных producer'ах (Bridge/handler).
func writePump(conn *websocket.Conn, client *Client) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-client.done:
			// Соединение закрывается — выходим до обращения к conn.
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			_ = conn.WriteMessage(websocket.CloseMessage, nil)
			return
		case message := <-client.send:
			if err := conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				return
			}
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
