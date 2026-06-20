package auth

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestMagicLinkService поднимает miniredis + redis client + MagicLinkService
// без Resend (dev-mode: код пишется в логи, но возвращается из SendMagicLink).
// Регистрирует cleanup'ы через t.Cleanup.
func newTestMagicLinkService(t *testing.T) (*MagicLinkService, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	svc := NewMagicLinkService(rdb, nil, "test@example.com")
	require.True(t, svc.IsDevMode(), "service must be in dev mode (nil resend)")
	return svc, mr
}

// TestMagicLink_VerifyCode_Success — happy path:
//  1. SendMagicLink возвращает код.
//  2. VerifyCode(тот же код) → nil.
//  3. Повторный VerifyCode(тот же код) → ErrMagicLinkInvalid (одноразовость).
//
// Это базовый контракт: код использовать ровно один раз.
func TestMagicLink_VerifyCode_Success(t *testing.T) {
	svc, _ := newTestMagicLinkService(t)
	ctx := context.Background()
	email := "alice@example.com"

	code, err := svc.SendMagicLink(ctx, email)
	require.NoError(t, err)
	require.NotEmpty(t, code)
	assert.Len(t, code, 6, "magic-link code must be 6 digits")

	// Первый verify — успех.
	err = svc.VerifyCode(ctx, email, code)
	require.NoError(t, err)

	// Повторный verify — код уже удалён, должен вернуться ErrMagicLinkInvalid.
	err = svc.VerifyCode(ctx, email, code)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrMagicLinkInvalid),
		"повторный verify должен вернуть ErrMagicLinkInvalid, got %v", err)
}

// TestMagicLink_VerifyCode_BruteForceLockout — HI-04:
// после magicLinkVerifyMax (=5) неудачных попыток для одного email
// 6-й и последующие VerifyCode возвращают ErrMagicLinkVerifyLimited.
//
// Дополнительно: проверяем, что счётчик `magiclink:fails:<email>` в Redis
// действительно равен magicLinkVerifyMax.
func TestMagicLink_VerifyCode_BruteForceLockout(t *testing.T) {
	svc, mr := newTestMagicLinkService(t)
	ctx := context.Background()
	email := "bruce@example.com"

	// Сначала отправим валидный код, чтобы у атакующего был осмысленный таргет.
	_, err := svc.SendMagicLink(ctx, email)
	require.NoError(t, err)

	// magicLinkVerifyMax = 5 неправильных попыток.
	for i := 0; i < magicLinkVerifyMax; i++ {
		err := svc.VerifyCode(ctx, email, "000000")
		require.Error(t, err, "попытка %d должна провалиться", i+1)
		assert.True(t, errors.Is(err, ErrMagicLinkInvalid),
			"попытка %d: ожидался ErrMagicLinkInvalid, got %v", i+1, err)
	}

	// 6-я попытка — должна быть заблокирована, без попытки сравнить код.
	err = svc.VerifyCode(ctx, email, "000000")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrMagicLinkVerifyLimited),
		"6-я попытка должна вернуть ErrMagicLinkVerifyLimited, got %v", err)

	// 7-я — тоже заблокирована.
	err = svc.VerifyCode(ctx, email, "000000")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrMagicLinkVerifyLimited))

	// Даже с **правильным** кодом теперь нельзя залогиниться — это и есть
	// ключевое свойство lockout'а (защита от race brute-force).
	correctCode, err := mr.Get("magiclink:" + email)
	require.NoError(t, err)
	err = svc.VerifyCode(ctx, email, correctCode)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrMagicLinkVerifyLimited),
		"правильный код во время lockout должен возвращать ErrMagicLinkVerifyLimited")

	// Проверяем сам счётчик.
	failsRaw, err := mr.Get(magicLinkFailsPrefix + email)
	require.NoError(t, err)
	failsInt, err := strconv.Atoi(failsRaw)
	require.NoError(t, err)
	// После 5 invalid-попыток счётчик == 5; 6-я не инкрементирует, потому
	// что VerifyCode возвращается раньше любой записи.
	assert.GreaterOrEqual(t, failsInt, magicLinkVerifyMax,
		"fails-counter должен быть >= magicLinkVerifyMax после lockout'а")
}

