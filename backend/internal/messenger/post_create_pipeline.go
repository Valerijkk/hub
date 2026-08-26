package messenger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type TrustedMessageProducer string

const (
	TrustedMessageProducerUser              TrustedMessageProducer = "user"
	TrustedMessageProducerForward           TrustedMessageProducer = "forward"
	TrustedMessageProducerSystemBot         TrustedMessageProducer = "system_bot"
	TrustedMessageProducerPoll              TrustedMessageProducer = "poll"
	TrustedMessageProducerPromotionBacklink TrustedMessageProducer = "promotion_backlink"
)

var (
	ErrTrustedMessageProducerDenied = errors.New("trusted message producer denied")
	ErrPostCreatePipelineInvalid    = errors.New("post-create pipeline request invalid")
)

type AuthorizedMessageProducerScope struct {
	TenantID     uuid.UUID
	ActorID      uuid.UUID
	ChannelID    uuid.UUID
	Producer     TrustedMessageProducer
	BotKind      BotKind
	IntegrationID uuid.UUID
}

type PostCreatePipelineRequest struct {
	Producer TrustedMessageProducer
	Message  MessageDocument
	Parent   *MessageDocument
}

type Phase151PostCreatePipelineRepository interface {
	FindCanonicalCreateResult(context.Context, AuthorizedMessageProducerScope, string) (CreateMessageResult, bool, error)
	ResolveEligibleMentionUserIDs(context.Context, AuthorizedMessageProducerScope, PostCreatePipelineRequest) ([]uuid.UUID, error)
	PersistPendingCanonicalMessage(context.Context, AuthorizedMessageProducerScope, *MessageDocument) (CreateMessageResult, error)
	EnsureAutoThreadFollows(context.Context, AuthorizedMessageProducerScope, *MessageDocument) error
	ReconcileReplyCount(context.Context, AuthorizedMessageProducerScope, *MessageDocument) error
	CompletePostCreateEffects(context.Context, AuthorizedMessageProducerScope, string) (bool, error)
	PublishUnreadInvalidation(context.Context, AuthorizedMessageProducerScope, *MessageDocument) error
}

func ExportPhase151ValidateTrustedMessageProducer(
	scope AuthorizedMessageProducerScope,
	producer TrustedMessageProducer,
) error {
	if scope.TenantID == uuid.Nil || scope.ActorID == uuid.Nil || scope.ChannelID == uuid.Nil ||
		scope.Producer != producer || !isKnownTrustedMessageProducer(producer) {
		return ErrTrustedMessageProducerDenied
	}

	if producer == TrustedMessageProducerSystemBot {
		if scope.BotKind != BotKindSystem || scope.IntegrationID == uuid.Nil {
			return ErrTrustedMessageProducerDenied
		}
		return nil
	}
	if scope.BotKind != "" || scope.IntegrationID != uuid.Nil {
		return ErrTrustedMessageProducerDenied
	}
	return nil
}

func ExportPhase151ValidateAuthorizedBotProducer(
	expected AuthorizedMessageProducerScope,
	actual AuthorizedMessageProducerScope,
) error {
	if err := ExportPhase151ValidateTrustedMessageProducer(expected, TrustedMessageProducerSystemBot); err != nil {
		return err
	}
	if err := ExportPhase151ValidateTrustedMessageProducer(actual, TrustedMessageProducerSystemBot); err != nil {
		return err
	}
	if expected.TenantID != actual.TenantID ||
		expected.ActorID != actual.ActorID ||
		expected.ChannelID != actual.ChannelID ||
		expected.Producer != actual.Producer ||
		expected.BotKind != actual.BotKind ||
		expected.IntegrationID != actual.IntegrationID {
		return ErrTrustedMessageProducerDenied
	}
	return nil
}

