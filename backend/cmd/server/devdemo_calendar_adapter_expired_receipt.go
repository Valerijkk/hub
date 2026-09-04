//go:build !qa && !release

package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent/calendarmutationreceipt"
	"github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/user"
	calendarmod "github.com/valerii/hub/backend/internal/calendar"
	"github.com/valerii/hub/backend/internal/devdemo"
)

// recoverExpiredCalendarEventCreate keeps the deterministic dev dataset
// replayable after the production receipt TTL without changing that TTL or
// mutating receipt state. Every other create conflict remains terminal.
func (a *calendarDemoAdapter) recoverExpiredCalendarEventCreate(
	ctx context.Context,
	actorID uuid.UUID,
	idempotencyKey string,
	command calendarmod.CalendarEventCreateCommand,
	requested map[uuid.UUID]calendarDemoRequestedAttendee,
	requestedOrder []uuid.UUID,
) (devdemo.CalendarDemoCreateResult, error) {
	var zero devdemo.CalendarDemoCreateResult
	if a == nil || a.db == nil || a.eventReadService == nil || ctx == nil || actorID == uuid.Nil || len(requested) != len(requestedOrder) {
		return zero, errors.New("invalid expired calendar demo receipt recovery")
	}

	organizations, err := a.db.Organization.Query().
		Where(organization.HasUsersWith(
			user.IDEQ(actorID),
			user.StatusEQ(user.StatusActive),
		)).
		Limit(2).
		All(ctx)
	if err != nil || len(organizations) != 1 || organizations[0] == nil || organizations[0].ID == uuid.Nil {
		return zero, errors.New("calendar demo actor organization is unavailable")
	}

	keyHash := sha256.Sum256([]byte(idempotencyKey))
	receipts, err := a.db.CalendarMutationReceipt.Query().
		Where(
			calendarmutationreceipt.OrganizationIDEQ(organizations[0].ID),
			calendarmutationreceipt.ActorUserIDEQ(actorID),
			calendarmutationreceipt.OperationEQ(calendarmutationreceipt.OperationEventCreate),
			calendarmutationreceipt.IdempotencyKeyHashEQ(keyHash[:]),
		).
		Limit(2).
		All(ctx)
	if err != nil || len(receipts) != 1 || receipts[0] == nil {
		return zero, errors.New("expired calendar demo receipt is unavailable")
	}

	receipt := receipts[0]
	requestHash := command.RequestHash()
	now := time.Now().UTC().Truncate(time.Second)
	if receipt.State != calendarmutationreceipt.StateCompleted || receipt.ResponseStatus == nil || *receipt.ResponseStatus != 201 ||
		receipt.ResponseJSON == nil || receipt.CalendarEventID == nil || *receipt.CalendarEventID == uuid.Nil ||
		receipt.ExpiresAt.IsZero() || receipt.ExpiresAt.After(now) || !bytes.Equal(receipt.RequestHash, requestHash[:]) {
		return zero, errors.New("expired calendar demo receipt is invalid")
	}

	var stored calendarDemoCreateReceipt
	if err := json.Unmarshal(*receipt.ResponseJSON, &stored); err != nil || stored.ID != *receipt.CalendarEventID ||
		stored.Version == 0 || stored.OrganizerID == nil || *stored.OrganizerID != actorID || stored.Status != calendarmod.CalendarEventStatusConfirmed {
		return zero, errors.New("expired calendar demo receipt response is invalid")
	}

	live, err := a.GetEvent(ctx, actorID, *receipt.CalendarEventID)
	if err != nil || live.EventID != *receipt.CalendarEventID || live.Status != "confirmed" {
		return zero, errors.New("expired calendar demo event is not live")
	}
	liveAttendees, err := calendarDemoReplayCreateAttendees(actorID, requested, requestedOrder, live.Attendees)
	if err != nil {
		return zero, err
	}

	return devdemo.CalendarDemoCreateResult{
		EventID:   live.EventID,
		Version:   live.Version,
		Sequence:  live.Sequence,
		Attendees: liveAttendees,
		Replayed:  true,
	}, nil
}
