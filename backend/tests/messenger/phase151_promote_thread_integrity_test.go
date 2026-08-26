package messenger_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/messenger"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

type phase151PromoteThreadIntegrityCall struct {
	channelID uuid.UUID
	messageID string
}

type phase151PromoteThreadIntegrityRepliesCall struct {
	channelID uuid.UUID
	rootID    string
	limit     int
}

type phase151PromoteThreadIntegrityRepository struct {
	*phase151PromoteSourceBindingRepositoryBase

	order *[]string

	sourceCalls  []phase151PromoteThreadIntegrityCall
	parentCalls  []phase151PromoteThreadIntegrityCall
	repliesCalls []phase151PromoteThreadIntegrityRepliesCall

	parentDocument *messenger.MessageDocument
	parentFound    bool
	parentErr      error
	replies        []*messenger.MessageDocument
	repliesErr     error
	sourceID       string

	genericParentCalls  int
	genericRepliesCalls int
	getChannelCalls     int
	backlinkCalls       int
	backlinks           []messenger.MessageDocument
}

func (r *phase151PromoteThreadIntegrityRepository) record(operation string) {
	*r.order = append(*r.order, operation)
}

func (r *phase151PromoteThreadIntegrityRepository) FindLiveMessageInChannel(
	ctx context.Context,
	channelID uuid.UUID,
	messageID string,
) (*messenger.MessageDocument, bool, error) {
	if messageID == r.sourceID && len(r.sourceCalls) == 0 {
		r.record("scoped-source")
		r.sourceCalls = append(r.sourceCalls, phase151PromoteThreadIntegrityCall{channelID: channelID, messageID: messageID})
		return r.mockRepository.FindLiveMessageInChannel(ctx, channelID, messageID)
	}
	r.record("scoped-parent")
	r.parentCalls = append(r.parentCalls, phase151PromoteThreadIntegrityCall{channelID: channelID, messageID: messageID})
	return r.parentDocument, r.parentFound, r.parentErr
}

func (r *phase151PromoteThreadIntegrityRepository) ListLiveThreadRepliesInChannel(
	_ context.Context,
	channelID uuid.UUID,
	rootID string,
	limit int,
) ([]*messenger.MessageDocument, error) {
	r.record("scoped-replies")
	r.repliesCalls = append(r.repliesCalls, phase151PromoteThreadIntegrityRepliesCall{
		channelID: channelID,
		rootID:    rootID,
		limit:     limit,
	})
	return r.replies, r.repliesErr
}

func (r *phase151PromoteThreadIntegrityRepository) FindMessageByID(
	_ context.Context,
	_ string,
) (*messenger.MessageDocument, error) {
	r.record("generic-parent")
	r.genericParentCalls++
	if r.parentErr != nil {
		return nil, r.parentErr
	}
	if !r.parentFound {
		return nil, &ent.NotFoundError{}
	}
	return r.parentDocument, nil
}

func (r *phase151PromoteThreadIntegrityRepository) ListThreadReplies(
	_ context.Context,
	_ string,
	_ *time.Time,
	_ int,
) ([]*messenger.MessageDocument, error) {
	r.record("generic-replies")
	r.genericRepliesCalls++
	return r.replies, r.repliesErr
}

func (r *phase151PromoteThreadIntegrityRepository) GetChannel(
	ctx context.Context,
	channelID uuid.UUID,
) (*ent.Channel, error) {
	r.record("get-channel")
	r.getChannelCalls++
	return r.mockRepository.GetChannel(ctx, channelID)
}

func (r *phase151PromoteThreadIntegrityRepository) InsertMessage(
	_ context.Context,
	document *messenger.MessageDocument,
) (string, error) {
	r.record("backlink")
	r.backlinkCalls++
	if document != nil {
		copy := *document
		r.backlinks = append(r.backlinks, copy)
	}
	return "phase151-promotion-backlink", nil
}

type phase151PromoteThreadIntegrityAbsentRepliesRepository struct {
	messenger.Repository
	delegate *phase151PromoteThreadIntegrityRepository
}

func (r *phase151PromoteThreadIntegrityAbsentRepliesRepository) FindLiveMessageInChannel(
	ctx context.Context,
	channelID uuid.UUID,
	messageID string,
) (*messenger.MessageDocument, bool, error) {
	return r.delegate.FindLiveMessageInChannel(ctx, channelID, messageID)
}

type phase151PromoteThreadIntegrityTargetCall struct {
	actorID   uuid.UUID
	projectID uuid.UUID
}

type phase151PromoteThreadIntegrityTargetChecker struct {
	order     *[]string
	calls     int
	arguments []phase151PromoteThreadIntegrityTargetCall
}

func (c *phase151PromoteThreadIntegrityTargetChecker) AuthorizePromotionTarget(
	_ context.Context,
	actorID uuid.UUID,
	projectID uuid.UUID,
) (bool, bool, error) {
	*c.order = append(*c.order, "target-authority")
	c.calls++
	c.arguments = append(c.arguments, phase151PromoteThreadIntegrityTargetCall{actorID: actorID, projectID: projectID})
	return true, true, nil
}

