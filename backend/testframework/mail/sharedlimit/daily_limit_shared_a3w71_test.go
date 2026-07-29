package sharedlimit_test

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/mail"
)

// TestService_SendAlert_SharedDailyLimit_A3W7_1_001 — два Service на одном Redis
// не суммируют лимит выше DailyLimit; restart (новый Service) не обнуляет окно.
// Отдельный пакет без enttest — устойчив к параллельной генерации Ent.
func TestService_SendAlert_SharedDailyLimit_A3W7_1_001(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	slots := mail.NewRedisDailySlotStore(rdb)
	ctx := context.Background()
	req := mail.SendRequest{To: "shared@test.local", Subject: "cap"}

	a := mail.NewService(mail.Config{DailyLimit: 2})
	a.SetDailySlotStore(slots)
	b := mail.NewService(mail.Config{DailyLimit: 2})
	b.SetDailySlotStore(slots)

	require.NoError(t, a.SendAlert(ctx, req))
	require.NoError(t, b.SendAlert(ctx, req))
	err = a.SendAlert(ctx, req)
	assert.ErrorIs(t, err, mail.ErrDailyLimitExceeded)
	err = b.SendAlert(ctx, req)
	assert.ErrorIs(t, err, mail.ErrDailyLimitExceeded)

	c := mail.NewService(mail.Config{DailyLimit: 2})
	c.SetDailySlotStore(mail.NewRedisDailySlotStore(rdb))
	err = c.SendAlert(ctx, req)
	assert.ErrorIs(t, err, mail.ErrDailyLimitExceeded)
}

func TestMemoryDailySlotStore_IndependentServices(t *testing.T) {
	ctx := context.Background()
	req := mail.SendRequest{To: "local@test.local", Subject: "x"}

	a := mail.NewService(mail.Config{DailyLimit: 1})
	b := mail.NewService(mail.Config{DailyLimit: 1})
	require.NoError(t, a.SendAlert(ctx, req))
	require.NoError(t, b.SendAlert(ctx, req))
}

func TestService_SendAlert_DailyLimit_W7_A1_003_Regression(t *testing.T) {
	svc := mail.NewService(mail.Config{DailyLimit: 2})
	ctx := context.Background()
	req := mail.SendRequest{To: "user@test.local", Subject: "Test"}
	require.NoError(t, svc.SendAlert(ctx, req))
	require.NoError(t, svc.SendAlert(ctx, req))
	err := svc.SendAlert(ctx, req)
	assert.ErrorIs(t, err, mail.ErrDailyLimitExceeded)
}
