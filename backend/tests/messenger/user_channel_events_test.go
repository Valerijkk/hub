package messenger_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/messenger"
)

// lifecycleFailureRepository keeps mutation-failure coverage local to this
// producer unit. It deliberately reuses the repository fixture and its
// promoted methods instead of widening the shared test mock's surface.
type lifecycleFailureRepository struct {
	*mockRepository
	createChannelErr      error
	updateChannelErr      error
	getChannelMemberIDsErr error
	removeMemberErr       error
}

func (r *lifecycleFailureRepository) CreateChannel(ctx context.Context, input messenger.CreateChannelInput) (*ent.Channel, error) {
	if r.createChannelErr != nil {
		return nil, r.createChannelErr
	}
	return r.mockRepository.CreateChannel(ctx, input)
}

func (r *lifecycleFailureRepository) UpdateChannel(ctx context.Context, channelID uuid.UUID, input messenger.UpdateChannelInput) (*ent.Channel, error) {
	if r.updateChannelErr != nil {
		return nil, r.updateChannelErr
	}
	return r.mockRepository.UpdateChannel(ctx, channelID, input)
}

func (r *lifecycleFailureRepository) GetChannelMemberIDs(ctx context.Context, channelID uuid.UUID) ([]uuid.UUID, error) {
	if r.getChannelMemberIDsErr != nil {
		return nil, r.getChannelMemberIDsErr
	}
	return r.mockRepository.GetChannelMemberIDs(ctx, channelID)
}

func (r *lifecycleFailureRepository) RemoveMember(ctx context.Context, channelID, userID uuid.UUID) error {
	if r.removeMemberErr != nil {
		return r.removeMemberErr
	}
	return r.mockRepository.RemoveMember(ctx, channelID, userID)
}

// internalLifecycleRepository records only the persistence boundary needed by
// the internal-membership and DM lifecycle contracts. Keeping it in this
// producer test avoids broadening the shared messenger mock for a single
// event-delivery concern.
type internalLifecycleRepository struct {
	*mockRepository

	createChannelErr error
	findDMErr        error
	snapshotErr      error
	addErr           error

	createChannelCalls int
	createdChannelIDs  []uuid.UUID
	snapshotCalls      int
	addMemberInputs    [][]uuid.UUID
	deleteChannelIDs   []uuid.UUID
}

func (r *internalLifecycleRepository) CreateChannel(ctx context.Context, input messenger.CreateChannelInput) (*ent.Channel, error) {
	r.createChannelCalls++
	if r.createChannelErr != nil {
		return nil, r.createChannelErr
	}
	channel, err := r.mockRepository.CreateChannel(ctx, input)
	if err == nil {
		r.createdChannelIDs = append(r.createdChannelIDs, channel.ID)
	}
	return channel, err
}
func (r *internalLifecycleRepository) FindDMChannel(ctx context.Context, dmKey string) (*ent.Channel, error) {
	if r.findDMErr != nil {
		return nil, r.findDMErr
	}
	return r.mockRepository.FindDMChannel(ctx, dmKey)
}

func (r *internalLifecycleRepository) GetChannelMemberIDs(ctx context.Context, channelID uuid.UUID) ([]uuid.UUID, error) {
	r.snapshotCalls++
	if r.snapshotErr != nil {
		return nil, r.snapshotErr
	}
	return r.mockRepository.GetChannelMemberIDs(ctx, channelID)
}

