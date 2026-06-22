package tracker

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	apperr "github.com/valerii/hub/backend/internal/platform/errors"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// ListAutomationRules — Список правил автоматизации
// @Summary Список правил автоматизации
// @Description Возвращает workflow automation rules проекта (PROD-01)
// @Tags Трекер
// @Produce json
// @Param projectId path string true "ID проекта"
// @Success 200 {array} AutomationRuleResponse
// @Security BearerAuth
// @Router /projects/{projectId}/automation-rules [get]
func (h *Handler) ListAutomationRules(c fiber.Ctx) error {
	if h.automation == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "automation not configured")
	}
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	resp, err := h.automation.ListRules(c.Context(), projectID)
	if err != nil {
		return err
	}
	return c.JSON(resp)
}

// CreateAutomationRule — Создать правило автоматизации
// @Summary Создать правило автоматизации
// @Description Создаёт rule (max 3 per project, PROD-01)
// @Tags Трекер
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param body body CreateAutomationRuleRequest true "Rule"
// @Success 201 {object} AutomationRuleResponse
// @Security BearerAuth
// @Router /projects/{projectId}/automation-rules [post]
func (h *Handler) CreateAutomationRule(c fiber.Ctx) error {
	if h.automation == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "automation not configured")
	}
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, apperr.MsgUnauthorized)
	}
	var req CreateAutomationRuleRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}
	resp, err := h.automation.CreateRule(c.Context(), projectID, userID, req)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

// UpdateAutomationRule — Обновить правило автоматизации
// @Summary Обновить правило автоматизации
// @Tags Трекер
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param ruleId path string true "ID правила"
// @Param body body UpdateAutomationRuleRequest true "Updates"
// @Success 200 {object} AutomationRuleResponse
// @Security BearerAuth
// @Router /projects/{projectId}/automation-rules/{ruleId} [patch]
func (h *Handler) UpdateAutomationRule(c fiber.Ctx) error {
	if h.automation == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "automation not configured")
	}
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	ruleID, err := uuid.Parse(c.Params("ruleId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid rule id")
	}
	var req UpdateAutomationRuleRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}
	resp, err := h.automation.UpdateRule(c.Context(), projectID, ruleID, req)
	if err != nil {
		return err
	}
	return c.JSON(resp)
}

// DeleteAutomationRule — Удалить правило автоматизации
// @Summary Удалить правило автоматизации
// @Tags Трекер
// @Param projectId path string true "ID проекта"
// @Param ruleId path string true "ID правила"
// @Success 204 "No Content"
// @Security BearerAuth
// @Router /projects/{projectId}/automation-rules/{ruleId} [delete]
func (h *Handler) DeleteAutomationRule(c fiber.Ctx) error {
	if h.automation == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "automation not configured")
	}
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	ruleID, err := uuid.Parse(c.Params("ruleId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid rule id")
	}
	if err := h.automation.DeleteRule(c.Context(), projectID, ruleID); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}