type phase151PromoteThreadIntegrityPromoteCall struct {
	projectID uuid.UUID
	actorID   uuid.UUID
	title     string
	issueType string
}

type phase151PromoteThreadIntegrityPromoter struct {
	order       *[]string
	calls       int
	arguments   []phase151PromoteThreadIntegrityPromoteCall
	description map[string]interface{}
}

func (p *phase151PromoteThreadIntegrityPromoter) CreateIssueFromPromotion(
	_ context.Context,
	projectID uuid.UUID,
	actorID uuid.UUID,
	title string,
	issueType string,
	description map[string]interface{},
) (string, string, error) {
	*p.order = append(*p.order, "promoter")
	p.calls++
	p.arguments = append(p.arguments, phase151PromoteThreadIntegrityPromoteCall{
		projectID: projectID,
		actorID:   actorID,
		title:     title,
		issueType: issueType,
	})
	p.description = description
	return "HUB-151", phase151PromoteThreadIntegrityIssueID.String(), nil
}

type phase151PromoteThreadIntegrityLinkCall struct {
	projectID uuid.UUID
	issueID   uuid.UUID
	actorID   uuid.UUID
	channelID uuid.UUID
	rootID    string
}

type phase151PromoteThreadIntegrityLinker struct {
	order     *[]string
	calls     int
	arguments []phase151PromoteThreadIntegrityLinkCall
}

func (l *phase151PromoteThreadIntegrityLinker) LinkThreadToIssue(
	_ context.Context,
	projectID uuid.UUID,
	issueID uuid.UUID,
	actorID uuid.UUID,
	channelID uuid.UUID,
	rootID string,
) error {
	*l.order = append(*l.order, "linker")
	l.calls++
	l.arguments = append(l.arguments, phase151PromoteThreadIntegrityLinkCall{
		projectID: projectID,
		issueID:   issueID,
		actorID:   actorID,
		channelID: channelID,
		rootID:    rootID,
	})
	return nil
}

var phase151PromoteThreadIntegrityIssueID = uuid.MustParse("15151515-1515-4151-8151-151515151515")

func TestPhase151PromoteToIssueRootUsesSelfAndScopedRepliesInRepositoryOrder(t *testing.T) {
	ctx := context.Background()
	order := make([]string, 0, 8)
	repository, actorID, channelID, channel, source, projectID := phase151PromoteThreadIntegrityFixture(&order, false)
	repository.replies = []*messenger.MessageDocument{
		phase151PromoteThreadIntegrityReply("reply-new", channelID, source.ID, "newest reply", time.Date(2026, time.August, 26, 10, 2, 0, 0, time.UTC)),
		phase151PromoteThreadIntegrityReply("reply-old", channelID, source.ID, "older reply", time.Date(2026, time.August, 26, 10, 1, 0, 0, time.UTC)),
	}
	service, target, promoter, linker := phase151PromoteThreadIntegrityService(repository, actorID, channel, &order)

	response, err := service.PromoteToIssue(ctx, channelID, source.ID, actorID, phase151PromoteThreadIntegrityRequest(projectID))

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, []phase151PromoteThreadIntegrityCall{{channelID: channelID, messageID: source.ID}}, repository.sourceCalls)
	assert.Empty(t, repository.parentCalls, "root source must not issue a second parent lookup")
	assert.Equal(t, []phase151PromoteThreadIntegrityRepliesCall{{channelID: channelID, rootID: source.ID, limit: 100}}, repository.repliesCalls)
	phase151AssertPromoteThreadIntegrityNoGenericReads(t, repository)
	assert.Equal(t, []string{
		"source-action", "scoped-source", "target-authority", "scoped-replies", "get-channel", "promoter", "backlink", "linker",
	}, order)
	assert.Equal(t, []string{"- @user: root source", "- @user: newest reply", "- @user: older reply"}, phase151PromoteThreadIntegrityLines(t, promoter.description))
	assert.Equal(t, 1, linker.calls)
	assert.Equal(t, 1, repository.backlinkCalls)
	phase151AssertPromoteThreadIntegritySuccessfulEffects(t, response, target, promoter, repository, linker, projectID, actorID, channelID, source.ID)
}

