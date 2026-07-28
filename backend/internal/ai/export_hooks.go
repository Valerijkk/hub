//go:build hubtest

package ai

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	githubmod "github.com/valerii/hub/backend/internal/github"
	gitlabmod "github.com/valerii/hub/backend/internal/gitlab"
	messengermod "github.com/valerii/hub/backend/internal/messenger"
)

func ExportParseBotCommand(prompt string) *botCommand {
	return parseBotCommand(prompt)
}

func ExportBuildSummarizePrompt(messages []string) (string, string) {
	return buildSummarizePrompt(messages)
}

func ExportBuildTestCasesPrompt(issueKey, title, description string) (string, string) {
	return buildTestCasesPrompt(issueKey, title, description)
}

func ExportBuildExplainPrompt(code string) (string, string) {
	return buildExplainPrompt(code)
}

func ExportBuildTranslatePrompt(text string) (string, string) {
	return buildTranslatePrompt(text)
}

func ExportBuildDocsPrompt(contextText, query string) (string, string) {
	return buildDocsPrompt(contextText, query)
}

func ExportBuildSQLPrompt(question string) (string, string) {
	return buildSQLPrompt(question)
}

func ExportBuildHelpPrompt() (string, string) {
	return buildHelpPrompt()
}

func ExportExtractPRURL(args string) string {
	return extractPRURL(args)
}

func ExportExtractIssueKey(args string) string {
	return extractIssueKey(args)
}

func ExportGithubPRToReview(pr *githubmod.PRReviewContext) *ReviewContext {
	return githubPRToReview(pr)
}

func ExportGitlabMRToReview(mr *gitlabmod.MRReviewContext) *ReviewContext {
	return gitlabMRToReview(mr)
}

func ExportBuildReviewPrompt(rc *ReviewContext) (string, string) {
	return buildReviewPrompt(rc)
}

func ExportParseYandexCredentials(apiKey string) (folderID, key string, ok bool) {
	return parseYandexCredentials(apiKey)
}

func ExportParseGigaCredentials(apiKey string) (clientID, clientSecret string, ok bool) {
	return parseGigaCredentials(apiKey)
}

func ExportTruncate(s string, n int) string {
	return truncate(s, n)
}

func ExportUserIDFromCtx(c fiber.Ctx) (uuid.UUID, error) {
	return userIDFromCtx(c)
}

func ExportSecondsUntilMidnightUTC() string {
	return secondsUntilMidnightUTC()
}

func ExportNormalizeModels(models []string) []string {
	return normalizeModels(models)
}

func ExportFilterProvidersByConsent(providers []*ent.AIProvider, allowExternal bool) []*ent.AIProvider {
	return filterProvidersByConsent(providers, allowExternal)
}

func ExportConsentErrorForModel(model string, allowExternal bool, providers []*ent.AIProvider) error {
	return consentErrorForModel(model, allowExternal, providers)
}

func ExportValidateChatRequest(req ChatCompletionRequest) error {
	return validateChatRequest(req)
}

func ExportIsRetryableProviderError(err error) bool {
	return isRetryableProviderError(err)
}

func ExportIsOfflineProvider(p *ent.AIProvider) bool {
	return isOfflineProvider(p)
}

func ExportExtractUserPrompt(text string) string {
	return extractUserPrompt(text)
}

func ExportMarkdownToPlateJSON(text string) map[string]interface{} {
	return markdownToPlateJSON(text)
}

func ExportBuildChatRequest(systemPrompt string, history []messengermod.MessageResponse, botID uuid.UUID, userPrompt string) ChatCompletionRequest {
	return buildChatRequest(systemPrompt, history, botID, userPrompt)
}

func ExportExtractAssistantText(resp *ChatCompletionResponse) string {
	return extractAssistantText(resp)
}

func ExportSplitAnthropicMessages(msgs []ChatMessage) (string, []anthropicMessage) {
	return splitAnthropicMessages(msgs)
}

func ExportAnthropicTextFromBlocks(texts ...string) string {
	blocks := make([]struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}, len(texts))
	for i, t := range texts {
		blocks[i] = struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}{Type: "text", Text: t}
	}
	return anthropicText(&anthropicResponse{Content: blocks})
}

func ExportNormalizeComposerCompletion(prefix, raw string) string {
	return normalizeComposerCompletion(prefix, raw)
}

func ExportServiceResolveProviders(s *Service, ctx context.Context, userID uuid.UUID, model string) ([]*ent.AIProvider, error) {
	return s.resolveProviders(ctx, userID, model)
}

func ExportServiceEnforceQuota(s *Service, ctx context.Context, userID uuid.UUID) error {
	return s.enforceQuota(ctx, userID)
}

func ExportNewMessengerBotForShouldRespond(botID uuid.UUID) *MessengerBot {
	return &MessengerBot{botID: botID}
}

func ExportMessengerBotShouldRespond(b *MessengerBot, evt messengermod.MessageHookEvent, ch *ent.Channel, memberIDs []uuid.UUID) bool {
	return b.shouldRespond(evt, ch, memberIDs)
}

func ExportBotCommandName(cmd *botCommand) string {
	if cmd == nil {
		return ""
	}
	return cmd.name
}

func ExportBotCommandArgs(cmd *botCommand) string {
	if cmd == nil {
		return ""
	}
	return cmd.args
}

func ExportSettingBotSystemPrompt() string {
	return settingBotSystemPrompt
}

func ExportNewOpenAIAdapterWithClient(client *http.Client) *OpenAIAdapter {
	// httptest → 127.0.0.1; ValidateOutboundURL блокирует loopback (SSRF).
	// Custom client в hubtest = явный test transport, не user-controlled URL path.
	return &OpenAIAdapter{client: client, skipOutboundURLValidate: true}
}

func ExportNewYandexGPTAdapterWithClient(client *http.Client) *YandexGPTAdapter {
	return &YandexGPTAdapter{client: client}
}

func ExportNewGigaChatAdapterWithClient(client *http.Client) *GigaChatAdapter {
	return &GigaChatAdapter{client: client}
}

func ExportNewAnthropicAdapterWithClient(client *http.Client) *AnthropicAdapter {
	return &AnthropicAdapter{client: client}
}

func ExportYandexCompletionURL() string {
	return yandexCompletionURL
}

func ExportGigaChatAccessToken(a *GigaChatAdapter, ctx context.Context, clientID, secret string) (string, error) {
	return a.accessToken(ctx, clientID, secret)
}

func ExportServiceSetGigaAdapter(s *Service, a *GigaChatAdapter) {
	s.giga = a
}

func ExportServiceSetAnthropicAdapter(s *Service, a *AnthropicAdapter) {
	s.anthropic = a
}

func ExportServiceSetYandexAdapter(s *Service, a *YandexGPTAdapter) {
	s.yandex = a
}

func ExportOpenAIAdapterClient(a *OpenAIAdapter) *http.Client {
	if a == nil {
		return nil
	}
	return a.client
}
