package messenger_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/messenger"
)

type phase151RecoveryTargetFactSource struct {
	ids      []uuid.UUID
	orgErr   error
	facts    messenger.Phase151PendingRecoveryTargetFacts
	factsErr         error
	calls            []string
	organizationCtx  context.Context
	ctx              context.Context
	tenantID         uuid.UUID
	channelID        uuid.UUID
}

var _ messenger.Phase151PendingRecoveryTargetFactSource = (*phase151RecoveryTargetFactSource)(nil)

func (s *phase151RecoveryTargetFactSource) ListInstanceOrganizationIDs(ctx context.Context) ([]uuid.UUID, error) {
	s.calls = append(s.calls, "organization")
	s.organizationCtx = ctx
	if s.orgErr != nil {
		return nil, s.orgErr
	}
	return append([]uuid.UUID(nil), s.ids...), nil
}

func (s *phase151RecoveryTargetFactSource) ResolvePendingRecoveryTargetFacts(
	ctx context.Context,
	tenantID uuid.UUID,
	channelID uuid.UUID,
) (messenger.Phase151PendingRecoveryTargetFacts, error) {
	s.calls = append(s.calls, "facts")
	s.ctx = ctx
	s.tenantID = tenantID
	s.channelID = channelID
	if s.factsErr != nil {
		return messenger.Phase151PendingRecoveryTargetFacts{}, s.factsErr
	}
	return s.facts, nil
}

type phase151TypedNilContext struct{}

func (*phase151TypedNilContext) Deadline() (time.Time, bool) { return time.Time{}, false }
func (*phase151TypedNilContext) Done() <-chan struct{}       { return nil }
func (*phase151TypedNilContext) Err() error                  { return nil }
func (*phase151TypedNilContext) Value(any) any               { return nil }

func phase151RecoveryTargetFixture(channelType, kind string) (
	messenger.Phase151PendingRecoveryTarget,
	*phase151RecoveryTargetFactSource,
) {
	tenantID := uuid.New()
	channelID := uuid.New()
	target := messenger.Phase151PendingRecoveryTarget{
		TenantID: tenantID,
		ChannelID: channelID,
		Kind:     kind,
	}
	return target, &phase151RecoveryTargetFactSource{
		ids: []uuid.UUID{tenantID},
		facts: messenger.Phase151PendingRecoveryTargetFacts{
			TenantID:               tenantID,
			ChannelID:              channelID,
			ChannelType:            channelType,
			IsChannelLive:          true,
			HasActiveTenantProject: true,
			HasTenantCreator:       true,
		},
	}
}

func requireZeroRecoveryAuthority(t *testing.T, authority messenger.Phase151RecoveryTargetAuthority) {
	t.Helper()
	assert.Equal(t, messenger.Phase151RecoveryTargetAuthority{}, authority)
}

func TestPGPhase151PendingRecoveryTargetReauthorizerFactsContract(t *testing.T) {
	typeOfFacts := reflect.TypeOf(messenger.Phase151PendingRecoveryTargetFacts{})
	fieldNames := make([]string, typeOfFacts.NumField())
	for index := range fieldNames {
		fieldNames[index] = typeOfFacts.Field(index).Name
	}
	assert.Equal(t, []string{
		"TenantID", "ChannelID", "ChannelType", "IsChannelLive", "IsProjectChannel",
		"HasActiveTenantProject", "HasTenantCreator",
	}, fieldNames)
}

func TestPGPhase151PendingRecoveryTargetReauthorizerMapsOnlyCanonicalChannelTypes(t *testing.T) {
	for _, tt := range []struct {
		name        string
		channelType string
		kind        string
	}{
		{name: "public channel", channelType: "public", kind: "channel"},
		{name: "private channel", channelType: "private", kind: "channel"},
		{name: "direct message", channelType: "dm", kind: "dm"},
		{name: "group direct message", channelType: "group", kind: "dm"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			target, source := phase151RecoveryTargetFixture(tt.channelType, tt.kind)
			reauthorizer := messenger.NewPGPhase151PendingRecoveryTargetReauthorizer(source)
			ctx := context.WithValue(context.Background(), "phase151-recovery-context", tt.name)

			authority, err := reauthorizer.ReauthorizePendingRecoveryTarget(ctx, target)

			require.NoError(t, err)
			assert.Equal(t, messenger.Phase151RecoveryTargetAuthority{
				Principal: messenger.Phase151PostCreateRecoveryServicePrincipal,
				TenantID:  target.TenantID,
				ChannelID: target.ChannelID,
				Kind:      target.Kind,
			}, authority)
			assert.Equal(t, []string{"organization", "facts"}, source.calls)
			assert.Same(t, ctx, source.organizationCtx)
			assert.Same(t, ctx, source.ctx)
			assert.Equal(t, target.TenantID, source.tenantID)
			assert.Equal(t, target.ChannelID, source.channelID)
		})
	}

	t.Run("project private channel with active tenant project", func(t *testing.T) {
		target, source := phase151RecoveryTargetFixture("private", "channel")
		source.facts.IsProjectChannel = true
		source.facts.HasActiveTenantProject = true
		source.facts.HasTenantCreator = false
		reauthorizer := messenger.NewPGPhase151PendingRecoveryTargetReauthorizer(source)
		ctx := context.WithValue(context.Background(), "phase151-recovery-context", "project-channel")

		authority, err := reauthorizer.ReauthorizePendingRecoveryTarget(ctx, target)

		require.NoError(t, err)
		assert.Equal(t, messenger.Phase151RecoveryTargetAuthority{
			Principal: messenger.Phase151PostCreateRecoveryServicePrincipal,
			TenantID:  target.TenantID,
			ChannelID: target.ChannelID,
			Kind:      "channel",
		}, authority)
		assert.Equal(t, []string{"organization", "facts"}, source.calls)
		assert.Same(t, ctx, source.organizationCtx)
		assert.Same(t, ctx, source.ctx)
		assert.Equal(t, target.TenantID, source.tenantID)
		assert.Equal(t, target.ChannelID, source.channelID)
	})
}

