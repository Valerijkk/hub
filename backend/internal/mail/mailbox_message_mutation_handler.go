package mail

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

const mailboxMessageMutationHTTPJSONBodyMaxBytes = 16 << 10

// MailboxMessageMutationHandler exposes the single authenticated, receipt-backed
// bulk mutation surface. It owns HTTP validation and response mapping only.
type MailboxMessageMutationHandler struct {
	service MailboxMessageMutationServiceAPI
}

// NewMailboxMessageMutationHandler creates the isolated public mutation handler.
func NewMailboxMessageMutationHandler(service MailboxMessageMutationServiceAPI) (*MailboxMessageMutationHandler, error) {
	if mailboxHTTPDependencyNil(service) {
		return nil, mailboxMessageMutationError(MailboxMessageMutationErrorInvalid)
	}
	return &MailboxMessageMutationHandler{service: service}, nil
}

// RegisterRoutes registers only POST /api/mail/messages/bulk on the supplied /api router.
func (handler *MailboxMessageMutationHandler) RegisterRoutes(api fiber.Router, authMiddleware fiber.Handler) error {
	if handler == nil || mailboxHTTPDependencyNil(handler.service) || mailboxHTTPDependencyNil(api) || mailboxHTTPDependencyNil(authMiddleware) {
		return mailboxMessageMutationError(MailboxMessageMutationErrorInvalid)
	}
	api.Group("/mail", authMiddleware).Post("/messages/bulk", handler.mutate)
	return nil
}

type mailboxMessageMutationHTTPRequest struct {
	TargetType string          `json:"target_type"`
	TargetIDs  []string        `json:"target_ids"`
	Action     string          `json:"action"`
	FolderID   json.RawMessage `json:"folder_id"`
}

type mailboxMessageMutationHTTPResponse struct {
	Results []mailboxMessageMutationTargetHTTPResponse `json:"results"`
}

type mailboxMessageMutationTargetHTTPResponse struct {
	TargetType            string  `json:"target_type"`
	TargetID              string  `json:"target_id"`
	Status                string  `json:"status"`
	AffectedMessageCount  int     `json:"affected_message_count"`
	SucceededMessageCount int     `json:"succeeded_message_count"`
	FailedMessageCount    int     `json:"failed_message_count"`
	TrainingStatus        string  `json:"training_status"`
	Code                  *string `json:"code"`
}

func (handler *MailboxMessageMutationHandler) mutate(ctx fiber.Ctx) error {
	if err := mailboxMessageMutationHTTPRejectQuery(ctx); err != nil {
		return err
	}
	actorID, err := mailboxActorID(ctx)
	if err != nil {
		return err
	}
	idempotencyKey, err := mailboxMessageMutationHTTPIdempotencyKey(ctx)
	if err != nil {
		return err
	}
	var request mailboxMessageMutationHTTPRequest
	if err := mailboxMessageMutationHTTPDecodeJSON(ctx, &request); err != nil {
		return err
	}
	input, err := mailboxMessageMutationHTTPInput(request, idempotencyKey)
	if err != nil {
		return err
	}

	result, err := handler.service.MutateMailboxMessages(ctx.Context(), actorID, input)
	if err != nil {
		return mailboxMessageMutationHTTPServiceError(ctx, err)
	}
	response, valid := mailboxMessageMutationHTTPResponseFromResult(result, input)
	if !valid {
		return mailboxMessageMutationHTTPUnavailable()
	}
	switch result.State {
	case MailboxMessageMutationResultCompleted:
		return ctx.Status(http.StatusOK).JSON(response)
	case MailboxMessageMutationResultUnknown:
		// A validated unknown result is receipt-backed. Without that durable result,
		// the service must return an error and the transport reports 503 instead.
		return ctx.Status(http.StatusAccepted).JSON(response)
	default:
		return mailboxMessageMutationHTTPUnavailable()
	}
}

