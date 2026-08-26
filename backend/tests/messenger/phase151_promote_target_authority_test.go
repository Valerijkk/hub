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
)

type phase151PromoteTargetAuthoritySourceCall struct {
	channelID uuid.UUID
	messageID string
}

type phase151PromoteTargetAuthorityTargetCall struct {
	actorID   uuid.UUID
	projectID uuid.UUID
}

type phase151PromoteTargetAuthorityRepository struct {
	*phase151PromoteSourceBindingRepositoryBase
	order              *[]string
	sourceDocument     *messenger.MessageDocument
	sourceFound        bool
	sourceErr          error
	sourceCalls        []phase151PromoteTargetAuthoritySourceCall
	getChannelCalls    int
	genericParentCalls int
	repliesCalls       int
	backlinkCalls      int
}

func (r *phase151PromoteTargetAuthorityRepository) record(operation string) {
	*r.order = append(*r.order, operation)
}

func (r *phase151PromoteTargetAuthorityRepository) FindLiveMessageInChannel(
	_ context.Context,
	channelID uuid.UUID,
	messageID string,
) (*messenger.MessageDocument, bool, error) {
	r.record("scoped-live-source")
	r.sourceCalls = append(r.sourceCalls, phase151PromoteTargetAuthoritySourceCall{
		channelID: channelID,
		messageID: messageID,
	})
	return r.sourceDocument, r.sourceFound, r.sourceErr
}

func (r *phase151PromoteTargetAuthorityRepository) GetChannel(
	ctx context.Context,
	channelID uuid.UUID,
) (*ent.Channel, error) {
	r.record("get-channel")
	r.getChannelCalls++
	return r.mockRepository.GetChannel(ctx, channelID)
}

func (r *phase151PromoteTargetAuthorityRepository) FindMessageByID(
	ctx context.Context,
	messageID string,
) (*messenger.MessageDocument, error) {
	r.record("generic-parent")
	r.genericParentCalls++
	return r.mockRepository.FindMessageByID(ctx, messageID)
}

func (r *phase151PromoteTargetAuthorityRepository) ListThreadReplies(
	ctx context.Context,
	threadID string,
	before *time.Time,
	limit int,
) ([]*messenger.MessageDocument, error) {
	r.record("list-replies")
	r.repliesCalls++
	return r.mockRepository.ListThreadReplies(ctx, threadID, before, limit)
}

func (r *phase151PromoteTargetAuthorityRepository) InsertMessage(
	ctx context.Context,
	document *messenger.MessageDocument,
) (string, error) {
	r.record("backlink")
	r.backlinkCalls++
	return r.mockRepository.InsertMessage(ctx, document)
}

type phase151PromoteTargetAuthorityChecker struct {
	order     *[]string
	found     bool
	canCreate bool
	err       error
	calls     []phase151PromoteTargetAuthorityTargetCall
}

func (c *phase151PromoteTargetAuthorityChecker) AuthorizePromotionTarget(
	_ context.Context,
	actorID uuid.UUID,
	projectID uuid.UUID,
) (targetFound bool, canCreate bool, err error) {
	*c.order = append(*c.order, "target-authority")
	c.calls = append(c.calls, phase151PromoteTargetAuthorityTargetCall{
		actorID: actorID,
		projectID: projectID,
	})
	return c.found, c.canCreate, c.err
}

type phase151PromoteTargetAuthorityPromoter struct {
	order *[]string
	calls int
}

func (p *phase151PromoteTargetAuthorityPromoter) CreateIssueFromPromotion(
	_ context.Context,
	_ uuid.UUID,
	_ uuid.UUID,
	_ string,
	_ string,
	_ map[string]interface{},
) (string, string, error) {
	*p.order = append(*p.order, "promoter")
	p.calls++
	return "HUB-151", uuid.NewString(), nil
}

type phase151PromoteTargetAuthorityLinker struct {
	order *[]string
	calls int
}

func (l *phase151PromoteTargetAuthorityLinker) LinkThreadToIssue(
	_ context.Context,
	_ uuid.UUID,
	_ uuid.UUID,
	_ uuid.UUID,
	_ uuid.UUID,
	_ string,
) error {
	*l.order = append(*l.order, "linker")
	l.calls++
	return nil
}

