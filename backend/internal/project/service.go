// Package project реализует управление проектами: CRUD, участники, роли, команды, шаблоны.
package project

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/projectmember"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
	safehttp "github.com/valerii/hub/backend/internal/platform/http"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
	"github.com/valerii/hub/backend/internal/platform/observability"
)

// ProjectService реализует бизнес-логику управления проектами.
type ProjectService struct {
	repo         Repository
	cache        CacheInvalidator
	nats         *natsclient.Client
	trackerProv  TrackerProvisioner
	tmsProv      TMSProvisioner
	dashProv     DashboardProvisioner
	frontendURL  string // P3-BE-09 / W1-A5-011: base для invite Link
}

// SetFrontendURL задаёт публичный URL фронтенда для invite-ссылок.
func (s *ProjectService) SetFrontendURL(url string) {
	s.frontendURL = strings.TrimRight(strings.TrimSpace(url), "/")
}

// resolveFrontendURL — FRONTEND_URL env, иначе поле сервиса, иначе первый CORS_ORIGINS.
func (s *ProjectService) resolveFrontendURL() string {
	if u := strings.TrimRight(strings.TrimSpace(os.Getenv("FRONTEND_URL")), "/"); u != "" {
		return u
	}
	if s != nil && s.frontendURL != "" {
		return s.frontendURL
	}
	if cors := strings.TrimSpace(os.Getenv("CORS_ORIGINS")); cors != "" {
		first := strings.TrimSpace(strings.Split(cors, ",")[0])
		return strings.TrimRight(first, "/")
	}
	return ""
}

// memberCountOrZero безопасно читает количество участников проекта для UI-ответа.
// Если запрос упал — логируем как warn и возвращаем 0, не блокируя основной флоу
// (раньше count, _ := игнорировал ошибку молча — MEDIUM из аудита).
func (s *ProjectService) memberCountOrZero(ctx context.Context, projectID uuid.UUID) int {
	count, err := s.repo.MemberCount(ctx, projectID)
	if err != nil {
		slog.Warn("project member count failed; using 0 fallback",
			"project_id", projectID, "error", err)
		return 0
	}
	return count
}

// NewProjectService создаёт новый сервис проектов.
// nats может быть nil (graceful degradation для notifyLinkDeleted).
func NewProjectService(repo Repository, cache CacheInvalidator, nats *natsclient.Client) *ProjectService {
	return &ProjectService{
		repo:  repo,
		cache: cache,
		nats:  nats,
	}
}

// CreateProject создаёт проект с указанным пользователем как владельцем.
// Публикует outbox-событие project.created для downstream-потребителей.
func (s *ProjectService) CreateProject(ctx context.Context, userID uuid.UUID, name, key, description string) (*ProjectResponse, error) {
	return observability.Timed(ctx, "project.CreateProject", func(ctx context.Context) (*ProjectResponse, error) {
		// Проверяем уникальность ключа
		_, err := s.repo.FindByKey(ctx, key)
		if err == nil {
			return nil, apperrors.NewConflict(fmt.Sprintf("project key %q already exists", key))
		}
		if !ent.IsNotFound(err) {
			return nil, apperrors.NewInternal("failed to check key uniqueness")
		}

		p, err := s.repo.Create(ctx, name, key, description, userID)
		if err != nil {
			if ent.IsConstraintError(err) {
				return nil, apperrors.NewConflict(fmt.Sprintf("project key %q already exists", key))
			}
			return nil, apperrors.NewInternal("failed to create project")
		}

		// Автопровизия: создаём workflow, доску, тест-сьют и дашборд
		s.provisionProject(ctx, p.ID, userID)

		count, err := s.repo.MemberCount(ctx, p.ID)
		if err != nil {
			count = 1 // запасной вариант: создатель всегда участник
		}

		return toProjectResponse(p, count, "owner"), nil
	})
}

// GetProject возвращает проект по ID и роль текущего пользователя (если userID задан).
// W1-A5-006 / P2-BE-12: soft-deleted (status=deleted) → 404; не член → 404 (без утечки метаданных).
func (s *ProjectService) GetProject(ctx context.Context, projectID, userID uuid.UUID) (*ProjectResponse, error) {
	p, err := s.repo.FindByID(ctx, projectID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("project not found")
		}
		return nil, apperrors.NewInternal("failed to get project")
	}
	if p.Status == project.StatusDeleted {
		return nil, apperrors.NewNotFound("project not found")
	}

	myRole := ""
	if userID != uuid.Nil {
		mem, err := s.repo.FindMember(ctx, projectID, userID)
		if err != nil {
			if ent.IsNotFound(err) {
				return nil, apperrors.NewNotFound("project not found")
			}
			return nil, apperrors.NewInternal("failed to get project membership")
		}
		myRole = string(mem.ProjectRole)
	}

	return toProjectResponse(p, s.memberCountOrZero(ctx, projectID), myRole), nil
}

