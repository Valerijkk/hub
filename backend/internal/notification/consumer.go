package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/valerii/hub/backend/ent/projectmember"
	"github.com/valerii/hub/backend/internal/auth"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
	"github.com/valerii/hub/backend/internal/tracker"
	"github.com/valerii/hub/backend/internal/tms"
)

// Consumer — JetStream durable consumer для обработки событий всех модулей
// и создания уведомлений.
type Consumer struct {
	nc      *natsclient.Client
	repo    *Repository
	service *NotificationService

	// debounce — in-memory карта для 2-секундного окна дедупликации.
	debounce   map[string]time.Time
	debounceMu sync.Mutex

	cancel context.CancelFunc
}

// NewConsumer создаёт новый consumer уведомлений.
func NewConsumer(nc *natsclient.Client, repo *Repository, service *NotificationService) *Consumer {
	return &Consumer{
		nc:       nc,
		repo:     repo,
		service:  service,
		debounce: make(map[string]time.Time),
	}
}

// Start запускает JetStream consumer для обработки событий.
func (c *Consumer) Start(ctx context.Context) error {
	if c.nc == nil {
		slog.Warn("NATS client nil, consumer уведомлений не запущен")
		return nil
	}

	js := c.nc.JetStream

	// Получаем общий стрим EVENTS (создаётся при старте сервера).
	stream, err := js.Stream(ctx, natsclient.StreamName)
	if err != nil {
		slog.Error("ошибка получения JetStream stream", "error", err)
		return err
	}

	// Создаём durable consumer на общем стриме.
	cons, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:    "notification-dispatcher",
		AckPolicy:  jetstream.AckExplicitPolicy,
		MaxDeliver: 5,
		AckWait:    30 * time.Second,
	})
	if err != nil {
		slog.Error("ошибка создания JetStream consumer", "error", err)
		return err
	}

	consCtx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	// Запускаем обработку сообщений.
	go func() {
		iter, err := cons.Messages(jetstream.PullMaxMessages(10))
		if err != nil {
			slog.Error("ошибка получения итератора сообщений", "error", err)
			return
		}
		defer iter.Stop()

		for {
			select {
			case <-consCtx.Done():
				return
			default:
			}

			msg, err := iter.Next()
			if err != nil {
				if consCtx.Err() != nil {
					return
				}
				slog.Error("ошибка получения сообщения", "error", err)
				continue
			}

			c.processEvent(consCtx, msg)
		}
	}()

	// Горутина очистки дебаунс-карты каждые 30 секунд.
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-consCtx.Done():
				return
			case <-ticker.C:
				c.cleanDebounce()
			}
		}
	}()

	slog.Info("consumer уведомлений запущен")
	return nil
}

// Stop останавливает consumer.
func (c *Consumer) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
}

// processEvent обрабатывает одно событие из JetStream.
// ACK отправляется только при успешной обработке, иначе NATS повторит доставку.
func (c *Consumer) processEvent(ctx context.Context, msg jetstream.Msg) {
	subject := msg.Subject()
	notifType := SubjectToType(subject)
	if notifType == "" {
		// Неизвестный тип — ACK чтобы не копилось.
		_ = msg.Ack()
		return
	}

	switch notifType {
	case TypeAssigned:
		c.handleAssigned(ctx, msg.Data(), subject)
	case TypeStatusChanged:
		c.handleStatusChanged(ctx, msg.Data(), subject)
	case TypeCommentAdded:
		c.handleCommentAdded(ctx, msg.Data(), subject)
	case TypeBlockerAdded:
		c.handleBlockerAdded(ctx, msg.Data(), subject)
	case TypeMessageReceived:
		c.handleMessageReceived(ctx, msg.Data(), subject)
	case TypeCallStarted:
		c.handleCallStarted(ctx, msg.Data(), subject)
	case TypeRunCompleted:
		c.handleRunCompleted(ctx, msg.Data(), subject)
	case TypeNewLogin:
		c.handleNewLogin(ctx, msg.Data())
	}

	// ACK только после успешной обработки. При ошибке NATS повторит доставку (MaxDeliver: 5).
	if err := msg.Ack(); err != nil {
		slog.Error("ошибка ACK сообщения", "error", err)
	}
}

