// Package notification реализует систему уведомлений: in-app, Telegram, Web Push.
package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/issue"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

// NotificationService реализует бизнес-логику уведомлений:
// проверка предпочтений, тихие часы, сохранение и доставка.
type NotificationService struct {
	repo     *Repository
	nats     *natsclient.Client
	telegram *TelegramSender
	push     *PushSender
	db       *ent.Client
}

// NewNotificationService создаёт новый сервис уведомлений.
// push может быть nil (Web Push уведомления не будут отправляться).
func NewNotificationService(repo *Repository, nc *natsclient.Client, telegram *TelegramSender, db *ent.Client, push *PushSender) *NotificationService {
	return &NotificationService{
		repo:     repo,
		nats:     nc,
		telegram: telegram,
		db:       db,
		push:     push,
	}
}

// PushSender возвращает отправителя push-уведомлений (для handler).
func (s *NotificationService) PushSender() *PushSender {
	return s.push
}

// DispatchEvent описывает уведомление, которое нужно доставить пользователю.
//
// LO-2: до рефакторинга Dispatch принимал 10 позиционных аргументов
// (userID, notifType, title, body, entityType, entityID, projectID, actorID,
// isUrgent + ctx). Это делало callsite'ы хрупкими (легко перепутать порядок
// одинаково-типовых string-полей, например title/body или entityType/entityID).
// Struct делает каждое поле именованным и позволяет легко добавлять новые
// поля без ломки сигнатуры.
type DispatchEvent struct {
	// UserID — получатель уведомления.
	UserID uuid.UUID
	// Type — тип уведомления (TypeAssigned, TypeMentioned и т. п.).
	Type string
	// Title — заголовок (показывается в push/telegram/in-app).
	Title string
	// Body — тело сообщения (может быть пустым).
	Body string
	// EntityType — тип связанной сущности ("issue", "channel", "testrun", "call").
	EntityType string
	// EntityID — ID связанной сущности (string, т.к. не всегда UUID).
	EntityID string
	// ProjectID — необязательная привязка к проекту (используется в IsEnabled).
	ProjectID *uuid.UUID
	// ActorID — пользователь, инициировавший событие (может быть nil для системных).
	ActorID *uuid.UUID
	// IsUrgent — флаг "срочное", обходит тихие часы.
	IsUrgent bool
}

// Dispatch создаёт и отправляет уведомление пользователю.
// Проверяет предпочтения и тихие часы перед отправкой.
func (s *NotificationService) Dispatch(ctx context.Context, evt DispatchEvent) error {
	// Проверяем предпочтения in_app.
	inAppEnabled, err := s.repo.IsEnabled(ctx, evt.UserID, evt.Type, ChannelInApp, evt.ProjectID)
	if err != nil {
		slog.Error("ошибка проверки предпочтений", "error", err, "user_id", evt.UserID)
		return err
	}

	// Проверяем тихие часы.
	inQuietHours := s.isQuietHours(ctx, evt.UserID)
	if inQuietHours && !evt.IsUrgent {
		slog.Debug("уведомление подавлено тихими часами", "user_id", evt.UserID, "type", evt.Type)
		return nil
	}

	// Сохраняем уведомление в БД если in_app включён.
	var notif *ent.Notification
	if inAppEnabled {
		notif, err = s.repo.CreateNotification(ctx, evt.UserID, evt.Type, evt.Title, evt.Body,
			evt.EntityType, evt.EntityID, evt.ProjectID, evt.ActorID, evt.IsUrgent)
		if err != nil {
			slog.Error("ошибка создания уведомления", "error", err, "user_id", evt.UserID)
			return err
		}
		if notif == nil {
			// Дедупликация: пропущено.
			return nil
		}

		// Отправляем через NATS -> WS bridge.
		s.publishToWS(evt.UserID, notif)
	}

	// Отправляем в Telegram если канал включён и у пользователя есть telegram_id.
	telegramEnabled, err := s.repo.IsEnabled(ctx, evt.UserID, evt.Type, ChannelTelegram, evt.ProjectID)
	if err != nil {
		slog.Error("ошибка проверки предпочтений telegram", "error", err, "user_id", evt.UserID)
	}
	if telegramEnabled && s.telegram != nil {
		s.sendTelegram(ctx, evt.UserID, evt.Title, evt.Body)
	}

	// Отправляем через Web Push если канал включён и есть подписки.
	pushEnabled, err := s.repo.IsEnabled(ctx, evt.UserID, evt.Type, ChannelPush, evt.ProjectID)
	if err != nil {
		slog.Error("ошибка проверки предпочтений push", "error", err, "user_id", evt.UserID)
	}
	if pushEnabled && s.push != nil {
		// URL для навигации при клике на уведомление.
		pushURL := buildPushURL(evt.EntityType, evt.EntityID)
		s.push.Send(evt.UserID, evt.Title, evt.Body, pushURL)
	}

	return nil
}

