package messenger

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/google/uuid"
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
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// userIDFromCtx извлекает user ID из Fiber Locals.
func userIDFromCtx(c fiber.Ctx) (uuid.UUID, error) {
	idStr, ok := c.Locals("userID").(string)
	if !ok || idStr == "" {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "пользователь не аутентифицирован")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "некорректный user ID")
	}
	return id, nil
}

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

// CreateChannel обрабатывает POST /api/messenger/channels
func (h *Handler) CreateChannel(c fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	var req CreateChannelRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.CreateChannel(c.Context(), userID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListMyChannels обрабатывает GET /api/messenger/channels
func (h *Handler) ListMyChannels(c fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	resp, err := h.service.ListMyChannels(c.Context(), userID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// GetChannel обрабатывает GET /api/messenger/channels/:channelId
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

// UpdateChannel обрабатывает PUT /api/messenger/channels/:channelId
func (h *Handler) UpdateChannel(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	var req CreateChannelRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.UpdateChannel(c.Context(), channelID, userID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// DeleteChannel обрабатывает DELETE /api/messenger/channels/:channelId
func (h *Handler) DeleteChannel(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	if err := h.service.DeleteChannel(c.Context(), channelID, userID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// JoinChannel обрабатывает POST /api/messenger/channels/:channelId/join
func (h *Handler) JoinChannel(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	if err := h.service.JoinChannel(c.Context(), channelID, userID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

// LeaveChannel обрабатывает POST /api/messenger/channels/:channelId/leave
func (h *Handler) LeaveChannel(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	if err := h.service.LeaveChannel(c.Context(), channelID, userID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

// AddMembers обрабатывает POST /api/messenger/channels/:channelId/members
func (h *Handler) AddMembers(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	var req AddMembersRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	if err := h.service.AddMembers(c.Context(), channelID, userID, req.MemberIDs); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

// ListMembers обрабатывает GET /api/messenger/channels/:channelId/members
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

// SendMessage обрабатывает POST /api/messenger/channels/:channelId/messages
func (h *Handler) SendMessage(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	var req SendMessageRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.SendMessage(c.Context(), channelID, userID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListMessages обрабатывает GET /api/messenger/channels/:channelId/messages
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

// EditMessage обрабатывает PUT /api/messenger/channels/:channelId/messages/:messageId
func (h *Handler) EditMessage(c fiber.Ctx) error {
	messageID := parseMessageID(c)
	if messageID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "message ID обязателен")
	}
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	var req EditMessageRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.EditMessage(c.Context(), messageID, userID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// DeleteMessage обрабатывает DELETE /api/messenger/channels/:channelId/messages/:messageId
func (h *Handler) DeleteMessage(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	messageID := parseMessageID(c)
	if messageID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "message ID обязателен")
	}
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	if err := h.service.DeleteMessage(c.Context(), messageID, userID, channelID.String(), false); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ToggleReaction обрабатывает POST /api/messenger/channels/:channelId/messages/:messageId/reactions
func (h *Handler) ToggleReaction(c fiber.Ctx) error {
	messageID := parseMessageID(c)
	if messageID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "message ID обязателен")
	}
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	var req ToggleReactionRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	added, err := h.service.ToggleReaction(c.Context(), messageID, userID, req.Emoji)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"added": added})
}

// TogglePin обрабатывает POST /api/messenger/channels/:channelId/messages/:messageId/pin
func (h *Handler) TogglePin(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	messageID := parseMessageID(c)
	if messageID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "message ID обязателен")
	}
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	pinned, err := h.service.TogglePin(c.Context(), channelID, messageID, userID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"pinned": pinned})
}

// ForwardMessage обрабатывает POST /api/messenger/channels/:channelId/messages/:messageId/forward
func (h *Handler) ForwardMessage(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	messageID := parseMessageID(c)
	if messageID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "message ID обязателен")
	}
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	var req ForwardMessageRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.ForwardMessage(c.Context(), channelID, userID, messageID, req.TargetChannelID)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListThreadReplies обрабатывает GET /api/messenger/channels/:channelId/messages/:messageId/replies
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

// SearchMessages обрабатывает GET /api/messenger/search
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

// ListBookmarks обрабатывает GET /api/messenger/bookmarks
func (h *Handler) ListBookmarks(c fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	resp, err := h.service.ListBookmarks(c.Context(), userID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// AddBookmark обрабатывает POST /api/messenger/bookmarks
func (h *Handler) AddBookmark(c fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	var req AddBookmarkRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
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

// RemoveBookmark обрабатывает DELETE /api/messenger/bookmarks/:bookmarkId
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
// Presence handlers
// ---------------------------------------------------------------------------

// GetPresence обрабатывает GET /api/messenger/presence?user_ids=id1,id2,...
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

// SetCustomStatus обрабатывает PUT /api/messenger/presence/status
func (h *Handler) SetCustomStatus(c fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	var req struct {
		Text string `json:"text"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	if err := h.service.SetCustomStatus(c.Context(), userID.String(), req.Text); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

// Heartbeat обрабатывает POST /api/messenger/presence/heartbeat
func (h *Handler) Heartbeat(c fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
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

// GetUnreadCounts обрабатывает GET /api/messenger/unread
func (h *Handler) GetUnreadCounts(c fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	counts, err := h.service.GetUnreadCounts(c.Context(), userID.String())
	if err != nil {
		return err
	}

	return c.JSON(counts)
}

// MarkChannelRead обрабатывает POST /api/messenger/channels/:channelId/read
func (h *Handler) MarkChannelRead(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	userID, err := userIDFromCtx(c)
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

// ListPins обрабатывает GET /api/messenger/channels/:channelId/pins
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
// DM handlers
// ---------------------------------------------------------------------------

// CreateOrGetDM обрабатывает POST /api/messenger/dm
func (h *Handler) CreateOrGetDM(c fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	var req CreateDMRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
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
