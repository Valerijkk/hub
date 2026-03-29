package auth

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/redis/go-redis/v9"
)

// Настройки Magic Link авторизации.
const (
	magicLinkTTL     = 5 * time.Minute
	magicLinkRateMax = 3
	magicLinkRateTTL = 10 * time.Minute
)

// ResendClient is an interface for sending emails. In production this calls the
// Resend API; in dev mode (empty API key) it logs the code instead.
type ResendClient interface {
	SendMagicLinkEmail(ctx context.Context, to, code string) error
}

// MagicLinkService handles generating, storing and verifying magic link codes.
type MagicLinkService struct {
	redis     *redis.Client
	resend    ResendClient
	fromEmail string
}

// NewMagicLinkService creates a new MagicLinkService.
func NewMagicLinkService(rdb *redis.Client, resend ResendClient, fromEmail string) *MagicLinkService {
	return &MagicLinkService{
		redis:     rdb,
		resend:    resend,
		fromEmail: fromEmail,
	}
}

// generateCode returns a cryptographically random 6-digit string.
func generateCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", fmt.Errorf("generate magic link code: %w", err)
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// SendMagicLink generates a code, stores it in Redis, and sends the email.
// Rate-limited to 3 requests per 10 minutes per email.
// IsDevMode returns true when Resend is not configured (dev mode — code logged, not emailed).
func (s *MagicLinkService) IsDevMode() bool {
	return s.resend == nil
}

func (s *MagicLinkService) SendMagicLink(ctx context.Context, email string) (string, error) {
	// Check rate limit
	rateKey := fmt.Sprintf("magiclink_rate:%s", email)
	count, err := s.redis.Incr(ctx, rateKey).Result()
	if err != nil {
		return "", fmt.Errorf("magic link rate check: %w", err)
	}
	if count == 1 {
		// First request -- set the TTL for the rate limit window.
		s.redis.Expire(ctx, rateKey, magicLinkRateTTL)
	}
	if count > magicLinkRateMax {
		return "", ErrMagicLinkRateLimited
	}

	code, err := generateCode()
	if err != nil {
		return "", err
	}

	codeKey := fmt.Sprintf("magiclink:%s", email)
	if err := s.redis.Set(ctx, codeKey, code, magicLinkTTL).Err(); err != nil {
		return "", fmt.Errorf("store magic link code: %w", err)
	}

	// Send email (or log in dev mode)
	if s.resend != nil {
		if err := s.resend.SendMagicLinkEmail(ctx, email, code); err != nil {
			slog.Error("не удалось отправить magic link email", "email", email, "error", err)
			return code, nil // Возвращаем код даже при ошибке отправки (можно посмотреть в логах)
		}
	} else {
		slog.Info("magic link code (dev mode)", "email", email, "code", code)
	}

	return code, nil
}

// VerifyCode checks the code stored in Redis. On match the key is deleted so
// each code can only be used once.
func (s *MagicLinkService) VerifyCode(ctx context.Context, email, code string) error {
	codeKey := fmt.Sprintf("magiclink:%s", email)
	stored, err := s.redis.Get(ctx, codeKey).Result()
	if err == redis.Nil {
		return ErrMagicLinkInvalid
	}
	if err != nil {
		return fmt.Errorf("verify magic link code: %w", err)
	}
	if stored != code {
		return ErrMagicLinkInvalid
	}

	// Delete so the code cannot be reused.
	s.redis.Del(ctx, codeKey)
	return nil
}

// sentinel errors
var (
	ErrMagicLinkRateLimited = fmt.Errorf("magic link rate limit exceeded")
	ErrMagicLinkInvalid     = fmt.Errorf("invalid or expired code")
)
