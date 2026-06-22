package project_test

// Тесты сервиса проектов: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/project"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/invite"
	entproject "github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/projectmember"
	"github.com/valerii/hub/backend/ent/projecttemplate"
)

// notFoundErr возвращает ошибку, проходящую ent.IsNotFound.
func notFoundErr() error {
	return &ent.NotFoundError{}
}

// constraintErr возвращает ошибку, проходящую ent.IsConstraintError.
func constraintErr() error {
	return &ent.ConstraintError{}
}

// mockRepo реализует интерфейс project.Repository для тестов.
type mockRepo struct {
	projects       map[uuid.UUID]*ent.Project
	members        map[uuid.UUID][]*ent.ProjectMember
	invites        map[string]*ent.Invite // ключ — код приглашения
	memberCountMap map[uuid.UUID]int

	// Инъекция ошибок
	createErr           error
	findByIDErr         error
	findKeyErr          error
	findInviteByCodeErr error
	findMemberErr       error
	removeTeamMemberErr error
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		projects:       make(map[uuid.UUID]*ent.Project),
		members:        make(map[uuid.UUID][]*ent.ProjectMember),
		invites:        make(map[string]*ent.Invite),
		memberCountMap: make(map[uuid.UUID]int),
	}
}

func (m *mockRepo) FindByID(_ context.Context, id uuid.UUID) (*ent.Project, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	p, ok := m.projects[id]
	if !ok {
		return nil, notFoundErr()
	}
	return p, nil
}

func (m *mockRepo) FindByKey(_ context.Context, key string) (*ent.Project, error) {
	if m.findKeyErr != nil {
		return nil, m.findKeyErr
	}
	for _, p := range m.projects {
		if p.Key == key {
			return p, nil
		}
	}
	return nil, notFoundErr()
}

func (m *mockRepo) ListByUserID(_ context.Context, _ uuid.UUID, includeArchived bool) ([]*ent.Project, error) {
	var result []*ent.Project
	for _, p := range m.projects {
		if p.Status == entproject.StatusDeleted {
			continue
		}
		if !includeArchived && p.Status == entproject.StatusArchived {
			continue
		}
		result = append(result, p)
	}
	return result, nil
}