func TestPhase151PromoteToIssueReplyUsesExactScopedRootAndRepliesInRepositoryOrder(t *testing.T) {
	ctx := context.Background()
	order := make([]string, 0, 9)
	repository, actorID, channelID, channel, source, projectID := phase151PromoteThreadIntegrityFixture(&order, true)
	repository.replies = []*messenger.MessageDocument{
		phase151PromoteThreadIntegrityReply("reply-later", channelID, source.ThreadID, "later reply", time.Date(2026, time.August, 26, 10, 3, 0, 0, time.UTC)),
		source,
		phase151PromoteThreadIntegrityReply("reply-earlier", channelID, source.ThreadID, "earlier reply", time.Date(2026, time.August, 26, 10, 2, 0, 0, time.UTC)),
	}
	service, target, promoter, linker := phase151PromoteThreadIntegrityService(repository, actorID, channel, &order)

	response, err := service.PromoteToIssue(ctx, channelID, source.ID, actorID, phase151PromoteThreadIntegrityRequest(projectID))

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, []phase151PromoteThreadIntegrityCall{{channelID: channelID, messageID: source.ID}}, repository.sourceCalls)
	assert.Equal(t, []phase151PromoteThreadIntegrityCall{{channelID: channelID, messageID: source.ThreadID}}, repository.parentCalls)
	assert.Equal(t, []phase151PromoteThreadIntegrityRepliesCall{{channelID: channelID, rootID: source.ThreadID, limit: 100}}, repository.repliesCalls)
	phase151AssertPromoteThreadIntegrityNoGenericReads(t, repository)
	assert.Equal(t, []string{
		"source-action", "scoped-source", "target-authority", "scoped-parent", "scoped-replies", "get-channel", "promoter", "backlink", "linker",
	}, order)
	replyLines := phase151PromoteThreadIntegrityLines(t, promoter.description)
	assert.Equal(t, []string{"- @user: root source", "- @user: later reply", "- @user: promoted reply", "- @user: earlier reply"}, replyLines)
	assert.Equal(t, 1, strings.Count(strings.Join(replyLines, "\n"), "- @user: promoted reply"))
	assert.Equal(t, 1, linker.calls)
	assert.Equal(t, 1, repository.backlinkCalls)
	phase151AssertPromoteThreadIntegritySuccessfulEffects(t, response, target, promoter, repository, linker, projectID, actorID, channelID, source.ThreadID)
}

func TestPhase151PromoteToIssueRetainsSelectedReplyOutsideNewestHundred(t *testing.T) {
	ctx := context.Background()
	order := make([]string, 0, 9)
	repository, actorID, channelID, channel, source, projectID := phase151PromoteThreadIntegrityFixture(&order, true)
	newestAt := time.Date(2026, time.August, 26, 12, 0, 0, 0, time.UTC)
	repository.replies = make([]*messenger.MessageDocument, 0, 100)
	expectedLines := []string{"- @user: root source"}
	for index := 0; index < 100; index++ {
		content := fmt.Sprintf("page reply %03d", index)
		repository.replies = append(repository.replies, phase151PromoteThreadIntegrityReply(
			fmt.Sprintf("phase151-page-reply-%03d", index),
			channelID,
			source.ThreadID,
			content,
			newestAt.Add(-time.Duration(index)*time.Minute),
		))
		if index < 99 {
			expectedLines = append(expectedLines, "- @user: "+content)
		}
	}
	source.CreatedAt = newestAt.Add(-101 * time.Minute)
	expectedLines = append(expectedLines, "- @user: promoted reply")
	originalReplies := append([]*messenger.MessageDocument(nil), repository.replies...)
	service, target, promoter, linker := phase151PromoteThreadIntegrityService(repository, actorID, channel, &order)

	response, err := service.PromoteToIssue(ctx, channelID, source.ID, actorID, phase151PromoteThreadIntegrityRequest(projectID))

	require.NoError(t, err)
	require.NotNil(t, response)
	require.Len(t, repository.replies, len(originalReplies))
	for index, originalReply := range originalReplies {
		assert.Same(t, originalReply, repository.replies[index])
	}
	require.NotNil(t, repository.replies[99])
	assert.Equal(t, "phase151-page-reply-099", repository.replies[99].ID)
	lines := phase151PromoteThreadIntegrityLines(t, promoter.description)
	assert.Equal(t, expectedLines, lines)
	assert.Len(t, lines, 101, "description must contain one root and at most 100 replies")
	assert.Equal(t, 1, strings.Count(strings.Join(lines, "\n"), "- @user: promoted reply"))
	assert.NotContains(t, lines, "- @user: page reply 099")
	assert.Equal(t, []phase151PromoteThreadIntegrityRepliesCall{{channelID: channelID, rootID: source.ThreadID, limit: 100}}, repository.repliesCalls)
	phase151AssertPromoteThreadIntegrityNoGenericReads(t, repository)
	phase151AssertPromoteThreadIntegritySuccessfulEffects(t, response, target, promoter, repository, linker, projectID, actorID, channelID, source.ThreadID)
}