// handleAssigned обрабатывает событие назначения задачи.
func (c *Consumer) handleAssigned(ctx context.Context, data []byte, _ string) {
	var evt tracker.IssueAssignedEvent
	if err := json.Unmarshal(data, &evt); err != nil {
		slog.Error("ошибка десериализации IssueAssignedEvent", "error", err)
		return
	}

	assigneeID, err := uuid.Parse(evt.AssigneeID)
	if err != nil {
		slog.Error("некорректный UUID в событии", "field", "assignee_id", "value", evt.AssigneeID, "error", err)
		return
	}
	actorID, err := uuid.Parse(evt.AssignedByID)
	if err != nil {
		slog.Error("некорректный UUID в событии", "field", "assigned_by_id", "value", evt.AssignedByID, "error", err)
		return
	}
	projectID, err := uuid.Parse(evt.ProjectID)
	if err != nil {
		slog.Error("некорректный UUID в событии", "field", "project_id", "value", evt.ProjectID, "error", err)
		return
	}

	if assigneeID == actorID {
		return // Не уведомляем о самоназначении.
	}

	if !c.shouldDispatch(assigneeID.String(), TypeAssigned, evt.IssueID) {
		return
	}

	if err := c.service.Dispatch(ctx, DispatchEvent{
		UserID:     assigneeID,
		Type:       TypeAssigned,
		Title:      "Вам назначена задача " + evt.IssueKey,
		EntityType: "issue",
		EntityID:   evt.IssueID,
		ProjectID:  &projectID,
		ActorID:    &actorID,
	}); err != nil {
		slog.Error("не удалось отправить уведомление", "event", "assigned", "error", err)
	}
}

// handleStatusChanged обрабатывает событие смены статуса задачи.
func (c *Consumer) handleStatusChanged(ctx context.Context, data []byte, _ string) {
	var evt tracker.IssueMovedEvent
	if err := json.Unmarshal(data, &evt); err != nil {
		slog.Error("ошибка десериализации IssueMovedEvent", "error", err)
		return
	}

	projectID, err := uuid.Parse(evt.ProjectID)
	if err != nil {
		slog.Error("некорректный UUID в событии", "field", "project_id", "value", evt.ProjectID, "error", err)
		return
	}
	actorID, err := uuid.Parse(evt.MovedByID)
	if err != nil {
		slog.Error("некорректный UUID в событии", "field", "moved_by_id", "value", evt.MovedByID, "error", err)
		return
	}

	// Уведомляем назначенного пользователя (если есть) — загрузим из БД.
	issueID, err := uuid.Parse(evt.IssueID)
	if err != nil {
		slog.Error("некорректный UUID в событии", "field", "issue_id", "value", evt.IssueID, "error", err)
		return
	}

	assignee, err := c.repo.GetIssueAssignee(ctx, issueID)
	if err != nil || assignee == nil {
		return
	}

	if assignee.ID == actorID {
		return
	}

	if !c.shouldDispatch(assignee.ID.String(), TypeStatusChanged, evt.IssueID) {
		return
	}

	if err := c.service.Dispatch(ctx, DispatchEvent{
		UserID:     assignee.ID,
		Type:       TypeStatusChanged,
		Title:      "Статус задачи " + evt.IssueKey + " изменён",
		EntityType: "issue",
		EntityID:   evt.IssueID,
		ProjectID:  &projectID,
		ActorID:    &actorID,
	}); err != nil {
		slog.Error("не удалось отправить уведомление", "event", "status_changed", "error", err)
	}
}

