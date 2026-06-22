package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// maxResendResponseBodyBytes — верхняя граница тела ответа от Resend API,
// которое мы читаем при ошибке. BE-H1: защищает от memory DoS, если
// upstream вернёт огромный body (или mitm-прокси подменит ответ).
const maxResendResponseBodyBytes = 10 * 1024 * 1024 // 10 MB

// ResendEmailClient реализует интерфейс ResendClient через HTTP API Resend.
type ResendEmailClient struct {
	apiKey    string
	fromEmail string
	client    *http.Client
}

// NewResendEmailClient создаёт новый клиент для отправки email через Resend API.
func NewResendEmailClient(apiKey, fromEmail string) *ResendEmailClient {
	return &ResendEmailClient{
		apiKey:    apiKey,
		fromEmail: fromEmail,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// resendRequest — тело запроса к Resend API /emails.
type resendRequest struct {
	From    string `json:"from"`
	To      []string `json:"to"`
	Subject string `json:"subject"`
	HTML    string `json:"html"`
}

// SendMagicLinkEmail отправляет email с кодом magic link через Resend API.
func (r *ResendEmailClient) SendMagicLinkEmail(ctx context.Context, to, code string) error {
	body := resendRequest{
		From:    r.fromEmail,
		To:      []string{to},
		Subject: "Ваш код входа в Hub",
		HTML: fmt.Sprintf(
			`<div style="font-family:sans-serif;max-width:400px;margin:0 auto;padding:32px">
				<h2 style="color:#1a1a2e">Код входа в Hub</h2>
				<p style="font-size:32px;font-weight:bold;letter-spacing:8px;color:#3b82f6;margin:24px 0">%s</p>
				<p style="color:#666">Код действителен 5 минут. Если вы не запрашивали вход — проигнорируйте это письмо.</p>
			</div>`, code),
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("resend: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("resend: create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+r.apiKey)

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("resend: send email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		// BE-H1: лимит тела ответа 10MB — защита от memory DoS через oversized upstream.
		respBody, readErr := io.ReadAll(io.LimitReader(resp.Body, maxResendResponseBodyBytes))
		if readErr != nil {
			return fmt.Errorf("resend: API error (status %d), failed to read body: %w", resp.StatusCode, readErr)
		}
		return fmt.Errorf("resend: API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}
