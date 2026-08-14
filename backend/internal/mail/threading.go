package mail

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

// ThreadKey is the raw, versioned identity key for a mail thread.
type ThreadKey [32]byte

// ThreadKeyVersion identifies the encoding used to derive a ThreadKey.
type ThreadKeyVersion uint8

const (
	// ThreadKeyVersionV1 is the first accepted thread-key encoding.
	ThreadKeyVersionV1 ThreadKeyVersion = 1
)

// ThreadIdentitySource records whether an existing ancestor or fallback produced an identity.
type ThreadIdentitySource string

const (
	// ThreadIdentitySourceAncestor reuses the exact identity of a resolved ancestor.
	ThreadIdentitySourceAncestor ThreadIdentitySource = "ancestor"
	// ThreadIdentitySourceFallback derives a bounded identity from subject and participants.
	ThreadIdentitySourceFallback ThreadIdentitySource = "fallback"
)

// ThreadAncestor is one validated resolver result eligible for thread selection.
type ThreadAncestor struct {
	MessageID  string
	KeyVersion ThreadKeyVersion
	Key        ThreadKey
}

// ThreadingInput contains the bounded projection data used to derive a thread identity.
type ThreadingInput struct {
	References             []string
	InReplyTo              string
	Subject                string
	From                   EngineAddress
	To                     []EngineAddress
	CC                     []EngineAddress
	Ancestors              []ThreadAncestor
	AncestorLookupComplete bool
}

// ThreadIdentity is the deterministic identity selected for one projected message.
type ThreadIdentity struct {
	KeyVersion           ThreadKeyVersion
	Key                  ThreadKey
	Source               ThreadIdentitySource
	NormalizedSubject    string
	CanonicalParticipants []string
}

// ThreadErrorKind classifies a content-free thread policy failure.
type ThreadErrorKind string

const (
	// ThreadErrorInvalid denotes invalid thread input or an incomplete ancestor lookup.
	ThreadErrorInvalid ThreadErrorKind = "invalid"
	// ThreadErrorTooLarge denotes an input which exceeds the thread bounds.
	ThreadErrorTooLarge ThreadErrorKind = "too_large"
)

var (
	// ErrThreadInvalid is the stable sentinel for invalid thread input.
	ErrThreadInvalid = errors.New("mail thread invalid")
	// ErrThreadTooLarge is the stable sentinel for thread input exceeding bounds.
	ErrThreadTooLarge = errors.New("mail thread exceeds bounds")
)

// ThreadPolicyError is a content-free typed thread policy error.
type ThreadPolicyError struct {
	Kind ThreadErrorKind
}

// Error returns a stable error string without exposing input values.
func (errorValue *ThreadPolicyError) Error() string {
	if errorValue == nil {
		return "mail thread policy error"
	}

	switch errorValue.Kind {
	case ThreadErrorInvalid:
		return ErrThreadInvalid.Error()
	case ThreadErrorTooLarge:
		return ErrThreadTooLarge.Error()
	default:
		return "mail thread policy error"
	}
}

// Unwrap returns the stable sentinel for a recognized thread policy error.
func (errorValue *ThreadPolicyError) Unwrap() error {
	if errorValue == nil {
		return nil
	}

	switch errorValue.Kind {
	case ThreadErrorInvalid:
		return ErrThreadInvalid
	case ThreadErrorTooLarge:
		return ErrThreadTooLarge
	default:
		return nil
	}
}

// DeriveThreadIdentity deterministically selects a resolved ancestor or fallback thread key.
func DeriveThreadIdentity(input ThreadingInput) (ThreadIdentity, error) {
	if !input.AncestorLookupComplete {
		return ThreadIdentity{}, threadPolicyError(ThreadErrorInvalid)
	}
	candidates, err := preflightThreadingInput(input)
	if err != nil {
		return ThreadIdentity{}, err
	}
	if err := validateThreadSubject(input.Subject); err != nil {
		return ThreadIdentity{}, err
	}

	participants, err := normalizeThreadParticipants(input)
	if err != nil {
		return ThreadIdentity{}, err
	}
	ancestors, err := indexThreadAncestors(input.Ancestors)
	if err != nil {
		return ThreadIdentity{}, err
	}
	normalizedSubject, err := normalizeThreadSubject(input.Subject)
	if err != nil {
		return ThreadIdentity{}, err
	}

	if ancestor, found := selectThreadAncestor(candidates.references, candidates.inReplyTo, ancestors); found {
		return ThreadIdentity{
			KeyVersion:            ancestor.KeyVersion,
			Key:                   ancestor.Key,
			Source:                ThreadIdentitySourceAncestor,
			CanonicalParticipants: make([]string, 0),
		}, nil
	}

	key := deriveFallbackThreadKey(normalizedSubject, participants)
	return ThreadIdentity{
		KeyVersion:            ThreadKeyVersionV1,
		Key:                   key,
		Source:                ThreadIdentitySourceFallback,
		NormalizedSubject:     normalizedSubject,
		CanonicalParticipants: participants,
	}, nil
}

