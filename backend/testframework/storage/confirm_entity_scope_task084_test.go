package storage_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/filemetadata"
	"github.com/valerii/hub/backend/internal/storage"
)

// TASK-084 / W7-A4-002: ConfirmUpload scope для message/avatar/voice/channel.
// Mutation: в validateEntityScope для этих типов вернуть только entityID!="" → denial RED.

func createPendingFile(t *testing.T, client *ent.Client, ctx context.Context, userID, projectID uuid.UUID) uuid.UUID {
	t.Helper()
	fileID := uuid.New()
	_, err := client.FileMetadata.Create().
		SetID(fileID).
		SetFilename("a.bin").
		SetContentType("application/octet-stream").
		SetSize(10).
		SetS3Key(projectID.String() + "/files/" + fileID.String() + "/a.bin").
		SetS3Bucket("hub-files").
		SetStatus(filemetadata.StatusPending).
		SetUploaderID(userID).
		SetProjectID(projectID).
		Save(ctx)
	require.NoError(t, err)
	return fileID
}

func createChannelWithMember(t *testing.T, client *ent.Client, ctx context.Context, creatorID, projectID, memberID uuid.UUID) uuid.UUID {
	t.Helper()
	ch, err := client.Channel.Create().
		SetName("scope-" + uuid.New().String()[:8]).
		SetChannelType("public").
		SetProjectID(projectID).
		SetCreatedBy(creatorID).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.ChannelMember.Create().
		SetChannelID(ch.ID).
		SetUserID(memberID).
		SetRole("member").
		Save(ctx)
	require.NoError(t, err)
	return ch.ID
}

func TestValidateEntityScope_ChannelMemberOK_TASK084(t *testing.T) {
	client, ctx, userID, projectID := newStorageTestClient(t)
	svc := storage.ExportNewStorageServiceForTest(storageRepo(client), "hub-files")
	chID := createChannelWithMember(t, client, ctx, userID, projectID, userID)

	err := storage.ExportValidateEntityScope(svc, ctx, projectID, userID, "channel", chID.String())
	require.NoError(t, err)
	err = storage.ExportValidateEntityScope(svc, ctx, projectID, userID, "message", chID.String())
	require.NoError(t, err)
	err = storage.ExportValidateEntityScope(svc, ctx, projectID, userID, "voice", chID.String())
	require.NoError(t, err)
}

func TestValidateEntityScope_ChannelNonMemberDenied_TASK084(t *testing.T) {
	client, ctx, userID, projectID := newStorageTestClient(t)
	svc := storage.ExportNewStorageServiceForTest(storageRepo(client), "hub-files")

	stranger, err := client.User.Create().
		SetEmail("stranger-" + uuid.New().String()[:8] + "@test.com").
		SetName("Stranger").
		Save(ctx)
	require.NoError(t, err)

	// Канал принадлежит stranger; uploader (userID) не член.
	chID := createChannelWithMember(t, client, ctx, stranger.ID, projectID, stranger.ID)

	for _, typ := range []string{"channel", "message", "voice"} {
		err := storage.ExportValidateEntityScope(svc, ctx, projectID, userID, typ, chID.String())
		require.ErrorIs(t, err, storage.ErrInvalidEntity, "type=%s foreign channel must deny", typ)
	}
}

func TestValidateEntityScope_SoftDeletedChannelDenied_TASK084(t *testing.T) {
	client, ctx, userID, projectID := newStorageTestClient(t)
	svc := storage.ExportNewStorageServiceForTest(storageRepo(client), "hub-files")
	chID := createChannelWithMember(t, client, ctx, userID, projectID, userID)

	_, err := client.Channel.UpdateOneID(chID).
		SetDeletedAt(time.Now().UTC()).
		Save(ctx)
	require.NoError(t, err)

	err = storage.ExportValidateEntityScope(svc, ctx, projectID, userID, "message", chID.String())
	require.ErrorIs(t, err, storage.ErrInvalidEntity)
}

