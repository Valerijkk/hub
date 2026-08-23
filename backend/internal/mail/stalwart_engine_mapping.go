package mail

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/valerii/hub/backend/internal/mail/stalwart"
)

const (
	stalwartCursorList              = "list"
	stalwartCursorChanges           = "changes"
	stalwartCursorEnumerationPage   = "enumeration-page"
	stalwartMaximumQueryPosition    = 1_000_000_000
	stalwartMaximumBodyStructureNodes = 200
)

var stalwartEmailProjectionProperties = []string{
	"id",
	"threadId",
	"mailboxIds",
	"keywords",
	"size",
	"receivedAt",
	"sentAt",
	"messageId",
	"inReplyTo",
	"references",
	"from",
	"to",
	"cc",
	"replyTo",
	"subject",
	"preview",
	"attachments",
}

var stalwartEmailDetailProperties = append(append([]string(nil), stalwartEmailProjectionProperties...), "bodyStructure", "textBody", "htmlBody", "bodyValues")

type stalwartMessageCursor struct {
	AccountID string `json:"a"`
	BaseState string `json:"b,omitempty"`
	FolderID  string `json:"f,omitempty"`
	Kind      string `json:"k"`
	Position  int    `json:"p"`
	State     string `json:"s"`
}

type stalwartEmailQueryArguments struct {
	AccountID      string                    `json:"accountId"`
	CalculateTotal bool                      `json:"calculateTotal"`
	Filter         stalwartEmailQueryFilter  `json:"filter"`
	Limit          int                       `json:"limit"`
	Position       int                       `json:"position"`
	Sort           []stalwartEmailComparator `json:"sort"`
}

type stalwartEmailQueryFilter struct {
	InMailbox string `json:"inMailbox,omitempty"`
}

type stalwartEmailComparator struct {
	IsAscending bool   `json:"isAscending"`
	Property    string `json:"property"`
}

type stalwartEmailQueryResponse struct {
	AccountID  string   `json:"accountId"`
	IDs        []string `json:"ids"`
	Position   int      `json:"position"`
	QueryState string   `json:"queryState"`
	Total      *int     `json:"total"`
}

type stalwartEmailGetArguments struct {
	AccountID           string   `json:"accountId"`
	FetchHTMLBodyValues bool     `json:"fetchHTMLBodyValues,omitempty"`
	FetchTextBodyValues bool     `json:"fetchTextBodyValues,omitempty"`
	IDs                 []string `json:"ids"`
	MaxBodyValueBytes   int64    `json:"maxBodyValueBytes,omitempty"`
	Properties          []string `json:"properties"`
}

type stalwartEmailGetResponse struct {
	AccountID string            `json:"accountId"`
	List      []json.RawMessage `json:"list"`
	NotFound  []string          `json:"notFound"`
	State     string            `json:"state"`
}

type stalwartEmailChangesArguments struct {
	AccountID  string `json:"accountId"`
	MaxChanges int    `json:"maxChanges"`
	SinceState string `json:"sinceState"`
}

type stalwartEmailChangesResponse struct {
	AccountID       string   `json:"accountId"`
	Created         []string `json:"created"`
	Destroyed       []string `json:"destroyed"`
	HasMoreChanges  *bool    `json:"hasMoreChanges"`
	NewState        string   `json:"newState"`
	OldState        string   `json:"oldState"`
	Updated         []string `json:"updated"`
}

