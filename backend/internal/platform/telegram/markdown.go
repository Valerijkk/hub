// Package telegram содержит общие хелперы для интеграции с Telegram Bot API.
//
// LO-01: ранее функция экранирования MarkdownV2 дублировалась в
// internal/notification/telegram.go и internal/auth/telegram_bot.go. Теперь —
// единственный источник истины здесь.
package telegram

import "strings"

// markdownV2Special — спецсимволы, требующие экранирования по правилам
// Telegram MarkdownV2 (https://core.telegram.org/bots/api#markdownv2-style).
var markdownV2Special = []string{
	"_", "*", "[", "]", "(", ")", "~", "`", ">",
	"#", "+", "-", "=", "|", "{", "}", ".", "!",
}

// EscapeMarkdownV2 экранирует все спецсимволы для безопасной вставки строки
// в сообщение с parse_mode=MarkdownV2. Любой текст из user-input, который
// попадает в Telegram-сообщение, должен пройти через эту функцию, иначе
// Telegram отвергнет сообщение или интерпретирует часть как разметку.
func EscapeMarkdownV2(s string) string {
	result := s
	for _, ch := range markdownV2Special {
		result = strings.ReplaceAll(result, ch, "\\"+ch)
	}
	return result
}
