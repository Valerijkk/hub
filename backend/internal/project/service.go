package project

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/issuelink"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/projectmember"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

// ProjectService реализует бизнес-логику управления проектами.
type ProjectService struct {
	repo         Repository
	cache        CacheInvalidator
	db           *ent.Client
	nats         *natsclient.Client
	trackerProv  TrackerProvisioner
	tmsProv      TMSProvisioner
	dashProv     DashboardProvisioner
}

// NewProjectService создаёт новый сервис проектов.
func NewProjectService(repo Repository, cache CacheInvalidator) *ProjectService {
	return &ProjectService{
		repo:  repo,
		cache: cache,
	}
}

// SetDB устанавливает Ent-клиент для очистки кросс-проектных связей.
func (s *ProjectService) SetDB(db *ent.Client) {
	s.db = db
}

// SetNATS устанавливает NATS-клиент для нотификации при удалении связей.
func (s *ProjectService) SetNATS(nats *natsclient.Client) {
	s.nats = nats
}

// CreateProject creates a new project with the given user as owner.
// Publishes a project.created outbox event for downstream consumers.
func (s *ProjectService) CreateProject(ctx context.Context, userID uuid.UUID, name, key, description string) (*ProjectResponse, error) {
	// Check key uniqueness
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
		count = 1 // fallback: creator is always a member
	}

	return toProjectResponse(p, count, "owner"), nil
}

// GetProject returns a project by ID.
func (s *ProjectService) GetProject(ctx context.Context, projectID uuid.UUID) (*ProjectResponse, error) {
	p, err := s.repo.FindByID(ctx, projectID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("project not found")
		}
		return nil, apperrors.NewInternal("failed to get project")
	}
	count, _ := s.repo.MemberCount(ctx, projectID)
	return toProjectResponse(p, count, ""), nil
}

// ListProjects returns all projects the user is a member of (excluding deleted).
func (s *ProjectService) ListProjects(ctx context.Context, userID uuid.UUID, includeArchived bool) ([]ProjectResponse, error) {
	projects, err := s.repo.ListByUserID(ctx, userID, includeArchived)
	if err != nil {
		return nil, apperrors.NewInternal("failed to list projects")
	}

	result := make([]ProjectResponse, 0, len(projects))
	for _, p := range projects {
		count, _ := s.repo.MemberCount(ctx, p.ID)
		result = append(result, *toProjectResponse(p, count, ""))
	}
	return result, nil
}

// UpdateProject updates a project's name, description, or avatar.
func (s *ProjectService) UpdateProject(ctx context.Context, projectID uuid.UUID, req UpdateProjectRequest) (*ProjectResponse, error) {
	p, err := s.repo.Update(ctx, projectID, req.Name, req.Description, req.AvatarURL)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("project not found")
		}
		return nil, apperrors.NewInternal("failed to update project")
	}
	count, _ := s.repo.MemberCount(ctx, projectID)
	return toProjectResponse(p, count, ""), nil
}

// ArchiveProject sets the project status to archived.
func (s *ProjectService) ArchiveProject(ctx context.Context, projectID uuid.UUID) (*ProjectResponse, error) {
	p, err := s.repo.UpdateStatus(ctx, projectID, project.StatusArchived, false)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("project not found")
		}
		return nil, apperrors.NewInternal("failed to archive project")
	}
	count, _ := s.repo.MemberCount(ctx, projectID)
	return toProjectResponse(p, count, ""), nil
}

// UnarchiveProject returns the project from archived to active status.
func (s *ProjectService) UnarchiveProject(ctx context.Context, projectID uuid.UUID) (*ProjectResponse, error) {
	p, err := s.repo.UpdateStatus(ctx, projectID, project.StatusActive, false)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("project not found")
		}
		return nil, apperrors.NewInternal("failed to unarchive project")
	}
	count, _ := s.repo.MemberCount(ctx, projectID)
	return toProjectResponse(p, count, ""), nil
}

// DeleteProject выполняет мягкое удаление проекта: статус = deleted, deleted_at = now.
// Также удаляет кросс-проектные связи задач и уведомляет исполнителей.
func (s *ProjectService) DeleteProject(ctx context.Context, projectID uuid.UUID) error {
	p, err := s.repo.FindByID(ctx, projectID)
	if err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("project not found")
		}
		return apperrors.NewInternal("failed to find project")
	}

	// Удаляем кросс-проектные связи до soft-delete.
	s.cleanupCrossProjectLinks(ctx, projectID, p.Name)

	_, err = s.repo.UpdateStatus(ctx, projectID, project.StatusDeleted, true)
	if err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("project not found")
		}
		return apperrors.NewInternal("failed to delete project")
	}
	return nil
}