// ExportPhase151ExecutePostCreatePipeline persists one canonical pending
// message before any derived effect. A Created=false canonical lookup or
// insert race returns immediately and never repeats effects.
func ExportPhase151ExecutePostCreatePipeline(
	ctx context.Context,
	scope AuthorizedMessageProducerScope,
	request PostCreatePipelineRequest,
	repository Phase151PostCreatePipelineRepository,
) (CreateMessageResult, error) {
	if repository == nil {
		return CreateMessageResult{}, ErrPostCreatePipelineInvalid
	}
	if err := validatePostCreatePipelineRequest(scope, request); err != nil {
		return CreateMessageResult{}, err
	}

	canonical, found, err := repository.FindCanonicalCreateResult(ctx, scope, request.Message.ClientNonce)
	if err != nil {
		return CreateMessageResult{}, fmt.Errorf("find canonical message: %w", err)
	}
	if found {
		if err := validatePostCreateCanonicalResult(scope, request, canonical, true); err != nil {
			return CreateMessageResult{}, err
		}
		return canonical, nil
	}

	resolvedMentions, err := repository.ResolveEligibleMentionUserIDs(ctx, scope, request)
	if err != nil {
		return CreateMessageResult{}, fmt.Errorf("resolve eligible mentions: %w", err)
	}
	document, err := ExportPhase151PreparePendingCanonicalMessage(scope, request, resolvedMentions, time.Now())
	if err != nil {
		return CreateMessageResult{}, err
	}

	result, err := repository.PersistPendingCanonicalMessage(ctx, scope, document)
	if err != nil {
		return CreateMessageResult{}, fmt.Errorf("persist pending canonical message: %w", err)
	}
	if err := validatePostCreateCanonicalResult(scope, request, result, false); err != nil {
		return CreateMessageResult{}, err
	}
	if !result.Created {
		return result, nil
	}
	if !ExportPhase151ShouldClaimPostCreateEffects(result) {
		return result, nil
	}
	if err := repository.EnsureAutoThreadFollows(ctx, scope, result.Message); err != nil {
		return result, fmt.Errorf("ensure automatic thread follows: %w", err)
	}
	if request.Message.ThreadID != "" {
		if err := repository.ReconcileReplyCount(ctx, scope, result.Message); err != nil {
			return result, fmt.Errorf("reconcile reply count: %w", err)
		}
	}
	completed, err := repository.CompletePostCreateEffects(ctx, scope, result.Message.ID)
	if err != nil {
		return result, fmt.Errorf("complete post-create effects: %w", err)
	}
	if !completed {
		return result, nil
	}
	if err := repository.PublishUnreadInvalidation(ctx, scope, result.Message); err != nil {
		return result, fmt.Errorf("publish unread invalidation: %w", err)
	}
	return result, nil
}

func validatePostCreateCanonicalResult(
	scope AuthorizedMessageProducerScope,
	request PostCreatePipelineRequest,
	result CreateMessageResult,
	foundByLookup bool,
) error {
	if foundByLookup && result.Created {
		return ErrPostCreatePipelineInvalid
	}
	if result.Message == nil || result.Message.ID == "" ||
		result.Message.ChannelID != scope.ChannelID ||
		result.Message.SenderID != scope.ActorID ||
		result.Message.ClientNonce != request.Message.ClientNonce {
		return ErrPostCreatePipelineInvalid
	}

	expectedFingerprint, err := buildPostCreateClientNonceFingerprint(request)
	if err != nil ||
		result.Message.ClientNonceFingerprintVersion != clientNonceFingerprintVersionV1 ||
		result.Message.ClientNonceFingerprint != expectedFingerprint {
		return ErrPostCreatePipelineInvalid
	}

	messageState := result.Message.PostCreateEffectsState
	if messageState == "" {
		messageState = PostCreateEffectsStateComplete
	}
	if messageState != PostCreateEffectsStatePending && messageState != PostCreateEffectsStateComplete {
		return ErrPostCreatePipelineInvalid
	}
	if result.PostCreateEffectsState != messageState {
		return ErrPostCreatePipelineInvalid
	}
	return nil
}

func ExportPhase151PreparePendingCanonicalMessage(
	scope AuthorizedMessageProducerScope,
	request PostCreatePipelineRequest,
	resolvedMentions []uuid.UUID,
	now time.Time,
) (*MessageDocument, error) {
	if err := validatePostCreatePipelineRequest(scope, request); err != nil {
		return nil, err
	}
	if now.IsZero() {
		return nil, ErrPostCreatePipelineInvalid
	}

	contentJSON, err := clonePostCreateContentJSON(request.Message.ContentJSON)
	if err != nil {
		return nil, fmt.Errorf("%w: content_json", ErrPostCreatePipelineInvalid)
	}
	fingerprint, err := buildPostCreateClientNonceFingerprint(request)
	if err != nil {
		return nil, fmt.Errorf("%w: client_nonce_fingerprint", ErrPostCreatePipelineInvalid)
	}
	document := request.Message
	document.ContentJSON = contentJSON
	document.Attachments = append([]AttachmentRef(nil), request.Message.Attachments...)
	document.Mentions = append([]string(nil), request.Message.Mentions...)
	document.MentionGroups = append([]string(nil), request.Message.MentionGroups...)
	document.ResolvedMentionUserIDs = append([]uuid.UUID(nil), resolvedMentions...)
	document.ClientNonceFingerprint = fingerprint
	document.ClientNonceFingerprintVersion = clientNonceFingerprintVersionV1
	document.PostCreateEffectsState = PostCreateEffectsStatePending
	document.CreatedAt = now.UTC()
	return &document, nil
}