func TestPGPhase151PendingRecoveryTargetReauthorizerRejectsInvalidInputWithoutCalls(t *testing.T) {
	validTarget, _ := phase151RecoveryTargetFixture("public", "channel")
	for _, tt := range []struct {
		name   string
		mutate func(*messenger.Phase151PendingRecoveryTarget)
	}{
		{name: "zero tenant", mutate: func(target *messenger.Phase151PendingRecoveryTarget) { target.TenantID = uuid.Nil }},
		{name: "zero channel", mutate: func(target *messenger.Phase151PendingRecoveryTarget) { target.ChannelID = uuid.Nil }},
		{name: "empty kind", mutate: func(target *messenger.Phase151PendingRecoveryTarget) { target.Kind = "" }},
		{name: "whitespace kind", mutate: func(target *messenger.Phase151PendingRecoveryTarget) { target.Kind = " " }},
		{name: "padded channel kind", mutate: func(target *messenger.Phase151PendingRecoveryTarget) { target.Kind = " channel" }},
		{name: "case drifted channel kind", mutate: func(target *messenger.Phase151PendingRecoveryTarget) { target.Kind = "Channel" }},
		{name: "padded dm kind", mutate: func(target *messenger.Phase151PendingRecoveryTarget) { target.Kind = "dm " }},
		{name: "case drifted dm kind", mutate: func(target *messenger.Phase151PendingRecoveryTarget) { target.Kind = "DM" }},
		{name: "unknown kind", mutate: func(target *messenger.Phase151PendingRecoveryTarget) { target.Kind = "future_kind" }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			target := validTarget
			tt.mutate(&target)
			_, source := phase151RecoveryTargetFixture("public", "channel")
			reauthorizer := messenger.NewPGPhase151PendingRecoveryTargetReauthorizer(source)

			authority, err := reauthorizer.ReauthorizePendingRecoveryTarget(context.Background(), target)

			require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
			requireZeroRecoveryAuthority(t, authority)
			assert.Empty(t, source.calls)
		})
	}
}

func TestPGPhase151PendingRecoveryTargetReauthorizerFailsClosedBeforeTargetOrSourceUse(t *testing.T) {
	target, source := phase151RecoveryTargetFixture("public", "channel")
	var typedNilSource *phase151RecoveryTargetFactSource
	assert.Nil(t, messenger.NewPGPhase151PendingRecoveryTargetReauthorizer(nil))
	assert.Nil(t, messenger.NewPGPhase151PendingRecoveryTargetReauthorizer(typedNilSource))

	var nilReauthorizer *messenger.PGPhase151PendingRecoveryTargetReauthorizer
	authority, err := nilReauthorizer.ReauthorizePendingRecoveryTarget(context.Background(), target)
	require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
	requireZeroRecoveryAuthority(t, authority)

	reauthorizer := messenger.NewPGPhase151PendingRecoveryTargetReauthorizer(source)
	for _, ctx := range []context.Context{nil, (*phase151TypedNilContext)(nil)} {
		authority, err = reauthorizer.ReauthorizePendingRecoveryTarget(ctx, target)
		require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
		requireZeroRecoveryAuthority(t, authority)
	}
	assert.Empty(t, source.calls)

	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	malformed := target
	malformed.Kind = "bad"
	authority, err = reauthorizer.ReauthorizePendingRecoveryTarget(cancelled, malformed)
	require.ErrorIs(t, err, context.Canceled)
	requireZeroRecoveryAuthority(t, authority)
	assert.Empty(t, source.calls)
}