// ListProjects возвращает проекты пользователя (без удалённых).
func (s *ProjectService) ListProjects(ctx context.Context, userID uuid.UUID, includeArchived bool) ([]ProjectResponse, error) {
	projects, err := s.repo.ListByUserID(ctx, userID, includeArchived)
	if err != nil {
		return nil, apperrors.NewInternal("failed to list projects")
	}

	// Пакетно получаем счётчики участников одним запросом, чтобы избежать N+1.
	ids := make([]uuid.UUID, len(projects))
	for i, p := range projects {
		ids[i] = p.ID
	}
	counts, err := s.repo.MemberCountBatch(ctx, ids)
	if err != nil {
		counts = make(map[uuid.UUID]int) // запасной вариант: нули
	}

	result := make([]ProjectResponse, 0, len(projects))
	for _, p := range projects {
		result = append(result, *toProjectResponse(p, counts[p.ID], ""))
	}
	return result, nil
}

// UpdateProject обновляет имя, описание или аватар проекта.
// W1-A5-007 / P2-BE-12: avatar_url — только https/http или относительный путь `/…`.
func (s *ProjectService) UpdateProject(ctx context.Context, projectID uuid.UUID, req UpdateProjectRequest) (*ProjectResponse, error) {
	if req.AvatarURL != nil {
		if err := validateProjectAvatarURL(*req.AvatarURL); err != nil {
			return nil, apperrors.NewBadRequest("invalid avatar_url")
		}
	}
	p, err := s.repo.Update(ctx, projectID, req.Name, req.Description, req.AvatarURL)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("project not found")
		}
		return nil, apperrors.NewInternal("failed to update project")
	}
	return toProjectResponse(p, s.memberCountOrZero(ctx, projectID), ""), nil
}

// validateProjectAvatarURL отклоняет javascript:/data: и прочие не-http(s) схемы.
// Пустая строка и относительные пути (`/files/...`) допустимы.
func validateProjectAvatarURL(raw string) error {
	if raw == "" {
		return nil
	}
	if strings.HasPrefix(raw, "/") && !strings.HasPrefix(raw, "//") {
		return nil
	}
	if err := safehttp.ValidateOutboundURL(raw); err != nil {
		return err
	}
	// Доп. защита: отклоняем URL без явного host после parse (ValidateOutboundURL уже проверяет).
	u, err := url.Parse(raw)
	if err != nil {
		return err
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("unsupported avatar scheme: %s", scheme)
	}
	return nil
}

// ArchiveProject переводит проект в статус archived.
func (s *ProjectService) ArchiveProject(ctx context.Context, projectID uuid.UUID) (*ProjectResponse, error) {
	p, err := s.repo.UpdateStatus(ctx, projectID, project.StatusArchived, false)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("project not found")
		}
		return nil, apperrors.NewInternal("failed to archive project")
	}
	return toProjectResponse(p, s.memberCountOrZero(ctx, projectID), ""), nil
}

// UnarchiveProject возвращает проект из archived в active.
func (s *ProjectService) UnarchiveProject(ctx context.Context, projectID uuid.UUID) (*ProjectResponse, error) {
	p, err := s.repo.UpdateStatus(ctx, projectID, project.StatusActive, false)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("project not found")
		}
		return nil, apperrors.NewInternal("failed to unarchive project")
	}
	return toProjectResponse(p, s.memberCountOrZero(ctx, projectID), ""), nil
}

