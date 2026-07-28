package search_test
import (
	"github.com/valerii/hub/backend/internal/search"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
)

// statefulSearchRepo хранит сохранённые фильтры для CRUD-тестов сервиса.
type statefulSearchRepo struct {
	mockSearchRepo
	filters map[uuid.UUID]*ent.SavedFilter
}

func newStatefulSearchRepo() *statefulSearchRepo {
	return &statefulSearchRepo{filters: make(map[uuid.UUID]*ent.SavedFilter)}
}

func (m *statefulSearchRepo) CreateSavedFilter(_ context.Context, input search.CreateSavedFilterInput) (*ent.SavedFilter, error) {
	now := time.Now().UTC()
	f := &ent.SavedFilter{
		ID:           uuid.New(),
		Name:         input.Name,
		FilterConfig: input.FilterConfig,
		OwnerID:      input.OwnerID,
		ProjectID:    input.ProjectID,
		Shared:       false,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	m.filters[f.ID] = f
	return f, nil
}

func (m *statefulSearchRepo) ListSavedFilters(_ context.Context, userID, projectID uuid.UUID) ([]*ent.SavedFilter, error) {
	var result []*ent.SavedFilter
	for _, f := range m.filters {
		if f.ProjectID != projectID {
			continue
		}
		if f.OwnerID == userID || f.Shared {
			result = append(result, f)
		}
	}
	return result, nil
}

func (m *statefulSearchRepo) GetSavedFilter(_ context.Context, filterID uuid.UUID) (*ent.SavedFilter, error) {
	f, ok := m.filters[filterID]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	return f, nil
}

func (m *statefulSearchRepo) DeleteSavedFilter(_ context.Context, filterID uuid.UUID) error {
	if _, ok := m.filters[filterID]; !ok {
		return &ent.NotFoundError{}
	}
	delete(m.filters, filterID)
	return nil
}

func (m *statefulSearchRepo) UpdateSavedFilterShared(_ context.Context, filterID uuid.UUID, shared bool) (*ent.SavedFilter, error) {
	f, ok := m.filters[filterID]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	f.Shared = shared
	f.UpdatedAt = time.Now().UTC()
	return f, nil
}

func newSearchServiceTestEnv(t *testing.T) (*search.SearchService, *statefulSearchRepo, context.Context, uuid.UUID, uuid.UUID) {
	t.Helper()
	repo := newStatefulSearchRepo()
	svc := search.NewSearchService(repo, newScopeTestDB(t), nil)
	userID := uuid.New()
	projectID := uuid.New()
	return svc, repo, context.Background(), userID, projectID
}

func TestSavedFilterToDTO(t *testing.T) {
	now := time.Now().UTC()
	ownerID := uuid.New()
	projectID := uuid.New()
	filterID := uuid.New()

	dto := search.ExportSavedFilterToDTO(&ent.SavedFilter{
		ID:           filterID,
		Name:         "My bugs",
		FilterConfig: map[string]interface{}{"status": "open"},
		Shared:       true,
		OwnerID:      ownerID,
		ProjectID:    projectID,
		CreatedAt:    now,
		UpdatedAt:    now,
	})

	require.NotNil(t, dto)
	assert.Equal(t, filterID.String(), dto.ID)
	assert.Equal(t, "My bugs", dto.Name)
	assert.True(t, dto.Shared)
	assert.Equal(t, ownerID.String(), dto.OwnerID)
	assert.Equal(t, projectID.String(), dto.ProjectID)
}

func TestCreateSavedFilter_InvalidProjectID(t *testing.T) {
	svc, _, ctx, userID, _ := newSearchServiceTestEnv(t)

	_, err := svc.CreateSavedFilter(ctx, userID, search.SavedFilterCreateRequest{
		Name:         "F",
		FilterConfig: map[string]interface{}{"q": "bug"},
		ProjectID:    "not-uuid",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "project_id")
}

// TestCreateSavedFilter_NonMemberForbidden — A4-003: CreateSavedFilter требует
// IsProjectMember до repo.CreateSavedFilter. Mutation guard: при снятии проверки
// stranger создаст filter (len>0) и тест упадёт; при живой проверке ErrFilterForbidden
// и 0 filters в repo.
func TestCreateSavedFilter_NonMemberForbidden(t *testing.T) {
	deny := false
	repo := newStatefulSearchRepo()
	repo.memberOK = &deny
	svc := search.NewSearchService(repo, newScopeTestDB(t), nil)
	ctx := context.Background()
	stranger := uuid.New()
	projectID := uuid.New()

	_, err := svc.CreateSavedFilter(ctx, stranger, search.SavedFilterCreateRequest{
		Name:         "Leak",
		FilterConfig: map[string]interface{}{"status": "open"},
		ProjectID:    projectID.String(),
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, search.ErrFilterForbidden)
	assert.Empty(t, repo.filters, "non-member must not create a saved filter (A4-003)")
}

// TestListSavedFilters_NonMemberForbidden — A4-004: ListSavedFilters требует
// IsProjectMember до repo.ListSavedFilters. Mutation guard: при снятии проверки
// stranger получит shared filter (len==1) и тест упадёт; при живой проверке
// ErrFilterForbidden без утечки filter_config.
func TestListSavedFilters_NonMemberForbidden(t *testing.T) {
	deny := false
	repo := newStatefulSearchRepo()
	repo.memberOK = &deny
	svc := search.NewSearchService(repo, newScopeTestDB(t), nil)
	ctx := context.Background()
	owner := uuid.New()
	stranger := uuid.New()
	projectID := uuid.New()
	now := time.Now().UTC()
	filterID := uuid.New()
	repo.filters[filterID] = &ent.SavedFilter{
		ID: filterID, Name: "Shared bugs", OwnerID: owner, ProjectID: projectID,
		Shared: true, FilterConfig: map[string]interface{}{"secret": "leak"},
		CreatedAt: now, UpdatedAt: now,
	}

	list, err := svc.ListSavedFilters(ctx, stranger, projectID)
	require.Error(t, err)
	assert.ErrorIs(t, err, search.ErrFilterForbidden)
	assert.Nil(t, list, "non-member must not see shared filters (A4-004)")
}

func TestSavedFilter_CRUDAndShare(t *testing.T) {
	svc, _, ctx, userID, projectID := newSearchServiceTestEnv(t)
	otherUser := uuid.New()

	created, err := svc.CreateSavedFilter(ctx, userID, search.SavedFilterCreateRequest{
		Name:         "Open issues",
		FilterConfig: map[string]interface{}{"status": "open"},
		ProjectID:    projectID.String(),
	})
	require.NoError(t, err)
	assert.Equal(t, "Open issues", created.Name)
	assert.False(t, created.Shared)

	list, err := svc.ListSavedFilters(ctx, userID, projectID)
	require.NoError(t, err)
	require.Len(t, list, 1)

	otherList, err := svc.ListSavedFilters(ctx, otherUser, projectID)
	require.NoError(t, err)
	assert.Empty(t, otherList)

	filterID := uuid.MustParse(created.ID)
	shared, err := svc.ShareFilter(ctx, userID, filterID, true)
	require.NoError(t, err)
	assert.True(t, shared.Shared)

	otherList, err = svc.ListSavedFilters(ctx, otherUser, projectID)
	require.NoError(t, err)
	assert.Len(t, otherList, 1)

	err = svc.DeleteSavedFilter(ctx, otherUser, filterID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "доступ запрещён")

	require.NoError(t, svc.DeleteSavedFilter(ctx, userID, filterID))
	list, err = svc.ListSavedFilters(ctx, userID, projectID)
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestShareFilter_NotOwnerForbidden(t *testing.T) {
	svc, repo, ctx, ownerID, projectID := newSearchServiceTestEnv(t)
	now := time.Now().UTC()
	filterID := uuid.New()
	repo.filters[filterID] = &ent.SavedFilter{
		ID: filterID, Name: "X", OwnerID: ownerID, ProjectID: projectID,
		FilterConfig: map[string]interface{}{}, CreatedAt: now, UpdatedAt: now,
	}

	_, err := svc.ShareFilter(ctx, uuid.New(), filterID, true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "доступ запрещён")
}

func TestSearch_NormalizesLimit(t *testing.T) {
	db := newScopeTestDB(t)
	userID := uuid.New()
	svc := search.NewSearchService(&mockSearchRepo{}, db, nil)

	for _, limit := range []int{0, -1, 100} {
		results, _, err := svc.Search(context.Background(), "hub", userID, nil, limit)
		require.NoError(t, err)
		assert.Empty(t, results)
	}
}

func TestSearch_FiltersByType(t *testing.T) {
	db := newScopeTestDB(t)
	userID := uuid.New()
	projectID := uuid.New()

	_, err := db.Exec(
		`INSERT INTO project_members (project_members, user_project_members) VALUES (?, ?)`,
		projectID.String(), userID.String(),
	)
	require.NoError(t, err)

	svc := search.NewSearchService(&mockSearchRepo{}, db, nil)

	results, _, err := svc.Search(context.Background(), "hub", userID, []string{"project"}, 5)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestSearch_AdminUsesUnscopedUserSearch(t *testing.T) {
	db := newScopeTestDB(t)
	userID := uuid.New()
	svc := search.NewSearchService(&mockSearchRepo{isAdmin: true}, db, nil)

	results, _, err := svc.Search(context.Background(), "admin", userID, []string{"user"}, 10)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestSearch_ContinuesOnPartialShardErrors(t *testing.T) {
	db := newScopeTestDB(t)
	userID := uuid.New()
	// Без project membership: issues shard = OK empty; user shard fail → partial.
	repo := &mockSearchRepo{adminErr: errors.New("admin lookup failed")}
	svc := search.NewSearchService(repo, db, nil)

	results, partial, err := svc.Search(context.Background(), "hub", userID, []string{"issue", "user"}, 20)
	require.NoError(t, err)
	assert.True(t, partial, "user shard fail + issue ok → partial")
	assert.Empty(t, results)
}

func TestEntRepository_SavedFilterCRUD(t *testing.T) {
	dsn := "file:searchtest_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	owner, err := client.User.Create().SetEmail("search-" + uuid.New().String()[:8] + "@test.com").Save(ctx)
	require.NoError(t, err)
	proj, err := client.Project.Create().SetName("Search Proj").SetKey("SRC").Save(ctx)
	require.NoError(t, err)

	repo := search.NewEntRepository(client)
	cfg := map[string]interface{}{"priority": "high"}

	created, err := repo.CreateSavedFilter(ctx, search.CreateSavedFilterInput{
		Name: "High priority", FilterConfig: cfg, OwnerID: owner.ID, ProjectID: proj.ID,
	})
	require.NoError(t, err)
	assert.Equal(t, "High priority", created.Name)

	list, err := repo.ListSavedFilters(ctx, owner.ID, proj.ID)
	require.NoError(t, err)
	require.Len(t, list, 1)

	updated, err := repo.UpdateSavedFilterShared(ctx, created.ID, true)
	require.NoError(t, err)
	assert.True(t, updated.Shared)

	require.NoError(t, repo.DeleteSavedFilter(ctx, created.ID))
	list, err = repo.ListSavedFilters(ctx, owner.ID, proj.ID)
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestEntRepository_IsSystemAdmin(t *testing.T) {
	dsn := "file:searchadm_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	admin, err := client.User.Create().
		SetEmail("admin-" + uuid.New().String()[:8] + "@test.com").
		SetSystemRole("admin").
		Save(ctx)
	require.NoError(t, err)
	member, err := client.User.Create().
		SetEmail("member-" + uuid.New().String()[:8] + "@test.com").
		Save(ctx)
	require.NoError(t, err)

	repo := search.NewEntRepository(client)
	ok, err := repo.IsSystemAdmin(ctx, admin.ID)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = repo.IsSystemAdmin(ctx, member.ID)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestEntRepository_IsProjectMember(t *testing.T) {
	dsn := "file:searchpm_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	member, err := client.User.Create().SetEmail("pm-" + uuid.New().String()[:8] + "@test.com").Save(ctx)
	require.NoError(t, err)
	stranger, err := client.User.Create().SetEmail("str-" + uuid.New().String()[:8] + "@test.com").Save(ctx)
	require.NoError(t, err)
	proj, err := client.Project.Create().SetName("PM Proj").SetKey("PMP").Save(ctx)
	require.NoError(t, err)
	_, err = client.ProjectMember.Create().
		SetUserID(member.ID).
		SetProjectID(proj.ID).
		SetProjectRole("developer").
		Save(ctx)
	require.NoError(t, err)

	repo := search.NewEntRepository(client)
	ok, err := repo.IsProjectMember(ctx, member.ID, proj.ID)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = repo.IsProjectMember(ctx, stranger.ID, proj.ID)
	require.NoError(t, err)
	assert.False(t, ok)
}
