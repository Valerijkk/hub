package ai

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/aiprovider"
	"github.com/valerii/hub/backend/internal/search"
)

var (
	ErrProviderDisabled       = fmt.Errorf("ai provider disabled")
	ErrModelNotAllowed        = fmt.Errorf("model not in provider allowlist")
	ErrNoProvider             = fmt.Errorf("no provider configured for model")
	ErrDailyQuotaExceeded     = fmt.Errorf("daily AI request quota exceeded")
	ErrSettingsTooLong        = fmt.Errorf("system prompt too long")
	ErrSettingsNotConfigured  = fmt.Errorf("settings store not configured")
	ErrProviderNotFound       = fmt.Errorf("provider not found")
	// ErrQueryRequired — AssistJQL: пустой/blank query (A3W5-1-002 → HTTP 400).
	ErrQueryRequired = fmt.Errorf("query is required")
	// ErrVersionConflict — optimistic lock stale (A3W5-3-006 / NFR-24 → HTTP 409).
	ErrVersionConflict = fmt.Errorf("version conflict")
)

// DefaultDailyRequestLimit — requests/day на пользователя (M3-027).
const DefaultDailyRequestLimit = 500

// msgAIStreamFailed — фиксированный WS error payload (A3W5-3-007); сырой err только в slog.
const msgAIStreamFailed = "AI stream failed"

const jqlAssistSystemPrompt = `Ты конвертируешь запрос на естественном языке в Hub JQL (Jira Query Language).
Поддерживаемые поля: status, priority, type, assignee, created, updated.
Операторы: =, !=, IN, NOT IN, >, <, >=, <=.
Функции: currentUser().
Примеры:
- "мои открытые задачи" → assignee = currentUser() AND status = "Open"
- "критичные баги" → type = "bug" AND priority = "critical"
Ответь ТОЛЬКО валидной JQL-строкой без пояснений и markdown.`

// Service — бизнес-логика AI gateway.
type Service struct {
	repo         Repository
	adapter      ChatAdapter
	giga         *GigaChatAdapter
	anthropic    *AnthropicAdapter
	yandex       *YandexGPTAdapter
	rpm          ProviderRPMLimiter
	settings     SettingsStore
	offlineOnly  bool
	audit        AuditLogger
	streams      StreamPublisher
}

// NewService создаёт AI gateway service.
func NewService(repo Repository, adapter ChatAdapter, audit AuditLogger) *Service {
	return &Service{
		repo:      repo,
		adapter:   adapter,
		giga:      NewGigaChatAdapter(),
		anthropic: NewAnthropicAdapter(),
		yandex:    NewYandexGPTAdapter(),
		rpm:       NoopProviderRPMLimiter{},
		audit:     audit,
	}
}

// SetProviderRPMLimiter — применение Redis RPM (W9).
func (s *Service) SetProviderRPMLimiter(l ProviderRPMLimiter) {
	if l != nil {
		s.rpm = l
	}
}

// SetSettingsStore — SystemSetting для bot prompt (W10).
func (s *Service) SetSettingsStore(store SettingsStore) {
	s.settings = store
}

// SetOfflineOnly — только local/Ollama (AI-SPEC S8).
func (s *Service) SetOfflineOnly(v bool) {
	s.offlineOnly = v
}

// SetStreamPublisher подключает доставку WebSocket-стрима (Phase 17 W3-03).
func (s *Service) SetStreamPublisher(p StreamPublisher) {
	s.streams = p
}

func (s *Service) chatAdapterFor(p *ent.AIProvider) ChatAdapter {
	switch p.ProviderType {
	case aiprovider.ProviderTypeGigachat:
		return s.giga
	case aiprovider.ProviderTypeAnthropic:
		return s.anthropic
	default:
		if IsYandexProvider(p.BaseURL) {
			return s.yandex
		}
		return s.adapter
	}
}

// ListProviders — admin-safe список провайдеров.
func (s *Service) ListProviders(ctx context.Context) ([]ProviderResponse, error) {
	items, err := s.repo.ListProviders(ctx, "system")
	if err != nil {
		return nil, err
	}
	out := make([]ProviderResponse, 0, len(items))
	for _, p := range items {
		out = append(out, toProviderResponse(p))
	}
	return out, nil
}

// ListUserProviders — BYOK провайдеры пользователя.
func (s *Service) ListUserProviders(ctx context.Context, userID uuid.UUID) ([]ProviderResponse, error) {
	items, err := s.repo.ListUserProviders(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]ProviderResponse, 0, len(items))
	for _, p := range items {
		out = append(out, toProviderResponse(p))
	}
	return out, nil
}

