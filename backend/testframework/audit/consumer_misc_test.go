package audit_test

// Тесты NATS consumer аудита: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/audit"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestAuditConsumer_Stop_NilContext проверяет сценарий: audit consumer stop nil context.
func TestAuditConsumer_Stop_NilContext(t *testing.T) {
	c := audit.ExportNewAuditConsumerForTest(nil)
	assert.NotPanics(t, func() {
		c.Stop()
	})
}

// TestRetentionService_PurgeOldAllureResults_NilMongo проверяет бизнес-логику: хранение сервис очистка old allure results nil mongo.
func TestRetentionService_PurgeOldAllureResults_NilMongo(t *testing.T) {
	svc := audit.NewRetentionService(newRetentionTestClient(t), nil, nil, "")
	err := audit.ExportRetentionServicePurgeOldAllureResults(svc, context.Background(), 365)
	assert.NoError(t, err)
}

// TestMapSubjectToAction_AuthUnknownSuffix проверяет сценарий: map subject to action auth unknown suffix.
func TestMapSubjectToAction_AuthUnknownSuffix(t *testing.T) {
	consumer := audit.ExportNewAuditConsumerForTest(nil)
	action, entity := audit.ExportConsumerMapSubjectToAction(consumer, "auth.session.expired")
	assert.Empty(t, action)
	assert.Empty(t, entity)
}

// TestExtractStringField_NonStringValueSkipped проверяет сценарий: extract string field non string value skipped.
func TestExtractStringField_NonStringValueSkipped(t *testing.T) {
	payload := map[string]interface{}{
		"id": 12345,
	}
	got := audit.ExportExtractStringField(payload, "id")
	assert.Empty(t, got)
}
