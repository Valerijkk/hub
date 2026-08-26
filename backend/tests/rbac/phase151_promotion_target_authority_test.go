package rbac_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/projectmember"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/rbac"
)

// promotionTargetAuthorityRepo supplies only the new fresh authority boundary;
// embedded stubRepo keeps all unrelated RBAC persistence paths unusable.
type promotionTargetAuthorityRepo struct {
	*stubRepo
	facts rbac.PromotionTargetAuthorityFacts
	err   error

	resolveCalls int
	actorID      uuid.UUID
	projectID    uuid.UUID
}

func (r *promotionTargetAuthorityRepo) ResolveFreshPromotionTargetAuthority(
	_ context.Context,
	actorID, projectID uuid.UUID,
) (rbac.PromotionTargetAuthorityFacts, error) {
	r.resolveCalls++
	r.actorID = actorID
	r.projectID = projectID
	return r.facts, r.err
}

func newPromotionTargetAuthorityRBAC(t *testing.T, repo *promotionTargetAuthorityRepo) (*rbac.RBACService, *rbac.PermissionCache) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	cache := rbac.NewPermissionCache(rdb)
	return rbac.NewRBACService(repo, cache), cache
}

func assertFreshPromotionAuthorityLookup(t *testing.T, repo *promotionTargetAuthorityRepo, actorID, projectID uuid.UUID) {
	t.Helper()
	assert.Equal(t, 1, repo.resolveCalls, "promotion authorization must resolve fresh facts exactly once")
	assert.Equal(t, actorID, repo.actorID)
	assert.Equal(t, projectID, repo.projectID)
}

// Mutation: treating resolver-reported TargetFound=false as an existing
// project would disclose or authorize a target unavailable to this actor.
func TestAuthorizePromotionTarget_ReturnsNotFoundWhenResolverReportsTargetUnavailable(t *testing.T) {
	actorID := uuid.New()
	projectID := uuid.New()
	repo := &promotionTargetAuthorityRepo{
		stubRepo: &stubRepo{},
		facts:    rbac.PromotionTargetAuthorityFacts{TargetFound: false},
	}
	svc, _ := newPromotionTargetAuthorityRBAC(t, repo)

	targetFound, canCreate, err := svc.AuthorizePromotionTarget(context.Background(), actorID, projectID)

	require.NoError(t, err)
	assert.False(t, targetFound)
	assert.False(t, canCreate)
	assertFreshPromotionAuthorityLookup(t, repo, actorID, projectID)
}

// Mutation: accepting member data together with resolver-reported
// TargetFound=false creates internally contradictory authority facts.
func TestAuthorizePromotionTarget_FailsClosedForMemberOnUnavailableTargetFacts(t *testing.T) {
	actorID := uuid.New()
	projectID := uuid.New()
	repo := &promotionTargetAuthorityRepo{
		stubRepo: &stubRepo{},
		facts: rbac.PromotionTargetAuthorityFacts{
			TargetFound: false,
			Member:      &ent.ProjectMember{ProjectRole: projectmember.ProjectRoleDeveloper},
		},
	}
	svc, _ := newPromotionTargetAuthorityRBAC(t, repo)

	targetFound, canCreate, err := svc.AuthorizePromotionTarget(context.Background(), actorID, projectID)

	require.Error(t, err)
	assert.False(t, targetFound)
	assert.False(t, canCreate)
	assertFreshPromotionAuthorityLookup(t, repo, actorID, projectID)
}

