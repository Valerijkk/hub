package mail

const (
	// MaxEngineOpaqueIDBytes ограничивает непрозрачный идентификатор движка.
	MaxEngineOpaqueIDBytes = 1024
	// MaxEngineCursorBytes ограничивает непрозрачный курсор или checkpoint движка.
	MaxEngineCursorBytes = 4096
	// MaxEnginePageItems ограничивает одну страницу движка.
	MaxEnginePageItems = 100
	// MaxEngineFolderItems ограничивает список папок одной учётной записи.
	MaxEngineFolderItems = 100
	// MaxEngineAddressItems ограничивает каждый набор адресов содержимого.
	MaxEngineAddressItems = 200
	// MaxEngineKeywordItems ограничивает набор зарезервированных markers.
	MaxEngineKeywordItems = 200
	// MaxEngineSpamTrainingItems ограничивает feedback-команду classifier.
	MaxEngineSpamTrainingItems = 200
	// MaxEngineDNSRecordItems ограничивает публичные DNS-требования домена.
	MaxEngineDNSRecordItems = 32
)

// EngineCapabilities — подтверждённые безопасные возможности движка.
type EngineCapabilities struct {
	ExclusiveCustomFolderWriter bool
	StableProjectionEnumeration bool
	SubmissionReconciliation    bool
	DraftReconciliation         bool
}

// EngineHealth — безопасное состояние доступности движка.
type EngineHealth struct {
	Available bool
	State     EngineHealthState
}

// EngineHealthState — безопасное состояние работоспособности движка.
type EngineHealthState string

const (
	// EngineHealthUnknown обозначает неизвестное состояние движка.
	EngineHealthUnknown EngineHealthState = "unknown"
	// EngineHealthReady обозначает готовность движка к обслуживанию.
	EngineHealthReady EngineHealthState = "ready"
	// EngineHealthDegraded обозначает деградированное состояние движка.
	EngineHealthDegraded EngineHealthState = "degraded"
	// EngineHealthUnavailable обозначает недоступность движка.
	EngineHealthUnavailable EngineHealthState = "unavailable"
)

// DomainSpec — неизменяемое желаемое состояние домена для provisioning.
type DomainSpec struct {
	ProvisioningKey   string
	Domain            string
	DesiredGeneration int64
	DesiredHash       string
	Present           bool
	Enabled           bool
}

// EngineDomainState — безопасное наблюдаемое состояние домена движка.
type EngineDomainState struct {
	DomainID     string
	State        EngineDomainStateKind
	Applied      bool
	DNSRecords   []EngineDNSRecord
	Verification EngineVerificationState
	DKIM         EngineVerificationState
}

// EngineDomainStateKind — типизированное состояние домена движка.
type EngineDomainStateKind string

const (
	// EngineDomainAbsent обозначает подтверждённое отсутствие домена.
	EngineDomainAbsent EngineDomainStateKind = "absent"
	// EngineDomainDNSRequired обозначает необходимость публичной DNS-настройки.
	EngineDomainDNSRequired EngineDomainStateKind = "dns_required"
	// EngineDomainVerified обозначает подтверждённый домен.
	EngineDomainVerified EngineDomainStateKind = "verified"
	// EngineDomainIndeterminate обозначает неопределённый результат операции с доменом.
	EngineDomainIndeterminate EngineDomainStateKind = "indeterminate"
	// EngineDomainRejected обозначает отклонённое состояние домена.
	EngineDomainRejected EngineDomainStateKind = "rejected"
)

// EngineVerificationState — безопасное состояние проверки домена или DKIM.
type EngineVerificationState string

const (
	// EngineVerificationUnknown обозначает неизвестное состояние проверки.
	EngineVerificationUnknown EngineVerificationState = "unknown"
	// EngineVerificationPending обозначает ожидающую проверку.
	EngineVerificationPending EngineVerificationState = "pending"
	// EngineVerificationVerified обозначает успешную проверку.
	EngineVerificationVerified EngineVerificationState = "verified"
	// EngineVerificationFailed обозначает неуспешную проверку.
	EngineVerificationFailed EngineVerificationState = "failed"
)

// EngineDNSRecord — ограниченная публичная DNS-запись, требуемая доменом.
type EngineDNSRecord struct {
	Name  string
	Type  EngineDNSRecordType
	Value string
	TTL   int
}

