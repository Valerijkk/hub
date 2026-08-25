package mail

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"mime"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/valerii/hub/backend/internal/ical"
	"github.com/valerii/hub/backend/internal/mail/stalwart"
)

const (
	stalwartCalendarEmailGetCallID  = "email-get-calendar"
	stalwartCalendarLocatorDomain   = "hub.mail.calendar-part-locator"
	stalwartCalendarLocatorVersion  = byte(1)
	stalwartCalendarMaximumHeaders  = MaxEngineReferenceItems
	stalwartCalendarMaximumTreeDepth = 64
	stalwartCalendarJMAPResponseMaxBytes int64 = MaxCalendarPartBytes
)

var (
	stalwartCalendarEmailProperties = []string{
		"id",
		"header:Authentication-Results:asRaw:all",
		"bodyStructure",
	}
	stalwartCalendarBodyProperties = []string{
		"partId",
		"blobId",
		"size",
		"type",
		"header:Content-Type:asRaw:all",
	}
)

type stalwartCalendarHydratedPart struct {
	version      string
	partID       string
	blobID       string
	method       MailCalendarMethod
	declaredSize int64
}

type stalwartCalendarBodyBounds struct {
	headerBytes int
	headerCount int
	nodeCount   int
}

// stalwartCalendarBoundedMailboxClient is deliberately narrower than
// StalwartMailboxClient so existing non-calendar fakes do not gain a new
// transport requirement. Calendar hydration fails closed if its production
// response-bound capability is unavailable.
type stalwartCalendarBoundedMailboxClient interface {
	CallMailboxBounded(context.Context, stalwart.JMAPRequest, int64) (stalwart.JMAPResponse, error)
}

