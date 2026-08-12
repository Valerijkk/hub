//go:build hubtest

package auth_test

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/internal/auth"
)

func TestWSTicket_IssueRedeemOneShot(t *testing.T) {
	dsn := "file:wsticket_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	magic := auth.NewMagicLinkService(rdb, nil, "test@example.com")
	svc := auth.NewAuthService(auth.NewEntRepository(client), rdb, nil, "ws-ticket-secret", 15*time.Minute, 30*24*time.Hour, magic, nil)

	uid := uuid.New().String()
	sid := uuid.New().String()
	ticket, expiresIn, err := svc.IssueWSTicket(t.Context(), uid, sid)
	require.NoError(t, err)
	assert.Equal(t, auth.WSTicketTTLSeconds(), expiresIn)
	assert.NotEmpty(t, ticket)

	claims, err := svc.RedeemWSTicket(t.Context(), ticket)
	require.NoError(t, err)
	assert.Equal(t, uid, claims.UserID)
	assert.Equal(t, sid, claims.SessionID)

	_, err = svc.RedeemWSTicket(t.Context(), ticket)
	require.Error(t, err)
}
