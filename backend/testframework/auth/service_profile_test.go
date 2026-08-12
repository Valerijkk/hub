package auth_test

// Тесты сервиса аутентификации: интеграционные и unit-сценарии testframework.
import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/internal/auth"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

func newProfileTestEnv(t *testing.T) (*auth.ProfileService, context.Context, uuid.UUID) {
	t.Helper()
	dsn := "file:profiletest_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()
	u, err := client.User.Create().SetEmail("profile-" + uuid.New().String()[:6] + "@example.com").Save(ctx)
	require.NoError(t, err)
	return auth.NewProfileService(auth.NewEntRepository(client)), ctx, u.ID
}

// TestProfileService_GetMe проверяет бизнес-логику: профиль сервис получение me.
func TestProfileService_GetMe(t *testing.T) {
	svc, ctx, uid := newProfileTestEnv(t)
	u, err := svc.GetMe(ctx, uid)
	require.NoError(t, err)
	assert.Equal(t, uid, u.ID)
}

// TestProfileService_UpdateProfile проверяет бизнес-логику: профиль сервис обновление профиль.
func TestProfileService_UpdateProfile(t *testing.T) {
	svc, ctx, uid := newProfileTestEnv(t)
	name := "Updated Name"
	u, err := svc.UpdateProfile(ctx, uid, auth.UpdateProfileRequest{Name: &name})
	require.NoError(t, err)
	assert.Equal(t, name, u.Name)
}

// TestProfileService_CompleteOnboarding проверяет бизнес-логику: профиль сервис complete онбординг.
func TestProfileService_CompleteOnboarding(t *testing.T) {
	svc, ctx, uid := newProfileTestEnv(t)
	display := "Display"
	u, err := svc.CompleteOnboarding(ctx, uid, auth.OnboardingCompleteRequest{
		ProfileData: auth.UpdateProfileRequest{DisplayName: &display},
	})
	require.NoError(t, err)
	assert.True(t, u.OnboardingCompleted)
	assert.Equal(t, display, u.DisplayName)
}

// TestProfileService_CompleteOnboarding_InvalidUser проверяет отклонение при профиль сервис complete онбординг некорректный user.
func TestProfileService_CompleteOnboarding_InvalidUser(t *testing.T) {
	svc, ctx, _ := newProfileTestEnv(t)
	name := "Ghost"
	_, err := svc.CompleteOnboarding(ctx, uuid.New(), auth.OnboardingCompleteRequest{
		ProfileData: auth.UpdateProfileRequest{Name: &name},
	})
	require.Error(t, err)
}

// TestUserToResponse проверяет сценарий: user to response.
func TestUserToResponse(t *testing.T) {
	svc, ctx, uid := newProfileTestEnv(t)
	u, err := svc.GetMe(ctx, uid)
	require.NoError(t, err)

	resp := auth.UserToResponse(u)
	require.NotNil(t, resp)
	assert.Equal(t, uid.String(), resp.ID)
}

