package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"github.com/redis/go-redis/v9"
	telegramutil "github.com/valerii/hub/backend/internal/platform/telegram"
)

// Настройки Telegram бот-авторизации.
const (
	tgCodeTTL         = 5 * time.Minute
	tgCodePrefix      = "tglogin:"
	tgVerifyFailsMax  = 5 // max failed VerifyCode attempts per IP within TTL window
	tgFailsPrefix     = "tglogin:fails:"
)

// TelegramBotAuth обеспечивает вход через Telegram-бот.
// Бот слушает /start и /login, генерирует 6-значный код,
// сохраняет в Redis привязку код → telegram_id, шлёт код юзеру.
type TelegramBotAuth struct {
	bot   *telego.Bot
	redis *redis.Client
}

// NewTelegramBotAuth создаёт новый сервис Telegram-бот-авторизации.
// Возвращает nil если botToken пустой.
func NewTelegramBotAuth(botToken string, rdb *redis.Client) *TelegramBotAuth {
	if botToken == "" {
		return nil
	}
	bot, err := telego.NewBot(botToken)
	if err != nil {
		slog.Error("telegram bot auth: не удалось создать бота", "error", err)
		return nil
	}
	return &TelegramBotAuth{bot: bot, redis: rdb}
}

// StartPolling запускает long polling для обработки команд /start и /login.
func (t *TelegramBotAuth) StartPolling(ctx context.Context) {
	if t == nil {
		return
	}

	// Удаляем webhook если был — иначе getUpdates не работает.
	_ = t.bot.DeleteWebhook(&telego.DeleteWebhookParams{})

	go func() {
		offset := 0
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			updates, err := t.bot.GetUpdates(&telego.GetUpdatesParams{
				Offset:  offset,
				Timeout: 30,
			})
			if err != nil {
				slog.Error("telegram bot auth: ошибка получения обновлений", "error", err)
				time.Sleep(3 * time.Second)
				continue
			}

			for _, update := range updates {
				offset = update.UpdateID + 1
				if update.Message == nil {
					continue
				}
				text := strings.TrimSpace(update.Message.Text)
				if text == "/start" || text == "/login" || strings.HasPrefix(text, "/start ") {
					t.handleLogin(ctx, update.Message)
				}
			}
		}
	}()

	slog.Info("telegram bot auth polling запущен")
}

// handleLogin генерирует код и отправляет его пользователю.
func (t *TelegramBotAuth) handleLogin(ctx context.Context, msg *telego.Message) {
	chatID := msg.Chat.ID
	telegramID := msg.From.ID

	code, err := generateTGCode()
	if err != nil {
		slog.Error("telegram bot auth: ошибка генерации кода", "error", err)
		return
	}

	// Сохраняем код → telegram_id + username/name
	data := fmt.Sprintf("%d|%s|%s|%s",
		telegramID,
		msg.From.FirstName,
		msg.From.LastName,
		msg.From.Username,
	)
	key := tgCodePrefix + code
	if err := t.redis.Set(ctx, key, data, tgCodeTTL).Err(); err != nil {
		slog.Error("telegram bot auth: ошибка сохранения кода", "error", err)
		return
	}

	text := fmt.Sprintf("Ваш код входа в Hub:\n\n*%s*\n\nВведите его на сайте\\. Код действителен 5 минут\\.", telegramutil.EscapeMarkdownV2(code))
	params := tu.Message(tu.ID(chatID), text)
	params.WithParseMode(telego.ModeMarkdownV2)

	if _, err := t.bot.SendMessage(params); err != nil {
		slog.Error("telegram bot auth: ошибка отправки кода", "error", err, "chat_id", chatID)
	}
}

// ErrTelegramVerifyLimited indicates the caller (identified by IP) has exceeded
// the per-IP failed-attempt limit for Telegram code verification.
var ErrTelegramVerifyLimited = fmt.Errorf("too many telegram verification attempts, try again later")

// VerifyCode проверяет код и возвращает данные Telegram-пользователя.
// clientKey — стабильный идентификатор клиента (обычно IP); используется только
// для per-client учёта неудачных попыток (HI-04). Пустая строка отключает
// проверку счётчика (для тестов).
func (t *TelegramBotAuth) VerifyCode(ctx context.Context, code, clientKey string) (*TelegramUser, error) {
	if t == nil {
		return nil, fmt.Errorf("telegram bot auth не настроен")
	}

	failsKey := ""
	if clientKey != "" {
		failsKey = tgFailsPrefix + clientKey
		if cnt, err := t.redis.Get(ctx, failsKey).Int(); err == nil && cnt >= tgVerifyFailsMax {
			return nil, ErrTelegramVerifyLimited
		} else if err != nil && !errors.Is(err, redis.Nil) {
			slog.Warn("telegram verify fails counter read error", "client", clientKey, "error", err)
		}
	}

	key := tgCodePrefix + code
	data, err := t.redis.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		t.recordVerifyFailure(ctx, failsKey)
		return nil, fmt.Errorf("неверный или истёкший код")
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки кода: %w", err)
	}

	// Парсим: telegramID|firstName|lastName|username
	parts := strings.SplitN(data, "|", 4)
	if len(parts) < 4 {
		t.recordVerifyFailure(ctx, failsKey)
		return nil, fmt.Errorf("повреждённые данные кода")
	}

	tgID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		t.recordVerifyFailure(ctx, failsKey)
		return nil, fmt.Errorf("некорректный telegram ID: %w", err)
	}

	// Успех: удаляем одноразовый код и счётчик неудач.
	t.redis.Del(ctx, key)
	if failsKey != "" {
		t.redis.Del(ctx, failsKey)
	}

	return &TelegramUser{
		ID:        tgID,
		FirstName: parts[1],
		LastName:  parts[2],
		Username:  parts[3],
	}, nil
}

// recordVerifyFailure increments the per-client failed-attempt counter and sets
// the TTL on first failure to roughly the code TTL.
func (t *TelegramBotAuth) recordVerifyFailure(ctx context.Context, failsKey string) {
	if failsKey == "" {
		return
	}
	cnt, err := t.redis.Incr(ctx, failsKey).Result()
	if err != nil {
		slog.Warn("telegram verify fails counter incr error", "key", failsKey, "error", err)
		return
	}
	if cnt == 1 {
		if expErr := t.redis.Expire(ctx, failsKey, tgCodeTTL).Err(); expErr != nil {
			slog.Warn("telegram verify fails counter expire error", "key", failsKey, "error", expErr)
		}
	}
}

func generateTGCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
