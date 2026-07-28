package project

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/invite"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/issuelink"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/projectmember"
	"github.com/valerii/hub/backend/ent/projecttemplate"
	"github.com/valerii/hub/backend/ent/team"
	"github.com/valerii/hub/backend/ent/teammember"
	"github.com/valerii/hub/backend/ent/user"
)

// Repository определяет интерфейс доступа к данным модуля project.
// Позволяет тестировать service-слой с mock-реализациями.
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
	MemberCountBatch(ctx context.Context, projectIDs []uuid.UUID) (map[uuid.UUID]int, error)
	ListMembers(ctx context.Context, projectID uuid.UUID) ([]*ent.ProjectMember, error)
	AddMember(ctx context.Context, projectID, userID uuid.UUID, role projectmember.ProjectRole) (*ent.ProjectMember, error)
	RemoveMember(ctx context.Context, projectID, userID uuid.UUID) error
	UpdateMemberRole(ctx context.Context, projectID, userID uuid.UUID, role projectmember.ProjectRole) (*ent.ProjectMember, error)
	FindMember(ctx context.Context, projectID, userID uuid.UUID) (*ent.ProjectMember, error)
	CountOwners(ctx context.Context, projectID uuid.UUID) (int, error)
	FindInviteByCode(ctx context.Context, code string) (*ent.Invite, error)
	CreateInvite(ctx context.Context, projectID, creatorID uuid.UUID, code string, role invite.ProjectRole, maxUses int, expiresAt *time.Time) (*ent.Invite, error)
	IncrementInviteUseCount(ctx context.Context, inviteID uuid.UUID) error
	// TryConsumeInvite атомарно резервирует одну единицу использования инвайта,
	// уважая лимит MaxUses (0 = без лимита). Возвращает true, если место было
	// успешно зарезервировано; false без ошибки означает, что лимит исчерпан
	// (race-condition safe — см. HI-06).
	TryConsumeInvite(ctx context.Context, inviteID uuid.UUID, maxUses int) (bool, error)
	ListInvites(ctx context.Context, projectID uuid.UUID) ([]*ent.Invite, error)
	// DeleteInvite удаляет invite только внутри projectID (P1-BE-06 IDOR).
	DeleteInvite(ctx context.Context, projectID, inviteID uuid.UUID) error

	// Операции с командами
	CreateTeam(ctx context.Context, name, description string) (*ent.Team, error)
	// ListTeamsByUserID возвращает команды, где userID — участник (P1-BE-08 ACL).
	ListTeamsByUserID(ctx context.Context, userID uuid.UUID) ([]*ent.Team, error)
	FindTeamByID(ctx context.Context, id uuid.UUID) (*ent.Team, error)
	FindTeamMember(ctx context.Context, teamID, userID uuid.UUID) (*ent.TeamMember, error)
	AddTeamMember(ctx context.Context, teamID, userID uuid.UUID) (*ent.TeamMember, error)
	RemoveTeamMember(ctx context.Context, teamID, userID uuid.UUID) error
	TeamMemberCount(ctx context.Context, teamID uuid.UUID) (int, error)
	ListTeamMembers(ctx context.Context, teamID uuid.UUID) ([]*ent.TeamMember, error)
	AddTeamToProject(ctx context.Context, teamID, projectID uuid.UUID) error

	// Очистка межпроектных связей задач (используется в DeleteProject)
	// ListOutgoingCrossProjectLinks возвращает связи, у которых source принадлежит
	// заданному проекту, а target — другому. Source и target preload'ятся,
	// у target дополнительно подгружается assignee для уведомлений.
	ListOutgoingCrossProjectLinks(ctx context.Context, projectID uuid.UUID) ([]*ent.IssueLink, error)
	// ListIncomingCrossProjectLinks возвращает связи, у которых target принадлежит
	// заданному проекту, а source — другому. Target и source preload'ятся,
	// у source дополнительно подгружается assignee для уведомлений.
	ListIncomingCrossProjectLinks(ctx context.Context, projectID uuid.UUID) ([]*ent.IssueLink, error)
	// DeleteIssueLink удаляет связь между задачами по ID.
	DeleteIssueLink(ctx context.Context, linkID uuid.UUID) error

	// Операции с шаблонами проектов (используются в TemplateService)
	ListTemplates(ctx context.Context) ([]*ent.ProjectTemplate, error)
	GetTemplate(ctx context.Context, id uuid.UUID) (*ent.ProjectTemplate, error)
	// GetProjectWithWorkflowStatusesAndLabels возвращает проект с preload'ом
	// workflow_statuses и labels (для сохранения проекта как шаблона).
	GetProjectWithWorkflowStatusesAndLabels(ctx context.Context, projectID uuid.UUID) (*ent.Project, []string, []string, error)
	CreateTemplate(ctx context.Context, name, description string, templateType projecttemplate.TemplateType, config map[string]interface{}, createdByID *uuid.UUID, isSystem bool) (*ent.ProjectTemplate, error)
	DeleteTemplate(ctx context.Context, id uuid.UUID) error
	CountSystemTemplates(ctx context.Context) (int, error)
	// IsSystemAdmin — founder/admin (P1-BE-09 DeleteTemplate).
	IsSystemAdmin(ctx context.Context, userID uuid.UUID) (bool, error)
}

