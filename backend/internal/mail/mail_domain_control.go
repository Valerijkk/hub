package mail

import (
	"context"
	"crypto/sha256"
	"errors"
	"strings"

	"github.com/google/uuid"
)

const (
	mailDomainControlFingerprintVersion int32 = 1
	mailDomainControlOperation                 = "domain_create"
	mailDomainControlMaximumList               = 100
)

var (
	ErrMailDomainControlInvalid     = errors.New("mail domain control invalid")
	ErrMailDomainControlNotFound    = errors.New("mail domain control not found")
	ErrMailDomainControlConflict    = errors.New("mail domain control conflict")
	ErrMailDomainControlUnavailable = errors.New("mail domain control unavailable")
)

// MailDomainControlErrorKind is the closed error vocabulary for domain control.
type MailDomainControlErrorKind string

const (
	MailDomainControlErrorInvalid     MailDomainControlErrorKind = "invalid"
	MailDomainControlErrorNotFound    MailDomainControlErrorKind = "not_found"
	MailDomainControlErrorConflict    MailDomainControlErrorKind = "conflict"
	MailDomainControlErrorUnavailable MailDomainControlErrorKind = "unavailable"
)

// MailDomainControlError never carries provider, DNS, or database text.
type MailDomainControlError struct {
	Kind MailDomainControlErrorKind
}

func (errorValue *MailDomainControlError) Error() string {
	if errorValue == nil {
		return "mail domain control error"
	}
	switch errorValue.Kind {
	case MailDomainControlErrorInvalid:
		return ErrMailDomainControlInvalid.Error()
	case MailDomainControlErrorNotFound:
		return ErrMailDomainControlNotFound.Error()
	case MailDomainControlErrorConflict:
		return ErrMailDomainControlConflict.Error()
	case MailDomainControlErrorUnavailable:
		return ErrMailDomainControlUnavailable.Error()
	default:
		return "mail domain control error"
	}
}

func (errorValue *MailDomainControlError) Unwrap() error {
	if errorValue == nil {
		return nil
	}
	switch errorValue.Kind {
	case MailDomainControlErrorInvalid:
		return ErrMailDomainControlInvalid
	case MailDomainControlErrorNotFound:
		return ErrMailDomainControlNotFound
	case MailDomainControlErrorConflict:
		return ErrMailDomainControlConflict
	case MailDomainControlErrorUnavailable:
		return ErrMailDomainControlUnavailable
	default:
		return nil
	}
}

func mailDomainControlError(kind MailDomainControlErrorKind) error {
	return &MailDomainControlError{Kind: kind}
}

// MailDomainControlCreateRequest is the sealed request accepted from the future
// administrator boundary. Organization scope is resolved durably from ActorID.
type MailDomainControlCreateRequest struct {
	ActorID        uuid.UUID
	IdempotencyKey uuid.UUID
	Domain         string
}

// MailDomainControlListRequest requests the bounded, safe domain projection.
type MailDomainControlListRequest struct {
	ActorID uuid.UUID
}

// MailDomainControlSetDefaultRequest changes only an already verified domain.
type MailDomainControlSetDefaultRequest struct {
	ActorID  uuid.UUID
	DomainID uuid.UUID
}

// MailDomainControlView deliberately contains only user-safe domain state.
// It never exposes raw provider output, provisioning keys, leases, or errors.
type MailDomainControlView struct {
	ID                 uuid.UUID
	Domain             string
	State              string
	VerificationStatus string
	DKIMStatus         string
	RequiredDNSRecords RequiredDNSRecords
	IsDefault          bool
}

// MailDomainControlCreateResult is always the safe accepted representation.
type MailDomainControlCreateResult struct {
	Domain   MailDomainControlView
	Replayed bool
}

type mailDomainControlCreateRepositoryRequest struct {
	ActorID            uuid.UUID
	CanonicalDomain    string
	DomainID           uuid.UUID
	IdempotencyKeyHash [sha256.Size]byte
	RequestHash        [sha256.Size]byte
	DesiredHash        [sha256.Size]byte
	ProvisioningKey    string
	FingerprintVersion int32
}

// MailDomainControlRepository is intentionally small: no provider/DNS action
// is authorized by this control-plane core.
type MailDomainControlRepository interface {
	CreateMailDomain(context.Context, mailDomainControlCreateRepositoryRequest) (MailDomainControlCreateResult, error)
	ListMailDomains(context.Context, MailDomainControlListRequest) ([]MailDomainControlView, error)
	SetDefaultMailDomain(context.Context, MailDomainControlSetDefaultRequest) (MailDomainControlView, error)
}

// MailDomainControl is the closed MailDomain control-plane service.
type MailDomainControl struct {
	repository MailDomainControlRepository
}

func NewMailDomainControl(repository MailDomainControlRepository) (*MailDomainControl, error) {
	if mailboxEntNil(repository) {
		return nil, mailDomainControlError(MailDomainControlErrorInvalid)
	}
	return &MailDomainControl{repository: repository}, nil
}

