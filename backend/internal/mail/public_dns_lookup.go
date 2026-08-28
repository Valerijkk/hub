package mail

import (
	"context"
	"errors"
	"net"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	publicDNSMaximumRecords     = 32
	publicDNSMaximumRRSet       = 64
	publicDNSMaximumTXTBytes    = 2048
	publicDNSLookupTimeout      = 2 * time.Second
	publicDNSMaximumConcurrency = 4
	publicDNSMaximumNameBytes   = 253
)

// PublicDNSMX is one public MX answer without provider-specific metadata.
type PublicDNSMX struct {
	Preference uint16
	Host       string
}

// PublicDNSLookup resolves one public DNS record-set through an injected
// provider-neutral boundary.
type PublicDNSLookup interface {
	LookupTXT(context.Context, string) ([]string, error)
	LookupMX(context.Context, string) ([]PublicDNSMX, error)
	LookupCNAME(context.Context, string) (string, error)
}

// NewNetResolverPublicDNSLookup adapts one explicitly injected standard
// library resolver. It never selects a global/default resolver.
func NewNetResolverPublicDNSLookup(resolver *net.Resolver) (PublicDNSLookup, error) {
	if resolver == nil {
		return nil, publicDNSEngineError(EngineErrorRejectedPolicy)
	}
	return netResolverPublicDNSLookup{resolver: resolver}, nil
}

type netResolverPublicDNSLookup struct {
	resolver *net.Resolver
}

func (lookup netResolverPublicDNSLookup) LookupTXT(ctx context.Context, name string) ([]string, error) {
	if lookup.resolver == nil || ctx == nil {
		return nil, publicDNSEngineError(EngineErrorUnavailable)
	}
	values, err := lookup.resolver.LookupTXT(ctx, name)
	return append([]string(nil), values...), err
}

func (lookup netResolverPublicDNSLookup) LookupMX(ctx context.Context, name string) ([]PublicDNSMX, error) {
	if lookup.resolver == nil || ctx == nil {
		return nil, publicDNSEngineError(EngineErrorUnavailable)
	}
	values, err := lookup.resolver.LookupMX(ctx, name)
	if err != nil {
		return nil, err
	}
	converted := make([]PublicDNSMX, len(values))
	for index, value := range values {
		if value == nil {
			return nil, publicDNSEngineError(EngineErrorUnavailable)
		}
		converted[index] = PublicDNSMX{Preference: value.Pref, Host: strings.Clone(value.Host)}
	}
	return converted, nil
}

func (lookup netResolverPublicDNSLookup) LookupCNAME(ctx context.Context, name string) (string, error) {
	if lookup.resolver == nil || ctx == nil {
		return "", publicDNSEngineError(EngineErrorUnavailable)
	}
	value, err := lookup.resolver.LookupCNAME(ctx, name)
	if err != nil {
		return "", err
	}
	queried, validQueried := canonicalPublicDNSTarget(name)
	target, validTarget := canonicalPublicDNSTarget(value)
	if validQueried && validTarget && queried == target {
		return "", nil
	}
	return strings.Clone(value), nil
}

// PublicDNSVerification is the public-DNS verification outcome for a required
// record set. It contains no provider response data.
type PublicDNSVerification struct {
	Verification EngineVerificationState
	DKIM         EngineVerificationState
}

// Valid reports whether both result states are closed public-DNS outcomes.
func (verification PublicDNSVerification) Valid() bool {
	if !validPublicDNSVerificationState(verification.Verification) || !validPublicDNSVerificationState(verification.DKIM) {
		return false
	}
	if verification.Verification == EngineVerificationVerified && verification.DKIM != EngineVerificationVerified {
		return false
	}
	return verification.DKIM != EngineVerificationFailed || verification.Verification == EngineVerificationFailed
}

// Clone returns an independent value copy of the verification outcome.
func (verification PublicDNSVerification) Clone() PublicDNSVerification {
	return verification
}

