package tracker

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/issuelink"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// ---------------------------------------------------------------------------
// Связи задач
// ---------------------------------------------------------------------------

// CreateIssueLink создаёт направленную связь между двумя задачами.
// Проверяет отсутствие дубликата. Публикует outbox-событие project.{projectID}.issue.link_created.
func (r *EntRepository) CreateIssueLink(ctx context.Context, sourceID, targetID uuid.UUID, linkType string, customType string) (*ent.IssueLink, error) {
	// Проверяем дубликат
	exists, err := r.client.IssueLink.Query().
		Where(
			issuelink.HasSourceWith(issue.IDEQ(sourceID)),
			issuelink.HasTargetWith(issue.IDEQ(targetID)),
		).
		Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("проверка дубликата связи: %w", err)
	}
	if exists {
		return nil, apperrors.NewConflict("связь между этими задачами уже существует")
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
			_ = r.publisher.Publish(ctx, r.client, fmt.Sprintf("project.%s.issue.link_created", projectID), IssueLinkCreatedEvent{
				SourceID:  sourceID.String(),
				TargetID:  targetID.String(),
				LinkType:  linkType,
				ProjectID: projectID,
			})
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
