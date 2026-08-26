package messenger_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/messenger"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

type phase151PromoteSourceBindingCall struct {
	channelID uuid.UUID
	messageID string
}

type phase151PromoteSourceBindingResult struct {
	document *messenger.MessageDocument
	found    bool
	err      error
}

type phase151PromoteSourceBindingRepositoryBase struct {
	*phase151MessageAccessRepository
	genericFindCalls int
	backlinkCalls    int
}

func newPhase151PromoteSourceBindingRepositoryBase() *phase151PromoteSourceBindingRepositoryBase {
	return &phase151PromoteSourceBindingRepositoryBase{
		phase151MessageAccessRepository: newPhase151MessageAccessRepository(),
	}
}

func (r *phase151PromoteSourceBindingRepositoryBase) FindMessageByID(
	ctx context.Context,
	messageID string,
) (*messenger.MessageDocument, error) {
	r.genericFindCalls++
	return r.mockRepository.FindMessageByID(ctx, messageID)
}

func (r *phase151PromoteSourceBindingRepositoryBase) InsertMessage(
	ctx context.Context,
	document *messenger.MessageDocument,
) (string, error) {
	r.backlinkCalls++
	return r.mockRepository.InsertMessage(ctx, document)
}

type phase151PromoteSourceBindingRepository struct {
	*phase151PromoteSourceBindingRepositoryBase
	scopedOverride bool
	scopedDocument *messenger.MessageDocument
	scopedFound    bool
	scopedErr      error
	scopedCalls    []phase151PromoteSourceBindingCall
	scopedResults  []phase151PromoteSourceBindingResult
}

func (r *phase151PromoteSourceBindingRepository) FindLiveMessageInChannel(
	ctx context.Context,
	channelID uuid.UUID,
	messageID string,
) (*messenger.MessageDocument, bool, error) {
	callIndex := len(r.scopedCalls)
	r.scopedCalls = append(r.scopedCalls, phase151PromoteSourceBindingCall{
		channelID: channelID,
		messageID: messageID,
	})
	if callIndex < len(r.scopedResults) {
		result := r.scopedResults[callIndex]
		return result.document, result.found, result.err
	}
	if r.scopedOverride {
		return r.scopedDocument, r.scopedFound, r.scopedErr
	}
	return r.mockRepository.FindLiveMessageInChannel(ctx, channelID, messageID)
}

type phase151PromoteSourceBindingAbsentRepository struct {
	messenger.Repository
}

type phase151PromoteSourceBindingTargetChecker struct {
	calls int
}

func (c *phase151PromoteSourceBindingTargetChecker) AuthorizePromotionTarget(
	_ context.Context,
	_ uuid.UUID,
	_ uuid.UUID,
) (bool, bool, error) {
	c.calls++
	return true, true, nil
}

type phase151PromoteSourceBindingPromoter struct {
	calls int
}

func (p *phase151PromoteSourceBindingPromoter) CreateIssueFromPromotion(
	_ context.Context,
	_ uuid.UUID,
	_ uuid.UUID,
	_ string,
	_ string,
	_ map[string]interface{},
) (string, string, error) {
	p.calls++
	return "HUB-98", uuid.NewString(), nil
}

type phase151PromoteSourceBindingLinker struct {
	calls int
}

func (l *phase151PromoteSourceBindingLinker) LinkThreadToIssue(
	_ context.Context,
	_ uuid.UUID,
	_ uuid.UUID,
	_ uuid.UUID,
	_ uuid.UUID,
	_ string,
) error {
	l.calls++
	return nil
}