func TestPhase151PromoteToIssueAppendsSelectedReplyToShortPage(t *testing.T) {
	ctx := context.Background()
	order := make([]string, 0, 9)
	repository, actorID, channelID, channel, source, projectID := phase151PromoteThreadIntegrityFixture(&order, true)
	repository.replies = []*messenger.MessageDocument{
		phase151PromoteThreadIntegrityReply("reply-later", channelID, source.ThreadID, "later reply", time.Date(2026, time.August, 26, 10, 3, 0, 0, time.UTC)),
		phase151PromoteThreadIntegrityReply("reply-earlier", channelID, source.ThreadID, "earlier reply", time.Date(2026, time.August, 26, 10, 2, 0, 0, time.UTC)),
	}
	service, target, promoter, linker := phase151PromoteThreadIntegrityService(repository, actorID, channel, &order)

	response, err := service.PromoteToIssue(ctx, channelID, source.ID, actorID, phase151PromoteThreadIntegrityRequest(projectID))

	require.NoError(t, err)
	require.NotNil(t, response)
	lines := phase151PromoteThreadIntegrityLines(t, promoter.description)
	assert.Equal(t, []string{
		"- @user: root source",
		"- @user: later reply",
		"- @user: earlier reply",
		"- @user: promoted reply",
	}, lines)
	assert.Equal(t, 1, strings.Count(strings.Join(lines, "\n"), "- @user: promoted reply"))
	phase151AssertPromoteThreadIntegrityNoGenericReads(t, repository)
	phase151AssertPromoteThreadIntegritySuccessfulEffects(t, response, target, promoter, repository, linker, projectID, actorID, channelID, source.ThreadID)
}

func TestPhase151PromoteToIssueRejectsDuplicateAndOversizedReplyPagesBeforeEffects(t *testing.T) {
	ctx := context.Background()
	for _, scenario := range []struct {
		name      string
		configure func(*phase151PromoteThreadIntegrityRepository, uuid.UUID, string, *messenger.MessageDocument) string
	}{
		{
			name: "duplicate reply ID",
			configure: func(repository *phase151PromoteThreadIntegrityRepository, channelID uuid.UUID, rootID string, _ *messenger.MessageDocument) string {
				duplicateID := "phase151-duplicate-reply"
				repository.replies = []*messenger.MessageDocument{
					phase151PromoteThreadIntegrityReply(duplicateID, channelID, rootID, "first duplicate", time.Date(2026, time.August, 26, 10, 3, 0, 0, time.UTC)),
					phase151PromoteThreadIntegrityReply(duplicateID, channelID, rootID, "second duplicate", time.Date(2026, time.August, 26, 10, 2, 0, 0, time.UTC)),
				}
				return duplicateID
			},
		},
		{
			name: "more than one hundred replies",
			configure: func(repository *phase151PromoteThreadIntegrityRepository, channelID uuid.UUID, rootID string, _ *messenger.MessageDocument) string {
				newestAt := time.Date(2026, time.August, 26, 12, 0, 0, 0, time.UTC)
				repository.replies = make([]*messenger.MessageDocument, 0, 101)
				for index := 0; index < 101; index++ {
					repository.replies = append(repository.replies, phase151PromoteThreadIntegrityReply(
						fmt.Sprintf("phase151-oversized-reply-%03d", index),
						channelID,
						rootID,
						fmt.Sprintf("oversized reply %03d", index),
						newestAt.Add(-time.Duration(index)*time.Minute),
					))
				}
				return "phase151-oversized-reply-100"
			},
		},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			order := make([]string, 0, 6)
			repository, actorID, channelID, channel, source, projectID := phase151PromoteThreadIntegrityFixture(&order, true)
			raw := scenario.configure(repository, channelID, source.ThreadID, source)
			service, _, promoter, linker := phase151PromoteThreadIntegrityService(repository, actorID, channel, &order)

			response, err := service.PromoteToIssue(ctx, channelID, source.ID, actorID, phase151PromoteThreadIntegrityRequest(projectID))

			require.Nil(t, response)
			require.Error(t, err)
			requireAppErrorCode(t, err, 500)
			phase151AssertPromoteThreadIntegrityRedactedError(t, err, raw)
			assert.Equal(t, []phase151PromoteThreadIntegrityRepliesCall{{channelID: channelID, rootID: source.ThreadID, limit: 100}}, repository.repliesCalls)
			phase151AssertPromoteThreadIntegrityNoEffects(t, repository, promoter, linker)
			phase151AssertPromoteThreadIntegrityNoGenericReads(t, repository)
			assert.Equal(t, []string{"source-action", "scoped-source", "target-authority", "scoped-parent", "scoped-replies"}, order)
		})
	}
}
func TestPhase151PromoteToIssueRejectsMalformedFoundRootSourceIdentityBeforeEffects(t *testing.T) {
	ctx := context.Background()
	for _, scenario := range []struct {
		name               string
		mutateSource       func(*messenger.MessageDocument)
		redactMismatchedID bool
	}{
		{
			name: "empty source ID",
			mutateSource: func(source *messenger.MessageDocument) {
				source.ID = ""
			},
		},
		{
			name: "mismatched source ID",
			mutateSource: func(source *messenger.MessageDocument) {
				source.ID = "phase151-mismatched-source-" + uuid.NewString()
			},
			redactMismatchedID: true,
		},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			order := make([]string, 0, 2)
			repository, actorID, channelID, channel, source, projectID := phase151PromoteThreadIntegrityFixture(&order, false)
			routeMessageID := source.ID
			scenario.mutateSource(source)
			service, target, promoter, linker := phase151PromoteThreadIntegrityService(repository, actorID, channel, &order)

			response, err := service.PromoteToIssue(ctx, channelID, routeMessageID, actorID, phase151PromoteThreadIntegrityRequest(projectID))

			require.Nil(t, response)
			require.Error(t, err)
			requireAppErrorCode(t, err, 500)
			phase151AssertPromoteThreadIntegrityRedactedError(t, err, routeMessageID)
			if scenario.redactMismatchedID {
				phase151AssertPromoteThreadIntegrityRedactedError(t, err, source.ID)
			}
			assert.Equal(t, []phase151PromoteThreadIntegrityCall{{channelID: channelID, messageID: routeMessageID}}, repository.sourceCalls)
			assert.Empty(t, repository.parentCalls)
			assert.Empty(t, repository.repliesCalls)
			assert.Zero(t, target.calls)
			assert.Zero(t, repository.genericFindCalls)
			phase151AssertPromoteThreadIntegrityNoEffects(t, repository, promoter, linker)
			phase151AssertPromoteThreadIntegrityNoGenericReads(t, repository)
			assert.Equal(t, []string{"source-action", "scoped-source"}, order)
		})
	}
}

