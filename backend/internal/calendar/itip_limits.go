package calendar

const (
	itipHardMaxBytes             = 1048576
	itipHardMaxUnfoldedLineBytes = 8192
	itipHardMaxProperties        = 4096
	itipHardMaxAttendees         = 200
	itipHardMaxExDates           = 1000
)

// ITIPParseLimits задаёт проверяемые до выделения памяти пределы входного iTIP payload.
type ITIPParseLimits struct {
	MaxBytes             int
	MaxUnfoldedLineBytes int
	MaxProperties        int
	MaxAttendees         int
	MaxExDates           int
}

// DefaultITIPParseLimits возвращает замороженные безопасные пределы входного iTIP.
func DefaultITIPParseLimits() ITIPParseLimits {
	return ITIPParseLimits{
		MaxBytes:             itipHardMaxBytes,
		MaxUnfoldedLineBytes: itipHardMaxUnfoldedLineBytes,
		MaxProperties:        itipHardMaxProperties,
		MaxAttendees:         itipHardMaxAttendees,
		MaxExDates:           itipHardMaxExDates,
	}
}

func (limits ITIPParseLimits) valid() bool {
	return limits.MaxBytes > 0 && limits.MaxBytes <= itipHardMaxBytes &&
		limits.MaxUnfoldedLineBytes > 0 && limits.MaxUnfoldedLineBytes <= itipHardMaxUnfoldedLineBytes &&
		limits.MaxProperties > 0 && limits.MaxProperties <= itipHardMaxProperties &&
		limits.MaxAttendees > 0 && limits.MaxAttendees <= itipHardMaxAttendees &&
		limits.MaxExDates > 0 && limits.MaxExDates <= itipHardMaxExDates
}