func (r *internalLifecycleRepository) AddMembers(ctx context.Context, channelID uuid.UUID, memberIDs []uuid.UUID, role string) error {
	r.addMemberInputs = append(r.addMemberInputs, append([]uuid.UUID(nil), memberIDs...))
	if r.addErr != nil {
		return r.addErr
	}
	return r.mockRepository.AddMembers(ctx, channelID, memberIDs, role)
}
func (r *internalLifecycleRepository) DeleteChannel(ctx context.Context, channelID uuid.UUID) error {
	r.deleteChannelIDs = append(r.deleteChannelIDs, channelID)
	return r.mockRepository.DeleteChannel(ctx, channelID)
}
func TestUserChannelLifecycleCreatePublishesCreatedForSuccessfullyPersistedMembers(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	creatorID := uuid.New()
	inviteeID := uuid.New()
	outsiderID := uuid.New()
	publisher := &capturePublisher{}
	service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)

	response, err := service.CreateChannel(ctx, creatorID, messenger.CreateChannelRequest{
		ChannelType: "public",
		Name:        "lifecycle-created",
		MemberIDs: []string{
			inviteeID.String(),
			inviteeID.String(),
			creatorID.String(),
			"not-a-uuid",
		},
	})
	require.NoError(t, err)
	require.NotNil(t, response)

	requireUserChannelLifecycleEvents(t, publisher, "channel_created", response.ID, creatorID, inviteeID)
	requireNoUserChannelLifecycleEventFor(t, publisher, outsiderID)
}

func TestUserChannelLifecycleUpdatePublishesUpdatedForCurrentMembers(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	ownerID := uuid.New()
	memberID := uuid.New()
	outsiderID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID, memberID)
	publisher := &capturePublisher{}
	service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)

	response, err := service.UpdateChannel(ctx, channelID, ownerID, messenger.CreateChannelRequest{
		Name:    "lifecycle-renamed",
		Version: 1,
	})
	require.NoError(t, err)
	require.NotNil(t, response)

	requireUserChannelLifecycleEvents(t, publisher, "channel_updated", channelID.String(), ownerID, memberID)
	requireNoUserChannelLifecycleEventFor(t, publisher, outsiderID)
}

func TestUserChannelLifecycleDeletePublishesRemovedForPreDeleteMembers(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	ownerID := uuid.New()
	memberID := uuid.New()
	outsiderID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID, memberID)
	publisher := &capturePublisher{}
	service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)

	require.NoError(t, service.DeleteChannel(ctx, channelID, ownerID))

	requireUserChannelLifecycleEvents(t, publisher, "channel_removed", channelID.String(), ownerID, memberID)
	requireNoUserChannelLifecycleEventFor(t, publisher, outsiderID)
}

func TestUserChannelLifecycleJoinPublishesCreatedOnlyForNewMember(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	ownerID := uuid.New()
	joiningID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)
	publisher := &capturePublisher{}
	service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)

	require.NoError(t, service.JoinChannel(ctx, channelID, joiningID))
	require.NoError(t, service.JoinChannel(ctx, channelID, joiningID))

	requireUserChannelLifecycleEvents(t, publisher, "channel_created", channelID.String(), joiningID)
	requireNoUserChannelLifecycleEventFor(t, publisher, ownerID)
}

