package messenger_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valerii/hub/backend/internal/messenger"
)

func phase151PostCreateProductionConstructorDependencies() (
	*mockRepository,
	messenger.ChannelAccessResolver,
	messenger.AuthorizedUnreadScopeResolver,
	messenger.AuthorizedMessageSearchScopeResolver,
	*messenger.Phase151PostCreateRepository,
	*phase151CompositionMentions,
	*phase151CompositionInvalidation,
) {
	repository := newMockRepository()
	access := &phase151ThreadAccessResolver{grants: make(map[phase151ThreadAccessKey]messenger.ChannelAccess)}
	unreadScope := &phase151UnreadScopeFixture{}
	searchScope := &phase151MessageSearchScopeFixture{}
	mentions := &phase151CompositionMentions{}
	invalidation := &phase151CompositionInvalidation{}
	postCreate := messenger.NewPhase151PostCreateRepository(
		phase151CompositionCombinedRepository(),
		mentions,
		invalidation,
	)
	return repository, access, unreadScope, searchScope, postCreate, mentions, invalidation
}

func TestPhase151MessengerServiceWithAuthorityAndPostCreateStoresExactConcreteCompositionWithoutInfrastructureIO(t *testing.T) {
	repository, access, unreadScope, searchScope, postCreate, mentions, invalidation :=
		phase151PostCreateProductionConstructorDependencies()

	service := messenger.NewMessengerServiceWithAuthorityAndPostCreate(
		repository,
		nil,
		nil,
		access,
		unreadScope,
		searchScope,
		postCreate,
	)

	stored := messenger.ExportMessengerServicePostCreate(service)
	assert.IsType(t, &messenger.Phase151PostCreateRepository{}, stored)
	assert.Same(t, postCreate, stored)
	assert.Zero(t, repository.createMessageCalls)
	assert.Empty(t, mentions.calls)
	assert.Empty(t, invalidation.calls)
}

func TestPhase151MessengerServiceWithAuthorityAndPostCreateRejectsNilAndTypedNilDependencies(t *testing.T) {
	repository, access, unreadScope, searchScope, postCreate, _, _ :=
		phase151PostCreateProductionConstructorDependencies()

	assert.Panics(t, func() {
		messenger.NewMessengerServiceWithAuthorityAndPostCreate(
			repository, nil, nil, access, unreadScope, searchScope, nil,
		)
	})
	var typedNil *messenger.Phase151PostCreateRepository
	assert.Panics(t, func() {
		messenger.NewMessengerServiceWithAuthorityAndPostCreate(
			repository, nil, nil, access, unreadScope, searchScope, typedNil,
		)
	})

	legacy := messenger.NewMessengerServiceWithAuthority(
		repository, nil, nil, access, unreadScope, searchScope,
	)
	assert.Nil(t, messenger.ExportMessengerServicePostCreate(legacy))
	assert.Nil(t, messenger.ExportMessengerServicePostCreate(nil))
	assert.NotNil(t, postCreate)
}
