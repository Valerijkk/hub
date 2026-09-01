//go:build qa && !release

package qamailcalendar

import (
	"context"
	"errors"
	"sort"

	"entgo.io/ent/dialect/sql"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/aiprovider"
	"github.com/valerii/hub/backend/ent/calendarattendee"
	"github.com/valerii/hub/backend/ent/calendarevent"
	"github.com/valerii/hub/backend/ent/calendarmutationreceipt"
	"github.com/valerii/hub/backend/ent/crmcontact"
	"github.com/valerii/hub/backend/ent/hubrecord"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/leaverequest"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/mailattachment"
	"github.com/valerii/hub/backend/ent/mailboxpolicy"
	"github.com/valerii/hub/backend/ent/maildomain"
	"github.com/valerii/hub/backend/ent/maildraft"
	"github.com/valerii/hub/backend/ent/mailfolder"
	"github.com/valerii/hub/backend/ent/mailidentity"
	"github.com/valerii/hub/backend/ent/maillabel"
	"github.com/valerii/hub/backend/ent/mailmessage"
	"github.com/valerii/hub/backend/ent/mailmessagefolder"
	"github.com/valerii/hub/backend/ent/mailmessagelabel"
	"github.com/valerii/hub/backend/ent/mailmutationreceipt"
	"github.com/valerii/hub/backend/ent/mailoauthcredential"
	"github.com/valerii/hub/backend/ent/mailsendratebucket"
	"github.com/valerii/hub/backend/ent/mailsubmission"
	"github.com/valerii/hub/backend/ent/mailthread"
	"github.com/valerii/hub/backend/ent/marketplacelisting"
	"github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/projecttemplate"
	"github.com/valerii/hub/backend/ent/recentcontact"
	"github.com/valerii/hub/backend/ent/testcaseversion"
	"github.com/valerii/hub/backend/ent/testresult"
	"github.com/valerii/hub/backend/ent/user"
)

// errThreadSearchFixtureDeleteGuardDependent is private to the QA fixture
// deletion flow. HTTP owns its redacted response and header mapping.
var errThreadSearchFixtureDeleteGuardDependent = errors.New("qa mail thread search fixture delete guard dependent")

