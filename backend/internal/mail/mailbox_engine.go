package mail

import "context"

// MailEngine — независимый от провайдера порт движка почтового ящика.
type MailEngine interface {
	Capabilities(context.Context) (EngineCapabilities, error)
	GetDomain(context.Context, string) (EngineDomainState, error)
	EnsureDomain(context.Context, DomainSpec) (EngineDomainState, error)
	EnsureAccount(context.Context, AccountSpec) (EngineAccount, error)
	ListFolders(context.Context, string) ([]EngineFolder, error)
	CreateFolder(context.Context, string, FolderSpec) (EngineFolder, error)
	RenameFolder(context.Context, string, string, string) error
	DeleteFolder(context.Context, string, string) error
	ListMessages(context.Context, EngineMessageQuery) (EngineProjectionPage, error)
	GetMessage(context.Context, EngineMessageReadRequest) (EngineMessage, error)
	GetDraft(context.Context, EngineDraftReadRequest) (EngineMessage, error)
	GetBlob(context.Context, EngineBlobReadRequest) (EngineBlob, error)
	ReadCalendarPart(context.Context, EngineCalendarPartReadRequest) (EngineCalendarPartReadResult, error)
	Changes(context.Context, EngineChangeRequest) (EngineChangePage, error)
	EnumerateProjection(context.Context, EngineEnumerationRequest) (EngineEnumerationPage, error)
	SaveDraft(context.Context, EngineDraft) (EngineMessage, error)
	ReconcileDraft(context.Context, EngineDraftReconciliation) (EngineDraftResult, error)
	DeleteDraft(context.Context, EngineDraftDeletion) error
	ReconcileDraftDeletion(context.Context, EngineDraftDeletion) (EngineDraftDeletionResult, error)
	Submit(context.Context, EngineSubmission) (EngineSubmissionResult, error)
	ReconcileSubmission(context.Context, EngineSubmission) (EngineSubmissionResult, error)
	MoveMessages(context.Context, string, []string, string) error
	SetFlags(context.Context, string, []string, FlagPatch) error
	SetKeywords(context.Context, string, []string, KeywordPatch) error
	DeleteMessages(context.Context, string, []string, bool) error
	GetSieve(context.Context, string, string) (EngineSieveState, error)
	UpsertSieve(context.Context, string, SieveScript) (EngineSieveState, error)
	ClassifyMessages(context.Context, EngineSpamTraining) (EngineSpamTrainingResult, error)
	Health(context.Context) (EngineHealth, error)
}