// Mutation: allowing an ordinary user when resolver facts have no project
// membership grants task creation without an authority proof.
func TestAuthorizePromotionTarget_TargetFactsWithoutMembershipDeny(t *testing.T) {
	actorID := uuid.New()
	projectID := uuid.New()
	repo := &promotionTargetAuthorityRepo{
		stubRepo: &stubRepo{},
		facts: rbac.PromotionTargetAuthorityFacts{
			TargetFound:     true,
			ActorSystemRole: user.SystemRoleMember,
		},
	}
	svc, _ := newPromotionTargetAuthorityRBAC(t, repo)

	targetFound, canCreate, err := svc.AuthorizePromotionTarget(context.Background(), actorID, projectID)

	require.NoError(t, err)
	assert.True(t, targetFound)
	assert.False(t, canCreate)
	assertFreshPromotionAuthorityLookup(t, repo, actorID, projectID)
}

// Mutation: using the wrong default project-role map could allow viewers or
// deny developers and testers after a fresh authority lookup.
func TestAuthorizePromotionTarget_UsesFreshDefaultProjectRole(t *testing.T) {
	for _, tc := range []struct {
		name      string
		role      projectmember.ProjectRole
		canCreate bool
	}{
		{name: "developer can create", role: projectmember.ProjectRoleDeveloper, canCreate: true},
		{name: "tester can create", role: projectmember.ProjectRoleTester, canCreate: true},
		{name: "viewer cannot create", role: projectmember.ProjectRoleViewer, canCreate: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			actorID := uuid.New()
			projectID := uuid.New()
			repo := &promotionTargetAuthorityRepo{
				stubRepo: &stubRepo{},
				facts: rbac.PromotionTargetAuthorityFacts{
					TargetFound:     true,
					ActorSystemRole: user.SystemRoleMember,
					Member:          &ent.ProjectMember{ProjectRole: tc.role},
				},
			}
			svc, _ := newPromotionTargetAuthorityRBAC(t, repo)

			targetFound, canCreate, err := svc.AuthorizePromotionTarget(context.Background(), actorID, projectID)

			require.NoError(t, err)
			assert.True(t, targetFound)
			assert.Equal(t, tc.canCreate, canCreate)
			assertFreshPromotionAuthorityLookup(t, repo, actorID, projectID)
		})
	}
}

// Mutation: ignoring permission_names when present, or ignoring the legacy
// bitmask when permission_names is empty, breaks the documented role authority.
func TestAuthorizePromotionTarget_UsesCustomRolePermissionNamesAsAuthority(t *testing.T) {
	for _, tc := range []struct {
		name            string
		permissionNames []string
		legacyMask      uint64
		canCreate       bool
	}{
		{name: "names grant while legacy mask denies", permissionNames: []string{"task.create"}, legacyMask: 0, canCreate: true},
		{name: "names deny while legacy mask grants", permissionNames: []string{"project.view"}, legacyMask: uint64(rbac.PermTaskCreate), canCreate: false},
		{name: "legacy mask grants when names are empty", permissionNames: nil, legacyMask: uint64(rbac.PermTaskCreate), canCreate: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			actorID := uuid.New()
			projectID := uuid.New()
			repo := &promotionTargetAuthorityRepo{
				stubRepo: &stubRepo{},
				facts: rbac.PromotionTargetAuthorityFacts{
					TargetFound:     true,
					ActorSystemRole: user.SystemRoleMember,
					Member: &ent.ProjectMember{
						ProjectRole: projectmember.ProjectRoleViewer,
						Edges: ent.ProjectMemberEdges{CustomRole: &ent.Role{
							PermissionNames: tc.permissionNames,
							Permissions:     tc.legacyMask,
						}},
					},
				},
			}
			svc, _ := newPromotionTargetAuthorityRBAC(t, repo)

			targetFound, canCreate, err := svc.AuthorizePromotionTarget(context.Background(), actorID, projectID)

			require.NoError(t, err)
			assert.True(t, targetFound)
			assert.Equal(t, tc.canCreate, canCreate)
			assertFreshPromotionAuthorityLookup(t, repo, actorID, projectID)
		})
	}
}

