package mail

import (
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

const (
	// MaxAttachmentPolicyItems limits the number of attachments in one policy decision.
	MaxAttachmentPolicyItems = 50
	// MaxAttachmentPolicyBytes limits one attachment under the policy contract.
	MaxAttachmentPolicyBytes int64 = 26_214_400
	// MaxAttachmentTextBytes limits one attachment metadata text value in bytes.
	MaxAttachmentTextBytes = 1020
	// MaxAttachmentTextRunes limits one attachment metadata text value in runes.
	MaxAttachmentTextRunes = 255
	// AttachmentSHA256Bytes is the required SHA-256 digest length.
	AttachmentSHA256Bytes = 32
)

// AttachmentPolicyOperation identifies the deterministic attachment policy operation.
type AttachmentPolicyOperation string

const (
	// AttachmentPolicyOperationValidate validates metadata without requiring a clean scan.
	AttachmentPolicyOperationValidate AttachmentPolicyOperation = "validate"
	// AttachmentPolicyOperationSend authorizes attachments for sending.
	AttachmentPolicyOperationSend AttachmentPolicyOperation = "send"
	// AttachmentPolicyOperationRead authorizes attachments for reading.
	AttachmentPolicyOperationRead AttachmentPolicyOperation = "read"
)

// AttachmentByteOrigin identifies how attachment size semantics were obtained.
type AttachmentByteOrigin string

const (
	// AttachmentByteOriginUpload identifies a newly uploaded attachment.
	AttachmentByteOriginUpload AttachmentByteOrigin = "upload"
	// AttachmentByteOriginEngineMIME identifies attachment metadata produced by the mail engine.
	AttachmentByteOriginEngineMIME AttachmentByteOrigin = "engine_mime"
)

// AttachmentScanStatus identifies the current malware-scan state.
type AttachmentScanStatus string

const (
	// AttachmentScanStatusPending identifies an attachment awaiting a scan decision.
	AttachmentScanStatusPending AttachmentScanStatus = "pending"
	// AttachmentScanStatusClean identifies an attachment accepted by the scanner.
	AttachmentScanStatusClean AttachmentScanStatus = "clean"
	// AttachmentScanStatusRejected identifies an attachment rejected by the scanner.
	AttachmentScanStatusRejected AttachmentScanStatus = "rejected"
	// AttachmentScanStatusError identifies a failed or unavailable scan decision.
	AttachmentScanStatusError AttachmentScanStatus = "error"
)

// AttachmentErrorKind classifies a content-free attachment policy failure.
type AttachmentErrorKind string

const (
	// AttachmentErrorInvalid identifies invalid attachment input or configuration.
	AttachmentErrorInvalid AttachmentErrorKind = "invalid"
	// AttachmentErrorTooLarge identifies an attachment bound violation.
	AttachmentErrorTooLarge AttachmentErrorKind = "too_large"
	// AttachmentErrorUnsafeType identifies an unsafe filename extension or media type.
	AttachmentErrorUnsafeType AttachmentErrorKind = "unsafe_type"
	// AttachmentErrorNotClean identifies an attachment that is not clean for send or read.
	AttachmentErrorNotClean AttachmentErrorKind = "not_clean"
)

var (
	// ErrAttachmentInvalid reports invalid attachment input without exposing its content.
	ErrAttachmentInvalid = errors.New("mail attachment invalid")
	// ErrAttachmentTooLarge reports an attachment bound violation without exposing its content.
	ErrAttachmentTooLarge = errors.New("mail attachment exceeds bounds")
	// ErrAttachmentUnsafeType reports an unsafe attachment type without exposing its content.
	ErrAttachmentUnsafeType = errors.New("mail attachment type rejected")
	// ErrAttachmentNotClean reports an unavailable clean scan decision without exposing content.
	ErrAttachmentNotClean = errors.New("mail attachment not clean")
)

// AttachmentPolicyError stores a stable content-free attachment policy failure kind.
type AttachmentPolicyError struct {
	Kind AttachmentErrorKind
}

// Error returns the stable content-free attachment policy error text.
func (errorValue *AttachmentPolicyError) Error() string {
	if errorValue == nil {
		return "mail attachment policy error"
	}

	switch errorValue.Kind {
	case AttachmentErrorInvalid:
		return ErrAttachmentInvalid.Error()
	case AttachmentErrorTooLarge:
		return ErrAttachmentTooLarge.Error()
	case AttachmentErrorUnsafeType:
		return ErrAttachmentUnsafeType.Error()
	case AttachmentErrorNotClean:
		return ErrAttachmentNotClean.Error()
	default:
		return "mail attachment policy error"
	}
}

// Unwrap returns the stable sentinel associated with a known attachment error kind.
func (errorValue *AttachmentPolicyError) Unwrap() error {
	if errorValue == nil {
		return nil
	}

	switch errorValue.Kind {
	case AttachmentErrorInvalid:
		return ErrAttachmentInvalid
	case AttachmentErrorTooLarge:
		return ErrAttachmentTooLarge
	case AttachmentErrorUnsafeType:
		return ErrAttachmentUnsafeType
	case AttachmentErrorNotClean:
		return ErrAttachmentNotClean
	default:
		return nil
	}
}

// AttachmentPolicyAttachment contains bounded metadata for one attachment policy decision.
type AttachmentPolicyAttachment struct {
	Origin AttachmentByteOrigin
	Filename string
	MediaType string
	ContentID string
	Disposition EngineAttachmentDisposition
	ScanStatus AttachmentScanStatus
	SizeBytes int64
	SHA256 []byte
}

// AttachmentPolicyInput contains all explicit inputs to deterministic attachment evaluation.
type AttachmentPolicyInput struct {
	Operation AttachmentPolicyOperation
	MaxAttachmentBytes int64
	Attachments []AttachmentPolicyAttachment
}

// AttachmentPolicyDecision contains independent normalized attachment metadata.
type AttachmentPolicyDecision struct {
	Attachments []AttachmentPolicyAttachment
}

// EvaluateAttachmentPolicy validates and normalizes attachment metadata without I/O or global state.
func EvaluateAttachmentPolicy(input AttachmentPolicyInput) (AttachmentPolicyDecision, error) {
	if !validAttachmentOperation(input.Operation) || input.MaxAttachmentBytes < 1 || input.MaxAttachmentBytes > MaxAttachmentPolicyBytes {
		return AttachmentPolicyDecision{}, attachmentPolicyError(AttachmentErrorInvalid)
	}
	if len(input.Attachments) > MaxAttachmentPolicyItems {
		return AttachmentPolicyDecision{}, attachmentPolicyError(AttachmentErrorTooLarge)
	}

	normalized := make([]AttachmentPolicyAttachment, len(input.Attachments))
	for index, attachment := range input.Attachments {
		canonicalAttachment, err := normalizeAttachmentShape(attachment, input.MaxAttachmentBytes)
		if err != nil {
			return AttachmentPolicyDecision{}, err
		}
		normalized[index] = canonicalAttachment
	}
	for _, attachment := range normalized {
		if unsafeAttachmentType(attachment.Filename, attachment.MediaType) {
			return AttachmentPolicyDecision{}, attachmentPolicyError(AttachmentErrorUnsafeType)
		}
	}
	if input.Operation != AttachmentPolicyOperationValidate {
		for _, attachment := range normalized {
			if attachment.ScanStatus != AttachmentScanStatusClean {
				return AttachmentPolicyDecision{}, attachmentPolicyError(AttachmentErrorNotClean)
			}
		}
	}

	return AttachmentPolicyDecision{Attachments: normalized}, nil
}

func validAttachmentOperation(operation AttachmentPolicyOperation) bool {
	return operation == AttachmentPolicyOperationValidate || operation == AttachmentPolicyOperationSend || operation == AttachmentPolicyOperationRead
}

func normalizeAttachmentShape(attachment AttachmentPolicyAttachment, maximum int64) (AttachmentPolicyAttachment, error) {
	if attachment.Origin != AttachmentByteOriginUpload && attachment.Origin != AttachmentByteOriginEngineMIME {
		return AttachmentPolicyAttachment{}, attachmentPolicyError(AttachmentErrorInvalid)
	}
	if attachment.Disposition != EngineAttachmentDispositionAttachment && attachment.Disposition != EngineAttachmentDispositionInline {
		return AttachmentPolicyAttachment{}, attachmentPolicyError(AttachmentErrorInvalid)
	}
	if attachment.ScanStatus != AttachmentScanStatusPending && attachment.ScanStatus != AttachmentScanStatusClean && attachment.ScanStatus != AttachmentScanStatusRejected && attachment.ScanStatus != AttachmentScanStatusError {
		return AttachmentPolicyAttachment{}, attachmentPolicyError(AttachmentErrorInvalid)
	}

	filename, err := normalizeAttachmentFilename(attachment.Filename)
	if err != nil {
		return AttachmentPolicyAttachment{}, err
	}
	mediaType, err := normalizeAttachmentMediaType(attachment.MediaType)
	if err != nil {
		return AttachmentPolicyAttachment{}, err
	}
	if err := validateAttachmentContentID(attachment.ContentID); err != nil {
		return AttachmentPolicyAttachment{}, err
	}
	if attachment.Disposition == EngineAttachmentDispositionInline && attachment.ContentID == "" {
		return AttachmentPolicyAttachment{}, attachmentPolicyError(AttachmentErrorInvalid)
	}
	if attachment.SizeBytes < 0 || attachment.Origin == AttachmentByteOriginUpload && attachment.SizeBytes == 0 {
		return AttachmentPolicyAttachment{}, attachmentPolicyError(AttachmentErrorInvalid)
	}
	if attachment.SizeBytes > maximum {
		return AttachmentPolicyAttachment{}, attachmentPolicyError(AttachmentErrorTooLarge)
	}
	if len(attachment.SHA256) != 0 && len(attachment.SHA256) != AttachmentSHA256Bytes {
		return AttachmentPolicyAttachment{}, attachmentPolicyError(AttachmentErrorInvalid)
	}
	if (attachment.ScanStatus == AttachmentScanStatusClean || attachment.ScanStatus == AttachmentScanStatusRejected) && len(attachment.SHA256) != AttachmentSHA256Bytes {
		return AttachmentPolicyAttachment{}, attachmentPolicyError(AttachmentErrorInvalid)
	}

	return AttachmentPolicyAttachment{
		Origin: attachment.Origin,
		Filename: filename,
		MediaType: mediaType,
		ContentID: attachment.ContentID,
		Disposition: attachment.Disposition,
		ScanStatus: attachment.ScanStatus,
		SizeBytes: attachment.SizeBytes,
		SHA256: append([]byte(nil), attachment.SHA256...),
	}, nil
}

func normalizeAttachmentFilename(value string) (string, error) {
	if !utf8.ValidString(value) {
		return "", attachmentPolicyError(AttachmentErrorInvalid)
	}
	value = norm.NFC.String(value)
	if err := validateAttachmentText(value, false); err != nil {
		return "", err
	}
	if err := validateAttachmentFilenameStructure(value); err != nil {
		return "", err
	}
	if err := validateAttachmentFilenameStructure(norm.NFKC.String(value)); err != nil {
		return "", err
	}
	return value, nil
}

func validateAttachmentFilenameStructure(value string) error {
	if value == "." || value == ".." || unicode.IsSpace(firstRune(value)) || unicode.IsSpace(lastRune(value)) || strings.HasPrefix(value, ".") || strings.HasSuffix(value, ".") || strings.ContainsAny(value, "/\\<>:\"|?*") {
		return attachmentPolicyError(AttachmentErrorInvalid)
	}
	for _, segment := range strings.Split(value, ".") {
		if segment == "" {
			return attachmentPolicyError(AttachmentErrorInvalid)
		}
	}
	return nil
}

func normalizeAttachmentMediaType(value string) (string, error) {
	if len(value) > MaxAttachmentTextBytes || len(value) > MaxAttachmentTextRunes {
		return "", attachmentPolicyError(AttachmentErrorTooLarge)
	}
	if value == "" || strings.Count(value, "/") != 1 {
		return "", attachmentPolicyError(AttachmentErrorInvalid)
	}
	parts := strings.Split(value, "/")
	if !validAttachmentMIMEToken(parts[0]) || !validAttachmentMIMEToken(parts[1]) {
		return "", attachmentPolicyError(AttachmentErrorInvalid)
	}
	return strings.ToLower(value), nil
}

func validAttachmentMIMEToken(value string) bool {
	if value == "" {
		return false
	}
	for index := 0; index < len(value); index++ {
		character := value[index]
		if character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9' || strings.ContainsRune("!#$&^_.+-", rune(character)) {
			continue
		}
		return false
	}
	return true
}