func TestUserChannelLifecycleAddMembersPublishesCreatedOnlyForActuallyNewMembers(t *testing.T) {
	ctx := context.Background()

	t.Run("success deduplicates input and excludes existing members", func(t *testing.T) {
		repo := newMockRepository()
		ownerID := uuid.New()
		existingID := uuid.New()
		newMemberA := uuid.New()
		newMemberB := uuid.New()
		channelID := setupChannelWithMembers(repo, "private", false, ownerID, existingID)
		publisher := &capturePublisher{}
		service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)

		err := service.AddMembers(ctx, channelID, ownerID, []string{
			existingID.String(),
			newMemberA.String(),
			newMemberA.String(),
			newMemberB.String(),
			newMemberB.String(),
		})
		require.NoError(t, err)

		requireUserChannelLifecycleEvents(t, publisher, "channel_created", channelID.String(), newMemberA, newMemberB)
		requireNoUserChannelLifecycleEventFor(t, publisher, existingID)
		requireNoUserChannelLifecycleEventFor(t, publisher, ownerID)
	})

	t.Run("nil member IDs are rejected atomically", func(t *testing.T) {
		testCases := []struct {
			name       string
			memberIDs  func(uuid.UUID) []string
		}{
			{
				name: "nil only",
				memberIDs: func(uuid.UUID) []string {
					return []string{uuid.Nil.String()}
				},
			},
			{
				name: "valid new member followed by nil",
				memberIDs: func(validNew uuid.UUID) []string {
					return []string{validNew.String(), uuid.Nil.String()}
				},
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				repo := newMockRepository()
				ownerID := uuid.New()
				validNewID := uuid.New()
				channelID := setupChannelWithMembers(repo, "private", false, ownerID)
				publisher := &capturePublisher{}
				service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)
				addMembersCallsBefore := repo.addMembersCalls

				err := service.AddMembers(ctx, channelID, ownerID, testCase.memberIDs(validNewID))
				require.Error(t, err)
				requireAppErrorCode(t, err, 400)
				require.Equal(t, addMembersCallsBefore, repo.addMembersCalls, "invalid batch must not reach repository AddMembers")
				require.NotContains(t, repo.members[channelID], validNewID, "mixed invalid batch must not partially add its valid member")
				require.NotContains(t, repo.members[channelID], uuid.Nil, "nil UUID must never be persisted as a member")
				requireNoUserChannelLifecycleEvents(t, publisher)
			})
		}
	})

	t.Run("membership mutation failure publishes no user lifecycle event", func(t *testing.T) {
		repo := newMockRepository()
		ownerID := uuid.New()
		channelID := setupChannelWithMembers(repo, "private", false, ownerID)
		repo.addMembersErr = errors.New("add members failed")
		publisher := &capturePublisher{}
		service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)

		err := service.AddMembers(ctx, channelID, ownerID, []string{uuid.NewString()})
		require.Error(t, err)
		requireNoUserChannelLifecycleEvents(t, publisher)
	})
}