// cleanupCrossProjectLinks удаляет связи задач, где одна сторона принадлежит
// удаляемому проекту, а другая — другому проекту. Отправляет NATS-уведомление
// исполнителю задачи из другого проекта.
func (s *ProjectService) cleanupCrossProjectLinks(ctx context.Context, projectID uuid.UUID, projectName string) {
	if s.db == nil {
		return
	}

	// Связи, где source принадлежит этому проекту, а target — другому.
	outgoing, err := s.db.IssueLink.Query().
		Where(
			issuelink.HasSourceWith(issue.HasProjectWith(project.IDEQ(projectID))),
			issuelink.HasTargetWith(issue.HasProjectWith(project.IDNEQ(projectID))),
		).
		WithSource().
		WithTarget(func(q *ent.IssueQuery) { q.WithAssignee() }).
		All(ctx)
	if err != nil {
		slog.Error("ошибка поиска исходящих кросс-проектных связей", "error", err)
	}

	// Связи, где target принадлежит этому проекту, а source — другому.
	incoming, err := s.db.IssueLink.Query().
		Where(
			issuelink.HasTargetWith(issue.HasProjectWith(project.IDEQ(projectID))),
			issuelink.HasSourceWith(issue.HasProjectWith(project.IDNEQ(projectID))),
		).
		WithTarget().
		WithSource(func(q *ent.IssueQuery) { q.WithAssignee() }).
		All(ctx)
	if err != nil {
		slog.Error("ошибка поиска входящих кросс-проектных связей", "error", err)
	}

	// Обрабатываем исходящие: уведомляем assignee target-задачи.
	for _, link := range outgoing {
		if dErr := s.db.IssueLink.DeleteOneID(link.ID).Exec(ctx); dErr != nil {
			slog.Error("ошибка удаления кросс-проектной связи", "link_id", link.ID, "error", dErr)
			continue
		}
		if link.Edges.Source != nil && link.Edges.Target != nil {
			s.notifyLinkDeleted(ctx, link.Edges.Target, link.Edges.Source.IssueKey, projectName)
		}
	}

	// Обрабатываем входящие: уведомляем assignee source-задачи.
	for _, link := range incoming {
		if dErr := s.db.IssueLink.DeleteOneID(link.ID).Exec(ctx); dErr != nil {
			slog.Error("ошибка удаления кросс-проектной связи", "link_id", link.ID, "error", dErr)
			continue
		}
		if link.Edges.Target != nil && link.Edges.Source != nil {
			s.notifyLinkDeleted(ctx, link.Edges.Source, link.Edges.Target.IssueKey, projectName)
		}
	}
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

// RestoreProject restores a project from the trash (sets status to active, clears deleted_at).
func (s *ProjectService) RestoreProject(ctx context.Context, projectID uuid.UUID) (*ProjectResponse, error) {
	p, err := s.repo.UpdateStatus(ctx, projectID, project.StatusActive, false)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("project not found")
		}
		return nil, apperrors.NewInternal("failed to restore project")
	}
	count, _ := s.repo.MemberCount(ctx, projectID)
	return toProjectResponse(p, count, ""), nil
}

// PermanentDelete performs a hard delete of a project. Admin only.
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

// ListTrash returns projects in the trash (status=deleted) for the given user.
func (s *ProjectService) ListTrash(ctx context.Context, userID uuid.UUID) ([]ProjectResponse, error) {
	projects, err := s.repo.ListTrash(ctx, userID)
	if err != nil {
		return nil, apperrors.NewInternal("failed to list trash")
	}

	result := make([]ProjectResponse, 0, len(projects))
	for _, p := range projects {
		count, _ := s.repo.MemberCount(ctx, p.ID)
		result = append(result, *toProjectResponse(p, count, ""))
	}
	return result, nil
}

// AddMember adds a user to a project with the specified role.
// Invalidates the RBAC cache for the project.
func (s *ProjectService) AddMember(ctx context.Context, projectID, userID uuid.UUID, role string) (*MemberResponse, error) {
	prole := projectmember.ProjectRole(role)
	mem, err := s.repo.AddMember(ctx, projectID, userID, prole)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, apperrors.NewConflict("user is already a member of this project")
		}
		return nil, apperrors.NewInternal("failed to add member")
	}

	// Invalidate RBAC cache
	if s.cache != nil {
		if err := s.cache.InvalidateProject(ctx, projectID.String()); err != nil {
			slog.Error("failed to invalidate RBAC cache", "error", err, "project_id", projectID)
		}
	}

	return toMemberResponse(mem), nil
}

