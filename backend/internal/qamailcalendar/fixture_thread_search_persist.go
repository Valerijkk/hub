//go:build qa && !release

package qamailcalendar

import (
	"bytes"
	"context"
	"errors"
	"log/slog"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/maildomain"
	"github.com/valerii/hub/backend/ent/mailfolder"
	"github.com/valerii/hub/backend/ent/mailidentity"
	"github.com/valerii/hub/backend/ent/mailmessage"
	"github.com/valerii/hub/backend/ent/mailmessagefolder"
	"github.com/valerii/hub/backend/ent/mailthread"
	organizationpredicate "github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/schema"
	"github.com/valerii/hub/backend/ent/user"
)

var errThreadSearchFixturePersist = errors.New("qa mail thread search fixture persistence failed")

// threadSearchFixtureCommitOutcomeUnknownError marks the sole mutation whose
// durable result cannot be established because transaction commit failed.
// Its public text intentionally omits the database error; callers in this
// package may inspect the wrapped cause when they own reconciliation.
type threadSearchFixtureCommitOutcomeUnknownError struct {
	cause error
}

func (err *threadSearchFixtureCommitOutcomeUnknownError) Error() string {
	return "qa mail thread search fixture commit outcome unknown"
}

func (err *threadSearchFixtureCommitOutcomeUnknownError) Unwrap() error {
	return err.cause
}