// TestMagicLink_VerifyCode_LockoutTTL — после lockout'а fails-counter имеет
// TTL = magicLinkTTL. miniredis.FastForward сдвигает время на TTL+1, после
// чего fails-counter истекает и новый код снова можно проверить.
//
// Это страхует от вечной блокировки на корректные коды.
func TestMagicLink_VerifyCode_LockoutTTL(t *testing.T) {
	svc, mr := newTestMagicLinkService(t)
	ctx := context.Background()
	email := "carol@example.com"

	_, err := svc.SendMagicLink(ctx, email)
	require.NoError(t, err)

	// Spam fails until lockout.
	for i := 0; i < magicLinkVerifyMax; i++ {
		_ = svc.VerifyCode(ctx, email, "wrong-1")
	}
	err = svc.VerifyCode(ctx, email, "still-wrong")
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrMagicLinkVerifyLimited))

	// Перематываем время на TTL+минута (фактически на 6 минут — гарантированно
	// больше magicLinkTTL = 5 минут).
	mr.FastForward(magicLinkTTL + time.Minute)

	// Старый код истёк (TTL ровно как fails). Отправляем новый.
	newCode, err := svc.SendMagicLink(ctx, email)
	require.NoError(t, err)

	// Новый код должен пройти — lockout снят (fails-key истёк).
	err = svc.VerifyCode(ctx, email, newCode)
	require.NoError(t, err, "после истечения TTL fails-counter'а валидный код должен пройти")
}

// TestMagicLink_SendMagicLink_RateLimit — ME-05:
// magicLinkRateMax (=3) запросов на отправку в окне magicLinkRateTTL.
// 4-й вызов возвращает ErrMagicLinkRateLimited.
func TestMagicLink_SendMagicLink_RateLimit(t *testing.T) {
	svc, _ := newTestMagicLinkService(t)
	ctx := context.Background()
	email := "dave@example.com"

	// magicLinkRateMax успешных вызовов.
	for i := 0; i < magicLinkRateMax; i++ {
		_, err := svc.SendMagicLink(ctx, email)
		require.NoError(t, err, "вызов %d/%d должен пройти", i+1, magicLinkRateMax)
	}

	// (magicLinkRateMax + 1)-й — rate-limited.
	_, err := svc.SendMagicLink(ctx, email)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrMagicLinkRateLimited),
		"должен быть ErrMagicLinkRateLimited, got %v", err)

	// Дальнейшие тоже rate-limited.
	_, err = svc.SendMagicLink(ctx, email)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrMagicLinkRateLimited))
}

// TestMagicLink_VerifyCode_NoCodeStored — VerifyCode, вызванный без
// предварительного SendMagicLink (или после истечения TTL кода), возвращает
// ErrMagicLinkInvalid и инкрементирует fails-counter.
func TestMagicLink_VerifyCode_NoCodeStored(t *testing.T) {
	svc, mr := newTestMagicLinkService(t)
	ctx := context.Background()
	email := "eve@example.com"

	err := svc.VerifyCode(ctx, email, "123456")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrMagicLinkInvalid))

	// fails-counter инкрементировался (== 1).
	failsRaw, err := mr.Get(magicLinkFailsPrefix + email)
	require.NoError(t, err)
	assert.Equal(t, "1", failsRaw)
}

// TestMagicLink_VerifyCode_SuccessClearsFails — успешный verify сбрасывает
// fails-counter. Это правильный side-effect: после login не должно быть
// "хвоста" из неудачных попыток.
func TestMagicLink_VerifyCode_SuccessClearsFails(t *testing.T) {
	svc, mr := newTestMagicLinkService(t)
	ctx := context.Background()
	email := "frank@example.com"

	code, err := svc.SendMagicLink(ctx, email)
	require.NoError(t, err)

	// Накопим пару неудач.
	_ = svc.VerifyCode(ctx, email, "wrong")
	_ = svc.VerifyCode(ctx, email, "wrong")

	failsBefore, err := mr.Get(magicLinkFailsPrefix + email)
	require.NoError(t, err)
	assert.Equal(t, "2", failsBefore)

	// Успешный verify.
	require.NoError(t, svc.VerifyCode(ctx, email, code))

	// fails-counter удалён.
	_, err = mr.Get(magicLinkFailsPrefix + email)
	assert.Error(t, err, "fails-counter должен быть удалён после успешного login'а")
}

type stubResendClient struct {
	called bool
	err    error
}

func (s *stubResendClient) SendMagicLinkEmail(_ context.Context, _, _ string) error {
	s.called = true
	return s.err
}

func TestMagicLink_SendMagicLink_ResendPath(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	resend := &stubResendClient{}
	svc := NewMagicLinkService(rdb, resend, "hub@example.com")
	require.False(t, svc.IsDevMode())

	code, err := svc.SendMagicLink(context.Background(), "prod-"+uuid.New().String()[:6]+"@example.com")
	require.NoError(t, err)
	require.NotEmpty(t, code)
	assert.True(t, resend.called)
}

func TestMagicLink_SendMagicLink_ResendErrorStillReturnsCode(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	resend := &stubResendClient{err: errors.New("smtp down")}
	svc := NewMagicLinkService(rdb, resend, "hub@example.com")

	code, err := svc.SendMagicLink(context.Background(), "fail-"+uuid.New().String()[:6]+"@example.com")
	require.NoError(t, err)
	require.NotEmpty(t, code)
	assert.True(t, resend.called)
}
