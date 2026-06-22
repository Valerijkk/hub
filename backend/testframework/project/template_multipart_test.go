package project_test

// Тесты модуля проектов: интеграционные и unit-сценарии testframework.
import (
	"bytes"
	"mime/multipart"
	"testing"

	"github.com/stretchr/testify/require"
)

func multipartWriter(t *testing.T, body *bytes.Buffer, field, filename string, content []byte) *multipart.Writer {
	t.Helper()
	w := multipart.NewWriter(body)
	part, err := w.CreateFormFile(field, filename)
	require.NoError(t, err)
	_, err = part.Write(content)
	require.NoError(t, err)
	return w
}