func TestPhase151PromoteToIssueBindsSourceLookupToAuthorizedChannel(t *testing.T) {
	ctx := context.Background()

	t.Run("authorized route A uses two scoped live proofs without generic fallback", func(t *testing.T) {
		repository := &phase151PromoteSourceBindingRepository{
			phase151PromoteSourceBindingRepositoryBase: newPhase151PromoteSourceBindingRepositoryBase(),
		}
		actorID, channelID, channel, source := phase151PromoteSourceBindingFixture(repository.phase151PromoteSourceBindingRepositoryBase)

		service, target, promoter, linker := phase151PromoteSourceBindingService(repository, actorID, channel)

		response, err := service.PromoteToIssue(ctx, channelID, source.ID, actorID, phase151PromoteSourceBindingRequest())

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, "HUB-98", response.IssueKey)
		phase151AssertPromoteScopedCall(t, repository, channelID, source.ID, 2)
		assert.Zero(t, repository.genericFindCalls)
		assert.Equal(t, 2, target.calls)
		assert.Equal(t, 1, promoter.calls)
		assert.Equal(t, 1, linker.calls)
		assert.Equal(t, 1, repository.backlinkCalls)
	})

	t.Run("foreign and deleted live-scope misses are indistinguishable", func(t *testing.T) {
		var notFoundMessage string
		for _, scenario := range []string{"foreign B", "deleted source"} {
			t.Run(scenario, func(t *testing.T) {
				repository := &phase151PromoteSourceBindingRepository{
					phase151PromoteSourceBindingRepositoryBase: newPhase151PromoteSourceBindingRepositoryBase(),
				}
				actorID, channelID, channel, source := phase151PromoteSourceBindingFixture(repository.phase151PromoteSourceBindingRepositoryBase)
				if scenario == "foreign B" {
					foreign := *source
					foreign.ChannelID = uuid.New()
					repository.mockRepository.messages[source.ID] = &foreign
				} else {
					deletedAt := time.Now().UTC()
					deleted := *source
					deleted.DeletedAt = &deletedAt
					repository.mockRepository.messages[source.ID] = &deleted
				}
				service, target, promoter, linker := phase151PromoteSourceBindingService(repository, actorID, channel)

				response, err := service.PromoteToIssue(ctx, channelID, source.ID, actorID, phase151PromoteSourceBindingRequest())

				require.Nil(t, response)
				require.Error(t, err)
				requireAppErrorCode(t, err, 404)
				var appErr *apperrors.AppError
				require.ErrorAs(t, err, &appErr)
				if notFoundMessage == "" {
					notFoundMessage = appErr.Message
				} else {
					assert.Equal(t, notFoundMessage, appErr.Message)
				}
				phase151AssertPromoteScopedCall(t, repository, channelID, source.ID, 1)
				phase151AssertPromoteNoDownstream(t, repository.phase151PromoteSourceBindingRepositoryBase, target, promoter, linker)
			})
		}
	})

	t.Run("missing scoped lookup capability fails closed", func(t *testing.T) {
		base := newPhase151PromoteSourceBindingRepositoryBase()
		repository := &phase151PromoteSourceBindingAbsentRepository{Repository: base}
		actorID, channelID, channel, source := phase151PromoteSourceBindingFixture(base)
		service, target, promoter, linker := phase151PromoteSourceBindingService(repository, actorID, channel)

		response, err := service.PromoteToIssue(ctx, channelID, source.ID, actorID, phase151PromoteSourceBindingRequest())

		require.Nil(t, response)
		require.Error(t, err)
		requireAppErrorCode(t, err, 500)
		phase151AssertPromoteNoDownstream(t, base, target, promoter, linker)
	})

	for _, scenario := range []struct {
		name          string
		scopedDocument func(*messenger.MessageDocument) *messenger.MessageDocument
		scopedErr      error
	}{
		{name: "found nil document", scopedDocument: func(*messenger.MessageDocument) *messenger.MessageDocument { return nil }},
		{name: "found foreign document", scopedDocument: func(source *messenger.MessageDocument) *messenger.MessageDocument {
			foreign := *source
			foreign.ChannelID = uuid.New()
			return &foreign
		}},
		{name: "found deleted document", scopedDocument: func(source *messenger.MessageDocument) *messenger.MessageDocument {
			deletedAt := time.Now().UTC()
			deleted := *source
			deleted.DeletedAt = &deletedAt
			return &deleted
		}},
		{name: "scoped store error", scopedDocument: func(source *messenger.MessageDocument) *messenger.MessageDocument { return source }, scopedErr: errors.New("phase151 scoped source store error")},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			repository := &phase151PromoteSourceBindingRepository{
				phase151PromoteSourceBindingRepositoryBase: newPhase151PromoteSourceBindingRepositoryBase(),
				scopedOverride:                           true,
				scopedFound:                              true,
				scopedErr:                                scenario.scopedErr,
			}
			actorID, channelID, channel, source := phase151PromoteSourceBindingFixture(repository.phase151PromoteSourceBindingRepositoryBase)
			repository.scopedDocument = scenario.scopedDocument(source)
			service, target, promoter, linker := phase151PromoteSourceBindingService(repository, actorID, channel)

			response, err := service.PromoteToIssue(ctx, channelID, source.ID, actorID, phase151PromoteSourceBindingRequest())

			require.Nil(t, response)
			require.Error(t, err)
			requireAppErrorCode(t, err, 500)
			phase151AssertPromoteScopedCall(t, repository, channelID, source.ID, 1)
			phase151AssertPromoteNoDownstream(t, repository.phase151PromoteSourceBindingRepositoryBase, target, promoter, linker)
		})
	}
}

