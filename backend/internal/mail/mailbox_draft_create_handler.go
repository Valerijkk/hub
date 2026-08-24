package mail

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"mime"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

const mailboxDraftCreateHTTPJSONBodyMaxBytes = 20 * 1024 * 1024

type mailboxDraftCreateService interface {
	CreateMailboxDraft(context.Context, uuid.UUID, MailboxDraftCreateInput) (MailboxDraftView, error)
	GetMailboxDraft(context.Context, uuid.UUID, uuid.UUID) (MailboxDraftView, error)
	SaveMailboxDraft(context.Context, uuid.UUID, uuid.UUID, MailboxDraftSaveInput) (MailboxDraftView, error)
}

// MailboxDraftCreateHandler exposes the authenticated durable draft-create transport.
type MailboxDraftCreateHandler struct {
	service mailboxDraftCreateService
}

// NewMailboxDraftCreateHandler creates the authenticated draft create/read handler.
func NewMailboxDraftCreateHandler(service mailboxDraftCreateService) (*MailboxDraftCreateHandler, error) {
	if mailboxHTTPDependencyNil(service) {
		return nil, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	return &MailboxDraftCreateHandler{service: service}, nil
}

// RegisterRoutes registers the authenticated draft create/read routes.
func (handler *MailboxDraftCreateHandler) RegisterRoutes(api fiber.Router, authMiddleware fiber.Handler) error {
	if handler == nil || mailboxHTTPDependencyNil(handler.service) || mailboxHTTPDependencyNil(api) || mailboxHTTPDependencyNil(authMiddleware) {
		return mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	routes := api.Group("/mail", authMiddleware)
	routes.Post("/drafts", handler.create)
	routes.Get("/drafts/:draftId", handler.get)
	routes.Patch("/drafts/:draftId", handler.save)
	return nil
}

type mailboxDraftCreateHTTPRequest struct {
	DraftClientID    string                           `json:"draft_client_id"`
	FromIdentityID   string                           `json:"from_identity_id"`
	To               *[]mailboxDraftCreateHTTPAddress `json:"to"`
	CC               *[]mailboxDraftCreateHTTPAddress `json:"cc"`
	BCC              *[]mailboxDraftCreateHTTPAddress `json:"bcc"`
	Subject          *string                          `json:"subject"`
	BodyFormat       string                           `json:"body_format"`
	RichBody         mailboxDraftCreateHTTPBodyField   `json:"rich_body"`
	PlainBody        mailboxDraftCreateHTTPBodyField   `json:"plain_body"`
	Priority         string                           `json:"priority"`
	AttachmentIDs    *[]string                        `json:"attachment_ids"`
	ReplyToMessageID *string                          `json:"reply_to_message_id"`
	ForwardMessageID *string                          `json:"forward_message_id"`
}

type mailboxDraftSaveHTTPRequest struct {
	DraftClientID  string                           `json:"draft_client_id"`
	FromIdentityID string                           `json:"from_identity_id"`
	To             *[]mailboxDraftCreateHTTPAddress `json:"to"`
	CC             *[]mailboxDraftCreateHTTPAddress `json:"cc"`
	BCC            *[]mailboxDraftCreateHTTPAddress `json:"bcc"`
	Subject        *string                          `json:"subject"`
	BodyFormat     string                           `json:"body_format"`
	RichBody       mailboxDraftCreateHTTPBodyField   `json:"rich_body"`
	PlainBody      mailboxDraftCreateHTTPBodyField   `json:"plain_body"`
	Priority       string                           `json:"priority"`
	AttachmentIDs  *[]string                        `json:"attachment_ids"`
}

var mailboxDraftSaveHTTPIfMatchPattern = regexp.MustCompile(`^"[1-9][0-9]*"$`)

type mailboxDraftCreateHTTPAddress struct {
	Email string  `json:"email"`
	Name  *string `json:"name"`
}

// mailboxDraftCreateHTTPBodyField retains JSON presence independently from a
// string value so oneOf accepts an omitted opposite body, never an explicit null.
type mailboxDraftCreateHTTPBodyField struct {
	present bool
	value   string
}

func (field *mailboxDraftCreateHTTPBodyField) UnmarshalJSON(raw []byte) error {
	if field == nil || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return errors.New("invalid mail draft body field")
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return err
	}
	field.present = true
	field.value = value
	return nil
}

type mailboxDraftCreateHTTPResponse struct {
	ID               string                          `json:"id"`
	DraftClientID    string                          `json:"draft_client_id"`
	Revision         int64                           `json:"revision"`
	Status           string                          `json:"status"`
	FromIdentityID   string                          `json:"from_identity_id"`
	To               []mailboxDraftCreateHTTPAddress `json:"to"`
	CC               []mailboxDraftCreateHTTPAddress `json:"cc"`
	BCC              []mailboxDraftCreateHTTPAddress `json:"bcc"`
	Subject          string                          `json:"subject"`
	BodyFormat       string                          `json:"body_format"`
	RichBody         *string                         `json:"rich_body"`
	PlainBody        string                          `json:"plain_body"`
	Priority         string                          `json:"priority"`
	AttachmentIDs    []string                        `json:"attachment_ids"`
	ReplyToMessageID *string                         `json:"reply_to_message_id"`
	ForwardMessageID *string                         `json:"forward_message_id"`
	UpdatedAt        time.Time                       `json:"updated_at"`
}

func (handler *MailboxDraftCreateHandler) create(ctx fiber.Ctx) error {
	if handler == nil || mailboxHTTPDependencyNil(handler.service) {
		return mailboxDraftCreateHTTPUnavailable()
	}
	if err := mailboxDraftCreateHTTPRejectQuery(ctx); err != nil {
		return err
	}
	actorID, err := mailboxActorID(ctx)
	if err != nil {
		return err
	}
	var request mailboxDraftCreateHTTPRequest
	if err := mailboxDraftCreateHTTPDecodeJSON(ctx, &request); err != nil {
		return err
	}
	input, err := mailboxDraftCreateHTTPInput(request, ctx.Get("Idempotency-Key"))
	if err != nil {
		return err
	}
	view, err := handler.service.CreateMailboxDraft(ctx.Context(), actorID, input)
	if err != nil {
		return mailboxDraftCreateHTTPServiceError(ctx, err)
	}
	response, ok := mailboxDraftCreateHTTPResponseFromView(view)
	if !ok {
		return mailboxDraftCreateHTTPUnavailable()
	}
	return ctx.Status(http.StatusCreated).JSON(response)
}

func (handler *MailboxDraftCreateHandler) get(ctx fiber.Ctx) error {
	if handler == nil || mailboxHTTPDependencyNil(handler.service) {
		return mailboxDraftCreateHTTPUnavailable()
	}
	if err := mailboxDraftCreateHTTPRejectQuery(ctx); err != nil {
		return err
	}
	if err := mailboxDraftReadHTTPRejectBody(ctx); err != nil {
		return err
	}
	actorID, err := mailboxActorID(ctx)
	if err != nil {
		return err
	}
	draftID, err := mailboxDraftCreateHTTPCanonicalUUID(ctx.Params("draftId"))
	if err != nil {
		return err
	}
	view, err := handler.service.GetMailboxDraft(ctx.Context(), actorID, draftID)
	if err != nil {
		return mailboxDraftReadHTTPServiceError(ctx, err)
	}
	response, ok := mailboxDraftCreateHTTPResponseFromView(view)
	if !ok {
		return mailboxDraftCreateHTTPUnavailable()
	}
	return ctx.Status(http.StatusOK).JSON(response)
}

func (handler *MailboxDraftCreateHandler) save(ctx fiber.Ctx) error {
	if handler == nil || mailboxHTTPDependencyNil(handler.service) {
		return mailboxDraftCreateHTTPUnavailable()
	}
	if err := mailboxDraftCreateHTTPRejectQuery(ctx); err != nil {
		return err
	}
	if err := mailboxDraftSaveHTTPRejectIdempotencyKey(ctx); err != nil {
		return err
	}
	actorID, err := mailboxActorID(ctx)
	if err != nil {
		return err
	}
	draftID, err := mailboxDraftCreateHTTPCanonicalUUID(ctx.Params("draftId"))
	if err != nil {
		return err
	}
	expectedRevision, err := mailboxDraftSaveHTTPExpectedRevision(ctx.Get("If-Match"))
	if err != nil {
		return err
	}
	var request mailboxDraftSaveHTTPRequest
	if err := mailboxDraftCreateHTTPDecodeJSON(ctx, &request); err != nil {
		return err
	}
	input, err := mailboxDraftSaveHTTPInput(request, expectedRevision)
	if err != nil {
		return err
	}
	view, err := handler.service.SaveMailboxDraft(ctx.Context(), actorID, draftID, input)
	if err != nil {
		return mailboxDraftSaveHTTPServiceError(ctx, err)
	}
	response, ok := mailboxDraftCreateHTTPResponseFromView(view)
	if !ok {
		return mailboxDraftCreateHTTPUnavailable()
	}
	return ctx.Status(http.StatusOK).JSON(response)
}

func mailboxDraftCreateHTTPRejectQuery(ctx fiber.Ctx) error {
	if ctx == nil || len(ctx.Request().URI().QueryString()) != 0 {
		return mailboxDraftCreateHTTPInvalidRequest()
	}
	return nil
}

func mailboxDraftReadHTTPRejectBody(ctx fiber.Ctx) error {
	if ctx == nil || len(ctx.BodyRaw()) != 0 || strings.TrimSpace(ctx.Get(fiber.HeaderContentType)) != "" {
		return mailboxDraftCreateHTTPInvalidRequest()
	}
	return nil
}

func mailboxDraftCreateHTTPDecodeJSON(ctx fiber.Ctx, target any) error {
	if ctx == nil || target == nil {
		return mailboxDraftCreateHTTPInvalidRequest()
	}
	mediaType, _, err := mime.ParseMediaType(strings.TrimSpace(ctx.Get(fiber.HeaderContentType)))
	if err != nil || !strings.EqualFold(mediaType, "application/json") {
		return mailboxDraftCreateHTTPInvalidRequest()
	}
	rawBody := ctx.BodyRaw()
	if len(rawBody) > mailboxDraftCreateHTTPJSONBodyMaxBytes {
		return fiber.NewError(http.StatusRequestEntityTooLarge, "mail draft request body too large")
	}
	body := bytes.TrimSpace(rawBody)
	if len(body) == 0 || body[0] != '{' {
		return mailboxDraftCreateHTTPInvalidRequest()
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return mailboxDraftCreateHTTPInvalidRequest()
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return mailboxDraftCreateHTTPInvalidRequest()
	}
	return nil
}

func mailboxDraftCreateHTTPInput(request mailboxDraftCreateHTTPRequest, idempotencyKey string) (MailboxDraftCreateInput, error) {
	draftClientID, err := mailboxDraftCreateHTTPCanonicalUUID(request.DraftClientID)
	if err != nil {
		return MailboxDraftCreateInput{}, err
	}
	fromIdentityID, err := mailboxDraftCreateHTTPCanonicalUUID(request.FromIdentityID)
	if err != nil {
		return MailboxDraftCreateInput{}, err
	}
	if request.To == nil || request.CC == nil || request.BCC == nil || request.Subject == nil || request.AttachmentIDs == nil {
		return MailboxDraftCreateInput{}, mailboxDraftCreateHTTPInvalidRequest()
	}
	to, err := mailboxDraftCreateHTTPAddresses(*request.To)
	if err != nil {
		return MailboxDraftCreateInput{}, err
	}
	cc, err := mailboxDraftCreateHTTPAddresses(*request.CC)
	if err != nil {
		return MailboxDraftCreateInput{}, err
	}
	bcc, err := mailboxDraftCreateHTTPAddresses(*request.BCC)
	if err != nil {
		return MailboxDraftCreateInput{}, err
	}
	attachments, err := mailboxDraftCreateHTTPAttachments(*request.AttachmentIDs)
	if err != nil {
		return MailboxDraftCreateInput{}, err
	}
	content, err := mailboxDraftCreateHTTPContent(request.BodyFormat, request.RichBody, request.PlainBody)
	if err != nil {
		return MailboxDraftCreateInput{}, err
	}
	priority, err := mailboxDraftCreateHTTPPriority(request.Priority)
	if err != nil {
		return MailboxDraftCreateInput{}, err
	}
	replyToMessageID, err := mailboxDraftCreateHTTPOptionalUUID(request.ReplyToMessageID)
	if err != nil {
		return MailboxDraftCreateInput{}, err
	}
	forwardMessageID, err := mailboxDraftCreateHTTPOptionalUUID(request.ForwardMessageID)
	if err != nil {
		return MailboxDraftCreateInput{}, err
	}
	return MailboxDraftCreateInput{
		MailboxDraftWriteInput: MailboxDraftWriteInput{
			DraftClientID:    draftClientID,
			FromIdentityID:   fromIdentityID,
			To:               to,
			CC:               cc,
			BCC:              bcc,
			Subject:          *request.Subject,
			Priority:         priority,
			ReplyToMessageID: replyToMessageID,
			ForwardMessageID: forwardMessageID,
			Attachments:      attachments,
			Content:          content,
		},
		IdempotencyKey: idempotencyKey,
	}, nil
}

func mailboxDraftSaveHTTPInput(request mailboxDraftSaveHTTPRequest, expectedRevision int64) (MailboxDraftSaveInput, error) {
	draftClientID, err := mailboxDraftCreateHTTPCanonicalUUID(request.DraftClientID)
	if err != nil {
		return MailboxDraftSaveInput{}, err
	}
	fromIdentityID, err := mailboxDraftCreateHTTPCanonicalUUID(request.FromIdentityID)
	if err != nil {
		return MailboxDraftSaveInput{}, err
	}
	if expectedRevision < 1 || request.To == nil || request.CC == nil || request.BCC == nil || request.Subject == nil || request.AttachmentIDs == nil {
		return MailboxDraftSaveInput{}, mailboxDraftCreateHTTPInvalidRequest()
	}
	if len(*request.AttachmentIDs) != 0 {
		return MailboxDraftSaveInput{}, fiber.NewError(http.StatusUnprocessableEntity, "mail draft attachments cannot be saved")
	}
	to, err := mailboxDraftCreateHTTPAddresses(*request.To)
	if err != nil {
		return MailboxDraftSaveInput{}, err
	}
	cc, err := mailboxDraftCreateHTTPAddresses(*request.CC)
	if err != nil {
		return MailboxDraftSaveInput{}, err
	}
	bcc, err := mailboxDraftCreateHTTPAddresses(*request.BCC)
	if err != nil {
		return MailboxDraftSaveInput{}, err
	}
	content, err := mailboxDraftCreateHTTPContent(request.BodyFormat, request.RichBody, request.PlainBody)
	if err != nil {
		return MailboxDraftSaveInput{}, err
	}
	priority, err := mailboxDraftCreateHTTPPriority(request.Priority)
	if err != nil {
		return MailboxDraftSaveInput{}, err
	}
	return MailboxDraftSaveInput{
		MailboxDraftWriteInput: MailboxDraftWriteInput{
			DraftClientID:    draftClientID,
			FromIdentityID:   fromIdentityID,
			To:               to,
			CC:               cc,
			BCC:              bcc,
			ReplyTo:          nil,
			Subject:          *request.Subject,
			InReplyTo:        "",
			References:       nil,
			Priority:         priority,
			ReplyToMessageID: nil,
			ForwardMessageID: nil,
			Attachments:      nil,
			Content:          content,
		},
		ExpectedRevision: expectedRevision,
	}, nil
}

func mailboxDraftSaveHTTPRejectIdempotencyKey(ctx fiber.Ctx) error {
	if ctx == nil || ctx.Request().Header.Peek("Idempotency-Key") != nil {
		return mailboxDraftCreateHTTPInvalidRequest()
	}
	return nil
}

func mailboxDraftSaveHTTPExpectedRevision(raw string) (int64, error) {
	if !mailboxDraftSaveHTTPIfMatchPattern.MatchString(raw) {
		return 0, mailboxDraftCreateHTTPInvalidRequest()
	}
	revision, err := strconv.ParseInt(raw[1:len(raw)-1], 10, 64)
	if err != nil || revision < 1 {
		return 0, mailboxDraftCreateHTTPInvalidRequest()
	}
	return revision, nil
}

func mailboxDraftCreateHTTPCanonicalUUID(raw string) (uuid.UUID, error) {
	value, err := uuid.Parse(raw)
	if err != nil || value == uuid.Nil || raw != value.String() {
		return uuid.Nil, mailboxDraftCreateHTTPInvalidRequest()
	}
	return value, nil
}

func mailboxDraftCreateHTTPOptionalUUID(raw *string) (*uuid.UUID, error) {
	if raw == nil {
		return nil, nil
	}
	value, err := mailboxDraftCreateHTTPCanonicalUUID(*raw)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func mailboxDraftCreateHTTPAddresses(request []mailboxDraftCreateHTTPAddress) ([]EngineAddress, error) {
	if len(request) > MaxEngineDraftRecipientItems {
		return nil, mailboxDraftCreateHTTPInvalidRequest()
	}
	result := make([]EngineAddress, len(request))
	for index, address := range request {
		result[index] = EngineAddress{Email: address.Email}
		if address.Name != nil {
			result[index].Name = *address.Name
		}
	}
	return result, nil
}

func mailboxDraftCreateHTTPAttachments(request []string) ([]MailboxDraftAttachmentReference, error) {
	if len(request) > MaxEngineAttachmentItems {
		return nil, mailboxDraftCreateHTTPInvalidRequest()
	}
	result := make([]MailboxDraftAttachmentReference, len(request))
	for index, rawAttachmentID := range request {
		attachmentID, err := mailboxDraftCreateHTTPCanonicalUUID(rawAttachmentID)
		if err != nil {
			return nil, err
		}
		result[index] = MailboxDraftAttachmentReference{AttachmentID: attachmentID}
	}
	return result, nil
}

func mailboxDraftCreateHTTPContent(bodyFormat string, richBody, plainBody mailboxDraftCreateHTTPBodyField) (MailboxContentInput, error) {
	switch MailboxContentFormat(bodyFormat) {
	case MailboxContentFormatPlain:
		if !plainBody.present || richBody.present {
			return MailboxContentInput{}, mailboxDraftCreateHTTPInvalidRequest()
		}
		return MailboxContentInput{SourceFormat: MailboxContentFormatPlain, Body: plainBody.value}, nil
	case MailboxContentFormatRich:
		if !richBody.present || plainBody.present {
			return MailboxContentInput{}, mailboxDraftCreateHTTPInvalidRequest()
		}
		return MailboxContentInput{SourceFormat: MailboxContentFormatRich, Body: richBody.value}, nil
	default:
		return MailboxContentInput{}, mailboxDraftCreateHTTPInvalidRequest()
	}
}

func mailboxDraftCreateHTTPPriority(raw string) (EngineMessagePriority, error) {
	priority := EngineMessagePriority(raw)
	switch priority {
	case EnginePriorityLow, EnginePriorityNormal, EnginePriorityHigh:
		return priority, nil
	default:
		return "", mailboxDraftCreateHTTPInvalidRequest()
	}
}

func mailboxDraftCreateHTTPResponseFromView(view MailboxDraftView) (mailboxDraftCreateHTTPResponse, bool) {
	if view.ID == uuid.Nil || view.DraftClientID == uuid.Nil || view.Revision < 1 || view.Status != MailboxDraftStatusActive || view.Effective.FromIdentityID == uuid.Nil || view.UpdatedAt.IsZero() {
		return mailboxDraftCreateHTTPResponse{}, false
	}
	to, ok := mailboxDraftCreateHTTPResponseAddresses(view.Content.To)
	if !ok {
		return mailboxDraftCreateHTTPResponse{}, false
	}
	cc, ok := mailboxDraftCreateHTTPResponseAddresses(view.Content.CC)
	if !ok {
		return mailboxDraftCreateHTTPResponse{}, false
	}
	bcc, ok := mailboxDraftCreateHTTPResponseAddresses(view.Content.BCC)
	if !ok {
		return mailboxDraftCreateHTTPResponse{}, false
	}
	attachments, ok := mailboxDraftCreateHTTPResponseAttachments(view.Effective.AttachmentIDs)
	if !ok {
		return mailboxDraftCreateHTTPResponse{}, false
	}
	bodyFormat, richBody, ok := mailboxDraftCreateHTTPResponseContent(view.Content.Content)
	if !ok {
		return mailboxDraftCreateHTTPResponse{}, false
	}
	if _, err := mailboxDraftCreateHTTPPriority(string(view.Content.Priority)); err != nil {
		return mailboxDraftCreateHTTPResponse{}, false
	}
	replyToMessageID, ok := mailboxDraftCreateHTTPResponseOptionalUUID(view.Effective.ReplyToMessageID)
	if !ok {
		return mailboxDraftCreateHTTPResponse{}, false
	}
	forwardMessageID, ok := mailboxDraftCreateHTTPResponseOptionalUUID(view.Effective.ForwardMessageID)
	if !ok {
		return mailboxDraftCreateHTTPResponse{}, false
	}
	return mailboxDraftCreateHTTPResponse{
		ID:               view.ID.String(),
		DraftClientID:    view.DraftClientID.String(),
		Revision:         view.Revision,
		Status:           string(view.Status),
		FromIdentityID:   view.Effective.FromIdentityID.String(),
		To:               to,
		CC:               cc,
		BCC:              bcc,
		Subject:          view.Content.Subject,
		BodyFormat:       bodyFormat,
		RichBody:         richBody,
		PlainBody:        view.Content.Content.PlainBody,
		Priority:         string(view.Content.Priority),
		AttachmentIDs:    attachments,
		ReplyToMessageID: replyToMessageID,
		ForwardMessageID: forwardMessageID,
		UpdatedAt:        view.UpdatedAt,
	}, true
}

func mailboxDraftCreateHTTPResponseAddresses(addresses []EngineAddress) ([]mailboxDraftCreateHTTPAddress, bool) {
	if len(addresses) > MaxEngineDraftRecipientItems {
		return nil, false
	}
	result := make([]mailboxDraftCreateHTTPAddress, len(addresses))
	for index, address := range addresses {
		if address.Email == "" {
			return nil, false
		}
		result[index].Email = address.Email
		if address.Name != "" {
			name := address.Name
			result[index].Name = &name
		}
	}
	return result, true
}

func mailboxDraftCreateHTTPResponseAttachments(attachmentIDs []uuid.UUID) ([]string, bool) {
	if len(attachmentIDs) > MaxEngineAttachmentItems {
		return nil, false
	}
	result := make([]string, len(attachmentIDs))
	seen := make(map[uuid.UUID]struct{}, len(attachmentIDs))
	for index, attachmentID := range attachmentIDs {
		if attachmentID == uuid.Nil {
			return nil, false
		}
		if _, exists := seen[attachmentID]; exists {
			return nil, false
		}
		seen[attachmentID] = struct{}{}
		result[index] = attachmentID.String()
	}
	return result, true
}

func mailboxDraftCreateHTTPResponseContent(content MailboxContentResult) (string, *string, bool) {
	switch content.SourceFormat {
	case MailboxContentFormatPlain:
		if content.SanitizedHTML != "" {
			return "", nil, false
		}
		return "plain", nil, true
	case MailboxContentFormatRich:
		richBody := content.SanitizedHTML
		return "rich", &richBody, true
	default:
		return "", nil, false
	}
}

func mailboxDraftCreateHTTPResponseOptionalUUID(value *uuid.UUID) (*string, bool) {
	if value == nil {
		return nil, true
	}
	if *value == uuid.Nil {
		return nil, false
	}
	encoded := value.String()
	return &encoded, true
}

func mailboxDraftCreateHTTPServiceError(ctx fiber.Ctx, err error) error {
	var serviceErr *MailboxDraftServiceError
	if !errors.As(err, &serviceErr) || serviceErr == nil {
		return mailboxDraftCreateHTTPUnavailable()
	}
	switch serviceErr.Kind() {
	case MailboxDraftServiceErrorInvalid:
		return mailboxDraftCreateHTTPInvalidRequest()
	case MailboxDraftServiceErrorNotFound:
		return fiber.NewError(http.StatusNotFound, "mail draft dependency unavailable")
	case MailboxDraftServiceErrorDeleting:
		ctx.Set("X-Hub-Error-Code", "MAIL_ACCOUNT_DELETING")
		return fiber.NewError(http.StatusConflict, "mail account is deleting")
	case MailboxDraftServiceErrorIdempotencyConflict:
		ctx.Set("X-Hub-Error-Code", "IDEMPOTENCY_KEY_REUSED")
		return fiber.NewError(http.StatusConflict, "mail draft idempotency conflict")
	case MailboxDraftServiceErrorPayloadTooLarge, MailboxDraftServiceErrorAttachmentLimit:
		return fiber.NewError(http.StatusRequestEntityTooLarge, "mail draft request exceeds limits")
	case MailboxDraftServiceErrorUnprocessable:
		return fiber.NewError(http.StatusUnprocessableEntity, "mail draft request cannot be processed")
	case MailboxDraftServiceErrorBusy:
		return mailboxDraftCreateHTTPBusy(ctx, serviceErr)
	default:
		return mailboxDraftCreateHTTPUnavailable()
	}
}

func mailboxDraftReadHTTPServiceError(ctx fiber.Ctx, err error) error {
	var serviceErr *MailboxDraftServiceError
	if !errors.As(err, &serviceErr) || serviceErr == nil {
		return mailboxDraftCreateHTTPUnavailable()
	}
	switch serviceErr.Kind() {
	case MailboxDraftServiceErrorNotFound:
		return fiber.NewError(http.StatusNotFound, "mail draft unavailable")
	case MailboxDraftServiceErrorDeleting:
		ctx.Set("X-Hub-Error-Code", "MAIL_ACCOUNT_DELETING")
		return fiber.NewError(http.StatusConflict, "mail account is deleting")
	case MailboxDraftServiceErrorConflict, MailboxDraftServiceErrorReconciliationRequired:
		ctx.Set("X-Hub-Error-Code", "MAIL_DRAFT_RECONCILIATION_REQUIRED")
		return fiber.NewError(http.StatusConflict, "mail draft reconciliation required")
	default:
		return mailboxDraftCreateHTTPUnavailable()
	}
}

func mailboxDraftSaveHTTPServiceError(ctx fiber.Ctx, err error) error {
	var serviceErr *MailboxDraftServiceError
	if !errors.As(err, &serviceErr) || serviceErr == nil {
		return mailboxDraftCreateHTTPUnavailable()
	}
	switch serviceErr.Kind() {
	case MailboxDraftServiceErrorInvalid:
		return mailboxDraftCreateHTTPInvalidRequest()
	case MailboxDraftServiceErrorNotFound:
		return fiber.NewError(http.StatusNotFound, "mail draft unavailable")
	case MailboxDraftServiceErrorDeleting:
		ctx.Set("X-Hub-Error-Code", "MAIL_ACCOUNT_DELETING")
		return fiber.NewError(http.StatusConflict, "mail account is deleting")
	case MailboxDraftServiceErrorPreconditionFailed:
		ctx.Set("X-Hub-Error-Code", "MAIL_DRAFT_VERSION_CONFLICT")
		return fiber.NewError(http.StatusConflict, "mail draft version conflict")
	case MailboxDraftServiceErrorConflict, MailboxDraftServiceErrorReconciliationRequired:
		ctx.Set("X-Hub-Error-Code", "MAIL_DRAFT_RECONCILIATION_REQUIRED")
		return fiber.NewError(http.StatusConflict, "mail draft reconciliation required")
	case MailboxDraftServiceErrorPayloadTooLarge, MailboxDraftServiceErrorAttachmentLimit:
		return fiber.NewError(http.StatusRequestEntityTooLarge, "mail draft request exceeds limits")
	case MailboxDraftServiceErrorAttachmentConflict, MailboxDraftServiceErrorUnprocessable:
		return fiber.NewError(http.StatusUnprocessableEntity, "mail draft request cannot be processed")
	case MailboxDraftServiceErrorBusy:
		return mailboxDraftCreateHTTPBusy(ctx, serviceErr)
	default:
		return mailboxDraftCreateHTTPUnavailable()
	}
}

func mailboxDraftCreateHTTPBusy(ctx fiber.Ctx, serviceErr *MailboxDraftServiceError) error {
	if ctx == nil || serviceErr == nil {
		return mailboxDraftCreateHTTPUnavailable()
	}
	busyUntil := serviceErr.BusyUntil()
	if busyUntil == nil {
		return mailboxDraftCreateHTTPUnavailable()
	}
	remaining := busyUntil.Sub(time.Now())
	if remaining <= 0 {
		return mailboxDraftCreateHTTPUnavailable()
	}
	retryAfterSeconds := int(math.Ceil(remaining.Seconds()))
	if retryAfterSeconds < 1 {
		retryAfterSeconds = 1
	}
	if retryAfterSeconds > 5 {
		retryAfterSeconds = 5
	}
	ctx.Set("X-Hub-Error-Code", "MAIL_ACCOUNT_BUSY")
	ctx.Set("Retry-After", strconv.Itoa(retryAfterSeconds))
	return fiber.NewError(http.StatusLocked, "mailbox draft is temporarily busy")
}

func mailboxDraftCreateHTTPInvalidRequest() error {
	return fiber.NewError(http.StatusBadRequest, "invalid mail draft request")
}

func mailboxDraftCreateHTTPUnavailable() error {
	return fiber.NewError(http.StatusServiceUnavailable, "mail draft unavailable")
}