func TestPhase151PromoteToIssueReplyRequiresExactScopedLiveRootBeforeEffects(t *testing.T) {
	ctx := context.Background()
	var nonDisclosureMessage string
	for _, scenario := range []struct {
		name            string
		configure       func(*phase151PromoteThreadIntegrityRepository, *messenger.MessageDocument, *messenger.MessageDocument)
		wantCode        int
		wantParentCalls int
		nonDisclosure   bool
		redact          string
	}{
		{
			name: "missing parent is non-disclosive",
			configure: func(repository *phase151PromoteThreadIntegrityRepository, _ *messenger.MessageDocument, _ *messenger.MessageDocument) {
				repository.parentFound = false
			},
			wantCode: 404, wantParentCalls: 1, nonDisclosure: true,
		},
		{
			name: "deleted parent is non-disclosive",
			configure: func(repository *phase151PromoteThreadIntegrityRepository, _ *messenger.MessageDocument, parent *messenger.MessageDocument) {
				deletedAt := time.Date(2026, time.August, 26, 10, 3, 0, 0, time.UTC)
				parent.DeletedAt = &deletedAt
				repository.parentFound = false
			},
			wantCode: 404, wantParentCalls: 1, nonDisclosure: true,
		},
		{
			name: "foreign parent is non-disclosive",
			configure: func(repository *phase151PromoteThreadIntegrityRepository, _ *messenger.MessageDocument, parent *messenger.MessageDocument) {
				parent.ChannelID = uuid.New()
				repository.parentFound = false
			},
			wantCode: 404, wantParentCalls: 1, nonDisclosure: true,
		},
		{
			name: "nested parent is non-disclosive",
			configure: func(_ *phase151PromoteThreadIntegrityRepository, _ *messenger.MessageDocument, parent *messenger.MessageDocument) {
				parent.ThreadID = "phase151-nested-root"
			},
			wantCode: 404, wantParentCalls: 1, nonDisclosure: true,
		},
		{
			name: "parent ID mismatch is an internal invariant",
			configure: func(_ *phase151PromoteThreadIntegrityRepository, _ *messenger.MessageDocument, parent *messenger.MessageDocument) {
				parent.ID = "phase151-wrong-root"
			},
			wantCode: 500, wantParentCalls: 1,
		},
		{
			name: "found nil parent is an internal invariant",
			configure: func(repository *phase151PromoteThreadIntegrityRepository, _ *messenger.MessageDocument, _ *messenger.MessageDocument) {
				repository.parentDocument = nil
				repository.parentFound = true
			},
			wantCode: 500, wantParentCalls: 1,
		},
		{
			name: "found foreign parent is an internal invariant",
			configure: func(repository *phase151PromoteThreadIntegrityRepository, _ *messenger.MessageDocument, parent *messenger.MessageDocument) {
				parent.ChannelID = uuid.New()
				repository.parentFound = true
			},
			wantCode: 500, wantParentCalls: 1,
		},
		{
			name: "found deleted parent is an internal invariant",
			configure: func(repository *phase151PromoteThreadIntegrityRepository, _ *messenger.MessageDocument, parent *messenger.MessageDocument) {
				deletedAt := time.Date(2026, time.August, 26, 10, 4, 0, 0, time.UTC)
				parent.DeletedAt = &deletedAt
				repository.parentFound = true
			},
			wantCode: 500, wantParentCalls: 1,
		},
		{
			name: "self thread ID is non-disclosive",
			configure: func(repository *phase151PromoteThreadIntegrityRepository, source *messenger.MessageDocument, _ *messenger.MessageDocument) {
				source.ThreadID = source.ID
				repository.parentDocument = source
				repository.parentFound = true
			},
			wantCode: 404, wantParentCalls: 1, nonDisclosure: true,
		},
		{
			name: "scoped parent infrastructure error is internal",
			configure: func(repository *phase151PromoteThreadIntegrityRepository, _ *messenger.MessageDocument, _ *messenger.MessageDocument) {
				repository.parentErr = errors.New("phase151 scoped parent failure")
			},
			wantCode: 500, wantParentCalls: 1, redact: "phase151 scoped parent failure",
		},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			order := make([]string, 0, 8)
			repository, actorID, channelID, channel, source, projectID := phase151PromoteThreadIntegrityFixture(&order, true)
			scenario.configure(repository, source, repository.parentDocument)
			service, _, promoter, linker := phase151PromoteThreadIntegrityService(repository, actorID, channel, &order)

			response, err := service.PromoteToIssue(ctx, channelID, source.ID, actorID, phase151PromoteThreadIntegrityRequest(projectID))

			require.Nil(t, response)
			require.Error(t, err)
			requireAppErrorCode(t, err, scenario.wantCode)
			if scenario.nonDisclosure {
				var appErr *apperrors.AppError
				require.ErrorAs(t, err, &appErr)
				if nonDisclosureMessage == "" {
					nonDisclosureMessage = appErr.Message
				}
				assert.Equal(t, nonDisclosureMessage, appErr.Message)
			}
			if scenario.redact != "" {
				phase151AssertPromoteThreadIntegrityRedactedError(t, err, scenario.redact)
			}
			assert.Len(t, repository.parentCalls, scenario.wantParentCalls)
			phase151AssertPromoteThreadIntegrityNoEffects(t, repository, promoter, linker)
			phase151AssertPromoteThreadIntegrityNoGenericReads(t, repository)
		})
	}
}