func TestUserChannelLifecycleAddMembersInternalPublishesCreatedOnlyForNewMembers(t *testing.T) {
	ctx := context.Background()

	t.Run("deduplicates in caller order, excludes current members, and publishes after persistence", func(t *testing.T) {
		base := newMockRepository()
		ownerID := uuid.New()
		existingID := uuid.New()
		newMemberA := uuid.New()
		newMemberB := uuid.New()
		channelID := setupChannelWithMembers(base, "private", false, ownerID, existingID)
		repo := &internalLifecycleRepository{mockRepository: base}
		publisher := &capturePublisher{}
		service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)

		err := service.AddMembersInternal(ctx, channelID, []uuid.UUID{
			existingID,
			newMemberA,
			newMemberA,
			newMemberB,
			newMemberA,
			newMemberB,
		})
		require.NoError(t, err)
		require.Equal(t, 1, repo.snapshotCalls, "internal membership must snapshot current members before writing")
		require.Equal(t, [][]uuid.UUID{{newMemberA, newMemberB}}, repo.addMemberInputs, "only stable, newly absent members may reach persistence")
		requireUserChannelLifecycleEvents(t, publisher, "channel_created", channelID.String(), newMemberA, newMemberB)
		requireNoUserChannelLifecycleEventFor(t, publisher, existingID)
		requireNoUserChannelLifecycleEventFor(t, publisher, ownerID)
	})

	t.Run("rejects nil channel and member IDs before any read or write", func(t *testing.T) {
		testCases := []struct {
			name      string
			channelID uuid.UUID
			memberIDs func(uuid.UUID) []uuid.UUID
		}{
			{
				name:      "nil channel",
				channelID: uuid.Nil,
				memberIDs: func(validMemberID uuid.UUID) []uuid.UUID { return []uuid.UUID{validMemberID} },
			},
			{
				name:      "nil member",
				channelID: uuid.New(),
				memberIDs: func(uuid.UUID) []uuid.UUID { return []uuid.UUID{uuid.Nil} },
			},
			{
				name:      "mixed valid and nil member",
				channelID: uuid.New(),
				memberIDs: func(validMemberID uuid.UUID) []uuid.UUID { return []uuid.UUID{validMemberID, uuid.Nil} },
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				base := newMockRepository()
				repo := &internalLifecycleRepository{mockRepository: base}
				publisher := &capturePublisher{}
				service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)

				err := service.AddMembersInternal(ctx, testCase.channelID, testCase.memberIDs(uuid.New()))
				require.Error(t, err)
				requireAppErrorCode(t, err, 400)
				require.Zero(t, repo.snapshotCalls, "invalid internal membership input must not query state")
				require.Empty(t, repo.addMemberInputs, "invalid internal membership input must not reach persistence")
				requireNoUserChannelLifecycleEvents(t, publisher)
			})
		}
	})

	t.Run("empty and entirely current inputs are idempotent no-ops", func(t *testing.T) {
		base := newMockRepository()
		ownerID := uuid.New()
		existingID := uuid.New()
		channelID := setupChannelWithMembers(base, "private", false, ownerID, existingID)
		repo := &internalLifecycleRepository{mockRepository: base}
		publisher := &capturePublisher{}
		service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)

		require.NoError(t, service.AddMembersInternal(ctx, channelID, nil))
		require.Zero(t, repo.snapshotCalls, "empty input must not perform an unnecessary snapshot")
		require.Empty(t, repo.addMemberInputs, "empty input must not write")

		require.NoError(t, service.AddMembersInternal(ctx, channelID, []uuid.UUID{existingID, existingID}))
		require.Equal(t, 1, repo.snapshotCalls, "non-empty input must compare against the current membership snapshot")
		require.Empty(t, repo.addMemberInputs, "current members must not be re-written")
		requireNoUserChannelLifecycleEvents(t, publisher)
	})

	t.Run("snapshot or write failures publish no lifecycle event", func(t *testing.T) {
		t.Run("snapshot failure is fail closed", func(t *testing.T) {
			base := newMockRepository()
			channelID := setupChannelWithMembers(base, "private", false, uuid.New())
			repo := &internalLifecycleRepository{
				mockRepository: base,
				snapshotErr:     errors.New("snapshot failed"),
			}
			publisher := &capturePublisher{}
			service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)

			err := service.AddMembersInternal(ctx, channelID, []uuid.UUID{uuid.New()})
			require.Error(t, err)
			require.Equal(t, 1, repo.snapshotCalls)
			require.Empty(t, repo.addMemberInputs)
			requireNoUserChannelLifecycleEvents(t, publisher)
		})

		t.Run("write failure publishes nothing", func(t *testing.T) {
			base := newMockRepository()
			channelID := setupChannelWithMembers(base, "private", false, uuid.New())
			repo := &internalLifecycleRepository{
				mockRepository: base,
				addErr:         errors.New("add members failed"),
			}
			publisher := &capturePublisher{}
			service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)

			err := service.AddMembersInternal(ctx, channelID, []uuid.UUID{uuid.New()})
			require.Error(t, err)
			require.Equal(t, 1, repo.snapshotCalls)
			require.Len(t, repo.addMemberInputs, 1)
			requireNoUserChannelLifecycleEvents(t, publisher)
		})
	})
}