// DeleteProject выполняет мягкое удаление проекта: статус = deleted, deleted_at = now.
// Также удаляет кросс-проектные связи задач и уведомляет исполнителей.
// A2W1-A5-001: cleanup + soft-delete атомарны (tx) и fail-closed.
func (s *ProjectService) DeleteProject(ctx context.Context, projectID uuid.UUID) error {
	p, err := s.repo.FindByID(ctx, projectID)
	if err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("project not found")
		}
		return apperrors.NewInternal("failed to find project")
	}

	outgoing, incoming, err := s.repo.SoftDeleteWithCrossLinkCleanup(ctx, projectID)
	if err != nil {
		slog.ErrorContext(ctx, "delete project: soft-delete with link cleanup failed",
			"project_id", projectID, "error", err)
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("project not found")
		}
		return apperrors.NewInternal("failed to delete project")
	}

	for _, link := range outgoing {
		if link.Edges.Source != nil && link.Edges.Target != nil {
			s.notifyLinkDeleted(ctx, link.Edges.Target, link.Edges.Source.IssueKey, p.Name)
		}
	}
	for _, link := range incoming {
		if link.Edges.Target != nil && link.Edges.Source != nil {
			s.notifyLinkDeleted(ctx, link.Edges.Source, link.Edges.Target.IssueKey, p.Name)
		}
	}
	return nil
}

// notifyLinkDeleted отправляет NATS-уведомление исполнителю задачи о том,
// что связь удалена из-за удаления проекта.
func (s *ProjectService) notifyLinkDeleted(ctx context.Context, otherIssue *ent.Issue, deletedIssueKey, deletedProjectName string) {
	if s.nats == nil || otherIssue == nil {
		return
	}

	assignee := otherIssue.Edges.Assignee
	if assignee == nil {
		return
	}

	payload := map[string]interface{}{
		"user_id": assignee.ID.String(),
		"type":    "link_deleted",
		"title":   fmt.Sprintf("Связь с %s удалена — проект %s удалён", deletedIssueKey, deletedProjectName),
		"entity":  "issue",
		"entity_id": otherIssue.ID.String(),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	if pErr := s.nats.Publish("project.link_deleted", data); pErr != nil {
		slog.Error("не удалось отправить уведомление об удалении связи", "error", pErr)
	}
}

// RestoreProject восстанавливает проект из корзины (active, cleared deleted_at).
func (s *ProjectService) RestoreProject(ctx context.Context, projectID uuid.UUID) (*ProjectResponse, error) {
	p, err := s.repo.UpdateStatus(ctx, projectID, project.StatusActive, false)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("project not found")
		}
		return nil, apperrors.NewInternal("failed to restore project")
	}
	return toProjectResponse(p, s.memberCountOrZero(ctx, projectID), ""), nil
}

// PermanentDelete выполняет жёсткое удаление проекта. Только admin.
func (s *ProjectService) PermanentDelete(ctx context.Context, projectID uuid.UUID) error {
	err := s.repo.Delete(ctx, projectID)
	if err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("project not found")
		}
		return apperrors.NewInternal("failed to permanently delete project")
	}
	return nil
}

// ListTrash возвращает проекты в корзине (status=deleted) для пользователя.
func (s *ProjectService) ListTrash(ctx context.Context, userID uuid.UUID) ([]ProjectResponse, error) {
	projects, err := s.repo.ListTrash(ctx, userID)
	if err != nil {
		return nil, apperrors.NewInternal("failed to list trash")
	}

	// Пакетно получаем счётчики участников одним запросом, чтобы избежать N+1.
	ids := make([]uuid.UUID, len(projects))
	for i, p := range projects {
		ids[i] = p.ID
	}
	counts, err := s.repo.MemberCountBatch(ctx, ids)
	if err != nil {
		counts = make(map[uuid.UUID]int) // запасной вариант: нули
	}

	result := make([]ProjectResponse, 0, len(projects))
	for _, p := range projects {
		result = append(result, *toProjectResponse(p, counts[p.ID], ""))
	}
	return result, nil
}

// AddMember добавляет пользователя в проект с указанной ролью.
// Инвалидирует RBAC-кэш проекта — иначе новый member не увидит права до TTL.
func (s *ProjectService) AddMember(ctx context.Context, projectID, userID uuid.UUID, role string) (*MemberResponse, error) {
	prole := projectmember.ProjectRole(role)
	mem, err := s.repo.AddMember(ctx, projectID, userID, prole)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, apperrors.NewConflict("user is already a member of this project")
		}
		return nil, apperrors.NewInternal("failed to add member")
	}

	// Сбрасываем кэш RBAC
	if s.cache != nil {
		if err := s.cache.InvalidateProject(ctx, projectID.String()); err != nil {
			slog.Error("failed to invalidate RBAC cache", "error", err, "project_id", projectID)
		}
	}

	return toMemberResponse(mem), nil
}

