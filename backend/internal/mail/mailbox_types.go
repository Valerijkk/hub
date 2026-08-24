package mail

import (
	"crypto/sha256"
	"time"
)

const (
	// MaxEngineOpaqueIDBytes ограничивает непрозрачный идентификатор движка.
	MaxEngineOpaqueIDBytes = 1024
	// MaxEngineCursorBytes ограничивает непрозрачный курсор или checkpoint движка.
	MaxEngineCursorBytes = 4096
	// MaxEnginePageItems ограничивает одну страницу движка.
	MaxEnginePageItems = 100
	// MaxEngineProjectionItemBytes ограничивает переменные байты одной проекции сообщения.
	MaxEngineProjectionItemBytes = 1048576
	// MaxEngineProjectionPageBytes ограничивает переменные байты одной страницы проекций.
	MaxEngineProjectionPageBytes = 8388608
	// MaxEngineMessageBytes ограничивает размер полного сообщения или черновика движка.
	MaxEngineMessageBytes = 52428800
	// MaxEngineAttachmentBytes ограничивает размер одного вложения движка.
	MaxEngineAttachmentBytes = 26214400
	// MaxEngineDraftRecipientItems ограничивает каждый список получателей черновика.
	MaxEngineDraftRecipientItems = 100
	// MaxEngineDraftTotalRecipientItems ограничивает суммарное число получателей черновика.
	MaxEngineDraftTotalRecipientItems = 300
	// MaxEngineReferenceItems ограничивает список ссылок Message-ID.
	MaxEngineReferenceItems = 100
	// MaxEngineFolderItems ограничивает список папок одной учётной записи.
	MaxEngineFolderItems = 100
	// MaxEngineAddressItems ограничивает каждый набор адресов содержимого.
	MaxEngineAddressItems = 200
	// MaxEngineKeywordItems ограничивает набор зарезервированных markers.
	MaxEngineKeywordItems = 200
	// MaxEngineAttachmentItems ограничивает набор вложений сообщения или черновика.
	MaxEngineAttachmentItems = 50
	// MaxEngineAddressEmailBytes ограничивает нормализованный адрес электронной почты.
	MaxEngineAddressEmailBytes = 320
	// MaxEngineAddressNameBytes ограничивает отображаемое имя адреса.
	MaxEngineAddressNameBytes = 640
	// MaxEngineHeaderValueBytes ограничивает значение одного заголовка сообщения.
	MaxEngineHeaderValueBytes = 998
	// MaxEngineSnippetBytes ограничивает безопасный текстовый сниппет.
	MaxEngineSnippetBytes = 4096
	// MaxEngineMIMEFieldBytes ограничивает поле MIME-вложения.
	MaxEngineMIMEFieldBytes = 1020
	// MaxEngineBodyPartBytes ограничивает одно нормализованное тело сообщения.
	MaxEngineBodyPartBytes = 8388608
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
	Limit int
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
	Cursor    string
	Page      EnginePageRequest
}

// EngineMessageReadRequest — account-scoped запрос полного сообщения с ограничением содержимого.
type EngineMessageReadRequest struct {
	AccountID       string
	MessageID       string
	MaxContentBytes int64
}

// EngineDraftReadRequest is one exact provider-authoritative draft read.
// AccountID is already resolved by the service and is never sourced from HTTP.
type EngineDraftReadRequest struct {
	AccountID       string
	DraftKey        string
	MessageID       string
	ExpectedVersion string
	MaxContentBytes int64
}

// EngineBlobReadRequest — account-scoped запрос двоичного вложения с ограничением размера.
type EngineBlobReadRequest struct {
	AccountID string
	BlobID    string
	MaxBytes  int64
}

// EngineCalendarPartReadRequest is one private engine read for a decoded iTIP part.
type EngineCalendarPartReadRequest struct {
	AccountID   string
	MessageID   string
	Version     string
	LocatorHash [sha256.Size]byte
	Method      MailCalendarMethod
	MaxBytes    int
}