func TestUserChannelLifecycleCreateOrGetDMPublishesCreatedOnlyForNewDM(t *testing.T) {
	ctx := context.Background()

	t.Run("existing DM returns without a new lifecycle event", func(t *testing.T) {
		base := newMockRepository()
		userID := uuid.New()
		targetUserID := uuid.New()
		dmKey := canonicalDMKey(userID, targetUserID)
		existing, err := base.CreateChannel(ctx, messenger.CreateChannelInput{
			ChannelType: "dm",
			CreatedBy:   userID,
			DMKey:       &dmKey,
		})
		require.NoError(t, err)
		require.NoError(t, base.AddMembers(ctx, existing.ID, []uuid.UUID{userID, targetUserID}, "member"))
		repo := &internalLifecycleRepository{mockRepository: base}
		publisher := &capturePublisher{}
		service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)

		response, err := service.CreateOrGetDM(ctx, userID, targetUserID)
		require.NoError(t, err)
		require.NotNil(t, response)
		require.Equal(t, existing.ID.String(), response.ID)
		require.Zero(t, repo.createChannelCalls, "existing DM must not be created again")
		require.Empty(t, repo.addMemberInputs, "existing DM must not mutate membership")
		requireNoUserChannelLifecycleEvents(t, publisher)
	})

	t.Run("new DM publishes exactly once to its persisted participants", func(t *testing.T) {
		base := newMockRepository()
		userID := uuid.New()
		targetUserID := uuid.New()
		repo := &internalLifecycleRepository{mockRepository: base}
		publisher := &capturePublisher{}
		service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)

		response, err := service.CreateOrGetDM(ctx, userID, targetUserID)
		require.NoError(t, err)
		require.NotNil(t, response)
		require.Equal(t, 1, repo.createChannelCalls)
		require.Equal(t, [][]uuid.UUID{{userID, targetUserID}}, repo.addMemberInputs)
		requireUserChannelLifecycleEvents(t, publisher, "channel_created", response.ID, userID, targetUserID)

		retry, err := service.CreateOrGetDM(ctx, targetUserID, userID)
		require.NoError(t, err)
		require.NotNil(t, retry)
		require.Equal(t, response.ID, retry.ID)
		require.Equal(t, 1, repo.createChannelCalls, "idempotent DM retry must not create another channel")
		require.Equal(t, [][]uuid.UUID{{userID, targetUserID}}, repo.addMemberInputs, "idempotent DM retry must not write membership")
		requireUserChannelLifecycleEvents(t, publisher, "channel_created", response.ID, userID, targetUserID)
	})

	t.Run("lookup, create, or membership persistence failure publishes no lifecycle event", func(t *testing.T) {
		testCases := []struct {
			name      string
			configure func(*internalLifecycleRepository)
		}{
			{
				name: "lookup failure",
				configure: func(repo *internalLifecycleRepository) {
					repo.findDMErr = errors.New("lookup failed")
				},
			},
			{
				name: "create failure",
				configure: func(repo *internalLifecycleRepository) {
					repo.createChannelErr = errors.New("create failed")
				},
			},
			{
				name: "membership persistence failure",
				configure: func(repo *internalLifecycleRepository) {
					repo.addErr = errors.New("membership persistence failed")
				},
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				base := newMockRepository()
				repo := &internalLifecycleRepository{mockRepository: base}
				testCase.configure(repo)
				publisher := &capturePublisher{}
				service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)

				response, err := service.CreateOrGetDM(ctx, uuid.New(), uuid.New())
				require.Error(t, err)
				require.Nil(t, response)
				if testCase.name == "lookup failure" {
					require.Zero(t, repo.createChannelCalls, "failed DM lookup must not attempt channel creation")
					require.Empty(t, repo.addMemberInputs, "failed DM lookup must not write membership")
				}
				if testCase.name == "membership persistence failure" {
					require.Len(t, repo.createdChannelIDs, 1)
					require.Equal(t, repo.createdChannelIDs, repo.deleteChannelIDs, "failed DM membership persistence must delete its created channel exactly once")
				}
				requireNoUserChannelLifecycleEvents(t, publisher)
			})
		}
	})
}

func canonicalDMKey(userID, targetUserID uuid.UUID) string {
	left, right := userID.String(), targetUserID.String()
	if right < left {
		left, right = right, left
	}
	return left + "-" + right
}
func TestUserChannelLifecycleAcceptInvitePublishesCreatedOnlyForNewMember(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	ownerID := uuid.New()
	invitedID := uuid.New()
	channelID := setupChannelWithMembers(repo, "private", false, ownerID)
	_, err := repo.CreateChannelInvite(ctx, messenger.CreateChannelInviteInput{
		ChannelID: channelID,
		CreatorID: ownerID,
		Token:     "lifecycle-invite-token",
	})
	require.NoError(t, err)
	publisher := &capturePublisher{}
	service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)

	response, err := service.AcceptInvite(ctx, "lifecycle-invite-token", invitedID)
	require.NoError(t, err)
	require.NotNil(t, response)
	response, err = service.AcceptInvite(ctx, "lifecycle-invite-token", invitedID)
	require.NoError(t, err)
	require.NotNil(t, response)

	requireUserChannelLifecycleEvents(t, publisher, "channel_created", channelID.String(), invitedID)
	requireNoUserChannelLifecycleEventFor(t, publisher, ownerID)
}

