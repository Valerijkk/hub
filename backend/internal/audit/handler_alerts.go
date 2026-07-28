package audit

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"strings"

	"github.com/gofiber/fiber/v3"
)

const alertWebhookAuthPrefix = "Bearer "

// AlertmanagerWebhookPayload — подмножество JSON webhook v4 Alertmanager.
type AlertmanagerWebhookPayload struct {
	Status            string              `json:"status"`
	Receiver          string              `json:"receiver"`
	GroupLabels       map[string]string   `json:"groupLabels"`
	CommonLabels      map[string]string   `json:"commonLabels"`
	CommonAnnotations map[string]string   `json:"commonAnnotations"`
	ExternalURL       string              `json:"externalURL"`
	Alerts            []AlertmanagerAlert `json:"alerts"`
}

// AlertmanagerAlert — одно оповещение в payload webhook.
type AlertmanagerAlert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     string            `json:"startsAt"`
	EndsAt       string            `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
}

// RegisterAlertWebhook регистрирует ingress Alertmanager (только secret, без JWT).
func (h *Handler) RegisterAlertWebhook(api fiber.Router, secret string) {
	if secret == "" {
		return
	}
	webhooks := api.Group("/webhooks", h.alertWebhookAuth(secret))
	webhooks.Post("/alertmanager", h.ReceiveAlertmanagerWebhook)
}

func (h *Handler) alertWebhookAuth(secret string) fiber.Handler {
	return func(c fiber.Ctx) error {
		auth := c.Get("Authorization")
		if !strings.HasPrefix(auth, alertWebhookAuthPrefix) {
			return fiber.NewError(fiber.StatusUnauthorized, "missing alert webhook authorization")
		}
		token := strings.TrimPrefix(auth, alertWebhookAuthPrefix)
		// W6-A4-010: сравниваем SHA-256 дайджесты — fixed-size, без length-oracle CTC.
		if !constantTimeEqualDigest(token, secret) {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid alert webhook authorization")
		}
		return c.Next()
	}
}

func constantTimeEqualDigest(a, b string) bool {
	sumA := sha256.Sum256([]byte(a))
	sumB := sha256.Sum256([]byte(b))
	return subtle.ConstantTimeCompare(sumA[:], sumB[:]) == 1
}

// ReceiveAlertmanagerWebhook godoc
// @Summary Alertmanager webhook
// @Description Ingests firing/resolved alerts into audit log (ALERT_WEBHOOK_SECRET Bearer auth)
// @Tags Webhooks
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {ALERT_WEBHOOK_SECRET}"
// @Success 204 "Alerts recorded"
// @Failure 401 {object} fiber.Error
// @Router /api/webhooks/alertmanager [post]
func (h *Handler) ReceiveAlertmanagerWebhook(c fiber.Ctx) error {
	var payload AlertmanagerWebhookPayload
	if err := json.Unmarshal(c.Body(), &payload); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid alertmanager payload")
	}

	if err := h.service.RecordAlertmanagerWebhook(c.Context(), payload, c.IP()); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to record alerts")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// RecordAlertmanagerWebhook сохраняет оповещения Alertmanager в журнале аудита.
func (s *AuditService) RecordAlertmanagerWebhook(ctx context.Context, payload AlertmanagerWebhookPayload, ipAddress string) error {
	for _, alert := range payload.Alerts {
		action := "alert.firing"
		if alert.Status == "resolved" {
			action = "alert.resolved"
		}

		entityID := alert.Fingerprint
		if entityID == "" && alert.Labels != nil {
			entityID = alert.Labels["alertname"]
		}
		if entityID == "" {
			entityID = "unknown"
		}

		details := map[string]interface{}{
			"webhook_status": payload.Status,
			"alert_status":   alert.Status,
			"receiver":       payload.Receiver,
			"labels":         alert.Labels,
			"annotations":    alert.Annotations,
			"starts_at":      alert.StartsAt,
			"ends_at":        alert.EndsAt,
			"generator_url":  alert.GeneratorURL,
			"external_url":   payload.ExternalURL,
		}

		if _, err := s.repo.CreateAuditLog(ctx, action, "alert", entityID, nil, nil, details, ipAddress, "alertmanager"); err != nil {
			return err
		}
	}
	return nil
}