// EngineCalendarPartReadResult is one private engine calendar-part response.
type EngineCalendarPartReadResult struct {
	LocatorHash [sha256.Size]byte
	ITIP        []byte
}

// EngineChangeRequest — account-scoped запрос инкрементальных изменений от заданного курсора.
type EngineChangeRequest struct {
	AccountID  string
	FromCursor string
	Page       EnginePageRequest
}

// EngineEnumerationRequest — account-scoped запрос возобновляемого перечисления проекций.
type EngineEnumerationRequest struct {
	AccountID  string
	BaseCursor string
	Checkpoint string
	Page       EnginePageRequest
}

// EngineProjectionPage — account-scoped ограниченная страница проекций сообщений.
type EngineProjectionPage struct {
	AccountID  string
	Cursor     string
	NextCursor string
	Bounds     EnginePageBounds
	Messages   []EngineMessageProjection
}

// EngineAddress — ограниченный адрес сообщения или черновика.
type EngineAddress struct {
	Email string
	Name  string
}

// EngineBodyFormat — формат нормализованного тела сообщения.
type EngineBodyFormat string

const (
	// EngineBodyPlain обозначает текстовое тело сообщения.
	EngineBodyPlain EngineBodyFormat = "plain"
	// EngineBodyRich обозначает форматированное тело сообщения.
	EngineBodyRich EngineBodyFormat = "rich"
)

// EngineMessagePriority — безопасный приоритет сообщения.
type EngineMessagePriority string

const (
	// EnginePriorityLow обозначает низкий приоритет сообщения.
	EnginePriorityLow EngineMessagePriority = "low"
	// EnginePriorityNormal обозначает обычный приоритет сообщения.
	EnginePriorityNormal EngineMessagePriority = "normal"
	// EnginePriorityHigh обозначает высокий приоритет сообщения.
	EnginePriorityHigh EngineMessagePriority = "high"
)

// EngineAttachmentDisposition — способ представления вложения в сообщении.
type EngineAttachmentDisposition string

const (
	// EngineAttachmentDispositionAttachment обозначает отдельное вложение.
	EngineAttachmentDispositionAttachment EngineAttachmentDisposition = "attachment"
	// EngineAttachmentDispositionInline обозначает встроенное вложение.
	EngineAttachmentDispositionInline EngineAttachmentDisposition = "inline"
)

// EngineProjectionAttachment — ограниченные метаданные вложения в проекции сообщения.
type EngineProjectionAttachment struct {
	BlobID        string
	AttachmentKey string
	ContentID     string
	Filename      string
	MediaType     string
	SizeBytes     int64
	Disposition   EngineAttachmentDisposition
	SHA256        []byte
}

// MailAuthenticationClass — безопасный класс аутентификации календарного сообщения.
type MailAuthenticationClass string

const (
	// MailAuthenticationClassTrustedInternal обозначает доверенное внутреннее сообщение.
	MailAuthenticationClassTrustedInternal MailAuthenticationClass = "trusted_internal"
	// MailAuthenticationClassDMARCAligned обозначает внешнее DMARC-выровненное сообщение.
	MailAuthenticationClassDMARCAligned MailAuthenticationClass = "dmarc_aligned"
	// MailAuthenticationClassUntrusted обозначает неподтверждённое сообщение.
	MailAuthenticationClassUntrusted MailAuthenticationClass = "untrusted"
)

// EngineMessageProjection — ограниченные метаданные одного сообщения без его тела.
type EngineMessageProjection struct {
	AccountID         string
	MessageID         string
	Version           string
	InternetMessageID string
	InReplyTo         string
	References        []string
	From              EngineAddress
	To                []EngineAddress
	CC                []EngineAddress
	BCC               []EngineAddress
	ReplyTo           []EngineAddress
	Subject           string
	Snippet           string
	SentAt            *time.Time
	ReceivedAt        *time.Time
	CalendarMethod          *MailCalendarMethod
	AuthenticationClass     MailAuthenticationClass
	CalendarPartLocatorHash []byte
	SizeBytes         int64
	Priority          EngineMessagePriority
	FolderIDs         []string
	Flags             EngineMessageFlags
	ReservedKeywords  []EngineKeywordMarker
	Attachments       []EngineProjectionAttachment
	Origin            EngineProjectionOrigin
}

