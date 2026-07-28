package apigateway

import "time"

// --- DTO API-ключей ---

// CreateAPIKeyRequest — запрос на создание API-ключа.
type CreateAPIKeyRequest struct {
	Name      string     `json:"name" validate:"required,min=1,max=255"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// APIKeyResponse — ответ с метаданными API-ключа (без полного ключа).
type APIKeyResponse struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Prefix     string     `json:"prefix"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// CreateAPIKeyResponse — ответ при создании, содержит полный ключ ОДИН раз.
type CreateAPIKeyResponse struct {
	APIKeyResponse
	Key string `json:"key"`
}

// --- DTO вебхуков ---

// CreateWebhookRequest — запрос на создание вебхук-эндпоинта.
type CreateWebhookRequest struct {
	URL    string   `json:"url" validate:"required,url"`
	Secret string   `json:"secret,omitempty"`
	Events []string `json:"events" validate:"required,min=1"`
}

// UpdateWebhookRequest — запрос на обновление вебхук-эндпоинта.
type UpdateWebhookRequest struct {
	URL      *string  `json:"url,omitempty"`
	Secret   *string  `json:"secret,omitempty"`
	Events   []string `json:"events,omitempty"`
	IsActive *bool    `json:"is_active,omitempty"`
}

// WebhookResponse — ответ с данными вебхук-эндпоинта (без secret).
type WebhookResponse struct {
	ID       string   `json:"id"`
	URL      string   `json:"url"`
	Events   []string `json:"events"`
	IsActive bool     `json:"is_active"`
	Created  string   `json:"created_at"`
}

// CreateWebhookResponse — ответ при создании: secret возвращается ОДИН раз
// (как API key `key`), list/update его не содержат (P1-BE-27 / W4-A5-003).
type CreateWebhookResponse struct {
	WebhookResponse
	Secret string `json:"secret"`
}

// WebhookDeliveryResponse — ответ с данными доставки вебхука.
type WebhookDeliveryResponse struct {
	ID           string                 `json:"id"`
	EventType    string                 `json:"event_type"`
	Payload      map[string]interface{} `json:"payload"`
	StatusCode   int                    `json:"status_code"`
	ResponseBody string                `json:"response_body,omitempty"`
	Error        string                 `json:"error,omitempty"`
	DeliveredAt  time.Time              `json:"delivered_at"`
}
