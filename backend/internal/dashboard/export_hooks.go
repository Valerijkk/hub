//go:build hubtest

package dashboard

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/valerii/hub/backend/ent/issue"
)

// ExportDashboardServiceRedis возвращает Redis-клиент сервиса для hubtest.
func ExportDashboardServiceRedis(s *DashboardService) *redis.Client {
	return s.rdb
}

// ExportProjectStatsTTL возвращает TTL кеша статистики проекта для hubtest.
func ExportProjectStatsTTL() time.Duration {
	return projectStatsTTL
}

// ExportPriorityRank возвращает числовой приоритет задачи для hubtest.
func ExportPriorityRank(p issue.Priority) int {
	return priorityRank(p)
}

// ExportGetUserID извлекает userID из Fiber context для hubtest.
func ExportGetUserID(c fiber.Ctx) (uuid.UUID, error) {
	return getUserID(c)
}