// buildPushURL формирует URL для навигации при клике на push-уведомление.
func buildPushURL(entityType, entityID string) string {
	switch entityType {
	case "issue":
		return "/tracker/issues/" + entityID
	case "channel":
		return "/messenger/channels/" + entityID
	case "testrun":
		return "/tms/runs/" + entityID
	case "call":
		return "/calls/" + entityID
	default:
		return "/"
	}
}

// publishToWS публикует уведомление в NATS для доставки через WS bridge.
func (s *NotificationService) publishToWS(userID uuid.UUID, notif *ent.Notification) {
	if s.nats == nil {
		return
	}

	payload := WsNotificationPayload{
		ID:         notif.ID,
		Type:       notif.Type,
		Title:      notif.Title,
		Body:       notif.Body,
		EntityType: notif.EntityType,
		EntityID:   notif.EntityID,
		IsUrgent:   notif.IsUrgent,
		CreatedAt:  notif.CreatedAt,
	}
	if notif.ProjectID != nil {
		payload.ProjectID = notif.ProjectID
	}
	if notif.ActorID != nil {
		payload.ActorID = notif.ActorID
	}

	data, err := json.Marshal(payload)
	if err != nil {
		slog.Error("ошибка сериализации WS payload", "error", err)
		return
	}

	subject := fmt.Sprintf("notification.user.%s", userID.String())
	if err := s.nats.Publish(subject, data); err != nil {
		slog.Error("ошибка публикации в NATS", "error", err, "subject", subject)
	}
}

// sendTelegram отправляет уведомление в Telegram.
func (s *NotificationService) sendTelegram(ctx context.Context, userID uuid.UUID, title, body string) {
	user, err := s.db.User.Get(ctx, userID)
	if err != nil {
		slog.Error("ошибка загрузки пользователя для telegram", "error", err, "user_id", userID)
		return
	}

	if user.TelegramID == nil || *user.TelegramID == "" {
		return
	}

	chatID, err := strconv.ParseInt(*user.TelegramID, 10, 64)
	if err != nil {
		slog.Error("невалидный telegram_id", "error", err, "telegram_id", *user.TelegramID)
		return
	}

	s.telegram.Send(chatID, title, body)
}

// isQuietHours проверяет, находится ли текущее время в тихих часах пользователя.
func (s *NotificationService) isQuietHours(ctx context.Context, userID uuid.UUID) bool {
	user, err := s.db.User.Get(ctx, userID)
	if err != nil {
		return false
	}

	if user.QuietHoursStart == nil || user.QuietHoursEnd == nil {
		return false
	}

	loc, err := time.LoadLocation(user.Timezone)
	if err != nil {
		loc = time.UTC
	}

	now := time.Now().In(loc)
	currentMinutes := now.Hour()*60 + now.Minute()

	startMinutes, err := parseTimeMinutes(*user.QuietHoursStart)
	if err != nil {
		return false
	}
	endMinutes, err := parseTimeMinutes(*user.QuietHoursEnd)
	if err != nil {
		return false
	}

	// Обработка перехода через полночь (например, 23:00 - 07:00).
	if startMinutes <= endMinutes {
		return currentMinutes >= startMinutes && currentMinutes < endMinutes
	}
	return currentMinutes >= startMinutes || currentMinutes < endMinutes
}

// CheckDeadlines находит задачи с приближающимся дедлайном (24ч) и отправляет уведомления.
// Запускается по крону из main.go.
func (s *NotificationService) CheckDeadlines(ctx context.Context) error {
	now := time.Now()
	deadline := now.Add(24 * time.Hour)

	// Запрашиваем задачи с due_date, не удалённые, с assignee и project.
	issues, err := s.db.Issue.Query().
		Where(
			issue.DueDateNotNil(),
			issue.DueDateGTE(now),
			issue.DueDateLTE(deadline),
			issue.DeletedAtIsNil(),
		).
		WithAssignee().
		WithProject().
		All(ctx)
	if err != nil {
		return fmt.Errorf("ошибка запроса задач: %w", err)
	}

	count := 0
	for _, iss := range issues {
		assignee := iss.Edges.Assignee
		if assignee == nil {
			continue
		}
		proj := iss.Edges.Project
		var projectID *uuid.UUID
		if proj != nil {
			projectID = &proj.ID
		}
		title := fmt.Sprintf("Дедлайн приближается: %s", iss.Title)
		body := fmt.Sprintf("Задача #%d должна быть завершена до %s",
			iss.IssueNumber, iss.DueDate.Format("02.01.2006 15:04"))

		if dispErr := s.Dispatch(ctx, DispatchEvent{
			UserID:     assignee.ID,
			Type:       TypeDeadlineApproaching,
			Title:      title,
			Body:       body,
			EntityType: "issue",
			EntityID:   iss.ID.String(),
			ProjectID:  projectID,
		}); dispErr != nil {
			slog.Error("ошибка отправки deadline уведомления", "error", dispErr, "issue_id", iss.ID)
		}
		count++
	}
	slog.Info("проверка дедлайнов завершена", "notifications_sent", count)
	return nil
}

