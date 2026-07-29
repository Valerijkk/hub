//go:build hubtest

// Package notification — тестовые export-хуки для backend/testframework/ (не компилируются в release-сборках).
package notification

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/mymmrac/telego"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/projectmember"
)

func ExportHandleStatusChanged(c *Consumer, ctx context.Context, data []byte, subject string) {
	_ = c.handleStatusChanged(ctx, data, subject)
}

func ExportHandleMessageReceived(c *Consumer, ctx context.Context, data []byte, subject string) {
	_ = c.handleMessageReceived(ctx, data, subject)
}

func ExportHandleCallStarted(c *Consumer, ctx context.Context, data []byte, subject string) {
	_ = c.handleCallStarted(ctx, data, subject)
}

func ExportHandleAssigned(c *Consumer, ctx context.Context, data []byte, subject string) {
	_ = c.handleAssigned(ctx, data, subject)
}

func ExportHandleBlockerAdded(c *Consumer, ctx context.Context, data []byte, subject string) {
	_ = c.handleBlockerAdded(ctx, data, subject)
}

func ExportHandleCommentAdded(c *Consumer, ctx context.Context, data []byte, subject string) {
	_ = c.handleCommentAdded(ctx, data, subject)
}

func ExportHandleRunCompleted(c *Consumer, ctx context.Context, data []byte, subject string) {
	_ = c.handleRunCompleted(ctx, data, subject)
}

func ExportHandleNewLogin(c *Consumer, ctx context.Context, data []byte) {
	_ = c.handleNewLogin(ctx, data)
}

func ExportProcessEvent(c *Consumer, ctx context.Context, msg jetstream.Msg) {
	c.processEvent(ctx, msg)
}

func ExportShouldDispatch(c *Consumer, userID, notifType, entityID string) bool {
	return c.shouldDispatch(userID, notifType, entityID)
}

func ExportSeedDebounceEntry(c *Consumer, key string, at time.Time) {
	c.debounce[key] = at
}

func ExportDebounceLen(c *Consumer) int {
	return len(c.debounce)
}

func ExportCleanDebounce(c *Consumer) {
	c.cleanDebounce()
}

func ExportRepositoryDB(r *Repository) *ent.Client {
	return r.db
}

func ExportBuildPushURL(entityType, entityID string) string {
	return buildPushURL(entityType, entityID)
}

func ExportIsRunFailedRecipient(role projectmember.ProjectRole, userID, completerID uuid.UUID) bool {
	return isRunFailedRecipient(role, userID, completerID)
}

func ExportTelegramBindRepository(s *TelegramBindService) *Repository {
	return s.repo
}

func ExportTelegramHandleMessage(s *TelegramBindService, ctx context.Context, msg *telego.Message) {
	s.handleMessage(ctx, msg)
}

func ExportTelegramHandleUnlinkCommand(s *TelegramBindService, ctx context.Context, chatID int64, tgID string) {
	s.handleUnlinkCommand(ctx, chatID, tgID)
}

func ExportTelegramHandleStartBind(s *TelegramBindService, ctx context.Context, chatID int64, tgID, token string) {
	s.handleStartBind(ctx, chatID, tgID, token)
}

func ExportTelegramHandleLinkCommand(s *TelegramBindService, ctx context.Context, chatID int64, tgID string) {
	s.handleLinkCommand(ctx, chatID, tgID)
}

func ExportTelegramHandleTestCommand(s *TelegramBindService, ctx context.Context, chatID int64, tgID string) {
	s.handleTestCommand(ctx, chatID, tgID)
}

// ExportTelegramMsg отражает поставленный в очередь telegramMessage для внешних тестов.
type ExportTelegramMsg struct {
	ChatID int64
	Title  string
	Body   string
}

func ExportNewTestTelegramSender(cap int) *TelegramSender {
	return &TelegramSender{queue: make(chan telegramMessage, cap)}
}

func ExportTelegramSenderQueueReady(s *TelegramSender) bool {
	return s != nil && s.queue != nil
}

func ExportTelegramTryRecv(s *TelegramSender) (ExportTelegramMsg, bool) {
	if s == nil || s.queue == nil {
		return ExportTelegramMsg{}, false
	}
	select {
	case msg := <-s.queue:
		return ExportTelegramMsg{ChatID: msg.chatID, Title: msg.title, Body: msg.body}, true
	default:
		return ExportTelegramMsg{}, false
	}
}

func ExportSendTelegram(svc *NotificationService, ctx context.Context, userID uuid.UUID, title, body string) {
	svc.sendTelegram(ctx, userID, title, body)
}

func ExportTelegramSendMessage(s *TelegramSender, ctx context.Context, chatID int64, title, body string) {
	if s == nil {
		return
	}
	s.sendMessage(ctx, telegramMessage{chatID: chatID, title: title, body: body})
}

func ExportFormatTelegramMessage(title, body string) string {
	return formatTelegramMessage(title, body)
}

func ExportSubjectToType(subject string) string {
	return SubjectToType(subject)
}

func ExportTelegramBindSender(s *TelegramBindService) *TelegramSender {
	if s == nil {
		return nil
	}
	return s.sender
}

func ExportPushSendToUser(p *PushSender, ctx context.Context, userID uuid.UUID, title, body, url string) {
	if p == nil {
		return
	}
	p.sendToUser(ctx, pushMessage{userID: userID, title: title, body: body, url: url})
}

// ExportPushQueueLen — глубина очереди push (cap 200); для assert drop при overflow (A3W7-1-002).
func ExportPushQueueLen(p *PushSender) int {
	if p == nil || p.queue == nil {
		return 0
	}
	return len(p.queue)
}

func ExportPushSendToSubscription(p *PushSender, ctx context.Context, sub PushSub, payload []byte) {
	if p == nil {
		return
	}
	p.sendToSubscription(ctx, sub, payload)
}

func ExportPushSubFromEnt(sub *ent.PushSubscription) PushSub {
	return pushSubFromEnt(sub)
}

func ExportIsQuietHours(svc *NotificationService, ctx context.Context, userID uuid.UUID) bool {
	return svc.isQuietHours(ctx, userID)
}

func ExportSetServicePush(svc *NotificationService, push *PushSender) {
	svc.push = push
}

func ExportGetUserID(c fiber.Ctx) (uuid.UUID, error) {
	return getUserID(c)
}

func ExportParseTimeMinutes(s string) (int, error) {
	return parseTimeMinutes(s)
}

func ExportMatchSuffix(parts []string, suffix string) bool {
	return matchSuffix(parts, suffix)
}

func ExportNotifToDTO(n *ent.Notification) NotificationDTO {
	return notifToDTO(n)
}

func ExportNewTestConsumerWithDebounce() *Consumer {
	return &Consumer{debounce: make(map[string]time.Time)}
}

func ExportPublishToWS(svc *NotificationService, userID uuid.UUID, notif *ent.Notification) {
	svc.publishToWS(userID, notif)
}

func ExportTGBindPrefix() string {
	return tgBindPrefix
}

func ExportTGBindTTL() time.Duration {
	return tgBindTTL
}

func ExportTGLinkCodePrefix() string {
	return tgLinkCodePrefix
}

func ExportTGLinkCodeFailsPrefix() string {
	return tgLinkCodeFailsPrefix
}

func ExportTGLinkCodeVerifyMax() int {
	return tgLinkCodeVerifyMax
}

func ExportTGLinkCodeTTL() time.Duration {
	return tgLinkCodeTTL
}

func ExportConsumerStart(c *Consumer, ctx context.Context) error {
	return c.Start(ctx)
}
