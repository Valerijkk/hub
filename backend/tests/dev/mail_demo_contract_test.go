package dev_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/devdemo"
)

func TestParseMailDemoRuntimeConfig_Selection(t *testing.T) {
	testCases := []struct {
		name        string
		present     bool
		value       string
		devMode     bool
		wantEnabled bool
		wantError   bool
	}{
		{name: "absent is disabled", devMode: true},
		{name: "empty is disabled", present: true, devMode: true},
		{name: "false is disabled", present: true, value: "false"},
		{name: "true enables only development mode", present: true, value: "true", devMode: true, wantEnabled: true},
		{name: "true rejects non development mode", present: true, value: "true", wantError: true},
		{name: "uppercase true is invalid", present: true, value: "TRUE", devMode: true, wantError: true},
		{name: "numeric selector is invalid", present: true, value: "1", devMode: true, wantError: true},
		{name: "whitespace selector is invalid", present: true, value: " true", devMode: true, wantError: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			config, err := devdemo.ParseMailDemoRuntimeConfig(func(key string) (string, bool) {
				if key != "MAILBOX_DEMO_ENABLED" || !testCase.present {
					return "", false
				}
				return testCase.value, true
			}, testCase.devMode)

			if testCase.wantError {
				require.Error(t, err)
				assert.False(t, config.Enabled)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, testCase.wantEnabled, config.Enabled)
		})
	}
}

func TestNewMailCalendarDemoSeederWithMail_RejectsTypedNilMailPort(t *testing.T) {
	var mailPort *mailDemoPortFake

	seeder, err := devdemo.NewMailCalendarDemoSeederWithMail(nil, nil, false, mailDemoCalendarPortFake{}, mailPort)

	require.ErrorContains(t, err, "mail port is required")
	assert.Nil(t, seeder)
}

type mailDemoCalendarPortFake struct{}

func (mailDemoCalendarPortFake) CreateEvent(context.Context, uuid.UUID, devdemo.CalendarDemoCreateInput) (devdemo.CalendarDemoCreateResult, error) {
	return devdemo.CalendarDemoCreateResult{}, nil
}

func (mailDemoCalendarPortFake) GetEvent(context.Context, uuid.UUID, uuid.UUID) (devdemo.CalendarDemoEventState, error) {
	return devdemo.CalendarDemoEventState{}, nil
}

func (mailDemoCalendarPortFake) RespondSeries(context.Context, uuid.UUID, devdemo.CalendarDemoRSVPInput) (devdemo.CalendarDemoRSVPResult, error) {
	return devdemo.CalendarDemoRSVPResult{}, nil
}

func (mailDemoCalendarPortFake) CancelOccurrence(context.Context, uuid.UUID, devdemo.CalendarDemoOccurrenceCancelInput) (devdemo.CalendarDemoOccurrenceCancelResult, error) {
	return devdemo.CalendarDemoOccurrenceCancelResult{}, nil
}

type mailDemoPortFake struct{}

func (*mailDemoPortFake) PrepareMailDemo(context.Context, devdemo.MailDemoPrepareInput) error {
	return nil
}

func (*mailDemoPortFake) ReconcileMailDemo(context.Context, devdemo.MailDemoInput) (devdemo.MailDemoResult, error) {
	return devdemo.MailDemoResult{}, nil
}
