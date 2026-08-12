package ai

// ChatMessage — OpenAI-совместимое сообщение.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionRequest — тело proxy-запроса.
type ChatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Stream      bool          `json:"stream"`
	StreamID    string        `json:"stream_id,omitempty"`
	Temperature *float64      `json:"temperature,omitempty"`
	MaxTokens   *int          `json:"max_tokens,omitempty"`
}

// StreamStartResponse — принятый streaming job.
type StreamStartResponse struct {
	StreamID string `json:"stream_id"`
	Status   string `json:"status"`
}

// JQLAssistRequest — естественный язык → JQL.
type JQLAssistRequest struct {
	Query     string `json:"query"`
	ProjectID string `json:"project_id,omitempty"`
}

// JQLAssistResponse — сгенерированный JQL.
type JQLAssistResponse struct {
	JQL         string `json:"jql"`
	Explanation string `json:"explanation,omitempty"`
}

// ChatCompletionResponse — non-streaming OpenAI-compatible ответ.
type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int         `json:"index"`
		Message      ChatMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// ProviderResponse — элемент admin list (без API key).
type ProviderResponse struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	ProviderType string   `json:"provider_type"`
	BaseURL      string   `json:"base_url"`
	Models       []string `json:"models"`
	Enabled      bool     `json:"enabled"`
	Priority     int      `json:"priority"`
	MaxRPM       int      `json:"max_rpm"`
	RuResident   bool     `json:"ru_resident"`
	Scope        string   `json:"scope"`
	HasAPIKey    bool     `json:"has_api_key"`
	Version      int      `json:"version"`
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
}

// CreateProviderRequest — тело admin POST.
type CreateProviderRequest struct {
	Name         string   `json:"name"`
	ProviderType string   `json:"provider_type"`
	BaseURL      string   `json:"base_url"`
	APIKey       string   `json:"api_key"`
	Models       []string `json:"models"`
	Enabled      *bool    `json:"enabled"`
	Priority     *int     `json:"priority"`
	MaxRPM       *int     `json:"max_rpm"`
	RuResident   *bool    `json:"ru_resident"`
}

// UpdateProviderRequest — тело admin PATCH.
type UpdateProviderRequest struct {
	Name       *string  `json:"name"`
	BaseURL    *string  `json:"base_url"`
	APIKey     *string  `json:"api_key"`
	Models     []string `json:"models"`
	Enabled    *bool    `json:"enabled"`
	Priority   *int     `json:"priority"`
	MaxRPM     *int     `json:"max_rpm"`
	RuResident *bool    `json:"ru_resident"`
	Version    int      `json:"version"`
}

// TestProviderResponse — результат ping.
type TestProviderResponse struct {
	OK       bool   `json:"ok"`
	Latency  int64  `json:"latency_ms"`
	Model    string `json:"model,omitempty"`
	Error    string `json:"error,omitempty"`
	Provider string `json:"provider"`
}

// ModelsListResponse — пользовательский список моделей.
type ModelsListResponse struct {
	Models []ModelInfo `json:"models"`
}

// UsageTodayResponse — дневные AI counters (M3-025).
type UsageTodayResponse struct {
	UsageDate    string `json:"usage_date"`
	TokensIn     int    `json:"tokens_in"`
	TokensOut    int    `json:"tokens_out"`
	RequestCount int    `json:"request_count"`
	DailyLimit   int    `json:"daily_limit"`
}

// UserPreferenceResponse — сохранённая default model.
type UserPreferenceResponse struct {
	DefaultModel    string `json:"default_model"`
	AllowExternalAI bool   `json:"allow_external_ai"`
	Version         int    `json:"version"`
}

// UpdateUserPreferenceRequest — тело PATCH.
type UpdateUserPreferenceRequest struct {
	DefaultModel    *string `json:"default_model"`
	AllowExternalAI *bool   `json:"allow_external_ai"`
	Version         int     `json:"version"`
}

// UsageSummaryResponse — admin aggregate за день.
type UsageSummaryResponse struct {
	UsageDate     string `json:"usage_date"`
	TotalRequests int    `json:"total_requests"`
	TokensIn      int    `json:"tokens_in"`
	TokensOut     int    `json:"tokens_out"`
	ActiveUsers   int    `json:"active_users"`
}

// UsageHistoryDayResponse — один день admin usage chart.
type UsageHistoryDayResponse struct {
	Date        string `json:"date"`
	Requests    int    `json:"requests"`
	TokensIn    int    `json:"tokens_in"`
	TokensOut   int    `json:"tokens_out"`
	ActiveUsers int    `json:"active_users"`
}

// UsageHistoryResponse — GET /admin/ai/usage/history.
type UsageHistoryResponse struct {
	Days []UsageHistoryDayResponse `json:"days"`
}

// AdminSettingsResponse — GET /admin/ai/settings.
type AdminSettingsResponse struct {
	SystemPrompt string `json:"system_prompt"`
	OfflineOnly  bool   `json:"offline_only"`
}

// UpdateAdminSettingsRequest — PATCH /admin/ai/settings.
type UpdateAdminSettingsRequest struct {
	SystemPrompt *string `json:"system_prompt"`
}

// BotInfoResponse — системный бот @hub-ai для DM / UI.
type BotInfoResponse struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
}