// TestProfileService_UpdateProfile_AllFields проверяет бизнес-логику: профиль сервис обновление профиль all fields.
func TestProfileService_UpdateProfile_AllFields(t *testing.T) {
	svc, ctx, uid := newProfileTestEnv(t)

	name := "Full Name"
	display := "Display"
	avatar := "https://cdn.example/avatar.png"
	phone := "+1234567890"
	position := "Engineer"
	department := "Platform"
	tz := "Europe/Moscow"
	lang := "ru"
	birthday := time.Date(1990, 1, 15, 0, 0, 0, 0, time.UTC)
	quietStart := "22:00"
	quietEnd := "08:00"
	custom := map[string]any{"team": "alpha"}

	u, err := svc.UpdateProfile(ctx, uid, auth.UpdateProfileRequest{
		Name:            &name,
		DisplayName:     &display,
		AvatarURL:       &avatar,
		Phone:           &phone,
		Position:        &position,
		Department:      &department,
		Timezone:        &tz,
		Language:        &lang,
		Birthday:        &birthday,
		QuietHoursStart: &quietStart,
		QuietHoursEnd:   &quietEnd,
		CustomFields:    custom,
	})
	require.NoError(t, err)
	assert.Equal(t, name, u.Name)
	assert.Equal(t, display, u.DisplayName)
	assert.Equal(t, avatar, u.AvatarURL)
	assert.Equal(t, phone, u.Phone)
	assert.Equal(t, position, u.Position)
	assert.Equal(t, department, u.Department)
	assert.Equal(t, tz, u.Timezone)
	assert.Equal(t, lang, u.Language)
	require.NotNil(t, u.Birthday)
	assert.Equal(t, birthday, *u.Birthday)
	require.NotNil(t, u.QuietHoursStart)
	assert.Equal(t, quietStart, *u.QuietHoursStart)
	require.NotNil(t, u.QuietHoursEnd)
	assert.Equal(t, quietEnd, *u.QuietHoursEnd)
	assert.Equal(t, custom, u.CustomFields)
}

// TestProfileService_UpdateProfile_AvatarURL_SchemeWhitelist_A2W6_A1_003
// отклоняет javascript:/data: и принимает https + относительный `/…`.
func TestProfileService_UpdateProfile_AvatarURL_SchemeWhitelist_A2W6_A1_003(t *testing.T) {
	svc, ctx, uid := newProfileTestEnv(t)

	for _, bad := range []string{
		"javascript:alert(1)",
		"data:image/png;base64,abc",
		"vbscript:msgbox(1)",
		"//evil.example/x.png",
	} {
		bad := bad
		_, err := svc.UpdateProfile(ctx, uid, auth.UpdateProfileRequest{AvatarURL: &bad})
		require.Error(t, err, "avatar %q must be rejected", bad)
		var appErr *apperrors.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, 400, appErr.Code, "avatar %q", bad)
	}

	httpsURL := "https://cdn.example/avatar.png"
	u, err := svc.UpdateProfile(ctx, uid, auth.UpdateProfileRequest{AvatarURL: &httpsURL})
	require.NoError(t, err)
	assert.Equal(t, httpsURL, u.AvatarURL)

	rel := "/files/avatars/me.png"
	u, err = svc.UpdateProfile(ctx, uid, auth.UpdateProfileRequest{AvatarURL: &rel})
	require.NoError(t, err)
	assert.Equal(t, rel, u.AvatarURL)

	empty := ""
	u, err = svc.UpdateProfile(ctx, uid, auth.UpdateProfileRequest{AvatarURL: &empty})
	require.NoError(t, err)
	assert.Equal(t, "", u.AvatarURL)
}

// TestProfileService_CompleteOnboarding_JavascriptAvatar_Rejected проверяет
// тот же whitelist на пути онбординга (не обходит UpdateProfile).
func TestProfileService_CompleteOnboarding_JavascriptAvatar_Rejected(t *testing.T) {
	svc, ctx, uid := newProfileTestEnv(t)
	bad := "javascript:alert(1)"
	_, err := svc.CompleteOnboarding(ctx, uid, auth.OnboardingCompleteRequest{
		ProfileData: auth.UpdateProfileRequest{AvatarURL: &bad},
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 400, appErr.Code)
}

// TestUserToResponse_WithTelegramID проверяет сценарий: user to response with Telegram id.
func TestUserToResponse_WithTelegramID(t *testing.T) {
	dsn := "file:profiletg_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()
	tgID := "123456789"
	u, err := client.User.Create().
		SetEmail("tg-" + uuid.New().String()[:6] + "@example.com").
		SetTelegramID(tgID).
		Save(ctx)
	require.NoError(t, err)

	resp := auth.UserToResponse(u)
	require.NotNil(t, resp)
	assert.Equal(t, tgID, resp.TelegramID)
}
