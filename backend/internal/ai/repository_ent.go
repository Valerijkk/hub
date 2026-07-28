package ai

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/aiprovider"
	"github.com/valerii/hub/backend/ent/aiusagedaily"
	"github.com/valerii/hub/backend/ent/aiuserpreference"
	"github.com/valerii/hub/backend/ent/user"
	cryptopkg "github.com/valerii/hub/backend/internal/platform/crypto"
	safehttp "github.com/valerii/hub/backend/internal/platform/http"
)

// EntRepository — реализация Repository через Ent ORM.
type EntRepository struct {
	client *ent.Client
	cipher *cryptopkg.Cipher
}

// NewEntRepository создаёт Ent-репозиторий AI.
func NewEntRepository(client *ent.Client, cipher *cryptopkg.Cipher) *EntRepository {
	return &EntRepository{client: client, cipher: cipher}
}

// ErrEncryptionRequired — cipher обязателен для записи/чтения непустого API-ключа (W5-A3-007).
var ErrEncryptionRequired = errors.New("encryption required: api key cannot be stored without cipher")

func (r *EntRepository) encryptField(v string) (string, error) {
	if v == "" {
		return "", nil
	}
	if r.cipher == nil {
		return "", ErrEncryptionRequired
	}
	return r.cipher.EncryptString(v)
}

func (r *EntRepository) decryptField(v string) (string, error) {
	if v == "" {
		return "", nil
	}
	if r.cipher == nil {
		return "", ErrEncryptionRequired
	}
	pt, err := r.cipher.DecryptString(v)
	if err != nil {
		return "", fmt.Errorf("decrypt api key: %w", err)
	}
	return pt, nil
}

func (r *EntRepository) ListProviders(ctx context.Context, scope string) ([]*ent.AIProvider, error) {
	q := r.client.AIProvider.Query().Order(ent.Asc(aiprovider.FieldPriority))
	if scope != "" {
		q = q.Where(aiprovider.ScopeEQ(aiprovider.Scope(scope)))
	}
	return q.All(ctx)
}

func (r *EntRepository) GetProvider(ctx context.Context, id uuid.UUID) (*ent.AIProvider, error) {
	p, err := r.client.AIProvider.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	key, err := r.decryptField(p.APIKeyEnc)
	if err != nil {
		return nil, err
	}
	p.APIKeyEnc = key
	return p, nil
}

func (r *EntRepository) CreateProvider(ctx context.Context, input CreateProviderInput) (*ent.AIProvider, error) {
	if err := safehttp.ValidateOutboundURL(strings.TrimSpace(input.BaseURL)); err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}
	keyEnc, err := r.encryptField(strings.TrimSpace(input.APIKey))
	if err != nil {
		return nil, fmt.Errorf("encrypt api key: %w", err)
	}
	scope := aiprovider.ScopeSystem
	if input.Scope == "user" {
		scope = aiprovider.ScopeUser
	}
	b := r.client.AIProvider.Create().
		SetName(strings.TrimSpace(input.Name)).
		SetProviderType(aiprovider.ProviderType(input.ProviderType)).
		SetBaseURL(strings.TrimRight(strings.TrimSpace(input.BaseURL), "/")).
		SetModels(normalizeModels(input.Models)).
		SetEnabled(input.Enabled).
		SetPriority(input.Priority).
		SetMaxRpm(input.MaxRPM).
		SetRuResident(input.RuResident).
		SetScope(scope)
	if keyEnc != "" {
		b = b.SetAPIKeyEnc(keyEnc)
	}
	if input.OwnerUserID != nil {
		b = b.SetOwnerID(*input.OwnerUserID)
	}
	return b.Save(ctx)
}

func (r *EntRepository) UpdateProvider(ctx context.Context, id uuid.UUID, input UpdateProviderInput) (*ent.AIProvider, error) {
	b := r.client.AIProvider.UpdateOneID(id)
	if input.Name != nil {
		b.SetName(strings.TrimSpace(*input.Name))
	}
	if input.BaseURL != nil {
		url := strings.TrimRight(strings.TrimSpace(*input.BaseURL), "/")
		if err := safehttp.ValidateOutboundURL(url); err != nil {
			return nil, fmt.Errorf("invalid base URL: %w", err)
		}
		b.SetBaseURL(url)
	}
	if input.APIKey != nil {
		keyEnc, err := r.encryptField(strings.TrimSpace(*input.APIKey))
		if err != nil {
			return nil, fmt.Errorf("encrypt api key: %w", err)
		}
		if keyEnc == "" {
			b.ClearAPIKeyEnc()
		} else {
			b.SetAPIKeyEnc(keyEnc)
		}
	}
	if input.Models != nil {
		b.SetModels(normalizeModels(input.Models))
	}
	if input.Enabled != nil {
		b.SetEnabled(*input.Enabled)
	}
	if input.Priority != nil {
		b.SetPriority(*input.Priority)
	}
	if input.MaxRPM != nil {
		b.SetMaxRpm(*input.MaxRPM)
	}
	if input.RuResident != nil {
		b.SetRuResident(*input.RuResident)
	}
	return b.Save(ctx)
}

