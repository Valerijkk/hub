package search

// truncateRunes обрезает s до maxRunes рун (не байт).
// Если строка длиннее — возвращает префикс + "...".
// P3-BE-04 / TASK-060: s[:N] режет UTF-8 посередине руны (emoji/кириллица).
func truncateRunes(s string, maxRunes int) string {
	if maxRunes < 0 {
		maxRunes = 0
	}
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes]) + "..."
}
