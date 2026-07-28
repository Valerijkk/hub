package ai

import (
	"context"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/channel"
	messengermod "github.com/valerii/hub/backend/internal/messenger"
)

const (
	botHistoryLimit       = 20
	botStreamEditInterval = 200 * time.Millisecond
	botStreamCursor       = "▍"
)

var (
	hubAIMentionRe    = regexp.MustCompile(`(?i)@hub-ai`)
	hubAIMentionStrip = regexp.MustCompile(`(?i)@hub-ai\s*`)
)

// MessengerBot — обработчик @hub-ai (M3-009..011).
type MessengerBot struct {
	ai     *Service
	msg    *messengermod.MessengerService
	botID  uuid.UUID
	issues IssueLookup
	docs   DocSearchProvider
	review CodeReviewProvider
}

// NewMessengerBot создаёт post-send hook для @hub-ai.
func NewMessengerBot(ai *Service, msg *messengermod.MessengerService, botUserID uuid.UUID) *MessengerBot {
	return &MessengerBot{
		ai:    ai,
		msg:   msg,
		botID: botUserID,
	}
}

// SetIssueLookup подключает tracker для /ai test-cases (M3-016).
func (b *MessengerBot) SetIssueLookup(l IssueLookup) {
	b.issues = l
}

// SetDocSearch подключает Hub search для /ai docs (M3-021).
func (b *MessengerBot) SetDocSearch(p DocSearchProvider) {
	b.docs = p
}

// SetCodeReview подключает GitHub/GitLab для /ai review (M3-017).
func (b *MessengerBot) SetCodeReview(p CodeReviewProvider) {
	b.review = p
}

// OnMessageSent реализует messenger.MessageSentHook.
func (b *MessengerBot) OnMessageSent(ctx context.Context, evt messengermod.MessageHookEvent) error {
	if evt.SenderID == b.botID {
		return nil
	}
	ch, err := b.msg.GetChannelEntity(ctx, evt.ChannelID)
	if err != nil {
		return err
	}
	memberIDs, err := b.msg.GetChannelMemberIDsInternal(ctx, evt.ChannelID)
	if err != nil {
		return err
	}
	if !b.shouldRespond(evt, ch, memberIDs) {
		return nil
	}

	model, err := b.ai.DefaultModel(ctx, evt.SenderID)
	if err != nil {
		slog.Warn("@hub-ai: no model configured", "error", err)
		return b.replyError(ctx, evt, ch, "AI провайдер не настроен. Admin → AI → добавьте Ollama или DeepSeek.")
	}

	prompt := extractUserPrompt(evt.ContentText)
	if prompt == "" {
		prompt = "Привет! Чем могу помочь?"
	}

	replyThread := evt.ThreadID
	if replyThread == "" && ch.ChannelType != channel.ChannelTypeDm {
		replyThread = evt.MessageID
	}

	if cmd := parseBotCommand(prompt); cmd != nil {
		return b.handleCommand(ctx, evt, ch, cmd, model, replyThread)
	}

	history, err := b.loadHistory(ctx, evt, ch, botHistoryLimit)
	if err != nil {
		slog.Error("@hub-ai: load history", "error", err)
		return err
	}

	chatReq := buildChatRequest(b.ai.BotSystemPrompt(ctx), history, b.botID, prompt)
	chatReq.Model = model
	return b.streamReply(ctx, evt, ch, evt.SenderID, chatReq, replyThread)
}

