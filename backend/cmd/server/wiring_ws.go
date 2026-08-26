package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	aimod "github.com/valerii/hub/backend/internal/ai"
	messengermod "github.com/valerii/hub/backend/internal/messenger"
	wsmod "github.com/valerii/hub/backend/internal/ws"
)

func wireWebSocket(d *serverDeps) {
	app := d.app
	authSvc := d.authSvc

	// --- WebSocket-шлюз ---

	wsHub := wsmod.NewHub()
	// ME-10: rootCtx (вместо локального wsCtx) — отменяется одним сигналом.
	go wsHub.Run(d.rootCtx)
	d.wsHub = wsHub

	channelAuthority := messengermod.NewWSChannelAuthority(
		messengermod.NewPGChannelAccessResolver(d.messengerEntRepo),
		d.messengerEntRepo,
	)
	channelAudience := wsmod.ChannelMembersResolver(func(ctx context.Context, channelID uuid.UUID) ([]uuid.UUID, error) {
		return channelAuthority.ResolveChannelEventAudience(ctx, channelID, "channel_event")
	})
	wsBridge := wsmod.NewBridgeWithMessengerUserEventGates(
		wsHub,
		d.nc,
		channelAudience,
		wsmod.CallChatAudienceResolver(d.callsRepo.ListCallChatAudienceUserIDs),
		channelAuthority, 		channelAuthority,
	)
	if err := wsBridge.Start(); err != nil {
		slog.Error("failed to start ws bridge", "error", err)
		os.Exit(1)
	}
	d.addCleanup(wsBridge.Stop)
	d.wsBridge = wsBridge

	d.aiSvc.SetStreamPublisher(aimod.NewWSStreamPublisher(wsHub))

	// SEC-H1: WS Authorizer гарантирует, что subscribe идёт только в
	// разрешённые комнаты (членство в канале/проекте/звонке).
	wsAuthz := wsmod.NewAuthorizer(wsmod.AuthorizerConfig{
		IsChannelMember: func(ctx context.Context, channelID, userID uuid.UUID) (bool, error) {
			return channelAuthority.AuthorizeChannelRoom(ctx, userID, channelID)
		},
		// A2-007: instant (ChannelID==nil) — IsActiveParticipant; channel — member OR participant.
		IsCallParticipantOrChannelMember: d.callsRepo.IsCallParticipantOrChannelMember,
		IsProjectMember: func(ctx context.Context, projectID, userID uuid.UUID) (bool, error) {
			_, err := d.projectRepo.FindMember(ctx, projectID, userID)
			if err != nil {
				if ent.IsNotFound(err) {
					return false, nil
				}
				return false, err
			}
			return true, nil
		},
	})
	app.Get("/ws", wsmod.Handler(wsHub, authSvc, wsAuthz))
}