func validPublicDNSVerificationState(value EngineVerificationState) bool {
	switch value {
	case EngineVerificationPending, EngineVerificationVerified, EngineVerificationFailed:
		return true
	default:
		return false
	}
}

// PublicDNSVerifier verifies one bounded expected DNS projection through a
// caller-provided lookup boundary.
type PublicDNSVerifier struct {
	lookup PublicDNSLookup
}

// NewPublicDNSVerifier constructs a verifier without invoking DNS.
func NewPublicDNSVerifier(lookup PublicDNSLookup) (*PublicDNSVerifier, error) {
	if publicDNSLookupNil(lookup) {
		return nil, publicDNSEngineError(EngineErrorRejectedPolicy)
	}
	return &PublicDNSVerifier{lookup: lookup}, nil
}

// Verify resolves every distinct expected owner/type once. It returns no
// partial result when DNS is unavailable, timed out, malformed, or cancelled.
func (verifier *PublicDNSVerifier) Verify(ctx context.Context, records []EngineDNSRecord) (PublicDNSVerification, error) {
	if verifier == nil || publicDNSLookupNil(verifier.lookup) || ctx == nil {
		return PublicDNSVerification{}, publicDNSEngineError(EngineErrorRejectedPolicy)
	}
	if ctx.Err() != nil {
		return PublicDNSVerification{}, publicDNSEngineError(EngineErrorUnavailable)
	}
	groups, validGroups := publicDNSExpectedGroups(records)
	if !validGroups {
		return PublicDNSVerification{}, publicDNSEngineError(EngineErrorRejectedPolicy)
	}

	workContext, cancel := context.WithCancel(ctx)
	defer cancel()
	results := make(chan publicDNSLookupResult, len(groups))
	semaphore := make(chan struct{}, publicDNSMaximumConcurrency)
	var workers sync.WaitGroup
	for _, group := range groups {
		group := group
		workers.Add(1)
		go func() {
			defer workers.Done()
			result := publicDNSLookupResult{unavailable: true}
			defer func() {
				if recover() != nil {
					result = publicDNSLookupResult{unavailable: true}
				}
				results <- result
			}()
			select {
			case semaphore <- struct{}{}:
			case <-workContext.Done():
				return
			}
			defer func() { <-semaphore }()
			lookupContext, lookupCancel := context.WithTimeout(workContext, publicDNSLookupTimeout)
			result = verifier.lookupGroup(lookupContext, group)
			lookupCancel()
		}()
	}
	go func() {
		workers.Wait()
		close(results)
	}()

	var (
		missing      bool
		mismatch     bool
		dkimMissing  bool
		dkimMismatch bool
		unavailable  bool
	)
	for result := range results {
		if result.unavailable {
			unavailable = true
			cancel()
			continue
		}
		if result.missing {
			missing = true
			if result.dkim {
				dkimMissing = true
			}
		}
		if result.mismatch {
			mismatch = true
			if result.dkim {
				dkimMismatch = true
			}
		}
	}
	if unavailable || ctx.Err() != nil {
		return PublicDNSVerification{}, publicDNSEngineError(EngineErrorUnavailable)
	}
	verification := EngineVerificationVerified
	if mismatch {
		verification = EngineVerificationFailed
	} else if missing {
		verification = EngineVerificationPending
	}
	dkim := EngineVerificationVerified
	if dkimMismatch {
		dkim = EngineVerificationFailed
	} else if dkimMissing {
		dkim = EngineVerificationPending
	}
	result := PublicDNSVerification{Verification: verification, DKIM: dkim}
	if !result.Valid() {
		return PublicDNSVerification{}, publicDNSEngineError(EngineErrorUnavailable)
	}
	return result, nil
}

type publicDNSExpectedGroup struct {
	owner      string
	recordType EngineDNSRecordType
	txt        map[string]struct{}
	mx         map[PublicDNSMX]struct{}
	cname      string
	dkim       bool
}

type publicDNSLookupResult struct {
	missing     bool
	mismatch    bool
	dkim        bool
	unavailable bool
}

