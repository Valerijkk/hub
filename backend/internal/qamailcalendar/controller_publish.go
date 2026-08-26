//go:build qa && !release

package qamailcalendar

import (
	"context"

	"github.com/valerii/hub/backend/internal/mail"
	"github.com/valerii/hub/backend/internal/platform/outbox"
)

func (c *Controller) publishInboundRequestFixture(
	ctx context.Context,
	nonce string,
) (inboundRequestFixtureSnapshot, error) {
	if c == nil ||
		ctx == nil ||
		c.client == nil ||
		c.engine == nil ||
		c.random == nil ||
		c.fixtures == nil ||
		!validToken(string(c.config.token)) ||
		!ValidNonce(nonce) {
		return inboundRequestFixtureSnapshot{}, errInboundRequestFixtureControllerInput
	}

	plan, err := newInboundRequestFixturePlan(nonce)
	if err != nil {
		return inboundRequestFixtureSnapshot{}, errInboundRequestFixtureLifecycle
	}

	c.mu.Lock()
	record, found := c.fixtures[nonce]
	if !found ||
		record.nonce != nonce ||
		record.state != inboundRequestFixtureStateCreated ||
		record.plan != plan ||
		!inboundRequestFixturePersistedMatchesPlan(record.persisted, plan) ||
		record.engineFixture.accountID == "" ||
		record.engineFixture.folderID == "" ||
		record.engineFixture.messageID != "qa-message-"+nonce ||
		record.engineFixture.version != "qa-version-"+nonce {
		c.mu.Unlock()
		return inboundRequestFixtureSnapshot{}, errInboundRequestFixtureLifecycle
	}
	record.state = inboundRequestFixtureStatePublishing
	c.fixtures[nonce] = record
	c.mu.Unlock()

	transition := func(state inboundRequestFixtureState) (inboundRequestFixtureSnapshot, bool) {
		c.mu.Lock()
		defer c.mu.Unlock()

		current, found := c.fixtures[nonce]
		if !found || current != record || current.state != inboundRequestFixtureStatePublishing {
			return inboundRequestFixtureSnapshot{}, false
		}
		current.state = state
		c.fixtures[nonce] = current
		if state != inboundRequestFixtureStatePublished {
			return inboundRequestFixtureSnapshot{}, true
		}
		return inboundRequestFixtureSnapshot{
			nonce:           current.nonce,
			scenario:        ScenarioInboundRequestV1,
			state:           current.state,
			ownerEmail:      current.plan.ownerAddress,
			organizationID:  current.persisted.organizationID,
			ownerUserID:     current.persisted.ownerUserID,
			mailAccountID:   current.persisted.mailAccountID,
			mailFolderID:    current.persisted.mailFolderID,
			mailThreadID:    current.persisted.mailThreadID,
			mailMessageID:   current.persisted.mailMessageID,
			envelopeEventID: current.persisted.envelopeEventID,
		}, true
	}

	method := mail.MailCalendarMethodRequest
	locatorHash := append([]byte(nil), record.persisted.calendarLocatorHash[:]...)
	envelope, err := mail.NewMailReceivedEnvelopeV1(
		record.persisted.envelopeEventID,
		record.plan.fixedAt,
		record.persisted.organizationID,
		nil,
		record.plan.correlationID,
		mail.MailReceivedEventPayload{
			Mailbox:        mail.MailboxEventID{AccountID: record.persisted.mailAccountID, ID: record.persisted.mailAccountID},
			Message:        mail.MailboxEventID{AccountID: record.persisted.mailAccountID, ID: record.persisted.mailMessageID},
			EngineMessage:  mail.MailboxEventOpaqueID{AccountID: record.persisted.mailAccountID, ID: record.engineFixture.messageID},
			Thread:         mail.MailboxEventID{AccountID: record.persisted.mailAccountID, ID: record.persisted.mailThreadID},
			From:           mail.MailEventAddress{Email: record.plan.organizerAddress},
			To:             []mail.MailEventAddress{{Email: record.plan.ownerAddress}},
			ReceivedAt:     record.plan.fixedAt,
			FolderIDs:      []mail.MailboxEventID{{AccountID: record.persisted.mailAccountID, ID: record.persisted.mailFolderID}},
			HasAttachments: false,
			CalendarMethod:  &method,
			AuthenticationClass:     mail.MailAuthenticationClassTrustedInternal,
			CalendarPartLocatorHash: locatorHash,
		},
	)
	if err != nil {
		if _, restored := transition(inboundRequestFixtureStateCreated); !restored {
			return inboundRequestFixtureSnapshot{}, errInboundRequestFixtureLifecycle
		}
		return inboundRequestFixtureSnapshot{}, errInboundRequestFixtureLifecycle
	}

	tx, err := c.client.Tx(ctx)
	if err != nil {
		if _, restored := transition(inboundRequestFixtureStateCreated); !restored {
			return inboundRequestFixtureSnapshot{}, errInboundRequestFixtureLifecycle
		}
		return inboundRequestFixtureSnapshot{}, errInboundRequestFixtureLifecycle
	}
	finished := false
	defer func() {
		if !finished {
			_ = tx.Rollback()
		}
	}()

	if err := outbox.NewPublisher().PublishInTx(ctx, tx, envelope.Subject(), envelope); err != nil {
		if tx.Rollback() != nil {
			return inboundRequestFixtureSnapshot{}, errInboundRequestFixtureLifecycle
		}
		finished = true
		if _, restored := transition(inboundRequestFixtureStateCreated); !restored {
			return inboundRequestFixtureSnapshot{}, errInboundRequestFixtureLifecycle
		}
		return inboundRequestFixtureSnapshot{}, errInboundRequestFixtureLifecycle
	}

	finished = true
	if err := tx.Commit(); err != nil {
		if _, marked := transition(inboundRequestFixtureStatePublishUnknown); !marked {
			return inboundRequestFixtureSnapshot{}, errInboundRequestFixtureLifecycle
		}
		return inboundRequestFixtureSnapshot{}, errInboundRequestFixtureLifecycle
	}

	snapshot, published := transition(inboundRequestFixtureStatePublished)
	if !published {
		return inboundRequestFixtureSnapshot{}, errInboundRequestFixtureLifecycle
	}
	return snapshot, nil
}
