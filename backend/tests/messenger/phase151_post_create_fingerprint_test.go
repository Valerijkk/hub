package messenger_test

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/messenger"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestPhase151PreparedCanonicalPersistsVersionedNonceFingerprint(t *testing.T) {
	scope, request := phase151PostCreateFingerprintFixture()
	first := phase151PrepareFingerprintDocument(
		t, scope, request, time.Date(2026, time.August, 27, 2, 0, 0, 0, time.UTC),
	)
	firstFields := phase151MessageBSONFields(t, *first)
	fingerprint, _ := firstFields["client_nonce_fingerprint"].(string)
	require.Len(t, fingerprint, 64)
	_, err := hex.DecodeString(fingerprint)
	require.NoError(t, err)
	assert.EqualValues(t, 1, firstFields["client_nonce_fingerprint_version"])

	equivalent := request
	equivalent.Message.ContentJSON = map[string]interface{}{
		"nested": map[string]interface{}{"a": "one", "b": "two"},
		"count": json.Number("1"),
		"type": "text",
	}
	equivalent.Message.Mentions = []string{"user-2", "user-1"}
	equivalent.Message.MentionGroups = []string{"online", "all"}
	second := phase151PrepareFingerprintDocument(
		t, scope, equivalent, time.Date(2026, time.August, 27, 3, 0, 0, 0, time.UTC),
	)
	secondFields := phase151MessageBSONFields(t, *second)
	assert.Equal(t, fingerprint, secondFields["client_nonce_fingerprint"],
		"JSON representation, mention order/duplicates, and server time must not change create intent")

	withoutAttachments := request
	withoutAttachments.Message.Attachments = nil
	emptyAttachments := withoutAttachments
	emptyAttachments.Message.Attachments = []messenger.AttachmentRef{}
	nilDocument := phase151PrepareFingerprintDocument(t, scope, withoutAttachments, time.Now().UTC())
	emptyDocument := phase151PrepareFingerprintDocument(t, scope, emptyAttachments, time.Now().UTC().Add(time.Minute))
	assert.Equal(t,
		phase151MessageBSONFields(t, *nilDocument)["client_nonce_fingerprint"],
		phase151MessageBSONFields(t, *emptyDocument)["client_nonce_fingerprint"],
		"nil and empty attachment lists retain legacy compatibility",
	)
}

func TestPhase151PreparedCanonicalRejectsCallerNonceFingerprintMetadata(t *testing.T) {
	for _, scenario := range []struct {
		name      string
		overrides map[string]interface{}
	}{
		{name: "fingerprint only", overrides: map[string]interface{}{"client_nonce_fingerprint": "attacker"}},
		{name: "version only", overrides: map[string]interface{}{"client_nonce_fingerprint_version": int32(1)}},
		{name: "fingerprint and version", overrides: map[string]interface{}{
			"client_nonce_fingerprint": "attacker", "client_nonce_fingerprint_version": int32(1),
		}},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			scope, request := phase151PostCreateFingerprintFixture()
			request.Message = phase151MessageDocumentWithBSONOverrides(t, request.Message, scenario.overrides)

			document, err := messenger.ExportPhase151PreparePendingCanonicalMessage(
				scope, request, nil, time.Now().UTC(),
			)

			assert.Nil(t, document)
			require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
		})
	}
}