func mailboxMessageMutationHTTPRejectQuery(ctx fiber.Ctx) error {
	if len(ctx.Request().URI().QueryString()) != 0 {
		return mailboxMessageMutationHTTPInvalidRequest()
	}
	return nil
}

func mailboxMessageMutationHTTPDecodeJSON(ctx fiber.Ctx, target any) error {
	mediaType, _, err := mime.ParseMediaType(strings.TrimSpace(ctx.Get(fiber.HeaderContentType)))
	if err != nil || !strings.EqualFold(mediaType, "application/json") {
		return fiber.NewError(http.StatusUnsupportedMediaType, "unsupported media type")
	}
	rawBody := ctx.BodyRaw()
	if len(rawBody) > mailboxMessageMutationHTTPJSONBodyMaxBytes {
		return fiber.NewError(http.StatusRequestEntityTooLarge, "request body too large")
	}
	body := bytes.TrimSpace(rawBody)
	if len(body) == 0 || body[0] != '{' {
		return mailboxMessageMutationHTTPInvalidRequest()
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return mailboxMessageMutationHTTPInvalidRequest()
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return mailboxMessageMutationHTTPInvalidRequest()
	}
	return nil
}

func mailboxMessageMutationHTTPIdempotencyKey(ctx fiber.Ctx) (uuid.UUID, error) {
	key, err := mailboxMessageMutationHTTPCanonicalUUID(ctx.Get("Idempotency-Key"))
	if err != nil || key.Version() != 4 && key.Version() != 7 {
		return uuid.Nil, mailboxMessageMutationHTTPInvalidRequest()
	}
	return key, nil
}

func mailboxMessageMutationHTTPInput(request mailboxMessageMutationHTTPRequest, idempotencyKey uuid.UUID) (MailboxMessageMutationInput, error) {
	if idempotencyKey == uuid.Nil || len(request.TargetIDs) < 1 || len(request.TargetIDs) > MailboxMessageMutationMaximumResolvedMessages {
		return MailboxMessageMutationInput{}, mailboxMessageMutationHTTPInvalidRequest()
	}
	input := MailboxMessageMutationInput{
		IdempotencyKey: idempotencyKey,
		TargetType:     MailboxMessageMutationTargetType(request.TargetType),
		Action:         MailboxMessageMutationAction(request.Action),
		TargetIDs:      make([]uuid.UUID, 0, len(request.TargetIDs)),
	}
	if !validMailboxMessageMutationTargetType(input.TargetType) || !validMailboxMessageMutationAction(input.Action) {
		return MailboxMessageMutationInput{}, mailboxMessageMutationHTTPInvalidRequest()
	}
	seen := make(map[uuid.UUID]struct{}, len(request.TargetIDs))
	for _, rawTargetID := range request.TargetIDs {
		targetID, err := mailboxMessageMutationHTTPCanonicalUUID(rawTargetID)
		if err != nil {
			return MailboxMessageMutationInput{}, mailboxMessageMutationHTTPInvalidRequest()
		}
		if _, exists := seen[targetID]; exists {
			return MailboxMessageMutationInput{}, mailboxMessageMutationHTTPInvalidRequest()
		}
		seen[targetID] = struct{}{}
		input.TargetIDs = append(input.TargetIDs, targetID)
	}
	if input.Action != MailboxMessageMutationActionMove {
		if len(request.FolderID) != 0 {
			return MailboxMessageMutationInput{}, mailboxMessageMutationHTTPInvalidRequest()
		}
		return input, nil
	}
	if len(request.FolderID) == 0 {
		return MailboxMessageMutationInput{}, mailboxMessageMutationHTTPInvalidRequest()
	}
	var rawFolderID string
	if err := json.Unmarshal(request.FolderID, &rawFolderID); err != nil {
		return MailboxMessageMutationInput{}, mailboxMessageMutationHTTPInvalidRequest()
	}
	folderID, err := mailboxMessageMutationHTTPCanonicalUUID(rawFolderID)
	if err != nil {
		return MailboxMessageMutationInput{}, mailboxMessageMutationHTTPInvalidRequest()
	}
	input.FolderID = &folderID
	return input, nil
}