// RemoveMember удаляет пользователя из проекта.
// Нельзя удалить последнего owner — иначе проект останется без управляющего.
func (s *ProjectService) RemoveMember(ctx context.Context, projectID, userID uuid.UUID) error {
	// Проверяем, является ли пользователь владельцем
	mem, err := s.repo.FindMember(ctx, projectID, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("member not found")
		}
		return apperrors.NewInternal("failed to find member")
	}

	if mem.ProjectRole == projectmember.ProjectRoleOwner {
		ownerCount, err := s.repo.CountOwners(ctx, projectID)
		if err != nil {
			return apperrors.NewInternal("failed to count owners")
		}
		if ownerCount <= 1 {
			return apperrors.NewBadRequest("cannot remove the last owner of a project")
		}
	}

	if err := s.repo.RemoveMember(ctx, projectID, userID); err != nil {
		return apperrors.NewInternal("failed to remove member")
	}

	// Сбрасываем кэш RBAC
	if s.cache != nil {
		if err := s.cache.InvalidateProject(ctx, projectID.String()); err != nil {
			slog.Error("failed to invalidate RBAC cache", "error", err, "project_id", projectID)
		}
	}

	return nil
}

// UpdateMemberRole меняет роль участника проекта.
// P1-BE-07: нельзя демоутить последнего owner (как в RemoveMember).
func (s *ProjectService) UpdateMemberRole(ctx context.Context, projectID, userID uuid.UUID, role string) (*MemberResponse, error) {
	prole := projectmember.ProjectRole(role)

	mem, err := s.repo.FindMember(ctx, projectID, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("member not found")
		}
		return nil, apperrors.NewInternal("failed to find member")
	}

	if mem.ProjectRole == projectmember.ProjectRoleOwner && prole != projectmember.ProjectRoleOwner {
		ownerCount, err := s.repo.CountOwners(ctx, projectID)
		if err != nil {
			return nil, apperrors.NewInternal("failed to count owners")
		}
		if ownerCount <= 1 {
			return nil, apperrors.NewBadRequest("cannot demote the last owner of a project")
		}
	}

	mem, err = s.repo.UpdateMemberRole(ctx, projectID, userID, prole)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("member not found")
		}
		return nil, apperrors.NewInternal("failed to update member role")
	}

	// Сбрасываем кэш RBAC для конкретного пользователя
	if s.cache != nil {
		if err := s.cache.Invalidate(ctx, userID.String(), projectID.String()); err != nil {
			slog.Error("failed to invalidate RBAC cache", "error", err, "user_id", userID, "project_id", projectID)
		}
	}

	return toMemberResponse(mem), nil
}

// ListMembers возвращает всех участников проекта.
func (s *ProjectService) ListMembers(ctx context.Context, projectID uuid.UUID) ([]MemberResponse, error) {
	members, err := s.repo.ListMembers(ctx, projectID)
	if err != nil {
		return nil, apperrors.NewInternal("failed to list members")
	}

	result := make([]MemberResponse, 0, len(members))
	for _, m := range members {
		result = append(result, *toMemberResponse(m))
	}
	return result, nil
}

// toProjectResponse конвертирует Ent Project в DTO ответа.
func toProjectResponse(p *ent.Project, memberCount int, myRole string) *ProjectResponse {
	return &ProjectResponse{
		ID:          p.ID.String(),
		Name:        p.Name,
		Key:         p.Key,
		Description: p.Description,
		AvatarURL:   p.AvatarURL,
		Status:      string(p.Status),
		MemberCount: memberCount,
		MyRole:      myRole,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

// toMemberResponse конвертирует Ent ProjectMember в DTO ответа.
func toMemberResponse(m *ent.ProjectMember) *MemberResponse {
	resp := &MemberResponse{
		Role:     string(m.ProjectRole),
		JoinedAt: m.JoinedAt,
	}
	if m.Edges.User != nil {
		resp.UserID = m.Edges.User.ID.String()
		resp.Email = m.Edges.User.Email
		resp.Name = m.Edges.User.Name
		resp.AvatarURL = m.Edges.User.AvatarURL
	}
	return resp
}

// toInviteResponse конвертирует Ent Invite в DTO ответа.
func toInviteResponse(inv *ent.Invite, frontendURL string) *InviteResponse {
	return &InviteResponse{
		ID:        inv.ID.String(),
		Code:      inv.Code,
		Role:      string(inv.ProjectRole),
		MaxUses:   inv.MaxUses,
		UseCount:  inv.UseCount,
		ExpiresAt: inv.ExpiresAt,
		Link:      fmt.Sprintf("%s/invite/%s", frontendURL, inv.Code),
		CreatedAt: inv.CreatedAt,
	}
}

// CreateTeam создаёт команду; создатель сразу становится участником (ownership без миграции Team).
func (s *ProjectService) CreateTeam(ctx context.Context, creatorID uuid.UUID, name, description string) (*TeamResponse, error) {
	t, err := s.repo.CreateTeam(ctx, name, description)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, apperrors.NewConflict("team name already exists")
		}
		return nil, apperrors.NewInternal("failed to create team")
	}
	if _, err := s.repo.AddTeamMember(ctx, t.ID, creatorID); err != nil {
		return nil, apperrors.NewInternal("failed to add team creator")
	}
	return toTeamResponse(t, 1), nil
}