func persistThreadSearchFixture(
	ctx context.Context,
	client *ent.Client,
	plan threadSearchFixturePlan,
) (threadSearchFixturePersisted, error) {
	if ctx == nil || client == nil || !threadSearchFixturePlanMatchesCanonical(plan) {
		return threadSearchFixturePersisted{}, errThreadSearchFixturePersist
	}

	persisted := threadSearchFixturePersistedFromPlan(plan)
	tx, err := client.Tx(ctx)
	if err != nil || tx == nil {
		return threadSearchFixturePersisted{}, errThreadSearchFixturePersist
	}
	commitAttempted := false
	defer func() {
		if !commitAttempted {
			_ = tx.Rollback()
		}
	}()

	organization, err := tx.Organization.Create().
		SetID(plan.organizationID).
		SetName("QA Thread Search " + plan.nonce).
		SetSlug(plan.organizationSlug).
		SetInviteToken(plan.correlationID).
		SetCreatedAt(plan.fixedAt).
		SetUpdatedAt(plan.fixedAt).
		Save(ctx)
	if err != nil || !threadSearchFixtureOrganizationMatchesPlan(organization, plan) {
		return threadSearchFixturePersisted{}, errThreadSearchFixturePersist
	}

	owner, err := tx.User.Create().
		SetID(plan.ownerUserID).
		SetEmail(plan.ownerAddress).
		SetName("QA Thread Search Owner").
		SetDisplayName("QA Thread Search Owner").
		SetSystemRole(user.SystemRoleMember).
		SetStatus(user.StatusActive).
		SetOrganizationID(plan.organizationID).
		SetTimezone("UTC").
		SetLanguage("ru").
		SetOnboardingCompleted(true).
		SetCreatedAt(plan.fixedAt).
		SetUpdatedAt(plan.fixedAt).
		Save(ctx)
	if err != nil || !threadSearchFixtureOwnerMatchesPlan(owner, plan) {
		return threadSearchFixturePersisted{}, errThreadSearchFixturePersist
	}

	domain, err := tx.MailDomain.Create().
		SetID(plan.mailDomainID).
		SetOrganizationID(plan.organizationID).
		SetDomain(plan.domain).
		SetEngineProvisioningKey(plan.domainProvisioningKey).
		SetEngineDomainID(plan.engineDomainID).
		SetDesiredPresent(true).
		SetEnabled(true).
		SetIsDefault(true).
		SetState(maildomain.State("verified")).
		SetVerificationStatus(maildomain.VerificationStatus("verified")).
		SetDkimStatus(maildomain.DkimStatus("valid")).
		SetDesiredGeneration(1).
		SetDesiredHashVersion(1).
		SetDesiredHash(append([]byte(nil), plan.domainControlHash[:]...)).
		SetDeployedGeneration(1).
		SetDeployedHashVersion(1).
		SetDeployedHash(append([]byte(nil), plan.domainControlHash[:]...)).
		SetCreatedAt(plan.fixedAt).
		SetUpdatedAt(plan.fixedAt).
		Save(ctx)
	if err != nil || !threadSearchFixtureDomainMatchesPlan(domain, plan) {
		return threadSearchFixturePersisted{}, errThreadSearchFixturePersist
	}

	account, err := tx.MailAccount.Create().
		SetID(plan.mailAccountID).
		SetOrganizationID(plan.organizationID).
		SetUserID(plan.ownerUserID).
		SetMailDomainID(plan.mailDomainID).
		SetEngineProvisioningKey(plan.accountProvisioningKey).
		SetLocalPart("owner").
		SetEngineAccountID(plan.engineAccountID).
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
		Save(ctx)
	if err != nil || !threadSearchFixtureAccountMatchesPlan(account, plan) {
		return threadSearchFixturePersisted{}, errThreadSearchFixturePersist
	}

	identity, err := tx.MailIdentity.Create().
		SetID(plan.mailIdentityID).
		SetMailAccountID(plan.mailAccountID).
		SetNormalizedAddress(plan.ownerAddress).
		SetKind(mailidentity.KindPrimary).
		SetStatus(mailidentity.StatusActive).
		SetSendAllowed(true).
		SetEngineIdentityID(plan.engineIdentityID).
		SetCreatedAt(plan.fixedAt).
		SetUpdatedAt(plan.fixedAt).
		Save(ctx)
	if err != nil || !threadSearchFixtureMailIdentityMatchesPlan(identity, plan) {
		return threadSearchFixturePersisted{}, errThreadSearchFixturePersist
	}

	folder, err := tx.MailFolder.Create().
		SetID(plan.mailFolderID).
		SetMailAccountID(plan.mailAccountID).
		SetEngineFolderID(plan.engineFolderID).
		SetName(plan.folderName).
		SetNormalizedName(plan.normalizedFolderName).
		SetControlState(mailfolder.ControlStateActive).
		SetCreatedAt(plan.fixedAt).
		SetUpdatedAt(plan.fixedAt).
		Save(ctx)
	if err != nil || !threadSearchFixtureFolderMatchesPlan(folder, plan) {
		return threadSearchFixturePersisted{}, errThreadSearchFixturePersist
	}

	for index, row := range plan.rows {
		thread, err := tx.MailThread.Create().
			SetID(row.threadID).
			SetMailAccountID(plan.mailAccountID).
			SetThreadKeyVersion(1).
			SetThreadKey(append([]byte(nil), row.threadKey[:]...)).
			SetNormalizedSubject(row.normalizedSubject).
			SetParticipantAddresses(schema.MailThreadParticipants{plan.ownerAddress}).
			SetLastMessageAt(row.lastMessageAt).
			SetMessageCount(1).
			SetUnreadCount(0).
			SetIsPinned(row.isPinned).
			SetHasAttachments(false).
			SetSnippet(row.snippet).
			SetLastSeenSyncEpoch(0).
			SetCreatedAt(plan.fixedAt).
			SetUpdatedAt(plan.fixedAt).
			Save(ctx)
		if err != nil || !threadSearchFixtureThreadMatchesPlan(thread, plan, row) {
			return threadSearchFixturePersisted{}, errThreadSearchFixturePersist
		}

		message, err := tx.MailMessage.Create().
			SetID(row.messageID).
			SetMailAccountID(plan.mailAccountID).
			SetMailThreadID(row.threadID).
			SetEngineMessageID(row.engineMessageID).
			SetEngineVersion(row.engineVersion).
			SetFrom(schema.MailStoredAddress{Email: plan.ownerAddress, Name: "QA Thread Search Owner"}).
			SetSubject(row.normalizedSubject).
			SetSnippet(row.snippet).
			SetReceivedAt(row.lastMessageAt).
			SetSizeBytes(int64(len(row.snippet))).
			SetPriority(mailmessage.Priority("normal")).
			SetHasAttachments(false).
			SetIsDraft(false).
			SetIsSeen(true).
			SetIsAnswered(false).
			SetIsFlagged(row.isFlagged).
			SetIsJunk(false).
			SetProjectionOrigin(mailmessage.ProjectionOrigin("inbound_or_unknown")).
			SetLastSeenSyncEpoch(0).
			SetCreatedAt(plan.fixedAt).
			SetUpdatedAt(plan.fixedAt).
			Save(ctx)
		if err != nil || !threadSearchFixtureMessageMatchesPlan(message, plan, row) {
			return threadSearchFixturePersisted{}, errThreadSearchFixturePersist
		}

		messageFolder, err := tx.MailMessageFolder.Create().
			SetID(row.messageFolderID).
			SetMailMessageID(row.messageID).
			SetMailFolderID(plan.mailFolderID).
			SetLastSeenSyncEpoch(0).
			SetProjectedAt(row.lastMessageAt).
			SetCreatedAt(plan.fixedAt).
			SetUpdatedAt(plan.fixedAt).
			Save(ctx)
		if err != nil || !threadSearchFixtureMessageFolderMatchesPlan(messageFolder, plan, row) {
			return threadSearchFixturePersisted{}, errThreadSearchFixturePersist
		}
		if persisted.threadIDs[index] != row.threadID ||
			persisted.messageIDs[index] != row.messageID ||
			persisted.messageFolderIDs[index] != row.messageFolderID {
			return threadSearchFixturePersisted{}, errThreadSearchFixturePersist
		}
	}

	if !threadSearchFixturePersistedMatchesPlan(persisted, plan) ||
		!threadSearchFixtureVerifyPersisted(ctx, tx, plan) {
		return threadSearchFixturePersisted{}, errThreadSearchFixturePersist
	}

	commitAttempted = true
	if err := tx.Commit(); err != nil {
		return threadSearchFixturePersisted{}, &threadSearchFixtureCommitOutcomeUnknownError{cause: err}
	}
	return persisted, nil
}

