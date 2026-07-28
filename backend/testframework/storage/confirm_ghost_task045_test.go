package storage_test

// TASK-045 / P2-BE-21: Confirm без объекта в S3 → ErrObjectNotFound (не helper-only).
import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/filemetadata"
	"github.com/valerii/hub/backend/internal/storage"
)

// TestConfirmUpload_MissingS3Object_NotFound — pending file + HeadObject 404 → ErrObjectNotFound.
// Мутация: убрать isS3NotFound→ErrObjectNotFound в assertObjectWithinLimit → ошибка не ErrObjectNotFound / confirm проходит → RED.
func TestConfirmUpload_MissingS3Object_NotFound(t *testing.T) {
	client, ctx, userID, projectID := newStorageTestClient(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><Error><Code>NoSuchKey</Code><Message>Not Found</Message></Error>`))
	}))
	t.Cleanup(srv.Close)

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion("us-east-1"),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider("test", "test", ""),
		),
	)
	require.NoError(t, err)
	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(srv.URL)
		o.UsePathStyle = true
	})
	svc := storage.ExportNewStorageServiceForTestWithS3(
		storage.NewEntRepository(client),
		"hub-files",
		s3Client,
		s3.NewPresignClient(s3Client),
	)

	fileID := uuid.New()
	_, err = client.FileMetadata.Create().
		SetID(fileID).
		SetFilename("ghost.pdf").
		SetContentType("application/pdf").
		SetSize(0).
		SetS3Key(projectID.String() + "/files/" + fileID.String() + "/ghost.pdf").
		SetS3Bucket("hub-files").
		SetStatus(filemetadata.StatusPending).
		SetUploaderID(userID).
		SetProjectID(projectID).
		Save(ctx)
	require.NoError(t, err)

	err = svc.ConfirmUpload(ctx, fileID, "message", "MSG-GHOST")
	require.ErrorIs(t, err, storage.ErrObjectNotFound)

	got, err := client.FileMetadata.Get(ctx, fileID)
	require.NoError(t, err)
	require.Equal(t, filemetadata.StatusPending, got.Status, "ghost confirm не должен переводить в confirmed")
}
