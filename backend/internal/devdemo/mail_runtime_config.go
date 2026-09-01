package devdemo

import "errors"

const mailDemoRuntimeEnabledEnv = "MAILBOX_DEMO_ENABLED"

// MailDemoRuntimeConfig описывает только разрешённое включение локального Mail demo.
type MailDemoRuntimeConfig struct {
	Enabled bool
}

// ParseMailDemoRuntimeConfig принимает только явный selector: отсутствующий,
// пустой и exact false выключают demo, а exact true допустим только в dev mode.
func ParseMailDemoRuntimeConfig(lookup func(string) (string, bool), devMode bool) (MailDemoRuntimeConfig, error) {
	if lookup == nil {
		return MailDemoRuntimeConfig{}, errors.New("mail demo runtime: environment lookup is required")
	}

	selected, present := lookup(mailDemoRuntimeEnabledEnv)
	switch {
	case !present || selected == "" || selected == "false":
		return MailDemoRuntimeConfig{}, nil
	case selected != "true":
		return MailDemoRuntimeConfig{}, errors.New("mail demo runtime: MAILBOX_DEMO_ENABLED must be true or false when set")
	case !devMode:
		return MailDemoRuntimeConfig{}, errors.New("mail demo runtime: MAILBOX_DEMO_ENABLED requires dev mode")
	default:
		return MailDemoRuntimeConfig{Enabled: true}, nil
	}
}