// handleCommentAdded обрабатывает событие добавления комментария.
func (c *Consumer) handleCommentAdded(ctx context.Context, data []byte, _ string) {
	var evt tracker.CommentCreatedEvent
	if err := json.Unmarshal(data, &evt); err != nil {
		slog.Error("ошибка десериализации CommentCreatedEvent", "error", err)
		return
	}

	projectID, err := uuid.Parse(evt.ProjectID)
	if err != nil {
		slog.Error("некорректный UUID в событии", "field", "project_id", "value", evt.ProjectID, "error", err)
		return
	}
	actorID, err := uuid.Parse(evt.AuthorID)
	if err != nil {
		slog.Error("некорректный UUID в событии", "field", "author_id", "value", evt.AuthorID, "error", err)
		return
	}

	// Уведомляем упомянутых пользователей.
	for _, mentionStr := range evt.Mentions {
		mentionID, err := uuid.Parse(mentionStr)
		if err != nil || mentionID == actorID {
			continue
		}

		if !c.shouldDispatch(mentionID.String(), TypeMentioned, evt.IssueID) {
			continue
		}

		if err := c.service.Dispatch(ctx, DispatchEvent{
			UserID:     mentionID,
			Type:       TypeMentioned,
			Title:      "Вас упомянули в задаче " + evt.IssueKey,
			EntityType: "issue",
			EntityID:   evt.IssueID,
			ProjectID:  &projectID,
			ActorID:    &actorID,
		}); err != nil {
			slog.Error("не удалось отправить уведомление", "event", "mentioned", "error", err)
		}
	}

	// Уведомляем назначенного пользователя о новом комментарии.
	issueID, err := uuid.Parse(evt.IssueID)
	if err != nil {
		slog.Error("некорректный UUID в событии", "field", "issue_id", "value", evt.IssueID, "error", err)
		return
	}
	assignee, err := c.repo.GetIssueAssignee(ctx, issueID)
	if err != nil || assignee == nil {
		return
	}
	if assignee.ID == actorID {
		return
	}

	if !c.shouldDispatch(assignee.ID.String(), TypeCommentAdded, evt.IssueID) {
		return
	}

	if err := c.service.Dispatch(ctx, DispatchEvent{
		UserID:     assignee.ID,
		Type:       TypeCommentAdded,
		Title:      "Новый комментарий к задаче " + evt.IssueKey,
		EntityType: "issue",
		EntityID:   evt.IssueID,
		ProjectID:  &projectID,
		ActorID:    &actorID,
	}); err != nil {
		slog.Error("не удалось отправить уведомление", "event", "comment_added", "error", err)
	}
}

// handleBlockerAdded обрабатывает событие добавления блокирующей связи.
func (c *Consumer) handleBlockerAdded(ctx context.Context, data []byte, _ string) {
	// Событие создания связи — структура упрощённая.
	var evt struct {
		SourceID  string `json:"source_id"`
		TargetID  string `json:"target_id"`
		LinkType  string `json:"link_type"`
		ProjectID string `json:"project_id"`
		CreatedBy string `json:"created_by"`
	}
	if err := json.Unmarshal(data, &evt); err != nil {
		return
	}
	if evt.LinkType != "blocks" {
		return
	}

	// Уведомляем назначенного на заблокированную задачу.
	targetID, err := uuid.Parse(evt.TargetID)
	if err != nil {
		slog.Error("некорректный UUID в событии", "field", "target_id", "value", evt.TargetID, "error", err)
		return
	}
	assignee, err := c.repo.GetIssueAssignee(ctx, targetID)
	if err != nil || assignee == nil {
		return
	}

	actorID, err := uuid.Parse(evt.CreatedBy)
	if err != nil {
		slog.Error("некорректный UUID в событии", "field", "created_by", "value", evt.CreatedBy, "error", err)
		return
	}
	projectID, err := uuid.Parse(evt.ProjectID)
	if err != nil {
		slog.Error("некорректный UUID в событии", "field", "project_id", "value", evt.ProjectID, "error", err)
		return
	}

	if assignee.ID == actorID {
		return
	}

	if err := c.service.Dispatch(ctx, DispatchEvent{
		UserID:     assignee.ID,
		Type:       TypeBlockerAdded,
		Title:      "Задача заблокирована",
		EntityType: "issue",
		EntityID:   evt.TargetID,
		ProjectID:  &projectID,
		ActorID:    &actorID,
		IsUrgent:   true,
	}); err != nil {
		slog.Error("не удалось отправить уведомление", "event", "blocker_added", "error", err)
	}
}