func mailboxMessageMutationHTTPCanonicalUUID(raw string) (uuid.UUID, error) {
	value, err := uuid.Parse(raw)
	if err != nil || value == uuid.Nil || raw != value.String() {
		return uuid.Nil, mailboxMessageMutationHTTPInvalidRequest()
	}
	return value, nil
}

func mailboxMessageMutationHTTPResponseFromResult(result MailboxMessageMutationResult, input MailboxMessageMutationInput) (mailboxMessageMutationHTTPResponse, bool) {
	if !mailboxMessageMutationResultValid(result, input) {
		return mailboxMessageMutationHTTPResponse{}, false
	}
	response := mailboxMessageMutationHTTPResponse{Results: make([]mailboxMessageMutationTargetHTTPResponse, len(result.Results))}
	for index, target := range result.Results {
		responseTarget := mailboxMessageMutationTargetHTTPResponse{
			TargetType:            string(target.TargetType),
			TargetID:              target.TargetID.String(),
			Status:                string(target.Status),
			AffectedMessageCount:  target.AffectedMessageCount,
			SucceededMessageCount: target.SucceededMessageCount,
			FailedMessageCount:    target.FailedMessageCount,
			TrainingStatus:        string(target.TrainingStatus),
		}
		if target.Code != nil {
			code := string(*target.Code)
			responseTarget.Code = &code
		}
		response.Results[index] = responseTarget
	}
	return response, true
}

func mailboxMessageMutationHTTPServiceError(ctx fiber.Ctx, err error) error {
	kind, ok := MailboxMessageMutationErrorKindOf(err)
	if !ok {
		return mailboxMessageMutationHTTPUnavailable()
	}
	switch kind {
	case MailboxMessageMutationErrorInvalid:
		return mailboxMessageMutationHTTPInvalidRequest()
	case MailboxMessageMutationErrorNotFound:
		return fiber.NewError(http.StatusNotFound, "mail message target not found")
	case MailboxMessageMutationErrorDeleting:
		ctx.Set("X-Hub-Error-Code", "MAIL_ACCOUNT_DELETING")
		return fiber.NewError(http.StatusConflict, "mailbox account deleting")
	case MailboxMessageMutationErrorIdempotencyConflict:
		ctx.Set("X-Hub-Error-Code", "IDEMPOTENCY_KEY_REUSED")
		return fiber.NewError(http.StatusConflict, "mail mutation idempotency conflict")
	case MailboxMessageMutationErrorExpansionLimit:
		ctx.Set("X-Hub-Error-Code", "MAIL_BULK_EXPANSION_LIMIT")
		return fiber.NewError(http.StatusUnprocessableEntity, "mail mutation expansion limit exceeded")
	case MailboxMessageMutationErrorBusy:
		ctx.Set("X-Hub-Error-Code", "MAIL_ACCOUNT_BUSY")
		ctx.Set("Retry-After", "1")
		return fiber.NewError(http.StatusLocked, "mailbox mutation is temporarily busy")
	case MailboxMessageMutationErrorConflict,
		MailboxMessageMutationErrorOutcomeUnknown,
		MailboxMessageMutationErrorUnavailable:
		return mailboxMessageMutationHTTPUnavailable()
	default:
		return mailboxMessageMutationHTTPUnavailable()
	}
}

func mailboxMessageMutationHTTPInvalidRequest() error {
	return fiber.NewError(http.StatusBadRequest, "invalid mail message mutation request")
}

func mailboxMessageMutationHTTPUnavailable() error {
	return fiber.NewError(http.StatusServiceUnavailable, "mail message mutation unavailable")
}