func clonePostCreateContentJSON(content map[string]interface{}) (map[string]interface{}, error) {
	if content == nil {
		return nil, nil
	}
	encoded, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}
	cloned := make(map[string]interface{}, len(content))
	if err := json.Unmarshal(encoded, &cloned); err != nil {
		return nil, err
	}
	return cloned, nil
}

func validatePostCreatePipelineRequest(
	scope AuthorizedMessageProducerScope,
	request PostCreatePipelineRequest,
) error {
	if err := ExportPhase151ValidateTrustedMessageProducer(scope, request.Producer); err != nil {
		return err
	}
	message := request.Message
	if _, err := json.Marshal(message.ContentJSON); err != nil {
		return ErrPostCreatePipelineInvalid
	}
	if message.ChannelID == uuid.Nil || message.SenderID == uuid.Nil ||
		message.ChannelID != scope.ChannelID || message.SenderID != scope.ActorID ||
		message.ClientNonce == "" {
		return ErrPostCreatePipelineInvalid
	}
	if message.ID != "" || message.ClientNonceFingerprint != "" ||
		message.ClientNonceFingerprintVersion != 0 || message.ReplyCount != 0 || len(message.Reactions) != 0 ||
		len(message.ResolvedMentionUserIDs) != 0 || message.PostCreateEffectsState != "" ||
		message.PostCreateEffectsCompletedAt != nil || message.EditedAt != nil ||
		message.DeletedAt != nil || !message.CreatedAt.IsZero() {
		return ErrPostCreatePipelineInvalid
	}
	if message.ThreadID == "" {
		if request.Parent != nil || message.AlsoSendToChannel {
			return ErrPostCreatePipelineInvalid
		}
		return nil
	}
	if request.Parent == nil || request.Parent.ID != message.ThreadID ||
		request.Parent.ChannelID != message.ChannelID || request.Parent.ThreadID != "" ||
		request.Parent.DeletedAt != nil {
		return ErrPostCreatePipelineInvalid
	}
	return nil
}

// ExportPhase151TrustedMessageProducers returns a copy of the closed producer
// allowlist used by the canonical post-create pipeline.
func ExportPhase151TrustedMessageProducers() []TrustedMessageProducer {
	return []TrustedMessageProducer{
		TrustedMessageProducerUser,
		TrustedMessageProducerForward,
		TrustedMessageProducerSystemBot,
		TrustedMessageProducerPoll,
		TrustedMessageProducerPromotionBacklink,
	}
}

// ExportPhase151TrustedProducerForBotScope maps only a current, enabled,
// project-eligible Hub system-bot send proof into the closed pipeline enum.
func ExportPhase151TrustedProducerForBotScope(
	scope *AuthorizedBotChannelScope,
) (TrustedMessageProducer, bool) {
	if scope == nil ||
		scope.TenantID == uuid.Nil ||
		scope.ActorID == uuid.Nil ||
		scope.ChannelID == uuid.Nil ||
		scope.ProjectID == uuid.Nil ||
		scope.IntegrationID == uuid.Nil ||
		scope.ActorKind != BotActorKindSystem ||
		scope.BotKind != BotKindHubAssistant ||
		!scope.BotEnabled ||
		!scope.ProjectEligible ||
		scope.Action != BotChannelAction(BotMongoPathSend) {
		return "", false
	}
	return TrustedMessageProducerSystemBot, true
}
func isKnownTrustedMessageProducer(producer TrustedMessageProducer) bool {
	switch producer {
	case TrustedMessageProducerUser,
		TrustedMessageProducerForward,
		TrustedMessageProducerSystemBot,
		TrustedMessageProducerPoll,
		TrustedMessageProducerPromotionBacklink:
		return true
	default:
		return false
	}
}