// CacheInvalidator абстрагирует инвалидацию RBAC-кэша для тестируемости.
type CacheInvalidator interface {
	InvalidateProject(ctx context.Context, projectID string) error
	Invalidate(ctx context.Context, userID, projectID string) error
}

// OutboxPublisher абстрагирует публикацию outbox-событий для репозитория.
type OutboxPublisher interface {
	PublishInTx(ctx context.Context, tx *ent.Tx, eventType string, payload any) error
}

// EntRepository реализует Repository через Ent ORM с outbox-событиями.
type EntRepository struct {
	client    *ent.Client
	publisher OutboxPublisher
}

// NewEntRepository создаёт Ent-репозиторий project.
func NewEntRepository(client *ent.Client, publisher OutboxPublisher) *EntRepository {
	return &EntRepository{client: client, publisher: publisher}
}

// FindByID возвращает проект по ID.
func (r *EntRepository) FindByID(ctx context.Context, id uuid.UUID) (*ent.Project, error) {
	return r.client.Project.Get(ctx, id)
}

// FindByKey ищет проект по уникальному ключу (например, "HUB").
func (r *EntRepository) FindByKey(ctx context.Context, key string) (*ent.Project, error) {
	return r.client.Project.Query().
		Where(project.KeyEQ(key)).
		Only(ctx)
}

// ListByUserID возвращает проекты пользователя (исключая удалённые, опционально архивные).
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

// ListTrash возвращает удалённые проекты пользователя (status=deleted).
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

// Create создаёт проект в транзакции: запись проекта, добавление владельца, outbox-событие.
func (r *EntRepository) Create(ctx context.Context, name, key, description string, ownerID uuid.UUID) (*ent.Project, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, err
	}

	create := tx.Project.Create().
		SetName(name).
		SetKey(key).
		SetDescription(description)

	if owner, err := tx.User.Query().Where(user.IDEQ(ownerID)).WithOrganization().Only(ctx); err == nil {
		if org := owner.Edges.Organization; org != nil {
			create.SetOrganization(org)
		}
	}

	p, err := create.Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	// Добавляем создателя как владельца
	_, err = tx.ProjectMember.Create().
		SetProjectID(p.ID).
		SetUserID(ownerID).
		SetProjectRole(projectmember.ProjectRoleOwner).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	// Публикуем outbox-событие project.created
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

// Update обновляет поля проекта (имя, описание, аватар).
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