// RemoveMember removes a user from a project.
// Cannot remove the last owner of a project.
func (s *ProjectService) RemoveMember(ctx context.Context, projectID, userID uuid.UUID) error {
	// Check if user is an owner
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

	// Invalidate RBAC cache
	if s.cache != nil {
		if err := s.cache.InvalidateProject(ctx, projectID.String()); err != nil {
			slog.Error("failed to invalidate RBAC cache", "error", err, "project_id", projectID)
		}
	}

	return nil
}

// UpdateMemberRole changes a member's role in a project.
func (s *ProjectService) UpdateMemberRole(ctx context.Context, projectID, userID uuid.UUID, role string) (*MemberResponse, error) {
	prole := projectmember.ProjectRole(role)
	mem, err := s.repo.UpdateMemberRole(ctx, projectID, userID, prole)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("member not found")
		}
		return nil, apperrors.NewInternal("failed to update member role")
	}

	// Invalidate RBAC cache for the specific user
	if s.cache != nil {
		if err := s.cache.Invalidate(ctx, userID.String(), projectID.String()); err != nil {
			slog.Error("failed to invalidate RBAC cache", "error", err, "user_id", userID, "project_id", projectID)
		}
	}

	return toMemberResponse(mem), nil
}

// ListMembers returns all members of a project.
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

// toProjectResponse converts an Ent Project entity to a response DTO.
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

// toMemberResponse converts an Ent ProjectMember entity to a response DTO.
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

// toInviteResponse converts an Ent Invite entity to a response DTO.
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

// CreateTeam creates a new team and returns its response representation.
func (s *ProjectService) CreateTeam(ctx context.Context, name, description string) (*TeamResponse, error) {
	t, err := s.repo.CreateTeam(ctx, name, description)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, apperrors.NewConflict("team name already exists")
		}
		return nil, apperrors.NewInternal("failed to create team")
	}
	return toTeamResponse(t, 0), nil
}

// ListTeams returns all teams with their member counts.
func (s *ProjectService) ListTeams(ctx context.Context) ([]TeamResponse, error) {
	teams, err := s.repo.ListTeams(ctx)
	if err != nil {
		return nil, apperrors.NewInternal("failed to list teams")
	}
	result := make([]TeamResponse, 0, len(teams))
	for _, t := range teams {
		result = append(result, *toTeamResponse(t, len(t.Edges.Members)))
	}
	return result, nil
}

// AddTeamMember adds a user to a team.
func (s *ProjectService) AddTeamMember(ctx context.Context, teamID, userID uuid.UUID) error {
	_, err := s.repo.AddTeamMember(ctx, teamID, userID)
	if err != nil {
		if ent.IsConstraintError(err) {
			return apperrors.NewConflict("user is already a member of this team")
		}
		return apperrors.NewInternal("failed to add team member")
	}
	return nil
}

// RemoveTeamMember removes a user from a team.
func (s *ProjectService) RemoveTeamMember(ctx context.Context, teamID, userID uuid.UUID) error {
	if err := s.repo.RemoveTeamMember(ctx, teamID, userID); err != nil {
		return apperrors.NewInternal("failed to remove team member")
	}
	return nil
}

// AddTeamToProject links a team to a project and bulk-adds all team members as project members.
func (s *ProjectService) AddTeamToProject(ctx context.Context, teamID, projectID uuid.UUID) error {
	if err := s.repo.AddTeamToProject(ctx, teamID, projectID); err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("team or project not found")
		}
		return apperrors.NewInternal("failed to add team to project")
	}
	return nil
}

// toTeamResponse converts an Ent Team entity to a response DTO.
func toTeamResponse(t *ent.Team, memberCount int) *TeamResponse {
	return &TeamResponse{
		ID:          t.ID.String(),
		Name:        t.Name,
		Description: t.Description,
		MemberCount: memberCount,
	}
}

// ListInvites returns all invites for a project.
func (s *ProjectService) ListInvites(ctx context.Context, projectID uuid.UUID) ([]InviteResponse, error) {
	invites, err := s.repo.ListInvites(ctx, projectID)
	if err != nil {
		return nil, apperrors.NewInternal("failed to list invites")
	}
	result := make([]InviteResponse, 0, len(invites))
	for _, inv := range invites {
		result = append(result, *toInviteResponse(inv, ""))
	}
	return result, nil
}

// RevokeInvite deletes an invite.
func (s *ProjectService) RevokeInvite(ctx context.Context, inviteID uuid.UUID) error {
	return s.repo.DeleteInvite(ctx, inviteID)
}