func threadSearchFixturePersistedMatchesPlan(
	persisted threadSearchFixturePersisted,
	plan threadSearchFixturePlan,
) bool {
	return threadSearchFixturePlanMatchesCanonical(plan) &&
		persisted == threadSearchFixturePersistedFromPlan(plan)
}

func threadSearchFixturePersistedFromPlan(plan threadSearchFixturePlan) threadSearchFixturePersisted {
	persisted := threadSearchFixturePersisted{
		organizationID: plan.organizationID,
		ownerUserID:    plan.ownerUserID,
		mailDomainID:   plan.mailDomainID,
		mailAccountID:  plan.mailAccountID,
		mailIdentityID: plan.mailIdentityID,
		mailFolderID:   plan.mailFolderID,
	}
	for index, row := range plan.rows {
		persisted.threadIDs[index] = row.threadID
		persisted.messageIDs[index] = row.messageID
		persisted.messageFolderIDs[index] = row.messageFolderID
	}
	return persisted
}

func threadSearchFixtureVerifyPersisted(
	ctx context.Context,
	tx *ent.Tx,
	plan threadSearchFixturePlan,
) bool {
	organization, err := tx.Organization.Get(ctx, plan.organizationID)
	if err != nil || !threadSearchFixtureOrganizationMatchesPlan(organization, plan) {
		return false
	}
	owner, err := tx.User.Query().
		Where(
			user.IDEQ(plan.ownerUserID),
			user.HasOrganizationWith(organizationpredicate.IDEQ(plan.organizationID)),
		).
		Only(ctx)
	if err != nil || !threadSearchFixtureOwnerMatchesPlan(owner, plan) {
		return false
	}
	domain, err := tx.MailDomain.Get(ctx, plan.mailDomainID)
	if err != nil || !threadSearchFixtureDomainMatchesPlan(domain, plan) {
		return false
	}
	account, err := tx.MailAccount.Get(ctx, plan.mailAccountID)
	if err != nil || !threadSearchFixtureAccountMatchesPlan(account, plan) {
		return false
	}
	identities, err := tx.MailIdentity.Query().
		Where(
			mailidentity.IDEQ(plan.mailIdentityID),
			mailidentity.MailAccountIDEQ(plan.mailAccountID),
		).
		All(ctx)
	if err != nil || len(identities) != 1 || !threadSearchFixtureMailIdentityMatchesPlan(identities[0], plan) {
		return false
	}
	folder, err := tx.MailFolder.Get(ctx, plan.mailFolderID)
	if err != nil || !threadSearchFixtureFolderMatchesPlan(folder, plan) {
		return false
	}

	threadCount, err := tx.MailThread.Query().
		Where(mailthread.MailAccountIDEQ(plan.mailAccountID)).
		Count(ctx)
	if err != nil || threadCount != threadSearchFixtureRowCount {
		return false
	}
	messageCount, err := tx.MailMessage.Query().
		Where(mailmessage.MailAccountIDEQ(plan.mailAccountID)).
		Count(ctx)
	if err != nil || messageCount != threadSearchFixtureRowCount {
		return false
	}
	messageFolderCount, err := tx.MailMessageFolder.Query().
		Where(mailmessagefolder.MailFolderIDEQ(plan.mailFolderID)).
		Count(ctx)
	if err != nil || messageFolderCount != threadSearchFixtureRowCount {
		return false
	}

	for _, row := range plan.rows {
		thread, err := tx.MailThread.Get(ctx, row.threadID)
		if err != nil || !threadSearchFixtureThreadMatchesPlan(thread, plan, row) {
			return false
		}
		message, err := tx.MailMessage.Get(ctx, row.messageID)
		if err != nil {
			slog.ErrorContext(ctx, "qa mail thread search fixture message read failed", "error", err)
			return false
		}
		if fieldCode := threadSearchFixtureMessageMismatchFieldCode(message, plan, row); fieldCode != "" {
			slog.ErrorContext(
				ctx,
				"qa mail thread search fixture message mismatch",
				"message_id", row.messageID.String(),
				"field_code", fieldCode,
			)
			return false
		}
		messageFolder, err := tx.MailMessageFolder.Get(ctx, row.messageFolderID)
		if err != nil || !threadSearchFixtureMessageFolderMatchesPlan(messageFolder, plan, row) {
			return false
		}
	}
	return true
}

