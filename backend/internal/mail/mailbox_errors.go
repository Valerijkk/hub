package mail

import "errors"

var (
	// ErrEngineNotFound обозначает отсутствие объекта в разрешённой области ящика.
	ErrEngineNotFound = errors.New("mail engine object not found")
	// ErrEngineConflict обозначает конфликт состояния или версии движка.
	ErrEngineConflict = errors.New("mail engine state conflict")
	// ErrEngineInvalidCursor обозначает недействительный или истёкший курсор движка.
	ErrEngineInvalidCursor = errors.New("mail engine invalid cursor")
	// ErrEngineUnavailable обозначает недоступность движка.
	ErrEngineUnavailable = errors.New("mail engine unavailable")
	// ErrEngineIndeterminate обозначает неопределённый результат операции движка.
	ErrEngineIndeterminate = errors.New("mail engine indeterminate result")
	// ErrEngineRejectedPolicy обозначает отклонение операции политикой движка.
	ErrEngineRejectedPolicy = errors.New("mail engine policy rejected")
	// ErrEngineUnauthorizedAccount обозначает недопустимую область учётной записи движка.
	ErrEngineUnauthorizedAccount = errors.New("mail engine unauthorized account")
	// ErrEngineRestartRequired обозначает необходимость перезапустить базовую синхронизацию.
	ErrEngineRestartRequired = errors.New("mail engine projection restart required")
)

// EngineErrorKind — безопасный стабильный вид ошибки порта движка.
type EngineErrorKind string

const (
	// EngineErrorNotFound обозначает отсутствие объекта.
	EngineErrorNotFound EngineErrorKind = "not_found"
	// EngineErrorConflict обозначает конфликт версии или состояния.
	EngineErrorConflict EngineErrorKind = "conflict"
	// EngineErrorInvalidCursor обозначает недействительный или истёкший курсор.
	EngineErrorInvalidCursor EngineErrorKind = "invalid_cursor"
	// EngineErrorUnavailable обозначает недоступность движка.
	EngineErrorUnavailable EngineErrorKind = "unavailable"
	// EngineErrorIndeterminate обозначает неопределённый результат операции.
	EngineErrorIndeterminate EngineErrorKind = "indeterminate"
	// EngineErrorRejectedPolicy обозначает отклонение политикой.
	EngineErrorRejectedPolicy EngineErrorKind = "rejected_policy"
	// EngineErrorUnauthorizedAccount обозначает отказ в области учётной записи.
	EngineErrorUnauthorizedAccount EngineErrorKind = "unauthorized_account"
	// EngineErrorRestartRequired обозначает необходимость перезапустить синхронизацию.
	EngineErrorRestartRequired EngineErrorKind = "restart_required"
)

// EngineError — безопасная типизированная ошибка без данных провайдера.
type EngineError struct {
	Kind EngineErrorKind
}

// Error возвращает безопасное стабильное описание ошибки движка.
func (e *EngineError) Error() string {
	if e == nil {
		return "mail engine error"
	}
	return "mail engine " + string(e.Kind)
}

// Unwrap возвращает стабильный sentinel для errors.Is.
func (e *EngineError) Unwrap() error {
	if e == nil {
		return nil
	}
	switch e.Kind {
	case EngineErrorNotFound:
		return ErrEngineNotFound
	case EngineErrorConflict:
		return ErrEngineConflict
	case EngineErrorInvalidCursor:
		return ErrEngineInvalidCursor
	case EngineErrorUnavailable:
		return ErrEngineUnavailable
	case EngineErrorIndeterminate:
		return ErrEngineIndeterminate
	case EngineErrorRejectedPolicy:
		return ErrEngineRejectedPolicy
	case EngineErrorUnauthorizedAccount:
		return ErrEngineUnauthorizedAccount
	case EngineErrorRestartRequired:
		return ErrEngineRestartRequired
	default:
		return nil
	}
}