func validateAttachmentContentID(value string) error {
	if value == "" {
		return nil
	}
	return validateAttachmentText(value, false)
}

func validateAttachmentText(value string, allowEmpty bool) error {
	if !utf8.ValidString(value) || !allowEmpty && value == "" {
		return attachmentPolicyError(AttachmentErrorInvalid)
	}
	if len(value) > MaxAttachmentTextBytes || utf8.RuneCountInString(value) > MaxAttachmentTextRunes {
		return attachmentPolicyError(AttachmentErrorTooLarge)
	}
	for _, character := range value {
		if unicode.IsControl(character) || unicode.In(character, unicode.Cf) || character == '\u2028' || character == '\u2029' {
			return attachmentPolicyError(AttachmentErrorInvalid)
		}
	}
	return nil
}

func firstRune(value string) rune {
	character, _ := utf8.DecodeRuneInString(value)
	return character
}

func lastRune(value string) rune {
	character, _ := utf8.DecodeLastRuneInString(value)
	return character
}

func unsafeAttachmentType(filename, mediaType string) bool {
	skeleton := strings.ToLower(norm.NFKC.String(filename))
	for _, extension := range strings.Split(skeleton, ".")[1:] {
		if unsafeAttachmentExtension(extension) {
			return true
		}
	}
	return unsafeAttachmentMediaType(mediaType)
}