// UpdateStatus обновляет статус проекта и опционально устанавливает deleted_at.
func (r *EntRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status project.Status, setDeletedAt bool) (*ent.Project, error) {
	update := r.client.Project.UpdateOneID(id).SetStatus(status)
	if setDeletedAt {
		update = update.SetDeletedAt(time.Now().UTC())
	} else {
		update = update.ClearDeletedAt()
	}
	return update.Save(ctx)
}

// Delete выполняет жёсткое удаление проекта из базы данных.
func (r *EntRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.client.Project.DeleteOneID(id).Exec(ctx)
}

// MemberCount возвращает количество участников проекта.
func (r *EntRepository) MemberCount(ctx context.Context, projectID uuid.UUID) (int, error) {
	return r.client.ProjectMember.Query().
		Where(projectmember.HasProjectWith(project.IDEQ(projectID))).
		Count(ctx)
}

// MemberCountBatch возвращает количество участников для нескольких проектов одним запросом.
func (r *EntRepository) MemberCountBatch(ctx context.Context, projectIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	if len(projectIDs) == 0 {
		return make(map[uuid.UUID]int), nil
	}

	var rows []struct {
		ProjectID uuid.UUID `json:"project_members"`
		Count     int       `json:"count"`
	}
	err := r.client.ProjectMember.Query().
		Where(projectmember.HasProjectWith(project.IDIn(projectIDs...))).
		GroupBy(projectmember.ProjectColumn).
		Aggregate(ent.Count()).
		Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}

	counts := make(map[uuid.UUID]int, len(rows))
	for _, row := range rows {
		counts[row.ProjectID] = row.Count
	}
	return counts, nil
}

// ListMembers возвращает всех участников проекта с загруженными данными пользователя.
func (r *EntRepository) ListMembers(ctx context.Context, projectID uuid.UUID) ([]*ent.ProjectMember, error) {
	return r.client.ProjectMember.Query().
		Where(projectmember.HasProjectWith(project.IDEQ(projectID))).
		WithUser().
		All(ctx)
}

// AddMember добавляет пользователя в проект с указанной ролью и возвращает участника с данными пользователя.
func (r *EntRepository) AddMember(ctx context.Context, projectID, userID uuid.UUID, role projectmember.ProjectRole) (*ent.ProjectMember, error) {
	mem, err := r.client.ProjectMember.Create().
		SetProjectID(projectID).
		SetUserID(userID).
		SetProjectRole(role).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	// Reload с user edge для DTO ответа.
	return r.client.ProjectMember.Query().
		Where(projectmember.IDEQ(mem.ID)).
		WithUser().
		Only(ctx)
}

// RemoveMember удаляет пользователя из проекта.
func (r *EntRepository) RemoveMember(ctx context.Context, projectID, userID uuid.UUID) error {
	mem, err := r.FindMember(ctx, projectID, userID)
	if err != nil {
		return err
	}
	return r.client.ProjectMember.DeleteOneID(mem.ID).Exec(ctx)
}

// UpdateMemberRole обновляет роль участника проекта.
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

// FindMember ищет участника проекта по ID проекта и пользователя.
func (r *EntRepository) FindMember(ctx context.Context, projectID, userID uuid.UUID) (*ent.ProjectMember, error) {
	return r.client.ProjectMember.Query().
		Where(
			projectmember.HasProjectWith(project.IDEQ(projectID)),
			projectmember.HasUserWith(user.IDEQ(userID)),
		).
		WithUser().
		Only(ctx)
}

// CountOwners возвращает количество владельцев проекта.
func (r *EntRepository) CountOwners(ctx context.Context, projectID uuid.UUID) (int, error) {
	return r.client.ProjectMember.Query().
		Where(
			projectmember.HasProjectWith(project.IDEQ(projectID)),
			projectmember.ProjectRoleEQ(projectmember.ProjectRoleOwner),
		).
		Count(ctx)
}

// FindInviteByCode ищет активное приглашение по коду с загруженным проектом.
func (r *EntRepository) FindInviteByCode(ctx context.Context, code string) (*ent.Invite, error) {
	return r.client.Invite.Query().
		Where(
			invite.CodeEQ(code),
			invite.RevokedAtIsNil(),
		).
		WithProject().
		Only(ctx)
}

