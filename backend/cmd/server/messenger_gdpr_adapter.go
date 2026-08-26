package main

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	auditmod "github.com/valerii/hub/backend/internal/audit"
	messengermod "github.com/valerii/hub/backend/internal/messenger"
)

type messengerGDPRDataEraser struct {
	tenantSource messengermod.InstanceOrganizationIDSource
	repository   messengermod.MessengerDataEraseRepository
}

func (e *messengerGDPRDataEraser) EraseMessengerUserData(
	ctx context.Context,
	userID uuid.UUID,
) (auditmod.MessengerDataEraseResult, error) {
	if e == nil || e.tenantSource == nil || e.repository == nil || userID == uuid.Nil {
		return auditmod.MessengerDataEraseResult{}, fmt.Errorf("messenger GDPR eraser is not configured")
	}
	tenantID, err := messengermod.ExportPhase151ResolveInstanceOrganizationID(ctx, e.tenantSource)
	if err != nil {
		return auditmod.MessengerDataEraseResult{}, fmt.Errorf("resolve messenger GDPR tenant: %w", err)
	}
	result, err := messengermod.ExportPhase151EraseMessengerUserData(
		ctx, tenantID, userID, e.repository,
	)
	return auditmod.MessengerDataEraseResult{
		DeletedMessages:        result.DeletedMessages,
		DeletedReadStates:      result.DeletedReadStates,
		DeletedThreadFollows:   result.DeletedThreadFollows,
		PulledResolvedMentions: result.PulledResolvedMentions,
	}, err
}
