package ai

import (
	"fmt"
	"regexp"
	"strings"
)

const botSummarizeLimit = 50

var (
	botCommandRe = regexp.MustCompile(`(?i)^/ai\s+(summarize|test-cases|explain|translate|review|docs|sql|help)\b(.*)$`)
	prURLInArgs  = regexp.MustCompile(`(?i)https?://(?:github\.com/[^\s]+/pull/\d+|[^\s]+/[^\s]+/-/merge_requests/\d+)`)
)

type botCommand struct {
	name string
	args string
}

func parseBotCommand(prompt string) *botCommand {
	prompt = strings.TrimSpace(prompt)
	m := botCommandRe.FindStringSubmatch(prompt)
	if len(m) < 3 {
		return nil
	}
	return &botCommand{
		name: strings.ToLower(strings.TrimSpace(m[1])),
		args: strings.TrimSpace(m[2]),
	}
}

func buildSummarizePrompt(messages []string) (string, string) {
	system := `Ты AI-ассистент команды разработки. Сделай краткую структурированную выжимку обсуждения на русском.
Выдели: ключевые решения, action items, открытые вопросы. Формат — Markdown с заголовками.`
	var b strings.Builder
	for i, m := range messages {
		if strings.TrimSpace(m) == "" {
			continue
		}
		fmt.Fprintf(&b, "[%d] %s\n", i+1, m)
	}
	transcript := strings.TrimSpace(b.String())
	if transcript == "" {
		return system, "Нет сообщений для суммаризации."
	}
	user := "Суммаризируй следующее обсуждение:\n\n" + transcript
	return system, user
}

func buildTestCasesPrompt(issueKey, title, description string) (string, string) {
	system := `Ты QA-инженер. Сгенерируй тест-кейсы для задачи в Markdown.
Для каждого кейса: **Название**, **Предусловия**, **Шаги** (нумерованный список), **Ожидаемый результат**.
Покрой happy path, граничные случаи и негативные сценарии.`
	desc := strings.TrimSpace(description)
	if desc == "" {
		desc = "(описание отсутствует)"
	}
	user := fmt.Sprintf("Задача: %s — %s\n\nОписание:\n%s", issueKey, title, desc)
	return system, user
}

func buildExplainPrompt(code string) (string, string) {
	system := `Объясни код понятно на русском. Опиши назначение, ключевые части и возможные edge cases. Формат — Markdown.`
	if code == "" {
		return system, "Объясни, что делает этот фрагмент кода (код не указан — попроси пользователя прислать код)."
	}
	return system, "Объясни этот код:\n\n```\n" + code + "\n```"
}

func buildTranslatePrompt(text string) (string, string) {
	system := `Переведи текст между русским и английским (определи направление автоматически). Сохрани технические термины. Ответ — только перевод.`
	if text == "" {
		return system, "Укажи текст для перевода после команды /ai translate"
	}
	return system, text
}

func buildDocsPrompt(context string, query string) (string, string) {
	system := `Ты AI-ассистент Hub. Ответь на вопрос пользователя, опираясь на найденный контекст из задач, тест-кейсов и сообщений команды.
Если контекста недостаточно — скажи об этом. Формат — Markdown.`
	user := fmt.Sprintf("Вопрос: %s\n\nКонтекст из Hub:\n%s", query, context)
	return system, user
}

func buildSQLPrompt(question string) (string, string) {
	system := `Ты SQL-эксперт PostgreSQL. Сгенерируй безопасный SELECT-запрос по описанию на русском.
Hub использует PostgreSQL + Ent ORM. Основные таблицы: issues, projects, users, test_cases.
Только SELECT, без INSERT/UPDATE/DELETE. Добавь комментарий к запросу.`
	if strings.TrimSpace(question) == "" {
		return system, "Опиши, какие данные нужны, после `/ai sql`"
	}
	return system, question
}

func buildHelpPrompt() (string, string) {
	system := DefaultBotSystemPrompt
	user := `Перечисли доступные slash-команды @hub-ai кратко с примерами:
/ai summarize, /ai test-cases KEY, /ai review URL, /ai docs вопрос, /ai explain code, /ai translate text, /ai sql вопрос`
	return system, user
}

func extractPRURL(args string) string {
	m := prURLInArgs.FindString(args)
	return strings.TrimSpace(m)
}

func extractIssueKey(args string) string {
	issueKeyRe := regexp.MustCompile(`(?i)([A-Z][A-Z0-9]+-\d+)`)
	m := issueKeyRe.FindStringSubmatch(args)
	if len(m) < 2 {
		return ""
	}
	return strings.ToUpper(m[1])
}
