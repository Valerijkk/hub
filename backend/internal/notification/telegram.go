package notification

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	telegramutil "github.com/valerii/hub/backend/internal/platform/telegram"
)

// telegramMessage — внутреннее представление сообщения для очереди.
type telegramMessage struct {
	chatID int64
	title  string
	body   string
}

// TelegramSender отправляет уведомления через Telegram-бот.
// Использует буферизованный канал и воркер с rate limiting (35мс между отправками).
type TelegramSender struct {
	bot       *telego.Bot
	queue     chan telegramMessage
	cancel    context.CancelFunc
	workerCtx context.Context
}

// NewTelegramSender создаёт нового отправителя Telegram уведомлений.
// Если botToken пустой — возвращает nil (nil-safe: Send на nil sender — no-op).
func NewTelegramSender(botToken string) *TelegramSender {
	if botToken == "" {
		return nil
	}

	bot, err := telego.NewBot(botToken)
	if err != nil {
		slog.Error("ошибка создания Telegram бота", "error", err)
		return nil
	}

	return &TelegramSender{
		bot:   bot,
		queue: make(chan telegramMessage, 100),
	}
}

// Start запускает воркер отправки сообщений.
func (t *TelegramSender) Start(ctx context.Context) {
	if t == nil {
		return
	}

	workerCtx, cancel := context.WithCancel(ctx)
	t.cancel = cancel
	t.workerCtx = workerCtx

	go func() {
		for {
			select {
			case <-workerCtx.Done():
				return
			case msg := <-t.queue:
				t.sendMessage(workerCtx, msg)
				// Rate limiting: 35мс между отправками (~28 сообщений/сек).
				// MEDIUM: ctx-aware — при shutdown воркер прерывается мгновенно.
				select {
				case <-workerCtx.Done():
					return
				case <-time.After(35 * time.Millisecond):
				}
			}
		}
	}()

	slog.Info("telegram sender запущен")
}

// Stop останавливает воркер отправки.
func (t *TelegramSender) Stop() {
	if t == nil || t.cancel == nil {
		return
	}
	t.cancel()
}

// Send добавляет сообщение в очередь отправки.
// Nil-safe: вызов на nil sender — no-op.
func (t *TelegramSender) Send(chatID int64, title, body string) {
	if t == nil {
		return
	}

	select {
	case t.queue <- telegramMessage{chatID: chatID, title: title, body: body}:
	default:
		slog.Warn("очередь telegram переполнена, сообщение отброшено")
	}
}

// sendMessage выполняет отправку одного сообщения через Telegram Bot API.
//
// ME-1: telego v0.31.0 SendMessage не принимает context.Context, поэтому
// оборачиваем вызов в горутину с каналом и select на ctx.Done(). Это
// гарантирует, что при shutdown воркер не виснет на сетевом таймауте API.
func (t *TelegramSender) sendMessage(ctx context.Context, msg telegramMessage) {
	text := formatTelegramMessage(msg.title, msg.body)

	params := tu.Message(tu.ID(msg.chatID), text)
	params.WithParseMode(telego.ModeMarkdownV2)

	errCh := make(chan error, 1)
	go func() {
		_, err := t.bot.SendMessage(params)
		errCh <- err
	}()

	select {
	case <-ctx.Done():
		slog.Warn("отправка telegram сообщения прервана shutdown", "chat_id", msg.chatID)
		return
	case err := <-errCh:
		if err != nil {
			slog.Error("ошибка отправки telegram сообщения", "error", err, "chat_id", msg.chatID)
		}
	}
}

// formatTelegramMessage форматирует уведомление для Telegram MarkdownV2.
// LO-01: использует общий хелпер из internal/platform/telegram (был дубликат).
func formatTelegramMessage(title, body string) string {
	escapedTitle := telegramutil.EscapeMarkdownV2(title)
	if body == "" {
		return fmt.Sprintf("*%s*", escapedTitle)
	}
	escapedBody := telegramutil.EscapeMarkdownV2(body)
	return fmt.Sprintf("*%s*\n\n%s", escapedTitle, escapedBody)
}
