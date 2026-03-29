package messenger

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/google/uuid"
	apperr "github.com/valerii/hub/backend/internal/platform/errors"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
	"github.com/valerii/hub/backend/internal/rbac"
)

// splitComma разделяет строку по запятым, убирая пробелы.
func splitComma(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// queryInt извлекает целочисленный параметр из строки запроса с дефолтным значением.
func queryInt(c fiber.Ctx, key string, defaultVal int) int {
	s := c.Query(key)
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return v
}

// Handler предоставляет HTTP-обработчики для модуля мессенджера.
type Handler struct {
	service *MessengerService
}

// NewHandler создаёт новый обработчик мессенджера.
func NewHandler(service *MessengerService) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes регистрирует все маршруты мессенджера под /api/messenger.
// authMw — middleware аутентификации, rbacSvc используется для проверки прав доступа.
func (h *Handler) RegisterRoutes(api fiber.Router, authMw fiber.Handler, rbacSvc *rbac.RBACService) {
	msg := api.Group("/messenger", authMw)

	// Каналы
	channels := msg.Group("/channels")
	channels.Post("/", h.CreateChannel)
	channels.Get("/", h.ListMyChannels)
	channels.Get("/:channelId", h.GetChannel)
	channels.Put("/:channelId", h.UpdateChannel)
	channels.Delete("/:channelId", h.DeleteChannel)
	channels.Post("/:channelId/join", h.JoinChannel)
	channels.Post("/:channelId/leave", h.LeaveChannel)
	channels.Post("/:channelId/members", h.AddMembers)
	channels.Get("/:channelId/members", h.ListMembers)
	channels.Post("/:channelId/star", h.ToggleStar)
	channels.Post("/:channelId/mute", h.MuteChannel)
	channels.Post("/:channelId/unmute", h.UnmuteChannel)

	// Индикатор набора текста
	channels.Post("/:channelId/typing", h.SendTyping)

	// Сообщения
	messages := channels.Group("/:channelId/messages")
	messages.Post("/", h.SendMessage)
	messages.Get("/", h.ListMessages)
	messages.Put("/:messageId", h.EditMessage)
	messages.Delete("/:messageId", h.DeleteMessage)
	messages.Post("/:messageId/reactions", h.ToggleReaction)
	messages.Post("/:messageId/pin", h.TogglePin)
	messages.Post("/:messageId/forward", h.ForwardMessage)
	messages.Get("/:messageId/replies", h.ListThreadReplies)

	// Присутствие
	msg.Get("/presence", h.GetPresence)
	msg.Put("/presence/status", h.SetCustomStatus)
	msg.Post("/presence/heartbeat", h.Heartbeat)

	// Непрочитанные
	msg.Get("/unread", h.GetUnreadCounts)
	channels.Post("/:channelId/read", h.MarkChannelRead)

	// Закреплённые сообщения
	channels.Get("/:channelId/pins", h.ListPins)

	// Поиск
	msg.Get("/search", h.SearchMessages)

	// Закладки
	msg.Get("/bookmarks", h.ListBookmarks)
	msg.Post("/bookmarks", h.AddBookmark)
	msg.Delete("/bookmarks/:bookmarkId", h.RemoveBookmark)

	// DM
	msg.Post("/dm", h.CreateOrGetDM)

	// Опросы
	channels.Post("/:channelId/polls", h.CreatePoll)
	polls := msg.Group("/polls")
	polls.Post("/:pollId/vote", h.VotePoll)
	polls.Post("/:pollId/close", h.ClosePoll)
	polls.Get("/:pollId", h.GetPollResults)

	// Пригласительные ссылки
	channels.Post("/:channelId/invites", h.CreateInviteLink)
	channels.Get("/:channelId/invites", h.ListChannelInvites)
	channels.Delete("/:channelId/invites/:inviteId", h.RevokeInvite)
	invites := msg.Group("/invites")
	invites.Get("/:token", h.ValidateInvite)
	invites.Post("/:token/accept", h.AcceptInvite)
}

// ---------------------------------------------------------------------------
// Invite Link Handlers
// ---------------------------------------------------------------------------

// CreateInviteLink godoc
// @Summary Создать пригласительную ссылку канала
// @Description Генерирует уникальную пригласительную ссылку для вступления в указанный канал.
// @Tags Мессенджер
// @Accept json
// @Produce json
// @Param channelId path string true "ID канала"
// @Param body body CreateInviteLinkRequest true "Параметры приглашения"
// @Success 201 {object} InviteLinkResponse "Ссылка создана"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/messenger/channels/{channelId}/invites [post]
func (h *Handler) CreateInviteLink(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	var req CreateInviteLinkRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	resp, err := h.service.CreateInviteLink(c.Context(), channelID, userID, req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListChannelInvites godoc
// @Summary Список пригласительных ссылок канала
// @Description Возвращает все активные пригласительные ссылки указанного канала.
// @Tags Мессенджер
// @Produce json
// @Param channelId path string true "ID канала"
// @Success 200 {array} InviteLinkResponse "Список приглашений"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/messenger/channels/{channelId}/invites [get]
func (h *Handler) ListChannelInvites(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	invites, err := h.service.ListChannelInvites(c.Context(), channelID, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(invites)
}

// ValidateInvite godoc
// @Summary Проверить пригласительную ссылку
// @Description Проверяет валидность пригласительного токена и возвращает информацию о канале.
// @Tags Мессенджер
// @Produce json
// @Param token path string true "Токен приглашения"
// @Success 200 {object} ValidateInviteResponse "Информация о приглашении"
// @Failure 400 {object} fiber.Error "Токен не указан"
// @Failure 404 {object} fiber.Error "Приглашение не найдено"
// @Security BearerAuth
// @Router /api/messenger/invites/{token} [get]
func (h *Handler) ValidateInvite(c fiber.Ctx) error {
	token := c.Params("token")
	if token == "" {
		return fiber.NewError(fiber.StatusBadRequest, "token required")
	}
	resp, err := h.service.ValidateInvite(c.Context(), token)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	return c.JSON(resp)
}

// AcceptInvite godoc
// @Summary Принять приглашение в канал
// @Description Принимает пригласительную ссылку и добавляет пользователя в канал.
// @Tags Мессенджер
// @Produce json
// @Param token path string true "Токен приглашения"
// @Success 200 {object} AcceptInviteResponse "Пользователь добавлен в канал"
// @Failure 400 {object} fiber.Error "Неверный токен или запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/messenger/invites/{token}/accept [post]
func (h *Handler) AcceptInvite(c fiber.Ctx) error {
	token := c.Params("token")
	if token == "" {
		return fiber.NewError(fiber.StatusBadRequest, "token required")
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	resp, err := h.service.AcceptInvite(c.Context(), token, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(resp)
}

// RevokeInvite godoc
// @Summary Отозвать пригласительную ссылку
// @Description Деактивирует пригласительную ссылку канала. Доступно только участникам канала.
// @Tags Мессенджер
// @Param channelId path string true "ID канала"
// @Param inviteId path string true "ID приглашения"
// @Success 200 "Приглашение отозвано"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/messenger/channels/{channelId}/invites/{inviteId} [delete]
func (h *Handler) RevokeInvite(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	inviteIDStr := c.Params("inviteId")
	inviteID, err := uuid.Parse(inviteIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid invite ID")
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	if err := h.service.RevokeInvite(c.Context(), channelID, inviteID, userID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.SendStatus(fiber.StatusOK)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// parseChannelID извлекает и валидирует channelId из параметров пути.
func parseChannelID(c fiber.Ctx) (uuid.UUID, error) {
	idStr := utils.CopyString(c.Params("channelId"))
	if idStr == "" {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "channel ID обязателен")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "некорректный channel ID")
	}
	return id, nil
}

// parseMessageID извлекает messageId из параметров пути.
func parseMessageID(c fiber.Ctx) string {
	return utils.CopyString(c.Params("messageId"))
}

// ---------------------------------------------------------------------------
// Channel handlers
// ---------------------------------------------------------------------------

// CreateChannel godoc
// @Summary Создать канал
// @Description Создаёт новый канал мессенджера. Создатель автоматически становится участником.
// @Tags Мессенджер
// @Accept json
// @Produce json
// @Param body body CreateChannelRequest true "Данные канала"
// @Success 201 {object} ChannelResponse "Канал создан"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/messenger/channels [post]
func (h *Handler) CreateChannel(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req CreateChannelRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.CreateChannel(c.Context(), userID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListMyChannels godoc
// @Summary Список моих каналов
// @Description Возвращает список каналов, в которых состоит текущий пользователь.
// @Tags Мессенджер
// @Produce json
// @Success 200 {array} ChannelResponse "Список каналов"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/messenger/channels [get]
func (h *Handler) ListMyChannels(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	resp, err := h.service.ListMyChannels(c.Context(), userID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// GetChannel godoc
// @Summary Получить канал
// @Description Возвращает детальную информацию о канале по его ID.
// @Tags Мессенджер
// @Produce json
// @Param channelId path string true "ID канала"
// @Success 200 {object} ChannelResponse "Данные канала"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 404 {object} fiber.Error "Канал не найден"
// @Security BearerAuth
// @Router /api/messenger/channels/{channelId} [get]
func (h *Handler) GetChannel(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}

	resp, err := h.service.GetChannel(c.Context(), channelID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// UpdateChannel godoc
// @Summary Обновить канал
// @Description Обновляет название и описание канала. Доступно только участникам.
// @Tags Мессенджер
// @Accept json
// @Produce json
// @Param channelId path string true "ID канала"
// @Param body body CreateChannelRequest true "Обновлённые данные"
// @Success 200 {object} ChannelResponse "Канал обновлён"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/messenger/channels/{channelId} [put]
func (h *Handler) UpdateChannel(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req CreateChannelRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.UpdateChannel(c.Context(), channelID, userID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// DeleteChannel godoc
// @Summary Удалить канал
// @Description Удаляет канал мессенджера. Доступно только создателю канала.
// @Tags Мессенджер
// @Param channelId path string true "ID канала"
// @Success 204 "Канал удалён"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Security BearerAuth
// @Router /api/messenger/channels/{channelId} [delete]
func (h *Handler) DeleteChannel(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	if err := h.service.DeleteChannel(c.Context(), channelID, userID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// JoinChannel godoc
// @Summary Вступить в канал
// @Description Текущий пользователь вступает в указанный канал.
// @Tags Мессенджер
// @Param channelId path string true "ID канала"
// @Success 200 "Пользователь вступил в канал"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/messenger/channels/{channelId}/join [post]
func (h *Handler) JoinChannel(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	if err := h.service.JoinChannel(c.Context(), channelID, userID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

// LeaveChannel godoc
// @Summary Покинуть канал
// @Description Текущий пользователь покидает указанный канал.
// @Tags Мессенджер
// @Param channelId path string true "ID канала"
// @Success 200 "Пользователь покинул канал"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/messenger/channels/{channelId}/leave [post]
func (h *Handler) LeaveChannel(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	if err := h.service.LeaveChannel(c.Context(), channelID, userID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

// AddMembers godoc
// @Summary Добавить участников в канал
// @Description Добавляет одного или нескольких пользователей в канал.
// @Tags Мессенджер
// @Accept json
// @Param channelId path string true "ID канала"
// @Param body body AddMembersRequest true "Список ID пользователей"
// @Success 200 "Участники добавлены"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/messenger/channels/{channelId}/members [post]
func (h *Handler) AddMembers(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req AddMembersRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	if err := h.service.AddMembers(c.Context(), channelID, userID, req.MemberIDs); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

// ListMembers godoc
// @Summary Список участников канала
// @Description Возвращает список всех участников указанного канала.
// @Tags Мессенджер
// @Produce json
// @Param channelId path string true "ID канала"
// @Success 200 {array} MemberResponse "Список участников"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/messenger/channels/{channelId}/members [get]
func (h *Handler) ListMembers(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}

	resp, err := h.service.ListMembers(c.Context(), channelID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// ---------------------------------------------------------------------------
// Message handlers
// ---------------------------------------------------------------------------

// SendMessage godoc
// @Summary Отправить сообщение
// @Description Отправляет новое сообщение в указанный канал. Поддерживает ответы в тредах.
// @Tags Мессенджер
// @Accept json
// @Produce json
// @Param channelId path string true "ID канала"
// @Param body body SendMessageRequest true "Данные сообщения"
// @Success 201 {object} MessageResponse "Сообщение отправлено"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/messenger/channels/{channelId}/messages [post]
func (h *Handler) SendMessage(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req SendMessageRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.SendMessage(c.Context(), channelID, userID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListMessages godoc
// @Summary Список сообщений канала
// @Description Возвращает сообщения канала с пагинацией по курсору (before) и лимиту.
// @Tags Мессенджер
// @Produce json
// @Param channelId path string true "ID канала"
// @Param before query string false "Курсор — дата RFC3339, сообщения до этой даты"
// @Param limit query int false "Количество сообщений (по умолчанию 50)"
// @Success 200 {array} MessageResponse "Список сообщений"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/messenger/channels/{channelId}/messages [get]
func (h *Handler) ListMessages(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}

	var before *time.Time
	beforeStr := c.Query("before")
	if beforeStr != "" {
		t, err := time.Parse(time.RFC3339, beforeStr)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "некорректный формат before (ожидается RFC3339)")
		}
		before = &t
	}

	limit := queryInt(c, "limit", 50)

	resp, err := h.service.ListMessages(c.Context(), channelID, before, limit)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// EditMessage godoc
// @Summary Редактировать сообщение
// @Description Редактирует текст ранее отправленного сообщения. Доступно только автору.
// @Tags Мессенджер
// @Accept json
// @Produce json
// @Param channelId path string true "ID канала"
// @Param messageId path string true "ID сообщения"
// @Param body body EditMessageRequest true "Новый текст"
// @Success 200 {object} MessageResponse "Сообщение обновлено"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Security BearerAuth
// @Router /api/messenger/channels/{channelId}/messages/{messageId} [put]
func (h *Handler) EditMessage(c fiber.Ctx) error {
	messageID := parseMessageID(c)
	if messageID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "message ID обязателен")
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req EditMessageRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.EditMessage(c.Context(), messageID, userID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// DeleteMessage godoc
// @Summary Удалить сообщение
// @Description Удаляет сообщение из канала. Доступно автору или администратору канала.
// @Tags Мессенджер
// @Param channelId path string true "ID канала"
// @Param messageId path string true "ID сообщения"
// @Success 204 "Сообщение удалено"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Security BearerAuth
// @Router /api/messenger/channels/{channelId}/messages/{messageId} [delete]
func (h *Handler) DeleteMessage(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	messageID := parseMessageID(c)
	if messageID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "message ID обязателен")
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	if err := h.service.DeleteMessage(c.Context(), messageID, userID, channelID.String(), false); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ToggleReaction godoc
// @Summary Переключить реакцию
// @Description Добавляет или убирает реакцию (эмодзи) на сообщение.
// @Tags Мессенджер
// @Accept json
// @Produce json
// @Param channelId path string true "ID канала"
// @Param messageId path string true "ID сообщения"
// @Param body body ToggleReactionRequest true "Эмодзи реакции"
// @Success 200 {object} map[string]bool "added: true/false"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/messenger/channels/{channelId}/messages/{messageId}/reactions [post]
func (h *Handler) ToggleReaction(c fiber.Ctx) error {
	messageID := parseMessageID(c)
	if messageID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "message ID обязателен")
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req ToggleReactionRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	added, err := h.service.ToggleReaction(c.Context(), messageID, userID, req.Emoji)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"added": added})
}

// TogglePin godoc
// @Summary Закрепить/открепить сообщение
// @Description Переключает закрепление сообщения в канале.
// @Tags Мессенджер
// @Produce json
// @Param channelId path string true "ID канала"
// @Param messageId path string true "ID сообщения"
// @Success 200 {object} map[string]bool "pinned: true/false"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/messenger/channels/{channelId}/messages/{messageId}/pin [post]
func (h *Handler) TogglePin(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	messageID := parseMessageID(c)
	if messageID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "message ID обязателен")
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	pinned, err := h.service.TogglePin(c.Context(), channelID, messageID, userID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"pinned": pinned})
}

// ForwardMessage godoc
// @Summary Переслать сообщение
// @Description Пересылает сообщение в другой канал.
// @Tags Мессенджер
// @Accept json
// @Produce json
// @Param channelId path string true "ID исходного канала"
// @Param messageId path string true "ID сообщения"
// @Param body body ForwardMessageRequest true "ID целевого канала"
// @Success 201 {object} MessageResponse "Сообщение переслано"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/messenger/channels/{channelId}/messages/{messageId}/forward [post]
func (h *Handler) ForwardMessage(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	messageID := parseMessageID(c)
	if messageID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "message ID обязателен")
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req ForwardMessageRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.ForwardMessage(c.Context(), channelID, userID, messageID, req.TargetChannelID)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListThreadReplies godoc
// @Summary Список ответов в треде
// @Description Возвращает ответы на указанное сообщение (тред) с пагинацией.
// @Tags Мессенджер
// @Produce json
// @Param channelId path string true "ID канала"
// @Param messageId path string true "ID родительского сообщения"
// @Param before query string false "Курсор — дата RFC3339"
// @Param limit query int false "Количество (по умолчанию 50)"
// @Success 200 {array} MessageResponse "Ответы в треде"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/messenger/channels/{channelId}/messages/{messageId}/replies [get]
func (h *Handler) ListThreadReplies(c fiber.Ctx) error {
	messageID := parseMessageID(c)
	if messageID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "message ID обязателен")
	}

	var before *time.Time
	beforeStr := c.Query("before")
	if beforeStr != "" {
		t, err := time.Parse(time.RFC3339, beforeStr)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "некорректный формат before (ожидается RFC3339)")
		}
		before = &t
	}

	limit := queryInt(c, "limit", 50)

	resp, err := h.service.ListThreadReplies(c.Context(), messageID, before, limit)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// ---------------------------------------------------------------------------
// Search handlers
// ---------------------------------------------------------------------------

// SearchMessages godoc
// @Summary Поиск сообщений
// @Description Полнотекстовый поиск по сообщениям мессенджера с опциональной фильтрацией по каналу.
// @Tags Мессенджер
// @Produce json
// @Param query query string true "Поисковый запрос"
// @Param channel_id query string false "ID канала для фильтрации"
// @Param limit query int false "Лимит результатов (по умолчанию 20)"
// @Success 200 {array} MessageResponse "Результаты поиска"
// @Failure 400 {object} fiber.Error "Параметр query обязателен"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/messenger/search [get]
func (h *Handler) SearchMessages(c fiber.Ctx) error {
	query := c.Query("query")
	if query == "" {
		return fiber.NewError(fiber.StatusBadRequest, "параметр query обязателен")
	}
	channelID := c.Query("channel_id")
	limit := queryInt(c, "limit", 20)

	resp, err := h.service.SearchMessages(c.Context(), query, channelID, limit)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// ---------------------------------------------------------------------------
// Bookmark handlers
// ---------------------------------------------------------------------------

// ListBookmarks godoc
// @Summary Список закладок
// @Description Возвращает все сохранённые закладки на сообщения текущего пользователя.
// @Tags Мессенджер
// @Produce json
// @Success 200 {array} BookmarkResponse "Список закладок"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/messenger/bookmarks [get]
func (h *Handler) ListBookmarks(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	resp, err := h.service.ListBookmarks(c.Context(), userID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// AddBookmark godoc
// @Summary Добавить закладку
// @Description Сохраняет сообщение в закладки текущего пользователя.
// @Tags Мессенджер
// @Accept json
// @Param body body AddBookmarkRequest true "ID сообщения и канала"
// @Success 201 "Закладка добавлена"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/messenger/bookmarks [post]
func (h *Handler) AddBookmark(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req AddBookmarkRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	chID, err := uuid.Parse(req.ChannelID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный channel_id")
	}

	if err := h.service.AddBookmark(c.Context(), userID, req.MessageID, chID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusCreated)
}

// RemoveBookmark godoc
// @Summary Удалить закладку
// @Description Удаляет закладку на сообщение.
// @Tags Мессенджер
// @Param bookmarkId path string true "ID закладки"
// @Success 204 "Закладка удалена"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/messenger/bookmarks/{bookmarkId} [delete]
func (h *Handler) RemoveBookmark(c fiber.Ctx) error {
	idStr := utils.CopyString(c.Params("bookmarkId"))
	if idStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "bookmark ID обязателен")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный bookmark ID")
	}

	if err := h.service.RemoveBookmark(c.Context(), id); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ---------------------------------------------------------------------------
// Typing handler
// ---------------------------------------------------------------------------

// SendTyping godoc
// @Summary Отправить индикатор набора
// @Description Публикует событие "пользователь печатает" для участников канала.
// @Tags Мессенджер
// @Param channelId path string true "ID канала"
// @Success 204 "Событие отправлено"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/messenger/channels/{channelId}/typing [post]
func (h *Handler) SendTyping(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	if err := h.service.PublishTyping(c.Context(), channelID.String(), userID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ---------------------------------------------------------------------------
// Presence handlers
// ---------------------------------------------------------------------------

// GetPresence godoc
// @Summary Получить статус присутствия
// @Description Возвращает статус онлайн/оффлайн для указанных пользователей.
// @Tags Мессенджер
// @Produce json
// @Param user_ids query string true "Список ID пользователей через запятую"
// @Success 200 {object} map[string]PresenceStatus "Статусы пользователей"
// @Failure 400 {object} fiber.Error "Параметр user_ids обязателен"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/messenger/presence [get]
func (h *Handler) GetPresence(c fiber.Ctx) error {
	userIDsStr := c.Query("user_ids")
	if userIDsStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "параметр user_ids обязателен")
	}

	userIDs := splitComma(userIDsStr)
	statuses, err := h.service.GetPresence(c.Context(), userIDs)
	if err != nil {
		return err
	}

	return c.JSON(statuses)
}

// SetCustomStatus godoc
// @Summary Установить пользовательский статус
// @Description Устанавливает текстовый статус пользователя с опциональной длительностью.
// @Tags Мессенджер
// @Accept json
// @Param body body object true "text и duration статуса"
// @Success 200 "Статус обновлён"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/messenger/presence/status [put]
func (h *Handler) SetCustomStatus(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req struct {
		Text     string `json:"text"`
		Duration string `json:"duration"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	if err := h.service.SetCustomStatus(c.Context(), userID.String(), req.Text, req.Duration); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

// Heartbeat godoc
// @Summary Heartbeat присутствия
// @Description Обновляет таймстамп активности пользователя для отслеживания онлайн-статуса.
// @Tags Мессенджер
// @Success 200 "Heartbeat принят"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/messenger/presence/heartbeat [post]
func (h *Handler) Heartbeat(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	if err := h.service.Heartbeat(c.Context(), userID.String()); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

// ---------------------------------------------------------------------------
// Unread handlers
// ---------------------------------------------------------------------------

// GetUnreadCounts godoc
// @Summary Счётчики непрочитанных
// @Description Возвращает количество непрочитанных сообщений по каналам для текущего пользователя.
// @Tags Мессенджер
// @Produce json
// @Success 200 {object} map[string]int "Счётчики непрочитанных по каналам"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/messenger/unread [get]
func (h *Handler) GetUnreadCounts(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	counts, err := h.service.GetUnreadCounts(c.Context(), userID.String())
	if err != nil {
		return err
	}

	return c.JSON(counts)
}

// MarkChannelRead godoc
// @Summary Отметить канал прочитанным
// @Description Сбрасывает счётчик непрочитанных сообщений для указанного канала.
// @Tags Мессенджер
// @Param channelId path string true "ID канала"
// @Success 200 "Канал отмечен прочитанным"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/messenger/channels/{channelId}/read [post]
func (h *Handler) MarkChannelRead(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	if err := h.service.MarkChannelRead(c.Context(), userID.String(), channelID.String()); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

// ---------------------------------------------------------------------------
// Pinned messages handlers
// ---------------------------------------------------------------------------

// ListPins godoc
// @Summary Список закреплённых сообщений
// @Description Возвращает все закреплённые сообщения в указанном канале.
// @Tags Мессенджер
// @Produce json
// @Param channelId path string true "ID канала"
// @Success 200 {array} MessageResponse "Закреплённые сообщения"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/messenger/channels/{channelId}/pins [get]
func (h *Handler) ListPins(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}

	pins, err := h.service.ListPins(c.Context(), channelID)
	if err != nil {
		return err
	}

	return c.JSON(pins)
}

// ---------------------------------------------------------------------------
// Star / Mute handlers
// ---------------------------------------------------------------------------

// ToggleStar godoc
// @Summary Переключить избранное
// @Description Добавляет или убирает канал из избранного текущего пользователя.
// @Tags Мессенджер
// @Produce json
// @Param channelId path string true "ID канала"
// @Success 200 {object} map[string]bool "starred: true/false"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/messenger/channels/{channelId}/star [post]
func (h *Handler) ToggleStar(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	starred, err := h.service.ToggleStar(c.Context(), channelID, userID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"starred": starred})
}

// MuteChannelRequest — тело запроса на мьют канала.
type MuteChannelRequest struct {
	Duration string `json:"duration"` // "1h", "2h", "8h", "1d", "forever"
}

// MuteChannel godoc
// @Summary Замьютить канал
// @Description Отключает уведомления канала на указанный период.
// @Tags Мессенджер
// @Accept json
// @Produce json
// @Param channelId path string true "ID канала"
// @Param body body MuteChannelRequest true "Длительность мьюта"
// @Success 200 {object} map[string]string "muted_until: дата"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/messenger/channels/{channelId}/mute [post]
func (h *Handler) MuteChannel(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req MuteChannelRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	mutedUntil, err := h.service.MuteChannel(c.Context(), channelID, userID, req.Duration)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"muted_until": mutedUntil})
}

// UnmuteChannel godoc
// @Summary Размьютить канал
// @Description Включает уведомления канала обратно.
// @Tags Мессенджер
// @Param channelId path string true "ID канала"
// @Success 200 "Уведомления включены"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/messenger/channels/{channelId}/unmute [post]
func (h *Handler) UnmuteChannel(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	if err := h.service.UnmuteChannel(c.Context(), channelID, userID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

// ---------------------------------------------------------------------------
// DM handlers
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Poll handlers
// ---------------------------------------------------------------------------

// parsePollID извлекает и валидирует pollId из параметров пути.
func parsePollID(c fiber.Ctx) (uuid.UUID, error) {
	idStr := utils.CopyString(c.Params("pollId"))
	if idStr == "" {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "poll ID обязателен")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "некорректный poll ID")
	}
	return id, nil
}

// CreatePoll godoc
// @Summary Создать опрос
// @Description Создаёт новый опрос в указанном канале мессенджера.
// @Tags Мессенджер
// @Accept json
// @Produce json
// @Param channelId path string true "ID канала"
// @Param body body CreatePollRequest true "Данные опроса"
// @Success 201 {object} PollResponse "Опрос создан"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/messenger/channels/{channelId}/polls [post]
func (h *Handler) CreatePoll(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req CreatePollRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.CreatePoll(c.Context(), channelID, userID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// VotePoll godoc
// @Summary Проголосовать в опросе
// @Description Отправляет голос текущего пользователя за выбранный вариант.
// @Tags Мессенджер
// @Accept json
// @Produce json
// @Param pollId path string true "ID опроса"
// @Param body body VotePollRequest true "Выбранный вариант"
// @Success 200 {object} PollResponse "Голос учтён"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/messenger/polls/{pollId}/vote [post]
func (h *Handler) VotePoll(c fiber.Ctx) error {
	pollID, err := parsePollID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req VotePollRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.VotePoll(c.Context(), pollID, userID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// ClosePoll godoc
// @Summary Закрыть опрос
// @Description Завершает опрос и фиксирует результаты. Доступно только создателю.
// @Tags Мессенджер
// @Produce json
// @Param pollId path string true "ID опроса"
// @Success 200 {object} PollResponse "Опрос закрыт"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Security BearerAuth
// @Router /api/messenger/polls/{pollId}/close [post]
func (h *Handler) ClosePoll(c fiber.Ctx) error {
	pollID, err := parsePollID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	resp, err := h.service.ClosePoll(c.Context(), pollID, userID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// GetPollResults godoc
// @Summary Результаты опроса
// @Description Возвращает текущие результаты опроса, включая голоса и процентное распределение.
// @Tags Мессенджер
// @Produce json
// @Param pollId path string true "ID опроса"
// @Success 200 {object} PollResponse "Результаты опроса"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/messenger/polls/{pollId} [get]
func (h *Handler) GetPollResults(c fiber.Ctx) error {
	pollID, err := parsePollID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	resp, err := h.service.GetPollResults(c.Context(), pollID, userID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// ---------------------------------------------------------------------------
// DM handlers
// ---------------------------------------------------------------------------

// CreateOrGetDM godoc
// @Summary Создать или получить личную переписку
// @Description Создаёт новый DM-канал с указанным пользователем или возвращает существующий.
// @Tags Мессенджер
// @Accept json
// @Produce json
// @Param body body CreateDMRequest true "ID целевого пользователя"
// @Success 200 {object} ChannelResponse "DM-канал"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/messenger/dm [post]
func (h *Handler) CreateOrGetDM(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req CreateDMRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	targetID, err := uuid.Parse(req.TargetUserID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный target_user_id")
	}

	resp, err := h.service.CreateOrGetDM(c.Context(), userID, targetID)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}