// EngineMessageBody — разрешённое нормализованное тело сообщения.
type EngineMessageBody struct {
	Format    EngineBodyFormat
	PlainBody string
	RichBody  string
}

// EngineDraftAttachment — полное вложение, передаваемое только при сохранении черновика.
type EngineDraftAttachment struct {
	AttachmentKey string
	ContentID     string
	Filename      string
	MediaType     string
	Disposition   EngineAttachmentDisposition
	SHA256        [32]byte
	Bytes         []byte
}

// EngineMessageContent — каноническое содержимое, передаваемое только движку.
type EngineMessageContent struct {
	From        EngineAddress
	To          []EngineAddress
	CC          []EngineAddress
	BCC         []EngineAddress
	ReplyTo     []EngineAddress
	Subject     string
	InReplyTo   string
	References  []string
	Format      EngineBodyFormat
	PlainBody   string
	RichBody    string
	Priority    EngineMessagePriority
	Attachments []EngineDraftAttachment
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

// EngineMessage — полное account-scoped сообщение с проекцией и разрешённым телом.
type EngineMessage struct {
	Projection EngineMessageProjection
	Body       EngineMessageBody
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
	CreatedOrUpdated []EngineMessageProjection
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
	Messages       []EngineMessageProjection
}

// EngineDraft — request-only полный черновик со стабильным ключом и ограничением содержимого.
type EngineDraft struct {
	AccountID       string
	DraftKey        string
	MessageID       string
	PreviousVersion string
	MutationID      string
	MaxContentBytes int64
	Message         EngineMessageContent
}

// EngineDraftReconciliation — account-scoped запрос сверки неопределённого черновика.
type EngineDraftReconciliation struct {
	AccountID       string
	DraftKey        string
	MessageID       string
	PreviousVersion string
	MutationID      string
	MaxContentBytes int64
}

// EngineDraftDeletion — request-only точное удаление черновика.
type EngineDraftDeletion struct {
	AccountID       string
	DraftKey        string
	MessageID       string
	ExpectedVersion string
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

// EngineDraftRetained is a confirmed observation that the current draft still
// has the exact previously confirmed message/version pair.
type EngineDraftRetained struct {
	Message EngineMessage
}

func (EngineDraftRetained) engineDraftResult() {}

// EngineDraftAbsent — подтверждённое отсутствие или отсутствие изменений черновика.
type EngineDraftAbsent struct{}

func (EngineDraftAbsent) engineDraftResult() {}

// EngineDraftDeletionResult — закрытый результат сверки удаления черновика.
type EngineDraftDeletionResult interface {
	engineDraftDeletionResult()
}

// EngineDraftDeleted — подтверждённое отсутствие черновика.
type EngineDraftDeleted struct{}

func (EngineDraftDeleted) engineDraftDeletionResult() {}

// EngineDraftRetainedUnchanged — подтверждённый неизменённый черновик.
type EngineDraftRetainedUnchanged struct{}

func (EngineDraftRetainedUnchanged) engineDraftDeletionResult() {}

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

// EngineDraftDeletionIndeterminate — неопределённый результат сверки удаления черновика.
type EngineDraftDeletionIndeterminate struct {
	Code EngineIndeterminateCode
}

func (EngineDraftDeletionIndeterminate) engineDraftDeletionResult() {}

// EngineSubmission — request-only ссылочная отправка подтверждённого черновика.
type EngineSubmission struct {
	AccountID       string
	DraftKey        string
	MessageID       string
	ExpectedVersion string
	SubmissionKey   string
	MaxMessageBytes int64
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