// CreateUserProvider — BYOK провайдер пользователя.
func (s *Service) CreateUserProvider(ctx context.Context, userID uuid.UUID, req CreateProviderRequest) (*ProviderResponse, error) {
	pt := req.ProviderType
	if pt == "" {
		pt = string(aiprovider.ProviderTypeOpenaiCompatible)
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	p, err := s.repo.CreateProvider(ctx, CreateProviderInput{
		Name:         req.Name,
		ProviderType: pt,
		BaseURL:      req.BaseURL,
		APIKey:       req.APIKey,
		Models:       req.Models,
		Enabled:      enabled,
		Priority:     0,
		MaxRPM:       60,
		RuResident:   req.RuResident != nil && *req.RuResident,
		Scope:        "user",
		OwnerUserID:  &userID,
	})
	if err != nil {
		return nil, err
	}
	resp := toProviderResponse(p)
	return &resp, nil
}

// UpdateUserProvider патчит BYOK провайдера.
func (s *Service) UpdateUserProvider(ctx context.Context, userID, providerID uuid.UUID, req UpdateProviderRequest) (*ProviderResponse, error) {
	if _, err := s.repo.GetUserProvider(ctx, userID, providerID); err != nil {
		return nil, err
	}
	input := UpdateProviderInput{
		Name: req.Name, BaseURL: req.BaseURL, Models: req.Models,
		Enabled: req.Enabled, RuResident: req.RuResident,
		Version: req.Version,
	}
	if req.APIKey != nil {
		input.APIKey = req.APIKey
	}
	p, err := s.repo.UpdateProvider(ctx, providerID, input)
	if err != nil {
		return nil, err
	}
	resp := toProviderResponse(p)
	return &resp, nil
}

// CreateProvider добавляет системного провайдера.
func (s *Service) CreateProvider(ctx context.Context, req CreateProviderRequest) (*ProviderResponse, error) {
	pt := req.ProviderType
	if pt == "" {
		pt = string(aiprovider.ProviderTypeOpenaiCompatible)
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	priority := 0
	if req.Priority != nil {
		priority = *req.Priority
	}
	maxRPM := 60
	if req.MaxRPM != nil {
		maxRPM = *req.MaxRPM
	}
	ru := false
	if req.RuResident != nil {
		ru = *req.RuResident
	}
	p, err := s.repo.CreateProvider(ctx, CreateProviderInput{
		Name:         req.Name,
		ProviderType: pt,
		BaseURL:      req.BaseURL,
		APIKey:       req.APIKey,
		Models:       req.Models,
		Enabled:      enabled,
		Priority:     priority,
		MaxRPM:       maxRPM,
		RuResident:   ru,
		Scope:        "system",
	})
	if err != nil {
		return nil, err
	}
	resp := toProviderResponse(p)
	return &resp, nil
}

// UpdateProvider патчит config системного провайдера (admin).
// W5-A1-010: user-scope BYOK недоступен через /admin/ai — 404.
func (s *Service) UpdateProvider(ctx context.Context, id uuid.UUID, req UpdateProviderRequest) (*ProviderResponse, error) {
	if _, err := s.requireSystemProvider(ctx, id); err != nil {
		return nil, err
	}
	input := UpdateProviderInput{
		Name: req.Name, BaseURL: req.BaseURL, Models: req.Models,
		Enabled: req.Enabled, Priority: req.Priority, MaxRPM: req.MaxRPM, RuResident: req.RuResident,
		Version: req.Version,
	}
	if req.APIKey != nil {
		input.APIKey = req.APIKey
	}
	p, err := s.repo.UpdateProvider(ctx, id, input)
	if err != nil {
		return nil, err
	}
	resp := toProviderResponse(p)
	return &resp, nil
}

// TestProvider — ping системного провайдера minimal chat request.
func (s *Service) TestProvider(ctx context.Context, id uuid.UUID) (*TestProviderResponse, error) {
	p, err := s.requireSystemProvider(ctx, id)
	if err != nil {
		return nil, err
	}
	model := pickTestModel(p.Models)
	if model == "" {
		return &TestProviderResponse{
			OK: false, Provider: p.Name, Error: "no models configured",
		}, nil
	}
	start := time.Now()
	_, err = s.chatAdapterFor(p).ChatCompletion(ctx, p.BaseURL, p.APIKeyEnc, model, ChatCompletionRequest{
		Model: model, Messages: []ChatMessage{{Role: "user", Content: "ping"}},
	})
	latency := time.Since(start).Milliseconds()
	if err != nil {
		// A2W5-A1-004: JSON.error без upstream URL/auth/internals.
		return &TestProviderResponse{
			OK: false, Provider: p.Name, Model: model, Latency: latency,
			Error: sanitizeProviderTestError(err),
		}, nil
	}
	return &TestProviderResponse{
		OK: true, Provider: p.Name, Model: model, Latency: latency,
	}, nil
}

// requireSystemProvider — admin-scope guard: user BYOK → ErrProviderNotFound (W5-A1-010).
func (s *Service) requireSystemProvider(ctx context.Context, id uuid.UUID) (*ent.AIProvider, error) {
	p, err := s.repo.GetProvider(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrProviderNotFound
		}
		return nil, err
	}
	// Пустой scope — legacy system (до появления BYOK).
	if p.Scope != "" && p.Scope != aiprovider.ScopeSystem {
		return nil, ErrProviderNotFound
	}
	return p, nil
}

// ListModels — enabled models для ModelPicker (system + BYOK).
func (s *Service) ListModels(ctx context.Context, userID uuid.UUID) (*ModelsListResponse, error) {
	models, err := s.repo.ListModels(ctx, &userID)
	if err != nil {
		return nil, err
	}
	return &ModelsListResponse{Models: models}, nil
}

// DefaultModel — preference пользователя или первая enabled модель.
func (s *Service) DefaultModel(ctx context.Context, userID uuid.UUID) (string, error) {
	if model, err := s.preferredModel(ctx, userID); err == nil && model != "" {
		return model, nil
	}
	resp, err := s.ListModels(ctx, userID)
	if err != nil {
		return "", err
	}
	if len(resp.Models) == 0 {
		return "", ErrNoProvider
	}
	return resp.Models[0].ID, nil
}

func (s *Service) preferredModel(ctx context.Context, userID uuid.UUID) (string, error) {
	model, _, err := s.repo.GetUserPreference(ctx, userID)
	if err != nil || model == "" {
		return "", err
	}
	if _, err := s.resolveProvider(ctx, userID, model); err != nil {
		return "", err
	}
	return model, nil
}

// GetUserPreference — default model и privacy flags.
func (s *Service) GetUserPreference(ctx context.Context, userID uuid.UUID) (*UserPreferenceResponse, error) {
	rec, err := s.repo.GetUserPreferenceRecord(ctx, userID)
	if err != nil {
		return nil, err
	}
	// W5-A3-012 / P2-BE-17: default AllowExternalAI=false (opt-in).
	resp := &UserPreferenceResponse{AllowExternalAI: false}
	if rec != nil {
		resp.DefaultModel = rec.DefaultModel
		resp.AllowExternalAI = rec.AllowExternalAI
		resp.Version = rec.Version
	}
	return resp, nil
}

// SaveUserPreference — default model и privacy для ModelPicker.
func (s *Service) SaveUserPreference(ctx context.Context, userID uuid.UUID, req UpdateUserPreferenceRequest) (*UserPreferenceResponse, error) {
	input := SaveUserPreferenceInput{Version: req.Version}
	if req.DefaultModel != nil {
		model := strings.TrimSpace(*req.DefaultModel)
		if model != "" {
			if _, err := s.resolveProvider(ctx, userID, model); err != nil {
				return nil, ErrModelNotAllowed
			}
		}
		input.DefaultModel = &model
	}
	if req.AllowExternalAI != nil {
		input.AllowExternalAI = req.AllowExternalAI
	}
	if input.DefaultModel == nil && input.AllowExternalAI == nil {
		return s.GetUserPreference(ctx, userID)
	}
	if err := s.repo.SaveUserPreference(ctx, userID, input); err != nil {
		return nil, err
	}
	return s.GetUserPreference(ctx, userID)
}

// GetUsageToday — token/request counters пользователя (M3-025).
func (s *Service) GetUsageToday(ctx context.Context, userID uuid.UUID) (*UsageTodayResponse, error) {
	in, out, count, err := s.repo.GetDailyUsage(ctx, userID)
	if err != nil {
		return nil, err
	}
	day := time.Now().UTC().Format("2006-01-02")
	return &UsageTodayResponse{
		UsageDate:    day,
		TokensIn:     in,
		TokensOut:    out,
		RequestCount: count,
		DailyLimit:   DefaultDailyRequestLimit,
	}, nil
}

// GetUsageSummary — admin aggregate за сегодня.
func (s *Service) GetUsageSummary(ctx context.Context) (*UsageSummaryResponse, error) {
	req, in, out, users, err := s.repo.GetUsageSummary(ctx, time.Now())
	if err != nil {
		return nil, err
	}
	return &UsageSummaryResponse{
		UsageDate:     time.Now().UTC().Format("2006-01-02"),
		TotalRequests: req,
		TokensIn:      in,
		TokensOut:     out,
		ActiveUsers:   users,
	}, nil
}

// GetUsageHistory — дневные агрегаты для admin charts (W9).
func (s *Service) GetUsageHistory(ctx context.Context, days int) (*UsageHistoryResponse, error) {
	if days <= 0 {
		days = 7
	}
	if days > 90 {
		days = 90
	}
	rows, err := s.repo.ListUsageHistory(ctx, days)
	if err != nil {
		return nil, err
	}
	out := make([]UsageHistoryDayResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, UsageHistoryDayResponse{
			Date:         r.Date,
			Requests:     r.Requests,
			TokensIn:     r.TokensIn,
			TokensOut:    r.TokensOut,
			ActiveUsers:  r.ActiveUsers,
		})
	}
	return &UsageHistoryResponse{Days: out}, nil
}

