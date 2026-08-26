//go:build qa && !release

package qamailcalendar

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/mailfolder"
	"github.com/valerii/hub/backend/ent/mailmessage"
	"github.com/valerii/hub/backend/ent/schema"
	"github.com/valerii/hub/backend/ent/user"
)

var errInboundRequestFixturePersistInput = errors.New("qa mail calendar fixture persistence input is invalid")

type inboundRequestFixturePersisted struct {
	organizationID       uuid.UUID
	ownerUserID          uuid.UUID
	mailAccountID        uuid.UUID
	mailFolderID         uuid.UUID
	mailThreadID         uuid.UUID
	mailMessageID        uuid.UUID
	envelopeEventID      uuid.UUID
	calendarLocatorHash  [sha256.Size]byte
}

func persistInboundRequestFixture(
	ctx context.Context,
	client *ent.Client,
	plan inboundRequestFixturePlan,
	engineFixture inboundRequestEngineFixture,
) (inboundRequestFixturePersisted, error) {
	if ctx == nil || client == nil {
		return inboundRequestFixturePersisted{}, errInboundRequestFixturePersistInput
	}
	canonicalPlan, err := newInboundRequestFixturePlan(plan.nonce)
	if err != nil || plan != canonicalPlan {
		return inboundRequestFixturePersisted{}, errInvalidFixtureNonce
	}
	if engineFixture.accountID == "" ||
		engineFixture.folderID == "" ||
		engineFixture.messageID != "qa-message-"+plan.nonce ||
		engineFixture.version != "qa-version-"+plan.nonce {
		return inboundRequestFixturePersisted{}, errInboundRequestFixturePersistInput
	}

	tx, err := client.Tx(ctx)
	if err != nil {
		return inboundRequestFixturePersisted{}, fmt.Errorf("begin inbound request fixture transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	if _, err := tx.Organization.Create().
		SetID(plan.organizationID).
		SetName("QA Mail Calendar " + plan.nonce).
		SetSlug(plan.organizationSlug).
		SetInviteToken(plan.correlationID).
		SetCreatedAt(plan.fixedAt).
		SetUpdatedAt(plan.fixedAt).
		Save(ctx); err != nil {
		return inboundRequestFixturePersisted{}, fmt.Errorf("create inbound request fixture organization: %w", err)
	}

	if _, err := tx.User.Create().
		SetID(plan.ownerUserID).
		SetEmail(plan.ownerAddress).
		SetName("QA Owner").
		SetDisplayName("QA Owner").
		SetSystemRole(user.SystemRoleMember).
		SetStatus(user.StatusActive).
		SetOrganizationID(plan.organizationID).
		SetTimezone("UTC").
		SetLanguage("ru").
		SetOnboardingCompleted(true).
		SetCreatedAt(plan.fixedAt).
		SetUpdatedAt(plan.fixedAt).
		Save(ctx); err != nil {
		return inboundRequestFixturePersisted{}, fmt.Errorf("create inbound request fixture owner: %w", err)
	}

	if _, err := tx.MailDomain.Create().
		SetID(plan.mailDomainID).
		SetOrganizationID(plan.organizationID).
		SetDomain(plan.domain).
		SetEngineProvisioningKey(plan.domainProvisioningKey).
		SetDesiredHash(append([]byte(nil), plan.domainControlHash[:]...)).
		SetCreatedAt(plan.fixedAt).
		SetUpdatedAt(plan.fixedAt).
		Save(ctx); err != nil {
		return inboundRequestFixturePersisted{}, fmt.Errorf("create inbound request fixture domain: %w", err)
	}

	if _, err := tx.MailAccount.Create().
		SetID(plan.mailAccountID).
		SetOrganizationID(plan.organizationID).
		SetUserID(plan.ownerUserID).
		SetMailDomainID(plan.mailDomainID).
		SetEngineProvisioningKey(plan.accountProvisioningKey).
		SetLocalPart("owner").
		SetEngineAccountID(engineFixture.accountID).
		SetDesiredEnabled(true).
		SetStatus(mailaccount.Status("active")).
		SetControlHash(append([]byte(nil), plan.accountControlHash[:]...)).
		SetDeployedGeneration(1).
		SetDeployedHashVersion(1).
		SetDeployedHash(append([]byte(nil), plan.accountControlHash[:]...)).
		SetProjectionState(mailaccount.ProjectionState("ready")).
		SetSyncCursor(plan.correlationID).
		SetLastSyncedAt(plan.fixedAt).
		SetSyncScheduleUpdatedAt(plan.fixedAt).
		SetEventCutoffAt(plan.fixedAt).
		SetCreatedAt(plan.fixedAt).
		SetUpdatedAt(plan.fixedAt).
		Save(ctx); err != nil {
		return inboundRequestFixturePersisted{}, fmt.Errorf("create inbound request fixture account: %w", err)
	}

	if _, err := tx.MailFolder.Create().
		SetID(plan.mailFolderID).
		SetMailAccountID(plan.mailAccountID).
		SetEngineFolderID(engineFixture.folderID).
		SetName(plan.folderName).
		SetNormalizedName(plan.normalizedFolderName).
		SetControlState(mailfolder.ControlStateActive).
		SetCreatedAt(plan.fixedAt).
		SetUpdatedAt(plan.fixedAt).
		Save(ctx); err != nil {
		return inboundRequestFixturePersisted{}, fmt.Errorf("create inbound request fixture folder: %w", err)
	}

	participants := schema.MailThreadParticipants{plan.organizerAddress, plan.ownerAddress}
	if _, err := tx.MailThread.Create().
		SetID(plan.mailThreadID).
		SetMailAccountID(plan.mailAccountID).
		SetThreadKeyVersion(1).
		SetThreadKey(append([]byte(nil), plan.threadKey[:]...)).
		SetNormalizedSubject(plan.normalizedSubject).
		SetParticipantAddresses(participants).
		SetLastMessageAt(plan.fixedAt).
		SetMessageCount(1).
		SetUnreadCount(1).
		SetIsPinned(false).
		SetHasAttachments(false).
		SetSnippet(plan.snippet).
		SetLastSeenSyncEpoch(0).
		SetCreatedAt(plan.fixedAt).
		SetUpdatedAt(plan.fixedAt).
		Save(ctx); err != nil {
		return inboundRequestFixturePersisted{}, fmt.Errorf("create inbound request fixture thread: %w", err)
	}

	if _, err := tx.MailMessage.Create().
		SetID(plan.mailMessageID).
		SetMailAccountID(plan.mailAccountID).
		SetMailThreadID(plan.mailThreadID).
		SetEngineMessageID(engineFixture.messageID).
		SetEngineVersion(engineFixture.version).
		SetInternetMessageID("<" + plan.itipUID + "@example.test>").
		SetReferences(schema.MailMessageReferences{}).
		SetFrom(schema.MailStoredAddress{Email: plan.organizerAddress}).
		SetTo(schema.MailStoredAddresses{{Email: plan.ownerAddress}}).
		SetCc(schema.MailStoredAddresses{}).
		SetBcc(schema.MailStoredAddresses{}).
		SetReplyTo(schema.MailStoredAddresses{}).
		SetSubject(plan.subject).
		SetSnippet(plan.snippet).
		SetReceivedAt(plan.fixedAt).
		SetSizeBytes(int64(len(plan.snippet))).
		SetPriority(mailmessage.Priority("normal")).
		SetHasAttachments(false).
		SetIsDraft(false).
		SetIsSeen(false).
		SetIsAnswered(false).
		SetIsFlagged(false).
		SetIsJunk(false).
		SetProjectionOrigin(mailmessage.ProjectionOrigin("inbound_or_unknown")).
		SetFlagsSyncedAt(plan.fixedAt).
		SetCalendarMethod(mailmessage.CalendarMethod("request")).
		SetCalendarPartLocatorHash(append([]byte(nil), plan.calendarLocatorHash[:]...)).
		SetAuthenticationClass(mailmessage.AuthenticationClass("trusted_internal")).
		SetLastSeenSyncEpoch(0).
		SetReceivedEventID(plan.envelopeEventID).
		SetCreatedAt(plan.fixedAt).
		SetUpdatedAt(plan.fixedAt).
		Save(ctx); err != nil {
		return inboundRequestFixturePersisted{}, fmt.Errorf("create inbound request fixture message: %w", err)
	}

	messageFolderID := inboundRequestFixtureUUID("mail-message-folder", plan.nonce)
	if _, err := tx.MailMessageFolder.Create().
		SetID(messageFolderID).
		SetMailMessageID(plan.mailMessageID).
		SetMailFolderID(plan.mailFolderID).
		SetLastSeenSyncEpoch(0).
		SetProjectedAt(plan.fixedAt).
		SetCreatedAt(plan.fixedAt).
		SetUpdatedAt(plan.fixedAt).
		Save(ctx); err != nil {
		return inboundRequestFixturePersisted{}, fmt.Errorf("create inbound request fixture message folder: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return inboundRequestFixturePersisted{}, fmt.Errorf("commit inbound request fixture transaction: %w", err)
	}
	committed = true

	return inboundRequestFixturePersisted{
		organizationID:      plan.organizationID,
		ownerUserID:         plan.ownerUserID,
		mailAccountID:       plan.mailAccountID,
		mailFolderID:        plan.mailFolderID,
		mailThreadID:        plan.mailThreadID,
		mailMessageID:       plan.mailMessageID,
		envelopeEventID:     plan.envelopeEventID,
		calendarLocatorHash: plan.calendarLocatorHash,
	}, nil
}
