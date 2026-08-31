//go:build !release

package qamailcalendar

// ThreadSearchDeleteGuardSlot identifies one exact dependent set that must be
// observed before a thread-search fixture can be deleted. The literal order is
// part of the private guard ABI; scans may not derive it dynamically.
type ThreadSearchDeleteGuardSlot uint8

const (
	ThreadSearchDeleteGuardSlotMailAttachmentByMessage ThreadSearchDeleteGuardSlot = iota
	ThreadSearchDeleteGuardSlotMailDraftReplyToMessage
	ThreadSearchDeleteGuardSlotMailDraftForwardMessage
	ThreadSearchDeleteGuardSlotMailDraftPendingReplyToMessage
	ThreadSearchDeleteGuardSlotMailDraftPendingForwardMessage
	ThreadSearchDeleteGuardSlotMailMessageFolderByMessage
	ThreadSearchDeleteGuardSlotMailMessageLabelByMessage
	ThreadSearchDeleteGuardSlotMailMessageByThread
	ThreadSearchDeleteGuardSlotMailFolderByParent
	ThreadSearchDeleteGuardSlotMailMessageFolderByFolder
	ThreadSearchDeleteGuardSlotMailDraftByAccount
	ThreadSearchDeleteGuardSlotMailIdentityByAccount
	ThreadSearchDeleteGuardSlotMailLabelByAccount
	ThreadSearchDeleteGuardSlotMailOAuthCredentialByAccount
	ThreadSearchDeleteGuardSlotMailSubmissionByAccount
	ThreadSearchDeleteGuardSlotRecentContactByAccount
	ThreadSearchDeleteGuardSlotMailFolderByAccount
	ThreadSearchDeleteGuardSlotMailThreadByAccount
	ThreadSearchDeleteGuardSlotMailMessageByAccount
	ThreadSearchDeleteGuardSlotMailMutationReceiptByAccount
	ThreadSearchDeleteGuardSlotMailAccountByDomain
	ThreadSearchDeleteGuardSlotAIProviderByOwner
	ThreadSearchDeleteGuardSlotCRMContactByLinkedUser
	ThreadSearchDeleteGuardSlotHubRecordByCreatedBy
	ThreadSearchDeleteGuardSlotIssueByAssignee
	ThreadSearchDeleteGuardSlotIssueByReporter
	ThreadSearchDeleteGuardSlotLeaveRequestByReviewer
	ThreadSearchDeleteGuardSlotMarketplaceListingByReviewer
	ThreadSearchDeleteGuardSlotProjectTemplateByCreatedBy
	ThreadSearchDeleteGuardSlotTestCaseVersionByCreatedBy
	ThreadSearchDeleteGuardSlotTestResultByExecutedBy
	ThreadSearchDeleteGuardSlotCalendarAttendeeByUser
	ThreadSearchDeleteGuardSlotCalendarEventByOrganizer
	ThreadSearchDeleteGuardSlotMailSendRateBucketByUser
	ThreadSearchDeleteGuardSlotRecentContactByUser
	ThreadSearchDeleteGuardSlotCalendarMutationReceiptByActor
	ThreadSearchDeleteGuardSlotMailSendRateBucketByOrganization
	ThreadSearchDeleteGuardSlotMailDomainByOrganization
	ThreadSearchDeleteGuardSlotMailboxPolicyByOrganization
	ThreadSearchDeleteGuardSlotCount
)

// ThreadSearchDeleteGuardSlotState is the closed result of one exact-ID,
// fixture-scoped dependent scan.
type ThreadSearchDeleteGuardSlotState uint8

const (
	ThreadSearchDeleteGuardSlotStateUnknown ThreadSearchDeleteGuardSlotState = iota
	ThreadSearchDeleteGuardSlotStateClean
	ThreadSearchDeleteGuardSlotStateDependent
	ThreadSearchDeleteGuardSlotStateGraphMismatch
)

// ThreadSearchDeleteGuardObservation is complete only when every literal slot
// is observed. A GraphMismatch represents any missing, extra, duplicate,
// foreign, or otherwise invalid exact-set result and therefore fails closed.
type ThreadSearchDeleteGuardObservation struct {
	States [ThreadSearchDeleteGuardSlotCount]ThreadSearchDeleteGuardSlotState
}

// ThreadSearchDeleteGuardDecision is deliberately private-detail-free: callers
// distinguish only clean deletion, a proven dependent, and a fail-closed
// invalid observation.
type ThreadSearchDeleteGuardDecision uint8

const (
	ThreadSearchDeleteGuardDecisionInvalid ThreadSearchDeleteGuardDecision = iota
	ThreadSearchDeleteGuardDecisionAllow
	ThreadSearchDeleteGuardDecisionDependent
)

// DecideThreadSearchDeleteGuard classifies the complete fixed-slot observation
// without performing I/O. Invalid input takes precedence over a dependent so a
// caller can expose the pre-delete sentinel only for a fully proven dependency.
func DecideThreadSearchDeleteGuard(observation ThreadSearchDeleteGuardObservation) ThreadSearchDeleteGuardDecision {
	dependent := false
	for slot := ThreadSearchDeleteGuardSlot(0); slot < ThreadSearchDeleteGuardSlotCount; slot++ {
		switch observation.States[slot] {
		case ThreadSearchDeleteGuardSlotStateClean:
		case ThreadSearchDeleteGuardSlotStateDependent:
			dependent = true
		case ThreadSearchDeleteGuardSlotStateUnknown, ThreadSearchDeleteGuardSlotStateGraphMismatch:
			return ThreadSearchDeleteGuardDecisionInvalid
		default:
			return ThreadSearchDeleteGuardDecisionInvalid
		}
	}
	if dependent {
		return ThreadSearchDeleteGuardDecisionDependent
	}
	return ThreadSearchDeleteGuardDecisionAllow
}
