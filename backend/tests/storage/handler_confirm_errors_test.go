package storage_test

// Покрытие ветвления ошибок в Handler.ConfirmUpload: сервис возвращает
// разные ошибки, хендлер обязан отдавать разные HTTP-коды.
//
// Без этого теста ветки маппинга не покрыты, и подмена кода — например
// 409 вместо 404 — прошла бы незамеченной: клиент различает эти случаи
// по коду, а не по тексту.
//
// Мутация: заменить любой из кодов в internal/storage/handler.go
// (StatusConflict / StatusNotFound / StatusBadRequest) на другой → RED.

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent/filemetadata"
)

// TestHandler_ConfirmUpload_ErrorCodes — каждый класс ошибки сервиса
// отображается в свой HTTP-код.
func TestHandler_ConfirmUpload_ErrorCodes(t *testing.T) {
	app, projectID, userID, client := newStorageRoutesEnv(t)
	ctx := context.Background()

	ch, err := client.Channel.Create().
		SetName("errcodes").
		SetChannelType("public").
		SetProjectID(projectID).
		SetCreatedBy(userID).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.ChannelMember.Create().
		SetChannelID(ch.ID).
		SetUserID(userID).
		SetRole("member").
		Save(ctx)
	require.NoError(t, err)

	// newFile создаёт файл в заданном статусе и возвращает его id.
	newFile := func(status filemetadata.Status, contentType string) uuid.UUID {
		id := uuid.New()
		_, e := client.FileMetadata.Create().
			SetID(id).
			SetFilename("f.txt").
			SetContentType(contentType).
			SetSize(0).
			SetS3Key(projectID.String() + "/files/" + id.String() + "/f.txt").
			SetS3Bucket("hub-files").
			SetStatus(status).
			SetUploaderID(userID).
			SetProjectID(projectID).
			Save(ctx)
		require.NoError(t, e)
		return id
	}

	confirm := func(fileID uuid.UUID, body string) int {
		req := httptest.NewRequest(http.MethodPost,
			"/api/files/"+fileID.String()+"/confirm",
			bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")
		resp, e := app.Test(req)
		require.NoError(t, e)
		defer resp.Body.Close()
		return resp.StatusCode
	}

	okBody := fmt.Sprintf(`{"entityType":"message","entityId":%q}`, ch.ID.String())

	t.Run("уже подтверждён -> 409", func(t *testing.T) {
		id := newFile(filemetadata.StatusConfirmed, "text/plain")
		assert.Equal(t, http.StatusConflict, confirm(id, okBody),
			"повторное подтверждение — конфликт, а не ошибка запроса")
	})

	t.Run("несуществующий файл -> 404", func(t *testing.T) {
		assert.Equal(t, http.StatusNotFound, confirm(uuid.New(), okBody),
			"файла нет — 404, иначе клиент не отличит его от ошибки данных")
	})

	t.Run("битое тело запроса -> 400", func(t *testing.T) {
		id := newFile(filemetadata.StatusPending, "text/plain")
		assert.Equal(t, http.StatusBadRequest, confirm(id, `{"entityType":`),
			"невалидный JSON — 400")
	})

	t.Run("неизвестный тип сущности -> 400", func(t *testing.T) {
		id := newFile(filemetadata.StatusPending, "text/plain")
		code := confirm(id, `{"entityType":"unknown","entityId":"X"}`)
		assert.Equal(t, http.StatusBadRequest, code,
			"неизвестный entityType — ошибка данных, а не 500")
	})

	t.Run("чужой канал -> 400", func(t *testing.T) {
		id := newFile(filemetadata.StatusPending, "text/plain")
		body := fmt.Sprintf(`{"entityType":"message","entityId":%q}`, uuid.New().String())
		code := confirm(id, body)
		assert.Equal(t, http.StatusBadRequest, code,
			"канал вне проекта файла — отказ, а не подтверждение")
	})
}

// TestHandler_ConfirmUpload_BadFileID — нераспарсенный UUID в пути
// не должен доходить до сервиса.
func TestHandler_ConfirmUpload_BadFileID(t *testing.T) {
	app, _, _, _ := newStorageRoutesEnv(t)

	req := httptest.NewRequest(http.MethodPost,
		"/api/files/not-a-uuid/confirm",
		bytes.NewReader([]byte(`{"entityType":"message","entityId":"x"}`)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.GreaterOrEqual(t, resp.StatusCode, 400,
		"мусор вместо UUID обязан отбиваться на разборе пути")
	assert.Less(t, resp.StatusCode, 500,
		"это ошибка клиента, а не сбой сервера")
}
