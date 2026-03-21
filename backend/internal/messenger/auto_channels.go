package messenger

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

// ProjectCreatedPayload — структура события создания проекта из outbox.
// Соответствует project.ProjectCreatedEvent.
type ProjectCreatedPayload struct {
	ProjectID   string `json:"project_id"`
	Name        string `json:"name"`
	Key         string `json:"key"`
	CreatedByID string `json:"created_by_id"`
}

// AutoChannelService автоматически создаёт каналы при создании проектов.
// Подписывается на NATS subject "project.created" и создаёт канал #project-{key},
// добавляя всех участников проекта.
type AutoChannelService struct {
	messengerSvc *MessengerService
	entRepo      *EntRepository
	sub          *nats.Subscription
}

// NewAutoChannelService создаёт новый сервис авто-каналов.
func NewAutoChannelService(messengerSvc *MessengerService, entRepo *EntRepository) *AutoChannelService {
	return &AutoChannelService{
		messengerSvc: messengerSvc,
		entRepo:      entRepo,
	}
}

// Start подписывается на NATS subject "project.created" для автоматического
// создания каналов проектов. При получении события создаёт публичный канал
// #project-{key} и добавляет всех участников проекта.
func (a *AutoChannelService) Start(nc *natsclient.Client) error {
	if nc == nil {
		slog.Warn("auto-channels: NATS клиент не задан, подписка пропущена")
		return nil
	}

	sub, err := nc.Subscribe("project.created", func(msg *nats.Msg) {
		a.handleProjectCreated(msg.Data)
	})
	if err != nil {
		return fmt.Errorf("auto-channels: ошибка подписки на project.created: %w", err)
	}
	a.sub = sub
	slog.Info("auto-channels: подписка на project.created активирована")
	return nil
}

// Stop отписывается от NATS subject.
func (a *AutoChannelService) Stop() {
	if a.sub != nil {
		if err := a.sub.Unsubscribe(); err != nil {
			slog.Error("auto-channels: ошибка отписки", "error", err)
		}
		a.sub = nil
		slog.Info("auto-channels: подписка остановлена")
	}
}

// handleProjectCreated обрабатывает событие создания проекта.
// Создаёт публичный канал #project-{key} и добавляет всех участников.
func (a *AutoChannelService) handleProjectCreated(data []byte) {
	var payload ProjectCreatedPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		slog.Error("auto-channels: ошибка десериализации события", "error", err)
		return
	}

	projectID, err := uuid.Parse(payload.ProjectID)
	if err != nil {
		slog.Error("auto-channels: некорректный project_id", "project_id", payload.ProjectID, "error", err)
		return
	}

	createdBy, err := uuid.Parse(payload.CreatedByID)
	if err != nil {
		slog.Error("auto-channels: некорректный created_by_id", "created_by_id", payload.CreatedByID, "error", err)
		return
	}

	ctx := context.Background()

	// Определяем имя канала: #project-{key} (в нижнем регистре)
	channelName := "#project-" + strings.ToLower(payload.Key)

	// Создаём публичный канал, привязанный к проекту
	ch, err := a.messengerSvc.CreateChannel(ctx, createdBy, CreateChannelRequest{
		ChannelType: "public",
		Name:        channelName,
		ProjectID:   payload.ProjectID,
	})
	if err != nil {
		slog.Error("auto-channels: ошибка создания канала",
			"project_id", payload.ProjectID,
			"channel_name", channelName,
			"error", err,
		)
		return
	}

	slog.Info("auto-channels: канал создан",
		"channel_id", ch.ID,
		"channel_name", channelName,
		"project_id", payload.ProjectID,
	)

	// Получаем всех участников проекта
	memberIDs, err := a.entRepo.ListProjectMembers(ctx, projectID)
	if err != nil {
		slog.Error("auto-channels: ошибка получения участников проекта",
			"project_id", payload.ProjectID,
			"error", err,
		)
		return
	}

	// Фильтруем создателя (уже добавлен как owner)
	channelID, err := uuid.Parse(ch.ID)
	if err != nil {
		slog.Error("auto-channels: некорректный channel_id", "error", err)
		return
	}

	membersToAdd := make([]uuid.UUID, 0, len(memberIDs))
	for _, mid := range memberIDs {
		if mid != createdBy {
			membersToAdd = append(membersToAdd, mid)
		}
	}

	if len(membersToAdd) > 0 {
		if err := a.messengerSvc.AddMembersInternal(ctx, channelID, membersToAdd); err != nil {
			slog.Error("auto-channels: ошибка добавления участников в канал",
				"channel_id", ch.ID,
				"error", err,
			)
			return
		}
		slog.Info("auto-channels: участники добавлены",
			"channel_id", ch.ID,
			"count", len(membersToAdd),
		)
	}
}
