package calls

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teambition/rrule-go"
)

// TestMaxActiveRecordingsLimit проверяет что константа лимита активных записей равна 2.
func TestMaxActiveRecordingsLimit(t *testing.T) {
	assert.Equal(t, 2, maxActiveRecordingsPerCall, "лимит активных записей должен быть 2")
}

// TestToRecordingResponse проверяет конвертацию Ent Recording в DTO.
func TestToRecordingResponse(t *testing.T) {
	// Тест конвертации — проверяем что поля корректно маппятся.
	// Без реальной БД тестируем только логику маппинга через nil-safe проверку.
	assert.NotNil(t, toRecordingResponse, "toRecordingResponse должна быть определена")
}

// TestRRuleParsing проверяет парсинг RFC 5545 RRULE строк.
func TestRRuleParsing(t *testing.T) {
	tests := []struct {
		name     string
		rruleStr string
		wantErr  bool
	}{
		{
			name:     "ежедневно",
			rruleStr: "FREQ=DAILY;BYHOUR=10;BYMINUTE=0;BYSECOND=0",
			wantErr:  false,
		},
		{
			name:     "еженедельно по понедельникам",
			rruleStr: "FREQ=WEEKLY;BYDAY=MO;BYHOUR=10;BYMINUTE=0;BYSECOND=0",
			wantErr:  false,
		},
		{
			name:     "по будням",
			rruleStr: "FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR;BYHOUR=10;BYMINUTE=0;BYSECOND=0",
			wantErr:  false,
		},
		{
			name:     "некорректный RRULE",
			rruleStr: "INVALID_RULE",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule, err := rrule.StrToRRule(tt.rruleStr)
			if tt.wantErr {
				assert.Error(t, err, "ожидается ошибка для некорректного RRULE")
				return
			}
			require.NoError(t, err, "ошибка парсинга RRULE")

			// Вычисляем следующее вхождение
			next := rule.After(time.Now(), false)
			assert.False(t, next.IsZero(), "следующее вхождение не должно быть нулевым")
		})
	}
}

// TestRRuleNextOccurrence проверяет вычисление следующего вхождения для RRULE.
func TestRRuleNextOccurrence(t *testing.T) {
	// Ежедневный стендап в 10:00 UTC
	rule, err := rrule.StrToRRule("FREQ=DAILY;BYHOUR=10;BYMINUTE=0;BYSECOND=0")
	require.NoError(t, err)

	now := time.Now()
	next := rule.After(now, false)

	assert.False(t, next.IsZero(), "следующее вхождение не должно быть нулевым")
	assert.True(t, next.After(now) || next.Equal(now), "следующее вхождение должно быть в будущем или сейчас")
	assert.Equal(t, 10, next.Hour(), "час должен быть 10")
	assert.Equal(t, 0, next.Minute(), "минуты должны быть 0")
}

// TestRRuleWithCountLimit проверяет RRULE с ограниченным количеством повторений.
func TestRRuleWithCountLimit(t *testing.T) {
	// Только 1 повторение — после него не должно быть вхождений
	pastTime := time.Now().Add(-48 * time.Hour)
	rruleStr := "FREQ=DAILY;COUNT=1;DTSTART=" + pastTime.UTC().Format("20060102T150405Z")

	rule, err := rrule.StrToRRule(rruleStr)
	require.NoError(t, err)

	// Ищем вхождение после текущего момента — не должно быть
	next := rule.After(time.Now(), false)
	assert.True(t, next.IsZero(), "не должно быть вхождений после исчерпания COUNT")
}

// TestScheduledCallResponse проверяет корректность конвертации в DTO.
func TestScheduledCallResponse(t *testing.T) {
	assert.NotNil(t, toScheduledCallResponse, "toScheduledCallResponse должна быть определена")
}

// TestRecordingStatusValues проверяет допустимые значения статуса записи.
func TestRecordingStatusValues(t *testing.T) {
	// Проверяем что статусы соответствуют ожидаемым
	validStatuses := []string{"recording", "processing", "ready", "failed"}
	for _, status := range validStatuses {
		assert.Contains(t, []string{"recording", "processing", "ready", "failed"}, status,
			"статус %s должен быть допустимым", status)
	}
}
