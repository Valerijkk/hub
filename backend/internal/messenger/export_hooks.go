//go:build hubtest

package messenger

import (
	"context"
	"reflect"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/valerii/hub/backend/ent"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

// ExportNewMessengerServiceTest создаёт MessengerService без NATS для hubtest.
func ExportNewMessengerServiceTest(repo Repository) *MessengerService {
	return newMessengerServiceWithEventPublisher(repo, nil, nil, nil)
}

// ExportNewMessengerServiceWithNATS создаёт MessengerService с NATS для hubtest (capture events).
func ExportNewMessengerServiceWithNATS(repo Repository, nats *natsclient.Client) *MessengerService {
	return NewMessengerService(repo, nats, nil, nil)
}

// ExportNewMessengerServiceWithEventPublisher creates a MessengerService with
// an isolated EventPublisher for hubtest producer assertions.
func ExportNewMessengerServiceWithEventPublisher(repo Repository, publisher EventPublisher) *MessengerService {
	if isNilEventPublisher(publisher) {
		panic("messenger hubtest event publisher must not be nil or typed-nil")
	}
	return newMessengerServiceWithEventPublisher(repo, publisher, nil, nil)
}

func isNilEventPublisher(publisher EventPublisher) bool {
	if publisher == nil {
		return true
	}

	value := reflect.ValueOf(publisher)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

// ExportAutoChannelServiceHandleProjectCreated вызывает handleProjectCreated для hubtest.
func ExportAutoChannelServiceHandleProjectCreated(s *AutoChannelService, payload []byte) {
	s.handleProjectCreated(payload)
}

// ExportAutoChannelServiceSetLifecycleCtx задаёт lifecycleCtx для hubtest.
func ExportAutoChannelServiceSetLifecycleCtx(s *AutoChannelService, ctx context.Context) {
	s.lifecycleCtx = ctx
}

// ExportAutoChannelServiceSubscription возвращает NATS-подписку для hubtest.
func ExportAutoChannelServiceSubscription(s *AutoChannelService) *nats.Subscription {
	return s.sub
}

// ExportToMessageResponse экспортирует toMessageResponse для hubtest.
func ExportToMessageResponse(doc *MessageDocument) *MessageResponse {
	return toMessageResponse(doc)
}

// ExportToChannelResponse экспортирует toChannelResponse для hubtest.
func ExportToChannelResponse(ch *ent.Channel) *ChannelResponse {
	return toChannelResponse(ch)
}

// ExportNewMessengerServiceEmpty создаёт пустой MessengerService для hubtest.
func ExportNewMessengerServiceEmpty() *MessengerService {
	return newMessengerServiceWithEventPublisher(nil, nil, nil, nil)
}

// ExportToChannelInviteResponse конвертирует ent.ChannelInvite в ChannelInviteResponse для hubtest.
func ExportToChannelInviteResponse(inv *ent.ChannelInvite) *ChannelInviteResponse {
	return toChannelInviteResponse(inv)
}

// ExportToPollResponse собирает PollResponse для hubtest.
func ExportToPollResponse(p *ent.Poll, options []*ent.PollOption, votes []*ent.PollVote, currentUserID uuid.UUID, isAnonymous bool) *PollResponse {
	return toPollResponse(p, options, votes, currentUserID, isAnonymous)
}

// ExportSplitComma разбивает строку по запятым для hubtest.
func ExportSplitComma(s string) []string {
	return splitComma(s)
}

// ExportQueryInt читает int query-параметр для hubtest.
func ExportQueryInt(c fiber.Ctx, key string, defaultVal int) int {
	return queryInt(c, key, defaultVal)
}

// ExportMessengerServiceUnread возвращает unread-сервис из MessengerService для hubtest.
func ExportMessengerServiceUnread(s *MessengerService) *UnreadService {
	return s.unread
}

// ExportMessengerServiceResolveGroupMentions вызывает resolveGroupMentions для hubtest.
func ExportMessengerServiceResolveGroupMentions(s *MessengerService, ctx context.Context, channelID uuid.UUID, projectID *uuid.UUID, groups []string, individualMentions []string, senderID uuid.UUID) []string {
	return s.resolveGroupMentions(ctx, channelID, projectID, groups, individualMentions, senderID)
}

// ExportParseChannelID парсит channel_id из Fiber-контекста для hubtest.
func ExportParseChannelID(c fiber.Ctx) (uuid.UUID, error) {
	return parseChannelID(c)
}

// ExportParseMessageID парсит message_id из Fiber-контекста для hubtest.
func ExportParseMessageID(c fiber.Ctx) string {
	return parseMessageID(c)
}

// ExportPresenceOnlineKey возвращает Redis-ключ online presence для hubtest.
func ExportPresenceOnlineKey() string {
	return presenceOnlineKey
}

// ExportMapKeys возвращает ключи map для hubtest.
func ExportMapKeys(m map[string]struct{}) []string {
	return mapKeys(m)
}

// ExportParseExpiresIn парсит expires_in для hubtest.
func ExportParseExpiresIn(expiresIn string) (*time.Time, error) {
	return parseExpiresIn(expiresIn)
}

// ExportMessengerServiceEmitMessageSent вызывает emitMessageSent для hubtest.
func ExportMessengerServiceEmitMessageSent(s *MessengerService, ctx context.Context, evt MessageHookEvent) {
	s.emitMessageSent(ctx, evt)
}

// ExportMessengerServiceIssueThreadLinker возвращает issueThreadLinker для hubtest.
func ExportMessengerServiceIssueThreadLinker(s *MessengerService) IssueThreadLinker {
	return s.issueThreadLinker
}

// ExportMessengerServicePromoteCheck возвращает promoteCheck для hubtest.
func ExportMessengerServicePromoteCheck(s *MessengerService) PromotePermissionChecker {
	return s.promoteCheck
}

// ExportGenerateInviteToken генерирует токен приглашения для hubtest.
func ExportGenerateInviteToken() (string, error) {
	return generateInviteToken()
}

// ExportInviteTokenLength возвращает длину токена приглашения для hubtest.
func ExportInviteTokenLength() int {
	return inviteTokenLength
}

// ExportBase62Chars возвращает алфавит base62 для hubtest.
func ExportBase62Chars() string {
	return base62Chars
}

// ExportTruncateRunes — P3-BE-04 / TASK-060: rune-safe truncate для hubtest.
func ExportTruncateRunes(s string, maxRunes int) string {
	return truncateRunes(s, maxRunes)
}