func (b *MessengerBot) handleCommand(
	ctx context.Context,
	evt messengermod.MessageHookEvent,
	ch *ent.Channel,
	cmd *botCommand,
	model, replyThread string,
) error {
	var systemPrompt, userPrompt string

	switch cmd.name {
	case "summarize":
		msgs, err := b.loadHistory(ctx, evt, ch, botSummarizeLimit)
		if err != nil {
			return err
		}
		texts := make([]string, 0, len(msgs))
		for _, m := range msgs {
			texts = append(texts, m.ContentText)
		}
		systemPrompt, userPrompt = buildSummarizePrompt(texts)
	case "test-cases":
		key := extractIssueKey(cmd.args)
		if key == "" {
			return b.replyError(ctx, evt, ch, "Укажите ключ задачи: `/ai test-cases HUB-123`")
		}
		if b.issues == nil {
			return b.replyError(ctx, evt, ch, "Поиск задач недоступен.")
		}
		title, desc, err := b.issues.IssueContext(ctx, evt.SenderID, key)
		if err != nil {
			return b.replyError(ctx, evt, ch, "Задача **"+key+"** не найдена.")
		}
		systemPrompt, userPrompt = buildTestCasesPrompt(key, title, desc)
	case "explain":
		systemPrompt, userPrompt = buildExplainPrompt(cmd.args)
	case "translate":
		systemPrompt, userPrompt = buildTranslatePrompt(cmd.args)
	case "review":
		prURL := extractPRURL(cmd.args)
		if prURL == "" {
			prURL = strings.TrimSpace(cmd.args)
		}
		if prURL == "" {
			return b.replyError(ctx, evt, ch, "Укажите URL PR/MR: `/ai review https://github.com/org/repo/pull/42`")
		}
		if b.review == nil {
			return b.replyError(ctx, evt, ch, "Code review недоступен (нет интеграции GitHub/GitLab).")
		}
		reviewCtx, err := b.review.FetchReview(ctx, evt.SenderID, prURL)
		if err != nil {
			return b.replyError(ctx, evt, ch, "Code review недоступен для этого репозитория.")
		}
		systemPrompt, userPrompt = buildReviewPrompt(reviewCtx)
	case "docs":
		query := strings.TrimSpace(cmd.args)
		if query == "" {
			return b.replyError(ctx, evt, ch, "Укажите вопрос: `/ai docs как настроить TMS`")
		}
		if b.docs == nil {
			return b.replyError(ctx, evt, ch, "Поиск по Hub недоступен.")
		}
		docCtx, err := b.docs.SearchContext(ctx, evt.SenderID, query)
		if err != nil {
			return b.replyError(ctx, evt, ch, "Ошибка поиска: "+err.Error())
		}
		systemPrompt, userPrompt = buildDocsPrompt(docCtx, query)
	case "sql":
		systemPrompt, userPrompt = buildSQLPrompt(cmd.args)
	case "help":
		systemPrompt, userPrompt = buildHelpPrompt()
	default:
		return b.replyError(ctx, evt, ch, "Неизвестная команда. Введите `/ai help`")
	}

	chatReq := ChatCompletionRequest{
		Model: model,
		Messages: []ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}
	return b.streamReply(ctx, evt, ch, evt.SenderID, chatReq, replyThread)
}

func (b *MessengerBot) streamReply(
	ctx context.Context,
	evt messengermod.MessageHookEvent,
	ch *ent.Channel,
	userID uuid.UUID,
	chatReq ChatCompletionRequest,
	replyThread string,
) error {
	_ = b.msg.PublishTyping(ctx, evt.ChannelID.String(), b.botID)

	placeholder, err := b.msg.SendBotMessage(ctx, evt.ChannelID, messengermod.SendMessageRequest{
		ContentText: botStreamCursor,
		ContentJSON: markdownToPlateJSON(botStreamCursor),
		ThreadID:    replyThread,
	})
	if err != nil {
		return err
	}

	var streamed strings.Builder
	lastEdit := time.Time{}
	answer, err := b.ai.StreamChatCompletion(ctx, userID, chatReq, "", "hub-ai-bot", func(chunk string) error {
		streamed.WriteString(chunk)
		if time.Since(lastEdit) < botStreamEditInterval {
			return nil
		}
		lastEdit = time.Now()
		partial := streamed.String() + botStreamCursor
		return b.msg.UpdateBotMessage(ctx, placeholder.ID, partial, markdownToPlateJSON(partial))
	})
	if err != nil {
		slog.Warn("@hub-ai: stream failed", "error", err)
		return b.replyError(ctx, evt, ch, "Не удалось получить ответ от AI. Проверьте провайдер в Admin → AI.")
	}

	if answer == "" {
		answer = "_(пустой ответ)_"
	}
	return b.msg.UpdateBotMessage(ctx, placeholder.ID, answer, markdownToPlateJSON(answer))
}