// Mutation: requiring a project-member edge for founder/admin blocks the
// explicit system-role bypass after a live target has been proven.
func TestAuthorizePromotionTarget_AllowsSystemAdminWithoutMembership(t *testing.T) {
	for _, role := range []user.SystemRole{user.SystemRoleFounder, user.SystemRoleAdmin} {
		t.Run(string(role), func(t *testing.T) {
			actorID := uuid.New()
			projectID := uuid.New()
			repo := &promotionTargetAuthorityRepo{
				stubRepo: &stubRepo{},
				facts: rbac.PromotionTargetAuthorityFacts{
					TargetFound:     true,
					ActorSystemRole: role,
				},
			}
			svc, _ := newPromotionTargetAuthorityRBAC(t, repo)

			targetFound, canCreate, err := svc.AuthorizePromotionTarget(context.Background(), actorID, projectID)

			require.NoError(t, err)
			assert.True(t, targetFound)
			assert.True(t, canCreate)
			assertFreshPromotionAuthorityLookup(t, repo, actorID, projectID)
		})
	}
}

// Mutation: swallowing a fresh-repository failure can turn an unknown target
// authority state into an authorization decision.
func TestAuthorizePromotionTarget_PreservesFreshRepositoryError(t *testing.T) {
	actorID := uuid.New()
	projectID := uuid.New()
	repositoryErr := errors.New("fresh promotion target authority unavailable")
	repo := &promotionTargetAuthorityRepo{
		stubRepo: &stubRepo{},
		err:      repositoryErr,
	}
	svc, _ := newPromotionTargetAuthorityRBAC(t, repo)

	targetFound, canCreate, err := svc.AuthorizePromotionTarget(context.Background(), actorID, projectID)

	require.ErrorIs(t, err, repositoryErr)
	assert.False(t, targetFound)
	assert.False(t, canCreate)
	assertFreshPromotionAuthorityLookup(t, repo, actorID, projectID)
}

// Mutation: accepting a target-found record with no fresh actor-role proof
// turns malformed authority facts into a potentially privileged decision.
func TestAuthorizePromotionTarget_FailsClosedForMissingActorInvariant(t *testing.T) {
	actorID := uuid.New()
	projectID := uuid.New()
	repo := &promotionTargetAuthorityRepo{
		stubRepo: &stubRepo{},
		facts: rbac.PromotionTargetAuthorityFacts{
			TargetFound: true,
			Member:      &ent.ProjectMember{ProjectRole: projectmember.ProjectRoleDeveloper},
		},
	}
	svc, _ := newPromotionTargetAuthorityRBAC(t, repo)

	_, canCreate, err := svc.AuthorizePromotionTarget(context.Background(), actorID, projectID)

	require.Error(t, err)
	assert.False(t, canCreate)
	assertFreshPromotionAuthorityLookup(t, repo, actorID, projectID)
}

// Mutation: consulting PermissionCache/GetEffectivePermissions lets a cached
// allow outlive a freshly observed viewer denial.
func TestAuthorizePromotionTarget_FreshDenyOverridesCachedAllow(t *testing.T) {
	actorID := uuid.New()
	projectID := uuid.New()
	repo := &promotionTargetAuthorityRepo{
		stubRepo: &stubRepo{},
		facts: rbac.PromotionTargetAuthorityFacts{
			TargetFound:     true,
			ActorSystemRole: user.SystemRoleMember,
			Member:          &ent.ProjectMember{ProjectRole: projectmember.ProjectRoleViewer},
		},
	}
	svc, cache := newPromotionTargetAuthorityRBAC(t, repo)
	require.NoError(t, cache.Set(context.Background(), actorID.String(), projectID.String(), rbac.PermTaskCreate, time.Minute))

	targetFound, canCreate, err := svc.AuthorizePromotionTarget(context.Background(), actorID, projectID)

	require.NoError(t, err)
	assert.True(t, targetFound)
	assert.False(t, canCreate)
	assertFreshPromotionAuthorityLookup(t, repo, actorID, projectID)
}