func (s *Service) enforceQuota(ctx context.Context, userID uuid.UUID) error {
	_, _, count, err := s.repo.GetDailyUsage(ctx, userID)
	if err != nil {
		return err
	}
	if count >= DefaultDailyRequestLimit {
		return ErrDailyQuotaExceeded
	}
	return nil
}

// ChatCompletion проксирует chat к настроенному провайдеру.
func (s *Service) ChatCompletion(ctx context.Context, userID uuid.UUID, req ChatCompletionRequest, ip, userAgent string) (*ChatCompletionResponse, error) {
	if err := validateChatRequest(req); err != nil {
		return nil, err
	}
	if err := s.enforceQuota(ctx, userID); err != nil {
		return nil, err
	}
	if req.Stream {
		return nil, fmt.Errorf("use handler streaming path")
	}
	resp, used, latency, err := s.chatCompletionWithFallback(ctx, userID, req.Model, req)
	if err != nil {
		return nil, err
	}
	s.recordUsage(ctx, userID, used, req.Model, resp.Usage.PromptTokens, resp.Usage.CompletionTokens, latency, ip, userAgent)
	return resp, nil
}

// StreamChatCompletion синхронный stream для bot (M3-010).
func (s *Service) StreamChatCompletion(ctx context.Context, userID uuid.UUID, req ChatCompletionRequest, ip, userAgent string, onDelta func(content string) error) (string, error) {
	if err := validateChatRequest(req); err != nil {
		return "", err
	}
	if err := s.enforceQuota(ctx, userID); err != nil {
		return "", err
	}
	var full strings.Builder
	used, latency, err := s.streamCompletionWithFallback(ctx, userID, req.Model, req, func(d StreamDelta) error {
		if d.Done {
			return nil
		}
		full.WriteString(d.Content)
		if onDelta != nil && d.Content != "" {
			if err := onDelta(d.Content); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	out := full.String()
	outTokens := len(out) / 4
	if outTokens < 1 {
		outTokens = 1
	}
	inTokens := estimateTokens(req.Messages)
	s.recordUsage(ctx, userID, used, req.Model, inTokens, outTokens, latency, ip, userAgent)
	return out, nil
}

func (s *Service) StartChatStream(ctx context.Context, userID uuid.UUID, req ChatCompletionRequest, ip, userAgent string) error {
	if err := validateChatRequest(req); err != nil {
		return err
	}
	if err := s.enforceQuota(ctx, userID); err != nil {
		return err
	}
	if strings.TrimSpace(req.StreamID) == "" {
		return fmt.Errorf("stream_id required")
	}
	if s.streams == nil {
		return fmt.Errorf("streaming not configured")
	}
	// W5-A1-002: preflight consent/RPM до 202 — те же sentinel, что non-stream.
	if _, err := s.resolveProviders(ctx, userID, req.Model); err != nil {
		return err
	}
	// W5-A3-005: не Background() — cancel родителя/таймаут останавливает SSE.
	// WithoutCancel: HTTP handler отдаёт 202 и завершает request ctx; стрим живёт
	// по собственному deadline, но отмена явного parent (тесты/WS) пробрасывается
	// через отдельный cancel, если ctx уже отменён до старта.
	streamCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Minute)
	if err := ctx.Err(); err != nil {
		cancel()
		return err
	}
	go func() {
		defer cancel()
		s.runChatStream(streamCtx, userID, req, ip, userAgent)
	}()
	return nil
}

func (s *Service) runChatStream(ctx context.Context, userID uuid.UUID, req ChatCompletionRequest, ip, userAgent string) {
	streamID := req.StreamID
	publish := func(event string, payload map[string]interface{}) {
		if s.streams != nil {
			s.streams.PublishStreamEvent(userID, streamID, event, payload)
		}
	}
	publish("start", map[string]interface{}{"model": req.Model})

	var full strings.Builder
	used, latency, err := s.streamCompletionWithFallback(ctx, userID, req.Model, req, func(d StreamDelta) error {
		if d.Done {
			return nil
		}
		full.WriteString(d.Content)
		publish("delta", map[string]interface{}{"content": d.Content})
		return nil
	})
	if err != nil {
		slog.Warn("ai stream failed", "error", err, "stream_id", streamID)
		// A3W5-3-007: клиенту фиксированная строка; detail только в slog.
		publish("error", map[string]interface{}{"message": msgAIStreamFailed})
		return
	}
	outTokens := len(full.String()) / 4
	if outTokens < 1 {
		outTokens = 1
	}
	inTokens := estimateTokens(req.Messages)
	_ = s.repo.IncrementUsage(ctx, userID, inTokens, outTokens)
	if s.audit != nil && used != nil {
		_ = s.audit.LogAIChat(ctx, userID, used.ID, req.Model, inTokens, outTokens, latency, ip, userAgent)
	}
	publish("done", map[string]interface{}{
		"content": full.String(),
		"model":   req.Model,
	})
}

// AssistJQL — NL → Hub JQL через LLM (Phase 17 W5).
func (s *Service) AssistJQL(ctx context.Context, userID uuid.UUID, req JQLAssistRequest) (*JQLAssistResponse, error) {
	query := strings.TrimSpace(req.Query)
	if query == "" {
		return nil, ErrQueryRequired
	}
	model, err := s.DefaultModel(ctx, userID)
	if err != nil {
		return nil, err
	}
	resp, err := s.ChatCompletion(ctx, userID, ChatCompletionRequest{
		Model: model,
		Messages: []ChatMessage{
			{Role: "system", Content: jqlAssistSystemPrompt},
			{Role: "user", Content: query},
		},
	}, "", "jql-assist")
	if err != nil {
		return nil, err
	}
	jql := strings.TrimSpace(extractAssistantText(resp))
	jql = strings.Trim(jql, "`")
	if _, parseErr := search.ParseJQL(jql); parseErr != nil {
		return &JQLAssistResponse{
			JQL:         jql,
			Explanation: "warning: generated JQL may need manual fix: " + parseErr.Error(),
		}, nil
	}
	return &JQLAssistResponse{JQL: jql}, nil
}

func (s *Service) resolveProvider(ctx context.Context, userID uuid.UUID, model string) (*ent.AIProvider, error) {
	providers, err := s.resolveProviders(ctx, userID, model)
	if err != nil {
		return nil, err
	}
	return providers[0], nil
}

func validateChatRequest(req ChatCompletionRequest) error {
	if strings.TrimSpace(req.Model) == "" {
		return fmt.Errorf("model is required")
	}
	if len(req.Messages) == 0 {
		return fmt.Errorf("messages are required")
	}
	if err := CheckContentPolicy(req.Messages); err != nil {
		return err
	}
	return nil
}

func (s *Service) recordUsage(ctx context.Context, userID uuid.UUID, p *ent.AIProvider, model string, tokensIn, tokensOut int, latency time.Duration, ip, userAgent string) {
	_ = s.repo.IncrementUsage(ctx, userID, tokensIn, tokensOut)
	if s.audit != nil {
		_ = s.audit.LogAIChat(ctx, userID, p.ID, model, tokensIn, tokensOut, latency, ip, userAgent)
	}
}

func estimateTokens(msgs []ChatMessage) int {
	n := 0
	for _, m := range msgs {
		n += len(m.Content) / 4
	}
	if n < 1 {
		n = 1
	}
	return n
}

func toProviderResponse(p *ent.AIProvider) ProviderResponse {
	return ProviderResponse{
		ID:           p.ID.String(),
		Name:         p.Name,
		ProviderType: string(p.ProviderType),
		BaseURL:      p.BaseURL,
		Models:       append([]string(nil), p.Models...),
		Enabled:      p.Enabled,
		Priority:     p.Priority,
		MaxRPM:       p.MaxRpm,
		RuResident:   p.RuResident,
		Scope:        string(p.Scope),
		HasAPIKey:    p.APIKeyEnc != "",
		Version:      p.Version,
		CreatedAt:    p.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:    p.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func pickTestModel(models []string) string {
	if len(models) == 0 {
		return ""
	}
	return models[0]
}