func TestPhase151PromoteToIssueRejectsEachMalformedScopedReplyBeforeEffects(t *testing.T) {
	ctx := context.Background()
	for _, replySource := range []bool{false, true} {
		sourceName := "root source"
		if replySource {
			sourceName = "reply source"
		}
		t.Run(sourceName, func(t *testing.T) {
			for _, scenario := range []struct {
				name      string
				makeReply func(*phase151PromoteThreadIntegrityRepository, uuid.UUID, string) *messenger.MessageDocument
			}{
				{
					name: "nil reply",
					makeReply: func(_ *phase151PromoteThreadIntegrityRepository, _ uuid.UUID, _ string) *messenger.MessageDocument {
						return nil
					},
				},
				{
					name: "empty reply ID",
					makeReply: func(_ *phase151PromoteThreadIntegrityRepository, channelID uuid.UUID, rootID string) *messenger.MessageDocument {
						return phase151PromoteThreadIntegrityReply("", channelID, rootID, "must not accept empty ID", time.Now().UTC())
					},
				},
				{
					name: "foreign reply",
					makeReply: func(_ *phase151PromoteThreadIntegrityRepository, _ uuid.UUID, rootID string) *messenger.MessageDocument {
						return phase151PromoteThreadIntegrityReply("reply-foreign", uuid.New(), rootID, "must not disclose foreign", time.Now().UTC())
					},
				},
				{
					name: "deleted reply",
					makeReply: func(_ *phase151PromoteThreadIntegrityRepository, channelID uuid.UUID, rootID string) *messenger.MessageDocument {
						reply := phase151PromoteThreadIntegrityReply("reply-deleted", channelID, rootID, "must not disclose deleted", time.Now().UTC())
						deletedAt := time.Now().UTC()
						reply.DeletedAt = &deletedAt
						return reply
					},
				},
				{
					name: "wrong root reply",
					makeReply: func(_ *phase151PromoteThreadIntegrityRepository, channelID uuid.UUID, _ string) *messenger.MessageDocument {
						return phase151PromoteThreadIntegrityReply("reply-wrong-root", channelID, "phase151-other-root", "must not disclose wrong root", time.Now().UTC())
					},
				},
				{
					name: "root row returned as reply",
					makeReply: func(repository *phase151PromoteThreadIntegrityRepository, _ uuid.UUID, rootID string) *messenger.MessageDocument {
						return repository.messages[rootID]
					},
				},
			} {
				scenario := scenario
				t.Run(scenario.name, func(t *testing.T) {
					order := make([]string, 0, 8)
					repository, actorID, channelID, channel, source, projectID := phase151PromoteThreadIntegrityFixture(&order, replySource)
					rootID := source.ID
					if replySource {
						rootID = source.ThreadID
					}
					repository.replies = []*messenger.MessageDocument{scenario.makeReply(repository, channelID, rootID)}
					service, _, promoter, linker := phase151PromoteThreadIntegrityService(repository, actorID, channel, &order)

					response, err := service.PromoteToIssue(ctx, channelID, source.ID, actorID, phase151PromoteThreadIntegrityRequest(projectID))

					require.Nil(t, response)
					require.Error(t, err)
					requireAppErrorCode(t, err, 500)
					phase151AssertPromoteThreadIntegrityNoEffects(t, repository, promoter, linker)
					phase151AssertPromoteThreadIntegrityNoGenericReads(t, repository)
				})
			}
		})
	}
}

