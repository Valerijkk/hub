package calendar

import (
	"errors"
	"time"
)

var (
	// ErrInvalidITIPMessage означает, что обязательное поле iTIP-сообщения отсутствует или содержит недопустимое значение.
	ErrInvalidITIPMessage = errors.New("calendar: invalid iTIP message")
	// ErrInvalidITIPMethod означает неподдерживаемый iTIP-метод.
	ErrInvalidITIPMethod = errors.New("calendar: invalid iTIP method")
	// ErrInvalidITIPStatus означает недопустимый статус события или его сочетание с методом.
	ErrInvalidITIPStatus = errors.New("calendar: invalid iTIP status")
	// ErrInvalidITIPParticipant означает недопустимого участника или сочетание участников с методом.
	ErrInvalidITIPParticipant = errors.New("calendar: invalid iTIP participant")
	// ErrInvalidITIPTemporal означает, что временной union не выбран ровно в одном варианте.
	ErrInvalidITIPTemporal = errors.New("calendar: invalid iTIP temporal value")
	// ErrInvalidITIPRecurrence означает некорректный RECURRENCE-ID или EXDATE для формы события.
	ErrInvalidITIPRecurrence = errors.New("calendar: invalid iTIP recurrence value")
	// ErrInvalidITIPStamp означает отсутствующий либо не-UTC DTSTAMP.
	ErrInvalidITIPStamp = errors.New("calendar: invalid iTIP stamp")
	// ErrInvalidITIPSequence означает отрицательный SEQUENCE.
	ErrInvalidITIPSequence = errors.New("calendar: invalid iTIP sequence")
	// ErrInvalidITIPEmail означает некорректный адрес организатора или участника.
	ErrInvalidITIPEmail = errors.New("calendar: invalid iTIP email")
	// ErrInvalidITIPLimits означает некорректную конфигурацию ограничений входного iTIP.
	ErrInvalidITIPLimits = errors.New("calendar: invalid iTIP limits")
	// ErrITIPLimitExceeded означает превышение одного из ограничений входного iTIP.
	ErrITIPLimitExceeded = errors.New("calendar: iTIP limit exceeded")
	// ErrInvalidITIPSyntax означает синтаксически некорректный iCalendar payload.
	ErrInvalidITIPSyntax = errors.New("calendar: invalid iTIP syntax")
	// ErrUnsupportedITIPContent означает корректную, но неподдерживаемую конструкцию iCalendar.
	ErrUnsupportedITIPContent = errors.New("calendar: unsupported iTIP content")
)

// ITIPMethod задаёт поддерживаемый метод календарного iTIP-сообщения.
type ITIPMethod string

const (
	// ITIPMethodRequest означает приглашение либо обновление события.
	ITIPMethodRequest ITIPMethod = "REQUEST"
	// ITIPMethodCancel означает отмену серии либо отдельного экземпляра.
	ITIPMethodCancel ITIPMethod = "CANCEL"
	// ITIPMethodReply означает ответ одного участника на приглашение.
	ITIPMethodReply ITIPMethod = "REPLY"
)

// ITIPStatus задаёт RFC 5545 STATUS события в сериализуемом сообщении.
type ITIPStatus string

const (
	// ITIPStatusConfirmed означает подтверждённое событие.
	ITIPStatusConfirmed ITIPStatus = "CONFIRMED"
	// ITIPStatusTentative означает предварительный статус события.
	ITIPStatusTentative ITIPStatus = "TENTATIVE"
	// ITIPStatusCancelled означает отменённое событие.
	ITIPStatusCancelled ITIPStatus = "CANCELLED"
)

// ITIPParticipantRole задаёт роль участника в календарном событии.
type ITIPParticipantRole string

const (
	// ITIPParticipantRequired означает обязательного участника.
	ITIPParticipantRequired ITIPParticipantRole = "required"
	// ITIPParticipantOptional означает необязательного участника.
	ITIPParticipantOptional ITIPParticipantRole = "optional"
)

// ITIPOrganizer содержит отображаемое имя и email организатора события.
type ITIPOrganizer struct {
	Email       string
	DisplayName string
}

// ITIPParticipant содержит сериализуемые свойства участника события.
type ITIPParticipant struct {
	Email       string
	DisplayName string
	Role        ITIPParticipantRole
	Status      RSVPStatus
	RSVP        bool
}

// ITIPTimed задаёт timed-вариант temporal union через два точных момента и IANA-зону.
type ITIPTimed struct {
	StartsAt time.Time
	EndsAt   time.Time
	Timezone string
}

// ITIPAllDay задаёт all-day-вариант temporal union через даты с исключающим окончанием.
type ITIPAllDay struct {
	StartDate        string
	EndDateExclusive string
	Timezone         string
}

// ITIPTemporal хранит ровно один временной вариант: Timed либо AllDay.
type ITIPTemporal struct {
	Timed  *ITIPTimed
	AllDay *ITIPAllDay
}

// ITIPRecurrenceID представляет один canonical RECURRENCE-ID или EXDATE в форме события.
type ITIPRecurrenceID struct {
	Time time.Time
	Date string
}

// ITIPMessage содержит все явные входы чистой сериализации RFC 5545/5546.
type ITIPMessage struct {
	Method       ITIPMethod
	UID          string
	Stamp        time.Time
	Sequence     int
	Status       ITIPStatus
	Summary      string
	Description  string
	Location     string
	Temporal     ITIPTemporal
	Organizer    ITIPOrganizer
	Attendees    []ITIPParticipant
	RRULE        string
	EXDates      []ITIPRecurrenceID
	RecurrenceID *ITIPRecurrenceID
}

// ITIPIdentity содержит canonical duplicate identity разобранного iTIP-сообщения.
type ITIPIdentity struct {
	CanonicalVersion uint8
	Method           ITIPMethod
	UID              string
	HasRecurrenceID  bool
	RecurrenceID     string
	Sequence         int
	PayloadHash      [32]byte
}
