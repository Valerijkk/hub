package audit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuditConsumer_Stop_NilContext(t *testing.T) {
	c := &AuditConsumer{}
	assert.NotPanics(t, func() {
		c.Stop()
	})
}

func TestRetentionService_PurgeOldAllureResults_NilMongo(t *testing.T) {
	svc := NewRetentionService(newRetentionTestClient(t), nil, nil, "")
	err := svc.purgeOldAllureResults(context.Background(), 365)
	assert.NoError(t, err)
}

func TestMapSubjectToAction_AuthUnknownSuffix(t *testing.T) {
	consumer := &AuditConsumer{}
	action, entity := consumer.mapSubjectToAction("auth.session.expired")
	assert.Empty(t, action)
	assert.Empty(t, entity)
}

func TestExtractStringField_NonStringValueSkipped(t *testing.T) {
	payload := map[string]interface{}{
		"id": 12345,
	}
	got := extractStringField(payload, "id")
	assert.Empty(t, got)
}
