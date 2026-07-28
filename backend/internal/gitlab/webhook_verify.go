package gitlab

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// webhookProjectHint — минимальный срез GitLab payload для привязки token к интеграции.
type webhookProjectHint struct {
	Project struct {
		PathWithNamespace string `json:"path_with_namespace"`
	} `json:"project"`
}

// VerifyWebhookToken проверяет X-Gitlab-Token секретом интеграции,
// найденной по project.path_with_namespace из payload (A5-009).
// Не перебирает секреты инстанса.
func (s *GitLabService) VerifyWebhookToken(ctx context.Context, body []byte, token string) error {
	if token == "" {
		return fiber.NewError(fiber.StatusForbidden, "отсутствует заголовок X-Gitlab-Token")
	}

	projectPath, err := extractProjectPathFromWebhookPayload(body)
	if err != nil {
		return fiber.NewError(fiber.StatusForbidden, "не удалось определить project_path webhook")
	}

	integration, err := s.repo.FindIntegrationByProjectPath(ctx, projectPath)
	if err != nil {
		return fiber.NewError(fiber.StatusForbidden, "интеграция для проекта не найдена")
	}
	if integration.WebhookSecret == "" {
		return fiber.NewError(fiber.StatusForbidden, "webhook secret не настроен")
	}

	secret := integration.WebhookSecret
	// W7-A3-003 / P2-BE-20: сравниваем SHA-256 дайджесты — равная длина,
	// ConstantTimeCompare без length-oracle на сырых строках.
	if !constantTimeEqualDigest(token, secret) {
		return fiber.NewError(fiber.StatusForbidden, "невалидный X-Gitlab-Token")
	}
	return nil
}

func constantTimeEqualDigest(a, b string) bool {
	sumA := sha256.Sum256([]byte(a))
	sumB := sha256.Sum256([]byte(b))
	return subtle.ConstantTimeCompare(sumA[:], sumB[:]) == 1
}

func extractProjectPathFromWebhookPayload(body []byte) (string, error) {
	var hint webhookProjectHint
	if err := json.Unmarshal(body, &hint); err != nil {
		return "", err
	}
	path := strings.TrimSpace(hint.Project.PathWithNamespace)
	if path == "" {
		return "", fmt.Errorf("project.path_with_namespace missing in webhook payload")
	}
	return path, nil
}
