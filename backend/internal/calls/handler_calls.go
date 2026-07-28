package calls

import (
	"errors"
	"log/slog"
	"net/url"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/google/uuid"
	apperr "github.com/valerii/hub/backend/internal/platform/errors"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// CreateCall — Создать звонок
// @Summary Создать звонок
// @Description Создаёт новый голосовой/видеозвонок в указанном канале.
// @Tags Звонки
// @Accept json
// @Produce json
// @Param body body CreateCallRequest true "Данные звонка (channel_id)"
// @Success 201 {object} CallResponse "Звонок создан"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет членства в канале"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/calls [post]
func (h *Handler) CreateCall(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req CreateCallRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	if req.ChannelID == uuid.Nil {
		return fiber.NewError(fiber.StatusBadRequest, "channel_id обязателен")
	}

	userName := userNameFromCtx(c)
	resp, err := h.service.CreateCall(c.Context(), req.ChannelID, userID, userName)
	if err != nil {
		// A2-002: non-member -> 403
		if errors.Is(err, ErrCallForbidden) {
			return fiber.NewError(fiber.StatusForbidden, "недостаточно прав для создания звонка")
		}
		slog.Error("ошибка создания звонка", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось создать звонок")
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// CreateInstantCall — Создать быстрый звонок
// @Summary Создать быстрый звонок
// @Description Создаёт мгновенный звонок без привязки к каналу. Генерирует invite_code для ссылки.
// @Tags Звонки
// @Accept json
// @Produce json
// @Param body body object false "Опциональный title"
// @Success 201 {object} CallResponse "Звонок создан"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/calls/instant [post]
func (h *Handler) CreateInstantCall(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req struct {
		Title string `json:"title"`
	}
	_ = c.Bind().JSON(&req)

	userName := userNameFromCtx(c)
	resp, err := h.service.CreateInstantCall(c.Context(), userID, userName, req.Title)
	if err != nil {
		slog.Error("ошибка создания instant звонка", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось создать звонок")
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// JoinByInviteCode — Присоединиться по invite-коду
// @Summary Присоединиться по invite-коду
// @Description Присоединяет пользователя к звонку по пригласительному коду.
// @Tags Звонки
// @Produce json
// @Param code path string true "Пригласительный код"
// @Success 200 {object} CallResponse "Данные звонка с токеном"
// @Failure 400 {object} fiber.Error "Код не указан"
// @Failure 404 {object} fiber.Error "Звонок не найден"
// @Security BearerAuth
// @Router /api/calls/invite/{code} [get]
func (h *Handler) JoinByInviteCode(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	rawCode := utils.CopyString(c.Params("code"))
	decoded, err := url.PathUnescape(rawCode)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invite code обязателен")
	}
	code := strings.TrimSpace(decoded)
	if code == "" {
		return fiber.NewError(fiber.StatusBadRequest, "invite code обязателен")
	}

	userName := userNameFromCtx(c)
	resp, err := h.service.JoinByInviteCode(c.Context(), code, userID, userName)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "звонок не найден или уже завершён")
	}

	return c.JSON(resp)
}

// GetActiveCall — Получить активный звонок канала
// @Summary Получить активный звонок канала
// @Description Возвращает текущий активный звонок в указанном канале (или null).
// @Tags Звонки
// @Produce json
// @Param channelId path string true "ID канала"
// @Success 200 {object} CallResponse "Активный звонок или null"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/calls/channel/{channelId} [get]
func (h *Handler) GetActiveCall(c fiber.Ctx) error {
	idStr := utils.CopyString(c.Params("channelId"))
	channelID, err := uuid.Parse(idStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный channel ID")
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	resp, err := h.service.GetActiveCallByChannel(c.Context(), channelID, userID)
	if err != nil {
		if errors.Is(err, ErrCallForbidden) {
			return fiber.NewError(fiber.StatusForbidden, "недостаточно прав для просмотра звонка")
		}
		slog.Error("ошибка получения активного звонка", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка поиска звонка")
	}

	if resp == nil {
		return c.JSON(fiber.Map{"call": nil})
	}

	return c.JSON(resp)
}

// JoinCall — Присоединиться к звонку
// @Summary Присоединиться к звонку
// @Description Присоединяет текущего пользователя к звонку и возвращает LiveKit токен.
// @Tags Звонки
// @Produce json
// @Param callId path string true "ID звонка"
// @Success 200 {object} CallResponse "Токен для подключения"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/calls/{callId}/join [post]
func (h *Handler) JoinCall(c fiber.Ctx) error {
	callID, err := parseCallID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	userName := userNameFromCtx(c)
	resp, err := h.service.JoinCall(c.Context(), callID, userID, userName)
	if err != nil {
		// HI-03: 403 для запретов, 500 для прочих ошибок.
		if errors.Is(err, ErrCallForbidden) {
			return fiber.NewError(fiber.StatusForbidden, "недостаточно прав для присоединения к звонку")
		}
		slog.Error("ошибка присоединения к звонку", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось присоединиться к звонку")
	}

	return c.JSON(resp)
}

// LeaveCall — Покинуть звонок
// @Summary Покинуть звонок
// @Description Текущий пользователь покидает звонок.
// @Tags Звонки
// @Param callId path string true "ID звонка"
// @Success 200 "Пользователь покинул звонок"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/calls/{callId}/leave [post]
func (h *Handler) LeaveCall(c fiber.Ctx) error {
	callID, err := parseCallID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	if err := h.service.LeaveCall(c.Context(), callID, userID); err != nil {
		slog.Error("ошибка выхода из звонка", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось покинуть звонок")
	}

	return c.SendStatus(fiber.StatusOK)
}

// EndCall — Завершить звонок
// @Summary Завершить звонок
// @Description Завершает звонок для всех участников.
// @Tags Звонки
// @Param callId path string true "ID звонка"
// @Success 200 "Звонок завершён"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/calls/{callId}/end [post]
func (h *Handler) EndCall(c fiber.Ctx) error {
	callID, err := parseCallID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	if err := h.service.EndCall(c.Context(), callID, userID); err != nil {
		// HI-03: 403 для запретов, 500 для прочих ошибок.
		if errors.Is(err, ErrCallForbidden) {
			return fiber.NewError(fiber.StatusForbidden, "только создатель или владелец/админ канала может завершить звонок")
		}
		slog.Error("ошибка завершения звонка", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось завершить звонок")
	}

	return c.SendStatus(fiber.StatusOK)
}

// RaiseHand — Поднять/опустить руку
// @Summary Поднять/опустить руку
// @Description Переключает состояние "поднятой руки" участника в звонке.
// @Tags Звонки
// @Accept json
// @Produce json
// @Param callId path string true "ID звонка"
// @Param body body RaiseHandRequest true "raised: true/false"
// @Success 200 {object} map[string]bool "raised: true/false"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Не участник звонка"
// @Failure 404 {object} fiber.Error "Звонок не найден"
// @Security BearerAuth
// @Router /api/calls/{callId}/raise-hand [post]
func (h *Handler) RaiseHand(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req RaiseHandRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	callID, err := parseCallID(c)
	if err != nil {
		return err
	}

	// A2-006: authz в сервисе (active participant), без прямого доступа к repo.
	if err := h.service.RaiseHand(c.Context(), callID, userID, req.Raised); err != nil {
		if errors.Is(err, ErrCallNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "звонок не найден")
		}
		if errors.Is(err, ErrCallForbidden) {
			return fiber.NewError(fiber.StatusForbidden, "недостаточно прав для поднятия руки")
		}
		slog.Error("ошибка поднятия руки", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось обновить статус руки")
	}

	return c.JSON(fiber.Map{"raised": req.Raised})
}

// SendChatMessage — Отправить сообщение в чат звонка
// @Summary Отправить сообщение в чат звонка
// @Description Отправляет текстовое сообщение в чат текущего звонка.
// @Tags Звонки
// @Accept json
// @Produce json
// @Param callId path string true "ID звонка"
// @Param body body object true "text: текст сообщения"
// @Success 201 {object} CallChatMessage "Сообщение отправлено"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа к чату звонка"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/calls/{callId}/chat [post]
func (h *Handler) SendChatMessage(c fiber.Ctx) error {
	callID, err := parseCallID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req struct {
		Text      string `json:"text"`
		Ephemeral bool   `json:"ephemeral"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}
	if req.Text == "" {
		return fiber.NewError(fiber.StatusBadRequest, "текст сообщения обязателен")
	}

	userName := userNameFromCtx(c)
	msg, err := h.service.SendChatMessage(c.Context(), callID.String(), userID.String(), userName, req.Text, req.Ephemeral)
	if err != nil {
		if errors.Is(err, ErrCallForbidden) {
			return fiber.NewError(fiber.StatusForbidden, "недостаточно прав для чата звонка")
		}
		slog.Error("ошибка отправки сообщения чата звонка", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось отправить сообщение")
	}

	return c.Status(fiber.StatusCreated).JSON(msg)
}

// ListChatMessages — Список сообщений чата звонка
// @Summary Список сообщений чата звонка
// @Description Возвращает сообщения чата звонка с пагинацией по лимиту.
// @Tags Звонки
// @Produce json
// @Param callId path string true "ID звонка"
// @Param limit query int false "Лимит сообщений (по умолчанию 50, макс 200)"
// @Success 200 {object} map[string][]CallChatMessage "messages: массив сообщений"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа к чату звонка"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/calls/{callId}/chat [get]
func (h *Handler) ListChatMessages(c fiber.Ctx) error {
	callID, err := parseCallID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, parseErr := strconv.Atoi(l); parseErr == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}

	msgs, err := h.service.ListChatMessages(c.Context(), callID.String(), userID.String(), limit)
	if err != nil {
		if errors.Is(err, ErrCallForbidden) {
			return fiber.NewError(fiber.StatusForbidden, "недостаточно прав для чтения чата звонка")
		}
		slog.Error("ошибка получения сообщений чата звонка", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось получить сообщения")
	}

	if msgs == nil {
		msgs = []CallChatMessage{}
	}

	return c.JSON(fiber.Map{"messages": msgs})
}

// GuestJoinCall — Гостевой вход в звонок по invite-коду
// @Summary Гостевой вход в звонок по invite-коду
// @Description Позволяет гостю без аккаунта присоединиться к звонку, указав имя и invite-код.
// @Tags Звонки
// @Accept json
// @Produce json
// @Param body body GuestJoinRequest true "Invite-код и имя гостя"
// @Success 200 {object} GuestJoinResponse "Данные звонка с токеном"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 404 {object} fiber.Error "Звонок не найден"
// @Router /api/calls/guest-join [post]
func (h *Handler) GuestJoinCall(c fiber.Ctx) error {
	var req GuestJoinRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный запрос")
	}
	if req.InviteCode == "" || req.GuestName == "" {
		return fiber.NewError(fiber.StatusBadRequest, "invite_code и guest_name обязательны")
	}

	resp, err := h.service.GuestJoin(c.Context(), req.InviteCode, req.GuestName)
	if err != nil {
		if errors.Is(err, ErrCallNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "звонок не найден или завершён")
		}
		slog.Error("ошибка гостевого входа", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка генерации токена")
	}

	return c.JSON(resp)
}