func (m *mockRepo) ListTrash(_ context.Context, _ uuid.UUID) ([]*ent.Project, error) {
	var result []*ent.Project
	for _, p := range m.projects {
		if p.Status == entproject.StatusDeleted {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *mockRepo) Create(_ context.Context, name, key, description string, _ uuid.UUID) (*ent.Project, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	// Проверяем уникальность ключа
	for _, p := range m.projects {
		if p.Key == key {
			return nil, constraintErr()
		}
	}
	id := uuid.New()
	now := time.Now()
	p := &ent.Project{
		ID:          id,
		Name:        name,
		Key:         key,
		Description: description,
		Status:      entproject.StatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	m.projects[id] = p
	m.memberCountMap[id] = 1
	return p, nil
}

func (m *mockRepo) Update(_ context.Context, id uuid.UUID, name, description, avatarURL *string) (*ent.Project, error) {
	p, ok := m.projects[id]
	if !ok {
		return nil, notFoundErr()
	}
	if name != nil {
		p.Name = *name
	}
	if description != nil {
		p.Description = *description
	}
	if avatarURL != nil {
		p.AvatarURL = *avatarURL
	}
	p.UpdatedAt = time.Now()
	return p, nil
}

func (m *mockRepo) UpdateStatus(_ context.Context, id uuid.UUID, status entproject.Status, setDeletedAt bool) (*ent.Project, error) {
	p, ok := m.projects[id]
	if !ok {
		return nil, notFoundErr()
	}
	p.Status = status
	if setDeletedAt {
		now := time.Now()
		p.DeletedAt = &now
	} else {
		p.DeletedAt = nil
	}
	p.UpdatedAt = time.Now()
	return p, nil
}

func (m *mockRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.projects, id)
	return nil
}

func (m *mockRepo) MemberCount(_ context.Context, projectID uuid.UUID) (int, error) {
	return m.memberCountMap[projectID], nil
}

func (m *mockRepo) MemberCountBatch(_ context.Context, projectIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	counts := make(map[uuid.UUID]int, len(projectIDs))
	for _, id := range projectIDs {
		counts[id] = m.memberCountMap[id]
	}
	return counts, nil
}

func (m *mockRepo) ListMembers(_ context.Context, projectID uuid.UUID) ([]*ent.ProjectMember, error) {
	return m.members[projectID], nil
}

func (m *mockRepo) AddMember(_ context.Context, projectID, userID uuid.UUID, role projectmember.ProjectRole) (*ent.ProjectMember, error) {
	// Проверяем, не является ли уже участником
	for _, mem := range m.members[projectID] {
		if mem.Edges.User != nil && mem.Edges.User.ID == userID {
			return nil, constraintErr()
		}
	}
	mem := &ent.ProjectMember{
		ID:          uuid.New(),
		ProjectRole: role,
		JoinedAt:    time.Now(),
		Edges: ent.ProjectMemberEdges{
			User: &ent.User{ID: userID, Email: "test@test.com", Name: "Test"},
		},
	}
	m.members[projectID] = append(m.members[projectID], mem)
	m.memberCountMap[projectID]++
	return mem, nil
}

func (m *mockRepo) RemoveMember(_ context.Context, projectID, userID uuid.UUID) error {
	members := m.members[projectID]
	for i, mem := range members {
		if mem.Edges.User != nil && mem.Edges.User.ID == userID {
			m.members[projectID] = append(members[:i], members[i+1:]...)
			m.memberCountMap[projectID]--
			return nil
		}
	}
	return notFoundErr()
}

func (m *mockRepo) UpdateMemberRole(_ context.Context, projectID, userID uuid.UUID, role projectmember.ProjectRole) (*ent.ProjectMember, error) {
	for _, mem := range m.members[projectID] {
		if mem.Edges.User != nil && mem.Edges.User.ID == userID {
			mem.ProjectRole = role
			return mem, nil
		}
	}
	return nil, notFoundErr()
}

func (m *mockRepo) FindMember(_ context.Context, projectID, userID uuid.UUID) (*ent.ProjectMember, error) {
	if m.findMemberErr != nil {
		return nil, m.findMemberErr
	}
	for _, mem := range m.members[projectID] {
		if mem.Edges.User != nil && mem.Edges.User.ID == userID {
			return mem, nil
		}
	}
	return nil, notFoundErr()
}

func (m *mockRepo) CountOwners(_ context.Context, projectID uuid.UUID) (int, error) {
	count := 0
	for _, mem := range m.members[projectID] {
		if mem.ProjectRole == projectmember.ProjectRoleOwner {
			count++
		}
	}
	return count, nil
}

func (m *mockRepo) FindInviteByCode(_ context.Context, code string) (*ent.Invite, error) {
	if m.findInviteByCodeErr != nil {
		return nil, m.findInviteByCodeErr
	}
	inv, ok := m.invites[code]
	if !ok {
		return nil, notFoundErr()
	}
	return inv, nil
}

func (m *mockRepo) CreateInvite(_ context.Context, _, _ uuid.UUID, code string, role invite.ProjectRole, maxUses int, expiresAt *time.Time) (*ent.Invite, error) {
	inv := &ent.Invite{
		ID:          uuid.New(),
		Code:        code,
		MaxUses:     maxUses,
		UseCount:    0,
		ProjectRole: role,
		ExpiresAt:   expiresAt,
		CreatedAt:   time.Now(),
	}
	m.invites[code] = inv
	return inv, nil
}

func (m *mockRepo) IncrementInviteUseCount(_ context.Context, inviteID uuid.UUID) error {
	for _, inv := range m.invites {
		if inv.ID == inviteID {
			inv.UseCount++
			return nil
		}
	}
	return nil
}

func (m *mockRepo) TryConsumeInvite(_ context.Context, inviteID uuid.UUID, maxUses int) (bool, error) {
	for _, inv := range m.invites {
		if inv.ID == inviteID {
			if maxUses > 0 && inv.UseCount >= maxUses {
				return false, nil
			}
			inv.UseCount++
			return true, nil
		}
	}
	return false, nil
}

func (m *mockRepo) ListInvites(_ context.Context, _ uuid.UUID) ([]*ent.Invite, error) {
	var result []*ent.Invite
	for _, inv := range m.invites {
		result = append(result, inv)
	}
	return result, nil
}

func (m *mockRepo) DeleteInvite(_ context.Context, inviteID uuid.UUID) error {
	for code, inv := range m.invites {
		if inv.ID == inviteID {
			delete(m.invites, code)
			return nil
		}
	}
	return nil
}

func (m *mockRepo) CreateTeam(_ context.Context, name, description string) (*ent.Team, error) {
	return &ent.Team{ID: uuid.New(), Name: name, Description: description}, nil
}

func (m *mockRepo) ListTeams(_ context.Context) ([]*ent.Team, error) {
	return nil, nil
}

func (m *mockRepo) FindTeamByID(_ context.Context, id uuid.UUID) (*ent.Team, error) {
	return nil, notFoundErr()
}

func (m *mockRepo) AddTeamMember(_ context.Context, _, _ uuid.UUID) (*ent.TeamMember, error) {
	return &ent.TeamMember{ID: uuid.New(), JoinedAt: time.Now()}, nil
}

func (m *mockRepo) RemoveTeamMember(_ context.Context, _, _ uuid.UUID) error {
	if m.removeTeamMemberErr != nil {
		return m.removeTeamMemberErr
	}
	return nil
}

func (m *mockRepo) TeamMemberCount(_ context.Context, _ uuid.UUID) (int, error) {
	return 0, nil
}

func (m *mockRepo) ListTeamMembers(_ context.Context, _ uuid.UUID) ([]*ent.TeamMember, error) {
	return nil, nil
}

func (m *mockRepo) AddTeamToProject(_ context.Context, _, _ uuid.UUID) error {
	return nil
}

// --- Очистка межпроектных связей (no-op для unit-тестов) ---

func (m *mockRepo) ListOutgoingCrossProjectLinks(_ context.Context, _ uuid.UUID) ([]*ent.IssueLink, error) {
	return nil, nil
}

func (m *mockRepo) ListIncomingCrossProjectLinks(_ context.Context, _ uuid.UUID) ([]*ent.IssueLink, error) {
	return nil, nil
}

func (m *mockRepo) DeleteIssueLink(_ context.Context, _ uuid.UUID) error {
	return nil
}

// --- Операции с шаблонами проектов (no-op для unit-тестов) ---

func (m *mockRepo) ListTemplates(_ context.Context) ([]*ent.ProjectTemplate, error) {
	return nil, nil
}

func (m *mockRepo) GetTemplate(_ context.Context, _ uuid.UUID) (*ent.ProjectTemplate, error) {
	return nil, notFoundErr()
}

func (m *mockRepo) GetProjectWithWorkflowStatusesAndLabels(_ context.Context, _ uuid.UUID) (*ent.Project, []string, []string, error) {
	return nil, nil, nil, notFoundErr()
}

func (m *mockRepo) CreateTemplate(_ context.Context, name, description string, templateType projecttemplate.TemplateType, config map[string]interface{}, _ *uuid.UUID, isSystem bool) (*ent.ProjectTemplate, error) {
	return &ent.ProjectTemplate{
		ID:           uuid.New(),
		Name:         name,
		Description:  description,
		TemplateType: templateType,
		Config:       config,
		IsSystem:     isSystem,
		CreatedAt:    time.Now(),
	}, nil
}

func (m *mockRepo) DeleteTemplate(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockRepo) CountSystemTemplates(_ context.Context) (int, error) {
	return 0, nil
}

// mockCache реализует CacheInvalidator для тестов.
type mockCache struct {
	invalidated []string
}

func (mc *mockCache) InvalidateProject(_ context.Context, projectID string) error {
	mc.invalidated = append(mc.invalidated, projectID)
	return nil
}

func (mc *mockCache) Invalidate(_ context.Context, userID, projectID string) error {
	mc.invalidated = append(mc.invalidated, userID+":"+projectID)
	return nil
}

func newTestService() (*project.ProjectService, *mockRepo, *mockCache) {
	repo := newMockRepo()
	cache := &mockCache{}
	svc := project.NewProjectService(repo, cache, nil)
	return svc, repo, cache
}

// TestCreateProject_Success проверяет сценарий: создание project успех.
func TestCreateProject_Success(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()
	userID := uuid.New()

	resp, err := svc.CreateProject(ctx, userID, "My Project", "MP", "A test project")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Name != "My Project" {
		t.Errorf("expected name 'My Project', got %q", resp.Name)
	}
	if resp.Key != "MP" {
		t.Errorf("expected key 'MP', got %q", resp.Key)
	}
	if resp.Status != "active" {
		t.Errorf("expected status 'active', got %q", resp.Status)
	}
	if resp.MemberCount != 1 {
		t.Errorf("expected member_count 1, got %d", resp.MemberCount)
	}
}

// TestCreateProject_DuplicateKey проверяет сценарий: создание project duplicate key.
func TestCreateProject_DuplicateKey(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()
	userID := uuid.New()

	_, err := svc.CreateProject(ctx, userID, "Project One", "DUP", "First")
	if err != nil {
		t.Fatalf("unexpected error on first create: %v", err)
	}

	_, err = svc.CreateProject(ctx, userID, "Project Two", "DUP", "Second")
	if err == nil {
		t.Fatal("expected conflict error for duplicate key, got nil")
	}
}

// TestArchiveProject_SetsStatusArchived проверяет сценарий: archive project sets status archived.
func TestArchiveProject_SetsStatusArchived(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()
	userID := uuid.New()

	resp, err := svc.CreateProject(ctx, userID, "Archive Test", "ARCH", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	projectID, _ := uuid.Parse(resp.ID)
	archived, err := svc.ArchiveProject(ctx, projectID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if archived.Status != "archived" {
		t.Errorf("expected status 'archived', got %q", archived.Status)
	}
}

// TestDeleteProject_SoftDelete проверяет сценарий: удаление project soft удаление.
func TestDeleteProject_SoftDelete(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	userID := uuid.New()

	resp, err := svc.CreateProject(ctx, userID, "Delete Test", "DEL", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	projectID, _ := uuid.Parse(resp.ID)
	err = svc.DeleteProject(ctx, projectID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p := repo.projects[projectID]
	if p.Status != entproject.StatusDeleted {
		t.Errorf("expected status 'deleted', got %q", p.Status)
	}
	if p.DeletedAt == nil {
		t.Error("expected deleted_at to be set, got nil")
	}
}

// TestJoinByInvite_ValidCode проверяет сценарий: join by приглашение корректный code.
func TestJoinByInvite_ValidCode(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	resp, err := svc.CreateProject(ctx, ownerID, "Invite Test", "INV", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	projectID, _ := uuid.Parse(resp.ID)

	// Создаём приглашение вручную в моке
	inv := &ent.Invite{
		ID:          uuid.New(),
		Code:        "TESTCODE",
		MaxUses:     10,
		UseCount:    0,
		ProjectRole: invite.ProjectRoleDeveloper,
		CreatedAt:   time.Now(),
		Edges: ent.InviteEdges{
			Project: repo.projects[projectID],
		},
	}
	repo.invites["TESTCODE"] = inv

	joinerID := uuid.New()
	joinResp, err := svc.JoinByInvite(ctx, "TESTCODE", joinerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if joinResp.ID != resp.ID {
		t.Errorf("expected project ID %s, got %s", resp.ID, joinResp.ID)
	}
}

// TestJoinByInvite_ExpiredCode проверяет отклонение при join by приглашение просроченный code.
func TestJoinByInvite_ExpiredCode(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	resp, err := svc.CreateProject(ctx, ownerID, "Expire Test", "EXP", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	projectID, _ := uuid.Parse(resp.ID)

	past := time.Now().Add(-1 * time.Hour)
	inv := &ent.Invite{
		ID:          uuid.New(),
		Code:        "EXPIRED",
		MaxUses:     10,
		UseCount:    0,
		ProjectRole: invite.ProjectRoleDeveloper,
		ExpiresAt:   &past,
		CreatedAt:   time.Now(),
		Edges: ent.InviteEdges{
			Project: repo.projects[projectID],
		},
	}
	repo.invites["EXPIRED"] = inv

	joinerID := uuid.New()
	_, err = svc.JoinByInvite(ctx, "EXPIRED", joinerID)
	if err == nil {
		t.Fatal("expected error for expired invite, got nil")
	}
}

// TestRemoveMember_LastOwner проверяет сценарий: remove участник last owner.
func TestRemoveMember_LastOwner(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	resp, err := svc.CreateProject(ctx, ownerID, "Owner Test", "OWN", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	projectID, _ := uuid.Parse(resp.ID)

	// Добавляем владельца как участника в мок
	repo.members[projectID] = []*ent.ProjectMember{
		{
			ID:          uuid.New(),
			ProjectRole: projectmember.ProjectRoleOwner,
			JoinedAt:    time.Now(),
			Edges: ent.ProjectMemberEdges{
				User: &ent.User{ID: ownerID, Email: "owner@test.com", Name: "Owner"},
			},
		},
	}

	err = svc.RemoveMember(ctx, projectID, ownerID)
	if err == nil {
		t.Fatal("expected error when removing last owner, got nil")
	}
}

// Подавляем предупреждение о неиспользуемом импорте.
var _ = fmt.Sprintf
