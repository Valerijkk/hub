//go:build qa && !release

package qamailcalendar

import (
	"context"

	"github.com/google/uuid"
)

type inboundRequestFixturePublishedExternalLockedExactSetLoader[T any] func(context.Context) ([]*T, error)

type inboundRequestFixturePublishedExternalLockedExactSetIDExtractor[T any] func(*T) (uuid.UUID, bool)

func observeInboundRequestFixturePublishedExternalLockedExactSet[T any](
	ctx context.Context,
	slot inboundRequestFixtureExternalPolicySlot,
	loader inboundRequestFixturePublishedExternalLockedExactSetLoader[T],
	extractID inboundRequestFixturePublishedExternalLockedExactSetIDExtractor[T],
) (inboundRequestFixtureExternalObservation, bool) {
	if ctx == nil || !slot.valid() || loader == nil || extractID == nil {
		return inboundRequestFixtureExternalObservation{}, false
	}

	rows, err := loader(ctx)
	if err != nil || len(rows) > 2 {
		return inboundRequestFixtureExternalObservation{}, false
	}

	var firstID uuid.UUID
	var secondID uuid.UUID
	for index, row := range rows {
		if row == nil {
			return inboundRequestFixtureExternalObservation{}, false
		}

		rowID, ok := extractID(row)
		if !ok || rowID == uuid.Nil {
			return inboundRequestFixtureExternalObservation{}, false
		}

		switch index {
		case 0:
			firstID = rowID
		case 1:
			secondID = rowID
		}
	}

	return newInboundRequestFixtureExternalObservation(slot, len(rows), firstID, secondID)
}
