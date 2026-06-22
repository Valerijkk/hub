package auth

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/valerii/hub/backend/ent/user"
)

// userStatusCachePrefix — префикс Redis-ключа кэша статуса пользователя.
// Используется auth middleware для избегания обращения к БД на каждом
// запросе (ME-13).
const userStatusCachePrefix = "user:status:"

// userStatusCacheTTL — короткий TTL, чтобы деактивация пользователя
// применилась максимум за это окно даже без явной инвалидации.
const userStatusCacheTTL = 30 * time.Second

// Маркеры значений в кэше (короткие, чтобы не раздувать Redis).
const (
	userStatusCacheActive      = "active"
	userStatusCacheDeactivated = "deactivated"
)

// UserStatusCacheKey возвращает Redis-ключ для кэша статуса пользователя.
// Экспортирован, чтобы другие пакеты (audit) могли инвалидировать кэш
// при изменении статуса пользователя.
func UserStatusCacheKey(uid uuid.UUID) string {
	return userStatusCachePrefix + uid.String()
}

// InvalidateUserStatusCache удаляет закэшированный статус пользователя
// из Redis. Безопасен при rdb == nil (no-op).
//
// Вызывается из audit.DeactivateUser после изменения статуса в БД,
// чтобы не ждать TTL (см. ME-13).
func InvalidateUserStatusCache(ctx context.Context, rdb *redis.Client, uid uuid.UUID) {
	if rdb == nil {
		return
	}
	if err := rdb.Del(ctx, UserStatusCacheKey(uid)).Err(); err != nil {
		slog.Warn("invalidate user status cache", "user_id", uid, "error", err)
	}
}

// RequireAuth возвращает Fiber middleware для валидации JWT access-токенов.
// Сохраняет userID и sessionID в Locals для последующих обработчиков.
//
// ME-13: статус пользователя (active/deactivated) кэшируется в Redis на
// userStatusCacheTTL (~30s), чтобы не дёргать БД на каждом запросе.
// При деактивации кэш явно инвалидируется (см. InvalidateUserStatusCache).
func RequireAuth(svc *AuthService) fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := utils.CopyString(c.Get("Authorization"))
		if authHeader == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing authorization header")
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid authorization format")
		}

		tokenStr := parts[1]
		if tokenStr == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "empty token")
		}

		claims, err := svc.ParseAccessToken(tokenStr)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid or expired token")
		}

		// Проверяем, не деактивирован ли пользователь (Redis-кэш + fallback в БД).
		uid, parseErr := uuid.Parse(claims.UserID)
		if parseErr != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid token claims")
		}
		deactivated, statusErr := svc.IsUserDeactivated(c.Context(), uid)
		if statusErr != nil {
			return fiber.NewError(fiber.StatusServiceUnavailable, "unable to verify account status")
		}
		if deactivated {
			return fiber.NewError(fiber.StatusForbidden, "user account is deactivated")
		}

		c.Locals("userID", claims.UserID)
		c.Locals("sessionID", claims.SessionID)

		return c.Next()
	}
}

// userIsDeactivated возвращает true, если пользователь помечен как
// deactivated. Сначала пробует Redis-кэш, при промахе — читает БД и
// записывает результат в кэш.
//
// Ошибки Redis (промах или сбой) — graceful degradation в БД. Ошибка
// GetUser из БД пробрасывается наверх (fail-closed).
func userIsDeactivated(ctx context.Context, repo Repository, rdb *redis.Client, uid uuid.UUID) (bool, error) {
	cacheKey := UserStatusCacheKey(uid)

	// Попадание в кэш?
	if rdb != nil {
		switch val, err := rdb.Get(ctx, cacheKey).Result(); {
		case err == nil:
			return val == userStatusCacheDeactivated, nil
		case errors.Is(err, redis.Nil):
			// промах — читаем статус из БД.
		default:
			slog.Warn("user status cache get", "user_id", uid, "error", err)
			// продолжаем в БД (graceful degradation).
		}
	}

	// Промах кэша — fallback в БД.
	u, uErr := repo.GetUser(ctx, uid)
	if uErr != nil {
		return false, uErr
	}
	deactivated := u.Status == user.StatusDeactivated

	// Пишем кэш (best-effort).
	if rdb != nil {
		val := userStatusCacheActive
		if deactivated {
			val = userStatusCacheDeactivated
		}
		if err := rdb.Set(ctx, cacheKey, val, userStatusCacheTTL).Err(); err != nil {
			slog.Warn("user status cache set", "user_id", uid, "error", err)
		}
	}

	return deactivated, nil
}
