package mail

import (
	"context"
	"fmt"
	"strings"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/user"
)

// BroadcastRequest — маркетинговая рассылка (NOTIF-09 lite).
type BroadcastRequest struct {
	Subject string `json:"subject" validate:"required,min=3,max=200"`
	HTMLBody string `json:"html_body" validate:"required,min=10"`
	DryRun  bool   `json:"dry_run,omitempty"`
}

// BroadcastResponse — результат broadcast.
type BroadcastResponse struct {
	Recipients int  `json:"recipients"`
	Sent       int  `json:"sent"`
	DryRun     bool `json:"dry_run"`
}

// BroadcastSender — одно email (hub-mail bridge).
type BroadcastSender interface {
	SendAlert(ctx context.Context, req SendRequest) error
}

// Broadcast — email всем active users (admin only).
func (s *AdminService) Broadcast(ctx context.Context, db *ent.Client, sender BroadcastSender, req BroadcastRequest) (*BroadcastResponse, error) {
	subject := strings.TrimSpace(req.Subject)
	body := strings.TrimSpace(req.HTMLBody)
	if subject == "" || body == "" {
		return nil, ErrInvalidRequest
	}

	users, err := db.User.Query().
		Where(user.EmailNEQ("")).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	resp := &BroadcastResponse{
		Recipients: len(users),
		DryRun:     req.DryRun,
	}
	if req.DryRun || sender == nil {
		return resp, nil
	}

	for _, u := range users {
		if err := sender.SendAlert(ctx, SendRequest{
			To:       u.Email,
			Subject:  subject,
			HTMLBody: body,
			TextBody: body,
			Template: "broadcast",
		}); err != nil {
			continue
		}
		resp.Sent++
	}
	return resp, nil
}