func TestPhase151PromoteToIssueTargetAuthorityConcealsAndStopsBeforeDownstream(t *testing.T) {
	ctx := context.Background()

	for _, scenario := range []struct {
		name         string
		checker      *phase151PromoteTargetAuthorityChecker
		wantCode     int
		wantTargetLog []string
	}{
		{
			name:         "missing target is concealed",
			checker:      &phase151PromoteTargetAuthorityChecker{found: false, canCreate: false},
			wantCode:     404,
			wantTargetLog: []string{"source-action", "scoped-live-source", "target-authority"},
		},
		{
			name:         "found target without create authority is forbidden",
			checker:      &phase151PromoteTargetAuthorityChecker{found: true, canCreate: false},
			wantCode:     403,
			wantTargetLog: []string{"source-action", "scoped-live-source", "target-authority"},
		},
		{
			name:         "target authority error is internal",
			checker:      &phase151PromoteTargetAuthorityChecker{err: errors.New("phase151 target authority error")},
			wantCode:     500,
			wantTargetLog: []string{"source-action", "scoped-live-source", "target-authority"},
		},
		{
			name:         "target authority invariant violation is internal",
			checker:      &phase151PromoteTargetAuthorityChecker{found: false, canCreate: true},
			wantCode:     500,
			wantTargetLog: []string{"source-action", "scoped-live-source", "target-authority"},
		},
		{
			name:         "missing target authority dependency is internal",
			checker:      nil,
			wantCode:     500,
			wantTargetLog: []string{"source-action", "scoped-live-source"},
		},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			order := make([]string, 0, 8)
			repository, actorID, channelID, channel, source, projectID := phase151PromoteTargetAuthorityFixture(&order)
			checker := scenario.checker
			if checker != nil {
				checker.order = &order
			}
			service, resolver, promoter, linker := phase151PromoteTargetAuthorityService(
				repository,
				actorID,
				channel,
				&order,
				checker,
			)

			response, err := service.PromoteToIssue(ctx, channelID, source.ID, actorID, phase151PromoteTargetAuthorityRequest(projectID))

			require.Nil(t, response)
			require.Error(t, err)
			requireAppErrorCode(t, err, scenario.wantCode)
			phase151AssertPromoteTargetSourceAuthority(t, resolver, repository, actorID, channelID, source.ID)
			if checker != nil {
				assert.Equal(t, []phase151PromoteTargetAuthorityTargetCall{{
					actorID: actorID, projectID: projectID,
				}}, checker.calls)
			}
			phase151AssertPromoteTargetNoDownstream(t, repository, promoter, linker)
			assert.Equal(t, scenario.wantTargetLog, order)
		})
	}
}

