package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// Настройки Magic Link авторизации.
const (
	magicLinkTTL         = 5 * time.Minute
	magicLinkRateMax     = 3
	magicLinkRateTTL     = 10 * time.Minute
	magicLinkVerifyMax   = 5 // лимит неудачных VerifyCode на email в окне TTL (HI-04)
	magicLinkFailsPrefix = "magiclink:fails:"
)

// ResendClient is the email delivery interface. A nil client disables delivery;
// public code disclosure is controlled only by the Handler's explicit dev mode.
type ResendClient interface {
	SendMagicLinkEmail(ctx context.Context, to, code string) error
}

// MagicLinkService генерирует, хранит в Redis и верифицирует одноразовые коды magic link.
type MagicLinkService struct {
	redis     *redis.Client
	resend    ResendClient
	fromEmail string
}

// NewMagicLinkService создаёт MagicLinkService.
func NewMagicLinkService(rdb *redis.Client, resend ResendClient, fromEmail string) *MagicLinkService {
	return &MagicLinkService{
		redis:     rdb,
		resend:    resend,
		fromEmail: fromEmail,
	}
}

// generateCode возвращает криптостойкую 6-значную строку.
func generateCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", fmt.Errorf("generate magic link code: %w", err)
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// maskEmailMinimized — A14-006: не писать полный email в stdout/Loki.
func maskEmailMinimized(email string) string {
	at := strings.LastIndex(email, "@")
	if at <= 0 {
		return "***"
	}
	local, domain := email[:at], email[at:]
	if len(local) == 0 {
		return "***" + domain
	}
	return string(local[0]) + "***" + domain
}

// SendMagicLink генерирует код, сохраняет в Redis и отправляет email.
// Лимит: 3 запроса на 10 минут на email (TZ mandatory).
func (s *MagicLinkService) SendMagicLink(ctx context.Context, email string) (string, error) {
	masked := maskEmailMinimized(email)
	// ME-05: атомарное окно rate-limit через SET NX EX — без гонки Incr+Expire,
	// при которой счётчик мог остаться без TTL после сбоя Redis между командами.
	rateKey := fmt.Sprintf("magiclink_rate:%s", email)
	ok, err := s.redis.SetNX(ctx, rateKey, 1, magicLinkRateTTL).Result()
	if err != nil {
		return "", fmt.Errorf("magic link rate check (setnx): %w", err)
	}
	if !ok {
		count, err := s.redis.Incr(ctx, rateKey).Result()
		if err != nil {
			return "", fmt.Errorf("magic link rate check (incr): %w", err)
		}
		if count > magicLinkRateMax {
			return "", ErrMagicLinkRateLimited
		}
	}

	code, err := generateCode()
	if err != nil {
		return "", err
	}

	codeKey := fmt.Sprintf("magiclink:%s", email)
	if err := s.redis.Set(ctx, codeKey, code, magicLinkTTL).Err(); err != nil {
		return "", fmt.Errorf("store magic link code: %w", err)
	}

	if s.resend != nil {
		if err := s.resend.SendMagicLinkEmail(ctx, email, code); err != nil {
			slog.Error("не удалось отправить magic link email", "email", masked, "error", err)
			// A1-007: не маскировать сбой Resend успехом — клиент получит 500, не «sent».
			return "", fmt.Errorf("send magic link email: %w", err)
		}
	} else {
		// A14-006: OTP is never written to logs, including when delivery is disabled.
		slog.Info("magic link issued without delivery", "email", masked)
	}

	return code, nil
}

// VerifyCode сверяет код в Redis; при успехе ключ удаляется (одноразовость).
// После magicLinkVerifyMax неудач в окне TTL дальнейшие попытки отклоняются (HI-04).
func (s *MagicLinkService) VerifyCode(ctx context.Context, email, code string) error {
	masked := maskEmailMinimized(email)
	codeKey := fmt.Sprintf("magiclink:%s", email)
	failsKey := magicLinkFailsPrefix + email

	if cnt, err := s.redis.Get(ctx, failsKey).Int(); err == nil && cnt >= magicLinkVerifyMax {
		return ErrMagicLinkVerifyLimited
	} else if err != nil && !errors.Is(err, redis.Nil) {
		slog.Warn("magiclink fails counter read error", "email", masked, "error", err)
	}

	stored, err := s.redis.Get(ctx, codeKey).Result()
	if errors.Is(err, redis.Nil) {
		s.recordFailure(ctx, failsKey)
		return ErrMagicLinkInvalid
	}
	if err != nil {
		return fmt.Errorf("verify magic link code: %w", err)
	}
	// SEC-07: constant-time compare — защита от timing side-channel при подборе кода.
	if subtle.ConstantTimeCompare([]byte(stored), []byte(code)) != 1 {
		s.recordFailure(ctx, failsKey)
		return ErrMagicLinkInvalid
	}

	// A1-010: при сбое Del код остаётся → повторное использование; fail-closed.
	if err := s.redis.Del(ctx, codeKey).Err(); err != nil {
		return fmt.Errorf("invalidate magic link code: %w", err)
	}
	if err := s.redis.Del(ctx, failsKey).Err(); err != nil {
		slog.Warn("magiclink fails counter delete error", "email", masked, "error", err)
	}
	return nil
}

// recordFailure увеличивает счётчик неудач email; TTL ~ magicLinkTTL при первой ошибке.
func (s *MagicLinkService) recordFailure(ctx context.Context, failsKey string) {
	cnt, err := s.redis.Incr(ctx, failsKey).Result()
	if err != nil {
		slog.Warn("magiclink fails counter incr error", "key", failsKey, "error", err)
		return
	}
	if cnt == 1 {
		if expErr := s.redis.Expire(ctx, failsKey, magicLinkTTL).Err(); expErr != nil {
			slog.Warn("magiclink fails counter expire error", "key", failsKey, "error", expErr)
		}
	}
}

// sentinel-ошибки magic link.
var (
	ErrMagicLinkRateLimited   = fmt.Errorf("magic link rate limit exceeded")
	ErrMagicLinkInvalid       = fmt.Errorf("invalid or expired code")
	ErrMagicLinkVerifyLimited = fmt.Errorf("too many verification attempts, try again later")
)