func TestUserChannelLifecycleAcceptInviteFiniteRetryIsIdempotentForNowMember(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	ownerID := uuid.New()
	invitedID := uuid.New()
	channelID := setupChannelWithMembers(repo, "private", false, ownerID)
	invite, err := repo.CreateChannelInvite(ctx, messenger.CreateChannelInviteInput{
		ChannelID: channelID,
		CreatorID: ownerID,
		Token:     "lifecycle-finite-invite-token",
		MaxUses:   1,
	})
	require.NoError(t, err)
	publisher := &capturePublisher{}
	service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)

	response, err := service.AcceptInvite(ctx, invite.Token, invitedID)
	require.NoError(t, err)
	require.NotNil(t, response)
	require.Equal(t, 1, invite.UseCount)

	retry, err := service.AcceptInvite(ctx, invite.Token, invitedID)
	require.NoError(t, err)
	require.NotNil(t, retry)
	require.Equal(t, response.ID, retry.ID)
	require.Equal(t, 1, invite.UseCount, "idempotent retry must not consume another invite use")
	requireUserChannelLifecycleEvents(t, publisher, "channel_created", channelID.String(), invitedID)
}

func TestUserChannelLifecycleAcceptInvitePreexistingMemberBypassesConsumableValidation(t *testing.T) {
	testCases := []struct {
		name   string
		mutate func(*ent.ChannelInvite)
	}{
		{
			name: "expired invite",
			mutate: func(invite *ent.ChannelInvite) {
				expiredAt := time.Now().UTC().Add(-time.Minute)
				invite.ExpiresAt = &expiredAt
			},
		},
		{
			name: "revoked invite",
			mutate: func(invite *ent.ChannelInvite) {
				invite.IsRevoked = true
			},
		},
		{
			name: "max-used invite",
			mutate: func(invite *ent.ChannelInvite) {
				invite.MaxUses = 1
				invite.UseCount = 1
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			repo := newMockRepository()
			ownerID := uuid.New()
			memberID := uuid.New()
			channelID := setupChannelWithMembers(repo, "private", false, ownerID, memberID)
			invite, err := repo.CreateChannelInvite(ctx, messenger.CreateChannelInviteInput{
				ChannelID: channelID,
				CreatorID: ownerID,
				Token:     "lifecycle-preexisting-" + strings.ReplaceAll(testCase.name, " ", "-"),
			})
			require.NoError(t, err)
			testCase.mutate(invite)
			publisher := &capturePublisher{}
			service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)

			response, err := service.AcceptInvite(ctx, invite.Token, memberID)
			require.NoError(t, err)
			require.NotNil(t, response)
			require.Equal(t, channelID.String(), response.ID)
			requireNoUserChannelLifecycleEvents(t, publisher)
		})
	}
}

func TestUserChannelLifecycleLeavePublishesRemovedOnlyForLeaver(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	ownerID := uuid.New()
	leavingID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID, leavingID)
	publisher := &capturePublisher{}
	service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)

	require.NoError(t, service.LeaveChannel(ctx, channelID, leavingID))

	requireUserChannelLifecycleEvents(t, publisher, "channel_removed", channelID.String(), leavingID)
	requireNoUserChannelLifecycleEventFor(t, publisher, ownerID)
}

