package status

import (
	"context"
	"time"
)

// Pinger проверяет доступность компонента.
type Pinger func(ctx context.Context) error

// Service агрегирует health probes для публичной status page.
type Service struct {
	postgres Pinger
	redis    Pinger
	mongo    Pinger
	nats     Pinger
}

// NewService создаёт сервис status page.
func NewService(postgres, redis, mongo, nats Pinger) *Service {
	return &Service{
		postgres: postgres,
		redis:    redis,
		mongo:    mongo,
		nats:     nats,
	}
}

// GetStatus возвращает агрегированный статус компонентов.
func (s *Service) GetStatus(ctx context.Context) Response {
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	components := []ComponentStatus{
		s.probe(checkCtx, "postgres", "PostgreSQL", "Primary database", s.postgres),
		s.probe(checkCtx, "redis", "Redis", "Sessions & cache", s.redis),
		s.probe(checkCtx, "mongo", "MongoDB", "Messages & reports", s.mongo),
		s.probe(checkCtx, "nats", "NATS", "Real-time events", s.nats),
		{ID: "api", Name: "Hub API", Status: "operational", Description: "Application layer"},
	}

	overall := "operational"
	for _, c := range components {
		if c.Status != "operational" {
			overall = "degraded"
			break
		}
	}

	return Response{
		Status:     overall,
		UpdatedAt:  time.Now().UTC(),
		Components: components,
	}
}

func (s *Service) probe(ctx context.Context, id, name, desc string, ping Pinger) ComponentStatus {
	st := "operational"
	if ping == nil {
		st = "unknown"
	} else if err := ping(ctx); err != nil {
		st = "outage"
	}
	return ComponentStatus{ID: id, Name: name, Status: st, Description: desc}
}
