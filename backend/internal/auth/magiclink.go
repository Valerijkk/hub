package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/redis/go-redis/v9"
)

// Настройки Magic Link авторизации.
const (
	magicLinkTTL         = 5 * time.Minute
	magicLinkRateMax     = 3
	magicLinkRateTTL     = 10 * time.Minute
	magicLinkVerifyMax   = 5 // max failed VerifyCode attempts per email within TTL window
	magicLinkFailsPrefix = "magiclink:fails:"
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
	// ME-05: Atomic rate-limit window establishment.
	// Use SET NX EX to atomically initialize counter=1 with TTL on first hit.
	// If SetNX returns false (key exists), Incr the existing counter.
	// This avoids the race between Incr returning count=1 and a separate
	// Expire being applied — a concurrent Incr in that window would have
	// no TTL and the counter would persist forever after a Redis crash
	// between the two commands.
	rateKey := fmt.Sprintf("magiclink_rate:%s", email)
	ok, err := s.redis.SetNX(ctx, rateKey, 1, magicLinkRateTTL).Result()
	if err != nil {
		return "", fmt.Errorf("magic link rate check (setnx): %w", err)
	}
	if !ok {
		// Key already exists — increment without touching TTL.
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
// Per-email brute-force protection: after magicLinkVerifyMax failed attempts in
// the code's TTL window, further verifications are rejected outright (HI-04).
func (s *MagicLinkService) VerifyCode(ctx context.Context, email, code string) error {
	codeKey := fmt.Sprintf("magiclink:%s", email)
	failsKey := magicLinkFailsPrefix + email

	// Per-email failed-attempt limit (independent of IP-based rate-limit).
	if cnt, err := s.redis.Get(ctx, failsKey).Int(); err == nil && cnt >= magicLinkVerifyMax {
		return ErrMagicLinkVerifyLimited
	} else if err != nil && err != redis.Nil {
		// Don't fail open on transient Redis error — log and continue cautiously.
		slog.Warn("magiclink fails counter read error", "email", email, "error", err)
	}

	stored, err := s.redis.Get(ctx, codeKey).Result()
	if err == redis.Nil {
		s.recordFailure(ctx, failsKey)
		return ErrMagicLinkInvalid
	}
	if err != nil {
		return fmt.Errorf("verify magic link code: %w", err)
	}
	// SEC-07: constant-time comparison — защита от timing-side-channel.
	if subtle.ConstantTimeCompare([]byte(stored), []byte(code)) != 1 {
		s.recordFailure(ctx, failsKey)
		return ErrMagicLinkInvalid
	}

	// Success: delete code (one-shot) and clear fails counter.
	s.redis.Del(ctx, codeKey)
	s.redis.Del(ctx, failsKey)
	return nil
}

// recordFailure increments the failed-attempt counter for an email and sets the
// TTL on first failure to roughly match the code TTL.
func (s *MagicLinkService) recordFailure(ctx context.Context, failsKey string) {
	cnt, err := s.redis.Incr(ctx, failsKey).Result()
	if err != nil {
		slog.Warn("magiclink fails counter incr error", "key", failsKey, "error", err)
		return
	}
	if cnt == 1 {
		// First failure — set TTL roughly matching the code TTL window.
		if expErr := s.redis.Expire(ctx, failsKey, magicLinkTTL).Err(); expErr != nil {
			slog.Warn("magiclink fails counter expire error", "key", failsKey, "error", expErr)
		}
	}
}

// sentinel errors
var (
	ErrMagicLinkRateLimited   = fmt.Errorf("magic link rate limit exceeded")
	ErrMagicLinkInvalid       = fmt.Errorf("invalid or expired code")
	ErrMagicLinkVerifyLimited = fmt.Errorf("too many verification attempts, try again later")
)
