package apigateway

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/apikey"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/ent/webhookdelivery"
	"github.com/valerii/hub/backend/ent/webhookendpoint"
)

// APIKeyRepository — персистентность API key (ARCH-01).
type APIKeyRepository interface {
	CreateAPIKey(ctx context.Context, userID uuid.UUID, name, hash, prefix string, expiresAt *time.Time) (*ent.APIKey, error)
	ListAPIKeysByUser(ctx context.Context, userID uuid.UUID) ([]*ent.APIKey, error)
	CountAPIKeyForUser(ctx context.Context, userID, keyID uuid.UUID) (int, error)
	DeleteAPIKey(ctx context.Context, keyID uuid.UUID) error
	GetAPIKeyByHashWithUser(ctx context.Context, hash string) (*ent.APIKey, error)
	UpdateAPIKeyLastUsedAt(ctx context.Context, keyID uuid.UUID, when time.Time) error
}

// CreateWebhookDeliveryInput — строка webhook delivery log.
type CreateWebhookDeliveryInput struct {
	EventType    string
	Payload      map[string]interface{}
	StatusCode   int
	ResponseBody string
	EndpointID   uuid.UUID
	ErrorMsg     string
}

// UpdateWebhookEndpointInput — частичное обновление webhook-эндпоинта.
type UpdateWebhookEndpointInput struct {
	URL      *string
	Secret   *string
	Events   []string
	IsActive *bool
}

// WebhookRepository — webhook endpoints и delivery logs (ARCH-01).
type WebhookRepository interface {
	ListActiveWebhookEndpoints(ctx context.Context) ([]*ent.WebhookEndpoint, error)
	CreateWebhookDelivery(ctx context.Context, input CreateWebhookDeliveryInput) error
	CreateWebhookEndpoint(ctx context.Context, projectID uuid.UUID, url, secret string, events []string) (*ent.WebhookEndpoint, error)
	ListWebhookEndpointsByProject(ctx context.Context, projectID uuid.UUID) ([]*ent.WebhookEndpoint, error)
	UpdateWebhookEndpoint(ctx context.Context, endpointID uuid.UUID, input UpdateWebhookEndpointInput) (*ent.WebhookEndpoint, error)
	DeleteWebhookEndpoint(ctx context.Context, endpointID uuid.UUID) error
	ListWebhookDeliveries(ctx context.Context, endpointID uuid.UUID, limit int) ([]*ent.WebhookDelivery, error)
	// WebhookEndpointProjectID возвращает projectID владеющего эндпоинтом проекта
	// (APIGW-01): используется для RBAC-проверки на маршрутах /webhooks/:id,
	// где projectId отсутствует в пути — нельзя доверять клиентскому ID.
	WebhookEndpointProjectID(ctx context.Context, endpointID uuid.UUID) (uuid.UUID, error)
}

// Repository объединяет API key и webhook — один Ent-клиент (ARCH-01).
type Repository interface {
	APIKeyRepository
	WebhookRepository
}

// EntRepository — реализация Repository через Ent ORM.
type EntRepository struct {
	client *ent.Client
}

// NewEntRepository создаёт Ent-репозиторий apigateway.
func NewEntRepository(client *ent.Client) *EntRepository {
	return &EntRepository{client: client}
}

func (r *EntRepository) CreateAPIKey(ctx context.Context, userID uuid.UUID, name, hash, prefix string, expiresAt *time.Time) (*ent.APIKey, error) {
	b := r.client.APIKey.Create().
		SetName(name).
		SetKeyHash(hash).
		SetPrefix(prefix).
		SetUserID(userID)
	if expiresAt != nil {
		b.SetExpiresAt(*expiresAt)
	}
	return b.Save(ctx)
}

func (r *EntRepository) ListAPIKeysByUser(ctx context.Context, userID uuid.UUID) ([]*ent.APIKey, error) {
	return r.client.APIKey.Query().
		Where(apikey.HasUserWith(user.ID(userID))).
		All(ctx)
}