func (r *EntRepository) FindProviderForModel(ctx context.Context, userID uuid.UUID, model string) (*ent.AIProvider, error) {
	list, err := r.ListProvidersForModel(ctx, userID, model)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, &ent.NotFoundError{}
	}
	return list[0], nil
}

func (r *EntRepository) ListProvidersForModel(ctx context.Context, userID uuid.UUID, model string) ([]*ent.AIProvider, error) {
	model = strings.TrimSpace(model)
	var out []*ent.AIProvider
	if userID != uuid.Nil {
		userProviders, err := r.client.AIProvider.Query().
			Where(
				aiprovider.EnabledEQ(true),
				aiprovider.ScopeEQ(aiprovider.ScopeUser),
				aiprovider.HasOwnerWith(user.IDEQ(userID)),
			).
			Order(ent.Asc(aiprovider.FieldPriority)).
			All(ctx)
		if err != nil {
			return nil, err
		}
		for _, p := range userProviders {
			if providerHasModel(p, model) {
				key, err := r.decryptField(p.APIKeyEnc)
				if err != nil {
					return nil, err
				}
				cp := *p
				cp.APIKeyEnc = key
				out = append(out, &cp)
			}
		}
	}
	systemProviders, err := r.client.AIProvider.Query().
		Where(aiprovider.EnabledEQ(true), aiprovider.ScopeEQ(aiprovider.ScopeSystem)).
		Order(ent.Asc(aiprovider.FieldPriority)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	for _, p := range systemProviders {
		if providerHasModel(p, model) {
			key, err := r.decryptField(p.APIKeyEnc)
			if err != nil {
				return nil, err
			}
			cp := *p
			cp.APIKeyEnc = key
			out = append(out, &cp)
		}
	}
	if len(out) == 0 {
		return nil, &ent.NotFoundError{}
	}
	return out, nil
}

func providerHasModel(p *ent.AIProvider, model string) bool {
	for _, m := range p.Models {
		if m == model {
			return true
		}
	}
	return false
}

func (r *EntRepository) ListUserProviders(ctx context.Context, userID uuid.UUID) ([]*ent.AIProvider, error) {
	return r.client.AIProvider.Query().
		Where(
			aiprovider.ScopeEQ(aiprovider.ScopeUser),
			aiprovider.HasOwnerWith(user.IDEQ(userID)),
		).
		Order(ent.Asc(aiprovider.FieldPriority)).
		All(ctx)
}

func (r *EntRepository) GetUserProvider(ctx context.Context, userID, providerID uuid.UUID) (*ent.AIProvider, error) {
	p, err := r.client.AIProvider.Query().
		Where(
			aiprovider.IDEQ(providerID),
			aiprovider.ScopeEQ(aiprovider.ScopeUser),
			aiprovider.HasOwnerWith(user.IDEQ(userID)),
		).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	p.APIKeyEnc = "" // никогда не отдаём ключ при чтении
	return p, nil
}

func (r *EntRepository) IncrementUsage(ctx context.Context, userID uuid.UUID, tokensIn, tokensOut int) error {
	day := utcDate(time.Now())
	existing, err := r.client.AIUsageDaily.Query().
		Where(
			aiusagedaily.HasUserWith(user.IDEQ(userID)),
			aiusagedaily.UsageDateEQ(day),
		).
		Only(ctx)
	if ent.IsNotFound(err) {
		_, err = r.client.AIUsageDaily.Create().
			SetUserID(userID).
			SetUsageDate(day).
			SetTokensIn(tokensIn).
			SetTokensOut(tokensOut).
			SetRequestCount(1).
			Save(ctx)
		return err
	}
	if err != nil {
		return err
	}
	return r.client.AIUsageDaily.UpdateOne(existing).
		AddTokensIn(tokensIn).
		AddTokensOut(tokensOut).
		AddRequestCount(1).
		Exec(ctx)
}

func (r *EntRepository) ListModels(ctx context.Context, userID *uuid.UUID) ([]ModelInfo, error) {
	var providers []*ent.AIProvider
	system, err := r.ListProviders(ctx, "system")
	if err != nil {
		return nil, err
	}
	providers = append(providers, system...)
	if userID != nil && *userID != uuid.Nil {
		userProviders, uErr := r.ListUserProviders(ctx, *userID)
		if uErr != nil {
			return nil, uErr
		}
		providers = append(providers, userProviders...)
	}
	var out []ModelInfo
	for _, p := range providers {
		if !p.Enabled {
			continue
		}
		for _, m := range p.Models {
			out = append(out, ModelInfo{
				ID:           m,
				ProviderID:   p.ID.String(),
				ProviderName: p.Name,
				RuResident:   p.RuResident,
			})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].ProviderName == out[j].ProviderName {
			return out[i].ID < out[j].ID
		}
		return out[i].ProviderName < out[j].ProviderName
	})
	return out, nil
}