func (b *MessengerBot) shouldRespond(evt messengermod.MessageHookEvent, ch *ent.Channel, memberIDs []uuid.UUID) bool {
	// P1-BE-31: бот отвечает только если он член канала — иначе mention
	// в private/public без bot∈members читал историю во внешний LLM.
	botIsMember := false
	for _, id := range memberIDs {
		if id == b.botID {
			botIsMember = true
			break
		}
	}
	if !botIsMember {
		return false
	}
	if ch.ChannelType == channel.ChannelTypeDm {
		return true
	}
	if hubAIMentionRe.MatchString(evt.ContentText) {
		return true
	}
	botStr := b.botID.String()
	for _, m := range evt.Mentions {
		if m == botStr {
			return true
		}
	}
	return false
}

func (b *MessengerBot) loadHistory(ctx context.Context, evt messengermod.MessageHookEvent, ch *ent.Channel, limit int) ([]messengermod.MessageResponse, error) {
	threadRoot := evt.ThreadID
	if threadRoot == "" && ch.ChannelType == channel.ChannelTypeDm {
		return b.msg.ListMessagesForBot(ctx, evt.ChannelID, nil, limit)
	}
	if threadRoot == "" {
		return b.msg.ListMessagesForBot(ctx, evt.ChannelID, nil, limit)
	}
	replies, err := b.msg.ListThreadRepliesForBot(ctx, threadRoot, nil, limit)
	if err != nil {
		return nil, err
	}
	parent, err := b.msg.FindMessageForBot(ctx, threadRoot)
	if err == nil && parent != nil {
		replies = append([]messengermod.MessageResponse{*parent}, replies...)
	}
	if len(replies) > limit {
		replies = replies[len(replies)-limit:]
	}
	return replies, nil
}

func buildChatRequest(systemPrompt string, history []messengermod.MessageResponse, botID uuid.UUID, userPrompt string) ChatCompletionRequest {
	msgs := []ChatMessage{{Role: "system", Content: systemPrompt}}
	for _, h := range history {
		text := strings.TrimSpace(h.ContentText)
		if text == "" {
			continue
		}
		role := "user"
		if h.SenderID == botID.String() {
			role = "assistant"
		}
		msgs = append(msgs, ChatMessage{Role: role, Content: text})
	}
	msgs = append(msgs, ChatMessage{Role: "user", Content: userPrompt})
	return ChatCompletionRequest{
		Messages: msgs,
	}
}

func extractUserPrompt(text string) string {
	return strings.TrimSpace(hubAIMentionStrip.ReplaceAllString(text, ""))
}

func extractAssistantText(resp *ChatCompletionResponse) string {
	if resp == nil || len(resp.Choices) == 0 {
		return ""
	}
	return strings.TrimSpace(resp.Choices[0].Message.Content)
}

func (b *MessengerBot) replyError(ctx context.Context, evt messengermod.MessageHookEvent, ch *ent.Channel, msg string) error {
	replyThread := evt.ThreadID
	if replyThread == "" && ch.ChannelType != channel.ChannelTypeDm {
		replyThread = evt.MessageID
	}
	_, err := b.msg.SendBotMessage(ctx, evt.ChannelID, messengermod.SendMessageRequest{
		ContentText: msg,
		ContentJSON: markdownToPlateJSON(msg),
		ThreadID:    replyThread,
	})
	return err
}

// markdownToPlateJSON — простой Plate doc с markdown-текстом (FE рендерит через BotMarkdown).
func markdownToPlateJSON(text string) map[string]interface{} {
	return map[string]interface{}{
		"type": "doc",
		"children": []map[string]interface{}{
			{
				"type": "p",
				"children": []map[string]interface{}{
					{"text": text},
				},
			},
		},
		"hub_ai_markdown": true,
	}
}
