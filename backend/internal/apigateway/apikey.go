package apigateway

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/apikey"
	"github.com/valerii/hub/backend/ent/user"
)

// APIKeyService управляет жизненным циклом API-ключей:
// генерация, верификация, отзыв, обновление last_used_at.
type APIKeyService struct {
	db *ent.Client
}

// NewAPIKeyService создаёт новый сервис API-ключей.
func NewAPIKeyService(db *ent.Client) *APIKeyService {
	return &APIKeyService{db: db}
}

// CreateAPIKey генерирует новый API-ключ для пользователя.
// Возвращает полный ключ ТОЛЬКО при создании — далее хранится лишь SHA-256 хеш.
func (s *APIKeyService) CreateAPIKey(ctx context.Context, userID uuid.UUID, name string, expiresAt *time.Time) (*CreateAPIKeyResponse, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return nil, fmt.Errorf("не удалось сгенерировать ключ: %w", err)
	}
	fullKey := base64.RawURLEncoding.EncodeToString(raw)
	prefix := fullKey[:8]
	hash := hashKey(fullKey)

	builder := s.db.APIKey.Create().
		SetName(name).
		SetKeyHash(hash).
		SetPrefix(prefix).
		SetUserID(userID)

	if expiresAt != nil {
		builder.SetExpiresAt(*expiresAt)
	}

	ak, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("не удалось сохранить API-ключ: %w", err)
	}

	return &CreateAPIKeyResponse{
		APIKeyResponse: APIKeyResponse{
			ID:        ak.ID.String(),
			Name:      ak.Name,
			Prefix:    ak.Prefix,
			ExpiresAt: ak.ExpiresAt,
			CreatedAt: ak.CreatedAt,
		},
		Key: fullKey,
	}, nil
}

// ListAPIKeys возвращает список API-ключей пользователя (без полных ключей).
func (s *APIKeyService) ListAPIKeys(ctx context.Context, userID uuid.UUID) ([]APIKeyResponse, error) {
	keys, err := s.db.APIKey.Query().
		Where(apikey.HasUserWith(user.ID(userID))).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить API-ключи: %w", err)
	}

	result := make([]APIKeyResponse, 0, len(keys))
	for _, k := range keys {
		result = append(result, APIKeyResponse{
			ID:         k.ID.String(),
			Name:       k.Name,
			Prefix:     k.Prefix,
			LastUsedAt: k.LastUsedAt,
			ExpiresAt:  k.ExpiresAt,
			CreatedAt:  k.CreatedAt,
		})
	}
	return result, nil
}

// RevokeAPIKey удаляет API-ключ пользователя.
func (s *APIKeyService) RevokeAPIKey(ctx context.Context, userID uuid.UUID, keyID uuid.UUID) error {
	count, err := s.db.APIKey.Query().
		Where(
			apikey.ID(keyID),
			apikey.HasUserWith(user.ID(userID)),
		).
		Count(ctx)
	if err != nil || count == 0 {
		return fiber.NewError(fiber.StatusNotFound, "API-ключ не найден")
	}
	return s.db.APIKey.DeleteOneID(keyID).Exec(ctx)
}

// VerifyAPIKey проверяет API-ключ по хешу и возвращает ID пользователя.
// Обновляет last_used_at в фоновой горутине.
func (s *APIKeyService) VerifyAPIKey(ctx context.Context, rawKey string) (uuid.UUID, error) {
	hash := hashKey(rawKey)

	ak, err := s.db.APIKey.Query().
		Where(apikey.KeyHash(hash)).
		WithUser().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "недействительный API-ключ")
		}
		return uuid.Nil, fmt.Errorf("ошибка проверки API-ключа: %w", err)
	}

	if ak.ExpiresAt != nil && ak.ExpiresAt.Before(time.Now()) {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "API-ключ истёк")
	}

	// Обновляем last_used_at в фоне (не блокируем запрос)
	go func() {
		bgCtx := context.Background()
		if err := s.db.APIKey.UpdateOneID(ak.ID).
			SetLastUsedAt(time.Now()).
			Exec(bgCtx); err != nil {
			slog.Error("не удалось обновить last_used_at для API-ключа", "error", err)
		}
	}()

	if ak.Edges.User == nil {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "пользователь API-ключа не найден")
	}

	return ak.Edges.User.ID, nil
}

// APIKeyAuth создаёт Fiber middleware для аутентификации через API-ключ.
// Работает параллельно с JWT: если X-API-Key присутствует — используется API key,
// иначе управление передаётся следующему middleware (JWT).
func APIKeyAuth(svc *APIKeyService) fiber.Handler {
	return func(c fiber.Ctx) error {
		key := c.Get("X-API-Key")
		if key == "" {
			return c.Next()
		}

		userID, err := svc.VerifyAPIKey(c.Context(), key)
		if err != nil {
			return err
		}

		c.Locals("userID", userID.String())
		return c.Next()
	}
}

// hashKey вычисляет SHA-256 хеш API-ключа и возвращает hex-строку.
func hashKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return fmt.Sprintf("%x", h[:])
}
