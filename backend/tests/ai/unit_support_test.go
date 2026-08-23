package ai_test

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/aiprovider"
	"github.com/valerii/hub/backend/internal/ai"
	githubmod "github.com/valerii/hub/backend/internal/github"
	gitlabmod "github.com/valerii/hub/backend/internal/gitlab"
)

type mockRepo struct {
	providers     []*ent.AIProvider
	owners        map[uuid.UUID]uuid.UUID
	requestCount  int
	allowExternal map[uuid.UUID]bool
	preferences   map[uuid.UUID]*ai.UserPreferenceRecord
}

func (r *mockRepo) ListProviders(_ context.Context, scope string) ([]*ent.AIProvider, error) {
	items := make([]*ent.AIProvider, 0, len(r.providers))
	for _, provider := range r.providers {
		if scope == "" || string(provider.Scope) == scope {
			items = append(items, provider)
		}
	}
	return items, nil
}

func (r *mockRepo) GetProvider(_ context.Context, id uuid.UUID) (*ent.AIProvider, error) {
	for _, provider := range r.providers {
		if provider.ID == id {
			return provider, nil
		}
	}
	return nil, &ent.NotFoundError{}
}

func (r *mockRepo) CreateProvider(_ context.Context, input ai.CreateProviderInput) (*ent.AIProvider, error) {
	provider := &ent.AIProvider{
		ID:           uuid.New(),
		Name:         input.Name,
		ProviderType: aiprovider.ProviderType(input.ProviderType),
		BaseURL:      input.BaseURL,
		APIKeyEnc:    input.APIKey,
		Models:       append([]string(nil), input.Models...),
		Enabled:      input.Enabled,
		Priority:     input.Priority,
		MaxRpm:       input.MaxRPM,
		RuResident:   input.RuResident,
		Scope:        aiprovider.Scope(input.Scope),
		Version:      1,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	r.providers = append(r.providers, provider)
	if input.OwnerUserID != nil {
		if r.owners == nil {
			r.owners = make(map[uuid.UUID]uuid.UUID)
		}
		r.owners[provider.ID] = *input.OwnerUserID
	}
	return provider, nil
}

func (r *mockRepo) UpdateProvider(ctx context.Context, id uuid.UUID, input ai.UpdateProviderInput) (*ent.AIProvider, error) {
	provider, err := r.GetProvider(ctx, id)
	if err != nil {
		return nil, err
	}
	if input.Name != nil {
		provider.Name = *input.Name
	}
	if input.BaseURL != nil {
		provider.BaseURL = *input.BaseURL
	}
	if input.APIKey != nil {
		provider.APIKeyEnc = *input.APIKey
	}
	if input.Models != nil {
		provider.Models = append([]string(nil), input.Models...)
	}
	if input.Enabled != nil {
		provider.Enabled = *input.Enabled
	}
	if input.Priority != nil {
		provider.Priority = *input.Priority
	}
	if input.MaxRPM != nil {
		provider.MaxRpm = *input.MaxRPM
	}
	if input.RuResident != nil {
		provider.RuResident = *input.RuResident
	}
	provider.Version++
	provider.UpdatedAt = time.Now().UTC()
	return provider, nil
}

func (r *mockRepo) FindProviderForModel(ctx context.Context, userID uuid.UUID, model string) (*ent.AIProvider, error) {
	items, err := r.ListProvidersForModel(ctx, userID, model)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, &ent.NotFoundError{}
	}
	return items[0], nil
}

func (r *mockRepo) ListProvidersForModel(_ context.Context, _ uuid.UUID, model string) ([]*ent.AIProvider, error) {
	items := make([]*ent.AIProvider, 0, len(r.providers))
	for _, provider := range r.providers {
		if !provider.Enabled || !mockProviderHasModel(provider, model) {
			continue
		}
		items = append(items, provider)
	}
	return items, nil
}

func mockProviderHasModel(provider *ent.AIProvider, model string) bool {
	for _, candidate := range provider.Models {
		if candidate == model {
			return true
		}
	}
	return false
}

func (r *mockRepo) IncrementUsage(_ context.Context, _ uuid.UUID, _, _ int) error {
	r.requestCount++
	return nil
}

func (r *mockRepo) ListModels(_ context.Context, _ *uuid.UUID) ([]ai.ModelInfo, error) {
	items := make([]ai.ModelInfo, 0)
	for _, provider := range r.providers {
		if !provider.Enabled {
			continue
		}
		for _, model := range provider.Models {
			items = append(items, ai.ModelInfo{
				ID: model, ProviderID: provider.ID.String(), ProviderName: provider.Name, RuResident: provider.RuResident,
			})
		}
	}
	return items, nil
}