func TestValidateEntityScope_ChannelProjectMismatch_TASK084(t *testing.T) {
	client, ctx, userID, projectID := newStorageTestClient(t)
	svc := storage.ExportNewStorageServiceForTest(storageRepo(client), "hub-files")

	otherProj, err := client.Project.Create().
		SetName("Other").
		SetKey("OTH2").
		Save(ctx)
	require.NoError(t, err)

	chID := createChannelWithMember(t, client, ctx, userID, otherProj.ID, userID)

	err = storage.ExportValidateEntityScope(svc, ctx, projectID, userID, "channel", chID.String())
	require.ErrorIs(t, err, storage.ErrEntityProjectMismatch)
}

func TestValidateEntityScope_AvatarSelfOK_TASK084(t *testing.T) {
	client, ctx, userID, projectID := newStorageTestClient(t)
	svc := storage.ExportNewStorageServiceForTest(storageRepo(client), "hub-files")

	require.NoError(t, storage.ExportValidateEntityScope(svc, ctx, projectID, userID, "avatar", "me"))
	require.NoError(t, storage.ExportValidateEntityScope(svc, ctx, projectID, userID, "avatar", userID.String()))
}

func TestValidateEntityScope_AvatarForeignDenied_TASK084(t *testing.T) {
	client, ctx, userID, projectID := newStorageTestClient(t)
	svc := storage.ExportNewStorageServiceForTest(storageRepo(client), "hub-files")

	other := uuid.New()
	err := storage.ExportValidateEntityScope(svc, ctx, projectID, userID, "avatar", other.String())
	require.ErrorIs(t, err, storage.ErrInvalidEntity)

	err = storage.ExportValidateEntityScope(svc, ctx, projectID, userID, "avatar", "")
	require.ErrorIs(t, err, storage.ErrInvalidEntity)
}

func TestValidateEntityScope_VoiceTempOK_TASK084(t *testing.T) {
	client, ctx, userID, projectID := newStorageTestClient(t)
	svc := storage.ExportNewStorageServiceForTest(storageRepo(client), "hub-files")
	require.NoError(t, storage.ExportValidateEntityScope(svc, ctx, projectID, userID, "voice", "temp"))
}

func TestConfirmUpload_RejectsForeignChannel_TASK084(t *testing.T) {
	client, ctx, userID, projectID := newStorageTestClient(t)
	svc := storage.ExportNewStorageServiceForTest(storageRepo(client), "hub-files")

	stranger, err := client.User.Create().
		SetEmail("own-" + uuid.New().String()[:8] + "@test.com").
		SetName("Owner").
		Save(ctx)
	require.NoError(t, err)
	chID := createChannelWithMember(t, client, ctx, stranger.ID, projectID, stranger.ID)
	fileID := createPendingFile(t, client, ctx, userID, projectID)

	err = svc.ConfirmUpload(ctx, fileID, "channel", chID.String())
	require.ErrorIs(t, err, storage.ErrInvalidEntity)

	meta, err := client.FileMetadata.Get(ctx, fileID)
	require.NoError(t, err)
	require.Equal(t, filemetadata.StatusPending, meta.Status)
}

func TestConfirmUpload_RejectsForeignAvatar_TASK084(t *testing.T) {
	client, ctx, userID, projectID := newStorageTestClient(t)
	svc := storage.ExportNewStorageServiceForTest(storageRepo(client), "hub-files")
	fileID := createPendingFile(t, client, ctx, userID, projectID)

	err := svc.ConfirmUpload(ctx, fileID, "avatar", uuid.New().String())
	require.ErrorIs(t, err, storage.ErrInvalidEntity)
}

func TestConfirmUpload_AvatarMeOK_TASK084(t *testing.T) {
	client, ctx, userID, projectID := newStorageTestClient(t)
	svc := storage.ExportNewStorageServiceForTest(storageRepo(client), "hub-files")
	fileID := createPendingFile(t, client, ctx, userID, projectID)

	require.NoError(t, svc.ConfirmUpload(ctx, fileID, "avatar", "me"))
}