func TestPhase151PromoteToIssueRejectsMissingScopedRepliesCapabilityAndError(t *testing.T) {
	ctx := context.Background()
	for _, replySource := range []bool{false, true} {
		sourceName := "root source"
		if replySource {
			sourceName = "reply source"
		}
		t.Run(sourceName, func(t *testing.T) {
			for _, scenario := range []struct {
				name              string
				serviceRepository func(*phase151PromoteThreadIntegrityRepository) messenger.Repository
				configure         func(*phase151PromoteThreadIntegrityRepository)
				redact            string
			}{
				{
					name: "missing scoped replies capability",
					serviceRepository: func(repository *phase151PromoteThreadIntegrityRepository) messenger.Repository {
						return &phase151PromoteThreadIntegrityAbsentRepliesRepository{Repository: repository, delegate: repository}
					},
					configure: func(_ *phase151PromoteThreadIntegrityRepository) {},
				},
				{
					name: "scoped replies infrastructure error",
					serviceRepository: func(repository *phase151PromoteThreadIntegrityRepository) messenger.Repository { return repository },
					configure: func(repository *phase151PromoteThreadIntegrityRepository) {
						repository.repliesErr = errors.New("phase151 scoped replies failure")
					},
					redact: "phase151 scoped replies failure",
				},
			} {
				scenario := scenario
				t.Run(scenario.name, func(t *testing.T) {
					order := make([]string, 0, 8)
					repository, actorID, channelID, channel, source, projectID := phase151PromoteThreadIntegrityFixture(&order, replySource)
					scenario.configure(repository)
					service, _, promoter, linker := phase151PromoteThreadIntegrityService(scenario.serviceRepository(repository), actorID, channel, &order)

					response, err := service.PromoteToIssue(ctx, channelID, source.ID, actorID, phase151PromoteThreadIntegrityRequest(projectID))

					require.Nil(t, response)
					require.Error(t, err)
					requireAppErrorCode(t, err, 500)
					if scenario.redact != "" {
						phase151AssertPromoteThreadIntegrityRedactedError(t, err, scenario.redact)
					}
					phase151AssertPromoteThreadIntegrityNoEffects(t, repository, promoter, linker)
					phase151AssertPromoteThreadIntegrityNoGenericReads(t, repository)
				})
			}
		})
	}
}

func phase151PromoteThreadIntegrityFixture(
	order *[]string,
	replySource bool,
) (*phase151PromoteThreadIntegrityRepository, uuid.UUID, uuid.UUID, *ent.Channel, *messenger.MessageDocument, uuid.UUID) {
	repository := &phase151PromoteThreadIntegrityRepository{
		phase151PromoteSourceBindingRepositoryBase: newPhase151PromoteSourceBindingRepositoryBase(),
		order:                                      order,
	}
	actorID := uuid.New()
	channelID := setupChannelWithMembers(repository.mockRepository, "private", false, actorID)
	channel := repository.channels[channelID]
	root := &messenger.MessageDocument{
		ID:          "phase151-promote-thread-root-" + uuid.NewString(),
		ChannelID:   channelID,
		SenderID:    actorID,
		ContentText: "root source",
		CreatedAt:   time.Date(2026, time.August, 26, 10, 0, 0, 0, time.UTC),
	}
	source := root
	if replySource {
		source = phase151PromoteThreadIntegrityReply(
			"phase151-promote-thread-reply-"+uuid.NewString(),
			channelID,
			root.ID,
			"promoted reply",
			time.Date(2026, time.August, 26, 10, 1, 0, 0, time.UTC),
		)
		repository.parentDocument = root
		repository.parentFound = true
	}
	repository.messages[root.ID] = root
	repository.messages[source.ID] = source
	repository.sourceID = source.ID
	return repository, actorID, channelID, channel, source, uuid.New()
}

func phase151PromoteThreadIntegrityReply(
	id string,
	channelID uuid.UUID,
	rootID string,
	content string,
	createdAt time.Time,
) *messenger.MessageDocument {
	return &messenger.MessageDocument{
		ID:          id,
		ChannelID:   channelID,
		SenderID:    uuid.New(),
		ContentText: content,
		ThreadID:    rootID,
		CreatedAt:   createdAt,
	}
}

