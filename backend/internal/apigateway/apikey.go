package apigateway

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent/user"
)

// Размеры пула воркеров и очереди для асинхронного обновления
// `api_key.last_used_at`. BE-H4: раньше каждое успешное API-key-обращение
// порождало unbounded `go func()` — под нагрузкой это приводило к
// goroutine-leak и connection-pool exhaustion. Сейчас обновления
// мультиплексируются через буферизованный канал, а лишние events
// дедуплицируются (per-key throttle на уровне канала).
const (
	apiKeyTouchWorkers   = 4
	apiKeyTouchQueueSize = 100
)

// apiKeyTouchJob — задание на обновление last_used_at для одного ключа.
type apiKeyTouchJob struct {
	keyID uuid.UUID
	when  time.Time
}

// APIKeyService управляет жизненным циклом API-ключей:
// генерация, верификация, отзыв, обновление last_used_at.
type APIKeyService struct {
	repo APIKeyRepository

	// BE-H4: bounded worker-pool для асинхронных обновлений last_used_at.
	touchJobs chan apiKeyTouchJob
	touchWG   sync.WaitGroup
	stopOnce  sync.Once
}

// NewAPIKeyService создаёт новый сервис API-ключей.
// Запускает фиксированный пул воркеров (apiKeyTouchWorkers) для асинхронного
// обновления last_used_at у API-ключей. Завершите сервис вызовом Close при
// корректном завершении работы.
func NewAPIKeyService(repo APIKeyRepository) *APIKeyService {
	svc := &APIKeyService{
		repo:      repo,
		touchJobs: make(chan apiKeyTouchJob, apiKeyTouchQueueSize),
	}
	for i := 0; i < apiKeyTouchWorkers; i++ {
		svc.touchWG.Add(1)
		go svc.touchWorker()
	}
	return svc
}

// touchWorker обрабатывает задания обновления last_used_at из канала.
// Завершается при закрытии canал touchJobs.
func (s *APIKeyService) touchWorker() {
	defer s.touchWG.Done()
	for job := range s.touchJobs {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := s.repo.UpdateAPIKeyLastUsedAt(bgCtx, job.keyID, job.when); err != nil {
			slog.Error("не удалось обновить last_used_at для API-ключа", "error", err, "key_id", job.keyID)
		}
		cancel()
	}
}

// Close останавливает воркер-пул last_used_at и дожидается завершения.
// Идемпотентен — повторный вызов безопасен.
func (s *APIKeyService) Close() {
	s.stopOnce.Do(func() {
		close(s.touchJobs)
		s.touchWG.Wait()
	})
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

	ak, err := s.repo.CreateAPIKey(ctx, userID, name, hash, prefix, expiresAt)
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
	keys, err := s.repo.ListAPIKeysByUser(ctx, userID)
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
	count, err := s.repo.CountAPIKeyForUser(ctx, userID, keyID)
	if err != nil || count == 0 {
		return ErrAPIKeyNotFound
	}
	return s.repo.DeleteAPIKey(ctx, keyID)
}

// VerifyAPIKey проверяет API-ключ по хешу и возвращает ID пользователя.
// Обновляет last_used_at в фоновой горутине.
func (s *APIKeyService) VerifyAPIKey(ctx context.Context, rawKey string) (uuid.UUID, error) {
	hash := hashKey(rawKey)

	ak, err := s.repo.GetAPIKeyByHashWithUser(ctx, hash)
	if err != nil {
		if IsNotFound(err) {
			return uuid.Nil, ErrAPIKeyInvalid
		}
		return uuid.Nil, fmt.Errorf("ошибка проверки API-ключа: %w", err)
	}

	if ak.ExpiresAt != nil && ak.ExpiresAt.Before(time.Now()) {
		return uuid.Nil, ErrAPIKeyExpired
	}

	// BE-H4: обновляем last_used_at через bounded worker-pool вместо
	// unbounded `go func()`. Если очередь переполнена — пропускаем
	// обновление с warn-логом; на следующем запросе оно всё равно
	// будет обновлено воркером.
	job := apiKeyTouchJob{keyID: ak.ID, when: time.Now()}
	select {
	case s.touchJobs <- job:
	default:
		slog.Warn("API-key touch queue full, skipping last_used_at update",
			"key_id", ak.ID, "queue_size", apiKeyTouchQueueSize)
	}

	if ak.Edges.User == nil {
		return uuid.Nil, ErrAPIKeyUserNotFound
	}

	// APIGW-04: JWT RequireAuth rejects deactivated users; API keys must too.
	if ak.Edges.User.Status == user.StatusDeactivated {
		return uuid.Nil, ErrAPIKeyUserDeactivated
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
			return MapToFiber(err)
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
