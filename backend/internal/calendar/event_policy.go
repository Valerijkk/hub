package calendar

// RSVPResetScope определяет набор неорганизаторов, чей RSVP должен быть сброшен при обновлении.
type RSVPResetScope string

const (
	// RSVPResetNone означает сохранение существующих RSVP.
	RSVPResetNone RSVPResetScope = "none"
	// RSVPResetAllNonOrganizers означает сброс RSVP всех неорганизаторов серии.
	RSVPResetAllNonOrganizers RSVPResetScope = "all_non_organizers"
	// RSVPResetAffectedNonOrganizers означает сброс RSVP только затронутой recurrence-области.
	RSVPResetAffectedNonOrganizers RSVPResetScope = "affected_non_organizers"
)

// CancelIntent определяет, кому адресован iTIP CANCEL для классифицированного обновления.
type CancelIntent string

const (
	// CancelIntentNone означает отсутствие iTIP CANCEL.
	CancelIntentNone CancelIntent = "none"
	// CancelIntentWholeEvent означает отмену всего события или occurrence.
	CancelIntentWholeEvent CancelIntent = "whole_event"
	// CancelIntentRemovedAttendees означает CANCEL только удалённым участникам.
	CancelIntentRemovedAttendees CancelIntent = "removed_attendees"
)

// EventUpdate описывает уже вычисленный набор изменившихся полей без зависимости от хранилища.
type EventUpdate struct {
	TitleOrAgendaChanged bool
	DescriptionChanged   bool
	NotifyAttendees      bool
	TimeChanged          bool
	TimezoneChanged      bool
	AllDayChanged        bool
	LocationChanged      bool
	RecurrenceChanged    bool
	ReminderChanged      bool
	AddedAttendees       bool
	RemovedAttendees     bool
	Cancelled            bool
}

// EventUpdatePolicy содержит детерминированный результат классификации обновления события.
type EventUpdatePolicy struct {
	VersionIncrement         uint
	SequenceIncrement        uint
	RSVPReset                RSVPResetScope
	PreserveExistingRSVP     bool
	InitializeAddedAttendees bool
	SendRequest              bool
	CancelIntent             CancelIntent
}

// ClassifyEventUpdate определяет изменение версии, iTIP sequence, RSVP и уведомлений без побочных эффектов.
func ClassifyEventUpdate(update EventUpdate) EventUpdatePolicy {
	if update.Cancelled {
		return EventUpdatePolicy{
			VersionIncrement:     1,
			SequenceIncrement:    1,
			RSVPReset:            RSVPResetNone,
			PreserveExistingRSVP: true,
			CancelIntent:         CancelIntentWholeEvent,
		}
	}

	policy := EventUpdatePolicy{
		RSVPReset:    RSVPResetNone,
		CancelIntent: CancelIntentNone,
	}
	if !update.hasChanges() {
		return policy
	}
	policy.VersionIncrement = 1

	switch {
	case update.TimeChanged || update.TimezoneChanged || update.AllDayChanged || update.LocationChanged:
		policy.SequenceIncrement = 1
		policy.RSVPReset = RSVPResetAllNonOrganizers
		policy.SendRequest = true
	case update.RecurrenceChanged:
		policy.SequenceIncrement = 1
		policy.RSVPReset = RSVPResetAffectedNonOrganizers
		policy.SendRequest = true
	}

	if update.TitleOrAgendaChanged || (update.DescriptionChanged && update.NotifyAttendees) {
		policy.SequenceIncrement = 1
		policy.SendRequest = true
	}
	if update.AddedAttendees {
		policy.SequenceIncrement = 1
		policy.InitializeAddedAttendees = true
		policy.SendRequest = true
	}
	if update.RemovedAttendees {
		policy.SequenceIncrement = 1
		policy.CancelIntent = CancelIntentRemovedAttendees
	}
	if policy.RSVPReset == RSVPResetNone {
		policy.PreserveExistingRSVP = true
	}

	return policy
}

func (update EventUpdate) hasChanges() bool {
	return update.TitleOrAgendaChanged ||
		update.DescriptionChanged ||
		update.TimeChanged ||
		update.TimezoneChanged ||
		update.AllDayChanged ||
		update.LocationChanged ||
		update.RecurrenceChanged ||
		update.ReminderChanged ||
		update.AddedAttendees ||
		update.RemovedAttendees
}
