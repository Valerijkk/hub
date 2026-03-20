package project

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/invite"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/projectmember"
	"github.com/valerii/hub/backend/ent/team"
	"github.com/valerii/hub/backend/ent/teammember"
	"github.com/valerii/hub/backend/ent/user"
)

// Repository defines the data access interface for the project module.
// This allows the service layer to be tested with mock implementations.
type Repository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*ent.Project, error)
	FindByKey(ctx context.Context, key string) (*ent.Project, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, includeArchived bool) ([]*ent.Project, error)
	ListTrash(ctx context.Context, userID uuid.UUID) ([]*ent.Project, error)
	Create(ctx context.Context, name, key, description string, ownerID uuid.UUID) (*ent.Project, error)
	Update(ctx context.Context, id uuid.UUID, name, description, avatarURL *string) (*ent.Project, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status project.Status, setDeletedAt bool) (*ent.Project, error)
	Delete(ctx context.Context, id uuid.UUID) error
	MemberCount(ctx context.Context, projectID uuid.UUID) (int, error)
	ListMembers(ctx context.Context, projectID uuid.UUID) ([]*ent.ProjectMember, error)
	AddMember(ctx context.Context, projectID, userID uuid.UUID, role projectmember.ProjectRole) (*ent.ProjectMember, error)
	RemoveMember(ctx context.Context, projectID, userID uuid.UUID) error
	UpdateMemberRole(ctx context.Context, projectID, userID uuid.UUID, role projectmember.ProjectRole) (*ent.ProjectMember, error)
	FindMember(ctx context.Context, projectID, userID uuid.UUID) (*ent.ProjectMember, error)
	CountOwners(ctx context.Context, projectID uuid.UUID) (int, error)
	FindInviteByCode(ctx context.Context, code string) (*ent.Invite, error)
	CreateInvite(ctx context.Context, projectID, creatorID uuid.UUID, code string, role invite.ProjectRole, maxUses int, expiresAt *time.Time) (*ent.Invite, error)
	IncrementInviteUseCount(ctx context.Context, inviteID uuid.UUID) error
	ListInvites(ctx context.Context, projectID uuid.UUID) ([]*ent.Invite, error)
	DeleteInvite(ctx context.Context, inviteID uuid.UUID) error

	// Team operations
	CreateTeam(ctx context.Context, name, description string) (*ent.Team, error)
	ListTeams(ctx context.Context) ([]*ent.Team, error)
	FindTeamByID(ctx context.Context, id uuid.UUID) (*ent.Team, error)
	AddTeamMember(ctx context.Context, teamID, userID uuid.UUID) (*ent.TeamMember, error)
	RemoveTeamMember(ctx context.Context, teamID, userID uuid.UUID) error
	TeamMemberCount(ctx context.Context, teamID uuid.UUID) (int, error)
	ListTeamMembers(ctx context.Context, teamID uuid.UUID) ([]*ent.TeamMember, error)
	AddTeamToProject(ctx context.Context, teamID, projectID uuid.UUID) error
}

// CacheInvalidator abstracts RBAC cache invalidation for testability.
type CacheInvalidator interface {
	InvalidateProject(ctx context.Context, projectID string) error
	Invalidate(ctx context.Context, userID, projectID string) error
}

// OutboxPublisher abstracts outbox event publishing for the repository.
type OutboxPublisher interface {
	PublishInTx(ctx context.Context, tx *ent.Tx, eventType string, payload any) error
}

// EntRepository implements Repository using Ent ORM with outbox event publishing.
type EntRepository struct {
	client    *ent.Client
	publisher OutboxPublisher
}

// NewEntRepository creates a new Ent-backed repository.
func NewEntRepository(client *ent.Client, publisher OutboxPublisher) *EntRepository {
	return &EntRepository{client: client, publisher: publisher}
}

func (r *EntRepository) FindByID(ctx context.Context, id uuid.UUID) (*ent.Project, error) {
	return r.client.Project.Get(ctx, id)
}

func (r *EntRepository) FindByKey(ctx context.Context, key string) (*ent.Project, error) {
	return r.client.Project.Query().
		Where(project.KeyEQ(key)).
		Only(ctx)
}

func (r *EntRepository) ListByUserID(ctx context.Context, userID uuid.UUID, includeArchived bool) ([]*ent.Project, error) {
	query := r.client.Project.Query().
		Where(
			project.StatusNEQ(project.StatusDeleted),
			project.HasMembersWith(
				projectmember.HasUserWith(user.IDEQ(userID)),
			),
		)

	if !includeArchived {
		query = query.Where(project.StatusNEQ(project.StatusArchived))
	}

	return query.All(ctx)
}