func TestUserChannelLifecycleMutationFailuresPublishNoEvents(t *testing.T) {
	ctx := context.Background()

	t.Run("create channel owner persistence failure rolls back before publication", func(t *testing.T) {
		repo := newMockRepository()
		repo.addMembersErr = errors.New("persist owner failed")
		publisher := &capturePublisher{}
		service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)

		response, err := service.CreateChannel(ctx, uuid.New(), messenger.CreateChannelRequest{ChannelType: "public", Name: "fail-owner"})
		require.Error(t, err)
		require.Nil(t, response)
		requireNoUserChannelLifecycleEvents(t, publisher)
		require.Len(t, repo.channels, 1)
		for _, channel := range repo.channels {
			require.NotNil(t, channel.DeletedAt, "channel must be rolled back when owner persistence fails")
		}
	})

	t.Run("create channel invited-members persistence failure rolls back before publication", func(t *testing.T) {
		repo := newMockRepository()
		repo.addMembersErrExceptOwner = errors.New("persist invited members failed")
		publisher := &capturePublisher{}
		service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)

		response, err := service.CreateChannel(ctx, uuid.New(), messenger.CreateChannelRequest{
			ChannelType: "public",
			Name:        "fail-invitees",
			MemberIDs:   []string{uuid.NewString()},
		})
		require.Error(t, err)
		require.Nil(t, response)
		requireNoUserChannelLifecycleEvents(t, publisher)
		require.Len(t, repo.channels, 1)
		for _, channel := range repo.channels {
			require.NotNil(t, channel.DeletedAt, "channel must be rolled back when invited-member persistence fails")
		}
	})

	t.Run("create channel repository failure", func(t *testing.T) {
		publisher := &capturePublisher{}
		repo := &lifecycleFailureRepository{
			mockRepository:    newMockRepository(),
			createChannelErr: errors.New("create channel failed"),
		}
		service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)

		_, err := service.CreateChannel(ctx, uuid.New(), messenger.CreateChannelRequest{ChannelType: "public", Name: "fail-create"})
		require.Error(t, err)
		requireNoUserChannelLifecycleEvents(t, publisher)
	})

	t.Run("update channel repository failure", func(t *testing.T) {
		base := newMockRepository()
		ownerID := uuid.New()
		channelID := setupChannelWithMembers(base, "public", false, ownerID)
		publisher := &capturePublisher{}
		repo := &lifecycleFailureRepository{
			mockRepository:    base,
			updateChannelErr: errors.New("update channel failed"),
		}
		service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)

		_, err := service.UpdateChannel(ctx, channelID, ownerID, messenger.CreateChannelRequest{Name: "fail-update", Version: 1})
		require.Error(t, err)
		requireNoUserChannelLifecycleEvents(t, publisher)
	})

	t.Run("delete channel member snapshot failure is fail closed", func(t *testing.T) {
		base := newMockRepository()
		ownerID := uuid.New()
		memberID := uuid.New()
		channelID := setupChannelWithMembers(base, "public", false, ownerID, memberID)
		publisher := &capturePublisher{}
		repo := &lifecycleFailureRepository{
			mockRepository:          base,
			getChannelMemberIDsErr: errors.New("snapshot members failed"),
		}
		service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)

		err := service.DeleteChannel(ctx, channelID, ownerID)
		require.Error(t, err)
		require.Nil(t, base.channels[channelID].DeletedAt, "delete must not proceed without a pre-delete member snapshot")
		requireNoUserChannelLifecycleEvents(t, publisher)
	})

	t.Run("delete channel repository failure", func(t *testing.T) {
		repo := newMockRepository()
		ownerID := uuid.New()
		channelID := setupChannelWithMembers(repo, "public", false, ownerID)
		repo.deleteChannelErr = errors.New("delete channel failed")
		publisher := &capturePublisher{}
		service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)

		err := service.DeleteChannel(ctx, channelID, ownerID)
		require.Error(t, err)
		requireNoUserChannelLifecycleEvents(t, publisher)
	})

	t.Run("join channel membership mutation failure", func(t *testing.T) {
		repo := newMockRepository()
		ownerID := uuid.New()
		joiningID := uuid.New()
		channelID := setupChannelWithMembers(repo, "public", false, ownerID)
		repo.addMembersErr = errors.New("join member failed")
		publisher := &capturePublisher{}
		service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)

		err := service.JoinChannel(ctx, channelID, joiningID)
		require.Error(t, err)
		requireNoUserChannelLifecycleEvents(t, publisher)
	})

	t.Run("accept invite membership mutation failure", func(t *testing.T) {
		repo := newMockRepository()
		ownerID := uuid.New()
		invitedID := uuid.New()
		channelID := setupChannelWithMembers(repo, "private", false, ownerID)
		_, err := repo.CreateChannelInvite(ctx, messenger.CreateChannelInviteInput{
			ChannelID: channelID,
			CreatorID: ownerID,
			Token:     "lifecycle-failing-invite-token",
		})
		require.NoError(t, err)
		repo.addMembersErr = errors.New("accept invite member failed")
		publisher := &capturePublisher{}
		service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)

		_, err = service.AcceptInvite(ctx, "lifecycle-failing-invite-token", invitedID)
		require.Error(t, err)
		requireNoUserChannelLifecycleEvents(t, publisher)
	})

	t.Run("leave channel membership mutation failure", func(t *testing.T) {
		base := newMockRepository()
		ownerID := uuid.New()
		leavingID := uuid.New()
		channelID := setupChannelWithMembers(base, "public", false, ownerID, leavingID)
		publisher := &capturePublisher{}
		repo := &lifecycleFailureRepository{
			mockRepository: base,
			removeMemberErr:  errors.New("leave member failed"),
		}
		service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)

		err := service.LeaveChannel(ctx, channelID, leavingID)
		require.Error(t, err)
		requireNoUserChannelLifecycleEvents(t, publisher)
	})
}