func TestPGPhase151PendingRecoveryTargetReauthorizerRejectsSingletonAndTenantFailuresBeforeFacts(t *testing.T) {
	target, _ := phase151RecoveryTargetFixture("public", "channel")
	lookupErr := errors.New("phase151 singleton lookup failed")
	for _, tt := range []struct {
		name      string
		ids       []uuid.UUID
		orgErr    error
		want      error
		wantCalls []string
	}{
		{name: "zero organizations", want: messenger.ErrInstanceOrganizationInvalid, wantCalls: []string{"organization"}},
		{name: "two organizations", ids: []uuid.UUID{target.TenantID, uuid.New()}, want: messenger.ErrInstanceOrganizationInvalid, wantCalls: []string{"organization"}},
		{name: "nil organization", ids: []uuid.UUID{uuid.Nil}, want: messenger.ErrInstanceOrganizationInvalid, wantCalls: []string{"organization"}},
		{name: "lookup error identity", orgErr: lookupErr, want: lookupErr, wantCalls: []string{"organization"}},
		{name: "target tenant mismatch", ids: []uuid.UUID{uuid.New()}, want: messenger.ErrPostCreatePipelineInvalid, wantCalls: []string{"organization"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			source := &phase151RecoveryTargetFactSource{
				ids: tt.ids, orgErr: tt.orgErr,
				facts: messenger.Phase151PendingRecoveryTargetFacts{TenantID: target.TenantID, ChannelID: target.ChannelID, IsChannelLive: true, ChannelType: "public"},
			}
			reauthorizer := messenger.NewPGPhase151PendingRecoveryTargetReauthorizer(source)

			authority, err := reauthorizer.ReauthorizePendingRecoveryTarget(context.Background(), target)

			require.ErrorIs(t, err, tt.want)
			requireZeroRecoveryAuthority(t, authority)
			assert.Equal(t, tt.wantCalls, source.calls)
		})
	}
}

func TestPGPhase151PendingRecoveryTargetReauthorizerRejectsFactFailures(t *testing.T) {
	factErr := errors.New("phase151 pending recovery facts failed")
	for _, tt := range []struct {
		name   string
		kind   string
		mutate func(*messenger.Phase151PendingRecoveryTargetFacts)
		want   error
	}{
		{name: "fact error identity", kind: "channel", want: factErr},
		{name: "tenant drift", kind: "channel", mutate: func(facts *messenger.Phase151PendingRecoveryTargetFacts) { facts.TenantID = uuid.New() }, want: messenger.ErrPostCreatePipelineInvalid},
		{name: "channel drift", kind: "channel", mutate: func(facts *messenger.Phase151PendingRecoveryTargetFacts) { facts.ChannelID = uuid.New() }, want: messenger.ErrPostCreatePipelineInvalid},
		{name: "inactive channel", kind: "channel", mutate: func(facts *messenger.Phase151PendingRecoveryTargetFacts) { facts.IsChannelLive = false }, want: messenger.ErrPostCreatePipelineInvalid},
		{name: "unknown type", kind: "channel", mutate: func(facts *messenger.Phase151PendingRecoveryTargetFacts) { facts.ChannelType = "unknown" }, want: messenger.ErrPostCreatePipelineInvalid},
		{name: "padded type", kind: "channel", mutate: func(facts *messenger.Phase151PendingRecoveryTargetFacts) { facts.ChannelType = " public" }, want: messenger.ErrPostCreatePipelineInvalid},
		{name: "case drifted type", kind: "channel", mutate: func(facts *messenger.Phase151PendingRecoveryTargetFacts) { facts.ChannelType = "Public" }, want: messenger.ErrPostCreatePipelineInvalid},
		{name: "channel mapped to dm target", kind: "dm", want: messenger.ErrPostCreatePipelineInvalid},
		{name: "dm mapped to channel target", kind: "channel", mutate: func(facts *messenger.Phase151PendingRecoveryTargetFacts) { facts.ChannelType = "dm" }, want: messenger.ErrPostCreatePipelineInvalid},
		{name: "project without active tenant project", kind: "channel", mutate: func(facts *messenger.Phase151PendingRecoveryTargetFacts) { facts.IsProjectChannel = true; facts.HasActiveTenantProject = false }, want: messenger.ErrPostCreatePipelineInvalid},
		{name: "projectless without tenant creator", kind: "channel", mutate: func(facts *messenger.Phase151PendingRecoveryTargetFacts) { facts.HasTenantCreator = false }, want: messenger.ErrPostCreatePipelineInvalid},
	} {
		t.Run(tt.name, func(t *testing.T) {
			target, source := phase151RecoveryTargetFixture("public", tt.kind)
			if tt.name == "fact error identity" {
				source.factsErr = factErr
			}
			if tt.mutate != nil {
				tt.mutate(&source.facts)
			}
			reauthorizer := messenger.NewPGPhase151PendingRecoveryTargetReauthorizer(source)

			authority, err := reauthorizer.ReauthorizePendingRecoveryTarget(context.Background(), target)

			require.ErrorIs(t, err, tt.want)
			requireZeroRecoveryAuthority(t, authority)
			assert.Equal(t, []string{"organization", "facts"}, source.calls)
		})
	}
}