func (r *mockRepo) ListUserProviders(_ context.Context, userID uuid.UUID) ([]*ent.AIProvider, error) {
	items := make([]*ent.AIProvider, 0)
	for _, provider := range r.providers {
		if string(provider.Scope) == "user" && r.owners[provider.ID] == userID {
			items = append(items, provider)
		}
	}
	return items, nil
}

func (r *mockRepo) GetUserProvider(ctx context.Context, userID, providerID uuid.UUID) (*ent.AIProvider, error) {
	provider, err := r.GetProvider(ctx, providerID)
	if err != nil || string(provider.Scope) != "user" || r.owners[providerID] != userID {
		return nil, &ent.NotFoundError{}
	}
	return provider, nil
}

func (r *mockRepo) GetDailyUsage(_ context.Context, _ uuid.UUID) (int, int, int, error) {
	return 0, 0, r.requestCount, nil
}

func (r *mockRepo) GetUserPreference(_ context.Context, userID uuid.UUID) (string, *uuid.UUID, error) {
	preference := r.preference(userID)
	if preference == nil {
		return "", nil, nil
	}
	return preference.DefaultModel, preference.DefaultProviderID, nil
}

func (r *mockRepo) GetUserPreferenceRecord(_ context.Context, userID uuid.UUID) (*ai.UserPreferenceRecord, error) {
	if preference := r.preference(userID); preference != nil {
		copy := *preference
		return &copy, nil
	}
	if r.allowExternal == nil {
		return nil, nil
	}
	allowed, exists := r.allowExternal[userID]
	if !exists {
		return nil, nil
	}
	return &ai.UserPreferenceRecord{AllowExternalAI: allowed}, nil
}

func (r *mockRepo) SaveUserPreference(_ context.Context, userID uuid.UUID, input ai.SaveUserPreferenceInput) error {
	if r.preferences == nil {
		r.preferences = make(map[uuid.UUID]*ai.UserPreferenceRecord)
	}
	preference := r.preference(userID)
	if preference == nil {
		preference = &ai.UserPreferenceRecord{}
		r.preferences[userID] = preference
	}
	if input.DefaultModel != nil {
		preference.DefaultModel = *input.DefaultModel
	}
	if input.DefaultProviderID != nil {
		preference.DefaultProviderID = input.DefaultProviderID
	}
	if input.AllowExternalAI != nil {
		preference.AllowExternalAI = *input.AllowExternalAI
	}
	preference.Version++
	return nil
}

func (r *mockRepo) preference(userID uuid.UUID) *ai.UserPreferenceRecord {
	if r.preferences == nil {
		return nil
	}
	return r.preferences[userID]
}

func (*mockRepo) GetUsageSummary(_ context.Context, _ time.Time) (int, int, int, int, error) {
	return 0, 0, 0, 0, nil
}

func (*mockRepo) ListUsageHistory(_ context.Context, _ int) ([]ai.UsageHistoryDay, error) {
	return nil, nil
}

type mockAdapter struct {
	resp *ai.ChatCompletionResponse
	err  error
}

func (a *mockAdapter) ChatCompletion(_ context.Context, _, _, _ string, _ ai.ChatCompletionRequest) (*ai.ChatCompletionResponse, error) {
	if a.err != nil {
		return nil, a.err
	}
	if a.resp != nil {
		return a.resp, nil
	}
	return &ai.ChatCompletionResponse{}, nil
}

func (a *mockAdapter) ChatCompletionStream(_ context.Context, _, _, _ string, _ ai.ChatCompletionRequest, onDelta func(ai.StreamDelta) error) error {
	if a.err != nil {
		return a.err
	}
	return onDelta(ai.StreamDelta{Done: true})
}

type memSettings struct {
	data map[string]string
}

func (s *memSettings) Get(_ context.Context, key string) (string, error) {
	return s.data[key], nil
}

func (s *memSettings) Upsert(_ context.Context, key, value, _ string) error {
	if s.data == nil {
		s.data = make(map[string]string)
	}
	s.data[key] = value
	return nil
}

type githubReviewRepo struct {
	integrations map[string]*ent.GitHubIntegration
}