// CreateInvite создаёт приглашение в проект с указанной ролью и ограничениями.
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

// IncrementInviteUseCount увеличивает счётчик использований приглашения на 1.
func (r *EntRepository) IncrementInviteUseCount(ctx context.Context, inviteID uuid.UUID) error {
	return r.client.Invite.UpdateOneID(inviteID).
		AddUseCount(1).
		Exec(ctx)
}

// TryConsumeInvite атомарно увеличивает UseCount на 1 только если ещё не
// достигнут лимит MaxUses. Возвращает (true, nil), если место было успешно
// зарезервировано; (false, nil), если лимит уже достигнут (race с другим
// joiner'ом); (false, err) при ошибке БД.
//
// maxUses == 0 трактуется как "без лимита" — тогда инкремент всегда успешен.
// Это эквивалентно SQL: UPDATE invites SET use_count = use_count + 1
//
//	WHERE id = $1 AND (max_uses = 0 OR use_count < max_uses).
//
// Ent возвращает ErrNotFound, если ни одна строка не удовлетворила условию,
// и мы трактуем это как "лимит достигнут" (HI-06).
func (r *EntRepository) TryConsumeInvite(ctx context.Context, inviteID uuid.UUID, maxUses int) (bool, error) {
	upd := r.client.Invite.UpdateOneID(inviteID).
		Where(invite.RevokedAtIsNil()).
		AddUseCount(1).
		AddVersion(1)
	if maxUses > 0 {
		upd = upd.Where(invite.UseCountLT(maxUses))
	}
	_, err := upd.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ListInvites возвращает активные (не revoked) приглашения проекта.
func (r *EntRepository) ListInvites(ctx context.Context, projectID uuid.UUID) ([]*ent.Invite, error) {
	return r.client.Invite.Query().
		Where(
			invite.HasProjectWith(project.IDEQ(projectID)),
			invite.RevokedAtIsNil(),
		).
		All(ctx)
}

// DeleteInvite мягко отзывает приглашение (A2W8-3-002: revoked_at вместо hard delete).
func (r *EntRepository) DeleteInvite(ctx context.Context, projectID, inviteID uuid.UUID) error {
	n, err := r.client.Invite.Update().
		Where(
			invite.IDEQ(inviteID),
			invite.HasProjectWith(project.IDEQ(projectID)),
			invite.RevokedAtIsNil(),
		).
		SetRevokedAt(time.Now().UTC()).
		AddVersion(1).
		Save(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return &ent.NotFoundError{}
	}
	return nil
}

// CreateTeam создаёт команду с указанным именем и описанием.
func (r *EntRepository) CreateTeam(ctx context.Context, name, description string) (*ent.Team, error) {
	create := r.client.Team.Create().SetName(name)
	if description != "" {
		create = create.SetDescription(description)
	}
	return create.Save(ctx)
}

// ListTeamsByUserID возвращает команды, в которых состоит пользователь.
func (r *EntRepository) ListTeamsByUserID(ctx context.Context, userID uuid.UUID) ([]*ent.Team, error) {
	return r.client.Team.Query().
		Where(team.HasMembersWith(teammember.HasUserWith(user.IDEQ(userID)))).
		WithMembers().
		All(ctx)
}

// FindTeamByID возвращает команду по ID с eager-loaded участниками.
func (r *EntRepository) FindTeamByID(ctx context.Context, id uuid.UUID) (*ent.Team, error) {
	return r.client.Team.Query().
		Where(team.IDEQ(id)).
		WithMembers().
		Only(ctx)
}

// FindTeamMember возвращает членство пользователя в команде.
func (r *EntRepository) FindTeamMember(ctx context.Context, teamID, userID uuid.UUID) (*ent.TeamMember, error) {
	return r.client.TeamMember.Query().
		Where(
			teammember.HasTeamWith(team.IDEQ(teamID)),
			teammember.HasUserWith(user.IDEQ(userID)),
		).
		Only(ctx)
}

// AddTeamMember добавляет пользователя в команду и возвращает TeamMember.
func (r *EntRepository) AddTeamMember(ctx context.Context, teamID, userID uuid.UUID) (*ent.TeamMember, error) {
	return r.client.TeamMember.Create().
		SetTeamID(teamID).
		SetUserID(userID).
		Save(ctx)
}

// RemoveTeamMember удаляет пользователя из команды.
func (r *EntRepository) RemoveTeamMember(ctx context.Context, teamID, userID uuid.UUID) error {
	_, err := r.client.TeamMember.Delete().
		Where(
			teammember.HasTeamWith(team.IDEQ(teamID)),
			teammember.HasUserWith(user.IDEQ(userID)),
		).
		Exec(ctx)
	return err
}

// TeamMemberCount возвращает количество участников команды.
func (r *EntRepository) TeamMemberCount(ctx context.Context, teamID uuid.UUID) (int, error) {
	return r.client.TeamMember.Query().
		Where(teammember.HasTeamWith(team.IDEQ(teamID))).
		Count(ctx)
}

// ListTeamMembers возвращает всех участников команды с eager-loaded user edge.
func (r *EntRepository) ListTeamMembers(ctx context.Context, teamID uuid.UUID) ([]*ent.TeamMember, error) {
	return r.client.TeamMember.Query().
		Where(teammember.HasTeamWith(team.IDEQ(teamID))).
		WithUser().
		All(ctx)
}

// ListOutgoingCrossProjectLinks возвращает связи задач, где source — в данном
// проекте, а target — в другом. Edges Source/Target/Target.Assignee preload'ятся.
func (r *EntRepository) ListOutgoingCrossProjectLinks(ctx context.Context, projectID uuid.UUID) ([]*ent.IssueLink, error) {
	return r.client.IssueLink.Query().
		Where(
			issuelink.HasSourceWith(issue.HasProjectWith(project.IDEQ(projectID))),
			issuelink.HasTargetWith(issue.HasProjectWith(project.IDNEQ(projectID))),
		).
		WithSource().
		WithTarget(func(q *ent.IssueQuery) { q.WithAssignee() }).
		All(ctx)
}

// ListIncomingCrossProjectLinks возвращает связи задач, где target — в данном
// проекте, а source — в другом. Edges Target/Source/Source.Assignee preload'ятся.
func (r *EntRepository) ListIncomingCrossProjectLinks(ctx context.Context, projectID uuid.UUID) ([]*ent.IssueLink, error) {
	return r.client.IssueLink.Query().
		Where(
			issuelink.HasTargetWith(issue.HasProjectWith(project.IDEQ(projectID))),
			issuelink.HasSourceWith(issue.HasProjectWith(project.IDNEQ(projectID))),
		).
		WithTarget().
		WithSource(func(q *ent.IssueQuery) { q.WithAssignee() }).
		All(ctx)
}

// DeleteIssueLink удаляет связь между задачами по ID.
func (r *EntRepository) DeleteIssueLink(ctx context.Context, linkID uuid.UUID) error {
	return r.client.IssueLink.DeleteOneID(linkID).Exec(ctx)
}

// ListTemplates возвращает все шаблоны проектов, упорядоченные по дате создания.
func (r *EntRepository) ListTemplates(ctx context.Context) ([]*ent.ProjectTemplate, error) {
	return r.client.ProjectTemplate.Query().
		Order(ent.Asc(projecttemplate.FieldCreatedAt)).
		All(ctx)
}

// GetTemplate возвращает шаблон проекта по ID (с created_by для ownership).
func (r *EntRepository) GetTemplate(ctx context.Context, id uuid.UUID) (*ent.ProjectTemplate, error) {
	return r.client.ProjectTemplate.Query().
		Where(projecttemplate.IDEQ(id)).
		WithCreatedBy().
		Only(ctx)
}

// GetProjectWithWorkflowStatusesAndLabels возвращает проект и имена его
// workflow_statuses и labels — используется при сохранении проекта как шаблона.
// Возвращает (project, statusNames, labelNames, err).
// W1-A5-010 / P3-BE-01: ошибки загрузки статусов/меток — fail-closed (не пустой шаблон).
func (r *EntRepository) GetProjectWithWorkflowStatusesAndLabels(ctx context.Context, projectID uuid.UUID) (*ent.Project, []string, []string, error) {
	proj, err := r.client.Project.Get(ctx, projectID)
	if err != nil {
		return nil, nil, nil, err
	}

	statuses, err := r.client.Project.QueryWorkflowStatuses(proj).All(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("query workflow statuses: %w", err)
	}
	statusNames := make([]string, 0, len(statuses))
	for _, st := range statuses {
		statusNames = append(statusNames, st.Name)
	}

	labels, err := r.client.Project.QueryLabels(proj).All(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("query labels: %w", err)
	}
	labelNames := make([]string, 0, len(labels))
	for _, l := range labels {
		labelNames = append(labelNames, l.Name)
	}

	return proj, statusNames, labelNames, nil
}

// CreateTemplate создаёт шаблон проекта. createdByID может быть nil
// (для системных шаблонов), isSystem помечает шаблон как системный.
func (r *EntRepository) CreateTemplate(ctx context.Context, name, description string, templateType projecttemplate.TemplateType, config map[string]interface{}, createdByID *uuid.UUID, isSystem bool) (*ent.ProjectTemplate, error) {
	create := r.client.ProjectTemplate.Create().
		SetName(name).
		SetDescription(description).
		SetTemplateType(templateType).
		SetConfig(config).
		SetIsSystem(isSystem)
	if createdByID != nil {
		create = create.SetCreatedByID(*createdByID)
	}
	return create.Save(ctx)
}

// DeleteTemplate удаляет шаблон проекта по ID.
func (r *EntRepository) DeleteTemplate(ctx context.Context, id uuid.UUID) error {
	return r.client.ProjectTemplate.DeleteOneID(id).Exec(ctx)
}

// CountSystemTemplates возвращает количество системных шаблонов.
func (r *EntRepository) CountSystemTemplates(ctx context.Context) (int, error) {
	return r.client.ProjectTemplate.Query().
		Where(projecttemplate.IsSystem(true)).
		Count(ctx)
}

// IsSystemAdmin возвращает true для founder/admin.
func (r *EntRepository) IsSystemAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	u, err := r.client.User.Get(ctx, userID)
	if err != nil {
		return false, err
	}
	return u.SystemRole == user.SystemRoleFounder || u.SystemRole == user.SystemRoleAdmin, nil
}

// AddTeamToProject привязывает команду к проекту и bulk-добавляет участников.
func (r *EntRepository) AddTeamToProject(ctx context.Context, teamID, projectID uuid.UUID) error {
	// Привязываем команду к проекту
	err := r.client.Project.UpdateOneID(projectID).
		AddTeamIDs(teamID).
		Exec(ctx)
	if err != nil {
		return err
	}

	// Массово добавляем участников команды как участников проекта с ролью "developer"
	members, err := r.ListTeamMembers(ctx, teamID)
	if err != nil {
		return err
	}

	for _, m := range members {
		if m.Edges.User == nil {
			continue
		}
		// W1-A5-009 / P3-BE-01: игнорировать только constraint (уже участник);
		// прочие ошибки БД — fail-closed, не частичный membership без err клиенту.
		if _, err := r.AddMember(ctx, projectID, m.Edges.User.ID, projectmember.ProjectRoleDeveloper); err != nil {
			if ent.IsConstraintError(err) {
				continue
			}
			return fmt.Errorf("add team member %s to project: %w", m.Edges.User.ID, err)
		}
	}

	return nil
}
