package messenger

import (
	"context"

	"github.com/google/uuid"
)

type AuthorizedMessageSearchTarget struct {
	ChannelID       uuid.UUID
	Kind            string
	IsProjectChannel bool
}

type AuthorizedMessageSearchScope struct {
	TenantID uuid.UUID
	UserID   uuid.UUID
	Targets  []AuthorizedMessageSearchTarget
}

type AuthorizedMessageSearchScopeResolver interface {
	ResolveAuthorizedMessageSearchScope(context.Context, uuid.UUID) (AuthorizedMessageSearchScope, error)
}