type threadAncestorCandidates struct {
	references []string
	inReplyTo  string
}

func preflightThreadingInput(input ThreadingInput) (threadAncestorCandidates, error) {
	if len(input.Subject) > MaxEngineHeaderValueBytes ||
		len(input.References) > MaxEngineReferenceItems ||
		len(input.To) > MaxEngineAddressItems ||
		len(input.CC) > MaxEngineAddressItems ||
		len(input.Ancestors) > 202 {
		return threadAncestorCandidates{}, threadPolicyError(ThreadErrorTooLarge)
	}

	candidates := threadAncestorCandidates{
		references: make([]string, len(input.References)),
	}
	for index, reference := range input.References {
		if len(reference) <= MaxEngineHeaderValueBytes {
			candidates.references[index] = reference
		}
	}
	if len(input.InReplyTo) <= MaxEngineHeaderValueBytes {
		candidates.inReplyTo = input.InReplyTo
	}

	for _, ancestor := range input.Ancestors {
		if len(ancestor.MessageID) > MaxEngineHeaderValueBytes {
			return threadAncestorCandidates{}, threadPolicyError(ThreadErrorTooLarge)
		}
	}
	if err := preflightThreadAddress(input.From); err != nil {
		return threadAncestorCandidates{}, err
	}
	for _, address := range input.To {
		if err := preflightThreadAddress(address); err != nil {
			return threadAncestorCandidates{}, err
		}
	}
	for _, address := range input.CC {
		if err := preflightThreadAddress(address); err != nil {
			return threadAncestorCandidates{}, err
		}
	}

	return candidates, nil
}

func preflightThreadAddress(address EngineAddress) error {
	if len(address.Email) > MaxEngineAddressEmailBytes || len(address.Name) > MaxEngineAddressNameBytes {
		return addressPolicyError(AddressErrorTooLarge)
	}
	return nil
}

func validateThreadSubject(subject string) error {
	if !utf8.ValidString(subject) {
		return threadPolicyError(ThreadErrorInvalid)
	}
	for _, character := range subject {
		if unicode.IsControl(character) {
			return threadPolicyError(ThreadErrorInvalid)
		}
	}
	return nil
}

func normalizeThreadParticipants(input ThreadingInput) ([]string, error) {
	participants := make([]string, 0, 1+len(input.To)+len(input.CC))
	seen := make(map[string]struct{}, cap(participants))
	addresses := make([]EngineAddress, 0, 1+len(input.To)+len(input.CC))
	addresses = append(addresses, input.From)
	addresses = append(addresses, input.To...)
	addresses = append(addresses, input.CC...)

	for _, address := range addresses {
		normalized, err := NormalizeAddress(address)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[normalized.Email]; exists {
			continue
		}
		seen[normalized.Email] = struct{}{}
		participants = append(participants, normalized.Email)
	}
	if len(participants) == 0 {
		return nil, addressPolicyError(AddressErrorInvalid)
	}
	if len(participants) > 401 {
		return nil, threadPolicyError(ThreadErrorTooLarge)
	}

	sort.Strings(participants)
	return participants, nil
}

type threadAncestorIndex struct {
	ancestor  ThreadAncestor
	ambiguous bool
}

func indexThreadAncestors(ancestors []ThreadAncestor) (map[string]threadAncestorIndex, error) {
	indexed := make(map[string]threadAncestorIndex, len(ancestors))
	for _, ancestor := range ancestors {
		if !validThreadMessageID(ancestor.MessageID) || ancestor.KeyVersion != ThreadKeyVersionV1 || threadKeyIsZero(ancestor.Key) {
			return nil, threadPolicyError(ThreadErrorInvalid)
		}

		current, exists := indexed[ancestor.MessageID]
		if !exists {
			indexed[ancestor.MessageID] = threadAncestorIndex{ancestor: ancestor}
			continue
		}
		if current.ancestor.KeyVersion != ancestor.KeyVersion || current.ancestor.Key != ancestor.Key {
			current.ambiguous = true
			indexed[ancestor.MessageID] = current
		}
	}
	return indexed, nil
}

