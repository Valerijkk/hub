package messenger

import (
	"time"

	"github.com/gofiber/fiber/v3"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

func (h *Handler) GetThread(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	rootMessageID := parseMessageID(c)
	if rootMessageID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "message ID обязателен")
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	var before *time.Time
	if value := c.Query("before"); value != "" {
		parsed, parseErr := time.Parse(time.RFC3339, value)
		if parseErr != nil {
			return fiber.NewError(fiber.StatusBadRequest, "некорректный формат before")
		}
		before = &parsed
	}
	response, err := h.service.GetThread(
		c.Context(),
		channelID,
		rootMessageID,
		userID,
		before,
		queryInt(c, "limit", 50),
	)
	if err != nil {
		return err
	}
	return c.JSON(response)
}

func (h *Handler) MarkThreadRead(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	rootMessageID := parseMessageID(c)
	if rootMessageID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "message ID обязателен")
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	var request MarkThreadReadRequest
	if err := c.Bind().JSON(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректная квитанция чтения треда")
	}
	advanced, err := h.service.MarkThreadRead(c.Context(), channelID, rootMessageID, userID, request)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"advanced": advanced})
}

func (h *Handler) FollowThread(c fiber.Ctx) error {
	return h.setThreadFollow(c, true)
}

func (h *Handler) UnfollowThread(c fiber.Ctx) error {
	return h.setThreadFollow(c, false)
}

func (h *Handler) setThreadFollow(c fiber.Ctx, follow bool) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	rootMessageID := parseMessageID(c)
	if rootMessageID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "message ID is required")
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	var changed bool
	if follow {
		changed, err = h.service.FollowThread(c.Context(), channelID, rootMessageID, userID)
	} else {
		changed, err = h.service.UnfollowThread(c.Context(), channelID, rootMessageID, userID)
	}
	if err != nil {
		return err
	}
	mode := ThreadFollowModeManualUnfollow
	if follow {
		mode = ThreadFollowModeManualFollow
	}
	return c.JSON(fiber.Map{"changed": changed, "mode": mode})
}