func TestPhase151PromoteToIssueRechecksLiveSourceBeforeMutation(t *testing.T) {
	ctx := context.Background()

	for _, scenario := range []struct {
		name     string
		second   func(*messenger.MessageDocument) phase151PromoteSourceBindingResult
		wantCode int
		redact   string
	}{
		{
			name:     "missing source is concealed after description",
			second:   func(*messenger.MessageDocument) phase151PromoteSourceBindingResult { return phase151PromoteSourceBindingResult{} },
			wantCode: 404,
		},
		{
			name: "source store error is redacted after description",
			second: func(source *messenger.MessageDocument) phase151PromoteSourceBindingResult {
				return phase151PromoteSourceBindingResult{document: source, found: true, err: errors.New("phase151 source recheck store error")}
			},
			wantCode: 500,
			redact:   "phase151 source recheck store error",
		},
		{
			name:     "found nil source is internal after description",
			second:   func(*messenger.MessageDocument) phase151PromoteSourceBindingResult { return phase151PromoteSourceBindingResult{found: true} },
			wantCode: 500,
		},
		{
			name: "mismatched source identity is internal after description",
			second: func(source *messenger.MessageDocument) phase151PromoteSourceBindingResult {
				mismatch := *source
				mismatch.ID += "-other"
				return phase151PromoteSourceBindingResult{document: &mismatch, found: true}
			},
			wantCode: 500,
		},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			repository := &phase151PromoteSourceBindingRepository{
				phase151PromoteSourceBindingRepositoryBase: newPhase151PromoteSourceBindingRepositoryBase(),
			}
			actorID, channelID, channel, source := phase151PromoteSourceBindingFixture(repository.phase151PromoteSourceBindingRepositoryBase)
			repository.scopedResults = []phase151PromoteSourceBindingResult{
				{document: source, found: true},
				scenario.second(source),
			}
			service, target, promoter, linker := phase151PromoteSourceBindingService(repository, actorID, channel)

			response, err := service.PromoteToIssue(ctx, channelID, source.ID, actorID, phase151PromoteSourceBindingRequest())

			require.Nil(t, response)
			require.Error(t, err)
			requireAppErrorCode(t, err, scenario.wantCode)
			if scenario.redact != "" {
				phase151AssertPromoteThreadIntegrityRedactedError(t, err, scenario.redact)
			}
			phase151AssertPromoteScopedCall(t, repository, channelID, source.ID, 2)
			assert.Equal(t, 1, target.calls)
			assert.Zero(t, repository.genericFindCalls)
			assert.Zero(t, promoter.calls)
			assert.Zero(t, linker.calls)
			assert.Zero(t, repository.backlinkCalls)
		})
	}
}