// threadSearchFixtureGuardDeleteDependents proves the complete fixed direct-FK
// inventory before a fixture root is deleted. Every query is typed, scoped to
// immutable fixture IDs, and locks its observed rows with NOWAIT.
func threadSearchFixtureGuardDeleteDependents(
	ctx context.Context,
	tx *ent.Tx,
	record threadSearchFixtureRecord,
) error {
	if ctx == nil ||
		tx == nil ||
		record.Nonce != record.plan.nonce ||
		record.Scenario != ThreadSearchFixtureLifecycleScenarioThreadSearchV1 ||
		!threadSearchFixtureLifecycleRecordMatches(record, ThreadSearchFixtureLifecycleStateDeleting) {
		return errThreadSearchFixtureLifecycle
	}

	threadIDs, ok := threadSearchFixtureGuardSortedIDs(record.persisted.threadIDs[:])
	if !ok {
		return errThreadSearchFixtureLifecycle
	}
	messageIDs, ok := threadSearchFixtureGuardSortedIDs(record.persisted.messageIDs[:])
	if !ok {
		return errThreadSearchFixtureLifecycle
	}
	messageFolderIDs, ok := threadSearchFixtureGuardSortedIDs(record.persisted.messageFolderIDs[:])
	if !ok {
		return errThreadSearchFixtureLifecycle
	}

	organizationRow, err := tx.Organization.Query().
		Where(
			organization.IDEQ(record.persisted.organizationID),
			organization.HasUsersWith(user.IDEQ(record.persisted.ownerUserID)),
			threadSearchFixtureGuardNoWait,
		).
		Only(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	owner, err := tx.User.Query().
		Where(user.IDEQ(record.persisted.ownerUserID), threadSearchFixtureGuardNoWait).
		Only(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	domain, err := tx.MailDomain.Query().
		Where(
			maildomain.IDEQ(record.persisted.mailDomainID),
			maildomain.OrganizationIDEQ(record.persisted.organizationID),
			threadSearchFixtureGuardNoWait,
		).
		Only(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	account, err := tx.MailAccount.Query().
		Where(
			mailaccount.IDEQ(record.persisted.mailAccountID),
			mailaccount.UserIDEQ(record.persisted.ownerUserID),
			mailaccount.MailDomainIDEQ(record.persisted.mailDomainID),
			mailaccount.OrganizationIDEQ(record.persisted.organizationID),
			threadSearchFixtureGuardNoWait,
		).
		Only(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	folder, err := tx.MailFolder.Query().
		Where(
			mailfolder.IDEQ(record.persisted.mailFolderID),
			mailfolder.MailAccountIDEQ(record.persisted.mailAccountID),
			threadSearchFixtureGuardNoWait,
		).
		Only(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}

	threads := make([]*ent.MailThread, 0, threadSearchFixtureRowCount)
	for _, id := range threadIDs {
		value, queryErr := tx.MailThread.Query().
			Where(
				mailthread.IDEQ(id),
				mailthread.MailAccountIDEQ(record.persisted.mailAccountID),
				threadSearchFixtureGuardNoWait,
			).
			Only(ctx)
		if queryErr != nil {
			return errThreadSearchFixtureLifecycle
		}
		threads = append(threads, value)
	}
	messages := make([]*ent.MailMessage, 0, threadSearchFixtureRowCount)
	for _, id := range messageIDs {
		value, queryErr := tx.MailMessage.Query().
			Where(
				mailmessage.IDEQ(id),
				mailmessage.MailAccountIDEQ(record.persisted.mailAccountID),
				threadSearchFixtureGuardNoWait,
			).
			Only(ctx)
		if queryErr != nil {
			return errThreadSearchFixtureLifecycle
		}
		messages = append(messages, value)
	}
	messageFolders := make([]*ent.MailMessageFolder, 0, threadSearchFixtureRowCount)
	for _, id := range messageFolderIDs {
		value, queryErr := tx.MailMessageFolder.Query().
			Where(mailmessagefolder.IDEQ(id), threadSearchFixtureGuardNoWait).
			Only(ctx)
		if queryErr != nil {
			return errThreadSearchFixtureLifecycle
		}
		messageFolders = append(messageFolders, value)
	}

	observation := ThreadSearchDeleteGuardObservation{}
	attachments, err := tx.MailAttachment.Query().
		Where(mailattachment.MailMessageIDIn(messageIDs...), threadSearchFixtureGuardNoWait).
		Limit(1).
		All(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	observation.States[ThreadSearchDeleteGuardSlotMailAttachmentByMessage] = threadSearchFixtureGuardEmptyState(len(attachments))

	draftsReply, err := tx.MailDraft.Query().
		Where(maildraft.ReplyToMessageIDIn(messageIDs...), threadSearchFixtureGuardNoWait).
		Limit(1).
		All(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	observation.States[ThreadSearchDeleteGuardSlotMailDraftReplyToMessage] = threadSearchFixtureGuardEmptyState(len(draftsReply))

	draftsForward, err := tx.MailDraft.Query().
		Where(maildraft.ForwardMessageIDIn(messageIDs...), threadSearchFixtureGuardNoWait).
		Limit(1).
		All(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	observation.States[ThreadSearchDeleteGuardSlotMailDraftForwardMessage] = threadSearchFixtureGuardEmptyState(len(draftsForward))

	draftsPendingReply, err := tx.MailDraft.Query().
		Where(maildraft.PendingReplyToMessageIDIn(messageIDs...), threadSearchFixtureGuardNoWait).
		Limit(1).
		All(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	observation.States[ThreadSearchDeleteGuardSlotMailDraftPendingReplyToMessage] = threadSearchFixtureGuardEmptyState(len(draftsPendingReply))

	draftsPendingForward, err := tx.MailDraft.Query().
		Where(maildraft.PendingForwardMessageIDIn(messageIDs...), threadSearchFixtureGuardNoWait).
		Limit(1).
		All(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	observation.States[ThreadSearchDeleteGuardSlotMailDraftPendingForwardMessage] = threadSearchFixtureGuardEmptyState(len(draftsPendingForward))

	messageFoldersByMessage, err := tx.MailMessageFolder.Query().
		Where(mailmessagefolder.MailMessageIDIn(messageIDs...), threadSearchFixtureGuardNoWait).
		All(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	messageFoldersByMessageIDs, ok := threadSearchFixtureGuardMailMessageFolderIDs(messageFoldersByMessage)
	if !ok {
		return errThreadSearchFixtureLifecycle
	}
	observation.States[ThreadSearchDeleteGuardSlotMailMessageFolderByMessage] = threadSearchFixtureGuardExactIDState(messageFolderIDs, messageFoldersByMessageIDs)

	messageLabels, err := tx.MailMessageLabel.Query().
		Where(mailmessagelabel.MailMessageIDIn(messageIDs...), threadSearchFixtureGuardNoWait).
		Limit(1).
		All(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	observation.States[ThreadSearchDeleteGuardSlotMailMessageLabelByMessage] = threadSearchFixtureGuardEmptyState(len(messageLabels))

	messagesByThread, err := tx.MailMessage.Query().
		Where(mailmessage.MailThreadIDIn(threadIDs...), threadSearchFixtureGuardNoWait).
		All(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	messagesByThreadIDs, ok := threadSearchFixtureGuardMailMessageIDs(messagesByThread)
	if !ok {
		return errThreadSearchFixtureLifecycle
	}
	observation.States[ThreadSearchDeleteGuardSlotMailMessageByThread] = threadSearchFixtureGuardExactIDState(messageIDs, messagesByThreadIDs)

	childFolders, err := tx.MailFolder.Query().
		Where(mailfolder.ParentIDEQ(record.persisted.mailFolderID), threadSearchFixtureGuardNoWait).
		Limit(1).
		All(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	observation.States[ThreadSearchDeleteGuardSlotMailFolderByParent] = threadSearchFixtureGuardEmptyState(len(childFolders))

	messageFoldersByFolder, err := tx.MailMessageFolder.Query().
		Where(mailmessagefolder.MailFolderIDEQ(record.persisted.mailFolderID), threadSearchFixtureGuardNoWait).
		All(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	messageFoldersByFolderIDs, ok := threadSearchFixtureGuardMailMessageFolderIDs(messageFoldersByFolder)
	if !ok {
		return errThreadSearchFixtureLifecycle
	}
	observation.States[ThreadSearchDeleteGuardSlotMailMessageFolderByFolder] = threadSearchFixtureGuardExactIDState(messageFolderIDs, messageFoldersByFolderIDs)

	draftsByAccount, err := tx.MailDraft.Query().
		Where(maildraft.MailAccountIDEQ(record.persisted.mailAccountID), threadSearchFixtureGuardNoWait).
		Limit(1).
		All(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	observation.States[ThreadSearchDeleteGuardSlotMailDraftByAccount] = threadSearchFixtureGuardEmptyState(len(draftsByAccount))

	identities, err := tx.MailIdentity.Query().
		Where(mailidentity.MailAccountIDEQ(record.persisted.mailAccountID), threadSearchFixtureGuardNoWait).
		All(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	identityIDs, ok := threadSearchFixtureGuardMailIdentityIDs(identities)
	if !ok {
		return errThreadSearchFixtureLifecycle
	}
	observation.States[ThreadSearchDeleteGuardSlotMailIdentityByAccount] = threadSearchFixtureGuardExactIDState(
		[]uuid.UUID{record.persisted.mailIdentityID},
		identityIDs,
	)
	identity := threadSearchFixtureGuardFindMailIdentity(identities, record.persisted.mailIdentityID)

	labelsByAccount, err := tx.MailLabel.Query().
		Where(maillabel.MailAccountIDEQ(record.persisted.mailAccountID), threadSearchFixtureGuardNoWait).
		Limit(1).
		All(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	observation.States[ThreadSearchDeleteGuardSlotMailLabelByAccount] = threadSearchFixtureGuardEmptyState(len(labelsByAccount))

	oauthCredentials, err := tx.MailOAuthCredential.Query().
		Where(mailoauthcredential.MailAccountIDEQ(record.persisted.mailAccountID), threadSearchFixtureGuardNoWait).
		Limit(1).
		All(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	observation.States[ThreadSearchDeleteGuardSlotMailOAuthCredentialByAccount] = threadSearchFixtureGuardEmptyState(len(oauthCredentials))

	submissions, err := tx.MailSubmission.Query().
		Where(mailsubmission.MailAccountIDEQ(record.persisted.mailAccountID), threadSearchFixtureGuardNoWait).
		Limit(1).
		All(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	observation.States[ThreadSearchDeleteGuardSlotMailSubmissionByAccount] = threadSearchFixtureGuardEmptyState(len(submissions))

	recentContactsByAccount, err := tx.RecentContact.Query().
		Where(recentcontact.MailAccountIDEQ(record.persisted.mailAccountID), threadSearchFixtureGuardNoWait).
		Limit(1).
		All(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	observation.States[ThreadSearchDeleteGuardSlotRecentContactByAccount] = threadSearchFixtureGuardEmptyState(len(recentContactsByAccount))

	foldersByAccount, err := tx.MailFolder.Query().
		Where(mailfolder.MailAccountIDEQ(record.persisted.mailAccountID), threadSearchFixtureGuardNoWait).
		All(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	foldersByAccountIDs, ok := threadSearchFixtureGuardMailFolderIDs(foldersByAccount)
	if !ok {
		return errThreadSearchFixtureLifecycle
	}
	observation.States[ThreadSearchDeleteGuardSlotMailFolderByAccount] = threadSearchFixtureGuardExactIDState(
		[]uuid.UUID{record.persisted.mailFolderID},
		foldersByAccountIDs,
	)

	threadsByAccount, err := tx.MailThread.Query().
		Where(mailthread.MailAccountIDEQ(record.persisted.mailAccountID), threadSearchFixtureGuardNoWait).
		All(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	threadsByAccountIDs, ok := threadSearchFixtureGuardMailThreadIDs(threadsByAccount)
	if !ok {
		return errThreadSearchFixtureLifecycle
	}
	observation.States[ThreadSearchDeleteGuardSlotMailThreadByAccount] = threadSearchFixtureGuardExactIDState(threadIDs, threadsByAccountIDs)

	messagesByAccount, err := tx.MailMessage.Query().
		Where(mailmessage.MailAccountIDEQ(record.persisted.mailAccountID), threadSearchFixtureGuardNoWait).
		All(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	messagesByAccountIDs, ok := threadSearchFixtureGuardMailMessageIDs(messagesByAccount)
	if !ok {
		return errThreadSearchFixtureLifecycle
	}
	observation.States[ThreadSearchDeleteGuardSlotMailMessageByAccount] = threadSearchFixtureGuardExactIDState(messageIDs, messagesByAccountIDs)

	mutationReceipts, err := tx.MailMutationReceipt.Query().
		Where(mailmutationreceipt.MailAccountIDEQ(record.persisted.mailAccountID), threadSearchFixtureGuardNoWait).
		Limit(1).
		All(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	observation.States[ThreadSearchDeleteGuardSlotMailMutationReceiptByAccount] = threadSearchFixtureGuardEmptyState(len(mutationReceipts))

	accountsByDomain, err := tx.MailAccount.Query().
		Where(mailaccount.MailDomainIDEQ(record.persisted.mailDomainID), threadSearchFixtureGuardNoWait).
		All(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	accountsByDomainIDs, ok := threadSearchFixtureGuardMailAccountIDs(accountsByDomain)
	if !ok {
		return errThreadSearchFixtureLifecycle
	}
	accountByDomainState := threadSearchFixtureGuardExactIDState(
		[]uuid.UUID{record.persisted.mailAccountID},
		accountsByDomainIDs,
	)
	if accountByDomainState == ThreadSearchDeleteGuardSlotStateDependent {
		accountByDomainState = ThreadSearchDeleteGuardSlotStateGraphMismatch
	}
	observation.States[ThreadSearchDeleteGuardSlotMailAccountByDomain] = accountByDomainState

	aiProviders, err := tx.AIProvider.Query().
		Where(aiprovider.HasOwnerWith(user.IDEQ(record.persisted.ownerUserID)), threadSearchFixtureGuardNoWait).
		Limit(1).
		All(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	observation.States[ThreadSearchDeleteGuardSlotAIProviderByOwner] = threadSearchFixtureGuardEmptyState(len(aiProviders))

	crmContacts, err := tx.CrmContact.Query().
		Where(crmcontact.HasLinkedUserWith(user.IDEQ(record.persisted.ownerUserID)), threadSearchFixtureGuardNoWait).
		Limit(1).
		All(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	observation.States[ThreadSearchDeleteGuardSlotCRMContactByLinkedUser] = threadSearchFixtureGuardEmptyState(len(crmContacts))

	hubRecords, err := tx.HubRecord.Query().
		Where(hubrecord.HasCreatedByWith(user.IDEQ(record.persisted.ownerUserID)), threadSearchFixtureGuardNoWait).
		Limit(1).
		All(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	observation.States[ThreadSearchDeleteGuardSlotHubRecordByCreatedBy] = threadSearchFixtureGuardEmptyState(len(hubRecords))

	issuesByAssignee, err := tx.Issue.Query().
		Where(issue.HasAssigneeWith(user.IDEQ(record.persisted.ownerUserID)), threadSearchFixtureGuardNoWait).
		Limit(1).
		All(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	observation.States[ThreadSearchDeleteGuardSlotIssueByAssignee] = threadSearchFixtureGuardEmptyState(len(issuesByAssignee))

	issuesByReporter, err := tx.Issue.Query().
		Where(issue.HasReporterWith(user.IDEQ(record.persisted.ownerUserID)), threadSearchFixtureGuardNoWait).
		Limit(1).
		All(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	observation.States[ThreadSearchDeleteGuardSlotIssueByReporter] = threadSearchFixtureGuardEmptyState(len(issuesByReporter))

	leaveRequests, err := tx.LeaveRequest.Query().
		Where(leaverequest.HasReviewerWith(user.IDEQ(record.persisted.ownerUserID)), threadSearchFixtureGuardNoWait).
		Limit(1).
		All(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	observation.States[ThreadSearchDeleteGuardSlotLeaveRequestByReviewer] = threadSearchFixtureGuardEmptyState(len(leaveRequests))

	marketplaceListings, err := tx.MarketplaceListing.Query().
		Where(marketplacelisting.HasReviewerWith(user.IDEQ(record.persisted.ownerUserID)), threadSearchFixtureGuardNoWait).
		Limit(1).
		All(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	observation.States[ThreadSearchDeleteGuardSlotMarketplaceListingByReviewer] = threadSearchFixtureGuardEmptyState(len(marketplaceListings))

	projectTemplates, err := tx.ProjectTemplate.Query().
		Where(projecttemplate.HasCreatedByWith(user.IDEQ(record.persisted.ownerUserID)), threadSearchFixtureGuardNoWait).
		Limit(1).
		All(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	observation.States[ThreadSearchDeleteGuardSlotProjectTemplateByCreatedBy] = threadSearchFixtureGuardEmptyState(len(projectTemplates))

	testCaseVersions, err := tx.TestCaseVersion.Query().
		Where(testcaseversion.HasCreatedByWith(user.IDEQ(record.persisted.ownerUserID)), threadSearchFixtureGuardNoWait).
		Limit(1).
		All(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	observation.States[ThreadSearchDeleteGuardSlotTestCaseVersionByCreatedBy] = threadSearchFixtureGuardEmptyState(len(testCaseVersions))

	testResults, err := tx.TestResult.Query().
		Where(testresult.HasExecutedByWith(user.IDEQ(record.persisted.ownerUserID)), threadSearchFixtureGuardNoWait).
		Limit(1).
		All(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	observation.States[ThreadSearchDeleteGuardSlotTestResultByExecutedBy] = threadSearchFixtureGuardEmptyState(len(testResults))

	calendarAttendees, err := tx.CalendarAttendee.Query().
		Where(calendarattendee.UserIDEQ(record.persisted.ownerUserID), threadSearchFixtureGuardNoWait).
		Limit(1).
		All(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	observation.States[ThreadSearchDeleteGuardSlotCalendarAttendeeByUser] = threadSearchFixtureGuardEmptyState(len(calendarAttendees))

	calendarEvents, err := tx.CalendarEvent.Query().
		Where(calendarevent.OrganizerIDEQ(record.persisted.ownerUserID), threadSearchFixtureGuardNoWait).
		Limit(1).
		All(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	observation.States[ThreadSearchDeleteGuardSlotCalendarEventByOrganizer] = threadSearchFixtureGuardEmptyState(len(calendarEvents))

	rateBucketsByUser, err := tx.MailSendRateBucket.Query().
		Where(mailsendratebucket.UserIDEQ(record.persisted.ownerUserID), threadSearchFixtureGuardNoWait).
		Limit(1).
		All(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	observation.States[ThreadSearchDeleteGuardSlotMailSendRateBucketByUser] = threadSearchFixtureGuardEmptyState(len(rateBucketsByUser))

	recentContactsByUser, err := tx.RecentContact.Query().
		Where(recentcontact.UserIDEQ(record.persisted.ownerUserID), threadSearchFixtureGuardNoWait).
		Limit(1).
		All(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	observation.States[ThreadSearchDeleteGuardSlotRecentContactByUser] = threadSearchFixtureGuardEmptyState(len(recentContactsByUser))

	calendarMutationReceipts, err := tx.CalendarMutationReceipt.Query().
		Where(calendarmutationreceipt.ActorUserIDEQ(record.persisted.ownerUserID), threadSearchFixtureGuardNoWait).
		Limit(1).
		All(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	observation.States[ThreadSearchDeleteGuardSlotCalendarMutationReceiptByActor] = threadSearchFixtureGuardEmptyState(len(calendarMutationReceipts))

	rateBucketsByOrganization, err := tx.MailSendRateBucket.Query().
		Where(mailsendratebucket.OrganizationIDEQ(record.persisted.organizationID), threadSearchFixtureGuardNoWait).
		Limit(1).
		All(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	observation.States[ThreadSearchDeleteGuardSlotMailSendRateBucketByOrganization] = threadSearchFixtureGuardEmptyState(len(rateBucketsByOrganization))

	domainsByOrganization, err := tx.MailDomain.Query().
		Where(maildomain.OrganizationIDEQ(record.persisted.organizationID), threadSearchFixtureGuardNoWait).
		All(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	domainsByOrganizationIDs, ok := threadSearchFixtureGuardMailDomainIDs(domainsByOrganization)
	if !ok {
		return errThreadSearchFixtureLifecycle
	}
	observation.States[ThreadSearchDeleteGuardSlotMailDomainByOrganization] = threadSearchFixtureGuardExactIDState(
		[]uuid.UUID{record.persisted.mailDomainID},
		domainsByOrganizationIDs,
	)

	mailboxPolicies, err := tx.MailboxPolicy.Query().
		Where(mailboxpolicy.OrganizationIDEQ(record.persisted.organizationID), threadSearchFixtureGuardNoWait).
		Limit(1).
		All(ctx)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}
	observation.States[ThreadSearchDeleteGuardSlotMailboxPolicyByOrganization] = threadSearchFixtureGuardEmptyState(len(mailboxPolicies))

	if !threadSearchFixtureDeleteGraphValid(
		record,
		organizationRow,
		owner,
		domain,
		account,
		identity,
		folder,
		threads,
		messages,
		messageFolders,
	) {
		return errThreadSearchFixtureLifecycle
	}

	switch DecideThreadSearchDeleteGuard(observation) {
	case ThreadSearchDeleteGuardDecisionAllow:
		return nil
	case ThreadSearchDeleteGuardDecisionDependent:
		return errThreadSearchFixtureDeleteGuardDependent
	default:
		return errThreadSearchFixtureLifecycle
	}
}

func threadSearchFixtureGuardNoWait(selector *sql.Selector) {
	selector.ForUpdate(sql.WithLockAction(sql.NoWait))
}

func threadSearchFixtureGuardEmptyState(count int) ThreadSearchDeleteGuardSlotState {
	if count == 0 {
		return ThreadSearchDeleteGuardSlotStateClean
	}
	return ThreadSearchDeleteGuardSlotStateDependent
}

func threadSearchFixtureGuardSortedIDs(ids []uuid.UUID) ([]uuid.UUID, bool) {
	result := append([]uuid.UUID(nil), ids...)
	sort.Slice(result, func(left, right int) bool {
		return result[left].String() < result[right].String()
	})
	for index, id := range result {
		if id == uuid.Nil || (index > 0 && id == result[index-1]) {
			return nil, false
		}
	}
	return result, true
}

func threadSearchFixtureGuardExactIDState(expected, actual []uuid.UUID) ThreadSearchDeleteGuardSlotState {
	if len(expected) == 0 {
		return ThreadSearchDeleteGuardSlotStateGraphMismatch
	}
	expectedSet := make(map[uuid.UUID]struct{}, len(expected))
	for _, id := range expected {
		if id == uuid.Nil {
			return ThreadSearchDeleteGuardSlotStateGraphMismatch
		}
		if _, exists := expectedSet[id]; exists {
			return ThreadSearchDeleteGuardSlotStateGraphMismatch
		}
		expectedSet[id] = struct{}{}
	}
	seen := make(map[uuid.UUID]struct{}, len(actual))
	extra := false
	for _, id := range actual {
		if id == uuid.Nil {
			return ThreadSearchDeleteGuardSlotStateGraphMismatch
		}
		if _, exists := seen[id]; exists {
			return ThreadSearchDeleteGuardSlotStateGraphMismatch
		}
		seen[id] = struct{}{}
		if _, expected := expectedSet[id]; !expected {
			extra = true
		}
	}
	for id := range expectedSet {
		if _, exists := seen[id]; !exists {
			return ThreadSearchDeleteGuardSlotStateGraphMismatch
		}
	}
	if extra {
		return ThreadSearchDeleteGuardSlotStateDependent
	}
	return ThreadSearchDeleteGuardSlotStateClean
}

func threadSearchFixtureGuardMailAccountIDs(values []*ent.MailAccount) ([]uuid.UUID, bool) {
	ids := make([]uuid.UUID, len(values))
	for index, value := range values {
		if value == nil {
			return nil, false
		}
		ids[index] = value.ID
	}
	return ids, true
}

func threadSearchFixtureGuardMailDomainIDs(values []*ent.MailDomain) ([]uuid.UUID, bool) {
	ids := make([]uuid.UUID, len(values))
	for index, value := range values {
		if value == nil {
			return nil, false
		}
		ids[index] = value.ID
	}
	return ids, true
}

func threadSearchFixtureGuardMailFolderIDs(values []*ent.MailFolder) ([]uuid.UUID, bool) {
	ids := make([]uuid.UUID, len(values))
	for index, value := range values {
		if value == nil {
			return nil, false
		}
		ids[index] = value.ID
	}
	return ids, true
}

func threadSearchFixtureGuardMailIdentityIDs(values []*ent.MailIdentity) ([]uuid.UUID, bool) {
	ids := make([]uuid.UUID, len(values))
	for index, value := range values {
		if value == nil {
			return nil, false
		}
		ids[index] = value.ID
	}
	return ids, true
}

func threadSearchFixtureGuardMailMessageIDs(values []*ent.MailMessage) ([]uuid.UUID, bool) {
	ids := make([]uuid.UUID, len(values))
	for index, value := range values {
		if value == nil {
			return nil, false
		}
		ids[index] = value.ID
	}
	return ids, true
}

func threadSearchFixtureGuardMailMessageFolderIDs(values []*ent.MailMessageFolder) ([]uuid.UUID, bool) {
	ids := make([]uuid.UUID, len(values))
	for index, value := range values {
		if value == nil {
			return nil, false
		}
		ids[index] = value.ID
	}
	return ids, true
}

func threadSearchFixtureGuardMailThreadIDs(values []*ent.MailThread) ([]uuid.UUID, bool) {
	ids := make([]uuid.UUID, len(values))
	for index, value := range values {
		if value == nil {
			return nil, false
		}
		ids[index] = value.ID
	}
	return ids, true
}

func threadSearchFixtureGuardFindMailIdentity(values []*ent.MailIdentity, id uuid.UUID) *ent.MailIdentity {
	for _, value := range values {
		if value != nil && value.ID == id {
			return value
		}
	}
	return nil
}
