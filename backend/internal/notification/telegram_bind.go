package notification

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"github.com/redis/go-redis/v9"

	"github.com/valerii/hub/backend/ent"
)

const (
	tgBindPrefix          = "tg:bind:"
	tgLinkCodePrefix      = "tg:linkcode:"
	tgLinkCodeFailsPrefix = "tg:linkcode:fails:" // ASN-002: per-user fails, как magiclink:fails:
	tgBindTTL             = 10 * time.Minute
	tgLinkCodeTTL         = 5 * time.Minute
	tgLinkCodeVerifyMax   = 5 // лимит неудачных ConfirmLinkCode на user в окне TTL (как HI-04)
)

// TelegramBindService управляет привязкой Telegram для уведомлений (NOTIF-02).
type TelegramBindService struct {
	repo        *Repository
	rdb         *redis.Client
	sender      *TelegramSender
	bot         *telego.Bot
	botUsername string
}

// TelegramBindResponse — ответ POST /notifications/telegram/bind.
type TelegramBindResponse struct {
	BindToken  string `json:"bind_token"`
	DeepLink   string `json:"deep_link,omitempty"`
	ExpiresIn  int    `json:"expires_in"`
	BotEnabled bool   `json:"bot_enabled"`
}

// TelegramStatusResponse — GET /notifications/telegram/status.
type TelegramStatusResponse struct {
	Linked     bool       `json:"linked"`
	TelegramID string     `json:"telegram_id,omitempty"`
	Username   string     `json:"username,omitempty"`
}

// TelegramConfirmRequest — POST /notifications/telegram/confirm.
type TelegramConfirmRequest struct {
	Code string `json:"code"`
}

// NewTelegramBindService создаёт bind-сервис; bot nil при пустом token.
func NewTelegramBindService(
	repo *Repository,
	rdb *redis.Client,
	sender *TelegramSender,
	botToken, botUsername string,
) *TelegramBindService {
	if botToken == "" || rdb == nil {
		return nil
	}
	bot, err := telego.NewBot(botToken)
	if err != nil {
		slog.Error("telegram bind: bot init failed", "error", err)
		return nil
	}
	return &TelegramBindService{
		repo:        repo,
		rdb:         rdb,
		sender:      sender,
		bot:         bot,
		botUsername: strings.TrimPrefix(botUsername, "@"),
	}
}

// CreateBindToken сохраняет одноразовый bind-токен для deep link flow.
func (s *TelegramBindService) CreateBindToken(ctx context.Context, userID uuid.UUID) (*TelegramBindResponse, error) {
	if s == nil {
		return nil, ErrTelegramBindUnavailable
	}
	token, err := randomHex(16)
	if err != nil {
		return nil, err
	}
	key := tgBindPrefix + token
	if err := s.rdb.Set(ctx, key, userID.String(), tgBindTTL).Err(); err != nil {
		return nil, fmt.Errorf("store bind token: %w", err)
	}
	resp := &TelegramBindResponse{
		BindToken:  token,
		ExpiresIn:  int(tgBindTTL.Seconds()),
		BotEnabled: true,
	}
	if s.botUsername != "" {
		resp.DeepLink = fmt.Sprintf("https://t.me/%s?start=bind_%s", s.botUsername, token)
	}
	return resp, nil
}

// ConfirmLinkCode завершает bind по 6-значному коду из команды /link.
// ASN-002: после tgLinkCodeVerifyMax неудач в окне TTL — ErrTelegramConfirmLimited (как HI-04).
func (s *TelegramBindService) ConfirmLinkCode(ctx context.Context, userID uuid.UUID, code string) error {
	if s == nil {
		return ErrTelegramBindUnavailable
	}
	failsKey := tgLinkCodeFailsPrefix + userID.String()
	if cnt, err := s.rdb.Get(ctx, failsKey).Int(); err == nil && cnt >= tgLinkCodeVerifyMax {
		return ErrTelegramConfirmLimited
	} else if err != nil && !errors.Is(err, redis.Nil) {
		slog.Warn("telegram linkcode fails counter read error", "user_id", userID, "error", err)
	}

	code = strings.TrimSpace(code)
	if len(code) != 6 {
		s.recordLinkCodeFailure(ctx, failsKey)
		return ErrTelegramInvalidCode
	}
	key := tgLinkCodePrefix + code
	tgID, err := s.rdb.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		s.recordLinkCodeFailure(ctx, failsKey)
		return ErrTelegramInvalidCode
	}
	if err != nil {
		return fmt.Errorf("get link code: %w", err)
	}
	if err := s.linkTelegramID(ctx, userID, tgID); err != nil {
		return err
	}
	// W7-A1-010 / P3-BE-01: не глотать Del — токен может остаться в Redis.
	if err := s.rdb.Del(ctx, key).Err(); err != nil {
		slog.WarnContext(ctx, "telegram bind: del link code failed", "error", err)
	}
	if err := s.rdb.Del(ctx, failsKey).Err(); err != nil {
		slog.WarnContext(ctx, "telegram bind: del fails counter failed", "error", err)
	}
	return nil
}