func requireUserChannelLifecycleEvents(t *testing.T, publisher *capturePublisher, eventType, channelID string, expectedUserIDs ...uuid.UUID) {
	t.Helper()
	expected := make(map[string]struct{}, len(expectedUserIDs))
	for _, userID := range expectedUserIDs {
		expected[userID.String()] = struct{}{}
	}

	seen := make(map[string]int, len(expected))
	for _, publication := range publisher.snapshot() {
		if !strings.HasPrefix(publication.subject, "messenger.user.") {
			continue
		}
		userID, ok := userIDFromLifecycleSubject(publication.subject, eventType)
		require.Truef(t, ok, "unexpected user lifecycle subject: %s", publication.subject)
		require.Containsf(t, expected, userID, "unexpected recipient for %s", eventType)
		seen[userID]++
		requireLifecyclePayload(t, publication.payload, channelID)
	}

	require.Len(t, seen, len(expected), "every expected user receives one %s event", eventType)
	for userID := range expected {
		require.Equalf(t, 1, seen[userID], "user %s must receive exactly one %s event", userID, eventType)
	}
}

func requireNoUserChannelLifecycleEventFor(t *testing.T, publisher *capturePublisher, userID uuid.UUID) {
	t.Helper()
	prefix := fmt.Sprintf("messenger.user.%s.", userID.String())
	for _, publication := range publisher.snapshot() {
		require.Falsef(t, strings.HasPrefix(publication.subject, prefix), "unexpected lifecycle publication to %s: %s", userID, publication.subject)
	}
}

func requireNoUserChannelLifecycleEvents(t *testing.T, publisher *capturePublisher) {
	t.Helper()
	for _, publication := range publisher.snapshot() {
		require.Falsef(t, strings.HasPrefix(publication.subject, "messenger.user."), "mutation failure must not publish user lifecycle event: %s", publication.subject)
	}
}

func userIDFromLifecycleSubject(subject, eventType string) (string, bool) {
	const prefix = "messenger.user."
	suffix := "." + eventType
	if !strings.HasPrefix(subject, prefix) || !strings.HasSuffix(subject, suffix) {
		return "", false
	}
	userID := strings.TrimSuffix(strings.TrimPrefix(subject, prefix), suffix)
	if _, err := uuid.Parse(userID); err != nil {
		return "", false
	}
	return userID, true
}

func requireLifecyclePayload(t *testing.T, payload any, expectedChannelID string) {
	t.Helper()
	raw, err := json.Marshal(payload)
	require.NoError(t, err)

	var object map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(raw, &object))
	require.Len(t, object, 1, "lifecycle payload must contain only channel_id")
	channelRaw, ok := object["channel_id"]
	require.True(t, ok, "lifecycle payload must contain channel_id")

	var actualChannelID string
	require.NoError(t, json.Unmarshal(channelRaw, &actualChannelID))
	require.Equal(t, expectedChannelID, actualChannelID)
}
