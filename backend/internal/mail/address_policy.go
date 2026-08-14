package mail

import (
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/net/idna"
)

// AddressErrorKind определяет безопасную категорию ошибки адресной политики.
type AddressErrorKind string

const (
	// AddressErrorInvalid обозначает синтаксически недопустимый адрес.
	AddressErrorInvalid AddressErrorKind = "invalid"
	// AddressErrorTooLarge обозначает адрес или набор адресов, превышающий лимит.
	AddressErrorTooLarge AddressErrorKind = "too_large"
	// AddressErrorDuplicate обозначает повтор канонического адреса в наборе.
	AddressErrorDuplicate AddressErrorKind = "duplicate"
)

var (
	// ErrAddressInvalid сообщает о недопустимом адресе без раскрытия входных данных.
	ErrAddressInvalid = errors.New("mail address invalid")
	// ErrAddressTooLarge сообщает о превышении лимитов адреса без раскрытия входных данных.
	ErrAddressTooLarge = errors.New("mail address exceeds bounds")
	// ErrAddressDuplicate сообщает о повторе адреса без раскрытия входных данных.
	ErrAddressDuplicate = errors.New("mail address duplicate")
)

// AddressPolicyError сохраняет безопасную типизированную причину отказа адресной политики.
type AddressPolicyError struct {
	Kind AddressErrorKind
}

// Error возвращает постоянный безопасный текст ошибки адресной политики.
func (errorValue *AddressPolicyError) Error() string {
	if errorValue == nil {
		return "mail address policy error"
	}

	switch errorValue.Kind {
	case AddressErrorInvalid:
		return ErrAddressInvalid.Error()
	case AddressErrorTooLarge:
		return ErrAddressTooLarge.Error()
	case AddressErrorDuplicate:
		return ErrAddressDuplicate.Error()
	default:
		return "mail address policy error"
	}
}

// Unwrap возвращает стабильную sentinel-ошибку известной категории.
func (errorValue *AddressPolicyError) Unwrap() error {
	if errorValue == nil {
		return nil
	}

	switch errorValue.Kind {
	case AddressErrorInvalid:
		return ErrAddressInvalid
	case AddressErrorTooLarge:
		return ErrAddressTooLarge
	case AddressErrorDuplicate:
		return ErrAddressDuplicate
	default:
		return nil
	}
}

// NormalizeAddress проверяет и канонизирует адрес движка без изменения отображаемого имени.
func NormalizeAddress(address EngineAddress) (EngineAddress, error) {
	if !utf8.ValidString(address.Email) {
		return EngineAddress{}, addressPolicyError(AddressErrorInvalid)
	}
	if len(address.Email) > MaxEngineAddressEmailBytes {
		return EngineAddress{}, addressPolicyError(AddressErrorTooLarge)
	}
	if err := validateAddressName(address.Name); err != nil {
		return EngineAddress{}, err
	}
	for _, character := range address.Email {
		if unicode.IsSpace(character) || unicode.IsControl(character) {
			return EngineAddress{}, addressPolicyError(AddressErrorInvalid)
		}
	}

	atIndex := strings.IndexByte(address.Email, '@')
	if atIndex <= 0 || atIndex != strings.LastIndexByte(address.Email, '@') || atIndex == len(address.Email)-1 {
		return EngineAddress{}, addressPolicyError(AddressErrorInvalid)
	}

	localPart := address.Email[:atIndex]
	domainPart := address.Email[atIndex+1:]
	if len(localPart) > 64 {
		return EngineAddress{}, addressPolicyError(AddressErrorTooLarge)
	}
	if !validLocalPart(localPart) {
		return EngineAddress{}, addressPolicyError(AddressErrorInvalid)
	}

	canonicalDomain, err := idna.Lookup.ToASCII(domainPart)
	if err != nil {
		return EngineAddress{}, addressPolicyError(AddressErrorInvalid)
	}
	canonicalDomain = strings.ToLower(canonicalDomain)
	if err := validateDomain(canonicalDomain); err != nil {
		return EngineAddress{}, err
	}

	canonicalEmail := strings.ToLower(localPart) + "@" + canonicalDomain
	if len(canonicalEmail) > MaxEngineAddressEmailBytes {
		return EngineAddress{}, addressPolicyError(AddressErrorTooLarge)
	}

	return EngineAddress{Email: canonicalEmail, Name: address.Name}, nil
}

// NormalizeAddressSet канонизирует упорядоченный набор уникальных адресов без изменения входного среза.
func NormalizeAddressSet(addresses []EngineAddress, maximum int) ([]EngineAddress, error) {
	if maximum < 1 || maximum > MaxEngineAddressItems || len(addresses) > maximum {
		return nil, addressPolicyError(AddressErrorTooLarge)
	}

	normalized := make([]EngineAddress, 0, len(addresses))
	seen := make(map[string]struct{}, len(addresses))
	for _, address := range addresses {
		canonicalAddress, err := NormalizeAddress(address)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[canonicalAddress.Email]; exists {
			return nil, addressPolicyError(AddressErrorDuplicate)
		}
		seen[canonicalAddress.Email] = struct{}{}
		normalized = append(normalized, canonicalAddress)
	}

	return normalized, nil
}

func addressPolicyError(kind AddressErrorKind) error {
	return &AddressPolicyError{Kind: kind}
}

func validateAddressName(name string) error {
	if !utf8.ValidString(name) {
		return addressPolicyError(AddressErrorInvalid)
	}
	if len(name) > MaxEngineAddressNameBytes || utf8.RuneCountInString(name) > 160 {
		return addressPolicyError(AddressErrorTooLarge)
	}
	for _, character := range name {
		if unicode.IsControl(character) {
			return addressPolicyError(AddressErrorInvalid)
		}
	}
	return nil
}

func validLocalPart(localPart string) bool {
	if len(localPart) == 0 || localPart[0] == '.' || localPart[len(localPart)-1] == '.' {
		return false
	}

	previousDot := false
	for index := 0; index < len(localPart); index++ {
		character := localPart[index]
		if character == '.' {
			if previousDot {
				return false
			}
			previousDot = true
			continue
		}
		if !isAText(character) {
			return false
		}
		previousDot = false
	}
	return true
}

func isAText(character byte) bool {
	return character >= 'a' && character <= 'z' ||
		character >= 'A' && character <= 'Z' ||
		character >= '0' && character <= '9' ||
		strings.ContainsRune("!#$%&'*+-/=?^_`{|}~", rune(character))
}

func validateDomain(domain string) error {
	if len(domain) > 253 {
		return addressPolicyError(AddressErrorTooLarge)
	}
	if domain == "" {
		return addressPolicyError(AddressErrorInvalid)
	}

	labels := strings.Split(domain, ".")
	for _, label := range labels {
		if len(label) == 0 || label[0] == '-' || label[len(label)-1] == '-' {
			return addressPolicyError(AddressErrorInvalid)
		}
		if len(label) > 63 {
			return addressPolicyError(AddressErrorTooLarge)
		}
		for index := 0; index < len(label); index++ {
			character := label[index]
			if character >= 'a' && character <= 'z' || character >= '0' && character <= '9' || character == '-' {
				continue
			}
			return addressPolicyError(AddressErrorInvalid)
		}
	}

	return nil
}
