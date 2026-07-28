package apigateway

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/apikey"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/ent/webhookdelivery"
	"github.com/valerii/hub/backend/ent/webhookendpoint"
	cryptopkg "github.com/valerii/hub/backend/internal/platform/crypto"
)

// ErrEncryptionRequired — cipher обязателен для записи непустого webhook secret (W4-A5-008).
var ErrEncryptionRequired = errors.New("encryption required: webhook secret cannot be stored without cipher")

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
	cipher *cryptopkg.Cipher
}

// NewEntRepository создаёт Ent-репозиторий apigateway.
// cipher шифрует webhook secret at-rest (W4-A5-008); nil допустим только если secrets не пишутся.
func NewEntRepository(client *ent.Client, cipher *cryptopkg.Cipher) *EntRepository {
	return &EntRepository{client: client, cipher: cipher}
}

func (r *EntRepository) encryptSecret(v string) (string, error) {
	if v == "" {
		return "", nil
	}
	if r.cipher == nil {
		return "", ErrEncryptionRequired
	}
	return r.cipher.EncryptString(v)
}

func (r *EntRepository) decryptSecret(v string) (string, error) {
	if v == "" {
		return "", nil
	}
	if r.cipher == nil {
		return "", ErrEncryptionRequired
	}
	pt, ok := r.cipher.MaybeDecryptString(v)
	if ok {
		return pt, nil
	}
	// Legacy plaintext до W4-A5-008.
	return v, nil
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
		Where(
			apikey.HasUserWith(user.ID(userID)),
			apikey.DeletedAtIsNil(),
		).
		All(ctx)
}

func (r *EntRepository) CountAPIKeyForUser(ctx context.Context, userID, keyID uuid.UUID) (int, error) {
	return r.client.APIKey.Query().
		Where(
			apikey.ID(keyID),
			apikey.DeletedAtIsNil(),
			apikey.HasUserWith(user.ID(userID)),
		).
		Count(ctx)
}

// DeleteAPIKey мягко отзывает ключ (корзина 30д, W4-A5-007).
func (r *EntRepository) DeleteAPIKey(ctx context.Context, keyID uuid.UUID) error {
	affected, err := r.client.APIKey.Update().
		Where(apikey.IDEQ(keyID), apikey.DeletedAtIsNil()).
		SetDeletedAt(time.Now().UTC()).
		Save(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return &ent.NotFoundError{}
	}
	return nil
}

func (r *EntRepository) GetAPIKeyByHashWithUser(ctx context.Context, hash string) (*ent.APIKey, error) {
	return r.client.APIKey.Query().
		Where(apikey.KeyHash(hash), apikey.DeletedAtIsNil()).
		WithUser().
		Only(ctx)
}

func (r *EntRepository) UpdateAPIKeyLastUsedAt(ctx context.Context, keyID uuid.UUID, when time.Time) error {
	_, err := r.client.APIKey.Update().
		Where(apikey.IDEQ(keyID), apikey.DeletedAtIsNil()).
		SetLastUsedAt(when).
		Save(ctx)
	return err
}

func (r *EntRepository) ListActiveWebhookEndpoints(ctx context.Context) ([]*ent.WebhookEndpoint, error) {
	// WithProject — нужен dispatchEvent для фильтрации по проекту эндпоинта (TASK-012).
	eps, err := r.client.WebhookEndpoint.Query().
		Where(
			webhookendpoint.IsActive(true),
			webhookendpoint.DeletedAtIsNil(),
		).
		WithProject().
		All(ctx)
	if err != nil {
		return nil, err
	}
	for _, ep := range eps {
		plain, err := r.decryptSecret(ep.Secret)
		if err != nil {
			return nil, err
		}
		ep.Secret = plain
	}
	return eps, nil
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
	enc, err := r.encryptSecret(secret)
	if err != nil {
		return nil, err
	}
	return r.client.WebhookEndpoint.Create().
		SetURL(url).
		SetSecret(enc).
		SetEvents(events).
		SetProjectID(projectID).
		Save(ctx)
}

func (r *EntRepository) ListWebhookEndpointsByProject(ctx context.Context, projectID uuid.UUID) ([]*ent.WebhookEndpoint, error) {
	return r.client.WebhookEndpoint.Query().
		Where(
			webhookendpoint.HasProjectWith(project.ID(projectID)),
			webhookendpoint.DeletedAtIsNil(),
		).
		All(ctx)
}

func (r *EntRepository) UpdateWebhookEndpoint(ctx context.Context, endpointID uuid.UUID, input UpdateWebhookEndpointInput) (*ent.WebhookEndpoint, error) {
	b := r.client.WebhookEndpoint.Update().
		Where(webhookendpoint.IDEQ(endpointID), webhookendpoint.DeletedAtIsNil())
	if input.URL != nil {
		b.SetURL(*input.URL)
	}
	if input.Secret != nil {
		enc, err := r.encryptSecret(*input.Secret)
		if err != nil {
			return nil, err
		}
		b.SetSecret(enc)
	}
	if input.Events != nil {
		b.SetEvents(input.Events)
	}
	if input.IsActive != nil {
		b.SetIsActive(*input.IsActive)
	}
	affected, err := b.Save(ctx)
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, &ent.NotFoundError{}
	}
	return r.client.WebhookEndpoint.Query().
		Where(webhookendpoint.IDEQ(endpointID), webhookendpoint.DeletedAtIsNil()).
		Only(ctx)
}

// DeleteWebhookEndpoint мягко удаляет эндпоинт (корзина 30д, W4-A5-007).
func (r *EntRepository) DeleteWebhookEndpoint(ctx context.Context, endpointID uuid.UUID) error {
	affected, err := r.client.WebhookEndpoint.Update().
		Where(webhookendpoint.IDEQ(endpointID), webhookendpoint.DeletedAtIsNil()).
		SetDeletedAt(time.Now().UTC()).
		SetIsActive(false).
		Save(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return &ent.NotFoundError{}
	}
	return nil
}

func (r *EntRepository) WebhookEndpointProjectID(ctx context.Context, endpointID uuid.UUID) (uuid.UUID, error) {
	return r.client.WebhookEndpoint.Query().
		Where(webhookendpoint.ID(endpointID), webhookendpoint.DeletedAtIsNil()).
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
