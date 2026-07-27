package github

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// webhookRepoHint — минимальный срез GitHub payload для привязки HMAC к интеграции.
type webhookRepoHint struct {
	Repository struct {
		Name  string `json:"name"`
		Owner struct {
			Login string `json:"login"`
		} `json:"owner"`
	} `json:"repository"`
}

// VerifyWebhookSignature проверяет X-Hub-Signature-256 секретом интеграции,
// найденной по repository из payload (A5-008). Не перебирает секреты инстанса.
func (s *GitHubService) VerifyWebhookSignature(ctx context.Context, body []byte, signature string) error {
	if signature == "" {
		return fiber.NewError(fiber.StatusForbidden, "отсутствует заголовок X-Hub-Signature-256")
	}
	const prefix = "sha256="
	if !strings.HasPrefix(signature, prefix) {
		return fiber.NewError(fiber.StatusForbidden, "некорректный формат подписи")
	}
	receivedMAC, err := hex.DecodeString(signature[len(prefix):])
	if err != nil {
		return fiber.NewError(fiber.StatusForbidden, "некорректная hex-подпись")
	}

	owner, repo, err := extractRepoFromWebhookPayload(body)
	if err != nil {
		return fiber.NewError(fiber.StatusForbidden, "не удалось определить репозиторий webhook")
	}

	integration, err := s.repo.FindIntegrationByRepo(ctx, owner, repo)
	if err != nil {
		return fiber.NewError(fiber.StatusForbidden, "интеграция для репозитория не найдена")
	}
	if integration.WebhookSecret == "" {
		return fiber.NewError(fiber.StatusForbidden, "webhook secret не настроен")
	}

	mac := hmac.New(sha256.New, []byte(integration.WebhookSecret))
	mac.Write(body)
	expectedMAC := mac.Sum(nil)
	if subtle.ConstantTimeCompare(receivedMAC, expectedMAC) != 1 {
		return fiber.NewError(fiber.StatusForbidden, "невалидная подпись webhook")
	}
	return nil
}

func extractRepoFromWebhookPayload(body []byte) (owner, repo string, err error) {
	var hint webhookRepoHint
	if err := json.Unmarshal(body, &hint); err != nil {
		return "", "", err
	}
	owner = strings.TrimSpace(hint.Repository.Owner.Login)
	repo = strings.TrimSpace(hint.Repository.Name)
	if owner == "" || repo == "" {
		return "", "", fmt.Errorf("repository missing in webhook payload")
	}
	return owner, repo, nil
}
