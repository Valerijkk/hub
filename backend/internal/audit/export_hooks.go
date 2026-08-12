//go:build hubtest

package audit

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/valerii/hub/backend/ent"
)

func ExportProcessEvent(c *AuditConsumer, ctx context.Context, msg jetstream.Msg) {
	c.processEvent(ctx, msg)
}

func ExportNewAuditConsumerForTest(repo Repository) *AuditConsumer {
	return &AuditConsumer{repo: repo}
}

func ExportDaysRemaining(deletedAt time.Time) int {
	return daysRemaining(deletedAt)
}

func ExportConsumerMapSubjectToAction(c *AuditConsumer, subject string) (string, string) {
	return c.mapSubjectToAction(subject)
}

func ExportExtractStringField(payload map[string]interface{}, keys ...string) string {
	return extractStringField(payload, keys...)
}

func ExportMapCRUDSuffix(suffix string) string {
	return mapCRUDSuffix(suffix)
}

func ExportExtractUUIDField(payload map[string]interface{}, keys ...string) *uuid.UUID {
	return extractUUIDField(payload, keys...)
}

func ExportSlowRequestThresholdMS() int {
	return slowRequestThresholdMS
}

func ExportRetentionServicePurgeOldAllureResults(s *RetentionService, ctx context.Context, days int) error {
	return s.purgeOldAllureResults(ctx, days)
}

func ExportRetentionServicePurgeOldMessages(s *RetentionService, ctx context.Context, days int) error {
	return s.purgeOldMessages(ctx, days)
}

func ExportRetentionServiceGetRetentionDays(s *RetentionService, ctx context.Context, key string, defaultVal int) int {
	return s.getRetentionDays(ctx, key, defaultVal)
}

func ExportRetentionServicePurgeOldCallRecordings(s *RetentionService, ctx context.Context, days int) error {
	return s.purgeOldCallRecordings(ctx, days)
}

func ExportDefaultChatRetentionDays() int      { return defaultChatRetentionDays }
func ExportDefaultAllureRetentionDays() int    { return defaultAllureRetentionDays }
func ExportDefaultCallRecordingRetention() int { return defaultCallRecordingRetention }

func ExportShouldSkipHTTPAudit(method, path string) bool {
	return shouldSkipHTTPAudit(method, path)
}

func ExportStripQuery(p string) string {
	return stripQuery(p)
}

func ExportRepositoryEntClient(repo Repository) *ent.Client {
	return repo.(*entRepository).client
}

// ExportConstantTimeEqualDigest — hubtest: fixed-size digest compare (W6-A4-010).
func ExportConstantTimeEqualDigest(a, b string) bool {
	return constantTimeEqualDigest(a, b)
}

// ExportConstantTimeEqualDigestOperandLen — hubtest: длина операндов CTC (должна быть sha256.Size).
func ExportConstantTimeEqualDigestOperandLen(a, b string) int {
	_, n := constantTimeDigestCompare(a, b)
	return n
}

// ExportMongoEraseRetryProcessMsg — hubtest: C-W19-002 consumer path (Ack/Nak/Term).
func ExportMongoEraseRetryProcessMsg(c *MongoEraseRetryConsumer, ctx context.Context, msg jetstream.Msg) {
	c.processMsg(ctx, msg)
}
