//go:build qa && !release

package qamailcalendar

import (
	"context"
	"errors"

	"entgo.io/ent/dialect/sql"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/notification"
	"github.com/valerii/hub/backend/ent/predicate"
	"github.com/valerii/hub/backend/ent/session"
	"github.com/valerii/hub/backend/ent/user"
)

var errFixtureAuthSideEffectsCleanup = errors.New("qa mail calendar fixture auth side effects cleanup failed")

// deleteFixtureAuthSideEffectsLocked deletes only the locked fixture owner's
// login notifications and sessions. The caller owns the owner FOR UPDATE lock
// and the surrounding transaction.
func deleteFixtureAuthSideEffectsLocked(ctx context.Context, tx *ent.Tx, ownerID uuid.UUID) error {
	if ctx == nil || tx == nil || ownerID == uuid.Nil {
		return errFixtureAuthSideEffectsCleanup
	}

	notifications, err := tx.Notification.Query().
		Where(
			notification.UserIDEQ(ownerID),
			fixtureAuthSideEffectsNotificationNoWait(),
		).
		Order(ent.Asc(notification.FieldID)).
		All(ctx)
	if err != nil {
		return errFixtureAuthSideEffectsCleanup
	}

	notificationIDs := make([]uuid.UUID, 0, len(notifications))
	seenNotificationIDs := make(map[uuid.UUID]struct{}, len(notifications))
	for _, row := range notifications {
		if row == nil ||
			row.ID == uuid.Nil ||
			row.UserID != ownerID ||
			row.Type != "new_login" ||
			row.EntityType != "session" ||
			row.EntityID == nil ||
			*row.EntityID == uuid.Nil {
			return errFixtureAuthSideEffectsCleanup
		}

		entityID, err := uuid.Parse(row.EntityID.String())
		if err != nil || entityID == uuid.Nil || entityID != *row.EntityID {
			return errFixtureAuthSideEffectsCleanup
		}
		if _, duplicate := seenNotificationIDs[row.ID]; duplicate {
			return errFixtureAuthSideEffectsCleanup
		}
		seenNotificationIDs[row.ID] = struct{}{}
		notificationIDs = append(notificationIDs, row.ID)
	}

	if len(notificationIDs) > 0 {
		deleted, err := tx.Notification.Delete().
			Where(
				notification.IDIn(notificationIDs...),
				notification.UserIDEQ(ownerID),
			).
			Exec(ctx)
		if err != nil || deleted != len(notificationIDs) {
			return errFixtureAuthSideEffectsCleanup
		}
	}

	sessions, err := tx.Session.Query().
		Where(
			session.HasUserWith(user.IDEQ(ownerID)),
			fixtureAuthSideEffectsSessionNoWait(),
		).
		Order(ent.Asc(session.FieldID)).
		All(ctx)
	if err != nil {
		return errFixtureAuthSideEffectsCleanup
	}

	sessionIDs := make([]uuid.UUID, 0, len(sessions))
	seenSessionIDs := make(map[uuid.UUID]struct{}, len(sessions))
	for _, row := range sessions {
		if row == nil || row.ID == uuid.Nil {
			return errFixtureAuthSideEffectsCleanup
		}
		if _, duplicate := seenSessionIDs[row.ID]; duplicate {
			return errFixtureAuthSideEffectsCleanup
		}
		seenSessionIDs[row.ID] = struct{}{}
		sessionIDs = append(sessionIDs, row.ID)
	}

	if len(sessionIDs) > 0 {
		deleted, err := tx.Session.Delete().
			Where(
				session.IDIn(sessionIDs...),
				session.HasUserWith(user.IDEQ(ownerID)),
			).
			Exec(ctx)
		if err != nil || deleted != len(sessionIDs) {
			return errFixtureAuthSideEffectsCleanup
		}
	}

	notificationsRemain, err := tx.Notification.Query().
		Where(notification.UserIDEQ(ownerID)).
		Exist(ctx)
	if err != nil || notificationsRemain {
		return errFixtureAuthSideEffectsCleanup
	}

	sessionsRemain, err := tx.Session.Query().
		Where(session.HasUserWith(user.IDEQ(ownerID))).
		Exist(ctx)
	if err != nil || sessionsRemain {
		return errFixtureAuthSideEffectsCleanup
	}

	return nil
}

func fixtureAuthSideEffectsNotificationNoWait() predicate.Notification {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func fixtureAuthSideEffectsSessionNoWait() predicate.Session {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}