// recordLinkCodeFailure увеличивает счётчик неудач user; TTL = tgLinkCodeTTL при первой ошибке.
// Mutation (ASN-002): убрать вызовы recordLinkCodeFailure / проверку failsKey → brute-force без 429 → RED.
func (s *TelegramBindService) recordLinkCodeFailure(ctx context.Context, failsKey string) {
	cnt, err := s.rdb.Incr(ctx, failsKey).Result()
	if err != nil {
		slog.Warn("telegram linkcode fails counter incr error", "key", failsKey, "error", err)
		return
	}
	if cnt == 1 {
		if expErr := s.rdb.Expire(ctx, failsKey, tgLinkCodeTTL).Err(); expErr != nil {
			slog.Warn("telegram linkcode fails counter expire error", "key", failsKey, "error", expErr)
		}
	}
}

// Unlink удаляет telegram_id из профиля пользователя.
func (s *TelegramBindService) Unlink(ctx context.Context, userID uuid.UUID) error {
	return s.repo.ClearUserTelegram(ctx, userID)
}

// GetStatus возвращает состояние bind для текущего пользователя.
func (s *TelegramBindService) GetStatus(ctx context.Context, userID uuid.UUID) (*TelegramStatusResponse, error) {
	u, err := s.repo.GetUserTelegram(ctx, userID)
	if err != nil {
		return nil, err
	}
	resp := &TelegramStatusResponse{Linked: u.TelegramID != nil && *u.TelegramID != ""}
	if u.TelegramID != nil {
		resp.TelegramID = *u.TelegramID
	}
	return resp, nil
}

// SendTest отправляет тестовое уведомление в привязанный Telegram-чат.
func (s *TelegramBindService) SendTest(ctx context.Context, userID uuid.UUID) error {
	if s == nil || s.sender == nil {
		return ErrTelegramBindUnavailable
	}
	u, err := s.repo.GetUserTelegram(ctx, userID)
	if err != nil {
		return err
	}
	if u.TelegramID == nil || *u.TelegramID == "" {
		return ErrTelegramNotLinked
	}
	chatID, err := strconv.ParseInt(*u.TelegramID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid telegram_id: %w", err)
	}
	s.sender.Send(chatID, "Hub notifications", "Тестовое уведомление — всё работает ✅")
	return nil
}

// StartPolling обрабатывает /start bind_*, /link, /unlink, /test у notify-бота.
func (s *TelegramBindService) StartPolling(ctx context.Context) {
	if s == nil || s.bot == nil {
		return
	}
	_ = s.bot.DeleteWebhook(&telego.DeleteWebhookParams{})

	go func() {
		offset := 0
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			updates, err := s.bot.GetUpdates(&telego.GetUpdatesParams{Offset: offset, Timeout: 30})
			if err != nil {
				slog.Warn("telegram bind bot: getUpdates error", "error", err)
				time.Sleep(3 * time.Second)
				continue
			}
			for _, update := range updates {
				offset = update.UpdateID + 1
				if update.Message == nil || update.Message.From == nil {
					continue
				}
				s.handleMessage(ctx, update.Message)
			}
		}
	}()
	slog.Info("telegram bind bot polling started")
}

