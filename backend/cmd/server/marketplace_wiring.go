package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	marketplacemod "github.com/valerii/hub/backend/internal/marketplace"
	messengermod "github.com/valerii/hub/backend/internal/messenger"
	notifmod "github.com/valerii/hub/backend/internal/notification"
)

type marketplaceNotifierAdapter struct {
	svc *notifmod.NotificationService
}

func (a *marketplaceNotifierAdapter) NotifyListingSubmitted(ctx context.Context, adminIDs []uuid.UUID, listing marketplacemod.ListingResponse, authorID uuid.UUID) error {
	if a.svc == nil {
		return nil
	}
	title := "Новое объявление на модерации"
	body := listing.AuthorName + ": " + listing.Title
	for _, adminID := range adminIDs {
		if adminID == authorID {
			continue
		}
		if err := a.svc.Dispatch(ctx, notifmod.DispatchEvent{
			UserID:     adminID,
			Type:       notifmod.TypeAssigned,
			Title:      title,
			Body:       body,
			EntityType: "marketplace_listing",
			EntityID:   listing.ID.String(),
			ActorID:    &authorID,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (a *marketplaceNotifierAdapter) NotifyListingDecision(ctx context.Context, authorID uuid.UUID, approved bool, listing marketplacemod.ListingResponse, reviewerID uuid.UUID) error {
	if a.svc == nil {
		return nil
	}
	title := "Объявление отклонено"
	if approved {
		title = "Объявление опубликовано"
	}
	body := listing.Title
	if listing.ReviewNote != "" {
		body += ": " + listing.ReviewNote
	}
	return a.svc.Dispatch(ctx, notifmod.DispatchEvent{
		UserID:     authorID,
		Type:       notifmod.TypeAssigned,
		Title:      title,
		Body:       body,
		EntityType: "marketplace_listing",
		EntityID:   listing.ID.String(),
		ActorID:    &reviewerID,
	})
}

type marketplaceChannelBroadcaster struct {
	messenger *messengermod.MessengerService
	repo      *messengermod.EntRepository
}

func (b *marketplaceChannelBroadcaster) BroadcastApprovedListing(ctx context.Context, listing marketplacemod.ListingResponse) error {
	if b.messenger == nil || b.repo == nil {
		return nil
	}
	ch, err := b.repo.FindGlobalChannelByName(ctx, "marketplace")
	if err != nil || ch == nil {
		return nil
	}
	price := "Договорная"
	if listing.PriceCents != nil {
		if *listing.PriceCents == 0 {
			price = "Бесплатно"
		} else {
			price = fmt.Sprintf("%s %.2f", listing.Currency, float64(*listing.PriceCents)/100)
		}
	}
	text := fmt.Sprintf("[Marketplace] %s\n%s\n%s · %s\n/marketplace", listing.Title, listing.Description, listing.AuthorName, price)
	_, err = b.messenger.SendBotMessage(ctx, ch.ID, messengermod.SendMessageRequest{
		ContentText: text,
	})
	if err != nil {
		slog.WarnContext(ctx, "marketplace channel broadcast failed", "error", err)
	}
	return nil
}
