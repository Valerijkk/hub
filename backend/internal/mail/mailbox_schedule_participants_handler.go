package mail

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

const mailboxScheduleParticipantsHTTPJSONBodyMaxBytes = 512 * 1024

// MailboxScheduleParticipantsHandler exposes the isolated authenticated
// scheduling-participant resolver without production registration.
type MailboxScheduleParticipantsHandler struct {
	service MailboxScheduleParticipantsServiceAPI
}

func NewMailboxScheduleParticipantsHandler(
	service MailboxScheduleParticipantsServiceAPI,
) (*MailboxScheduleParticipantsHandler, error) {
	if mailboxHTTPDependencyNil(service) {
		return nil, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
	}
	return &MailboxScheduleParticipantsHandler{service: service}, nil
}

func (handler *MailboxScheduleParticipantsHandler) RegisterRoutes(
	api fiber.Router,
	authMiddleware fiber.Handler,
) error {
	if handler == nil ||
		mailboxHTTPDependencyNil(handler.service) ||
		mailboxHTTPDependencyNil(api) ||
		mailboxHTTPDependencyNil(authMiddleware) {
		return NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
	}

	routes := api.Group("/mail", authMiddleware)
	routes.Post(
		"/messages/:messageId/scheduling-participants",
		handler.resolve,
	)
	return nil
}

type mailboxScheduleParticipantsHTTPRequest struct {
	Addresses *[]string `json:"addresses"`
}

type mailboxScheduleParticipantsHTTPParticipant struct {
	Email  string     `json:"email"`
	UserID *uuid.UUID `json:"user_id"`
}

type mailboxScheduleParticipantsHTTPResponse struct {
	Participants []mailboxScheduleParticipantsHTTPParticipant `json:"participants"`
}

func (handler *MailboxScheduleParticipantsHandler) resolve(ctx fiber.Ctx) error {
	if handler == nil ||
		mailboxHTTPDependencyNil(handler.service) ||
		mailboxHTTPDependencyNil(ctx) {
		return mailboxScheduleParticipantsHTTPInternal()
	}

	actorID, err := mailboxActorID(ctx)
	if err != nil {
		return err
	}
	if len(ctx.Request().URI().QueryString()) != 0 {
		return mailboxScheduleParticipantsHTTPInvalidRequest()
	}

	messageID, err := mailboxScheduleParticipantsHTTPCanonicalUUID(ctx.Params("messageId"))
	if err != nil {
		return mailboxScheduleParticipantsHTTPInvalidRequest()
	}
	request, err := mailboxScheduleParticipantsHTTPDecodeRequest(ctx)
	if err != nil {
		return err
	}

	addresses := make([]string, len(*request.Addresses))
	copy(addresses, *request.Addresses)
	result, err := handler.service.ResolveMailboxScheduleParticipants(
		ctx.Context(),
		actorID,
		MailboxScheduleParticipantsInput{
			MessageID: messageID,
			Addresses: addresses,
		},
	)
	if err != nil {
		return mailboxScheduleParticipantsHTTPServiceError(ctx, err)
	}

	response, valid := mailboxScheduleParticipantsHTTPResponseFromResult(result, addresses)
	if !valid {
		return mailboxScheduleParticipantsHTTPInternal()
	}
	if err := ctx.Status(http.StatusOK).JSON(response); err != nil {
		return mailboxScheduleParticipantsHTTPInternal()
	}
	return nil
}

func mailboxScheduleParticipantsHTTPCanonicalUUID(raw string) (uuid.UUID, error) {
	value, err := uuid.Parse(raw)
	if err != nil || value == uuid.Nil || raw != value.String() {
		return uuid.Nil, mailboxScheduleParticipantsHTTPInvalidRequest()
	}
	return value, nil
}