// EngineDNSRecordType — тип публичной DNS-записи.
type EngineDNSRecordType string

const (
	// EngineDNSRecordTXT обозначает TXT-запись.
	EngineDNSRecordTXT EngineDNSRecordType = "TXT"
	// EngineDNSRecordMX обозначает MX-запись.
	EngineDNSRecordMX EngineDNSRecordType = "MX"
	// EngineDNSRecordCNAME обозначает CNAME-запись.
	EngineDNSRecordCNAME EngineDNSRecordType = "CNAME"
)

// AccountSpec — неизменяемое желаемое состояние учётной записи для provisioning.
type AccountSpec struct {
	ProvisioningKey   string
	LocalPart         string
	Domain            string
	DesiredGeneration int64
	DesiredHash       string
	Enabled           bool
	Primary           AccountIdentitySpec
	Aliases           []AccountIdentitySpec
}

// AccountIdentitySpec — желаемая primary или alias идентичность без секретов.
type AccountIdentitySpec struct {
	LocalPart string
	Domain    string
	Enabled   bool
}

// EngineAccount — безопасное подтверждённое состояние учётной записи движка.
type EngineAccount struct {
	AccountID         string
	State             EngineAccountState
	AppliedGeneration int64
	AppliedHash       string
	PrimaryIdentityID string
	AliasIdentityIDs  []string
}

// EngineAccountState — состояние provisioned учётной записи движка.
type EngineAccountState string

const (
	// EngineAccountProvisioning обозначает ожидающую provisioning учётную запись.
	EngineAccountProvisioning EngineAccountState = "provisioning"
	// EngineAccountActive обозначает активную учётную запись.
	EngineAccountActive EngineAccountState = "active"
	// EngineAccountSuspended обозначает отключённую учётную запись.
	EngineAccountSuspended EngineAccountState = "suspended"
	// EngineAccountIndeterminate обозначает неопределённый provisioning-результат.
	EngineAccountIndeterminate EngineAccountState = "indeterminate"
)

// EngineFolder — ограниченная папка движка внутри одной учётной записи.
type EngineFolder struct {
	AccountID string
	FolderID  string
	ParentID  string
	Name      string
	Role      EngineFolderRole
}

// EngineFolderRole — системная роль папки движка.
type EngineFolderRole string

const (
	// EngineFolderRoleCustom обозначает пользовательскую папку.
	EngineFolderRoleCustom EngineFolderRole = "custom"
	// EngineFolderRoleInbox обозначает входящие.
	EngineFolderRoleInbox EngineFolderRole = "inbox"
	// EngineFolderRoleSent обозначает отправленные.
	EngineFolderRoleSent EngineFolderRole = "sent"
	// EngineFolderRoleDrafts обозначает черновики.
	EngineFolderRoleDrafts EngineFolderRole = "drafts"
	// EngineFolderRoleArchive обозначает архив.
	EngineFolderRoleArchive EngineFolderRole = "archive"
	// EngineFolderRoleTrash обозначает корзину.
	EngineFolderRoleTrash EngineFolderRole = "trash"
	// EngineFolderRoleJunk обозначает спам.
	EngineFolderRoleJunk EngineFolderRole = "junk"
)

// FolderSpec — желаемые свойства создаваемой папки.
type FolderSpec struct {
	Name     string
	ParentID string
}

// EnginePageRequest — запрошенная ограниченная страница данных движка.
type EnginePageRequest struct {
	Cursor string
	Limit  int
}

// EnginePageBounds — явные ограничения и фактический размер страницы движка.
type EnginePageBounds struct {
	Requested int
	Maximum   int
	Returned  int
}

// EngineMessageQuery — account-scoped запрос ограниченной страницы сообщений.
type EngineMessageQuery struct {
	AccountID string
	FolderID  string
	Page      EnginePageRequest
}

// EngineMessagePage — account-scoped ограниченная страница гидратированных сообщений.
type EngineMessagePage struct {
	AccountID  string
	Cursor     string
	NextCursor string
	Bounds     EnginePageBounds
	Messages   []EngineMessage
}

// EngineAddress — ограниченный адрес сообщения или черновика.
type EngineAddress struct {
	Email string
	Name  string
}

