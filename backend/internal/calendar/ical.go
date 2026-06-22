package calendar

import (
	"strings"
	"time"
)

// RenderICalendar сериализует события в RFC 5545 iCalendar.
func RenderICalendar(events []EventResponse) string {
	var b strings.Builder
	b.WriteString("BEGIN:VCALENDAR\r\n")
	b.WriteString("VERSION:2.0\r\n")
	b.WriteString("PRODID:-//Hub Platform//Unified Calendar//RU\r\n")
	b.WriteString("CALSCALE:GREGORIAN\r\n")
	b.WriteString("METHOD:PUBLISH\r\n")

	for _, ev := range events {
		b.WriteString("BEGIN:VEVENT\r\n")
		b.WriteString("UID:" + escapeICal(ev.ID) + "@hub.local\r\n")
		b.WriteString("DTSTAMP:" + time.Now().UTC().Format("20060102T150405Z") + "\r\n")

		if ev.AllDay {
			b.WriteString("DTSTART;VALUE=DATE:" + ev.StartsAt.UTC().Format("20060102") + "\r\n")
			if ev.EndsAt != nil {
				// iCal end exclusive для all-day — стандарт RFC 5545.
				endExclusive := ev.EndsAt.UTC().Add(24 * time.Hour)
				b.WriteString("DTEND;VALUE=DATE:" + endExclusive.Format("20060102") + "\r\n")
			}
		} else {
			b.WriteString("DTSTART:" + ev.StartsAt.UTC().Format("20060102T150405Z") + "\r\n")
			if ev.EndsAt != nil {
				b.WriteString("DTEND:" + ev.EndsAt.UTC().Format("20060102T150405Z") + "\r\n")
			}
		}

		b.WriteString("SUMMARY:" + escapeICal(ev.Title) + "\r\n")
		if ev.Description != "" {
			b.WriteString("DESCRIPTION:" + escapeICal(ev.Description) + "\r\n")
		}
		if ev.URL != "" {
			b.WriteString("URL:" + escapeICal(ev.URL) + "\r\n")
		}
		b.WriteString("CATEGORIES:" + escapeICal(string(ev.Type)) + "\r\n")
		b.WriteString("END:VEVENT\r\n")
	}

	b.WriteString("END:VCALENDAR\r\n")
	return b.String()
}

func escapeICal(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, ";", "\\;")
	s = strings.ReplaceAll(s, ",", "\\,")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "")
	return s
}
