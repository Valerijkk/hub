package storage_test

// Покрытие ветки «объект больше лимита» в assertObjectWithinLimit:
// HeadObject отдаёт Content-Length > maxUploadBytes → объект удаляется
// из S3 и из БД, наружу уходит ErrFileTooLarge.
//
// Мутация: убрать проверку `*out.ContentLength > maxUploadBytes` в
// internal/storage/service.go:assertObjectWithinLimit → confirm проходит,
// файл остаётся в БД → оба теста RED.

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/storage"
)

// oversizedS3 поднимает фейковый S3: HEAD отдаёт заданный размер,
// DELETE считается отдельно — так видно, что оверсайз реально подчищается.
func oversizedS3(t *testing.T, size int64, deletes *atomic.Int32) *s3.Client {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodHead:
			w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
			w.WriteHeader(http.StatusOK)
		case http.MethodDelete:
			deletes.Add(1)
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	t.Cleanup(srv.Close)

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion("us-east-1"),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider("test", "test", ""),
		),
	)
	require.NoError(t, err)

	return s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(srv.URL)
		o.UsePathStyle = true
	})
}

// TestConfirmUpload_OversizedObject_Rejected — объект превышает NF-09 лимит:
// confirm отбивается ErrFileTooLarge, запись из БД удаляется.
func TestConfirmUpload_OversizedObject_Rejected(t *testing.T) {
	client, ctx, userID, projectID := newStorageTestClient(t)

	var deletes atomic.Int32
	limit := int64(storage.ExportMaxUploadBytes())
	s3Client := oversizedS3(t, limit+1, &deletes)

	svc := storage.ExportNewStorageServiceForTestWithS3(
		storage.NewEntRepository(client),
		"hub-files",
		s3Client,
		s3.NewPresignClient(s3Client),
	)

	fileID := uuid.New()
	_, err := client.FileMetadata.Create().
		SetID(fileID).
		SetFilename("huge.bin").
		SetContentType("application/octet-stream").
		SetSize(0).
		SetS3Key(projectID.String() + "/files/" + fileID.String() + "/huge.bin").
		SetS3Bucket("hub-files").
		SetStatus("pending").
		SetUploaderID(userID).
		SetProjectID(projectID).
		Save(ctx)
	require.NoError(t, err)

	err = svc.ConfirmUpload(ctx, fileID, "message", "MSG-HUGE")
	require.ErrorIs(t, err, storage.ErrFileTooLarge,
		"объект больше лимита должен отбиваться ErrFileTooLarge")

	_, getErr := client.FileMetadata.Get(ctx, fileID)
	require.Error(t, getErr,
		"запись оверсайз-файла должна удаляться из БД, иначе останется мусор")

	require.Positive(t, deletes.Load(),
		"оверсайз-объект должен удаляться из S3, иначе место не освобождается")
}

// TestConfirmUpload_ObjectAtLimit_Accepted — ровно на границе лимита проходит.
// Отделяет «больше лимита» от «равно лимиту»: off-by-one в сравнении
// сделал бы граничный файл невалидным.
func TestConfirmUpload_ObjectAtLimit_Accepted(t *testing.T) {
	client, ctx, userID, projectID := newStorageTestClient(t)

	var deletes atomic.Int32
	limit := int64(storage.ExportMaxUploadBytes())
	s3Client := oversizedS3(t, limit, &deletes)

	svc := storage.ExportNewStorageServiceForTestWithS3(
		storage.NewEntRepository(client),
		"hub-files",
		s3Client,
		s3.NewPresignClient(s3Client),
	)

	fileID := uuid.New()
	_, err := client.FileMetadata.Create().
		SetID(fileID).
		SetFilename("exact.bin").
		SetContentType("application/octet-stream").
		SetSize(0).
		SetS3Key(projectID.String() + "/files/" + fileID.String() + "/exact.bin").
		SetS3Bucket("hub-files").
		SetStatus("pending").
		SetUploaderID(userID).
		SetProjectID(projectID).
		Save(ctx)
	require.NoError(t, err)

	err = svc.ConfirmUpload(ctx, fileID, "message", "MSG-EXACT")
	require.NotErrorIs(t, err, storage.ErrFileTooLarge,
		"файл ровно на лимите отбиваться не должен")

	require.Zero(t, deletes.Load(),
		"файл в пределах лимита удалять не нужно")
}