func selectThreadAncestor(references []string, inReplyTo string, ancestors map[string]threadAncestorIndex) (ThreadAncestor, bool) {
	for index := len(references) - 1; index >= 0; index-- {
		if ancestor, found := resolvedThreadAncestor(references[index], ancestors); found {
			return ancestor, true
		}
	}
	return resolvedThreadAncestor(inReplyTo, ancestors)
}

func resolvedThreadAncestor(messageID string, ancestors map[string]threadAncestorIndex) (ThreadAncestor, bool) {
	if !validThreadMessageID(messageID) {
		return ThreadAncestor{}, false
	}
	ancestor, found := ancestors[messageID]
	if !found || ancestor.ambiguous {
		return ThreadAncestor{}, false
	}
	return ancestor.ancestor, true
}

func validThreadMessageID(messageID string) bool {
	if len(messageID) < len("<a@b>") || len(messageID) > MaxEngineHeaderValueBytes || messageID[0] != '<' || messageID[len(messageID)-1] != '>' {
		return false
	}

	contents := messageID[1 : len(messageID)-1]
	atIndex := strings.IndexByte(contents, '@')
	if atIndex <= 0 || atIndex != strings.LastIndexByte(contents, '@') || atIndex == len(contents)-1 {
		return false
	}
	return validThreadDotAtom(contents[:atIndex]) && validThreadDotAtom(contents[atIndex+1:])
}

func validThreadDotAtom(value string) bool {
	if value == "" || value[0] == '.' || value[len(value)-1] == '.' {
		return false
	}

	previousDot := false
	for index := 0; index < len(value); index++ {
		character := value[index]
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

func threadKeyIsZero(key ThreadKey) bool {
	var zero ThreadKey
	return key == zero
}

func normalizeThreadSubject(subject string) (string, error) {
	normalized := norm.NFC.String(subject)
	normalized = cases.Fold().String(normalized)
	normalized = norm.NFC.String(normalized)
	normalized = collapseThreadSubjectSpaces(normalized)

	for removed := 0; removed < 16; removed++ {
		remaining, matched := removeThreadSubjectPrefix(normalized)
		if !matched {
			break
		}
		normalized = remaining
	}
	if len(normalized) > MaxEngineHeaderValueBytes {
		return "", threadPolicyError(ThreadErrorTooLarge)
	}
	return normalized, nil
}

func collapseThreadSubjectSpaces(subject string) string {
	var builder strings.Builder
	builder.Grow(len(subject))
	pendingSpace := false
	for _, character := range subject {
		if unicode.IsSpace(character) {
			if builder.Len() > 0 {
				pendingSpace = true
			}
			continue
		}
		if pendingSpace {
			builder.WriteByte(' ')
			pendingSpace = false
		}
		builder.WriteRune(character)
	}
	return builder.String()
}

func removeThreadSubjectPrefix(subject string) (string, bool) {
	for _, prefix := range []string{"пересылка", "пересл", "ответ", "отв", "fwd", "fw", "re"} {
		if !strings.HasPrefix(subject, prefix) {
			continue
		}

		index := len(prefix)
		if index < len(subject) && subject[index] == ' ' {
			index++
		}
		if index >= len(subject) || subject[index] != ':' {
			continue
		}
		index++
		if index < len(subject) && subject[index] == ' ' {
			index++
		}
		return strings.Trim(subject[index:], " "), true
	}
	return subject, false
}

func deriveFallbackThreadKey(subject string, participants []string) ThreadKey {
	preimage := make([]byte, 0, len("hub.mail.thread.fallback")+2+4+len(subject)+4+len(participants)*4)
	preimage = append(preimage, "hub.mail.thread.fallback"...)
	preimage = append(preimage, 0, byte(ThreadKeyVersionV1))
	preimage = appendThreadString(preimage, subject)
	preimage = appendThreadUint32(preimage, uint32(len(participants)))
	for _, participant := range participants {
		preimage = appendThreadString(preimage, participant)
	}
	return ThreadKey(sha256.Sum256(preimage))
}

func appendThreadString(destination []byte, value string) []byte {
	destination = appendThreadUint32(destination, uint32(len(value)))
	return append(destination, value...)
}

func appendThreadUint32(destination []byte, value uint32) []byte {
	var encoded [4]byte
	binary.BigEndian.PutUint32(encoded[:], value)
	return append(destination, encoded[:]...)
}

func threadPolicyError(kind ThreadErrorKind) error {
	return &ThreadPolicyError{Kind: kind}
}
