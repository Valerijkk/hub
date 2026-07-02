package audit_test

// Тесты сервиса аудита: интеграционные и unit-сценарии testframework.
import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/audit"
)

// TestRepository_GetSystemSetting проверяет слой репозитория: репозиторий получение system setting.
func TestRepository_GetSystemSetting(t *testing.T) {
	repo := newAuditRepoTestEnv(t)
	ctx := context.Background()

	val, err := repo.GetSystemSetting(ctx, "missing.key")
	require.NoError(t, err)
	assert.Equal(t, "", val)

	_, err = audit.ExportRepositoryEntClient(repo).SystemSetting.Create().
		SetKey("limit.test").
		SetValue("42").
		SetDescription("test").
		Save(ctx)
	require.NoError(t, err)

	val, err = repo.GetSystemSetting(ctx, "limit.test")
	require.NoError(t, err)
	assert.Equal(t, "42", val)
}

// TestRepository_GetSystemSettingsBatch проверяет слой репозитория: batch-чтение
// нескольких настроек одним запросом; отсутствующие ключи не попадают в результат.
func TestRepository_GetSystemSettingsBatch(t *testing.T) {
	repo := newAuditRepoTestEnv(t)
	ctx := context.Background()

	empty, err := repo.GetSystemSettingsBatch(ctx, nil)
	require.NoError(t, err)
	assert.Empty(t, empty)

	client := audit.ExportRepositoryEntClient(repo)
	for key, value := range map[string]string{"limit.a": "1", "limit.b": "2"} {
		_, err = client.SystemSetting.Create().
			SetKey(key).
			SetValue(value).
			SetDescription("test").
			Save(ctx)
		require.NoError(t, err)
	}

	values, err := repo.GetSystemSettingsBatch(ctx, []string{"limit.a", "limit.b", "limit.missing"})
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"limit.a": "1", "limit.b": "2"}, values)
}

// TestRepository_UpsertSystemSetting проверяет слой репозитория: репозиторий upsert system setting.
func TestRepository_UpsertSystemSetting(t *testing.T) {
	repo := newAuditRepoTestEnv(t)
	ctx := context.Background()

	require.NoError(t, repo.UpsertSystemSetting(ctx, "limit.test", "10", "desc"))
	val, err := repo.GetSystemSetting(ctx, "limit.test")
	require.NoError(t, err)
	assert.Equal(t, "10", val)

	require.NoError(t, repo.UpsertSystemSetting(ctx, "limit.test", "20", "desc"))
	val, err = repo.GetSystemSetting(ctx, "limit.test")
	require.NoError(t, err)
	assert.Equal(t, "20", val)
}

// TestRepository_IsSystemAdmin проверяет слой репозитория: репозиторий is system админ.
func TestRepository_IsSystemAdmin(t *testing.T) {
	repo := newAuditRepoTestEnv(t)
	ctx := context.Background()
	r := audit.ExportRepositoryEntClient(repo)

	admin, err := r.User.Create().
		SetEmail("admin-" + uuid.New().String()[:6] + "@example.com").
		SetSystemRole(user.SystemRoleAdmin).
		Save(ctx)
	require.NoError(t, err)
	member, err := r.User.Create().
		SetEmail("member-" + uuid.New().String()[:6] + "@example.com").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)

	ok, err := repo.IsSystemAdmin(ctx, admin.ID)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = repo.IsSystemAdmin(ctx, member.ID)
	require.NoError(t, err)
	assert.False(t, ok)
}

// TestAuditService_GetResourceLimits проверяет бизнес-логику: audit сервис получение resource лимиты.
func TestAuditService_GetResourceLimits(t *testing.T) {
	repo := newAuditRepoTestEnv(t)
	svc := audit.NewAuditService(repo, nil)
	ctx := context.Background()

	limits, err := svc.GetResourceLimits(ctx)
	require.NoError(t, err)
	require.Len(t, limits, 3)
	assert.Equal(t, "50", limits[0].Value)
}

// batchCountingMockRepo фиксирует вызовы GetSystemSettingsBatch, чтобы
// подтвердить устранение N+1: три лимита должны читаться одним запросом.
type batchCountingMockRepo struct {
	mockAuditRepo
	batchCalls  int
	batchKeys   []string
	batchResult map[string]string
}

func (m *batchCountingMockRepo) GetSystemSettingsBatch(_ context.Context, keys []string) (map[string]string, error) {
	m.batchCalls++
	m.batchKeys = keys
	return m.batchResult, nil
}

