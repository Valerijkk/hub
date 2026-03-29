// Package messenger реализует мессенджер: каналы, сообщения, треды, реакции, присутствие, поиск.
package messenger

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/projectmember"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

// MessengerService реализует бизнес-логику модуля мессенджера.
// Оркестрирует операции с каналами, сообщениями, реакциями,
// публикует NATS-события для real-time доставки.
type MessengerService struct {
	repo     Repository
	nats     *natsclient.Client
	presence *PresenceService
	unread   *UnreadService
	db       *ent.Client
}

// NewMessengerService создаёт новый сервис мессенджера.
// db может быть nil (групповые упоминания @role не будут работать).
func NewMessengerService(repo Repository, nats *natsclient.Client, presence *PresenceService, unread *UnreadService, db *ent.Client) *MessengerService {
	return &MessengerService{repo: repo, nats: nats, presence: presence, unread: unread, db: db}
}

// ---------------------------------------------------------------------------
// Presence
// ---------------------------------------------------------------------------

// Heartbeat обновляет время активности пользователя.
func (s *MessengerService) Heartbeat(ctx context.Context, userID string) error {
	if s.presence == nil {
		return nil
	}
	return s.presence.Heartbeat(ctx, userID)
}

// GetPresence возвращает статусы присутствия для списка пользователей.
func (s *MessengerService) GetPresence(ctx context.Context, userIDs []string) (map[string]string, error) {
	if s.presence == nil {
		return map[string]string{}, nil
	}
	return s.presence.GetBulkStatus(ctx, userIDs)
}

// SetCustomStatus устанавливает кастомный текст статуса пользователя.
// duration — строковое значение длительности ("30m", "1h", "4h", "today", "1d", "1w", "0").
func (s *MessengerService) SetCustomStatus(ctx context.Context, userID, text, duration string) error {
	if s.presence == nil {
		return nil
	}
	return s.presence.SetCustomStatus(ctx, userID, text, duration)
}

// GetCustomStatus возвращает кастомный текст статуса пользователя.
func (s *MessengerService) GetCustomStatus(ctx context.Context, userID string) (string, error) {
	if s.presence == nil {
		return "", nil
	}
	return s.presence.GetCustomStatus(ctx, userID)
}

// ---------------------------------------------------------------------------
// Typing
// ---------------------------------------------------------------------------

// PublishTyping публикует событие набора текста в NATS (ephemeral, без записи в БД).
func (s *MessengerService) PublishTyping(ctx context.Context, channelID string, userID uuid.UUID) error {
	// Определяем отображаемое имя пользователя.
	userName := "user"
	if s.db != nil {
		u, err := s.db.User.Get(ctx, userID)
		if err == nil {
			if u.DisplayName != "" {
				userName = u.DisplayName
			} else if u.Name != "" {
				userName = u.Name
			}
		}
	}

	subject := fmt.Sprintf(SubjectTyping, channelID)
	s.nats.PublishJSON(subject, TypingEvent{
		ChannelID: channelID,
		UserID:    userID.String(),
		UserName:  userName,
	})
	return nil
}

// ---------------------------------------------------------------------------
// Unread
// ---------------------------------------------------------------------------

// GetUnreadCounts возвращает счётчики непрочитанных для пользователя.
func (s *MessengerService) GetUnreadCounts(ctx context.Context, userID string) (map[string]int64, error) {
	if s.unread == nil {
		return map[string]int64{}, nil
	}
	return s.unread.GetAll(ctx, userID)
}

// MarkChannelRead сбрасывает счётчик непрочитанных для канала.
func (s *MessengerService) MarkChannelRead(ctx context.Context, userID, channelID string) error {
	if s.unread == nil {
		return nil
	}
	return s.unread.MarkRead(ctx, userID, channelID)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// resolveGroupMentions разрешает групповые упоминания (@all, @online, @role)
// в конкретные user ID. Объединяет с индивидуальными упоминаниями из поля Mentions.
// Исключает отправителя из результата.
func (s *MessengerService) resolveGroupMentions(ctx context.Context, channelID uuid.UUID, projectID *uuid.UUID, groups []string, individualMentions []string, senderID uuid.UUID) []string {
	seen := make(map[string]struct{})
	senderStr := senderID.String()

	// Добавляем индивидуальные упоминания.
	for _, uid := range individualMentions {
		if uid != senderStr {
			seen[uid] = struct{}{}
		}
	}

	if len(groups) == 0 {
		return mapKeys(seen)
	}

	// Получаем участников канала (нужно для @all и пересечения с @online).
	memberIDs, err := s.repo.GetChannelMemberIDs(ctx, channelID)
	if err != nil {
		slog.Error("не удалось получить участников канала для упоминаний", "error", err)
		return mapKeys(seen)
	}

	memberSet := make(map[string]struct{}, len(memberIDs))
	for _, m := range memberIDs {
		memberSet[m.String()] = struct{}{}
	}

	for _, g := range groups {
		switch g {
		case "all":
			// @all — все участники канала.
			for _, m := range memberIDs {
				uid := m.String()
				if uid != senderStr {
					seen[uid] = struct{}{}
				}
			}

		case "online":
			// @online — онлайн-пользователи, пересечённые с участниками канала.
			if s.presence == nil {
				continue
			}
			onlineIDs, pErr := s.presence.GetOnlineUserIDs(ctx)
			if pErr != nil {
				slog.Error("не удалось получить онлайн-пользователей", "error", pErr)
				continue
			}
			for _, oid := range onlineIDs {
				if _, ok := memberSet[oid]; ok && oid != senderStr {
					seen[oid] = struct{}{}
				}
			}

		default:
			// Предполагаем имя роли проекта (owner, developer, tester, viewer).
			if s.db == nil || projectID == nil {
				continue
			}
			members, qErr := s.db.ProjectMember.Query().
				Where(
					projectmember.HasProjectWith(project.IDEQ(*projectID)),
					projectmember.ProjectRoleEQ(projectmember.ProjectRole(g)),
				).
				WithUser().
				All(ctx)
			if qErr != nil {
				slog.Error("не удалось получить участников по роли", "role", g, "error", qErr)
				continue
			}
			for _, m := range members {
				if m.Edges.User == nil {
					continue
				}
				uid := m.Edges.User.ID.String()
				// Учитываем только тех, кто является участником канала.
				if _, ok := memberSet[uid]; ok && uid != senderStr {
					seen[uid] = struct{}{}
				}
			}
		}
	}

	return mapKeys(seen)
}

// mapKeys возвращает ключи map[string]struct{} как срез строк.
func mapKeys(m map[string]struct{}) []string {
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}

// toChannelResponse конвертирует Ent Channel в DTO ответа.
func toChannelResponse(ch *ent.Channel) *ChannelResponse {
	resp := &ChannelResponse{
		ID:          ch.ID.String(),
		ChannelType: ch.ChannelType.String(),
		Name:        ch.Name,
		Description: ch.Description,
		IsReadonly:  ch.IsReadonly,
		CreatedBy:   ch.CreatedBy.String(),
		CreatedAt:   ch.CreatedAt,
		UpdatedAt:   ch.UpdatedAt,
	}
	if ch.ProjectID != nil {
		resp.ProjectID = ch.ProjectID.String()
	}
	return resp
}