// CreateMailDomain records the desired pending domain and a durable 202 replay receipt.
func (control *MailDomainControl) CreateMailDomain(ctx context.Context, request MailDomainControlCreateRequest) (MailDomainControlCreateResult, error) {
	if control == nil || mailboxEntNil(control.repository) || ctx == nil || !validMailDomainControlCreateRequest(request) {
		return MailDomainControlCreateResult{}, mailDomainControlError(MailDomainControlErrorInvalid)
	}
	canonicalDomain, ok := canonicalMailDomainControlDomain(request.Domain)
	if !ok {
		return MailDomainControlCreateResult{}, mailDomainControlError(MailDomainControlErrorInvalid)
	}
	repositoryRequest := mailDomainControlCreateRepositoryRequest{
		ActorID:            request.ActorID,
		CanonicalDomain:    canonicalDomain,
		DomainID:           mailDomainControlDomainID(request.ActorID, request.IdempotencyKey, canonicalDomain),
		IdempotencyKeyHash: mailDomainControlIdempotencyKeyHash(request.IdempotencyKey),
		RequestHash:        mailDomainControlRequestHash(request.ActorID, canonicalDomain),
		DesiredHash:        mailDomainControlDesiredHash(canonicalDomain),
		FingerprintVersion: mailDomainControlFingerprintVersion,
	}
	repositoryRequest.ProvisioningKey = mailDomainControlProvisioningKey(repositoryRequest.DomainID)
	if !validMailDomainControlCreateRepositoryRequest(repositoryRequest) {
		return MailDomainControlCreateResult{}, mailDomainControlError(MailDomainControlErrorInvalid)
	}
	result, err := control.repository.CreateMailDomain(ctx, repositoryRequest)
	if err != nil {
		return MailDomainControlCreateResult{}, mailDomainControlRepositoryError(err)
	}
	if !validMailDomainControlCreateResult(result, canonicalDomain) {
		return MailDomainControlCreateResult{}, mailDomainControlError(MailDomainControlErrorUnavailable)
	}
	return mailDomainControlCreateResultCopy(result), nil
}

// ListMailDomains returns an all-or-nothing bounded projection for the actor organization.
func (control *MailDomainControl) ListMailDomains(ctx context.Context, request MailDomainControlListRequest) ([]MailDomainControlView, error) {
	if control == nil || mailboxEntNil(control.repository) || ctx == nil || request.ActorID == uuid.Nil {
		return nil, mailDomainControlError(MailDomainControlErrorInvalid)
	}
	views, err := control.repository.ListMailDomains(ctx, request)
	if err != nil {
		return nil, mailDomainControlRepositoryError(err)
	}
	if len(views) > mailDomainControlMaximumList {
		return nil, mailDomainControlError(MailDomainControlErrorUnavailable)
	}
	seen := make(map[uuid.UUID]struct{}, len(views))
	result := make([]MailDomainControlView, 0, len(views))
	for _, view := range views {
		if !validMailDomainControlView(view) {
			return nil, mailDomainControlError(MailDomainControlErrorUnavailable)
		}
		if _, exists := seen[view.ID]; exists {
			return nil, mailDomainControlError(MailDomainControlErrorUnavailable)
		}
		seen[view.ID] = struct{}{}
		result = append(result, mailDomainControlViewCopy(view))
	}
	return result, nil
}

// SetDefaultMailDomain changes the default only through the durable repository fence.
func (control *MailDomainControl) SetDefaultMailDomain(ctx context.Context, request MailDomainControlSetDefaultRequest) (MailDomainControlView, error) {
	if control == nil || mailboxEntNil(control.repository) || ctx == nil || request.ActorID == uuid.Nil || request.DomainID == uuid.Nil {
		return MailDomainControlView{}, mailDomainControlError(MailDomainControlErrorInvalid)
	}
	view, err := control.repository.SetDefaultMailDomain(ctx, request)
	if err != nil {
		return MailDomainControlView{}, mailDomainControlRepositoryError(err)
	}
	if view.ID != request.DomainID || !view.IsDefault || !validMailDomainControlVerifiedView(view) {
		return MailDomainControlView{}, mailDomainControlError(MailDomainControlErrorUnavailable)
	}
	return mailDomainControlViewCopy(view), nil
}

func validMailDomainControlCreateRequest(request MailDomainControlCreateRequest) bool {
	return request.ActorID != uuid.Nil && validMailDomainControlIdempotencyKey(request.IdempotencyKey)
}

func validMailDomainControlIdempotencyKey(value uuid.UUID) bool {
	return value != uuid.Nil && (value.Version() == 4 || value.Version() == 7)
}

func canonicalMailDomainControlDomain(value string) (string, bool) {
	canonical := strings.ToLower(strings.TrimSpace(value))
	if len(canonical) == 0 || len(canonical) > 253 {
		return "", false
	}
	labels := strings.Split(canonical, ".")
	if len(labels) < 2 {
		return "", false
	}
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return "", false
		}
		for _, character := range label {
			if (character < 'a' || character > 'z') && (character < '0' || character > '9') && character != '-' {
				return "", false
			}
		}
	}
	return canonical, true
}

