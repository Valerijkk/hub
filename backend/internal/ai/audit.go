package ai

import (
	"context"
	"time"

	"github.com/google/uuid"
	auditmod "github.com/valerii/hub/backend/internal/audit"
)

// AuditRepoLogger пишет AI chat metadata в audit_log.
type AuditRepoLogger struct {
	repo auditmod.Repository
}

// NewAuditRepoLogger — audit logger на audit repository.
func NewAuditRepoLogger(repo auditmod.Repository) *AuditRepoLogger {
	return &AuditRepoLogger{repo: repo}
}

// LogAIChat — provider/model/tokens без prompt (privacy).
func (l *AuditRepoLogger) LogAIChat(ctx context.Context, userID uuid.UUID, providerID uuid.UUID, model string, tokensIn, tokensOut int, latency time.Duration, ip, userAgent string) error {
	details := map[string]interface{}{
		"provider_id":   providerID.String(),
		"model":         model,
		"tokens_in":     tokensIn,
		"tokens_out":    tokensOut,
		"latency_ms":    latency.Milliseconds(),
	}
	_, err := l.repo.CreateAuditLog(ctx, "ai.chat", "ai_provider", providerID.String(), &userID, nil, details, ip, userAgent)
	return err
}