func TestPhase151PromoteToIssueRejectsFoundScopedSourceWithEmptyIdentity(t *testing.T) {
	ctx := context.Background()
	repository := &phase151PromoteSourceBindingRepository{
		phase151PromoteSourceBindingRepositoryBase: newPhase151PromoteSourceBindingRepositoryBase(),
		scopedOverride:                           true,
		scopedFound:                              true,
	}
	actorID, channelID, channel, source := phase151PromoteSourceBindingFixture(repository.phase151PromoteSourceBindingRepositoryBase)
	emptyIdentitySource := *source
	emptyIdentitySource.ID = ""
	repository.scopedDocument = &emptyIdentitySource
	service, target, promoter, linker := phase151PromoteSourceBindingService(repository, actorID, channel)

	response, err := service.PromoteToIssue(ctx, channelID, "", actorID, phase151PromoteSourceBindingRequest())

	require.Nil(t, response)
	require.Error(t, err)
	requireAppErrorCode(t, err, 500)
	phase151AssertPromoteScopedCall(t, repository, channelID, "", 1)
	phase151AssertPromoteNoDownstream(t, repository.phase151PromoteSourceBindingRepositoryBase, target, promoter, linker)
}

func phase151PromoteSourceBindingFixture(
	repository *phase151PromoteSourceBindingRepositoryBase,
) (uuid.UUID, uuid.UUID, *ent.Channel, *messenger.MessageDocument) {
	actorID := uuid.New()
	channelID := setupChannelWithMembers(repository.mockRepository, "private", false, actorID)
	channel := repository.channels[channelID]
	source := &messenger.MessageDocument{
		ID:          "phase151-promote-source-" + uuid.NewString(),
		ChannelID:   channelID,
		SenderID:    actorID,
		ContentText: "bound source",
		CreatedAt:   time.Now().UTC(),
	}
	repository.messages[source.ID] = source
	return actorID, channelID, channel, source
}

func phase151PromoteSourceBindingService(
	repository messenger.Repository,
	actorID uuid.UUID,
	channel *ent.Channel,
) (*messenger.MessengerService, *phase151PromoteSourceBindingTargetChecker, *phase151PromoteSourceBindingPromoter, *phase151PromoteSourceBindingLinker) {
	resolver := &phase151ChannelAccessResolver{grants: make(map[phase151ChannelAccessKey]messenger.ChannelAccess)}
	access := phase151ReadableAccess(channel, true)
	access.CanPromote = true
	phase151Grant(resolver, actorID, channel, messenger.ChannelActionPromote, access)
	service := messenger.ExportNewMessengerServiceWithChannelAccessTest(repository, resolver)
	target := &phase151PromoteSourceBindingTargetChecker{}
	promoter := &phase151PromoteSourceBindingPromoter{}
	linker := &phase151PromoteSourceBindingLinker{}
	service.SetPromotePermissionChecker(target)
	service.SetIssuePromoter(promoter)
	service.SetIssueThreadLinker(linker)
	return service, target, promoter, linker
}

func phase151PromoteSourceBindingRequest() messenger.PromoteToIssueRequest {
	return messenger.PromoteToIssueRequest{
		ProjectID: uuid.NewString(),
		Title:     "Bound source promotion",
	}
}

func phase151AssertPromoteScopedCall(
	t *testing.T,
	repository *phase151PromoteSourceBindingRepository,
	channelID uuid.UUID,
	messageID string,
	callCount int,
) {
	t.Helper()
	expectedCalls := make([]phase151PromoteSourceBindingCall, 0, callCount)
	for index := 0; index < callCount; index++ {
		expectedCalls = append(expectedCalls, phase151PromoteSourceBindingCall{channelID: channelID, messageID: messageID})
	}
	assert.Equal(t, expectedCalls, repository.scopedCalls)
}

func phase151AssertPromoteNoDownstream(
	t *testing.T,
	repository *phase151PromoteSourceBindingRepositoryBase,
	target *phase151PromoteSourceBindingTargetChecker,
	promoter *phase151PromoteSourceBindingPromoter,
	linker *phase151PromoteSourceBindingLinker,
) {
	t.Helper()
	assert.Zero(t, repository.genericFindCalls)
	assert.Zero(t, target.calls)
	assert.Zero(t, promoter.calls)
	assert.Zero(t, linker.calls)
	assert.Zero(t, repository.backlinkCalls)
}