func TestPhase151PromoteToIssueStopsBeforeTargetAuthorityWhenSourceFails(t *testing.T) {
	ctx := context.Background()

	for _, scenario := range []struct {
		name               string
		configureSource    func(*phase151PromoteTargetAuthorityRepository, *phase151ChannelAccessResolver, uuid.UUID, uuid.UUID)
		wantCode           int
		wantSourceCalls    int
		wantSourceOrderLog []string
	}{
		{
			name: "source action is denied",
			configureSource: func(
				_ *phase151PromoteTargetAuthorityRepository,
				resolver *phase151ChannelAccessResolver,
				actorID uuid.UUID,
				channelID uuid.UUID,
			) {
				delete(resolver.grants, phase151ChannelAccessKey{
					actorID: actorID, channelID: channelID, action: messenger.ChannelActionPromote,
				})
			},
			wantCode:           403,
			wantSourceCalls:    0,
			wantSourceOrderLog: []string{"source-action"},
		},
		{
			name: "scoped live source is missing",
			configureSource: func(
				repository *phase151PromoteTargetAuthorityRepository,
				_ *phase151ChannelAccessResolver,
				_ uuid.UUID,
				_ uuid.UUID,
			) {
				repository.sourceFound = false
			},
			wantCode:           404,
			wantSourceCalls:    1,
			wantSourceOrderLog: []string{"source-action", "scoped-live-source"},
		},
		{
			name: "scoped live source lookup fails",
			configureSource: func(
				repository *phase151PromoteTargetAuthorityRepository,
				_ *phase151ChannelAccessResolver,
				_ uuid.UUID,
				_ uuid.UUID,
			) {
				repository.sourceErr = errors.New("phase151 scoped source failure")
			},
			wantCode:           500,
			wantSourceCalls:    1,
			wantSourceOrderLog: []string{"source-action", "scoped-live-source"},
		},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			order := make([]string, 0, 8)
			repository, actorID, channelID, channel, source, projectID := phase151PromoteTargetAuthorityFixture(&order)
			checker := &phase151PromoteTargetAuthorityChecker{
				order: &order, found: true, canCreate: true,
			}
			service, resolver, promoter, linker := phase151PromoteTargetAuthorityService(
				repository,
				actorID,
				channel,
				&order,
				checker,
			)
			scenario.configureSource(repository, resolver, actorID, channelID)

			response, err := service.PromoteToIssue(ctx, channelID, source.ID, actorID, phase151PromoteTargetAuthorityRequest(projectID))

			require.Nil(t, response)
			require.Error(t, err)
			requireAppErrorCode(t, err, scenario.wantCode)
			assert.Len(t, repository.sourceCalls, scenario.wantSourceCalls)
			assert.Empty(t, checker.calls, "source failure must not probe target authority")
			phase151AssertPromoteTargetNoDownstream(t, repository, promoter, linker)
			assert.Equal(t, scenario.wantSourceOrderLog, order)
		})
	}
}

func TestPhase151PromoteToIssueValidatesProjectBeforeSourceAndTargetAuthority(t *testing.T) {
	ctx := context.Background()
	order := make([]string, 0, 8)
	repository, actorID, channelID, channel, source, _ := phase151PromoteTargetAuthorityFixture(&order)
	checker := &phase151PromoteTargetAuthorityChecker{
		order: &order, found: true, canCreate: true,
	}
	service, resolver, promoter, linker := phase151PromoteTargetAuthorityService(
		repository,
		actorID,
		channel,
		&order,
		checker,
	)

	response, err := service.PromoteToIssue(ctx, channelID, source.ID, actorID, messenger.PromoteToIssueRequest{
		ProjectID: "not-a-project-uuid",
		Title:     "Must stop before authority",
	})

	require.Nil(t, response)
	require.Error(t, err)
	requireAppErrorCode(t, err, 400)
	assert.Empty(t, resolver.calls)
	assert.Empty(t, repository.sourceCalls)
	assert.Empty(t, checker.calls)
	phase151AssertPromoteTargetNoDownstream(t, repository, promoter, linker)
	assert.Empty(t, order)
}

func TestPhase151PromoteToIssueAllowsAuthorizedRootSourceAfterTargetAuthority(t *testing.T) {
	ctx := context.Background()
	order := make([]string, 0, 8)
	repository, actorID, channelID, channel, source, projectID := phase151PromoteTargetAuthorityFixture(&order)
	checker := &phase151PromoteTargetAuthorityChecker{
		order: &order, found: true, canCreate: true,
	}
	service, resolver, promoter, linker := phase151PromoteTargetAuthorityService(
		repository,
		actorID,
		channel,
		&order,
		checker,
	)

	response, err := service.PromoteToIssue(ctx, channelID, source.ID, actorID, phase151PromoteTargetAuthorityRequest(projectID))

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, "HUB-151", response.IssueKey)
	phase151AssertPromoteTargetSourceAuthority(t, resolver, repository, actorID, channelID, source.ID)
	assert.Equal(t, []phase151PromoteTargetAuthorityTargetCall{{
		actorID: actorID, projectID: projectID,
	}}, checker.calls)
	assert.Equal(t, 1, repository.getChannelCalls)
	assert.Zero(t, repository.genericParentCalls, "a root source is already the parent")
	assert.Equal(t, 1, repository.repliesCalls)
	assert.Equal(t, 1, promoter.calls)
	assert.Equal(t, 1, linker.calls)
	assert.Equal(t, 1, repository.backlinkCalls)
	assert.Equal(t, []string{
		"source-action",
		"scoped-live-source",
		"target-authority",
		"get-channel",
		"list-replies",
		"promoter",
		"backlink",
		"linker",
	}, order)
}