func mailboxScheduleParticipantsHTTPDecodeRequest(ctx fiber.Ctx) (mailboxScheduleParticipantsHTTPRequest, error) {
	value, one := mailboxSubmissionHTTPExactlyOneHeaderValue(ctx, fiber.HeaderContentType)
	if !one {
		return mailboxScheduleParticipantsHTTPRequest{}, mailboxScheduleParticipantsHTTPInvalidRequest()
	}
	mediaType, _, err := mime.ParseMediaType(strings.TrimSpace(string(value)))
	if err != nil || !strings.EqualFold(mediaType, fiber.MIMEApplicationJSON) {
		return mailboxScheduleParticipantsHTTPRequest{}, mailboxScheduleParticipantsHTTPInvalidRequest()
	}

	rawBody := ctx.BodyRaw()
	if len(rawBody) == 0 ||
		len(rawBody) > mailboxScheduleParticipantsHTTPJSONBodyMaxBytes ||
		!utf8.Valid(rawBody) {
		return mailboxScheduleParticipantsHTTPRequest{}, mailboxScheduleParticipantsHTTPInvalidRequest()
	}
	body := mailboxScheduleParticipantsHTTPTrimJSONWhitespace(rawBody)
	if len(body) == 0 || body[0] != '{' || !mailboxScheduleParticipantsHTTPStrictObject(body) {
		return mailboxScheduleParticipantsHTTPRequest{}, mailboxScheduleParticipantsHTTPInvalidRequest()
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	var request mailboxScheduleParticipantsHTTPRequest
	if err := decoder.Decode(&request); err != nil || request.Addresses == nil {
		return mailboxScheduleParticipantsHTTPRequest{}, mailboxScheduleParticipantsHTTPInvalidRequest()
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return mailboxScheduleParticipantsHTTPRequest{}, mailboxScheduleParticipantsHTTPInvalidRequest()
	}
	return request, nil
}

func mailboxScheduleParticipantsHTTPTrimJSONWhitespace(raw []byte) []byte {
	start := 0
	for start < len(raw) && mailboxScheduleParticipantsHTTPJSONWhitespace(raw[start]) {
		start++
	}
	end := len(raw)
	for end > start && mailboxScheduleParticipantsHTTPJSONWhitespace(raw[end-1]) {
		end--
	}
	return raw[start:end]
}

func mailboxScheduleParticipantsHTTPJSONWhitespace(value byte) bool {
	switch value {
	case ' ', '\t', '\n', '\r':
		return true
	default:
		return false
	}
}

func mailboxScheduleParticipantsHTTPStrictObject(body []byte) bool {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()

	token, err := decoder.Token()
	if err != nil {
		return false
	}
	delimiter, ok := token.(json.Delim)
	if !ok || delimiter != '{' {
		return false
	}

	seenAddresses := false
	for decoder.More() {
		key, err := decoder.Token()
		if err != nil {
			return false
		}
		name, ok := key.(string)
		if !ok || name != "addresses" || seenAddresses {
			return false
		}
		seenAddresses = true
		if !mailboxScheduleParticipantsHTTPStrictStringArray(decoder) {
			return false
		}
	}

	token, err = decoder.Token()
	if err != nil {
		return false
	}
	delimiter, ok = token.(json.Delim)
	if !ok || delimiter != '}' || !seenAddresses {
		return false
	}
	_, err = decoder.Token()
	return errors.Is(err, io.EOF)
}

func mailboxScheduleParticipantsHTTPStrictStringArray(decoder *json.Decoder) bool {
	token, err := decoder.Token()
	if err != nil {
		return false
	}
	delimiter, ok := token.(json.Delim)
	if !ok || delimiter != '[' {
		return false
	}

	for decoder.More() {
		token, err = decoder.Token()
		if err != nil {
			return false
		}
		if _, ok := token.(string); !ok {
			return false
		}
	}

	token, err = decoder.Token()
	if err != nil {
		return false
	}
	delimiter, ok = token.(json.Delim)
	return ok && delimiter == ']'
}

func mailboxScheduleParticipantsHTTPResponseFromResult(
	result MailboxScheduleParticipantsResult,
	addresses []string,
) (mailboxScheduleParticipantsHTTPResponse, bool) {
	if len(result.Participants) < 1 ||
		len(result.Participants) > MaxMailboxScheduleParticipantAddresses ||
		len(result.Participants) != len(addresses) {
		return mailboxScheduleParticipantsHTTPResponse{}, false
	}

	participants := make([]mailboxScheduleParticipantsHTTPParticipant, len(result.Participants))
	for index, participant := range result.Participants {
		if participant.Address != addresses[index] ||
			(participant.UserID != nil && *participant.UserID == uuid.Nil) {
			return mailboxScheduleParticipantsHTTPResponse{}, false
		}

		participants[index].Email = strings.Clone(participant.Address)
		if participant.UserID != nil {
			userID := *participant.UserID
			participants[index].UserID = &userID
		}
	}
	return mailboxScheduleParticipantsHTTPResponse{Participants: participants}, true
}

func mailboxScheduleParticipantsHTTPServiceError(ctx fiber.Ctx, err error) error {
	kind, ok := MailboxScheduleParticipantsErrorKindOf(err)
	if !ok {
		return mailboxScheduleParticipantsHTTPInternal()
	}

	switch kind {
	case MailboxScheduleParticipantsErrorInvalidRequest:
		return mailboxScheduleParticipantsHTTPInvalidRequest()
	case MailboxScheduleParticipantsErrorNotFound:
		return fiber.NewError(http.StatusNotFound, "mail message not found")
	case MailboxScheduleParticipantsErrorDeleting:
		ctx.Set("X-Hub-Error-Code", "MAIL_ACCOUNT_DELETING")
		return fiber.NewError(http.StatusConflict, "mailbox account deleting")
	case MailboxScheduleParticipantsErrorUnavailable:
		return fiber.NewError(http.StatusServiceUnavailable, "mail message unavailable")
	case MailboxScheduleParticipantsErrorInternal:
		return mailboxScheduleParticipantsHTTPInternal()
	default:
		return mailboxScheduleParticipantsHTTPInternal()
	}
}

func mailboxScheduleParticipantsHTTPInvalidRequest() error {
	return fiber.NewError(http.StatusBadRequest, "invalid mail scheduling request")
}

func mailboxScheduleParticipantsHTTPInternal() error {
	return fiber.NewError(http.StatusInternalServerError, "mail message unavailable")
}