func TestPhase151RejectsInvalidCanonicalNonceFingerprintMetadataBeforeEffects(t *testing.T) {
	for _, scenario := range []struct {
		name      string
		overrides map[string]interface{}
	}{
		{name: "missing fingerprint", overrides: map[string]interface{}{"client_nonce_fingerprint": nil}},
		{name: "missing version", overrides: map[string]interface{}{"client_nonce_fingerprint_version": nil}},
		{name: "unknown version", overrides: map[string]interface{}{"client_nonce_fingerprint_version": int32(2)}},
		{name: "different fingerprint", overrides: map[string]interface{}{
			"client_nonce_fingerprint": "0000000000000000000000000000000000000000000000000000000000000000",
		}},
		{name: "malformed fingerprint", overrides: map[string]interface{}{"client_nonce_fingerprint": "not-sha256"}},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			scope, request := phase151PostCreateFingerprintFixture()
			canonical := phase151PostCreateCanonicalResult(
				t, scope, request, false, messenger.PostCreateEffectsStateComplete,
			)
			canonical.Message = phase151MessageDocumentPointer(
				phase151MessageDocumentWithBSONOverrides(t, *canonical.Message, scenario.overrides),
			)
			recorder := &phase151PostCreateResultRecorder{
				phase151PostCreateRecorder: newPhase151PostCreateRecorder(),
				findFound: true,
				findResult: canonical,
			}

			result, err := messenger.ExportPhase151ExecutePostCreatePipeline(
				context.Background(), scope, request, recorder,
			)

			require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
			assert.Equal(t, messenger.CreateMessageResult{}, result)
			assert.Equal(t, []string{"idempotency:" + request.Message.ClientNonce}, recorder.calls)
			phase151AssertPostCreateNoEffects(t, recorder.phase151PostCreateRecorder)
		})
	}
}

func TestPhase151RejectsNonceReuseForDifferentCreateIntentBeforeEffects(t *testing.T) {
	for _, scenario := range []struct {
		name   string
		mutate func(*messenger.AuthorizedMessageProducerScope, *messenger.PostCreatePipelineRequest)
	}{
		{name: "producer differs", mutate: func(scope *messenger.AuthorizedMessageProducerScope, request *messenger.PostCreatePipelineRequest) {
			scope.Producer = messenger.TrustedMessageProducerForward
			request.Producer = messenger.TrustedMessageProducerForward
		}},
		{name: "message type differs", mutate: func(_ *messenger.AuthorizedMessageProducerScope, request *messenger.PostCreatePipelineRequest) {
			request.Message.MessageType = "system"
		}},
		{name: "poll id differs", mutate: func(_ *messenger.AuthorizedMessageProducerScope, request *messenger.PostCreatePipelineRequest) {
			request.Message.PollID = "poll-changed"
		}},
		{name: "content json differs", mutate: func(_ *messenger.AuthorizedMessageProducerScope, request *messenger.PostCreatePipelineRequest) {
			request.Message.ContentJSON = map[string]interface{}{"type": "text", "count": float64(2)}
		}},
		{name: "content text differs", mutate: func(_ *messenger.AuthorizedMessageProducerScope, request *messenger.PostCreatePipelineRequest) {
			request.Message.ContentText = "changed"
		}},
		{name: "thread and root projection differ", mutate: func(scope *messenger.AuthorizedMessageProducerScope, request *messenger.PostCreatePipelineRequest) {
			request.Message.ThreadID = "different-root"
			request.Message.AlsoSendToChannel = true
			request.Parent = &messenger.MessageDocument{ID: "different-root", ChannelID: scope.ChannelID}
		}},
		{name: "forward source differs", mutate: func(_ *messenger.AuthorizedMessageProducerScope, request *messenger.PostCreatePipelineRequest) {
			request.Message.ForwardedFromID = "forwarded-changed"
		}},
		{name: "voice url differs", mutate: func(_ *messenger.AuthorizedMessageProducerScope, request *messenger.PostCreatePipelineRequest) {
			request.Message.VoiceURL = "https://hub.invalid/changed.ogg"
		}},
		{name: "voice duration differs", mutate: func(_ *messenger.AuthorizedMessageProducerScope, request *messenger.PostCreatePipelineRequest) {
			request.Message.VoiceDurationMs++
		}},
		{name: "attachment order differs", mutate: func(_ *messenger.AuthorizedMessageProducerScope, request *messenger.PostCreatePipelineRequest) {
			attachments := append([]messenger.AttachmentRef(nil), request.Message.Attachments...)
			attachments[0], attachments[1] = attachments[1], attachments[0]
			request.Message.Attachments = attachments
		}},
		{name: "attachment field differs", mutate: func(_ *messenger.AuthorizedMessageProducerScope, request *messenger.PostCreatePipelineRequest) {
			attachments := append([]messenger.AttachmentRef(nil), request.Message.Attachments...)
			attachments[0].FileSize++
			request.Message.Attachments = attachments
		}},
		{name: "mention set differs", mutate: func(_ *messenger.AuthorizedMessageProducerScope, request *messenger.PostCreatePipelineRequest) {
			request.Message.Mentions = []string{"user-1", "user-3"}
		}},
		{name: "mention group set differs", mutate: func(_ *messenger.AuthorizedMessageProducerScope, request *messenger.PostCreatePipelineRequest) {
			request.Message.MentionGroups = []string{"all", "offline"}
		}},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			canonicalScope, original := phase151PostCreateFingerprintFixture()
			canonical := phase151PostCreateCanonicalResult(
				t, canonicalScope, original, false, messenger.PostCreateEffectsStateComplete,
			)
			retryScope := canonicalScope
			retry := original
			scenario.mutate(&retryScope, &retry)
			recorder := &phase151PostCreateResultRecorder{
				phase151PostCreateRecorder: newPhase151PostCreateRecorder(),
				findFound: true,
				findResult: canonical,
			}

			result, err := messenger.ExportPhase151ExecutePostCreatePipeline(
				context.Background(), retryScope, retry, recorder,
			)

			require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
			assert.Equal(t, messenger.CreateMessageResult{}, result)
			assert.Equal(t, []string{"idempotency:" + retry.Message.ClientNonce}, recorder.calls)
			phase151AssertPostCreateNoEffects(t, recorder.phase151PostCreateRecorder)
		})
	}
}