func mailDomainControlIdempotencyKeyHash(value uuid.UUID) [sha256.Size]byte {
	return sha256.Sum256([]byte("hub.mail.domain.idempotency.v1\x00" + value.String()))
}

func mailDomainControlRequestHash(actorID uuid.UUID, canonicalDomain string) [sha256.Size]byte {
	return sha256.Sum256([]byte("hub.mail.domain.create.request.v1\x00" + actorID.String() + "\x00" + canonicalDomain))
}

func mailDomainControlDesiredHash(canonicalDomain string) [sha256.Size]byte {
	return sha256.Sum256([]byte("hub.mail.domain.desired.v1\x00" + canonicalDomain + "\x00present=1\x00enabled=1"))
}

func mailDomainControlDomainID(actorID uuid.UUID, idempotencyKey uuid.UUID, canonicalDomain string) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte("hub.mail.domain.id.v1\x00"+actorID.String()+"\x00"+idempotencyKey.String()+"\x00"+canonicalDomain))
}

func mailDomainControlProvisioningKey(domainID uuid.UUID) string {
	if domainID == uuid.Nil {
		return ""
	}
	return "hub-mail-domain-v1:" + domainID.String()
}

func validMailDomainControlCreateRepositoryRequest(request mailDomainControlCreateRepositoryRequest) bool {
	if request.ActorID == uuid.Nil || request.DomainID == uuid.Nil || request.FingerprintVersion != mailDomainControlFingerprintVersion || !validMailDomainControlDigest(request.IdempotencyKeyHash[:]) || !validMailDomainControlDigest(request.RequestHash[:]) || !validMailDomainControlDigest(request.DesiredHash[:]) || request.ProvisioningKey != mailDomainControlProvisioningKey(request.DomainID) {
		return false
	}
	canonicalDomain, ok := canonicalMailDomainControlDomain(request.CanonicalDomain)
	return ok && canonicalDomain == request.CanonicalDomain
}

func validMailDomainControlDigest(value []byte) bool {
	if len(value) != sha256.Size {
		return false
	}
	for _, byteValue := range value {
		if byteValue != 0 {
			return true
		}
	}
	return false
}

func validMailDomainControlCreateResult(result MailDomainControlCreateResult, canonicalDomain string) bool {
	return validMailDomainControlView(result.Domain) && result.Domain.Domain == canonicalDomain && result.Domain.State == "pending" && result.Domain.VerificationStatus == "pending" && result.Domain.DKIMStatus == "unknown" && !result.Domain.IsDefault && len(result.Domain.RequiredDNSRecords) == 0
}

func validMailDomainControlVerifiedView(view MailDomainControlView) bool {
	return validMailDomainControlView(view) && view.State == "verified" && view.VerificationStatus == "verified"
}

func validMailDomainControlView(view MailDomainControlView) bool {
	canonicalDomain, ok := canonicalMailDomainControlDomain(view.Domain)
	if !ok || canonicalDomain != view.Domain || view.ID == uuid.Nil || !validMailDomainControlState(view.State) || !validMailDomainControlVerificationStatus(view.VerificationStatus) || !validMailDomainControlDKIMStatus(view.DKIMStatus) || view.RequiredDNSRecords == nil || view.RequiredDNSRecords.Validate() != nil {
		return false
	}
	return !view.IsDefault || view.State == "verified" && view.VerificationStatus == "verified"
}

func validMailDomainControlState(value string) bool {
	switch value {
	case "pending", "provisioning", "dns_required", "verified", "error", "deleting":
		return true
	default:
		return false
	}
}

func validMailDomainControlVerificationStatus(value string) bool {
	return value == "pending" || value == "verified" || value == "failed"
}

func validMailDomainControlDKIMStatus(value string) bool {
	return value == "unknown" || value == "pending" || value == "valid" || value == "invalid"
}

func mailDomainControlViewCopy(view MailDomainControlView) MailDomainControlView {
	return MailDomainControlView{
		ID:                 view.ID,
		Domain:             view.Domain,
		State:              view.State,
		VerificationStatus: view.VerificationStatus,
		DKIMStatus:         view.DKIMStatus,
		RequiredDNSRecords: append(RequiredDNSRecords(nil), view.RequiredDNSRecords...),
		IsDefault:          view.IsDefault,
	}
}

func mailDomainControlCreateResultCopy(result MailDomainControlCreateResult) MailDomainControlCreateResult {
	return MailDomainControlCreateResult{Domain: mailDomainControlViewCopy(result.Domain), Replayed: result.Replayed}
}

func mailDomainControlRepositoryError(err error) error {
	var controlError *MailDomainControlError
	if errors.As(err, &controlError) && controlError != nil {
		return mailDomainControlError(controlError.Kind)
	}
	return mailDomainControlError(MailDomainControlErrorUnavailable)
}