func threadSearchFixtureOrganizationMatchesPlan(value *ent.Organization, plan threadSearchFixturePlan) bool {
	return value != nil &&
		value.ID == plan.organizationID &&
		value.Name == "QA Thread Search "+plan.nonce &&
		value.Slug == plan.organizationSlug &&
		value.InviteToken == plan.correlationID &&
		value.CreatedAt.Equal(plan.fixedAt) &&
		value.UpdatedAt.Equal(plan.fixedAt)
}

func threadSearchFixtureOwnerMatchesPlan(value *ent.User, plan threadSearchFixturePlan) bool {
	return value != nil &&
		value.ID == plan.ownerUserID &&
		value.Email == plan.ownerAddress &&
		value.Name == "QA Thread Search Owner" &&
		value.DisplayName == "QA Thread Search Owner" &&
		value.SystemRole == user.SystemRoleMember &&
		value.Status == user.StatusActive &&
		value.Timezone == "UTC" &&
		value.Language == "ru" &&
		value.OnboardingCompleted &&
		value.CreatedAt.Equal(plan.fixedAt) &&
		value.UpdatedAt.Equal(plan.fixedAt)
}

func threadSearchFixtureDomainMatchesPlan(value *ent.MailDomain, plan threadSearchFixturePlan) bool {
	return value != nil &&
		value.ID == plan.mailDomainID &&
		value.OrganizationID == plan.organizationID &&
		value.Domain == plan.domain &&
		value.EngineProvisioningKey == plan.domainProvisioningKey &&
		value.EngineDomainID != nil && *value.EngineDomainID == plan.engineDomainID &&
		value.DesiredPresent &&
		value.Enabled &&
		value.IsDefault &&
		value.State == maildomain.State("verified") &&
		value.VerificationStatus == maildomain.VerificationStatus("verified") &&
		value.DkimStatus == maildomain.DkimStatus("valid") &&
		value.DesiredGeneration == 1 &&
		value.DesiredHashVersion == 1 &&
		bytes.Equal(value.DesiredHash, plan.domainControlHash[:]) &&
		value.DeployedGeneration != nil && *value.DeployedGeneration == 1 &&
		value.DeployedHashVersion != nil && *value.DeployedHashVersion == 1 &&
		value.DeployedHash != nil && bytes.Equal(*value.DeployedHash, plan.domainControlHash[:]) &&
		value.DeletedAt == nil &&
		value.CreatedAt.Equal(plan.fixedAt) &&
		value.UpdatedAt.Equal(plan.fixedAt)
}

