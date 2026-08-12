package tracker

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/issuelink"
	"github.com/valerii/hub/backend/ent/project"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// ---------------------------------------------------------------------------
// Связи задач
// ---------------------------------------------------------------------------

// CreateIssueLink создаёт направленную связь между двумя задачами.
// Дубликат = та же тройка (source, target, link_type); blocks+relates на одной паре допустимы (A2W8-3-003).
// Публикует outbox-событие project.{projectID}.issue.link_created.
func (r *EntRepository) CreateIssueLink(ctx context.Context, sourceID, targetID uuid.UUID, linkType string, customType string) (*ent.IssueLink, error) {
	exists, err := r.client.IssueLink.Query().
		Where(
			issuelink.HasSourceWith(issue.IDEQ(sourceID)),
			issuelink.HasTargetWith(issue.IDEQ(targetID)),
			issuelink.LinkTypeEQ(issuelink.LinkType(linkType)),
		).
		Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("проверка дубликата связи: %w", err)
	}
	if exists {
		return nil, apperrors.NewConflict("связь этого типа между этими задачами уже существует")
	}

	create := r.client.IssueLink.Create().
		SetSourceID(sourceID).
		SetTargetID(targetID).
		SetLinkType(issuelink.LinkType(linkType))

	if customType != "" {
		create = create.SetCustomType(customType)
	}

	link, err := create.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("создание связи: %w", err)
	}

	// Публикуем outbox-событие создания связи.
	if r.publisher != nil {
		source, qErr := r.client.Issue.Query().
			Where(issue.IDEQ(sourceID)).
			WithProject().
			Only(ctx)
		if qErr == nil && source.Edges.Project != nil {
			projectID := source.Edges.Project.ID.String()
			subj := fmt.Sprintf("project.%s.issue.link_created", projectID)
			logOutboxPublishErr(ctx, subj, r.publisher.Publish(ctx, r.client, subj, IssueLinkCreatedEvent{
				SourceID:  sourceID.String(),
				TargetID:  targetID.String(),
				LinkType:  linkType,
				ProjectID: projectID,
			}))
		}
	}

	return link, nil
}

// ListIssueLinks возвращает все связи (входящие и исходящие) для задачи.
func (r *EntRepository) ListIssueLinks(ctx context.Context, issueID uuid.UUID) ([]*ent.IssueLink, error) {
	// Исходящие связи
	outgoing, err := r.client.IssueLink.Query().
		Where(issuelink.HasSourceWith(issue.IDEQ(issueID))).
		WithTarget().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("получение исходящих связей: %w", err)
	}

	// Входящие связи
	incoming, err := r.client.IssueLink.Query().
		Where(issuelink.HasTargetWith(issue.IDEQ(issueID))).
		WithSource().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("получение входящих связей: %w", err)
	}

	// Объединяем
	all := make([]*ent.IssueLink, 0, len(outgoing)+len(incoming))
	all = append(all, outgoing...)
	all = append(all, incoming...)
	return all, nil
}

// MaxProjectIssueLinks — A13-004: hard cap batch-загрузки связей проекта для Gantt.
const MaxProjectIssueLinks = 2000

// ListProjectIssueLinks возвращает связи, где source или target принадлежит проекту.
func (r *EntRepository) ListProjectIssueLinks(ctx context.Context, projectID uuid.UUID) ([]*ent.IssueLink, error) {
	return r.client.IssueLink.Query().
		Where(
			issuelink.Or(
				issuelink.HasSourceWith(
					issue.HasProjectWith(project.IDEQ(projectID)),
					issue.DeletedAtIsNil(),
				),
				issuelink.HasTargetWith(
					issue.HasProjectWith(project.IDEQ(projectID)),
					issue.DeletedAtIsNil(),
				),
			),
		).
		WithSource().
		WithTarget().
		Limit(MaxProjectIssueLinks).
		All(ctx)
}

// DeleteIssueLink удаляет связь только если она принадлежит sourceIssueID (A3-011-link).
func (r *EntRepository) DeleteIssueLink(ctx context.Context, sourceIssueID, linkID uuid.UUID) error {
	affected, err := r.client.IssueLink.Delete().
		Where(
			issuelink.IDEQ(linkID),
			issuelink.HasSourceWith(issue.IDEQ(sourceIssueID)),
		).
		Exec(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return &ent.NotFoundError{}
	}
	return nil
}
