package notification

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
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
	bot    *telego.Bot
	queue  chan telegramMessage
	cancel context.CancelFunc
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

	go func() {
		for {
			select {
			case <-workerCtx.Done():
				return
			case msg := <-t.queue:
				t.sendMessage(msg)
				// Rate limiting: 35мс между отправками (~28 сообщений/сек).
				time.Sleep(35 * time.Millisecond)
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
func (t *TelegramSender) sendMessage(msg telegramMessage) {
	text := formatTelegramMessage(msg.title, msg.body)

	params := tu.Message(tu.ID(msg.chatID), text)
	params.WithParseMode(telego.ModeMarkdownV2)

	_, err := t.bot.SendMessage(params)
	if err != nil {
		slog.Error("ошибка отправки telegram сообщения", "error", err, "chat_id", msg.chatID)
	}
}

// formatTelegramMessage форматирует уведомление для Telegram MarkdownV2.
func formatTelegramMessage(title, body string) string {
	escapedTitle := escapeMarkdownV2(title)
	if body == "" {
		return fmt.Sprintf("*%s*", escapedTitle)
	}
	escapedBody := escapeMarkdownV2(body)
	return fmt.Sprintf("*%s*\n\n%s", escapedTitle, escapedBody)
}

// escapeMarkdownV2 экранирует специальные символы для Telegram MarkdownV2.
func escapeMarkdownV2(s string) string {
	special := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
	result := s
	for _, ch := range special {
		result = strings.ReplaceAll(result, ch, "\\"+ch)
	}
	return result
}
