package auth

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
)

func newProfileTestEnv(t *testing.T) (*ProfileService, context.Context, uuid.UUID) {
	t.Helper()
	dsn := "file:profiletest_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()
	u, err := client.User.Create().SetEmail("profile-" + uuid.New().String()[:6] + "@example.com").Save(ctx)
	require.NoError(t, err)
	return NewProfileService(client), ctx, u.ID
}

func TestProfileService_GetMe(t *testing.T) {
	svc, ctx, uid := newProfileTestEnv(t)
	u, err := svc.GetMe(ctx, uid)
	require.NoError(t, err)
	assert.Equal(t, uid, u.ID)
}

func TestProfileService_UpdateProfile(t *testing.T) {
	svc, ctx, uid := newProfileTestEnv(t)
	name := "Updated Name"
	u, err := svc.UpdateProfile(ctx, uid, UpdateProfileRequest{Name: &name})
	require.NoError(t, err)
	assert.Equal(t, name, u.Name)
}

func TestProfileService_CompleteOnboarding(t *testing.T) {
	svc, ctx, uid := newProfileTestEnv(t)
	display := "Display"
	u, err := svc.CompleteOnboarding(ctx, uid, OnboardingCompleteRequest{
		ProfileData: UpdateProfileRequest{DisplayName: &display},
	})
	require.NoError(t, err)
	assert.True(t, u.OnboardingCompleted)
	assert.Equal(t, display, u.DisplayName)
}

func TestProfileService_CompleteOnboarding_InvalidUser(t *testing.T) {
	svc, ctx, _ := newProfileTestEnv(t)
	name := "Ghost"
	_, err := svc.CompleteOnboarding(ctx, uuid.New(), OnboardingCompleteRequest{
		ProfileData: UpdateProfileRequest{Name: &name},
	})
	require.Error(t, err)
}

func TestUserToResponse(t *testing.T) {
	svc, ctx, uid := newProfileTestEnv(t)
	u, err := svc.GetMe(ctx, uid)
	require.NoError(t, err)

	resp := UserToResponse(u)
	require.NotNil(t, resp)
	assert.Equal(t, uid.String(), resp.ID)
}

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

	u, err := svc.UpdateProfile(ctx, uid, UpdateProfileRequest{
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

	resp := UserToResponse(u)
	require.NotNil(t, resp)
	assert.Equal(t, tgID, resp.TelegramID)
}