// handleMessageReceived обрабатывает событие нового сообщения.
// Для DM — уведомляем второго участника.
// Для любого типа канала — уведомляем @упомянутых пользователей.
func (c *Consumer) handleMessageReceived(ctx context.Context, data []byte, subject string) {
	// Извлекаем channelID из subject: messenger.channel.{channelID}.message_sent
	parts := strings.Split(subject, ".")
	if len(parts) < 4 {
		return
	}

	var evt struct {
		MessageID        string   `json:"message_id"`
		ChannelID        string   `json:"channel_id"`
		SenderID         string   `json:"sender_id"`
		ContentText      string   `json:"content_text"`
		MentionedUserIDs []string `json:"mentioned_user_ids"`
	}
	if err := json.Unmarshal(data, &evt); err != nil {
		return
	}

	senderID, err := uuid.Parse(evt.SenderID)
	if err != nil {
		slog.Error("некорректный UUID в событии", "field", "sender_id", "value", evt.SenderID, "error", err)
		return
	}
	channelID, err := uuid.Parse(parts[2])
	if err != nil {
		slog.Error("некорректный UUID в событии", "field", "channel_id", "value", parts[2], "error", err)
		return
	}

	ch, err := c.repo.GetChannel(ctx, channelID)
	if err != nil {
		return
	}

	truncated := evt.ContentText
	if len(truncated) > 100 {
		truncated = truncated[:100] + "..."
	}

	// Уведомляем @упомянутых пользователей в любом типе канала.
	for _, mentionStr := range evt.MentionedUserIDs {
		mentionID, err := uuid.Parse(mentionStr)
		if err != nil || mentionID == senderID {
			continue
		}

		if !c.shouldDispatch(mentionID.String(), TypeMentioned, evt.MessageID) {
			continue
		}

		if err := c.service.Dispatch(ctx, DispatchEvent{
			UserID:     mentionID,
			Type:       TypeMentioned,
			Title:      "Вас упомянули в сообщении",
			Body:       truncated,
			EntityType: "channel",
			EntityID:   parts[2],
			ActorID:    &senderID,
		}); err != nil {
			slog.Error("не удалось отправить уведомление", "event", "mentioned_message", "error", err)
		}
	}

	// Для DM — дополнительно уведомляем второго участника.
	if string(ch.ChannelType) != "dm" {
		return
	}

	members, err := c.repo.GetChannelMembers(ctx, channelID)
	if err != nil {
		return
	}

	for _, m := range members {
		recipientID := m.UserID
		if recipientID == senderID {
			continue
		}

		if !c.shouldDispatch(recipientID.String(), TypeMessageReceived, evt.MessageID) {
			continue
		}

		if err := c.service.Dispatch(ctx, DispatchEvent{
			UserID:     recipientID,
			Type:       TypeMessageReceived,
			Title:      "Новое сообщение",
			Body:       truncated,
			EntityType: "channel",
			EntityID:   parts[2],
			ActorID:    &senderID,
		}); err != nil {
			slog.Error("не удалось отправить уведомление", "event", "message_received", "error", err)
		}
	}
}

// handleCallStarted обрабатывает событие начала звонка.
func (c *Consumer) handleCallStarted(ctx context.Context, data []byte, subject string) {
	parts := strings.Split(subject, ".")
	if len(parts) < 4 {
		return
	}

	var evt struct {
		CallID    string `json:"call_id"`
		ChannelID string `json:"channel_id"`
		StartedBy string `json:"started_by"`
	}
	if err := json.Unmarshal(data, &evt); err != nil {
		return
	}

	starterID, err := uuid.Parse(evt.StartedBy)
	if err != nil {
		slog.Error("некорректный UUID в событии", "field", "started_by", "value", evt.StartedBy, "error", err)
		return
	}
	channelID, err := uuid.Parse(parts[2])
	if err != nil {
		slog.Error("некорректный UUID в событии", "field", "channel_id", "value", parts[2], "error", err)
		return
	}

	channelMembers, err := c.repo.GetChannelMembers(ctx, channelID)
	if err != nil {
		return
	}

	for _, m := range channelMembers {
		recipientID := m.UserID
		if recipientID == starterID {
			continue
		}

		if !c.shouldDispatch(recipientID.String(), TypeCallStarted, evt.CallID) {
			continue
		}

		if err := c.service.Dispatch(ctx, DispatchEvent{
			UserID:     recipientID,
			Type:       TypeCallStarted,
			Title:      "Звонок начался",
			EntityType: "call",
			EntityID:   evt.CallID,
			ActorID:    &starterID,
		}); err != nil {
			slog.Error("не удалось отправить уведомление", "event", "call_started", "error", err)
		}
	}
}

