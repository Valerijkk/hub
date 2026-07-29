package ai

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
)

// Repository — реестр AI-провайдеров и учёт usage.
type Repository interface {
	ListProviders(ctx context.Context, scope string) ([]*ent.AIProvider, error)
	GetProvider(ctx context.Context, id uuid.UUID) (*ent.AIProvider, error)
	CreateProvider(ctx context.Context, input CreateProviderInput) (*ent.AIProvider, error)
	UpdateProvider(ctx context.Context, id uuid.UUID, input UpdateProviderInput) (*ent.AIProvider, error)
	FindProviderForModel(ctx context.Context, userID uuid.UUID, model string) (*ent.AIProvider, error)
	ListProvidersForModel(ctx context.Context, userID uuid.UUID, model string) ([]*ent.AIProvider, error)
	IncrementUsage(ctx context.Context, userID uuid.UUID, tokensIn, tokensOut int) error
	ListModels(ctx context.Context, userID *uuid.UUID) ([]ModelInfo, error)
	ListUserProviders(ctx context.Context, userID uuid.UUID) ([]*ent.AIProvider, error)
	GetUserProvider(ctx context.Context, userID, providerID uuid.UUID) (*ent.AIProvider, error)
	GetDailyUsage(ctx context.Context, userID uuid.UUID) (tokensIn, tokensOut, requestCount int, err error)
	GetUserPreference(ctx context.Context, userID uuid.UUID) (defaultModel string, providerID *uuid.UUID, err error)
	GetUserPreferenceRecord(ctx context.Context, userID uuid.UUID) (*UserPreferenceRecord, error)
	SaveUserPreference(ctx context.Context, userID uuid.UUID, input SaveUserPreferenceInput) error
	GetUsageSummary(ctx context.Context, day time.Time) (requests, tokensIn, tokensOut, activeUsers int, err error)
	ListUsageHistory(ctx context.Context, days int) ([]UsageHistoryDay, error)
}

// UsageHistoryDay — агрегат AI usage за UTC-день.
type UsageHistoryDay struct {
	Date         string
	Requests     int
	TokensIn     int
	TokensOut    int
	ActiveUsers  int
}

// CreateProviderInput — payload создания админом.
type CreateProviderInput struct {
	Name         string
	ProviderType string
	BaseURL      string
	APIKey       string
	Models       []string
	Enabled      bool
	Priority     int
	MaxRPM       int
	RuResident   bool
	Scope        string
	OwnerUserID  *uuid.UUID
}

// UpdateProviderInput — частичное обновление.
type UpdateProviderInput struct {
	Name       *string
	BaseURL    *string
	APIKey     *string
	Models     []string
	Enabled    *bool
	Priority   *int
	MaxRPM     *int
	RuResident *bool
	// Version — optimistic lock (A3W5-3-006 / NFR-24); stale → ErrVersionConflict.
	Version int
}

// ModelInfo — модель для frontend picker.
type ModelInfo struct {
	ID           string `json:"id"`
	ProviderID   string `json:"provider_id"`
	ProviderName string `json:"provider_name"`
	RuResident   bool   `json:"ru_resident"`
}

// AuditLogger — metadata AI usage без prompt/response (privacy).
type AuditLogger interface {
	LogAIChat(ctx context.Context, userID uuid.UUID, providerID uuid.UUID, model string, tokensIn, tokensOut int, latency time.Duration, ip, userAgent string) error
}
