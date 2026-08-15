package mail

import (
	"errors"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const maxSendLoopValues = 16

// SendErrorKind определяет безопасную категорию отказа политики отправки.
type SendErrorKind string

const (
	// SendErrorInvalid обозначает недопустимую конфигурацию или форму входа.
	SendErrorInvalid SendErrorKind = "invalid"
	// SendErrorTooLarge обозначает превышение безопасного лимита отправки.
	SendErrorTooLarge SendErrorKind = "too_large"
	// SendErrorDuplicateRecipient обозначает повтор канонического получателя.
	SendErrorDuplicateRecipient SendErrorKind = "duplicate_recipient"
	// SendErrorIdentityNotAllowed обозначает недопустимую identity отправителя.
	SendErrorIdentityNotAllowed SendErrorKind = "identity_not_allowed"
	// SendErrorExternalDisabled обозначает запрещённого внешнего получателя.
	SendErrorExternalDisabled SendErrorKind = "external_disabled"
	// SendErrorAutoResponseSuppressed обозначает подавленный автоматический ответ.
	SendErrorAutoResponseSuppressed SendErrorKind = "auto_response_suppressed"
	// SendErrorRateLimited обозначает исчерпанный лимит отправки.
	SendErrorRateLimited SendErrorKind = "rate_limited"
)

var (
	// ErrSendInvalid сообщает о недопустимом входе без раскрытия его значений.
	ErrSendInvalid = errors.New("mail send invalid")
	// ErrSendTooLarge сообщает о превышении лимита без раскрытия его значений.
	ErrSendTooLarge = errors.New("mail send exceeds bounds")
	// ErrSendDuplicateRecipient сообщает о повторе получателя без раскрытия адреса.
	ErrSendDuplicateRecipient = errors.New("mail send duplicate recipient")
	// ErrSendIdentityNotAllowed сообщает об отклонённой identity отправителя.
	ErrSendIdentityNotAllowed = errors.New("mail send identity not allowed")
	// ErrSendExternalDisabled сообщает о запрещённой внешней доставке.
	ErrSendExternalDisabled = errors.New("mail send external delivery disabled")
	// ErrSendAutoResponseSuppressed сообщает о подавленном автоматическом ответе.
	ErrSendAutoResponseSuppressed = errors.New("mail send auto response suppressed")
	// ErrSendRateLimited сообщает об исчерпанном лимите отправки.
	ErrSendRateLimited = errors.New("mail send rate limited")
)

// SendPolicyError хранит безопасную типизированную причину отказа отправки.
type SendPolicyError struct {
	Kind SendErrorKind
}

// Error возвращает стабильный безопасный текст ошибки отправки.
func (errorValue *SendPolicyError) Error() string {
	if errorValue == nil {
		return "mail send policy error"
	}

	switch errorValue.Kind {
	case SendErrorInvalid:
		return ErrSendInvalid.Error()
	case SendErrorTooLarge:
		return ErrSendTooLarge.Error()
	case SendErrorDuplicateRecipient:
		return ErrSendDuplicateRecipient.Error()
	case SendErrorIdentityNotAllowed:
		return ErrSendIdentityNotAllowed.Error()
	case SendErrorExternalDisabled:
		return ErrSendExternalDisabled.Error()
	case SendErrorAutoResponseSuppressed:
		return ErrSendAutoResponseSuppressed.Error()
	case SendErrorRateLimited:
		return ErrSendRateLimited.Error()
	default:
		return "mail send policy error"
	}
}

// Unwrap возвращает стабильную sentinel-ошибку известной категории.
func (errorValue *SendPolicyError) Unwrap() error {
	if errorValue == nil {
		return nil
	}

	switch errorValue.Kind {
	case SendErrorInvalid:
		return ErrSendInvalid
	case SendErrorTooLarge:
		return ErrSendTooLarge
	case SendErrorDuplicateRecipient:
		return ErrSendDuplicateRecipient
	case SendErrorIdentityNotAllowed:
		return ErrSendIdentityNotAllowed
	case SendErrorExternalDisabled:
		return ErrSendExternalDisabled
	case SendErrorAutoResponseSuppressed:
		return ErrSendAutoResponseSuppressed
	case SendErrorRateLimited:
		return ErrSendRateLimited
	default:
		return nil
	}
}

// SendRateState содержит полностью внедряемое состояние одного token bucket.
type SendRateState struct {
	Tokens          int
	LastRefillAt    time.Time
	RefillRemainder time.Duration
}

// SendAuthorizedIdentity описывает уже разрешённую identity отправителя.
type SendAuthorizedIdentity struct {
	CanonicalAddress string
	Active           bool
	SendAllowed      bool
}

// SendAutoResponseMetadata содержит ограниченные признаки автоматического ответа.
type SendAutoResponseMetadata struct {
	NullReversePath       bool
	AutoSubmitted         string
	Precedence            string
	XLoop                 []string
	XAutoResponseSuppress string
}

// SendPolicyInput содержит все явные значения для детерминированной проверки отправки.
type SendPolicyInput struct {
	Now                   time.Time
	MessageBytes          int64
	PolicyMaxMessageBytes int64
	PolicyMaxRecipients   int
	To                    []EngineAddress
	CC                    []EngineAddress
	BCC                   []EngineAddress
	From                  EngineAddress
	Identity              SendAuthorizedIdentity
	InternalDomains       []string
	ExternalMailEnabled   bool
	Automated             bool
	AutoResponse          SendAutoResponseMetadata
	LoopToken             string
	UserRatePerMinute     int
	UserRateState         SendRateState
	InstanceRatePerMinute int
	InstanceRateState     SendRateState
}

// SendPolicyDecision возвращает независимые нормализованные получатели и состояния bucket-ов.
type SendPolicyDecision struct {
	To                []EngineAddress
	CC                []EngineAddress
	BCC               []EngineAddress
	UserRateState     SendRateState
	InstanceRateState SendRateState
	RetryAfterSeconds int
}

// EvaluateSendPolicy проверяет отправку без ввода-вывода, глобального состояния или часов.
func EvaluateSendPolicy(input SendPolicyInput) (SendPolicyDecision, error) {
	internalDomains, err := validateSendConfiguration(input)
	if err != nil {
		return SendPolicyDecision{}, err
	}

	if input.MessageBytes <= 0 {
		return SendPolicyDecision{}, sendPolicyError(SendErrorInvalid)
	}
	if input.MessageBytes > input.PolicyMaxMessageBytes {
		return SendPolicyDecision{}, sendPolicyError(SendErrorTooLarge)
	}

	totalRecipients := len(input.To) + len(input.CC) + len(input.BCC)
	if totalRecipients == 0 {
		return SendPolicyDecision{}, sendPolicyError(SendErrorInvalid)
	}
	if len(input.To) > MaxEngineDraftRecipientItems || len(input.CC) > MaxEngineDraftRecipientItems || len(input.BCC) > MaxEngineDraftRecipientItems || totalRecipients > input.PolicyMaxRecipients {
		return SendPolicyDecision{}, sendPolicyError(SendErrorTooLarge)
	}

	normalizedTo, normalizedCC, normalizedBCC, err := normalizeSendRecipients(input.To, input.CC, input.BCC)
	if err != nil {
		return SendPolicyDecision{}, err
	}
	canonicalFrom, err := NormalizeAddress(input.From)
	if err != nil {
		return SendPolicyDecision{}, mapAddressPolicyError(err)
	}
	if !sendIdentityAllows(input.Identity, canonicalFrom.Email) {
		return SendPolicyDecision{}, sendPolicyError(SendErrorIdentityNotAllowed)
	}

	if hasExternalRecipient(internalDomains, normalizedTo, normalizedCC, normalizedBCC) && !input.ExternalMailEnabled {
		return SendPolicyDecision{}, sendPolicyError(SendErrorExternalDisabled)
	}
	if input.Automated && suppressesAutoResponse(input.AutoResponse, input.LoopToken) {
		return SendPolicyDecision{}, sendPolicyError(SendErrorAutoResponseSuppressed)
	}

	userState := refillSendRateState(input.UserRateState, input.UserRatePerMinute, input.Now)
	instanceState := refillSendRateState(input.InstanceRateState, input.InstanceRatePerMinute, input.Now)
	if userState.Tokens == 0 || instanceState.Tokens == 0 {
		return SendPolicyDecision{
			UserRateState:     userState,
			InstanceRateState: instanceState,
			RetryAfterSeconds: sendRetryAfterSeconds(userState, input.UserRatePerMinute, instanceState, input.InstanceRatePerMinute),
		}, sendPolicyError(SendErrorRateLimited)
	}

	userState.Tokens--
	instanceState.Tokens--
	return SendPolicyDecision{
		To:                normalizedTo,
		CC:                normalizedCC,
		BCC:               normalizedBCC,
		UserRateState:     userState,
		InstanceRateState: instanceState,
	}, nil
}

func validateSendConfiguration(input SendPolicyInput) (map[string]struct{}, error) {
	if input.Now.IsZero() || input.PolicyMaxMessageBytes < 1_048_576 || input.PolicyMaxMessageBytes > MaxEngineMessageBytes || input.PolicyMaxRecipients < 1 || input.PolicyMaxRecipients > MaxEngineDraftTotalRecipientItems {
		return nil, sendPolicyError(SendErrorInvalid)
	}
	internalDomains, err := normalizeInternalDomains(input.InternalDomains)
	if err != nil {
		return nil, err
	}
	if !validSendRateState(input.UserRateState, input.UserRatePerMinute, input.Now) || !validSendRateState(input.InstanceRateState, input.InstanceRatePerMinute, input.Now) {
		return nil, sendPolicyError(SendErrorInvalid)
	}
	if input.Automated && !validAutoResponseMetadata(input.AutoResponse, input.LoopToken) {
		return nil, sendPolicyError(SendErrorInvalid)
	}
	return internalDomains, nil
}

func normalizeInternalDomains(domains []string) (map[string]struct{}, error) {
	if len(domains) < 1 || len(domains) > 100 {
		return nil, sendPolicyError(SendErrorInvalid)
	}

	canonicalDomains := make(map[string]struct{}, len(domains))
	for _, domain := range domains {
		if len(domain) == 0 || len(domain) > 253 || !isLowerASCII(domain) {
			return nil, sendPolicyError(SendErrorInvalid)
		}
		normalized, err := NormalizeAddress(EngineAddress{Email: "domain-check@" + domain})
		if err != nil || normalized.Email[len("domain-check@"):] != domain {
			return nil, sendPolicyError(SendErrorInvalid)
		}
		if _, exists := canonicalDomains[domain]; exists {
			return nil, sendPolicyError(SendErrorInvalid)
		}
		canonicalDomains[domain] = struct{}{}
	}
	return canonicalDomains, nil
}

func isLowerASCII(value string) bool {
	for index := 0; index < len(value); index++ {
		if value[index] > 0x7f || value[index] >= 'A' && value[index] <= 'Z' {
			return false
		}
	}
	return true
}

func validSendRateState(state SendRateState, rate int, now time.Time) bool {
	return rate >= 1 && rate <= 1000 && state.Tokens >= 0 && state.Tokens <= rate && !state.LastRefillAt.IsZero() && !state.LastRefillAt.After(now) && state.RefillRemainder >= 0 && state.RefillRemainder < time.Minute && (state.Tokens != rate || state.RefillRemainder == 0)
}

func validAutoResponseMetadata(metadata SendAutoResponseMetadata, loopToken string) bool {
	if len(metadata.XLoop) > maxSendLoopValues || !validSendHeaderValue(metadata.AutoSubmitted) || !validOptionalSendToken(metadata.Precedence) || !validOptionalSendToken(loopToken) || !validAutoResponseSuppressTokens(metadata.XAutoResponseSuppress) {
		return false
	}
	for _, value := range metadata.XLoop {
		if !validRequiredSendToken(value) {
			return false
		}
	}
	return true
}

func validSendHeaderValue(value string) bool {
	if !utf8.ValidString(value) || len(value) > MaxEngineHeaderValueBytes {
		return false
	}
	for _, character := range value {
		if character != '\t' && unicode.IsControl(character) {
			return false
		}
	}
	return true
}

func validOptionalSendToken(value string) bool {
	return value == "" || validRequiredSendToken(value)
}

func validRequiredSendToken(value string) bool {
	if !validSendHeaderValue(value) {
		return false
	}
	return validVisibleSendToken(trimASCIIWhitespace(value))
}

func validAutoResponseSuppressTokens(value string) bool {
	if !validSendHeaderValue(value) {
		return false
	}
	if value == "" {
		return true
	}
	for _, token := range strings.Split(value, ",") {
		if !validVisibleSendToken(trimASCIIWhitespace(token)) {
			return false
		}
	}
	return true
}

func validVisibleSendToken(value string) bool {
	if value == "" {
		return false
	}
	for index := 0; index < len(value); index++ {
		if value[index] < 0x21 || value[index] > 0x7e || value[index] == ',' || value[index] == ';' {
			return false
		}
	}
	return true
}

func normalizeSendRecipients(to, cc, bcc []EngineAddress) ([]EngineAddress, []EngineAddress, []EngineAddress, error) {
	seen := make(map[string]struct{}, len(to)+len(cc)+len(bcc))
	normalizedTo, err := normalizeSendRecipientPartition(to, seen)
	if err != nil {
		return nil, nil, nil, err
	}
	normalizedCC, err := normalizeSendRecipientPartition(cc, seen)
	if err != nil {
		return nil, nil, nil, err
	}
	normalizedBCC, err := normalizeSendRecipientPartition(bcc, seen)
	if err != nil {
		return nil, nil, nil, err
	}
	return normalizedTo, normalizedCC, normalizedBCC, nil
}

func normalizeSendRecipientPartition(addresses []EngineAddress, seen map[string]struct{}) ([]EngineAddress, error) {
	normalized := make([]EngineAddress, 0, len(addresses))
	for _, address := range addresses {
		canonicalAddress, err := NormalizeAddress(address)
		if err != nil {
			return nil, mapAddressPolicyError(err)
		}
		if _, exists := seen[canonicalAddress.Email]; exists {
			return nil, sendPolicyError(SendErrorDuplicateRecipient)
		}
		seen[canonicalAddress.Email] = struct{}{}
		normalized = append(normalized, canonicalAddress)
	}
	return normalized, nil
}

func mapAddressPolicyError(err error) error {
	if errors.Is(err, ErrAddressTooLarge) {
		return sendPolicyError(SendErrorTooLarge)
	}
	if errors.Is(err, ErrAddressDuplicate) {
		return sendPolicyError(SendErrorDuplicateRecipient)
	}
	return sendPolicyError(SendErrorInvalid)
}

func sendIdentityAllows(identity SendAuthorizedIdentity, fromAddress string) bool {
	if !identity.Active || !identity.SendAllowed {
		return false
	}
	canonicalIdentity, err := NormalizeAddress(EngineAddress{Email: identity.CanonicalAddress})
	return err == nil && canonicalIdentity.Email == identity.CanonicalAddress && canonicalIdentity.Email == fromAddress
}

func hasExternalRecipient(internalDomains map[string]struct{}, to, cc, bcc []EngineAddress) bool {
	for _, partition := range [][]EngineAddress{to, cc, bcc} {
		for _, address := range partition {
			if _, exists := internalDomains[address.Email[strings.LastIndexByte(address.Email, '@')+1:]]; !exists {
				return true
			}
		}
	}
	return false
}

func suppressesAutoResponse(metadata SendAutoResponseMetadata, loopToken string) bool {
	if metadata.NullReversePath {
		return true
	}
	if metadata.AutoSubmitted != "" && !strings.EqualFold(trimASCIIWhitespace(metadata.AutoSubmitted), "no") {
		return true
	}
	precedence := trimASCIIWhitespace(metadata.Precedence)
	if strings.EqualFold(precedence, "bulk") || strings.EqualFold(precedence, "list") || strings.EqualFold(precedence, "junk") {
		return true
	}
	canonicalLoopToken := trimASCIIWhitespace(loopToken)
	if canonicalLoopToken != "" {
		for _, value := range metadata.XLoop {
			if strings.EqualFold(trimASCIIWhitespace(value), canonicalLoopToken) {
				return true
			}
		}
	}
	for _, token := range strings.Split(metadata.XAutoResponseSuppress, ",") {
		canonicalToken := trimASCIIWhitespace(token)
		if strings.EqualFold(canonicalToken, "all") || strings.EqualFold(canonicalToken, "autoreply") || strings.EqualFold(canonicalToken, "autorespond") {
			return true
		}
	}
	return false
}

func trimASCIIWhitespace(value string) string {
	return strings.Trim(value, " \t")
}

func refillSendRateState(state SendRateState, rate int, now time.Time) SendRateState {
	if state.Tokens == rate || now.Sub(state.LastRefillAt) >= time.Minute {
		return SendRateState{Tokens: rate, LastRefillAt: now}
	}

	elapsedNanoseconds := now.Sub(state.LastRefillAt).Nanoseconds()
	refillUnits := elapsedNanoseconds*int64(rate) + int64(state.RefillRemainder)
	state.Tokens += int(refillUnits / int64(time.Minute))
	state.RefillRemainder = time.Duration(refillUnits % int64(time.Minute))
	state.LastRefillAt = now
	if state.Tokens >= rate {
		return SendRateState{Tokens: rate, LastRefillAt: now}
	}
	return state
}

func sendRetryAfterSeconds(userState SendRateState, userRate int, instanceState SendRateState, instanceRate int) int {
	maximum := int64(1)
	for _, bucket := range []struct {
		state SendRateState
		rate  int
	}{
		{state: userState, rate: userRate},
		{state: instanceState, rate: instanceRate},
	} {
		if bucket.state.Tokens != 0 {
			continue
		}
		numerator := int64(time.Minute) - int64(bucket.state.RefillRemainder)
		denominator := int64(bucket.rate) * int64(time.Second)
		seconds := (numerator + denominator - 1) / denominator
		if seconds > maximum {
			maximum = seconds
		}
	}
	if maximum > 60 {
		return 60
	}
	return int(maximum)
}

func sendPolicyError(kind SendErrorKind) error {
	return &SendPolicyError{Kind: kind}
}