func phase151PromoteThreadIntegrityService(
	repository messenger.Repository,
	actorID uuid.UUID,
	channel *ent.Channel,
	order *[]string,
) (*messenger.MessengerService, *phase151PromoteThreadIntegrityTargetChecker, *phase151PromoteThreadIntegrityPromoter, *phase151PromoteThreadIntegrityLinker) {
	resolver := &phase151ChannelAccessResolver{
		grants: make(map[phase151ChannelAccessKey]messenger.ChannelAccess),
		onAuthorize: func(phase151ChannelAccessKey) {
			*order = append(*order, "source-action")
		},
	}
	access := phase151ReadableAccess(channel, true)
	access.CanPromote = true
	phase151Grant(resolver, actorID, channel, messenger.ChannelActionPromote, access)

	target := &phase151PromoteThreadIntegrityTargetChecker{order: order}
	promoter := &phase151PromoteThreadIntegrityPromoter{order: order}
	linker := &phase151PromoteThreadIntegrityLinker{order: order}
	service := messenger.ExportNewMessengerServiceWithChannelAccessTest(repository, resolver)
	service.SetPromotePermissionChecker(target)
	service.SetIssuePromoter(promoter)
	service.SetIssueThreadLinker(linker)
	return service, target, promoter, linker
}

func phase151PromoteThreadIntegrityRequest(projectID uuid.UUID) messenger.PromoteToIssueRequest {
	return messenger.PromoteToIssueRequest{
		ProjectID: projectID.String(),
		Title:     "Thread integrity promotion",
	}
}

func phase151PromoteThreadIntegrityLines(t *testing.T, description map[string]interface{}) []string {
	t.Helper()
	children, ok := description["children"].([]interface{})
	require.True(t, ok)
	require.GreaterOrEqual(t, len(children), 3)
	lines := make([]string, 0, len(children)-3)
	for _, child := range children[3:] {
		paragraph, ok := child.(map[string]interface{})
		require.True(t, ok)
		inline, ok := paragraph["children"].([]interface{})
		require.Len(t, inline, 1)
		text, ok := inline[0].(map[string]interface{})["text"].(string)
		require.True(t, ok)
		lines = append(lines, text)
	}
	return lines
}

func phase151AssertPromoteThreadIntegrityRedactedError(t *testing.T, err error, raw string) {
	t.Helper()
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.NotContains(t, appErr.Message, raw)
	assert.NotContains(t, appErr.Error(), raw)
	assert.NotContains(t, err.Error(), raw)
}

func phase151AssertPromoteThreadIntegrityNoGenericReads(
	t *testing.T,
	repository *phase151PromoteThreadIntegrityRepository,
) {
	t.Helper()
	assert.Zero(t, repository.genericParentCalls)
	assert.Zero(t, repository.genericRepliesCalls)
}

func phase151AssertPromoteThreadIntegritySuccessfulEffects(
	t *testing.T,
	response *messenger.PromoteToIssueResponse,
	target *phase151PromoteThreadIntegrityTargetChecker,
	promoter *phase151PromoteThreadIntegrityPromoter,
	repository *phase151PromoteThreadIntegrityRepository,
	linker *phase151PromoteThreadIntegrityLinker,
	projectID uuid.UUID,
	actorID uuid.UUID,
	channelID uuid.UUID,
	rootID string,
) {
	t.Helper()
	assert.Equal(t, phase151PromoteThreadIntegrityIssueID.String(), response.IssueID)
	assert.Equal(t, "HUB-151", response.IssueKey)
	assert.Equal(t, []phase151PromoteThreadIntegrityTargetCall{{actorID: actorID, projectID: projectID}}, target.arguments)
	assert.Equal(t, []phase151PromoteThreadIntegrityPromoteCall{{
		projectID: projectID,
		actorID:   actorID,
		title:     "Thread integrity promotion",
		issueType: "task",
	}}, promoter.arguments)
	require.Len(t, repository.backlinks, 1)
	backlink := repository.backlinks[0]
	assert.Equal(t, channelID, backlink.ChannelID)
	assert.Equal(t, rootID, backlink.ThreadID)
	assert.Equal(t, actorID, backlink.SenderID)
	assert.Contains(t, backlink.ContentText, "HUB-151")
	assert.Equal(t, []phase151PromoteThreadIntegrityLinkCall{{
		projectID: projectID,
		issueID:   phase151PromoteThreadIntegrityIssueID,
		actorID:   actorID,
		channelID: channelID,
		rootID:    rootID,
	}}, linker.arguments)
}

func phase151AssertPromoteThreadIntegrityNoEffects(
	t *testing.T,
	repository *phase151PromoteThreadIntegrityRepository,
	promoter *phase151PromoteThreadIntegrityPromoter,
	linker *phase151PromoteThreadIntegrityLinker,
) {
	t.Helper()
	assert.Zero(t, repository.getChannelCalls)
	assert.Zero(t, promoter.calls)
	assert.Zero(t, repository.backlinkCalls)
	assert.Zero(t, linker.calls)
}