type stalwartEmailRecord struct {
	Attachments   []stalwartEmailBodyPart            `json:"attachments"`
	BodyStructure *stalwartEmailBodyPart              `json:"bodyStructure"`
	BodyValues    map[string]stalwartEmailBodyValue   `json:"bodyValues"`
	CC            []stalwartEmailAddress             `json:"cc"`
	From          []stalwartEmailAddress             `json:"from"`
	HTMLBody      []stalwartEmailBodyPart             `json:"htmlBody"`
	ID            string                              `json:"id"`
	InReplyTo     []string                            `json:"inReplyTo"`
	Keywords      map[string]bool                     `json:"keywords"`
	MailboxIDs    map[string]bool                     `json:"mailboxIds"`
	MessageID     []string                            `json:"messageId"`
	Preview       string                              `json:"preview"`
	ReceivedAt    *time.Time                          `json:"receivedAt"`
	References    []string                            `json:"references"`
	ReplyTo       []stalwartEmailAddress             `json:"replyTo"`
	SentAt        *time.Time                          `json:"sentAt"`
	Size          int64                               `json:"size"`
	Subject       string                              `json:"subject"`
	TextBody      []stalwartEmailBodyPart             `json:"textBody"`
	ThreadID      string                              `json:"threadId"`
	To            []stalwartEmailAddress             `json:"to"`
}

type stalwartEmailAddress struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type stalwartEmailBodyPart struct {
	BlobID      string                    `json:"blobId"`
	CID         string                    `json:"cid"`
	Disposition string                    `json:"disposition"`
	Name        string                    `json:"name"`
	PartID      string                    `json:"partId"`
	Size        int64                     `json:"size"`
	SubParts    []stalwartEmailBodyPart   `json:"subParts"`
	Type        string                    `json:"type"`
}

type stalwartEmailBodyValue struct {
	IsEncodingProblem bool   `json:"isEncodingProblem"`
	IsTruncated       bool   `json:"isTruncated"`
	Value             string `json:"value"`
}

