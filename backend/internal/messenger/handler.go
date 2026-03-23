package messenger

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/google/uuid"
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

// CreateChannel обрабатывает POST /api/messenger/channels
func (h *Handler) CreateChannel(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
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
	userID, err := httputil.UserIDFromCtx(c)
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
	userID, err := httputil.UserIDFromCtx(c)
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
	userID, err := httputil.UserIDFromCtx(c)
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
	userID, err := httputil.UserIDFromCtx(c)
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
	userID, err := httputil.UserIDFromCtx(c)
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
	userID, err := httputil.UserIDFromCtx(c)
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
	userID, err := httputil.UserIDFromCtx(c)
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
	userID, err := httputil.UserIDFromCtx(c)
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
	userID, err := httputil.UserIDFromCtx(c)
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
	userID, err := httputil.UserIDFromCtx(c)
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

// AddBookmark обрабатывает POST /api/messenger/bookmarks
func (h *Handler) AddBookmark(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
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
// Typing handler
// ---------------------------------------------------------------------------

// SendTyping обрабатывает POST /api/messenger/channels/:channelId/typing
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
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req struct {
		Text     string `json:"text"`
		Duration string `json:"duration"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	if err := h.service.SetCustomStatus(c.Context(), userID.String(), req.Text, req.Duration); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

// Heartbeat обрабатывает POST /api/messenger/presence/heartbeat
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

// GetUnreadCounts обрабатывает GET /api/messenger/unread
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

// MarkChannelRead обрабатывает POST /api/messenger/channels/:channelId/read
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
// Star / Mute handlers
// ---------------------------------------------------------------------------

// ToggleStar обрабатывает POST /api/messenger/channels/:channelId/star
// Переключает флаг is_starred для текущего пользователя.
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

// MuteChannel обрабатывает POST /api/messenger/channels/:channelId/mute
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
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	mutedUntil, err := h.service.MuteChannel(c.Context(), channelID, userID, req.Duration)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"muted_until": mutedUntil})
}

// UnmuteChannel обрабатывает POST /api/messenger/channels/:channelId/unmute
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

// CreatePoll обрабатывает POST /api/messenger/channels/:channelId/polls
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
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.CreatePoll(c.Context(), channelID, userID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// VotePoll обрабатывает POST /api/messenger/polls/:pollId/vote
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
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.VotePoll(c.Context(), pollID, userID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// ClosePoll обрабатывает POST /api/messenger/polls/:pollId/close
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

// GetPollResults обрабатывает GET /api/messenger/polls/:pollId
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

// CreateOrGetDM обрабатывает POST /api/messenger/dm
func (h *Handler) CreateOrGetDM(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
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
