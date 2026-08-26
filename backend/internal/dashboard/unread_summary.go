package dashboard

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/internal/messenger"
)

type MessengerUnreadSummaryProvider interface {
	ResolveUnreadSummary(context.Context, uuid.UUID) (messenger.DashboardUnreadSummary, error)
}

func (s *DashboardService) SetMessengerUnreadSummaryProvider(provider MessengerUnreadSummaryProvider) {
	if provider == nil {
		panic("dashboard messenger unread summary provider is required")
	}
	s.unreadSummaryProvider = provider
}

func (s *DashboardService) MessengerUnreadSummary(
	ctx context.Context,
	userID uuid.UUID,
) (messenger.DashboardUnreadSummary, error) {
	if s == nil || s.unreadSummaryProvider == nil || userID == uuid.Nil {
		return messenger.DashboardUnreadSummary{}, errors.New("dashboard messenger unread summary provider is required")
	}
	return s.unreadSummaryProvider.ResolveUnreadSummary(ctx, userID)
}

func ExportNewServiceWithUnreadSummaryProviderTest(provider MessengerUnreadSummaryProvider) *DashboardService {
	service := &DashboardService{}
	service.SetMessengerUnreadSummaryProvider(provider)
	return service
}