func stalwartPackCursor(cursor stalwartMessageCursor) (string, error) {
	if !stalwartValidCursor(cursor) {
		return "", stalwartEngineError(EngineErrorInvalidCursor)
	}
	payload, err := json.Marshal(cursor)
	if err != nil || len(payload) == 0 || len(payload) > MaxEngineCursorBytes {
		return "", stalwartEngineError(EngineErrorInvalidCursor)
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	if !validCursor(encoded, false) {
		return "", stalwartEngineError(EngineErrorInvalidCursor)
	}
	return encoded, nil
}

func stalwartParseCursor(value, accountID, kind, folderID string) (stalwartMessageCursor, error) {
	if !validAccountID(accountID) || !validCursor(value, false) {
		return stalwartMessageCursor{}, stalwartEngineError(EngineErrorInvalidCursor)
	}
	payload, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil || len(payload) == 0 || len(payload) > MaxEngineCursorBytes {
		return stalwartMessageCursor{}, stalwartEngineError(EngineErrorInvalidCursor)
	}
	var cursor stalwartMessageCursor
	if err := json.Unmarshal(payload, &cursor); err != nil || !stalwartValidCursor(cursor) {
		return stalwartMessageCursor{}, stalwartEngineError(EngineErrorInvalidCursor)
	}
	canonical, err := json.Marshal(cursor)
	if err != nil || !bytes.Equal(payload, canonical) || base64.RawURLEncoding.EncodeToString(canonical) != value {
		return stalwartMessageCursor{}, stalwartEngineError(EngineErrorInvalidCursor)
	}
	if cursor.AccountID != accountID || cursor.Kind != kind || cursor.FolderID != folderID {
		return stalwartMessageCursor{}, stalwartEngineError(EngineErrorInvalidCursor)
	}
	return cursor, nil
}

func stalwartValidCursor(cursor stalwartMessageCursor) bool {
	if !validAccountID(cursor.AccountID) || !validOpaqueID(cursor.State) || cursor.Position < 0 || cursor.Position > stalwartMaximumQueryPosition {
		return false
	}
	switch cursor.Kind {
	case stalwartCursorList:
		return validOpaqueID(cursor.FolderID) && cursor.BaseState == ""
	case stalwartCursorChanges:
		return cursor.FolderID == "" && cursor.BaseState == ""
	case stalwartCursorEnumerationPage:
		return cursor.FolderID == "" && validOpaqueID(cursor.BaseState)
	default:
		return false
	}
}

func stalwartMethodPayload(response stalwart.JMAPResponse, name, callID string) (json.RawMessage, error) {
	if len(response.MethodResponses) != 1 {
		return nil, stalwartEngineError(EngineErrorUnavailable)
	}
	method := response.MethodResponses[0]
	if method.Name != name || method.CallID != callID || len(method.Arguments) == 0 {
		return nil, stalwartEngineError(EngineErrorUnavailable)
	}
	return append(json.RawMessage(nil), method.Arguments...), nil
}

func stalwartParseEmailQuery(response stalwart.JMAPResponse, callID, accountID string) (stalwartEmailQueryResponse, error) {
	payload, err := stalwartMethodPayload(response, "Email/query", callID)
	if err != nil {
		return stalwartEmailQueryResponse{}, err
	}
	var result stalwartEmailQueryResponse
	if err := json.Unmarshal(payload, &result); err != nil {
		return stalwartEmailQueryResponse{}, stalwartEngineError(EngineErrorUnavailable)
	}
	if result.AccountID != accountID {
		return stalwartEmailQueryResponse{}, stalwartEngineError(EngineErrorUnauthorizedAccount)
	}
	if !validOpaqueID(result.QueryState) || result.Position < 0 || result.Position > stalwartMaximumQueryPosition || result.Total == nil || *result.Total < result.Position || *result.Total > stalwartMaximumQueryPosition {
		return stalwartEmailQueryResponse{}, stalwartEngineError(EngineErrorUnavailable)
	}
	if len(result.IDs) > MaxEnginePageItems || !validOpaqueIDs(result.IDs) {
		return stalwartEmailQueryResponse{}, stalwartEngineError(EngineErrorPayloadTooLarge)
	}
	return result, nil
}

func stalwartParseEmailGet(response stalwart.JMAPResponse, callID, accountID string) (stalwartEmailGetResponse, error) {
	payload, err := stalwartMethodPayload(response, "Email/get", callID)
	if err != nil {
		return stalwartEmailGetResponse{}, err
	}
	var result stalwartEmailGetResponse
	if err := json.Unmarshal(payload, &result); err != nil {
		return stalwartEmailGetResponse{}, stalwartEngineError(EngineErrorUnavailable)
	}
	if result.AccountID != accountID {
		return stalwartEmailGetResponse{}, stalwartEngineError(EngineErrorUnauthorizedAccount)
	}
	if !validOpaqueID(result.State) || len(result.List) > MaxEnginePageItems || len(result.NotFound) > MaxEnginePageItems || !validOpaqueIDs(result.NotFound) {
		return stalwartEmailGetResponse{}, stalwartEngineError(EngineErrorPayloadTooLarge)
	}
	return result, nil
}

func stalwartParseEmailChanges(response stalwart.JMAPResponse, callID, accountID string) (stalwartEmailChangesResponse, error) {
	payload, err := stalwartMethodPayload(response, "Email/changes", callID)
	if err != nil {
		return stalwartEmailChangesResponse{}, err
	}
	var result stalwartEmailChangesResponse
	if err := json.Unmarshal(payload, &result); err != nil {
		return stalwartEmailChangesResponse{}, stalwartEngineError(EngineErrorUnavailable)
	}
	if result.AccountID != accountID {
		return stalwartEmailChangesResponse{}, stalwartEngineError(EngineErrorUnauthorizedAccount)
	}
	if result.HasMoreChanges == nil || !validOpaqueID(result.OldState) || !validOpaqueID(result.NewState) || len(result.Created) > MaxEnginePageItems || len(result.Updated) > MaxEnginePageItems || len(result.Destroyed) > MaxEnginePageItems || !validOpaqueIDs(result.Created) || !validOpaqueIDs(result.Updated) || !validOpaqueIDs(result.Destroyed) {
		return stalwartEmailChangesResponse{}, stalwartEngineError(EngineErrorPayloadTooLarge)
	}
	return result, nil
}

func stalwartDecodeEmailRecord(raw json.RawMessage, maximum int) (stalwartEmailRecord, error) {
	if len(raw) == 0 || len(raw) > maximum {
		return stalwartEmailRecord{}, stalwartEngineError(EngineErrorPayloadTooLarge)
	}
	var record stalwartEmailRecord
	if err := json.Unmarshal(raw, &record); err != nil {
		return stalwartEmailRecord{}, stalwartEngineError(EngineErrorUnavailable)
	}
	return record, nil
}

func stalwartMapEmailProjection(accountID, version string, record stalwartEmailRecord) (EngineMessageProjection, error) {
	if !validAccountID(accountID) || !validOpaqueID(version) || !validOpaqueID(record.ID) || !validOpaqueID(record.ThreadID) || record.Size < 0 || record.Size > MaxEngineMessageBytes || !validSafeText(record.Subject, MaxEngineHeaderValueBytes, false) || !validSafeText(record.Preview, MaxEngineSnippetBytes, false) {
		return EngineMessageProjection{}, stalwartEngineError(EngineErrorUnavailable)
	}
	from, err := stalwartMapAddresses(record.From, MaxEngineAddressItems)
	if err != nil || len(from) != 1 {
		return EngineMessageProjection{}, stalwartMappingResponseError(err)
	}
	to, err := stalwartMapAddresses(record.To, MaxEngineAddressItems)
	if err != nil {
		return EngineMessageProjection{}, stalwartMappingResponseError(err)
	}
	cc, err := stalwartMapAddresses(record.CC, MaxEngineAddressItems)
	if err != nil {
		return EngineMessageProjection{}, stalwartMappingResponseError(err)
	}
	replyTo, err := stalwartMapAddresses(record.ReplyTo, MaxEngineAddressItems)
	if err != nil {
		return EngineMessageProjection{}, stalwartMappingResponseError(err)
	}
	to, cc = stalwartDeduplicateRecipients(to, cc)
	internetMessageID, err := stalwartSingleMessageID(record.MessageID)
	if err != nil {
		return EngineMessageProjection{}, err
	}
	inReplyTo, err := stalwartSingleMessageID(record.InReplyTo)
	if err != nil {
		return EngineMessageProjection{}, err
	}
	references, err := stalwartMapReferences(record.References)
	if err != nil {
		return EngineMessageProjection{}, err
	}
	folderIDs, err := stalwartMapObjectKeys(record.MailboxIDs, MaxEngineFolderItems, true)
	if err != nil {
		return EngineMessageProjection{}, err
	}
	keywords, flags, err := stalwartMapKeywords(record.Keywords)
	if err != nil {
		return EngineMessageProjection{}, err
	}
	attachments, err := stalwartMapAttachments(record.Attachments)
	if err != nil {
		return EngineMessageProjection{}, err
	}
	sentAt, err := stalwartMapTime(record.SentAt, false)
	if err != nil {
		return EngineMessageProjection{}, err
	}
	receivedAt, err := stalwartMapTime(record.ReceivedAt, true)
	if err != nil {
		return EngineMessageProjection{}, err
	}
	projection := EngineMessageProjection{
		AccountID:           accountID,
		MessageID:           record.ID,
		Version:             version,
		InternetMessageID:   internetMessageID,
		InReplyTo:           inReplyTo,
		References:          references,
		From:                from[0],
		To:                  to,
		CC:                  cc,
		BCC:                 []EngineAddress{},
		ReplyTo:             replyTo,
		Subject:             record.Subject,
		Snippet:             record.Preview,
		SentAt:              sentAt,
		ReceivedAt:          receivedAt,
		AuthenticationClass: MailAuthenticationClassUntrusted,
		SizeBytes:           record.Size,
		Priority:            EnginePriorityNormal,
		FolderIDs:           folderIDs,
		Flags:               flags,
		ReservedKeywords:    keywords,
		Attachments:         attachments,
		Origin:              EngineProjectionOriginInboundOrUnknown,
	}
	if err := validateProjection(projection); err != nil {
		return EngineMessageProjection{}, stalwartMappingResponseError(err)
	}
	return projection, nil
}

func stalwartMapEmailBody(record stalwartEmailRecord, maximum int64) (EngineMessageBody, error) {
	if record.BodyStructure == nil || maximum < 1 || maximum > MaxEngineMessageBytes {
		return EngineMessageBody{}, stalwartEngineError(EngineErrorUnavailable)
	}
	partIDs := make(map[string]struct{})
	nodes := 0
	if err := stalwartValidateBodyStructure(*record.BodyStructure, 1, &nodes, partIDs); err != nil {
		return EngineMessageBody{}, err
	}
	if len(record.BodyValues) > nodes {
		return EngineMessageBody{}, stalwartEngineError(EngineErrorPayloadTooLarge)
	}
	for partID, value := range record.BodyValues {
		if !validOpaqueID(partID) {
			return EngineMessageBody{}, stalwartEngineError(EngineErrorUnavailable)
		}
		if _, exists := partIDs[partID]; !exists || value.IsTruncated {
			return EngineMessageBody{}, stalwartEngineError(EngineErrorPayloadTooLarge)
		}
		if value.IsEncodingProblem || !stalwartValidSourceBody(value.Value) {
			return EngineMessageBody{}, stalwartEngineError(EngineErrorUnavailable)
		}
	}
	plain, err := stalwartSelectBodyValue(record.TextBody, record.BodyValues, partIDs)
	if err != nil {
		return EngineMessageBody{}, err
	}
	rich, err := stalwartSelectBodyValue(record.HTMLBody, record.BodyValues, partIDs)
	if err != nil {
		return EngineMessageBody{}, err
	}
	if int64(len(plain))+int64(len(rich)) > maximum || len(plain) > MaxEngineBodyPartBytes || len(rich) > MaxEngineBodyPartBytes {
		return EngineMessageBody{}, stalwartEngineError(EngineErrorPayloadTooLarge)
	}
	if rich != "" {
		return EngineMessageBody{Format: EngineBodyRich, PlainBody: plain, RichBody: rich}, nil
	}
	return EngineMessageBody{Format: EngineBodyPlain, PlainBody: plain}, nil
}

func stalwartValidateBodyStructure(part stalwartEmailBodyPart, depth int, nodes *int, partIDs map[string]struct{}) error {
	if depth > 64 || nodes == nil || partIDs == nil {
		return stalwartEngineError(EngineErrorPayloadTooLarge)
	}
	*nodes = *nodes + 1
	if *nodes > stalwartMaximumBodyStructureNodes || !validOpaqueID(part.PartID) || (part.BlobID != "" && !validOpaqueID(part.BlobID)) || part.Size < 0 || part.Size > MaxEngineMessageBytes || !validSafeText(part.Type, MaxEngineMIMEFieldBytes, false) || !validSafeText(part.Name, MaxEngineMIMEFieldBytes, false) || !validSafeText(part.CID, MaxEngineMIMEFieldBytes, false) || !validSafeText(part.Disposition, MaxEngineMIMEFieldBytes, false) {
		return stalwartEngineError(EngineErrorUnavailable)
	}
	if _, duplicate := partIDs[part.PartID]; duplicate {
		return stalwartEngineError(EngineErrorUnavailable)
	}
	partIDs[part.PartID] = struct{}{}
	for _, child := range part.SubParts {
		if err := stalwartValidateBodyStructure(child, depth+1, nodes, partIDs); err != nil {
			return err
		}
	}
	return nil
}

func stalwartSelectBodyValue(parts []stalwartEmailBodyPart, values map[string]stalwartEmailBodyValue, known map[string]struct{}) (string, error) {
	if len(parts) > stalwartMaximumBodyStructureNodes {
		return "", stalwartEngineError(EngineErrorPayloadTooLarge)
	}
	for _, part := range parts {
		if !validOpaqueID(part.PartID) {
			return "", stalwartEngineError(EngineErrorUnavailable)
		}
		if _, exists := known[part.PartID]; !exists {
			return "", stalwartEngineError(EngineErrorUnavailable)
		}
		value, exists := values[part.PartID]
		if !exists || value.IsTruncated {
			return "", stalwartEngineError(EngineErrorPayloadTooLarge)
		}
		if value.IsEncodingProblem || !stalwartValidSourceBody(value.Value) {
			return "", stalwartEngineError(EngineErrorUnavailable)
		}
		if value.Value != "" {
			return value.Value, nil
		}
	}
	return "", nil
}

func stalwartValidSourceBody(value string) bool {
	if len(value) > MaxEngineBodyPartBytes || !utf8.ValidString(value) {
		return false
	}
	for _, character := range value {
		if character == '\t' || character == '\r' || character == '\n' {
			continue
		}
		if character < 0x20 || (character >= 0x7f && character <= 0x9f) {
			return false
		}
	}
	return true
}

func stalwartMapAddresses(values []stalwartEmailAddress, maximum int) ([]EngineAddress, error) {
	if len(values) > maximum {
		return nil, stalwartEngineError(EngineErrorPayloadTooLarge)
	}
	result := make([]EngineAddress, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		address, err := NormalizeAddress(EngineAddress{Email: value.Email, Name: value.Name})
		if err != nil {
			return nil, stalwartEngineError(EngineErrorUnavailable)
		}
		if _, duplicate := seen[address.Email]; duplicate {
			continue
		}
		seen[address.Email] = struct{}{}
		result = append(result, address)
	}
	return result, nil
}

func stalwartDeduplicateRecipients(to, cc []EngineAddress) ([]EngineAddress, []EngineAddress) {
	seen := make(map[string]struct{}, len(to)+len(cc))
	filter := func(values []EngineAddress) []EngineAddress {
		result := make([]EngineAddress, 0, len(values))
		for _, value := range values {
			if _, duplicate := seen[value.Email]; duplicate {
				continue
			}
			seen[value.Email] = struct{}{}
			result = append(result, value)
		}
		return result
	}
	return filter(to), filter(cc)
}

func stalwartSingleMessageID(values []string) (string, error) {
	if len(values) > MaxEngineReferenceItems {
		return "", stalwartEngineError(EngineErrorPayloadTooLarge)
	}
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if !validThreadMessageID(value) {
			return "", stalwartEngineError(EngineErrorUnavailable)
		}
		seen[value] = struct{}{}
	}
	if len(seen) > 1 {
		return "", stalwartEngineError(EngineErrorUnavailable)
	}
	for value := range seen {
		return value, nil
	}
	return "", nil
}