func (s *TelegramBindService) handleMessage(ctx context.Context, msg *telego.Message) {
	text := strings.TrimSpace(msg.Text)
	chatID := msg.Chat.ID
	tgID := strconv.FormatInt(msg.From.ID, 10)

	switch {
	case strings.HasPrefix(text, "/start bind_"):
		token := strings.TrimPrefix(text, "/start bind_")
		s.handleStartBind(ctx, chatID, tgID, token)
	case text == "/link" || strings.HasPrefix(text, "/link@"):
		s.handleLinkCommand(ctx, chatID, tgID)
	case text == "/unlink" || strings.HasPrefix(text, "/unlink@"):
		s.handleUnlinkCommand(ctx, chatID, tgID)
	case text == "/test" || strings.HasPrefix(text, "/test@"):
		s.handleTestCommand(ctx, chatID, tgID)
	case text == "/start" || strings.HasPrefix(text, "/start@"):
		s.reply(chatID, "Hub Notify Bot\n\n/link — получить код привязки\n/unlink — отвязать\n/test — тест")
	}
}

func (s *TelegramBindService) handleStartBind(ctx context.Context, chatID int64, tgID, token string) {
	key := tgBindPrefix + token
	userIDStr, err := s.rdb.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		s.reply(chatID, "Ссылка привязки недействительна или истекла.")
		return
	}
	if err != nil {
		slog.Error("telegram bind: redis get", "error", err)
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return
	}
	if err := s.linkTelegramID(ctx, userID, tgID); err != nil {
		if errors.Is(err, ErrTelegramIDTaken) {
			s.reply(chatID, "Этот Telegram уже привязан к другому аккаунту Hub.")
			return
		}
		s.reply(chatID, "Не удалось привязать аккаунт. Попробуйте позже.")
		return
	}
	// W7-A1-010 / P3-BE-01
	if err := s.rdb.Del(ctx, key).Err(); err != nil {
		slog.WarnContext(ctx, "telegram bind: del pending token failed", "error", err)
	}
	s.reply(chatID, "Telegram успешно привязан к Hub ✅")
}

func (s *TelegramBindService) handleLinkCommand(ctx context.Context, chatID int64, tgID string) {
	code, err := generateLinkCode()
	if err != nil {
		return
	}
	key := tgLinkCodePrefix + code
	if err := s.rdb.Set(ctx, key, tgID, tgLinkCodeTTL).Err(); err != nil {
		return
	}
	s.replyPlain(chatID, fmt.Sprintf("Код привязки Hub: %s\n\nВведите его в Настройки → Уведомления → Telegram. Код действителен 5 минут.", code))
}

func (s *TelegramBindService) handleUnlinkCommand(ctx context.Context, chatID int64, tgID string) {
	u, err := s.repo.FindUserByTelegramID(ctx, tgID)
	if ent.IsNotFound(err) {
		s.reply(chatID, "Аккаунт не привязан.")
		return
	}
	if err != nil {
		return
	}
	if err := s.repo.ClearUserTelegram(ctx, u.ID); err != nil {
		s.reply(chatID, "Не удалось отвязать.")
		return
	}
	s.reply(chatID, "Telegram отвязан от Hub.")
}

func (s *TelegramBindService) handleTestCommand(ctx context.Context, chatID int64, tgID string) {
	u, err := s.repo.FindUserByTelegramID(ctx, tgID)
	if ent.IsNotFound(err) {
		s.reply(chatID, "Сначала привяжите аккаунт через Hub.")
		return
	}
	if err != nil {
		return
	}
	_ = u
	if s.sender != nil {
		s.sender.Send(chatID, "Hub notifications", "Тестовое уведомление ✅")
	}
}

func (s *TelegramBindService) linkTelegramID(ctx context.Context, userID uuid.UUID, tgID string) error {
	taken, err := s.repo.IsTelegramIDTaken(ctx, tgID, userID)
	if err != nil {
		return err
	}
	if taken {
		return ErrTelegramIDTaken
	}
	return s.repo.SetUserTelegram(ctx, userID, tgID)
}

func (s *TelegramBindService) replyPlain(chatID int64, text string) {
	params := tu.Message(tu.ID(chatID), text)
	if _, err := s.bot.SendMessage(params); err != nil {
		slog.Warn("telegram bind: send message failed", "error", err)
	}
}

func (s *TelegramBindService) reply(chatID int64, text string) {
	s.replyPlain(chatID, text)
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func generateLinkCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

var (
	ErrTelegramBindUnavailable = fmt.Errorf("telegram bind not configured")
	ErrTelegramInvalidCode     = fmt.Errorf("invalid or expired link code")
	ErrTelegramConfirmLimited  = fmt.Errorf("too many verification attempts, try again later")
	ErrTelegramNotLinked       = fmt.Errorf("telegram not linked")
	ErrTelegramIDTaken         = fmt.Errorf("telegram id already linked to another user")
)
