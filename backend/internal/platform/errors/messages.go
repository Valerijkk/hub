package errors

// Стандартные сообщения об ошибках для единообразия API-ответов.
const (
	// MsgBadRequest — некорректное тело запроса.
	MsgBadRequest = "некорректное тело запроса"
	// MsgUnauthorized — пользователь не авторизован.
	MsgUnauthorized = "не авторизован"
	// MsgForbidden — недостаточно прав.
	MsgForbidden = "недостаточно прав"
	// MsgNotFound — ресурс не найден.
	MsgNotFound = "не найдено"
	// MsgInternal — внутренняя ошибка сервера.
	MsgInternal = "внутренняя ошибка сервера"
	// MsgConflict — конфликт данных (дубликат и т.д.).
	MsgConflict = "конфликт данных"
)