func (r *EntRepository) ListTrash(ctx context.Context, userID uuid.UUID) ([]*ent.Project, error) {
	return r.client.Project.Query().
		Where(
			project.StatusEQ(project.StatusDeleted),
			project.HasMembersWith(
				projectmember.HasUserWith(user.IDEQ(userID)),
			),
		).
		All(ctx)
}

func (r *EntRepository) Create(ctx context.Context, name, key, description string, ownerID uuid.UUID) (*ent.Project, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, err
	}

	p, err := tx.Project.Create().
		SetName(name).
		SetKey(key).
		SetDescription(description).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	// Add creator as owner
	_, err = tx.ProjectMember.Create().
		SetProjectID(p.ID).
		SetUserID(ownerID).
		SetProjectRole(projectmember.ProjectRoleOwner).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	// Publish outbox event for project.created
	if r.publisher != nil {
		err = r.publisher.PublishInTx(ctx, tx, "project.created", ProjectCreatedEvent{
			ProjectID:   p.ID.String(),
			Name:        p.Name,
			Key:         p.Key,
			CreatedByID: ownerID.String(),
		})
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return p, nil
}

func (r *EntRepository) Update(ctx context.Context, id uuid.UUID, name, description, avatarURL *string) (*ent.Project, error) {
	update := r.client.Project.UpdateOneID(id)
	if name != nil {
		update = update.SetName(*name)
	}
	if description != nil {
		update = update.SetDescription(*description)
	}
	if avatarURL != nil {
		update = update.SetAvatarURL(*avatarURL)
	}
	return update.Save(ctx)
}

func (r *EntRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status project.Status, setDeletedAt bool) (*ent.Project, error) {
	update := r.client.Project.UpdateOneID(id).SetStatus(status)
	if setDeletedAt {
		update = update.SetDeletedAt(time.Now())
	} else {
		update = update.ClearDeletedAt()
	}
	return update.Save(ctx)
}

func (r *EntRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.client.Project.DeleteOneID(id).Exec(ctx)
}

func (r *EntRepository) MemberCount(ctx context.Context, projectID uuid.UUID) (int, error) {
	return r.client.ProjectMember.Query().
		Where(projectmember.HasProjectWith(project.IDEQ(projectID))).
		Count(ctx)
}

func (r *EntRepository) ListMembers(ctx context.Context, projectID uuid.UUID) ([]*ent.ProjectMember, error) {
	return r.client.ProjectMember.Query().
		Where(projectmember.HasProjectWith(project.IDEQ(projectID))).
		WithUser().
		All(ctx)
}

func (r *EntRepository) AddMember(ctx context.Context, projectID, userID uuid.UUID, role projectmember.ProjectRole) (*ent.ProjectMember, error) {
	mem, err := r.client.ProjectMember.Create().
		SetProjectID(projectID).
		SetUserID(userID).
		SetProjectRole(role).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	// Reload with user edge
	return r.client.ProjectMember.Query().
		Where(projectmember.IDEQ(mem.ID)).
		WithUser().
		Only(ctx)
}

func (r *EntRepository) RemoveMember(ctx context.Context, projectID, userID uuid.UUID) error {
	mem, err := r.FindMember(ctx, projectID, userID)
	if err != nil {
		return err
	}
	return r.client.ProjectMember.DeleteOneID(mem.ID).Exec(ctx)
}

