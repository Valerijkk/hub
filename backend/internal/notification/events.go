package notification

import "strings"

// Константы типов уведомлений.
const (
	// TypeAssigned — назначение задачи на пользователя.
	TypeAssigned = "assigned"
	// TypeMentioned — упоминание пользователя в комментарии.
	TypeMentioned = "mentioned"
	// TypeStatusChanged — смена статуса задачи.
	TypeStatusChanged = "status_changed"
	// TypeCommentAdded — новый комментарий к назначенной задаче.
	TypeCommentAdded = "comment_added"
	// TypeMessageReceived — личное сообщение или упоминание в канале.
	TypeMessageReceived = "message_received"
	// TypeCallStarted — звонок в канале пользователя.
	TypeCallStarted = "call_started"
	// TypeRunCompleted — завершение тестового прогона.
	TypeRunCompleted = "run_completed"
	// TypeDeadlineApproaching — дедлайн задачи через 24ч (cron).
	TypeDeadlineApproaching = "deadline_approaching"
	// TypeBlockerAdded — добавлена блокирующая связь.
	TypeBlockerAdded = "blocker_added"
	// TypeNewLogin — новый вход в аккаунт (другое устройство/IP).
	TypeNewLogin = "new_login"
)

// Константы каналов доставки уведомлений.
const (
	// ChannelInApp — доставка в приложение (колокольчик).
	ChannelInApp = "in_app"
	// ChannelTelegram — доставка через Telegram-бот.
	ChannelTelegram = "telegram"
)

// SubjectToType сопоставляет NATS subject с типом уведомления.
// Возвращает пустую строку, если subject не требует уведомления.
func SubjectToType(subject string) string {
	parts := strings.Split(subject, ".")

	// project.{id}.issue.assigned
	if matchSuffix(parts, "issue.assigned") {
		return TypeAssigned
	}
	// project.{id}.issue.moved
	if matchSuffix(parts, "issue.moved") {
		return TypeStatusChanged
	}
	// project.{id}.issue.comment_created
	if matchSuffix(parts, "issue.comment_created") {
		return TypeCommentAdded
	}
	// project.{id}.issue.link_created (blocks)
	if matchSuffix(parts, "issue.link_created") {
		return TypeBlockerAdded
	}
	// messenger.channel.{id}.message_sent
	if len(parts) >= 4 && parts[0] == "messenger" && parts[3] == "message_sent" {
		return TypeMessageReceived
	}
	// calls.channel.{id}.started
	if len(parts) >= 4 && parts[0] == "calls" && parts[3] == "started" {
		return TypeCallStarted
	}
	// tms.testrun.completed
	if subject == "tms.testrun.completed" {
		return TypeRunCompleted
	}
	// auth.session.created
	if subject == "auth.session.created" {
		return TypeNewLogin
	}

	return ""
}

// matchSuffix проверяет совпадение последних элементов subject.
func matchSuffix(parts []string, suffix string) bool {
	sp := strings.Split(suffix, ".")
	if len(parts) < len(sp) {
		return false
	}
	tail := parts[len(parts)-len(sp):]
	for i, s := range sp {
		if tail[i] != s {
			return false
		}
	}
	return true
}