func unsafeAttachmentExtension(value string) bool {
	switch value {
	case "svg", "svgz", "exe", "com", "bat", "cmd", "msi", "msp", "mst", "scr", "pif", "cpl", "jar", "js", "jse", "vbs", "vbe", "wsf", "wsh", "ps1", "psm1", "psd1", "hta", "chm", "reg", "lnk", "url", "scf", "sh", "bash", "zsh", "fish", "app", "apk", "dmg", "pkg", "deb", "rpm", "iso", "img", "appimage":
		return true
	default:
		return false
	}
}

func unsafeAttachmentMediaType(value string) bool {
	switch value {
	case "image/svg+xml", "application/x-msdownload", "application/x-msdos-program", "application/vnd.microsoft.portable-executable", "application/x-dosexec", "application/x-msi", "application/java-archive", "application/x-java-archive", "application/javascript", "text/javascript", "application/ecmascript", "text/ecmascript", "application/x-sh", "application/x-csh", "application/x-powershell", "application/vnd.microsoft.windows-registry", "application/vnd.android.package-archive", "application/x-apple-diskimage", "application/x-iso9660-image", "application/wasm":
		return true
	default:
		return false
	}
}

func attachmentPolicyError(kind AttachmentErrorKind) error {
	return &AttachmentPolicyError{Kind: kind}
}