// ReadCalendarPart rehydrates exactly one private iTIP leaf. It does not use
// GetMessage, projection paging, body values, or raw-message fallback.
func (engine *StalwartEngine) ReadCalendarPart(ctx context.Context, request EngineCalendarPartReadRequest) (EngineCalendarPartReadResult, error) {
	if !validStalwartCalendarPartReadRequest(request) || ctx == nil {
		return EngineCalendarPartReadResult{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	client, err := engine.mailboxClient(ctx, request.AccountID)
	if err != nil {
		return EngineCalendarPartReadResult{}, err
	}
	before, err := stalwartHydrateCalendarPart(ctx, client, request.AccountID, request.MessageID)
	if err != nil {
		return EngineCalendarPartReadResult{}, err
	}
	if before.version != request.Version || before.method != request.Method || before.declaredSize > int64(request.MaxBytes) {
		return EngineCalendarPartReadResult{}, stalwartEngineError(EngineErrorConflict)
	}
	itip, err := stalwartDownloadCalendarPart(ctx, client, request.AccountID, before, request.MaxBytes)
	if err != nil {
		return EngineCalendarPartReadResult{}, err
	}
	method, err := stalwartParseCalendarITIP(itip)
	if err != nil {
		return EngineCalendarPartReadResult{}, err
	}
	if method != before.method || method != request.Method {
		return EngineCalendarPartReadResult{}, stalwartEngineError(EngineErrorConflict)
	}
	beforeHash, err := stalwartCalendarLocatorHash(request.AccountID, request.MessageID, before.partID, before.blobID, before.method, itip)
	if err != nil {
		return EngineCalendarPartReadResult{}, err
	}
	if beforeHash != request.LocatorHash {
		return EngineCalendarPartReadResult{}, stalwartEngineError(EngineErrorConflict)
	}
	after, err := stalwartHydrateCalendarPart(ctx, client, request.AccountID, request.MessageID)
	if err != nil {
		if stalwartErrorHasKind(err, EngineErrorNotFound) {
			return EngineCalendarPartReadResult{}, stalwartEngineError(EngineErrorConflict)
		}
		return EngineCalendarPartReadResult{}, err
	}
	if after.version != request.Version || after.method != request.Method || after.declaredSize > int64(request.MaxBytes) {
		return EngineCalendarPartReadResult{}, stalwartEngineError(EngineErrorConflict)
	}
	afterHash, err := stalwartCalendarLocatorHash(request.AccountID, request.MessageID, after.partID, after.blobID, after.method, itip)
	if err != nil {
		return EngineCalendarPartReadResult{}, err
	}
	if afterHash != request.LocatorHash {
		return EngineCalendarPartReadResult{}, stalwartEngineError(EngineErrorConflict)
	}
	return EngineCalendarPartReadResult{
		LocatorHash: afterHash,
		ITIP:        append([]byte(nil), itip...),
	}, nil
}

func validStalwartCalendarPartReadRequest(request EngineCalendarPartReadRequest) bool {
	return validAccountID(request.AccountID) &&
		validOpaqueID(request.MessageID) &&
		validOpaqueID(request.Version) &&
		!calendarPartLocatorHashZero(request.LocatorHash) &&
		validCalendarPartMethod(request.Method) &&
		request.MaxBytes >= 1 && request.MaxBytes <= MaxCalendarPartBytes
}

// stalwartHydrateCalendarPart is deliberately exact-one only. In particular,
// callers must not route a projection page through this helper because it asks
// Stalwart for bodyStructure and bounded raw MIME headers.
func stalwartHydrateCalendarPart(ctx context.Context, client StalwartMailboxClient, accountID, messageID string) (stalwartCalendarHydratedPart, error) {
	if ctx == nil || client == nil || client.AccountID() != accountID || !validAccountID(accountID) || !validOpaqueID(messageID) {
		return stalwartCalendarHydratedPart{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	boundedClient, ok := client.(stalwartCalendarBoundedMailboxClient)
	if !ok {
		return stalwartCalendarHydratedPart{}, stalwartEngineError(EngineErrorUnavailable)
	}
	arguments, err := json.Marshal(stalwartEmailGetArguments{
		AccountID:      accountID,
		BodyProperties: append([]string(nil), stalwartCalendarBodyProperties...),
		IDs:            []string{messageID},
		Properties:     append([]string(nil), stalwartCalendarEmailProperties...),
	})
	if err != nil {
		return stalwartCalendarHydratedPart{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	response, err := boundedClient.CallMailboxBounded(ctx, stalwart.JMAPRequest{
		Using: stalwartMailboxCapabilities,
		MethodCalls: []stalwart.JMAPMethodCall{{
			Name:      "Email/get",
			Arguments: arguments,
			CallID:    stalwartCalendarEmailGetCallID,
		}},
	}, stalwartCalendarJMAPResponseMaxBytes)
	if err != nil {
		return stalwartCalendarHydratedPart{}, mapStalwartEngineError(err, false)
	}
	getResult, err := stalwartParseEmailGet(response, stalwartCalendarEmailGetCallID, accountID)
	if err != nil {
		return stalwartCalendarHydratedPart{}, err
	}
	if len(getResult.NotFound) != 0 {
		if len(getResult.NotFound) == 1 && getResult.NotFound[0] == messageID && len(getResult.List) == 0 {
			return stalwartCalendarHydratedPart{}, stalwartEngineError(EngineErrorNotFound)
		}
		return stalwartCalendarHydratedPart{}, stalwartEngineError(EngineErrorUnavailable)
	}
	if len(getResult.List) != 1 {
		return stalwartCalendarHydratedPart{}, stalwartEngineError(EngineErrorUnavailable)
	}
	record, err := stalwartDecodeEmailRecord(getResult.List[0], MaxCalendarPartBytes)
	if err != nil {
		return stalwartCalendarHydratedPart{}, err
	}
	if record.ID != messageID {
		return stalwartCalendarHydratedPart{}, stalwartEngineError(EngineErrorUnavailable)
	}
	// Authentication-Results remains retained only as bounded raw evidence in
	// this atom. No deployment strip/overwrite proof exists here, so all
	// calendar auth classification remains untrusted.
	if err := stalwartValidateCalendarAuthenticationResults(record.AuthenticationResults); err != nil {
		return stalwartCalendarHydratedPart{}, err
	}
	part, err := stalwartSelectCalendarLeaf(record.BodyStructure)
	if err != nil {
		return stalwartCalendarHydratedPart{}, err
	}
	return stalwartCalendarHydratedPart{
		version:      getResult.State,
		partID:       part.partID,
		blobID:       part.blobID,
		method:       part.method,
		declaredSize: part.declaredSize,
	}, nil
}

func stalwartValidateCalendarAuthenticationResults(values []string) error {
	_, err := stalwartUnfoldCalendarRawHeaders(values, stalwartCalendarMaximumHeaders, MaxEngineHeaderValueBytes, MaxEngineProjectionItemBytes)
	return err
}

func stalwartSelectCalendarLeaf(root *stalwartEmailBodyPart) (stalwartCalendarHydratedPart, error) {
	if root == nil {
		return stalwartCalendarHydratedPart{}, stalwartEngineError(EngineErrorConflict)
	}
	bounds := stalwartCalendarBodyBounds{}
	var selected *stalwartCalendarHydratedPart
	var walk func(stalwartEmailBodyPart, int) error
	walk = func(part stalwartEmailBodyPart, depth int) error {
		if depth > stalwartCalendarMaximumTreeDepth {
			return stalwartEngineError(EngineErrorPayloadTooLarge)
		}
		bounds.nodeCount++
		if bounds.nodeCount > stalwartMaximumBodyStructureNodes {
			return stalwartEngineError(EngineErrorPayloadTooLarge)
		}
		if (part.PartID != "" && !validOpaqueID(part.PartID)) ||
			(part.BlobID != "" && !validOpaqueID(part.BlobID)) ||
			part.Size < 0 || part.Size > MaxEngineMessageBytes ||
			!validSafeText(part.Type, MaxEngineMIMEFieldBytes, true) {
			return stalwartEngineError(EngineErrorUnavailable)
		}
		if err := bounds.consumeContentTypeHeaders(part.ContentType); err != nil {
			return err
		}
		if len(part.SubParts) == 0 && part.Type == "text/calendar" {
			if !validOpaqueID(part.PartID) || !validOpaqueID(part.BlobID) {
				return stalwartEngineError(EngineErrorConflict)
			}
			if part.Size > MaxCalendarPartBytes {
				return stalwartEngineError(EngineErrorPayloadTooLarge)
			}
			method, err := stalwartCalendarMIMEMethod(part.ContentType)
			if err != nil {
				return err
			}
			if selected != nil {
				return stalwartEngineError(EngineErrorConflict)
			}
			selected = &stalwartCalendarHydratedPart{
				partID:       part.PartID,
				blobID:       part.BlobID,
				method:       method,
				declaredSize: part.Size,
			}
		}
		for _, child := range part.SubParts {
			if err := walk(child, depth+1); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(*root, 1); err != nil {
		return stalwartCalendarHydratedPart{}, err
	}
	if selected == nil {
		return stalwartCalendarHydratedPart{}, stalwartEngineError(EngineErrorConflict)
	}
	return *selected, nil
}

func (bounds *stalwartCalendarBodyBounds) consumeContentTypeHeaders(values []string) error {
	if bounds == nil || len(values) > stalwartCalendarMaximumHeaders-bounds.headerCount {
		return stalwartEngineError(EngineErrorPayloadTooLarge)
	}
	for _, value := range values {
		if len(value) > MaxEngineMIMEFieldBytes {
			return stalwartEngineError(EngineErrorPayloadTooLarge)
		}
		bounds.headerCount++
		if len(value) > MaxEngineProjectionItemBytes-bounds.headerBytes {
			return stalwartEngineError(EngineErrorPayloadTooLarge)
		}
		bounds.headerBytes += len(value)
	}
	return nil
}

func stalwartCalendarMIMEMethod(values []string) (MailCalendarMethod, error) {
	if len(values) != 1 {
		return "", stalwartEngineError(EngineErrorUnavailable)
	}
	unfolded, err := stalwartUnfoldCalendarRawHeaders(values, 1, MaxEngineMIMEFieldBytes, MaxEngineMIMEFieldBytes)
	if err != nil {
		return "", err
	}
	mediaTypeToken := unfolded[0]
	if delimiter := strings.IndexByte(mediaTypeToken, ';'); delimiter >= 0 {
		mediaTypeToken = mediaTypeToken[:delimiter]
	}
	if strings.Trim(mediaTypeToken, " \t") != "text/calendar" {
		return "", stalwartEngineError(EngineErrorUnavailable)
	}
	mediaType, parameters, err := mime.ParseMediaType(unfolded[0])
	if err != nil || mediaType != "text/calendar" {
		return "", stalwartEngineError(EngineErrorUnavailable)
	}
	switch parameters["method"] {
	case "REQUEST":
		return MailCalendarMethodRequest, nil
	case "CANCEL":
		return MailCalendarMethodCancel, nil
	case "REPLY":
		return MailCalendarMethodReply, nil
	default:
		return "", stalwartEngineError(EngineErrorUnavailable)
	}
}

func stalwartUnfoldCalendarRawHeaders(values []string, maximumCount, maximumValueBytes, maximumTotalBytes int) ([]string, error) {
	if maximumCount < 1 || maximumValueBytes < 1 || maximumTotalBytes < 1 || len(values) > maximumCount {
		return nil, stalwartEngineError(EngineErrorPayloadTooLarge)
	}
	rawBytes := 0
	for _, value := range values {
		if len(value) > maximumValueBytes || len(value) > maximumTotalBytes-rawBytes {
			return nil, stalwartEngineError(EngineErrorPayloadTooLarge)
		}
		rawBytes += len(value)
	}

	unfolded := make([]string, 0, len(values))
	unfoldedBytes := 0
	for _, value := range values {
		result, err := stalwartUnfoldCalendarRawHeader(value, maximumValueBytes)
		if err != nil {
			return nil, err
		}
		if len(result) > maximumTotalBytes-unfoldedBytes {
			return nil, stalwartEngineError(EngineErrorPayloadTooLarge)
		}
		unfoldedBytes += len(result)
		unfolded = append(unfolded, result)
	}
	return unfolded, nil
}

func stalwartUnfoldCalendarRawHeader(value string, maximumBytes int) (string, error) {
	if value == "" || maximumBytes < 1 || len(value) > maximumBytes || !utf8.ValidString(value) {
		return "", stalwartEngineError(EngineErrorUnavailable)
	}
	for _, character := range value {
		if character != '\r' && character != '\n' && character != '\t' && unicode.IsControl(character) {
			return "", stalwartEngineError(EngineErrorUnavailable)
		}
	}
	var unfolded strings.Builder
	unfolded.Grow(len(value))
	for index := 0; index < len(value); {
		character := value[index]
		switch character {
		case '\r':
			if index+2 >= len(value) || value[index+1] != '\n' || (value[index+2] != ' ' && value[index+2] != '\t') {
				return "", stalwartEngineError(EngineErrorUnavailable)
			}
			index += 2
			for index < len(value) && (value[index] == ' ' || value[index] == '\t') {
				index++
			}
			if unfolded.Len() >= maximumBytes {
				return "", stalwartEngineError(EngineErrorPayloadTooLarge)
			}
			unfolded.WriteByte(' ')
		case '\n':
			return "", stalwartEngineError(EngineErrorUnavailable)
		case '\t':
			if unfolded.Len() >= maximumBytes {
				return "", stalwartEngineError(EngineErrorPayloadTooLarge)
			}
			unfolded.WriteByte(' ')
			index++
		default:
			if character < 0x20 || character == 0x7f {
				return "", stalwartEngineError(EngineErrorUnavailable)
			}
			if unfolded.Len() >= maximumBytes {
				return "", stalwartEngineError(EngineErrorPayloadTooLarge)
			}
			unfolded.WriteByte(character)
			index++
		}
	}
	if unfolded.Len() == 0 {
		return "", stalwartEngineError(EngineErrorUnavailable)
	}
	return unfolded.String(), nil
}

func stalwartDownloadCalendarPart(ctx context.Context, client StalwartMailboxClient, accountID string, part stalwartCalendarHydratedPart, maximum int) ([]byte, error) {
	if ctx == nil || client == nil || client.AccountID() != accountID || !validAccountID(accountID) ||
		!validOpaqueID(part.blobID) || part.declaredSize < 0 || part.declaredSize > MaxCalendarPartBytes ||
		maximum < 1 || maximum > MaxCalendarPartBytes || part.declaredSize > int64(maximum) {
		return nil, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	result, err := client.DownloadBlob(ctx, part.blobID, int64(maximum))
	if err != nil {
		return nil, mapStalwartEngineError(err, false)
	}
	byteLength := int64(len(result.Bytes))
	if result.AccountID != accountID {
		return nil, stalwartEngineError(EngineErrorUnauthorizedAccount)
	}
	if result.BlobID != part.blobID || result.SizeBytes < 0 || result.SizeBytes != byteLength || result.SizeBytes != part.declaredSize {
		return nil, stalwartEngineError(EngineErrorConflict)
	}
	if result.SizeBytes > int64(maximum) || result.SizeBytes > MaxCalendarPartBytes {
		return nil, stalwartEngineError(EngineErrorPayloadTooLarge)
	}
	return append([]byte(nil), result.Bytes...), nil
}

func stalwartParseCalendarITIP(payload []byte) (MailCalendarMethod, error) {
	if len(payload) == 0 || len(payload) > MaxCalendarPartBytes {
		return "", stalwartEngineError(EngineErrorPayloadTooLarge)
	}
	limits := ical.DefaultITIPParseLimits()
	if limits.MaxBytes > MaxCalendarPartBytes {
		limits.MaxBytes = MaxCalendarPartBytes
	}
	message, err := ical.ParseITIP(payload, limits)
	if err != nil {
		if errors.Is(err, ical.ErrITIPLimitExceeded) {
			return "", stalwartEngineError(EngineErrorPayloadTooLarge)
		}
		return "", stalwartEngineError(EngineErrorUnavailable)
	}
	switch message.Method {
	case ical.ITIPMethodRequest:
		return MailCalendarMethodRequest, nil
	case ical.ITIPMethodCancel:
		return MailCalendarMethodCancel, nil
	case ical.ITIPMethodReply:
		return MailCalendarMethodReply, nil
	default:
		return "", stalwartEngineError(EngineErrorUnavailable)
	}
}

func stalwartCalendarLocatorHash(accountID, messageID, partID, blobID string, method MailCalendarMethod, payload []byte) ([sha256.Size]byte, error) {
	if !validAccountID(accountID) || !validOpaqueID(messageID) || len(payload) == 0 || len(payload) > MaxCalendarPartBytes {
		return [sha256.Size]byte{}, stalwartEngineError(EngineErrorUnavailable)
	}
	locator, err := stalwartCalendarPrivateLocator(partID, blobID, method)
	if err != nil {
		return [sha256.Size]byte{}, err
	}
	payloadHash := sha256.Sum256(payload)
	hasher := sha256.New()
	_, _ = hasher.Write([]byte(stalwartCalendarLocatorDomain))
	_, _ = hasher.Write([]byte{0x00, stalwartCalendarLocatorVersion})
	stalwartCalendarHashFrame(hasher, []byte(accountID))
	stalwartCalendarHashFrame(hasher, []byte(messageID))
	stalwartCalendarHashFrame(hasher, locator)
	var payloadHashLength [4]byte
	binary.BigEndian.PutUint32(payloadHashLength[:], sha256.Size)
	_, _ = hasher.Write(payloadHashLength[:])
	_, _ = hasher.Write(payloadHash[:])
	var result [sha256.Size]byte
	copy(result[:], hasher.Sum(nil))
	return result, nil
}

func stalwartCalendarPrivateLocator(partID, blobID string, method MailCalendarMethod) ([]byte, error) {
	methodValue, ok := stalwartCalendarMIMEMethodValue(method)
	if !ok || !validOpaqueID(partID) || !validOpaqueID(blobID) {
		return nil, stalwartEngineError(EngineErrorUnavailable)
	}
	result := make([]byte, 0, 12+len(partID)+len(blobID)+len(methodValue))
	result = stalwartCalendarAppendFrame(result, []byte(partID))
	result = stalwartCalendarAppendFrame(result, []byte(blobID))
	result = stalwartCalendarAppendFrame(result, []byte(methodValue))
	return result, nil
}

func stalwartCalendarMIMEMethodValue(method MailCalendarMethod) (string, bool) {
	switch method {
	case MailCalendarMethodRequest:
		return "REQUEST", true
	case MailCalendarMethodCancel:
		return "CANCEL", true
	case MailCalendarMethodReply:
		return "REPLY", true
	default:
		return "", false
	}
}

func stalwartCalendarHashFrame(hasher interface{ Write([]byte) (int, error) }, value []byte) {
	var length [4]byte
	binary.BigEndian.PutUint32(length[:], uint32(len(value)))
	_, _ = hasher.Write(length[:])
	_, _ = hasher.Write(value)
}

func stalwartCalendarAppendFrame(target, value []byte) []byte {
	var length [4]byte
	binary.BigEndian.PutUint32(length[:], uint32(len(value)))
	target = append(target, length[:]...)
	return append(target, value...)
}