// parseTimeMinutes парсит строку "HH:MM" в минуты с начала дня.
func parseTimeMinutes(s string) (int, error) {
	t, err := time.Parse("15:04", s)
	if err != nil {
		return 0, err
	}
	return t.Hour()*60 + t.Minute(), nil
}

// ListNotifications возвращает пагинированный список уведомлений пользователя.
func (s *NotificationService) ListNotifications(ctx context.Context, userID uuid.UUID, page, perPage int, isRead *bool) (*NotificationListResponse, error) {
	items, total, err := s.repo.ListByUser(ctx, userID, page, perPage, isRead)
	if err != nil {
		return nil, err
	}

	dtos := make([]NotificationDTO, 0, len(items))
	for _, n := range items {
		dtos = append(dtos, notifToDTO(n))
	}

	return &NotificationListResponse{
		Notifications: dtos,
		Total:         total,
		Page:          page,
		PerPage:       perPage,
	}, nil
}

// MarkRead помечает уведомление как прочитанное.
func (s *NotificationService) MarkRead(ctx context.Context, notifID, userID uuid.UUID) error {
	return s.repo.MarkRead(ctx, notifID, userID)
}

// MarkAllRead помечает все уведомления пользователя как прочитанные.
func (s *NotificationService) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	return s.repo.MarkAllRead(ctx, userID)
}

// DeleteNotification удаляет уведомление пользователя.
func (s *NotificationService) DeleteNotification(ctx context.Context, notifID, userID uuid.UUID) error {
	return s.repo.DeleteNotification(ctx, notifID, userID)
}

// GetUnreadCount возвращает количество непрочитанных уведомлений.
func (s *NotificationService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	return s.repo.CountUnread(ctx, userID)
}

// GetPreferences возвращает настройки уведомлений пользователя.
func (s *NotificationService) GetPreferences(ctx context.Context, userID uuid.UUID) ([]PreferenceDTO, error) {
	prefs, err := s.repo.GetPreferences(ctx, userID)
	if err != nil {
		return nil, err
	}

	dtos := make([]PreferenceDTO, 0, len(prefs))
	for _, p := range prefs {
		dto := PreferenceDTO{
			ID:               p.ID,
			NotificationType: p.NotificationType,
			Channel:          p.Channel,
			Enabled:          p.Enabled,
		}
		if p.ProjectID != nil {
			dto.ProjectID = p.ProjectID
		}
		dtos = append(dtos, dto)
	}
	return dtos, nil
}

// UpdatePreference создаёт или обновляет настройку уведомлений.
func (s *NotificationService) UpdatePreference(ctx context.Context, userID uuid.UUID, req PreferenceUpdateRequest) (*PreferenceDTO, error) {
	pref, err := s.repo.UpsertPreference(ctx, userID, req.NotificationType, req.Channel, req.Enabled, req.ProjectID)
	if err != nil {
		return nil, err
	}

	dto := &PreferenceDTO{
		ID:               pref.ID,
		NotificationType: pref.NotificationType,
		Channel:          pref.Channel,
		Enabled:          pref.Enabled,
	}
	if pref.ProjectID != nil {
		dto.ProjectID = pref.ProjectID
	}
	return dto, nil
}

// notifToDTO конвертирует Ent Notification в DTO.
func notifToDTO(n *ent.Notification) NotificationDTO {
	dto := NotificationDTO{
		ID:         n.ID,
		Type:       n.Type,
		Title:      n.Title,
		Body:       n.Body,
		EntityType: n.EntityType,
		EntityID:   n.EntityID,
		IsRead:     n.IsRead,
		IsUrgent:   n.IsUrgent,
		CreatedAt:  n.CreatedAt,
	}
	if n.ProjectID != nil {
		dto.ProjectID = n.ProjectID
	}
	if n.ActorID != nil {
		dto.ActorID = n.ActorID
	}
	return dto
}