func (r *EntRepository) GetDailyUsage(ctx context.Context, userID uuid.UUID) (int, int, int, error) {
	day := utcDate(time.Now())
	row, err := r.client.AIUsageDaily.Query().
		Where(
			aiusagedaily.HasUserWith(user.IDEQ(userID)),
			aiusagedaily.UsageDateEQ(day),
		).
		Only(ctx)
	if ent.IsNotFound(err) {
		return 0, 0, 0, nil
	}
	if err != nil {
		return 0, 0, 0, err
	}
	return row.TokensIn, row.TokensOut, row.RequestCount, nil
}

func (r *EntRepository) GetUserPreference(ctx context.Context, userID uuid.UUID) (string, *uuid.UUID, error) {
	rec, err := r.GetUserPreferenceRecord(ctx, userID)
	if err != nil || rec == nil {
		return "", nil, err
	}
	return rec.DefaultModel, rec.DefaultProviderID, nil
}

func (r *EntRepository) GetUserPreferenceRecord(ctx context.Context, userID uuid.UUID) (*UserPreferenceRecord, error) {
	pref, err := r.client.AIUserPreference.Query().
		Where(aiuserpreference.HasUserWith(user.IDEQ(userID))).
		Only(ctx)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &UserPreferenceRecord{
		DefaultModel:      pref.DefaultModel,
		DefaultProviderID: pref.DefaultProviderID,
		AllowExternalAI:   pref.AllowExternalAi,
	}, nil
}

func (r *EntRepository) SaveUserPreference(ctx context.Context, userID uuid.UUID, input SaveUserPreferenceInput) error {
	existing, err := r.client.AIUserPreference.Query().
		Where(aiuserpreference.HasUserWith(user.IDEQ(userID))).
		Only(ctx)
	if ent.IsNotFound(err) {
		b := r.client.AIUserPreference.Create().SetUserID(userID)
		if input.DefaultModel != nil {
			if *input.DefaultModel != "" {
				b = b.SetDefaultModel(*input.DefaultModel)
			}
		}
		if input.DefaultProviderID != nil {
			b = b.SetDefaultProviderID(*input.DefaultProviderID)
		}
		if input.AllowExternalAI != nil {
			b = b.SetAllowExternalAi(*input.AllowExternalAI)
		}
		_, err = b.Save(ctx)
		return err
	}
	if err != nil {
		return err
	}
	u := r.client.AIUserPreference.UpdateOne(existing)
	if input.DefaultModel != nil {
		if *input.DefaultModel == "" {
			u.ClearDefaultModel()
		} else {
			u.SetDefaultModel(*input.DefaultModel)
		}
	}
	if input.DefaultProviderID != nil {
		if *input.DefaultProviderID == uuid.Nil {
			u.ClearDefaultProviderID()
		} else {
			u.SetDefaultProviderID(*input.DefaultProviderID)
		}
	}
	if input.AllowExternalAI != nil {
		u.SetAllowExternalAi(*input.AllowExternalAI)
	}
	_, err = u.Save(ctx)
	return err
}

func (r *EntRepository) GetUsageSummary(ctx context.Context, day time.Time) (int, int, int, int, error) {
	day = utcDate(day)
	rows, err := r.client.AIUsageDaily.Query().
		Where(aiusagedaily.UsageDateEQ(day)).
		All(ctx)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	var req, in, out int
	for _, row := range rows {
		req += row.RequestCount
		in += row.TokensIn
		out += row.TokensOut
	}
	return req, in, out, len(rows), nil
}

func (r *EntRepository) ListUsageHistory(ctx context.Context, days int) ([]UsageHistoryDay, error) {
	if days <= 0 {
		days = 7
	}
	if days > 90 {
		days = 90
	}
	since := utcDate(time.Now().UTC().AddDate(0, 0, -(days - 1)))
	rows, err := r.client.AIUsageDaily.Query().
		Where(aiusagedaily.UsageDateGTE(since)).
		Order(ent.Asc(aiusagedaily.FieldUsageDate)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	type agg struct {
		req, in, out, users int
	}
	byDay := make(map[string]*agg)
	for _, row := range rows {
		key := row.UsageDate.UTC().Format("2006-01-02")
		a, ok := byDay[key]
		if !ok {
			a = &agg{}
			byDay[key] = a
		}
		a.req += row.RequestCount
		a.in += row.TokensIn
		a.out += row.TokensOut
		a.users++
	}
	out := make([]UsageHistoryDay, 0, len(byDay))
	for d, a := range byDay {
		out = append(out, UsageHistoryDay{
			Date:        d,
			Requests:    a.req,
			TokensIn:    a.in,
			TokensOut:   a.out,
			ActiveUsers: a.users,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Date < out[j].Date })
	return out, nil
}

func normalizeModels(models []string) []string {
	if len(models) == 0 {
		return []string{}
	}
	seen := make(map[string]struct{}, len(models))
	out := make([]string, 0, len(models))
	for _, m := range models {
		m = strings.TrimSpace(m)
		if m == "" {
			continue
		}
		if _, ok := seen[m]; ok {
			continue
		}
		seen[m] = struct{}{}
		out = append(out, m)
	}
	sort.Strings(out)
	return out
}

func utcDate(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