// ListTeams возвращает команды, в которых состоит пользователь (P1-BE-08).
func (s *ProjectService) ListTeams(ctx context.Context, userID uuid.UUID) ([]TeamResponse, error) {
	teams, err := s.repo.ListTeamsByUserID(ctx, userID)
	if err != nil {
		return nil, apperrors.NewInternal("failed to list teams")
	}
	result := make([]TeamResponse, 0, len(teams))
	for _, t := range teams {
		result = append(result, *toTeamResponse(t, len(t.Edges.Members)))
	}
	return result, nil
}

// requireTeamMember — ACL: мутации команды только для её участников.
func (s *ProjectService) requireTeamMember(ctx context.Context, teamID, actorID uuid.UUID) error {
	if _, err := s.repo.FindTeamMember(ctx, teamID, actorID); err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewForbidden("not a team member")
		}
		return apperrors.NewInternal("failed to check team membership")
	}
	return nil
}

// AddTeamMember добавляет пользователя в команду (только участник команды).
func (s *ProjectService) AddTeamMember(ctx context.Context, teamID, actorID, userID uuid.UUID) error {
	if err := s.requireTeamMember(ctx, teamID, actorID); err != nil {
		return err
	}
	_, err := s.repo.AddTeamMember(ctx, teamID, userID)
	if err != nil {
		if ent.IsConstraintError(err) {
			return apperrors.NewConflict("user is already a member of this team")
		}
		return apperrors.NewInternal("failed to add team member")
	}
	return nil
}

// RemoveTeamMember удаляет пользователя из команды (только участник команды).
func (s *ProjectService) RemoveTeamMember(ctx context.Context, teamID, actorID, userID uuid.UUID) error {
	if err := s.requireTeamMember(ctx, teamID, actorID); err != nil {
		return err
	}
	if err := s.repo.RemoveTeamMember(ctx, teamID, userID); err != nil {
		return apperrors.NewInternal("failed to remove team member")
	}
	return nil
}

// AddTeamToProject привязывает команду к проекту и bulk-добавляет участников.
// P1-BE-08: только участник команды (actor) — stranger → 403, даже при PermTeamManage.
func (s *ProjectService) AddTeamToProject(ctx context.Context, teamID, projectID, actorID uuid.UUID) error {
	if err := s.requireTeamMember(ctx, teamID, actorID); err != nil {
		return err
	}
	if err := s.repo.AddTeamToProject(ctx, teamID, projectID); err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("team or project not found")
		}
		return apperrors.NewInternal("failed to add team to project")
	}
	return nil
}

// toTeamResponse конвертирует Ent Team в DTO ответа.
func toTeamResponse(t *ent.Team, memberCount int) *TeamResponse {
	return &TeamResponse{
		ID:          t.ID.String(),
		Name:        t.Name,
		Description: t.Description,
		MemberCount: memberCount,
	}
}

// ListInvites возвращает все приглашения проекта.
func (s *ProjectService) ListInvites(ctx context.Context, projectID uuid.UUID) ([]InviteResponse, error) {
	invites, err := s.repo.ListInvites(ctx, projectID)
	if err != nil {
		return nil, apperrors.NewInternal("failed to list invites")
	}
	base := s.resolveFrontendURL()
	result := make([]InviteResponse, 0, len(invites))
	for _, inv := range invites {
		result = append(result, *toInviteResponse(inv, base))
	}
	return result, nil
}

// RevokeInvite отзывает приглашение в рамках projectID (P1-BE-06).
func (s *ProjectService) RevokeInvite(ctx context.Context, projectID, inviteID uuid.UUID) error {
	if err := s.repo.DeleteInvite(ctx, projectID, inviteID); err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("invite not found")
		}
		return apperrors.NewInternal("failed to revoke invite")
	}
	return nil
}

