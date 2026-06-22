package onec

import (
	"context"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
)

// EntUserLookup загружает контактные поля пользователя через Ent.
type EntUserLookup struct {
	db *ent.Client
}

// NewEntUserLookup создаёт lookup пользователя.
func NewEntUserLookup(db *ent.Client) *EntUserLookup {
	return &EntUserLookup{db: db}
}

// GetUserEmailPhone возвращает email, телефон и отображаемое имя для маппинга в 1С.
func (l *EntUserLookup) GetUserEmailPhone(ctx context.Context, userID uuid.UUID) (email, phone, displayName string, err error) {
	u, err := l.db.User.Get(ctx, userID)
	if err != nil {
		return "", "", "", err
	}
	displayName = u.DisplayName
	if displayName == "" {
		displayName = u.Name
	}
	return u.Email, u.Phone, displayName, nil
}