// handleRunCompleted обрабатывает событие завершения тестового прогона.
// При failed_count > 0 отправляет run_failed QA/tester + автору прогона.
func (c *Consumer) handleRunCompleted(ctx context.Context, data []byte, _ string) {
	var evt tms.TestRunCompletedEvent
	if err := json.Unmarshal(data, &evt); err != nil {
		return
	}

	completedByID, err := uuid.Parse(evt.CompletedBy)
	if err != nil {
		slog.Error("некорректный UUID в событии", "field", "completed_by", "value", evt.CompletedBy, "error", err)
		return
	}
	projectID, err := uuid.Parse(evt.ProjectID)
	if err != nil {
		slog.Error("некорректный UUID в событии", "field", "project_id", "value", evt.ProjectID, "error", err)
		return
	}

	members, err := c.repo.GetProjectMembersWithUser(ctx, projectID)
	if err != nil {
		return
	}

	hasFailures := evt.Failed > 0
	notifType := TypeRunCompleted
	title := "Прогон завершён: " + evt.Title
	isUrgent := false
	if hasFailures {
		notifType = TypeRunFailed
		title = "Прогон с ошибками: " + evt.Title
		isUrgent = true
	}

	body := fmt.Sprintf("Passed: %d, Failed: %d, Skipped: %d", evt.Passed, evt.Failed, evt.Skipped)

	for _, pm := range members {
		if pm.Edges.User == nil {
			continue
		}
		recipientID := pm.Edges.User.ID

		if hasFailures {
			if !isRunFailedRecipient(pm.ProjectRole, recipientID, completedByID) {
				continue
			}
		} else if recipientID == completedByID {
			continue
		}

		if !c.shouldDispatch(recipientID.String(), notifType, evt.RunID) {
			continue
		}

		if err := c.service.Dispatch(ctx, DispatchEvent{
			UserID:     recipientID,
			Type:       notifType,
			Title:      title,
			Body:       body,
			EntityType: "testrun",
			EntityID:   evt.RunID,
			ProjectID:  &projectID,
			ActorID:    &completedByID,
			IsUrgent:   isUrgent,
		}); err != nil {
			slog.Error("не удалось отправить уведомление", "event", notifType, "error", err)
		}
	}
}

// isRunFailedRecipient — QA/tester роль или автор прогона (NOTIF-04).
func isRunFailedRecipient(role projectmember.ProjectRole, userID, completedByID uuid.UUID) bool {
	if userID == completedByID {
		return true
	}
	return role == projectmember.ProjectRoleTester
}

// handleNewLogin обрабатывает событие нового входа в аккаунт.
// Уведомляет пользователя о логине с информацией об устройстве и IP-адресе.
func (c *Consumer) handleNewLogin(ctx context.Context, data []byte) {
	var evt auth.SessionCreatedEvent
	if err := json.Unmarshal(data, &evt); err != nil {
		slog.Error("ошибка десериализации SessionCreatedEvent", "error", err)
		return
	}

	userID, err := uuid.Parse(evt.UserID)
	if err != nil {
		slog.Error("некорректный UUID в событии", "field", "user_id", "value", evt.UserID, "error", err)
		return
	}

	if !c.shouldDispatch(userID.String(), TypeNewLogin, evt.SessionID) {
		return
	}

	device := evt.Device
	if device == "" {
		device = "неизвестное устройство"
	}

	body := fmt.Sprintf("Устройство: %s, IP: %s", device, evt.IPAddress)

	if err := c.service.Dispatch(ctx, DispatchEvent{
		UserID:     userID,
		Type:       TypeNewLogin,
		Title:      "Новый вход в аккаунт",
		Body:       body,
		EntityType: "session",
		EntityID:   evt.SessionID,
	}); err != nil {
		slog.Error("не удалось отправить уведомление", "event", "new_login", "error", err)
	}
}

// shouldDispatch проверяет 2-секундное окно дебаунса.
func (c *Consumer) shouldDispatch(userID, notifType, entityID string) bool {
	key := userID + ":" + notifType + ":" + entityID
	c.debounceMu.Lock()
	defer c.debounceMu.Unlock()

	if last, ok := c.debounce[key]; ok {
		if time.Since(last) < 2*time.Second {
			return false
		}
	}
	c.debounce[key] = time.Now()
	return true
}

// cleanDebounce очищает устаревшие записи в дебаунс-карте.
func (c *Consumer) cleanDebounce() {
	c.debounceMu.Lock()
	defer c.debounceMu.Unlock()

	cutoff := time.Now().Add(-5 * time.Second)
	for key, t := range c.debounce {
		if t.Before(cutoff) {
			delete(c.debounce, key)
		}
	}
}
