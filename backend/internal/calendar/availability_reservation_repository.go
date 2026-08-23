package calendar

import "reflect"

func cloneCalendarAvailabilityReservationSet(set CalendarAvailabilityReservationSet) CalendarAvailabilityReservationSet {
	cloned, err := NewCalendarAvailabilityReservationSet(set.OrganizationID(), set.EventID(), set.OrganizerID(), set.participantsWithoutOrganizer(), set.Occurrences())
	if err != nil {
		return CalendarAvailabilityReservationSet{}
	}
	return cloned
}

func nilAvailabilityReservationTransactionBackend(backend AvailabilityReservationTransactionBackend) bool {
	if backend == nil {
		return true
	}
	value := reflect.ValueOf(backend)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
