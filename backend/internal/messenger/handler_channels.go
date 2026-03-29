package messenger

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	apperr "github.com/valerii/hub/backend/internal/platform/errors"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

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