func TestPhase151RejectsPersistedCanonicalNonceFingerprintMismatchBeforeEffects(t *testing.T) {
	scope, request := phase151PostCreateFingerprintFixture()
	canonical := phase151PostCreateCanonicalResult(
		t, scope, request, true, messenger.PostCreateEffectsStatePending,
	)
	canonical.Message = phase151MessageDocumentPointer(phase151MessageDocumentWithBSONOverrides(
		t,
		*canonical.Message,
		map[string]interface{}{
			"client_nonce_fingerprint": "0000000000000000000000000000000000000000000000000000000000000000",
		},
	))
	recorder := &phase151PostCreateResultRecorder{
		phase151PostCreateRecorder: newPhase151PostCreateRecorder(),
		persistResult: &canonical,
	}

	result, err := messenger.ExportPhase151ExecutePostCreatePipeline(
		context.Background(), scope, request, recorder,
	)

	require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
	assert.Equal(t, messenger.CreateMessageResult{}, result)
	assert.Equal(t, []string{
		"idempotency:" + request.Message.ClientNonce,
		"resolve-mentions",
		"persist-pending",
	}, recorder.calls)
	phase151AssertPostCreateNoEffects(t, recorder.phase151PostCreateRecorder)
}

func TestPhase151CompatibleNonceRetriesUseStoredCreateIntentFingerprint(t *testing.T) {
	t.Run("equivalent JSON and mention sets reuse canonical", func(t *testing.T) {
		scope, original := phase151PostCreateFingerprintFixture()
		canonical := phase151PostCreateCanonicalResult(
			t, scope, original, false, messenger.PostCreateEffectsStateComplete,
		)
		retry := original
		retry.Message.ContentJSON = map[string]interface{}{
			"count": json.Number("1"),
			"nested": map[string]interface{}{"b": "two", "a": "one"},
			"type": "text",
		}
		retry.Message.Mentions = []string{"user-2", "user-1"}
		retry.Message.MentionGroups = []string{"online", "all"}
		recorder := &phase151PostCreateResultRecorder{
			phase151PostCreateRecorder: newPhase151PostCreateRecorder(),
			findFound: true,
			findResult: canonical,
		}

		result, err := messenger.ExportPhase151ExecutePostCreatePipeline(
			context.Background(), scope, retry, recorder,
		)

		require.NoError(t, err)
		assert.Same(t, canonical.Message, result.Message)
		assert.Equal(t, []string{"idempotency:" + retry.Message.ClientNonce}, recorder.calls)
		phase151AssertPostCreateNoEffects(t, recorder.phase151PostCreateRecorder)
	})

	t.Run("later message edit does not rewrite original create intent", func(t *testing.T) {
		scope, request := phase151PostCreateFingerprintFixture()
		canonical := phase151PostCreateCanonicalResult(
			t, scope, request, false, messenger.PostCreateEffectsStateComplete,
		)
		editedAt := time.Date(2026, time.August, 27, 4, 0, 0, 0, time.UTC)
		canonical.Message.ContentText = "edited after create"
		canonical.Message.ContentJSON = map[string]interface{}{"type": "text", "body": "edited"}
		canonical.Message.Attachments = nil
		canonical.Message.EditedAt = &editedAt
		recorder := &phase151PostCreateResultRecorder{
			phase151PostCreateRecorder: newPhase151PostCreateRecorder(),
			findFound: true,
			findResult: canonical,
		}

		result, err := messenger.ExportPhase151ExecutePostCreatePipeline(
			context.Background(), scope, request, recorder,
		)

		require.NoError(t, err)
		assert.Same(t, canonical.Message, result.Message)
		assert.Equal(t, []string{"idempotency:" + request.Message.ClientNonce}, recorder.calls)
		phase151AssertPostCreateNoEffects(t, recorder.phase151PostCreateRecorder)
	})
}

