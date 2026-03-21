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

	// Создаём stream для уведомлений.
	_, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name: "NOTIFICATIONS",
		Subjects: []string{
			"project.>",
			"messenger.channel.>",
			"tms.>",
			"calls.channel.>",
		},
		Retention: jetstream.WorkQueuePolicy,
		MaxAge:    7 * 24 * time.Hour,
	})
	if err != nil {
		slog.Error("ошибка создания JetStream stream", "error", err)
		return err
	}

	// Создаём durable consumer.
	cons, err := js.CreateOrUpdateConsumer(ctx, "NOTIFICATIONS", jetstream.ConsumerConfig{
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
func (c *Consumer) processEvent(ctx context.Context, msg jetstream.Msg) {
	defer func() {
		if err := msg.Ack(); err != nil {
			slog.Error("ошибка ACK сообщения", "error", err)
		}
	}()

	subject := msg.Subject()
	notifType := SubjectToType(subject)
	if notifType == "" {
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
		return
	}
	actorID, _ := uuid.Parse(evt.AssignedByID)
	projectID, _ := uuid.Parse(evt.ProjectID)

	if assigneeID == actorID {
		return // Не уведомляем о самоназначении.
	}

	if !c.shouldDispatch(assigneeID.String(), TypeAssigned, evt.IssueID) {
		return
	}

	_ = c.service.Dispatch(ctx, assigneeID, TypeAssigned,
		"Вам назначена задача "+evt.IssueKey,
		"",
		"issue", evt.IssueID,
		&projectID, &actorID, false)
}

// handleStatusChanged обрабатывает событие смены статуса задачи.
func (c *Consumer) handleStatusChanged(ctx context.Context, data []byte, _ string) {
	var evt tracker.IssueMovedEvent
	if err := json.Unmarshal(data, &evt); err != nil {
		slog.Error("ошибка десериализации IssueMovedEvent", "error", err)
		return
	}

	projectID, _ := uuid.Parse(evt.ProjectID)
	actorID, _ := uuid.Parse(evt.MovedByID)

	// Уведомляем назначенного пользователя (если есть) — загрузим из БД.
	issueID, err := uuid.Parse(evt.IssueID)
	if err != nil {
		return
	}

	issue, err := c.repo.db.Issue.Get(ctx, issueID)
	if err != nil {
		return
	}

	assignee, err := issue.QueryAssignee().First(ctx)
	if err != nil || assignee == nil {
		return
	}

	if assignee.ID == actorID {
		return
	}

	if !c.shouldDispatch(assignee.ID.String(), TypeStatusChanged, evt.IssueID) {
		return
	}

	_ = c.service.Dispatch(ctx, assignee.ID, TypeStatusChanged,
		"Статус задачи "+evt.IssueKey+" изменён",
		"",
		"issue", evt.IssueID,
		&projectID, &actorID, false)
}

// handleCommentAdded обрабатывает событие добавления комментария.
func (c *Consumer) handleCommentAdded(ctx context.Context, data []byte, _ string) {
	var evt tracker.CommentCreatedEvent
	if err := json.Unmarshal(data, &evt); err != nil {
		slog.Error("ошибка десериализации CommentCreatedEvent", "error", err)
		return
	}

	projectID, _ := uuid.Parse(evt.ProjectID)
	actorID, _ := uuid.Parse(evt.AuthorID)

	// Уведомляем упомянутых пользователей.
	for _, mentionStr := range evt.Mentions {
		mentionID, err := uuid.Parse(mentionStr)
		if err != nil || mentionID == actorID {
			continue
		}

		if !c.shouldDispatch(mentionID.String(), TypeMentioned, evt.IssueID) {
			continue
		}

		_ = c.service.Dispatch(ctx, mentionID, TypeMentioned,
			"Вас упомянули в задаче "+evt.IssueKey,
			"",
			"issue", evt.IssueID,
			&projectID, &actorID, false)
	}

	// Уведомляем назначенного пользователя о новом комментарии.
	issueID, err := uuid.Parse(evt.IssueID)
	if err != nil {
		return
	}
	issue, err := c.repo.db.Issue.Get(ctx, issueID)
	if err != nil {
		return
	}
	assignee, err := issue.QueryAssignee().First(ctx)
	if err != nil || assignee == nil {
		return
	}
	if assignee.ID == actorID {
		return
	}

	if !c.shouldDispatch(assignee.ID.String(), TypeCommentAdded, evt.IssueID) {
		return
	}

	_ = c.service.Dispatch(ctx, assignee.ID, TypeCommentAdded,
		"Новый комментарий к задаче "+evt.IssueKey,
		"",
		"issue", evt.IssueID,
		&projectID, &actorID, false)
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
		return
	}
	issue, err := c.repo.db.Issue.Get(ctx, targetID)
	if err != nil {
		return
	}
	assignee, err := issue.QueryAssignee().First(ctx)
	if err != nil || assignee == nil {
		return
	}

	actorID, _ := uuid.Parse(evt.CreatedBy)
	projectID, _ := uuid.Parse(evt.ProjectID)

	if assignee.ID == actorID {
		return
	}

	_ = c.service.Dispatch(ctx, assignee.ID, TypeBlockerAdded,
		"Задача заблокирована",
		"",
		"issue", evt.TargetID,
		&projectID, &actorID, true)
}

// handleMessageReceived обрабатывает событие нового сообщения.
func (c *Consumer) handleMessageReceived(ctx context.Context, data []byte, subject string) {
	// Извлекаем channelID из subject: messenger.channel.{channelID}.message_sent
	parts := strings.Split(subject, ".")
	if len(parts) < 4 {
		return
	}

	var evt struct {
		MessageID   string `json:"message_id"`
		ChannelID   string `json:"channel_id"`
		SenderID    string `json:"sender_id"`
		ContentText string `json:"content_text"`
	}
	if err := json.Unmarshal(data, &evt); err != nil {
		return
	}

	senderID, _ := uuid.Parse(evt.SenderID)
	channelID, err := uuid.Parse(parts[2])
	if err != nil {
		return
	}

	// Для DM — уведомляем второго участника.
	channel, err := c.repo.db.Channel.Get(ctx, channelID)
	if err != nil {
		return
	}

	if string(channel.ChannelType) != "dm" {
		return // Уведомления только для DM.
	}

	members, err := channel.QueryMembers().All(ctx)
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

		truncated := evt.ContentText
		if len(truncated) > 100 {
			truncated = truncated[:100] + "..."
		}

		_ = c.service.Dispatch(ctx, recipientID, TypeMessageReceived,
			"Новое сообщение",
			truncated,
			"channel", parts[2],
			nil, &senderID, false)
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

	starterID, _ := uuid.Parse(evt.StartedBy)
	channelID, err := uuid.Parse(parts[2])
	if err != nil {
		return
	}

	members, err := c.repo.db.Channel.Get(ctx, channelID)
	if err != nil {
		return
	}
	channelMembers, err := members.QueryMembers().All(ctx)
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

		_ = c.service.Dispatch(ctx, recipientID, TypeCallStarted,
			"Звонок начался",
			"",
			"call", evt.CallID,
			nil, &starterID, false)
	}
}