func (verifier *PublicDNSVerifier) lookupGroup(ctx context.Context, group publicDNSExpectedGroup) publicDNSLookupResult {
	if ctx == nil || ctx.Err() != nil {
		return publicDNSLookupResult{unavailable: true}
	}
	name := group.owner + "."
	switch group.recordType {
	case EngineDNSRecordTXT:
		values, err := verifier.lookup.LookupTXT(ctx, name)
		if publicDNSLookupUnavailable(err) {
			return publicDNSLookupResult{unavailable: true}
		}
		if len(values) == 0 {
			return publicDNSLookupResult{missing: true, dkim: group.dkim}
		}
		if len(values) > publicDNSMaximumRRSet {
			return publicDNSLookupResult{unavailable: true}
		}
		observed := make(map[string]struct{}, len(values))
		for _, value := range values {
			if !validPublicDNSText(value) {
				return publicDNSLookupResult{unavailable: true}
			}
			observed[value] = struct{}{}
		}
		for value := range group.txt {
			if _, present := observed[value]; !present {
				return publicDNSLookupResult{mismatch: true, dkim: group.dkim}
			}
		}
	case EngineDNSRecordMX:
		values, err := verifier.lookup.LookupMX(ctx, name)
		if publicDNSLookupUnavailable(err) {
			return publicDNSLookupResult{unavailable: true}
		}
		if len(values) == 0 {
			return publicDNSLookupResult{missing: true}
		}
		if len(values) > publicDNSMaximumRRSet {
			return publicDNSLookupResult{unavailable: true}
		}
		observed := make(map[PublicDNSMX]struct{}, len(values))
		for _, value := range values {
			host, validHost := canonicalPublicDNSTarget(value.Host)
			if !validHost {
				return publicDNSLookupResult{unavailable: true}
			}
			observed[PublicDNSMX{Preference: value.Preference, Host: host}] = struct{}{}
		}
		for value := range group.mx {
			if _, present := observed[value]; !present {
				return publicDNSLookupResult{mismatch: true}
			}
		}
	case EngineDNSRecordCNAME:
		value, err := verifier.lookup.LookupCNAME(ctx, name)
		if publicDNSLookupUnavailable(err) {
			return publicDNSLookupResult{unavailable: true}
		}
		if value == "" {
			return publicDNSLookupResult{missing: true}
		}
		target, validTarget := canonicalPublicDNSTarget(value)
		if !validTarget {
			return publicDNSLookupResult{unavailable: true}
		}
		if target == group.owner {
			return publicDNSLookupResult{missing: true}
		}
		if target != group.cname {
			return publicDNSLookupResult{mismatch: true}
		}
	default:
		return publicDNSLookupResult{unavailable: true}
	}
	if ctx.Err() != nil {
		return publicDNSLookupResult{unavailable: true}
	}
	return publicDNSLookupResult{dkim: group.dkim}
}