func threadSearchFixtureAccountMatchesPlan(value *ent.MailAccount, plan threadSearchFixturePlan) bool {
	return value != nil &&
		value.ID == plan.mailAccountID &&
		value.OrganizationID == plan.organizationID &&
		value.UserID == plan.ownerUserID &&
		value.MailDomainID == plan.mailDomainID &&
		value.EngineProvisioningKey == plan.accountProvisioningKey &&
		value.LocalPart == "owner" &&
		value.EngineAccountID != nil && *value.EngineAccountID == plan.engineAccountID &&
		value.DesiredEnabled &&
		value.Status == mailaccount.Status("active") &&
		bytes.Equal(value.ControlHash, plan.accountControlHash[:]) &&
		value.DeployedGeneration != nil && *value.DeployedGeneration == 1 &&
		value.DeployedHashVersion != nil && *value.DeployedHashVersion == 1 &&
		value.DeployedHash != nil && bytes.Equal(*value.DeployedHash, plan.accountControlHash[:]) &&
		value.ProjectionState == mailaccount.ProjectionState("ready") &&
		value.SyncCursor != nil && *value.SyncCursor == plan.correlationID &&
		value.LastSyncedAt != nil && value.LastSyncedAt.Equal(plan.fixedAt) &&
		value.SyncScheduleUpdatedAt.Equal(plan.fixedAt) &&
		value.EventCutoffAt.Equal(plan.fixedAt) &&
		value.CreatedAt.Equal(plan.fixedAt) &&
		value.UpdatedAt.Equal(plan.fixedAt)
}

func threadSearchFixtureMailIdentityMatchesPlan(value *ent.MailIdentity, plan threadSearchFixturePlan) bool {
	return value != nil &&
		value.ID == plan.mailIdentityID &&
		value.MailAccountID == plan.mailAccountID &&
		value.NormalizedAddress == plan.ownerAddress &&
		value.DisplayName == nil &&
		value.Kind == mailidentity.KindPrimary &&
		value.SendAllowed &&
		value.Status == mailidentity.StatusActive &&
		value.EngineIdentityID != nil && *value.EngineIdentityID == plan.engineIdentityID &&
		value.CreatedAt.Equal(plan.fixedAt) &&
		value.UpdatedAt.Equal(plan.fixedAt)
}

func threadSearchFixtureFolderMatchesPlan(value *ent.MailFolder, plan threadSearchFixturePlan) bool {
	return value != nil &&
		value.ID == plan.mailFolderID &&
		value.MailAccountID == plan.mailAccountID &&
		value.EngineFolderID != nil && *value.EngineFolderID == plan.engineFolderID &&
		value.ParentID == nil &&
		value.Role == nil &&
		value.Name == plan.folderName &&
		value.NormalizedName == plan.normalizedFolderName &&
		value.ControlState == mailfolder.ControlStateActive &&
		value.DeletedAt == nil &&
		value.CreatedAt.Equal(plan.fixedAt) &&
		value.UpdatedAt.Equal(plan.fixedAt)
}

func threadSearchFixtureThreadMatchesPlan(
	value *ent.MailThread,
	plan threadSearchFixturePlan,
	row threadSearchFixtureRow,
) bool {
	return value != nil &&
		value.ID == row.threadID &&
		value.MailAccountID == plan.mailAccountID &&
		value.ThreadKeyVersion == 1 &&
		bytes.Equal(value.ThreadKey, row.threadKey[:]) &&
		value.NormalizedSubject == row.normalizedSubject &&
		len(value.ParticipantAddresses) == 1 && value.ParticipantAddresses[0] == plan.ownerAddress &&
		value.LastMessageAt.Equal(row.lastMessageAt) &&
		value.MessageCount == 1 &&
		value.UnreadCount == 0 &&
		value.IsPinned == row.isPinned &&
		!value.HasAttachments &&
		value.Snippet == row.snippet &&
		value.LastSeenSyncEpoch == 0 &&
		value.DeletedAt == nil &&
		value.CreatedAt.Equal(plan.fixedAt) &&
		value.UpdatedAt.Equal(plan.fixedAt)
}