func phase151PostCreateFingerprintFixture() (
	messenger.AuthorizedMessageProducerScope,
	messenger.PostCreatePipelineRequest,
) {
	scope := messenger.AuthorizedMessageProducerScope{
		TenantID: uuid.New(), ActorID: uuid.New(), ChannelID: uuid.New(),
		Producer: messenger.TrustedMessageProducerUser,
	}
	request := messenger.PostCreatePipelineRequest{
		Producer: messenger.TrustedMessageProducerUser,
		Message: messenger.MessageDocument{
			ChannelID: scope.ChannelID,
			SenderID: scope.ActorID,
			MessageType: "message",
			PollID: "poll-original",
			ContentJSON: map[string]interface{}{
				"type": "text",
				"count": float64(1),
				"nested": map[string]interface{}{"b": "two", "a": "one"},
			},
			ContentText: "phase151 fingerprint",
			ClientNonce: "phase151-fingerprint-v1",
			ForwardedFromID: "forwarded-original",
			VoiceURL: "https://hub.invalid/original.ogg",
			VoiceDurationMs: 1500,
			Attachments: []messenger.AttachmentRef{
				{FileID: "file-a", FileName: "a.txt", FileSize: 10, MimeType: "text/plain", URL: "/a"},
				{FileID: "file-b", FileName: "b.txt", FileSize: 20, MimeType: "text/plain", URL: "/b"},
			},
			Mentions: []string{"user-1", "user-2", "user-1"},
			MentionGroups: []string{"all", "online", "all"},
		},
	}
	return scope, request
}

func phase151PrepareFingerprintDocument(
	t *testing.T,
	scope messenger.AuthorizedMessageProducerScope,
	request messenger.PostCreatePipelineRequest,
	now time.Time,
) *messenger.MessageDocument {
	t.Helper()
	document, err := messenger.ExportPhase151PreparePendingCanonicalMessage(scope, request, nil, now)
	require.NoError(t, err)
	return document
}

func phase151MessageBSONFields(t *testing.T, document messenger.MessageDocument) bson.M {
	t.Helper()
	encoded, err := bson.Marshal(document)
	require.NoError(t, err)
	var fields bson.M
	require.NoError(t, bson.Unmarshal(encoded, &fields))
	return fields
}

func phase151MessageDocumentWithBSONOverrides(
	t *testing.T,
	document messenger.MessageDocument,
	overrides map[string]interface{},
) messenger.MessageDocument {
	t.Helper()
	fields := phase151MessageBSONFields(t, document)
	for key, value := range overrides {
		if value == nil {
			delete(fields, key)
			continue
		}
		fields[key] = value
	}
	encoded, err := bson.Marshal(fields)
	require.NoError(t, err)
	var result messenger.MessageDocument
	require.NoError(t, bson.Unmarshal(encoded, &result))
	return result
}

func phase151MessageDocumentPointer(document messenger.MessageDocument) *messenger.MessageDocument {
	return &document
}