// TestAuditService_GetResourceLimits_SingleBatchCall проверяет, что
// GetResourceLimits делает ровно один batch-запрос вместо трёх
// последовательных GetSystemSetting, а отсутствующие ключи получают дефолты.
func TestAuditService_GetResourceLimits_SingleBatchCall(t *testing.T) {
	repo := &batchCountingMockRepo{
		batchResult: map[string]string{
			audit.LimitFileMaxSizeMB: "100",
			// остальные ключи отсутствуют — сервис должен подставить дефолты
		},
	}
	svc := audit.NewAuditService(repo, nil)

	limits, err := svc.GetResourceLimits(context.Background())
	require.NoError(t, err)

	assert.Equal(t, 1, repo.batchCalls, "лимиты должны читаться одним batch-запросом")
	assert.ElementsMatch(t,
		[]string{audit.LimitFileMaxSizeMB, audit.LimitUserMaxProjects, audit.LimitProjectMaxStorageMB},
		repo.batchKeys)

	require.Len(t, limits, 3)
	byKey := make(map[string]string, len(limits))
	for _, l := range limits {
		byKey[l.Key] = l.Value
	}
	assert.Equal(t, "100", byKey[audit.LimitFileMaxSizeMB])
	assert.Equal(t, "20", byKey[audit.LimitUserMaxProjects])
	assert.Equal(t, "5120", byKey[audit.LimitProjectMaxStorageMB])
}

// TestAuditService_CheckLimit проверяет бизнес-логику: audit сервис проверка limit.
func TestAuditService_CheckLimit(t *testing.T) {
	repo := newAuditRepoTestEnv(t)
	svc := audit.NewAuditService(repo, nil)
	ctx := context.Background()

	require.NoError(t, repo.UpsertSystemSetting(ctx, "limit.user.max_projects", "5", "test"))

	tests := []struct {
		name    string
		current int
		wantOK  bool
	}{
		{"under limit", 3, true},
		{"at limit", 5, false},
		{"over limit", 6, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := svc.CheckLimit(ctx, "limit.user.max_projects", tt.current)
			require.NoError(t, err)
			assert.Equal(t, tt.wantOK, ok)
		})
	}
}

// TestAuditService_UpdateResourceLimit проверяет бизнес-логику: audit сервис обновление resource limit.
func TestAuditService_UpdateResourceLimit(t *testing.T) {
	repo := newAuditRepoTestEnv(t)
	svc := audit.NewAuditService(repo, nil)
	ctx := context.Background()

	require.NoError(t, svc.UpdateResourceLimit(ctx, audit.LimitFileMaxSizeMB, "100"))
	val, err := repo.GetSystemSetting(ctx, audit.LimitFileMaxSizeMB)
	require.NoError(t, err)
	assert.Equal(t, "100", val)
}

// TestAuditService_CheckLimit_NoSetting проверяет бизнес-логику: audit сервис проверка limit no setting.
func TestAuditService_CheckLimit_NoSetting(t *testing.T) {
	repo := newAuditRepoTestEnv(t)
	svc := audit.NewAuditService(repo, nil)
	ctx := context.Background()

	ok, err := svc.CheckLimit(ctx, "limit.missing", 999)
	require.NoError(t, err)
	assert.True(t, ok)
}

// TestAuditService_CheckLimit_InvalidValue проверяет отклонение при audit сервис проверка limit некорректный value.
func TestAuditService_CheckLimit_InvalidValue(t *testing.T) {
	repo := newAuditRepoTestEnv(t)
	svc := audit.NewAuditService(repo, nil)
	ctx := context.Background()

	require.NoError(t, repo.UpsertSystemSetting(ctx, "limit.bad", "not-a-number", "test"))
	ok, err := svc.CheckLimit(ctx, "limit.bad", 999)
	require.NoError(t, err)
	assert.True(t, ok)
}

// TestAuditService_ListTrash_UnknownType проверяет бизнес-логику: audit сервис список корзина unknown type.
func TestAuditService_ListTrash_UnknownType(t *testing.T) {
	repo := newAuditRepoTestEnv(t)
	svc := audit.NewAuditService(repo, nil)
	ctx := context.Background()

	resp, err := svc.ListTrash(ctx, "unknown", nil, 1, 50)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Total)
	assert.Empty(t, resp.Items)
}

// TestAuditService_DeactivateUser проверяет бизнес-логику: audit сервис deactivate user.
func TestAuditService_DeactivateUser(t *testing.T) {
	repo := newAuditRepoTestEnv(t)
	svc := audit.NewAuditService(repo, nil)
	ctx := context.Background()
	r := audit.ExportRepositoryEntClient(repo)

	u, err := r.User.Create().SetEmail("deact-" + uuid.New().String()[:6] + "@example.com").Save(ctx)
	require.NoError(t, err)

	require.NoError(t, svc.DeactivateUser(ctx, u.ID))

	updated, err := r.User.Get(ctx, u.ID)
	require.NoError(t, err)
	assert.Equal(t, user.StatusDeactivated, updated.Status)
}

// TestAuditService_ReassignTasks проверяет бизнес-логику: audit сервис reassign tasks.
func TestAuditService_ReassignTasks(t *testing.T) {
	ctx := context.Background()
	from := uuid.New()
	to := uuid.New()
	repo := &mockAuditRepo{}
	svc := audit.NewAuditService(repo, nil)

	count, err := svc.ReassignTasks(ctx, from, to)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}
