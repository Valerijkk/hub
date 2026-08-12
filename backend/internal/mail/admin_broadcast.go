package mail

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/valerii/hub/backend/ent/user"
)

// BroadcastRequest — маркетинговая рассылка (NOTIF-09 lite).
type BroadcastRequest struct {
	Subject  string `json:"subject" validate:"required,min=3,max=200"`
	HTMLBody string `json:"html_body" validate:"required,min=10"`
	DryRun   bool   `json:"dry_run,omitempty"`
}

// BroadcastResponse — результат broadcast.
type BroadcastResponse struct {
	Recipients int  `json:"recipients"`
	Sent       int  `json:"sent"`
	Failed     int  `json:"failed"`
	DryRun     bool `json:"dry_run"`
}

// BroadcastSender — одно email (hub-mail bridge).
type BroadcastSender interface {
	SendAlert(ctx context.Context, req SendRequest) error
}

// ActiveUserEmailSource — список email active users без *ent.Client в handler (W7-A1-005).
type ActiveUserEmailSource interface {
	ListActiveUserEmails(ctx context.Context) ([]string, error)
}

// Broadcast — email всем active users (admin only).
func (s *AdminService) Broadcast(ctx context.Context, users ActiveUserEmailSource, sender BroadcastSender, req BroadcastRequest) (*BroadcastResponse, error) {
	subject := strings.TrimSpace(req.Subject)
	body := strings.TrimSpace(req.HTMLBody)
	if subject == "" || body == "" {
		return nil, ErrInvalidRequest
	}
	if users == nil {
		return nil, fmt.Errorf("user source required")
	}

	emails, err := users.ListActiveUserEmails(ctx)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	resp := &BroadcastResponse{
		Recipients: len(emails),
		DryRun:     req.DryRun,
	}
	if req.DryRun || sender == nil {
		return resp, nil
	}

	for _, email := range emails {
		if err := sender.SendAlert(ctx, SendRequest{
			To:       email,
			Subject:  subject,
			HTMLBody: body,
			TextBody: body,
			Template: "broadcast",
		}); err != nil {
			// W7-A1-004: не глотать ошибку — считаем Failed + лог.
			slog.WarnContext(ctx, "hub-mail broadcast send failed",
				"to", email, "error", err)
			resp.Failed++
			continue
		}
		resp.Sent++
	}
	return resp, nil
}

// ListActiveUserEmails — active users с непустым email (A7-008).
func (s *EntSettingsStore) ListActiveUserEmails(ctx context.Context) ([]string, error) {
	users, err := s.db.User.Query().
		Where(
			user.EmailNEQ(""),
			user.StatusEQ(user.StatusActive),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(users))
	for _, u := range users {
		out = append(out, u.Email)
	}
	return out, nil
}

// Ensure EntSettingsStore satisfies ActiveUserEmailSource.
var _ ActiveUserEmailSource = (*EntSettingsStore)(nil)
