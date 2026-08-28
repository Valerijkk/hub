//go:build release

// Заглушка пакета dev для release-сборки.
//
// ME-09: при сборке бинарника с `-tags release` все dev-эндпоинты
// (seed, login, status) физически исключаются из компиляции — единственный способ
// «включить» их — пересобрать без тега `release`, что
// невозможно сделать в рантайме. Это строже, чем проверка
// переменной окружения DEV_MODE (которую может изменить любой, кто редактирует
// манифест деплоя).
//
// Тип dev.Handler и методы NewHandler / RegisterRoutes
// переобъявлены здесь как no-op, чтобы cmd/server/main.go продолжал
// компилироваться с обоими наборами тегов без изменений кода.

package dev

import (
	"github.com/gofiber/fiber/v3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/auth"
)

// Handler — заглушка для release-сборки. Все методы намеренно пустые.
type Handler struct{}

// NewHandler возвращает заглушку. db и authSvc принимаются для сохранения
// идентичного API с dev-сборкой NewHandler, но ни один из них не сохраняется —
// в release-сборке намеренно нет dev-функциональности.
func NewHandler(_ *ent.Client, _ *auth.AuthService, _ MailCalendarDemoSeeder) *Handler {
	return &Handler{}
}

// RegisterRoutes — no-op в release-сборках. /api/dev/* просто
// не существует.
func (h *Handler) RegisterRoutes(_ fiber.Router, _ bool) {
	// без действий
}