// handleRunCompleted обрабатывает событие завершения тестового прогона.
func (c *Consumer) handleRunCompleted(ctx context.Context, data []byte, _ string) {
	var evt tms.TestRunCompletedEvent
	if err := json.Unmarshal(data, &evt); err != nil {
		return
	}

	completedByID, _ := uuid.Parse(evt.CompletedBy)
	projectID, _ := uuid.Parse(evt.ProjectID)

	// Уведомляем участников проекта через edge-запрос.
	project, err := c.repo.db.Project.Get(ctx, projectID)
	if err != nil {
		return
	}
	members, err := project.QueryMembers().WithUser().All(ctx)
	if err != nil {
		return
	}

	body := fmt.Sprintf("Passed: %d, Failed: %d, Skipped: %d", evt.Passed, evt.Failed, evt.Skipped)

	for _, pm := range members {
		if pm.Edges.User == nil {
			continue
		}
		recipientID := pm.Edges.User.ID
		if recipientID == completedByID {
			continue
		}

		if !c.shouldDispatch(recipientID.String(), TypeRunCompleted, evt.RunID) {
			continue
		}

		_ = c.service.Dispatch(ctx, recipientID, TypeRunCompleted,
			"Прогон завершён: "+evt.Title,
			body,
			"testrun", evt.RunID,
			&projectID, &completedByID, false)
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