func (r *EntRepository) CountAPIKeyForUser(ctx context.Context, userID, keyID uuid.UUID) (int, error) {
	return r.client.APIKey.Query().
		Where(apikey.ID(keyID), apikey.HasUserWith(user.ID(userID))).
		Count(ctx)
}

func (r *EntRepository) DeleteAPIKey(ctx context.Context, keyID uuid.UUID) error {
	return r.client.APIKey.DeleteOneID(keyID).Exec(ctx)
}

func (r *EntRepository) GetAPIKeyByHashWithUser(ctx context.Context, hash string) (*ent.APIKey, error) {
	return r.client.APIKey.Query().
		Where(apikey.KeyHash(hash)).
		WithUser().
		Only(ctx)
}

func (r *EntRepository) UpdateAPIKeyLastUsedAt(ctx context.Context, keyID uuid.UUID, when time.Time) error {
	return r.client.APIKey.UpdateOneID(keyID).
		SetLastUsedAt(when).
		Exec(ctx)
}

func (r *EntRepository) ListActiveWebhookEndpoints(ctx context.Context) ([]*ent.WebhookEndpoint, error) {
	return r.client.WebhookEndpoint.Query().
		Where(webhookendpoint.IsActive(true)).
		All(ctx)
}

func (r *EntRepository) CreateWebhookDelivery(ctx context.Context, input CreateWebhookDeliveryInput) error {
	b := r.client.WebhookDelivery.Create().
		SetEventType(input.EventType).
		SetPayload(input.Payload).
		SetStatusCode(input.StatusCode).
		SetResponseBody(input.ResponseBody).
		SetEndpointID(input.EndpointID)
	if input.ErrorMsg != "" {
		b.SetError(input.ErrorMsg)
	}
	_, err := b.Save(ctx)
	return err
}

func (r *EntRepository) CreateWebhookEndpoint(ctx context.Context, projectID uuid.UUID, url, secret string, events []string) (*ent.WebhookEndpoint, error) {
	return r.client.WebhookEndpoint.Create().
		SetURL(url).
		SetSecret(secret).
		SetEvents(events).
		SetProjectID(projectID).
		Save(ctx)
}

func (r *EntRepository) ListWebhookEndpointsByProject(ctx context.Context, projectID uuid.UUID) ([]*ent.WebhookEndpoint, error) {
	return r.client.WebhookEndpoint.Query().
		Where(webhookendpoint.HasProjectWith(project.ID(projectID))).
		All(ctx)
}

func (r *EntRepository) UpdateWebhookEndpoint(ctx context.Context, endpointID uuid.UUID, input UpdateWebhookEndpointInput) (*ent.WebhookEndpoint, error) {
	b := r.client.WebhookEndpoint.UpdateOneID(endpointID)
	if input.URL != nil {
		b.SetURL(*input.URL)
	}
	if input.Secret != nil {
		b.SetSecret(*input.Secret)
	}
	if input.Events != nil {
		b.SetEvents(input.Events)
	}
	if input.IsActive != nil {
		b.SetIsActive(*input.IsActive)
	}
	return b.Save(ctx)
}

func (r *EntRepository) DeleteWebhookEndpoint(ctx context.Context, endpointID uuid.UUID) error {
	return r.client.WebhookEndpoint.DeleteOneID(endpointID).Exec(ctx)
}

func (r *EntRepository) WebhookEndpointProjectID(ctx context.Context, endpointID uuid.UUID) (uuid.UUID, error) {
	return r.client.WebhookEndpoint.Query().
		Where(webhookendpoint.ID(endpointID)).
		QueryProject().
		OnlyID(ctx)
}

func (r *EntRepository) ListWebhookDeliveries(ctx context.Context, endpointID uuid.UUID, limit int) ([]*ent.WebhookDelivery, error) {
	return r.client.WebhookDelivery.Query().
		Where(webhookdelivery.HasEndpointWith(webhookendpoint.ID(endpointID))).
		Order(ent.Desc(webhookdelivery.FieldDeliveredAt)).
		Limit(limit).
		All(ctx)
}

// IsNotFound — сообщает об ошибках ent not-found.
func IsNotFound(err error) bool {
	return ent.IsNotFound(err)
}