func threadSearchFixtureMessageMatchesPlan(
	value *ent.MailMessage,
	plan threadSearchFixturePlan,
	row threadSearchFixtureRow,
) bool {
	return threadSearchFixtureMessageMismatchFieldCode(value, plan, row) == ""
}

func threadSearchFixtureMessageMismatchFieldCode(
	value *ent.MailMessage,
	plan threadSearchFixturePlan,
	row threadSearchFixtureRow,
) string {
	if value == nil {
		return "message_nil"
	}
	if value.ID != row.messageID {
		return "id"
	}
	if value.MailAccountID != plan.mailAccountID {
		return "mail_account_id"
	}
	if value.MailThreadID != row.threadID {
		return "mail_thread_id"
	}
	if value.EngineMessageID != row.engineMessageID {
		return "engine_message_id"
	}
	if value.EngineVersion != row.engineVersion {
		return "engine_version"
	}
	if value.InternetMessageID != nil {
		return "internet_message_id"
	}
	if value.InReplyTo != nil {
		return "in_reply_to"
	}
	if value.From.Email != plan.ownerAddress || value.From.Name != "QA Thread Search Owner" {
		return "from"
	}
	if len(value.To) != 0 || len(value.Cc) != 0 || len(value.Bcc) != 0 || len(value.ReplyTo) != 0 {
		return "recipient_lists"
	}
	if value.Subject != row.normalizedSubject {
		return "subject"
	}
	if value.Snippet != row.snippet {
		return "snippet"
	}
	if value.ReceivedAt == nil || !value.ReceivedAt.Equal(row.lastMessageAt) {
		return "received_at"
	}
	if value.SizeBytes != int64(len(row.snippet)) {
		return "size_bytes"
	}
	if value.Priority != mailmessage.Priority("normal") {
		return "priority"
	}
	if value.HasAttachments {
		return "has_attachments"
	}
	if value.IsDraft {
		return "is_draft"
	}
	if !value.IsSeen {
		return "is_seen"
	}
	if value.IsAnswered {
		return "is_answered"
	}
	if value.IsFlagged != row.isFlagged {
		return "is_flagged"
	}
	if value.IsJunk {
		return "is_junk"
	}
	if value.EngineState != nil {
		return "engine_state"
	}
	if value.ProjectionOrigin != mailmessage.ProjectionOrigin("inbound_or_unknown") {
		return "projection_origin"
	}
	if value.CalendarMethod != nil {
		return "calendar_method"
	}
	if value.CalendarPartLocatorHash != nil && len(*value.CalendarPartLocatorHash) != 0 {
		return "calendar_part_locator_hash"
	}
	if value.CalendarEventID != nil {
		return "calendar_event_id"
	}
	if value.LastSeenSyncEpoch != 0 {
		return "last_seen_sync_epoch"
	}
	if value.ReceivedEventID != nil {
		return "received_event_id"
	}
	if value.DeletedAt != nil {
		return "deleted_at"
	}
	if !value.CreatedAt.Equal(plan.fixedAt) {
		return "created_at"
	}
	if !value.UpdatedAt.Equal(plan.fixedAt) {
		return "updated_at"
	}
	return ""
}

func threadSearchFixtureMessageFolderMatchesPlan(
	value *ent.MailMessageFolder,
	plan threadSearchFixturePlan,
	row threadSearchFixtureRow,
) bool {
	return value != nil &&
		value.ID == row.messageFolderID &&
		value.MailMessageID == row.messageID &&
		value.MailFolderID == plan.mailFolderID &&
		value.EngineState == nil &&
		value.LastSeenSyncEpoch == 0 &&
		value.ProjectedAt.Equal(row.lastMessageAt) &&
		value.CreatedAt.Equal(plan.fixedAt) &&
		value.UpdatedAt.Equal(plan.fixedAt)
}