func (*githubReviewRepo) CreateIntegration(context.Context, uuid.UUID, githubmod.CreateIntegrationInput) (*ent.GitHubIntegration, error) {
	return nil, &ent.NotFoundError{}
}
func (*githubReviewRepo) ListIntegrations(context.Context, uuid.UUID) ([]*ent.GitHubIntegration, error) {
	return nil, nil
}
func (*githubReviewRepo) GetIntegration(context.Context, uuid.UUID) (*ent.GitHubIntegration, error) {
	return nil, &ent.NotFoundError{}
}
func (*githubReviewRepo) GetIntegrationInProject(context.Context, uuid.UUID, uuid.UUID) (*ent.GitHubIntegration, error) {
	return nil, &ent.NotFoundError{}
}
func (*githubReviewRepo) DeleteIntegration(context.Context, uuid.UUID, uuid.UUID) error { return nil }
func (r *githubReviewRepo) FindIntegrationByRepo(_ context.Context, owner, repo string) (*ent.GitHubIntegration, error) {
	if integration := r.integrations[owner+"/"+repo]; integration != nil {
		return integration, nil
	}
	return nil, &ent.NotFoundError{}
}
func (*githubReviewRepo) FindIssueByKeyInProject(context.Context, uuid.UUID, string) (*ent.Issue, error) {
	return nil, &ent.NotFoundError{}
}
func (*githubReviewRepo) CreateGitHubLink(context.Context, uuid.UUID, uuid.UUID, githubmod.CreateGitHubLinkInput) (*ent.GitHubLink, error) {
	return nil, &ent.NotFoundError{}
}
func (*githubReviewRepo) FindGitHubLink(context.Context, string, string, uuid.UUID) (*ent.GitHubLink, error) {
	return nil, &ent.NotFoundError{}
}
func (*githubReviewRepo) UpdateGitHubLinkStatus(context.Context, uuid.UUID, string) error { return nil }
func (*githubReviewRepo) ListGitHubLinks(context.Context, uuid.UUID) ([]*ent.GitHubLink, error) {
	return nil, nil
}
func (*githubReviewRepo) FindIssueByKey(context.Context, string) (*ent.Issue, error) {
	return nil, &ent.NotFoundError{}
}
func (*githubReviewRepo) UpdateIssueStatus(context.Context, uuid.UUID, uuid.UUID) error { return nil }
func (*githubReviewRepo) FindDoneStatus(context.Context, uuid.UUID) (*ent.WorkflowStatus, error) {
	return nil, &ent.NotFoundError{}
}
func (*githubReviewRepo) LogActivity(context.Context, uuid.UUID, string) error { return nil }
func (*githubReviewRepo) IsProjectMember(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return false, nil
}
func (*githubReviewRepo) FindTestCaseByKey(context.Context, string) (*ent.TestCase, error) {
	return nil, &ent.NotFoundError{}
}
func (*githubReviewRepo) MarkTestCasesOutdatedByFiles(context.Context, []string) ([]string, error) {
	return nil, nil
}

type gitlabReviewRepo struct {
	integrations map[string]*ent.GitLabIntegration
}

func (*gitlabReviewRepo) CreateIntegration(context.Context, uuid.UUID, gitlabmod.CreateIntegrationInput) (*ent.GitLabIntegration, error) {
	return nil, &ent.NotFoundError{}
}
func (*gitlabReviewRepo) ListIntegrations(context.Context, uuid.UUID) ([]*ent.GitLabIntegration, error) {
	return nil, nil
}
func (*gitlabReviewRepo) GetIntegration(context.Context, uuid.UUID) (*ent.GitLabIntegration, error) {
	return nil, &ent.NotFoundError{}
}
func (*gitlabReviewRepo) GetIntegrationInProject(context.Context, uuid.UUID, uuid.UUID) (*ent.GitLabIntegration, error) {
	return nil, &ent.NotFoundError{}
}
func (*gitlabReviewRepo) DeleteIntegration(context.Context, uuid.UUID, uuid.UUID) error { return nil }
func (r *gitlabReviewRepo) FindIntegrationByProjectPath(_ context.Context, projectPath string) (*ent.GitLabIntegration, error) {
	if integration := r.integrations[projectPath]; integration != nil {
		return integration, nil
	}
	return nil, &ent.NotFoundError{}
}
func (*gitlabReviewRepo) FindIssueByKey(context.Context, string) (*ent.Issue, error) {
	return nil, &ent.NotFoundError{}
}
func (*gitlabReviewRepo) FindIssueByKeyInProject(context.Context, uuid.UUID, string) (*ent.Issue, error) {
	return nil, &ent.NotFoundError{}
}

var (
	_ ai.Repository         = (*mockRepo)(nil)
	_ ai.ChatAdapter        = (*mockAdapter)(nil)
	_ ai.SettingsStore      = (*memSettings)(nil)
	_ githubmod.Repository = (*githubReviewRepo)(nil)
	_ gitlabmod.Repository = (*gitlabReviewRepo)(nil)
)
