//go:build qa && !release

package qamailcalendar

import (
	"context"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailidentity"
)

func (c *Controller) deleteInboundRequestFixture(ctx context.Context, nonce string) error {
	if c == nil ||
		ctx == nil ||
		c.client == nil ||
		c.engine == nil ||
		c.random == nil ||
		c.fixtures == nil ||
		!validToken(string(c.config.token)) ||
		!ValidNonce(nonce) {
		return errInboundRequestFixtureControllerInput
	}

	plan, err := newInboundRequestFixturePlan(nonce)
	if err != nil {
		return errInboundRequestFixtureLifecycle
	}

	c.mu.Lock()
	record, found := c.fixtures[nonce]
	if !found ||
		record.nonce != nonce ||
		record.state != inboundRequestFixtureStateCreated ||
		record.plan != plan ||
		!inboundRequestFixturePersistedMatchesPlan(record.persisted, plan) ||
		record.engineFixture.accountID == "" ||
		record.engineFixture.primaryIdentityID == "" ||
		record.engineFixture.folderID == "" ||
		record.engineFixture.messageID != "qa-message-"+nonce ||
		record.engineFixture.version != "qa-version-"+nonce {
		c.mu.Unlock()
		return errInboundRequestFixtureLifecycle
	}
	record.state = inboundRequestFixtureStateDeleting
	c.fixtures[nonce] = record
	c.mu.Unlock()

	transition := func(state inboundRequestFixtureState) bool {
		c.mu.Lock()
		defer c.mu.Unlock()

		current, found := c.fixtures[nonce]
		if !found || current != record || current.state != inboundRequestFixtureStateDeleting {
			return false
		}
		current.state = state
		c.fixtures[nonce] = current
		return true
	}

	tx, err := c.client.Tx(ctx)
	if err != nil {
		if !transition(inboundRequestFixtureStateCreated) {
			return errInboundRequestFixtureLifecycle
		}
		return errInboundRequestFixtureLifecycle
	}
	finished := false
	defer func() {
		if !finished {
			_ = tx.Rollback()
		}
	}()

	identities, err := tx.MailIdentity.Query().
		Where(mailidentity.MailAccountIDEQ(record.persisted.mailAccountID)).
		Order(ent.Asc(mailidentity.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil || len(identities) != 1 {
		_ = tx.Rollback()
		finished = true
		if !transition(inboundRequestFixtureStateCreated) {
			return errInboundRequestFixtureLifecycle
		}
		return errInboundRequestFixtureLifecycle
	}
	identity := identities[0]
	if identity == nil ||
		identity.ID != record.persisted.mailIdentityID ||
		identity.MailAccountID != record.persisted.mailAccountID ||
		identity.NormalizedAddress != plan.ownerAddress ||
		identity.Kind != mailidentity.KindPrimary ||
		identity.Status != mailidentity.StatusActive ||
		!identity.SendAllowed ||
		identity.EngineIdentityID == nil ||
		*identity.EngineIdentityID != record.engineFixture.primaryIdentityID ||
		!identity.CreatedAt.Equal(plan.fixedAt) ||
		!identity.UpdatedAt.Equal(plan.fixedAt) {
		_ = tx.Rollback()
		finished = true
		if !transition(inboundRequestFixtureStateCreated) {
			return errInboundRequestFixtureLifecycle
		}
		return errInboundRequestFixtureLifecycle
	}

	messageFolderID := inboundRequestFixtureUUID("mail-message-folder", nonce)
	if err := tx.MailMessageFolder.DeleteOneID(messageFolderID).Exec(ctx); err != nil {
		_ = tx.Rollback()
		finished = true
		if !transition(inboundRequestFixtureStateCreated) {
			return errInboundRequestFixtureLifecycle
		}
		return errInboundRequestFixtureLifecycle
	}
	if err := tx.MailMessage.DeleteOneID(record.persisted.mailMessageID).Exec(ctx); err != nil {
		_ = tx.Rollback()
		finished = true
		if !transition(inboundRequestFixtureStateCreated) {
			return errInboundRequestFixtureLifecycle
		}
		return errInboundRequestFixtureLifecycle
	}
	if err := tx.MailThread.DeleteOneID(record.persisted.mailThreadID).Exec(ctx); err != nil {
		_ = tx.Rollback()
		finished = true
		if !transition(inboundRequestFixtureStateCreated) {
			return errInboundRequestFixtureLifecycle
		}
		return errInboundRequestFixtureLifecycle
	}
	if err := tx.MailFolder.DeleteOneID(record.persisted.mailFolderID).Exec(ctx); err != nil {
		_ = tx.Rollback()
		finished = true
		if !transition(inboundRequestFixtureStateCreated) {
			return errInboundRequestFixtureLifecycle
		}
		return errInboundRequestFixtureLifecycle
	}
	if err := tx.MailIdentity.DeleteOneID(record.persisted.mailIdentityID).Exec(ctx); err != nil {
		_ = tx.Rollback()
		finished = true
		if !transition(inboundRequestFixtureStateCreated) {
			return errInboundRequestFixtureLifecycle
		}
		return errInboundRequestFixtureLifecycle
	}
	if err := tx.MailAccount.DeleteOneID(record.persisted.mailAccountID).Exec(ctx); err != nil {
		_ = tx.Rollback()
		finished = true
		if !transition(inboundRequestFixtureStateCreated) {
			return errInboundRequestFixtureLifecycle
		}
		return errInboundRequestFixtureLifecycle
	}
	if err := tx.MailDomain.DeleteOneID(plan.mailDomainID).Exec(ctx); err != nil {
		_ = tx.Rollback()
		finished = true
		if !transition(inboundRequestFixtureStateCreated) {
			return errInboundRequestFixtureLifecycle
		}
		return errInboundRequestFixtureLifecycle
	}
	if err := tx.User.DeleteOneID(record.persisted.ownerUserID).Exec(ctx); err != nil {
		_ = tx.Rollback()
		finished = true
		if !transition(inboundRequestFixtureStateCreated) {
			return errInboundRequestFixtureLifecycle
		}
		return errInboundRequestFixtureLifecycle
	}
	if err := tx.Organization.DeleteOneID(record.persisted.organizationID).Exec(ctx); err != nil {
		_ = tx.Rollback()
		finished = true
		if !transition(inboundRequestFixtureStateCreated) {
			return errInboundRequestFixtureLifecycle
		}
		return errInboundRequestFixtureLifecycle
	}

	finished = true
	if err := tx.Commit(); err != nil {
		if !transition(inboundRequestFixtureStateCleanupFailed) {
			return errInboundRequestFixtureLifecycle
		}
		return errInboundRequestFixtureLifecycle
	}

	if err := destroyInboundRequestFixtureEngine(c.engine, record.engineFixture); err != nil {
		if !transition(inboundRequestFixtureStateCleanupFailed) {
			return errInboundRequestFixtureLifecycle
		}
		return errInboundRequestFixtureLifecycle
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	current, found := c.fixtures[nonce]
	if !found || current != record || current.state != inboundRequestFixtureStateDeleting {
		return errInboundRequestFixtureLifecycle
	}
	delete(c.fixtures, nonce)
	return nil
}