func publicDNSExpectedGroups(records []EngineDNSRecord) ([]publicDNSExpectedGroup, bool) {
	if len(records) == 0 || len(records) > publicDNSMaximumRecords {
		return nil, false
	}
	validated := make(RequiredDNSRecords, len(records))
	for index, record := range records {
		validated[index] = RequiredDNSRecord{Type: record.Type, Name: record.Name, Value: record.Value, TTL: record.TTL}
	}
	if err := validated.Validate(); err != nil {
		return nil, false
	}
	groups := make(map[string]*publicDNSExpectedGroup, len(validated))
	for _, record := range validated {
		owner, validOwner := canonicalPublicDNSOwner(record.Name)
		if !validOwner {
			return nil, false
		}
		key := string(record.Type) + "\x00" + owner
		group := groups[key]
		if group == nil {
			group = &publicDNSExpectedGroup{owner: owner, recordType: record.Type}
			groups[key] = group
		}
		switch record.Type {
		case EngineDNSRecordTXT:
			if !validPublicDNSText(record.Value) {
				return nil, false
			}
			if group.txt == nil {
				group.txt = make(map[string]struct{})
			}
			group.txt[record.Value] = struct{}{}
			if publicDNSDKIMExpected(owner, record.Value) {
				group.dkim = true
			}
		case EngineDNSRecordMX:
			preference, target, validMX := parsePublicDNSExpectedMX(record.Value)
			if !validMX {
				return nil, false
			}
			if group.mx == nil {
				group.mx = make(map[PublicDNSMX]struct{})
			}
			group.mx[PublicDNSMX{Preference: preference, Host: target}] = struct{}{}
		case EngineDNSRecordCNAME:
			target, validTarget := canonicalPublicDNSTarget(record.Value)
			if !validTarget || group.cname != "" {
				return nil, false
			}
			group.cname = target
		default:
			return nil, false
		}
	}
	ordered := make([]publicDNSExpectedGroup, 0, len(groups))
	hasDKIM := false
	for _, group := range groups {
		if group.recordType == EngineDNSRecordTXT && len(group.txt) == 0 || group.recordType == EngineDNSRecordMX && len(group.mx) == 0 || group.recordType == EngineDNSRecordCNAME && group.cname == "" {
			return nil, false
		}
		hasDKIM = hasDKIM || group.dkim
		ordered = append(ordered, *group)
	}
	if !hasDKIM {
		return nil, false
	}
	sort.Slice(ordered, func(left, right int) bool {
		if ordered[left].owner != ordered[right].owner {
			return ordered[left].owner < ordered[right].owner
		}
		return ordered[left].recordType < ordered[right].recordType
	})
	return ordered, true
}

func parsePublicDNSExpectedMX(value string) (uint16, string, bool) {
	if strings.Count(value, " ") != 1 {
		return 0, "", false
	}
	parts := strings.Split(value, " ")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return 0, "", false
	}
	preference, err := strconv.ParseUint(parts[0], 10, 16)
	if err != nil {
		return 0, "", false
	}
	target, validTarget := canonicalPublicDNSTarget(parts[1])
	if !validTarget {
		return 0, "", false
	}
	return uint16(preference), target, true
}

func publicDNSDKIMExpected(owner, value string) bool {
	labels := strings.Split(owner, ".")
	return len(labels) >= 4 && labels[0] != "" && labels[1] == "_domainkey" && strings.HasPrefix(value, "v=DKIM1;")
}

func canonicalPublicDNSOwner(value string) (string, bool) {
	return canonicalPublicDNSName(value, false, true)
}

func canonicalPublicDNSTarget(value string) (string, bool) {
	return canonicalPublicDNSName(value, true, false)
}

func canonicalPublicDNSName(value string, allowTrailingDot, allowUnderscore bool) (string, bool) {
	if allowTrailingDot && strings.HasSuffix(value, ".") {
		value = strings.TrimSuffix(value, ".")
	}
	if len(value) == 0 || len(value) > publicDNSMaximumNameBytes || strings.HasSuffix(value, ".") {
		return "", false
	}
	labels := strings.Split(value, ".")
	if len(labels) < 2 {
		return "", false
	}
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return "", false
		}
		for _, character := range []byte(label) {
			if character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9' || character == '-' || allowUnderscore && character == '_' {
				continue
			}
			return "", false
		}
	}
	return strings.ToLower(value), true
}

func validPublicDNSText(value string) bool {
	if len(value) == 0 || len(value) > publicDNSMaximumTXTBytes || !utf8.ValidString(value) {
		return false
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return false
		}
	}
	return true
}

func publicDNSLookupUnavailable(err error) bool {
	if err == nil {
		return false
	}
	var dnsError *net.DNSError
	return !errors.As(err, &dnsError) || !dnsError.IsNotFound
}

func publicDNSLookupNil(lookup PublicDNSLookup) bool {
	if lookup == nil {
		return true
	}
	value := reflect.ValueOf(lookup)
	switch value.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Interface, reflect.Func, reflect.Chan:
		return value.IsNil()
	default:
		return false
	}
}

func publicDNSEngineError(kind EngineErrorKind) error {
	return &EngineError{Kind: kind}
}