// EngineMessageContent — каноническое содержимое, передаваемое только движку.
type EngineMessageContent struct {
	From      EngineAddress
	To        []EngineAddress
	CC        []EngineAddress
	BCC       []EngineAddress
	Subject   string
	PlainBody string
	RichBody  string
}

// EngineProjectionOrigin — серверно подтверждённое происхождение проекции сообщения.
type EngineProjectionOrigin string

const (
	// EngineProjectionOriginInboundOrUnknown обозначает inbound либо неизвестное происхождение.
	EngineProjectionOriginInboundOrUnknown EngineProjectionOrigin = "inbound_or_unknown"
	// EngineProjectionOriginConfirmedOwnerSubmission обозначает подтверждённую отправку владельца.
	EngineProjectionOriginConfirmedOwnerSubmission EngineProjectionOrigin = "confirmed_owner_submission"
	// EngineProjectionOriginConfirmedOwnerDraft обозначает подтверждённый черновик владельца.
	EngineProjectionOriginConfirmedOwnerDraft EngineProjectionOrigin = "confirmed_owner_draft"
)

// EngineKeywordMarker — непрозрачный зарезервированный marker ключевого слова движка.
type EngineKeywordMarker string

// EngineMessage — гидратированное account-scoped сообщение движка.
type EngineMessage struct {
	AccountID        string
	MessageID        string
	Version          string
	Content          EngineMessageContent
	From             EngineAddress
	FolderIDs        []string
	Flags            EngineMessageFlags
	ReservedKeywords []EngineKeywordMarker
	Origin           EngineProjectionOrigin
}

// EngineMessageFlags — серверно подтверждённые флаги гидратированного сообщения.
type EngineMessageFlags struct {
	Seen     bool
	Answered bool
	Draft    bool
	Flagged  bool
	Junk     bool
}

// EngineBlob — account-scoped поток байтов движка с ограниченными метаданными.
type EngineBlob struct {
	AccountID string
	BlobID    string
	MediaType string
	SizeBytes int64
	Bytes     []byte
}

// EngineChangePage — ограниченная инкрементальная страница движка.
type EngineChangePage struct {
	AccountID        string
	FromCursor       string
	NextCursor       string
	Bounds           EnginePageBounds
	CreatedOrUpdated []EngineMessage
	DestroyedIDs     []string
}

// EngineEnumerationPage — ограниченная возобновляемая базовая страница проекции.
type EngineEnumerationPage struct {
	AccountID      string
	BaseCursor     string
	Checkpoint     string
	NextCheckpoint string
	Complete       bool
	Bounds         EnginePageBounds
	Messages       []EngineMessage
}

// EngineDraft — request-only полный черновик с прежней версией и стабильной мутацией Hub.
type EngineDraft struct {
	AccountID       string
	MessageID       string
	PreviousVersion string
	MutationID      string
	Message         EngineMessageContent
}

// EngineDraftReconciliation — account-scoped запрос сверки неопределённого черновика.
type EngineDraftReconciliation struct {
	AccountID       string
	MessageID       string
	PreviousVersion string
	MutationID      string
}

// EngineDraftResult — закрытый результат сверки черновика.
type EngineDraftResult interface {
	engineDraftResult()
}

// EngineDraftApplied — подтверждённый результат сохранения или сверки черновика.
type EngineDraftApplied struct {
	Message EngineMessage
}

func (EngineDraftApplied) engineDraftResult() {}

// EngineDraftAbsent — подтверждённое отсутствие или отсутствие изменений черновика.
type EngineDraftAbsent struct{}

func (EngineDraftAbsent) engineDraftResult() {}

// EngineIndeterminateCode — безопасная причина неопределённого результата.
type EngineIndeterminateCode string

const (
	// EngineIndeterminateTimeout обозначает потерянный результат по тайм-ауту.
	EngineIndeterminateTimeout EngineIndeterminateCode = "timeout"
	// EngineIndeterminateUnavailable обозначает потерянный результат из-за недоступности.
	EngineIndeterminateUnavailable EngineIndeterminateCode = "unavailable"
)

// EngineDraftIndeterminate — неопределённый результат сверки черновика.
type EngineDraftIndeterminate struct {
	Code EngineIndeterminateCode
}

func (EngineDraftIndeterminate) engineDraftResult() {}

// EngineSubmission — request-only отправка с неизменяемым ключом движка.
type EngineSubmission struct {
	AccountID     string
	DraftID       string
	MessageID     string
	SubmissionKey string
	Message       EngineMessageContent
}