func stalwartMapReferences(values []string) ([]string, error) {
	if len(values) > MaxEngineReferenceItems {
		return nil, stalwartEngineError(EngineErrorPayloadTooLarge)
	}
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if !validThreadMessageID(value) {
			return nil, stalwartEngineError(EngineErrorUnavailable)
		}
		if _, duplicate := seen[value]; duplicate {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result, nil
}

func stalwartMapObjectKeys(values map[string]bool, maximum int, required bool) ([]string, error) {
	if len(values) > maximum {
		return nil, stalwartEngineError(EngineErrorPayloadTooLarge)
	}
	result := make([]string, 0, len(values))
	for value, present := range values {
		if !present || !validOpaqueID(value) {
			return nil, stalwartEngineError(EngineErrorUnavailable)
		}
		result = append(result, value)
	}
	sort.Strings(result)
	if required && len(result) == 0 {
		return nil, stalwartEngineError(EngineErrorUnavailable)
	}
	return result, nil
}

func stalwartMapKeywords(values map[string]bool) ([]EngineKeywordMarker, EngineMessageFlags, error) {
	if len(values) > MaxEngineKeywordItems {
		return nil, EngineMessageFlags{}, stalwartEngineError(EngineErrorPayloadTooLarge)
	}
	keys := make([]string, 0, len(values))
	for value, present := range values {
		if !present || !validOpaqueID(value) || !validSafeText(value, MaxEngineMIMEFieldBytes, true) {
			return nil, EngineMessageFlags{}, stalwartEngineError(EngineErrorUnavailable)
		}
		keys = append(keys, value)
	}
	sort.Strings(keys)
	keywords := make([]EngineKeywordMarker, len(keys))
	for index, value := range keys {
		keywords[index] = EngineKeywordMarker(value)
	}
	_, seen := values["$seen"]
	_, answered := values["$answered"]
	_, draft := values["$draft"]
	_, flagged := values["$flagged"]
	_, junk := values["$junk"]
	return keywords, EngineMessageFlags{Seen: seen, Answered: answered, Draft: draft, Flagged: flagged, Junk: junk}, nil
}

func stalwartMapAttachments(values []stalwartEmailBodyPart) ([]EngineProjectionAttachment, error) {
	if len(values) > MaxEngineAttachmentItems {
		return nil, stalwartEngineError(EngineErrorPayloadTooLarge)
	}
	result := make([]EngineProjectionAttachment, 0, len(values))
	byBlob := make(map[string]EngineProjectionAttachment, len(values))
	for _, value := range values {
		if !validOpaqueID(value.BlobID) || value.Size < 0 || value.Size > MaxEngineAttachmentBytes || !validSafeText(value.CID, MaxEngineMIMEFieldBytes, false) || !validSafeText(value.Name, MaxEngineMIMEFieldBytes, false) || !validSafeText(value.Type, MaxEngineMIMEFieldBytes, false) {
			return nil, stalwartEngineError(EngineErrorUnavailable)
		}
		disposition, ok := stalwartAttachmentDisposition(value.Disposition)
		if !ok || disposition == EngineAttachmentDispositionInline && value.CID == "" {
			return nil, stalwartEngineError(EngineErrorUnavailable)
		}
		attachment := EngineProjectionAttachment{BlobID: value.BlobID, ContentID: value.CID, Filename: value.Name, MediaType: strings.ToLower(value.Type), SizeBytes: value.Size, Disposition: disposition}
		if existing, duplicate := byBlob[value.BlobID]; duplicate {
			if !stalwartSameProjectionAttachment(existing, attachment) {
				return nil, stalwartEngineError(EngineErrorUnavailable)
			}
			continue
		}
		byBlob[value.BlobID] = attachment
		result = append(result, attachment)
	}
	sort.Slice(result, func(left, right int) bool {
		if result[left].BlobID != result[right].BlobID {
			return result[left].BlobID < result[right].BlobID
		}
		if result[left].ContentID != result[right].ContentID {
			return result[left].ContentID < result[right].ContentID
		}
		return result[left].Filename < result[right].Filename
	})
	return result, nil
}

func stalwartSameProjectionAttachment(left, right EngineProjectionAttachment) bool {
	return left.BlobID == right.BlobID && left.ContentID == right.ContentID && left.Filename == right.Filename && left.MediaType == right.MediaType && left.SizeBytes == right.SizeBytes && left.Disposition == right.Disposition && len(left.SHA256) == 0 && len(right.SHA256) == 0
}

func stalwartAttachmentDisposition(value string) (EngineAttachmentDisposition, bool) {
	switch value {
	case "", "attachment":
		return EngineAttachmentDispositionAttachment, true
	case "inline":
		return EngineAttachmentDispositionInline, true
	default:
		return "", false
	}
}

func stalwartMapTime(value *time.Time, required bool) (*time.Time, error) {
	if value == nil {
		if required {
			return nil, stalwartEngineError(EngineErrorUnavailable)
		}
		return nil, nil
	}
	canonical := value.UTC()
	if !syncValidTime(canonical) {
		return nil, stalwartEngineError(EngineErrorUnavailable)
	}
	return &canonical, nil
}

func stalwartMappingResponseError(err error) error {
	if err == nil {
		return nil
	}
	if engineError, ok := err.(*EngineError); ok && engineError.Kind == EngineErrorPayloadTooLarge {
		return stalwartEngineError(EngineErrorPayloadTooLarge)
	}
	return stalwartEngineError(EngineErrorUnavailable)
}
