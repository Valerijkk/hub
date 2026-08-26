package messenger

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

func (s *MessengerService) ResolveUnreadSummary(
	ctx context.Context,
	userID uuid.UUID,
) (DashboardUnreadSummary, error) {
	if s == nil || s.unreadScope == nil {
		return DashboardUnreadSummary{}, errors.New("authorized unread scope resolver is required")
	}
	repository, ok := s.repo.(Phase151DashboardUnreadRepository)
	if !ok {
		return DashboardUnreadSummary{}, errors.New("durable dashboard unread repository is required")
	}
	return ExportPhase151BuildDashboardUnreadSummary(ctx, userID, s.unreadScope, repository)
}
