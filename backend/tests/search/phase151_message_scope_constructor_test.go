package search_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/messenger"
	"github.com/valerii/hub/backend/internal/search"
)

type phase151AuthorizedMessageScopeResolverFixture struct{}

func (*phase151AuthorizedMessageScopeResolverFixture) ResolveAuthorizedMessageSearchScope(
	context.Context,
	uuid.UUID,
) (messenger.AuthorizedMessageSearchScope, error) {
	return messenger.AuthorizedMessageSearchScope{}, nil
}

func TestPhase151NewSearchServiceWithAuthorizedMessageScopeRequiresResolver(t *testing.T) {
	assert.Panics(t, func() {
		search.NewSearchServiceWithAuthorizedMessageSearchScope(nil, nil, nil, nil)
	})

	service := search.NewSearchServiceWithAuthorizedMessageSearchScope(
		nil,
		nil,
		nil,
		&phase151AuthorizedMessageScopeResolverFixture{},
	)
	require.NotNil(t, service)
}