func (r *EntRepository) UpdateMemberRole(ctx context.Context, projectID, userID uuid.UUID, role projectmember.ProjectRole) (*ent.ProjectMember, error) {
	mem, err := r.FindMember(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	_, err = r.client.ProjectMember.UpdateOneID(mem.ID).
		SetProjectRole(role).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return r.client.ProjectMember.Query().
		Where(projectmember.IDEQ(mem.ID)).
		WithUser().
		Only(ctx)
}

func (r *EntRepository) FindMember(ctx context.Context, projectID, userID uuid.UUID) (*ent.ProjectMember, error) {
	return r.client.ProjectMember.Query().
		Where(
			projectmember.HasProjectWith(project.IDEQ(projectID)),
			projectmember.HasUserWith(user.IDEQ(userID)),
		).
		WithUser().
		Only(ctx)
}

func (r *EntRepository) CountOwners(ctx context.Context, projectID uuid.UUID) (int, error) {
	return r.client.ProjectMember.Query().
		Where(
			projectmember.HasProjectWith(project.IDEQ(projectID)),
			projectmember.ProjectRoleEQ(projectmember.ProjectRoleOwner),
		).
		Count(ctx)
}

func (r *EntRepository) FindInviteByCode(ctx context.Context, code string) (*ent.Invite, error) {
	return r.client.Invite.Query().
		Where(invite.CodeEQ(code)).
		WithProject().
		Only(ctx)
}

func (r *EntRepository) CreateInvite(ctx context.Context, projectID, creatorID uuid.UUID, code string, role invite.ProjectRole, maxUses int, expiresAt *time.Time) (*ent.Invite, error) {
	create := r.client.Invite.Create().
		SetProjectID(projectID).
		SetCreatedByID(creatorID).
		SetCode(code).
		SetProjectRole(role).
		SetMaxUses(maxUses)
	if expiresAt != nil {
		create = create.SetExpiresAt(*expiresAt)
	}
	return create.Save(ctx)
}

func (r *EntRepository) IncrementInviteUseCount(ctx context.Context, inviteID uuid.UUID) error {
	return r.client.Invite.UpdateOneID(inviteID).
		AddUseCount(1).
		Exec(ctx)
}

func (r *EntRepository) ListInvites(ctx context.Context, projectID uuid.UUID) ([]*ent.Invite, error) {
	return r.client.Invite.Query().
		Where(invite.HasProjectWith(project.IDEQ(projectID))).
		All(ctx)
}

func (r *EntRepository) DeleteInvite(ctx context.Context, inviteID uuid.UUID) error {
	return r.client.Invite.DeleteOneID(inviteID).Exec(ctx)
}

// CreateTeam creates a new team with the given name and description.
func (r *EntRepository) CreateTeam(ctx context.Context, name, description string) (*ent.Team, error) {
	create := r.client.Team.Create().SetName(name)
	if description != "" {
		create = create.SetDescription(description)
	}
	return create.Save(ctx)
}

// ListTeams returns all teams with their members eagerly loaded.
func (r *EntRepository) ListTeams(ctx context.Context) ([]*ent.Team, error) {
	return r.client.Team.Query().WithMembers().All(ctx)
}

// FindTeamByID returns a team by ID with members eagerly loaded.
func (r *EntRepository) FindTeamByID(ctx context.Context, id uuid.UUID) (*ent.Team, error) {
	return r.client.Team.Query().
		Where(team.IDEQ(id)).
		WithMembers().
		Only(ctx)
}

// AddTeamMember adds a user to a team, returning the created TeamMember.
func (r *EntRepository) AddTeamMember(ctx context.Context, teamID, userID uuid.UUID) (*ent.TeamMember, error) {
	return r.client.TeamMember.Create().
		SetTeamID(teamID).
		SetUserID(userID).
		Save(ctx)
}

// RemoveTeamMember removes a user from a team.
func (r *EntRepository) RemoveTeamMember(ctx context.Context, teamID, userID uuid.UUID) error {
	_, err := r.client.TeamMember.Delete().
		Where(
			teammember.HasTeamWith(team.IDEQ(teamID)),
			teammember.HasUserWith(user.IDEQ(userID)),
		).
		Exec(ctx)
	return err
}

// TeamMemberCount returns the number of members in a team.
func (r *EntRepository) TeamMemberCount(ctx context.Context, teamID uuid.UUID) (int, error) {
	return r.client.TeamMember.Query().
		Where(teammember.HasTeamWith(team.IDEQ(teamID))).
		Count(ctx)
}

// ListTeamMembers returns all members of a team with user edges eagerly loaded.
func (r *EntRepository) ListTeamMembers(ctx context.Context, teamID uuid.UUID) ([]*ent.TeamMember, error) {
	return r.client.TeamMember.Query().
		Where(teammember.HasTeamWith(team.IDEQ(teamID))).
		WithUser().
		All(ctx)
}

// AddTeamToProject links a team to a project and bulk-adds all team members as project members.
func (r *EntRepository) AddTeamToProject(ctx context.Context, teamID, projectID uuid.UUID) error {
	// Link team to project
	err := r.client.Project.UpdateOneID(projectID).
		AddTeamIDs(teamID).
		Exec(ctx)
	if err != nil {
		return err
	}

	// Bulk-add team members as project members with "developer" role
	members, err := r.ListTeamMembers(ctx, teamID)
	if err != nil {
		return err
	}

	for _, m := range members {
		if m.Edges.User == nil {
			continue
		}
		// Ignore constraint errors (user already a project member)
		_, _ = r.AddMember(ctx, projectID, m.Edges.User.ID, projectmember.ProjectRoleDeveloper)
	}

	return nil
}
