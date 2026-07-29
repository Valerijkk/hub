package tracker

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)
// ThreadMirror зеркалирует комментарии задачи в связанные треды мессенджера.
type ThreadMirror interface {
	MirrorIssueComment(ctx context.Context, channelID, senderID uuid.UUID, threadRootMessageID, text string) error
}

// SetThreadMirror подключает адаптер мессенджера для зеркалирования комментариев в тред.
func (s *TrackerService) SetThreadMirror(m ThreadMirror) {
	s.threadMirror = m
}

// IssueMessageLinkResponse — DTO связи задача↔тред для API.
type IssueMessageLinkResponse struct {
	ID              string    `json:"id"`
	IssueID         string    `json:"issue_id"`
	IssueKey        string    `json:"issue_key,omitempty"`
	IssueTitle      string    `json:"issue_title,omitempty"`
	ChannelID       string    `json:"channel_id"`
	MessageID       string    `json:"message_id"`
	MirrorToThread  bool      `json:"mirror_to_thread"`
	CreatedByID     string    `json:"created_by_id"`
	CreatedAt       time.Time `json:"created_at"`
}

// CreateIssueMessageLinkRequest — привязка задачи к корневому сообщению треда.
type CreateIssueMessageLinkRequest struct {
	ChannelID      string `json:"channel_id" validate:"required"`
	MessageID      string `json:"message_id" validate:"required"`
	MirrorToThread *bool  `json:"mirror_to_thread,omitempty"`
}

func toIssueMessageLinkResponse(link *ent.IssueMessageLink) IssueMessageLinkResponse {
	resp := IssueMessageLinkResponse{
		ID:             link.ID.String(),
		IssueID:        link.IssueID.String(),
		ChannelID:      link.ChannelID.String(),
		MessageID:      link.MessageID,
		MirrorToThread: link.MirrorToThread,
		CreatedByID:    link.CreatedByID.String(),
		CreatedAt:      link.CreatedAt,
	}
	if link.Edges.Issue != nil {
		resp.IssueKey = link.Edges.Issue.IssueKey
		resp.IssueTitle = link.Edges.Issue.Title
	}
	return resp
}

// CreateIssueMessageLink связывает задачу с корневым сообщением треда мессенджера.
func (s *TrackerService) CreateIssueMessageLink(
	ctx context.Context,
	projectID uuid.UUID,
	issueKey string,
	userID uuid.UUID,
	req CreateIssueMessageLinkRequest,
) (*IssueMessageLinkResponse, error) {
	issue, err := s.repo.FindIssueByKey(ctx, projectID, issueKey)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("задача не найдена")
		}
		return nil, apperrors.NewInternal("ошибка получения задачи")
	}
	if issue.Edges.Project == nil || issue.Edges.Project.ID != projectID {
		return nil, apperrors.NewNotFound("задача не найдена")
	}

	channelID, err := uuid.Parse(strings.TrimSpace(req.ChannelID))
	if err != nil {
		return nil, apperrors.NewBadRequest("некорректный channel_id")
	}
	messageID := strings.TrimSpace(req.MessageID)
	if messageID == "" {
		return nil, apperrors.NewBadRequest("message_id обязателен")
	}

	// P1-BE-04: канал должен принадлежать проекту; пользователь — member канала.
	if _, err := s.repo.FindChannelInProject(ctx, projectID, channelID); err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("канал не найден")
		}
		return nil, apperrors.NewInternal("ошибка проверки канала")
	}
	ok, err := s.repo.IsChannelMember(ctx, channelID, userID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка проверки membership канала")
	}
	if !ok {
		return nil, apperrors.NewForbidden("нет доступа к каналу")
	}

	mirror := true
	if req.MirrorToThread != nil {
		mirror = *req.MirrorToThread
	}

	link, err := s.repo.CreateIssueMessageLink(ctx, issue.ID, channelID, userID, messageID, mirror)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка создания связи issue↔thread")
	}
	link.Edges.Issue = issue
	resp := toIssueMessageLinkResponse(link)
	return &resp, nil
}