// EngineSubmissionResult — закрытый результат отправки или её сверки.
type EngineSubmissionResult interface {
	engineSubmissionResult()
}

// EngineSubmissionSent — подтверждённая отправка с известными идентификаторами результата.
type EngineSubmissionSent struct {
	SubmissionID string
	MessageID    string
}

func (EngineSubmissionSent) engineSubmissionResult() {}

// EngineSubmissionAbsent — подтверждённое отсутствие отправки.
type EngineSubmissionAbsent struct{}

func (EngineSubmissionAbsent) engineSubmissionResult() {}

// EngineSubmissionIndeterminate — неопределённый результат отправки или сверки.
type EngineSubmissionIndeterminate struct {
	Code EngineIndeterminateCode
}

func (EngineSubmissionIndeterminate) engineSubmissionResult() {}

// FlagValue — трёхсостояние флага: не менять либо явно установить значение.
type FlagValue struct {
	Present bool
	Value   bool
}

// FlagPatch — account-scoped частичное изменение серверных флагов.
type FlagPatch struct {
	Seen     FlagValue
	Answered FlagValue
	Flagged  FlagValue
	Junk     FlagValue
}

// KeywordPatch — изменение только серверно разрешённых зарезервированных markers.
type KeywordPatch struct {
	Add    []EngineKeywordMarker
	Remove []EngineKeywordMarker
}

// SieveScript — фиксированный Hub-скрипт с маркером поколения и хеша.
type SieveScript struct {
	Name       string
	Generation int64
	Hash       string
}

// EngineSieveState — безопасное состояние фиксированного Hub Sieve-скрипта.
type EngineSieveState struct {
	Name       string
	Active     bool
	Generation int64
	Hash       string
	State      EngineSieveStateKind
}

// EngineSieveStateKind — состояние применения фиксированного Sieve-скрипта.
type EngineSieveStateKind string

const (
	// EngineSieveAbsent обозначает отсутствие фиксированного скрипта.
	EngineSieveAbsent EngineSieveStateKind = "absent"
	// EngineSieveApplied обозначает подтверждённое применение скрипта.
	EngineSieveApplied EngineSieveStateKind = "applied"
	// EngineSieveIndeterminate обозначает неопределённый результат применения.
	EngineSieveIndeterminate EngineSieveStateKind = "indeterminate"
)

// EngineSpamTraining — account-scoped явная команда classifier feedback.
type EngineSpamTraining struct {
	AccountID string
	ReceiptID string
	CommandID string
	Kind      EngineSpamTrainingKind
	Samples   []EngineSpamTrainingSample
}

// EngineSpamTrainingKind — явный вид feedback для классификатора.
type EngineSpamTrainingKind string

const (
	// EngineSpamTrainingSpam обозначает feedback spam.
	EngineSpamTrainingSpam EngineSpamTrainingKind = "spam"
	// EngineSpamTrainingHam обозначает feedback ham.
	EngineSpamTrainingHam EngineSpamTrainingKind = "ham"
)

// EngineSpamTrainingSample — один уникальный account-scoped образец classifier feedback.
type EngineSpamTrainingSample struct {
	MessageID string
}

// EngineSpamTrainingResult — ограниченный результат classifier feedback по сообщениям.
type EngineSpamTrainingResult struct {
	Outcomes []EngineSpamTrainingOutcome
}

// EngineSpamTrainingOutcome — результат classifier feedback одного сообщения.
type EngineSpamTrainingOutcome struct {
	MessageID string
	Status    EngineSpamTrainingStatus
	Code      EngineIndeterminateCode
}

// EngineSpamTrainingStatus — явный результат classifier feedback.
type EngineSpamTrainingStatus string

const (
	// EngineSpamTrainingAccepted обозначает принятое classifier feedback.
	EngineSpamTrainingAccepted EngineSpamTrainingStatus = "accepted"
	// EngineSpamTrainingRejected обозначает отклонённый classifier feedback.
	EngineSpamTrainingRejected EngineSpamTrainingStatus = "rejected"
	// EngineSpamTrainingIndeterminate обозначает неопределённый classifier feedback.
	EngineSpamTrainingIndeterminate EngineSpamTrainingStatus = "indeterminate"
)
