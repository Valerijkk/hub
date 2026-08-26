package messenger

import (
	"strconv"
	"strings"

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

// queryIntMax — верхняя граница для query-int (совпадает с mongo ListMessages cap).
const queryIntMax = 100

// queryInt извлекает целочисленный параметр из строки запроса с дефолтным значением.
// P3-BE-09 / W3-A1-011: отрицательные и out-of-range значения → default (не уходят в List*).
func queryInt(c fiber.Ctx, key string, defaultVal int) int {
	s := c.Query(key)
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < 1 || v > queryIntMax {
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
	channels.Get("/discover", h.DiscoverChannels)
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
	messages.Post("/:messageId/promote-to-issue", h.PromoteToIssue)
	messages.Get("/:messageId/replies", h.ListThreadReplies)
	channels.Get("/:channelId/threads/:messageId", h.GetThread)
	channels.Post("/:channelId/threads/:messageId/read", h.MarkThreadRead)
	channels.Post("/:channelId/threads/:messageId/follow", h.FollowThread)
	channels.Delete("/:channelId/threads/:messageId/follow", h.UnfollowThread)

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
// Вспомогательные функции
// ---------------------------------------------------------------------------

// parseChannelID извлекает и валидирует channelId из параметров пути.
// ME-12: отклоняет uuid.Nil наряду с пустой строкой / невалидным форматом.
func parseChannelID(c fiber.Ctx) (uuid.UUID, error) {
	idStr := utils.CopyString(c.Params("channelId"))
	if idStr == "" {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "channel ID обязателен")
	}
	id, err := uuid.Parse(idStr)
	if err != nil || id == uuid.Nil {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "некорректный channel ID")
	}
	return id, nil
}

// parseMessageID извлекает messageId из параметров пути.
func parseMessageID(c fiber.Ctx) string {
	return utils.CopyString(c.Params("messageId"))
}