func phase151PromoteTargetAuthorityFixture(
	order *[]string,
) (*phase151PromoteTargetAuthorityRepository, uuid.UUID, uuid.UUID, *ent.Channel, *messenger.MessageDocument, uuid.UUID) {
	repository := &phase151PromoteTargetAuthorityRepository{
		phase151PromoteSourceBindingRepositoryBase: newPhase151PromoteSourceBindingRepositoryBase(),
		order:                                      order,
	}
	actorID := uuid.New()
	channelID := setupChannelWithMembers(repository.mockRepository, "private", false, actorID)
	channel := repository.channels[channelID]
	source := &messenger.MessageDocument{
		ID:          "phase151-promote-target-authority-source-" + uuid.NewString(),
		ChannelID:   channelID,
		SenderID:    actorID,
		ContentText: "target authority root source",
		CreatedAt:   time.Now().UTC(),
	}
	repository.messages[source.ID] = source
	repository.sourceDocument = source
	repository.sourceFound = true
	return repository, actorID, channelID, channel, source, uuid.New()
}

func phase151PromoteTargetAuthorityService(
	repository *phase151PromoteTargetAuthorityRepository,
	actorID uuid.UUID,
	channel *ent.Channel,
	order *[]string,
	checker *phase151PromoteTargetAuthorityChecker,
) (*messenger.MessengerService, *phase151ChannelAccessResolver, *phase151PromoteTargetAuthorityPromoter, *phase151PromoteTargetAuthorityLinker) {
	resolver := &phase151ChannelAccessResolver{
		grants: make(map[phase151ChannelAccessKey]messenger.ChannelAccess),
		onAuthorize: func(phase151ChannelAccessKey) {
			*order = append(*order, "source-action")
		},
	}
	access := phase151ReadableAccess(channel, true)
	access.CanPromote = true
	phase151Grant(resolver, actorID, channel, messenger.ChannelActionPromote, access)

	promoter := &phase151PromoteTargetAuthorityPromoter{order: order}
	linker := &phase151PromoteTargetAuthorityLinker{order: order}
	service := messenger.ExportNewMessengerServiceWithChannelAccessTest(repository, resolver)
	service.SetIssuePromoter(promoter)
	service.SetIssueThreadLinker(linker)
	if checker != nil {
		service.SetPromotePermissionChecker(checker)
	}
	return service, resolver, promoter, linker
}

func phase151PromoteTargetAuthorityRequest(projectID uuid.UUID) messenger.PromoteToIssueRequest {
	return messenger.PromoteToIssueRequest{
		ProjectID: " " + projectID.String() + " ",
		Title:     "Target authority promotion",
	}
}

func phase151AssertPromoteTargetSourceAuthority(
	t *testing.T,
	resolver *phase151ChannelAccessResolver,
	repository *phase151PromoteTargetAuthorityRepository,
	actorID uuid.UUID,
	channelID uuid.UUID,
	messageID string,
) {
	t.Helper()
	assert.Equal(t, []phase151ChannelAccessKey{{
		actorID: actorID, channelID: channelID, action: messenger.ChannelActionPromote,
	}}, resolver.calls)
	assert.Equal(t, []phase151PromoteTargetAuthoritySourceCall{{
		channelID: channelID, messageID: messageID,
	}}, repository.sourceCalls)
}

func phase151AssertPromoteTargetNoDownstream(
	t *testing.T,
	repository *phase151PromoteTargetAuthorityRepository,
	promoter *phase151PromoteTargetAuthorityPromoter,
	linker *phase151PromoteTargetAuthorityLinker,
) {
	t.Helper()
	assert.Zero(t, repository.getChannelCalls)
	assert.Zero(t, repository.genericParentCalls)
	assert.Zero(t, repository.repliesCalls)
	assert.Zero(t, promoter.calls)
	assert.Zero(t, linker.calls)
	assert.Zero(t, repository.backlinkCalls)
}
