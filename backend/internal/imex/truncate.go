package imex

// truncateCell обрезает значение ячейки PDF до maxRunes рун.
// При обрезке оставляет место под "..." (maxRunes-3 рун + суффикс).
// P3-BE-04 / TASK-060: cellVal[:47] по байтам ломал кириллицу/emoji.
func truncateCell(s string, maxRunes int) string {
	if maxRunes < 0 {
		maxRunes = 0
	}
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	keep := maxRunes - 3
	if keep < 0 {
		keep = 0
	}
	return string(runes[:keep]) + "..."
}
