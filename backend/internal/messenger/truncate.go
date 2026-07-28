package messenger

// truncateRunes обрезает s до maxRunes рун (не байт), без суффикса.
// P3-BE-04 / TASK-060: title[:500] по байтам ломал UTF-8 в promote-to-issue.
func truncateRunes(s string, maxRunes int) string {
	if maxRunes < 0 {
		maxRunes = 0
	}
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes])
}