// LinkThreadToIssue реализует messenger.IssueThreadLinker (автосвязь при promote).
func (s *TrackerService) LinkThreadToIssue(
	ctx context.Context,
	projectID, issueID, userID uuid.UUID,
	channelID uuid.UUID,
	threadRootMessageID string,
) error {
	issue, err := s.repo.FindIssueByID(ctx, issueID)
	if err != nil {
		return err
	}
	if issue.Edges.Project == nil || issue.Edges.Project.ID != projectID {
		return fmt.Errorf("issue not in project")
	}
	// P1-BE-04: канал должен принадлежать проекту (promote-path).
	if _, err := s.repo.FindChannelInProject(ctx, projectID, channelID); err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("канал не найден")
		}
		return fmt.Errorf("check channel in project: %w", err)
	}
	_, err = s.repo.CreateIssueMessageLink(ctx, issueID, channelID, userID, threadRootMessageID, true)
	if err != nil && ent.IsConstraintError(err) {
		return nil
	}
	return err
}

// ListIssueMessageLinks возвращает связи для задачи.
func (s *TrackerService) ListIssueMessageLinks(ctx context.Context, projectID uuid.UUID, issueKey string) ([]IssueMessageLinkResponse, error) {
	issue, err := s.repo.FindIssueByKey(ctx, projectID, issueKey)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("задача не найдена")
		}
		return nil, apperrors.NewInternal("ошибка получения задачи")
	}
	if issue.Edges.Project == nil || issue.Edges.Project.ID != projectID {
		return nil, apperrors.NewNotFound("задача не найдена")
	}

	links, err := s.repo.ListIssueMessageLinksByIssue(ctx, issue.ID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения связей")
	}
	out := make([]IssueMessageLinkResponse, 0, len(links))
	for _, l := range links {
		l.Edges.Issue = issue
		out = append(out, toIssueMessageLinkResponse(l))
	}
	return out, nil
}

// ListThreadIssueLinks возвращает связи для корневого сообщения треда.
func (s *TrackerService) ListThreadIssueLinks(
	ctx context.Context,
	projectID uuid.UUID,
	channelID uuid.UUID,
	messageID string,
) ([]IssueMessageLinkResponse, error) {
	links, err := s.repo.ListIssueMessageLinksByThread(ctx, channelID, messageID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения связей")
	}
	out := make([]IssueMessageLinkResponse, 0, len(links))
	for _, l := range links {
		// A3W1-2-004: Issue подгружается batch/eager в ListIssueMessageLinksByThread.
		if l.Edges.Issue == nil {
			continue
		}
		if l.Edges.Issue.Edges.Project == nil || l.Edges.Issue.Edges.Project.ID != projectID {
			continue
		}
		out = append(out, toIssueMessageLinkResponse(l))
	}
	return out, nil
}

// DeleteIssueMessageLink удаляет связь задача↔тред.
func (s *TrackerService) DeleteIssueMessageLink(ctx context.Context, projectID uuid.UUID, issueKey string, linkID uuid.UUID) error {
	issue, err := s.repo.FindIssueByKey(ctx, projectID, issueKey)
	if err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("задача не найдена")
		}
		return apperrors.NewInternal("ошибка получения задачи")
	}
	if issue.Edges.Project == nil || issue.Edges.Project.ID != projectID {
		return apperrors.NewNotFound("задача не найдена")
	}
	if err := s.repo.DeleteIssueMessageLink(ctx, linkID, issue.ID); err != nil {
		return apperrors.NewNotFound("связь не найдена")
	}
	return nil
}

// mirrorIssueCommentsToThreads синхронизирует новый комментарий в связанные треды.
func (s *TrackerService) mirrorIssueCommentsToThreads(
	ctx context.Context,
	issueID, userID uuid.UUID,
	bodyText string,
) {
	if s.threadMirror == nil || bodyText == "" {
		return
	}
	if strings.HasPrefix(bodyText, "[Hub sync]") {
		return
	}
	links, err := s.repo.ListIssueMessageLinksByIssue(ctx, issueID)
	if err != nil {
		// W1-A2-013 / P3-BE-01: не глотать сбой списка связей без сигнала в ops.
		slog.WarnContext(ctx, "mirror comments: list issue-message links failed",
			"issue_id", issueID, "error", err)
		return
	}
	text := fmt.Sprintf("[Hub sync] Комментарий к задаче:\n%s", bodyText)
	for _, link := range links {
		if !link.MirrorToThread {
			continue
		}
		if err := s.threadMirror.MirrorIssueComment(ctx, link.ChannelID, userID, link.MessageID, text); err != nil {
			slog.WarnContext(ctx, "mirror comments: thread mirror failed",
				"issue_id", issueID, "channel_id", link.ChannelID, "error", err)
			continue
		}
	}
}
