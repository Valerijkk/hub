package mail_test

import (
	"context"
	"errors"
	"net"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/valerii/hub/backend/internal/mail"
)

func TestPublicDNSVerifierVerify_ExactValuesAllowExtrasAndUseAbsoluteNames(t *testing.T) {
	records := publicDNSExpectedRecords()
	lookup := publicDNSLookupFixtureFor(records)
	lookup.txt["example.test."] = publicDNSTXTReply{values: []string{"v=spf1 -all", "unrelated"}}
	lookup.mx["example.test."] = publicDNSMXReply{values: []mail.PublicDNSMX{{Preference: 10, Host: "MAIL.EXAMPLE.TEST."}}}
	lookup.cname["alias.example.test."] = publicDNSCNAMEReply{value: "TARGET.EXAMPLE.TEST."}

	verifier, err := mail.NewPublicDNSVerifier(lookup)
	if err != nil {
		t.Fatalf("NewPublicDNSVerifier() error = %v", err)
	}
	result, err := verifier.Verify(context.Background(), records)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	want := mail.PublicDNSVerification{Verification: mail.EngineVerificationVerified, DKIM: mail.EngineVerificationVerified}
	if got := result; got != want {
		t.Fatalf("Verify() = %#v, want %#v", got, want)
	}
	if !result.Valid() || result.Clone() != result {
		t.Fatal("verified PublicDNSVerification is not valid or does not clone by value")
	}
	for _, name := range lookup.queries() {
		if !strings.HasSuffix(name, ".") {
			t.Fatalf("lookup name %q is not absolute", name)
		}
	}
}

func TestPublicDNSVerifierVerify_TXTComparisonIsByteExact(t *testing.T) {
	records := publicDNSExpectedRecords()
	lookup := publicDNSLookupFixtureFor(records)
	lookup.txt["example.test."] = publicDNSTXTReply{values: []string{"V=SPF1 -ALL"}}

	verifier, err := mail.NewPublicDNSVerifier(lookup)
	if err != nil {
		t.Fatalf("NewPublicDNSVerifier() error = %v", err)
	}
	result, err := verifier.Verify(context.Background(), records)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if got, want := result, (mail.PublicDNSVerification{Verification: mail.EngineVerificationFailed, DKIM: mail.EngineVerificationVerified}); got != want {
		t.Fatalf("Verify() = %#v, want %#v", got, want)
	}
}

func TestPublicDNSVerifierVerify_MissingAndPositiveMismatchHaveDifferentStates(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		lookup *publicDNSLookupFixture
		want   mail.PublicDNSVerification
	}{
		{
			name: "empty successful txt rrset is pending",
			lookup: func() *publicDNSLookupFixture {
				lookup := publicDNSLookupFixtureFor(publicDNSExpectedRecords())
				lookup.txt["example.test."] = publicDNSTXTReply{values: []string{}}
				return lookup
			}(),
			want: mail.PublicDNSVerification{Verification: mail.EngineVerificationPending, DKIM: mail.EngineVerificationVerified},
		},
		{
			name: "positive txt mismatch is failed",
			lookup: func() *publicDNSLookupFixture {
				lookup := publicDNSLookupFixtureFor(publicDNSExpectedRecords())
				lookup.txt["example.test."] = publicDNSTXTReply{values: []string{"v=spf1 +all"}}
				return lookup
			}(),
			want: mail.PublicDNSVerification{Verification: mail.EngineVerificationFailed, DKIM: mail.EngineVerificationVerified},
		},
		{
			name: "nxdomain is pending",
			lookup: func() *publicDNSLookupFixture {
				lookup := publicDNSLookupFixtureFor(publicDNSExpectedRecords())
				lookup.txt["_dmarc.example.test."] = publicDNSTXTReply{err: &net.DNSError{IsNotFound: true}}
				return lookup
			}(),
			want: mail.PublicDNSVerification{Verification: mail.EngineVerificationPending, DKIM: mail.EngineVerificationVerified},
		},
		{
			name: "resolver cname self result is missing",
			lookup: func() *publicDNSLookupFixture {
				lookup := publicDNSLookupFixtureFor(publicDNSExpectedRecords())
				lookup.cname["alias.example.test."] = publicDNSCNAMEReply{value: "ALIAS.EXAMPLE.TEST."}
				return lookup
			}(),
			want: mail.PublicDNSVerification{Verification: mail.EngineVerificationPending, DKIM: mail.EngineVerificationVerified},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			verifier, err := mail.NewPublicDNSVerifier(testCase.lookup)
			if err != nil {
				t.Fatalf("NewPublicDNSVerifier() error = %v", err)
			}
			result, err := verifier.Verify(context.Background(), publicDNSExpectedRecords())
			if err != nil {
				t.Fatalf("Verify() error = %v", err)
			}
			if got := result; got != testCase.want {
				t.Fatalf("Verify() = %#v, want %#v", got, testCase.want)
			}
		})
	}
}

func TestPublicDNSVerifierVerify_IsNotFoundAnswersAreDiscarded(t *testing.T) {
	for _, testCase := range []struct {
		name  string
		apply func(*publicDNSLookupFixture)
	}{
		{
			name: "txt",
			apply: func(lookup *publicDNSLookupFixture) {
				lookup.txt["example.test."] = publicDNSTXTReply{
					values: []string{"v=spf1 -all"},
					err:    &net.DNSError{IsNotFound: true},
				}
			},
		},
		{
			name: "mx",
			apply: func(lookup *publicDNSLookupFixture) {
				lookup.mx["example.test."] = publicDNSMXReply{
					values: []mail.PublicDNSMX{{Preference: 10, Host: "mail.example.test."}},
					err:    &net.DNSError{IsNotFound: true},
				}
			},
		},
		{
			name: "cname",
			apply: func(lookup *publicDNSLookupFixture) {
				lookup.cname["alias.example.test."] = publicDNSCNAMEReply{
					value: "target.example.test.",
					err:   &net.DNSError{IsNotFound: true},
				}
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			lookup := publicDNSLookupFixtureFor(publicDNSExpectedRecords())
			testCase.apply(lookup)
			verifier, err := mail.NewPublicDNSVerifier(lookup)
			if err != nil {
				t.Fatalf("NewPublicDNSVerifier() error = %v", err)
			}
			result, err := verifier.Verify(context.Background(), publicDNSExpectedRecords())
			if err != nil {
				t.Fatalf("Verify() error = %v", err)
			}
			want := mail.PublicDNSVerification{Verification: mail.EngineVerificationPending, DKIM: mail.EngineVerificationVerified}
			if result != want || result.Verification == mail.EngineVerificationVerified {
				t.Fatalf("Verify() = %#v, want missing non-verified result %#v", result, want)
			}
		})
	}
}

func TestPublicDNSVerifierVerify_IsNotFoundTimeoutAndTemporaryAreUnavailable(t *testing.T) {
	for _, dnsError := range []*net.DNSError{
		{IsNotFound: true, IsTimeout: true},
		{IsNotFound: true, IsTemporary: true},
	} {
		lookup := publicDNSLookupFixtureFor(publicDNSExpectedRecords())
		lookup.txt["example.test."] = publicDNSTXTReply{values: []string{"v=spf1 -all"}, err: dnsError}
		verifier, err := mail.NewPublicDNSVerifier(lookup)
		if err != nil {
			t.Fatalf("NewPublicDNSVerifier() error = %v", err)
		}
		result, err := verifier.Verify(context.Background(), publicDNSExpectedRecords())
		if !errors.Is(err, mail.ErrEngineUnavailable) || result != (mail.PublicDNSVerification{}) {
			t.Fatalf("Verify() = %#v, %v; want zero unavailable result", result, err)
		}
	}
}

func TestPublicDNSVerificationValid_RejectsFailedDKIMWithoutFailedVerification(t *testing.T) {
	for _, verification := range []mail.PublicDNSVerification{
		{Verification: mail.EngineVerificationPending, DKIM: mail.EngineVerificationFailed},
		{Verification: mail.EngineVerificationVerified, DKIM: mail.EngineVerificationFailed},
	} {
		if verification.Valid() {
			t.Fatalf("PublicDNSVerification %#v Valid() = true, want false", verification)
		}
	}
	for _, verification := range []mail.PublicDNSVerification{
		{Verification: mail.EngineVerificationFailed, DKIM: mail.EngineVerificationPending},
		{Verification: mail.EngineVerificationFailed, DKIM: mail.EngineVerificationVerified},
	} {
		if !verification.Valid() {
			t.Fatalf("PublicDNSVerification %#v Valid() = false, want true", verification)
		}
	}
}

func TestPublicDNSVerifierVerify_DKIMStatusTracksOnlyDKIMTXT(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		lookup *publicDNSLookupFixture
		want   mail.PublicDNSVerification
	}{
		{
			name: "missing dkim is pending",
			lookup: func() *publicDNSLookupFixture {
				lookup := publicDNSLookupFixtureFor(publicDNSExpectedRecords())
				lookup.txt["selector._domainkey.example.test."] = publicDNSTXTReply{}
				return lookup
			}(),
			want: mail.PublicDNSVerification{Verification: mail.EngineVerificationPending, DKIM: mail.EngineVerificationPending},
		},
		{
			name: "positive dkim mismatch is failed",
			lookup: func() *publicDNSLookupFixture {
				lookup := publicDNSLookupFixtureFor(publicDNSExpectedRecords())
				lookup.txt["selector._domainkey.example.test."] = publicDNSTXTReply{values: []string{"v=DKIM1; p=other"}}
				return lookup
			}(),
			want: mail.PublicDNSVerification{Verification: mail.EngineVerificationFailed, DKIM: mail.EngineVerificationFailed},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			verifier, err := mail.NewPublicDNSVerifier(testCase.lookup)
			if err != nil {
				t.Fatalf("NewPublicDNSVerifier() error = %v", err)
			}
			result, err := verifier.Verify(context.Background(), publicDNSExpectedRecords())
			if err != nil {
				t.Fatalf("Verify() error = %v", err)
			}
			if got := result; got != testCase.want {
				t.Fatalf("Verify() = %#v, want %#v", got, testCase.want)
			}
		})
	}
}

func TestPublicDNSVerifierVerify_MapsTimeoutCancellationAndMalformedResponsesToUnavailable(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		ctx    context.Context
		lookup *publicDNSLookupFixture
	}{
		{
			name: "lookup timeout",
			ctx:  context.Background(),
			lookup: &publicDNSLookupFixture{txtFn: func(ctx context.Context, _ string) ([]string, error) {
				<-ctx.Done()
				return nil, ctx.Err()
			}},
		},
		{
			name: "parent cancellation",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			lookup: publicDNSLookupFixtureFor(publicDNSExpectedRecords()),
		},
		{
			name: "malformed observed text",
			ctx:  context.Background(),
			lookup: func() *publicDNSLookupFixture {
				lookup := publicDNSLookupFixtureFor(publicDNSExpectedRecords())
				lookup.txt["example.test."] = publicDNSTXTReply{values: []string{""}}
				return lookup
			}(),
		},
		{
			name: "excess response values",
			ctx:  context.Background(),
			lookup: func() *publicDNSLookupFixture {
				lookup := publicDNSLookupFixtureFor(publicDNSExpectedRecords())
				lookup.txt["example.test."] = publicDNSTXTReply{values: make([]string, 65)}
				return lookup
			}(),
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			verifier, err := mail.NewPublicDNSVerifier(testCase.lookup)
			if err != nil {
				t.Fatalf("NewPublicDNSVerifier() error = %v", err)
			}
			_, err = verifier.Verify(testCase.ctx, publicDNSExpectedRecords())
			if !errors.Is(err, mail.ErrEngineUnavailable) {
				t.Fatalf("Verify() error = %v, want ErrEngineUnavailable", err)
			}
		})
	}
}

func TestPublicDNSVerifierVerify_PanicInLookupFailsClosed(t *testing.T) {
	lookup := publicDNSLookupFixtureFor(publicDNSExpectedRecords())
	verifier, err := mail.NewPublicDNSVerifier(publicDNSPanicLookup{fallback: lookup})
	if err != nil {
		t.Fatalf("NewPublicDNSVerifier() error = %v", err)
	}
	result, err := verifier.Verify(context.Background(), publicDNSExpectedRecords())
	if !errors.Is(err, mail.ErrEngineUnavailable) {
		t.Fatalf("Verify() error = %v, want ErrEngineUnavailable", err)
	}
	if result != (mail.PublicDNSVerification{}) {
		t.Fatalf("Verify() result = %#v, want zero result", result)
	}
}

func TestPublicDNSVerifierVerify_UncooperativeLookupReturnsBoundedUnavailable(t *testing.T) {
	lookup := publicDNSLookupFixtureFor(publicDNSExpectedRecords())
	lookup.ignoreContextBlock = make(chan struct{})
	defer close(lookup.ignoreContextBlock)
	verifier, err := mail.NewPublicDNSVerifier(lookup)
	if err != nil {
		t.Fatalf("NewPublicDNSVerifier() error = %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	type verificationOutcome struct {
		result mail.PublicDNSVerification
		err    error
	}
	outcomes := make(chan verificationOutcome, 1)
	go func() {
		result, verifyErr := verifier.Verify(ctx, publicDNSExpectedRecords())
		outcomes <- verificationOutcome{result: result, err: verifyErr}
	}()
	select {
	case outcome := <-outcomes:
		if !errors.Is(outcome.err, mail.ErrEngineUnavailable) || outcome.result != (mail.PublicDNSVerification{}) {
			t.Fatalf("Verify() = %#v, %v; want zero unavailable result", outcome.result, outcome.err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Verify() did not return after its parent deadline")
	}
	if started := lookup.started.Load(); started > 4 {
		t.Fatalf("stuck lookups = %d, want <= 4", started)
	}
}

func TestPublicDNSVerifierVerify_UncooperativeLookupCannotAccumulateStuckCalls(t *testing.T) {
	lookup := publicDNSLookupFixtureFor(publicDNSExpectedRecords())
	lookup.ignoreContextBlock = make(chan struct{})
	defer close(lookup.ignoreContextBlock)
	verifier, err := mail.NewPublicDNSVerifier(lookup)
	if err != nil {
		t.Fatalf("NewPublicDNSVerifier() error = %v", err)
	}

	firstContext, firstCancel := context.WithCancel(context.Background())
	firstDone := make(chan error, 1)
	go func() {
		_, verifyErr := verifier.Verify(firstContext, publicDNSExpectedRecords())
		firstDone <- verifyErr
	}()
	deadline := time.After(500 * time.Millisecond)
	for lookup.started.Load() < 4 {
		select {
		case <-deadline:
			t.Fatalf("stuck lookups = %d, want 4", lookup.started.Load())
		case <-time.After(time.Millisecond):
		}
	}
	firstCancel()
	if err := <-firstDone; !errors.Is(err, mail.ErrEngineUnavailable) {
		t.Fatalf("first Verify() error = %v, want unavailable", err)
	}

	var workers sync.WaitGroup
	errorsByCall := make(chan error, 8)
	for range 8 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()
			_, verifyErr := verifier.Verify(ctx, publicDNSExpectedRecords())
			errorsByCall <- verifyErr
		}()
	}
	workers.Wait()
	close(errorsByCall)
	for verifyErr := range errorsByCall {
		if !errors.Is(verifyErr, mail.ErrEngineUnavailable) {
			t.Fatalf("subsequent Verify() error = %v, want unavailable", verifyErr)
		}
	}
	if started := lookup.started.Load(); started != 4 || lookup.maximum.Load() > 4 {
		t.Fatalf("stuck lookup accounting = started %d maximum %d, want exactly four", started, lookup.maximum.Load())
	}
}

func TestPublicDNSVerifierVerify_BoundsConcurrencyAndDoesNotMutateInput(t *testing.T) {
	records := append(publicDNSExpectedRecords(),
		mail.EngineDNSRecord{Type: mail.EngineDNSRecordTXT, Name: "one.example.test", Value: "one", TTL: 300},
		mail.EngineDNSRecord{Type: mail.EngineDNSRecordTXT, Name: "two.example.test", Value: "two", TTL: 300},
		mail.EngineDNSRecord{Type: mail.EngineDNSRecordTXT, Name: "three.example.test", Value: "three", TTL: 300},
	)
	before := append([]mail.EngineDNSRecord(nil), records...)
	lookup := publicDNSLookupFixtureFor(records)
	lookup.block = make(chan struct{})
	verifier, err := mail.NewPublicDNSVerifier(lookup)
	if err != nil {
		t.Fatalf("NewPublicDNSVerifier() error = %v", err)
	}

	done := make(chan error, 1)
	go func() {
		_, verifyErr := verifier.Verify(context.Background(), records)
		done <- verifyErr
	}()
	deadline := time.After(500 * time.Millisecond)
	for lookup.started.Load() < 4 {
		select {
		case <-deadline:
			t.Fatalf("concurrent lookups = %d, want at least 4", lookup.started.Load())
		case <-time.After(time.Millisecond):
		}
	}
	if got := lookup.maximum.Load(); got > 4 {
		t.Fatalf("maximum concurrent lookups = %d, want <= 4", got)
	}
	close(lookup.block)
	if err := <-done; err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if !reflect.DeepEqual(records, before) {
		t.Fatalf("Verify() mutated input: got %#v, want %#v", records, before)
	}
}

func TestPublicDNSVerifierRejectsInvalidDependenciesAndExpectedRecords(t *testing.T) {
	if _, err := mail.NewPublicDNSVerifier(nil); !errors.Is(err, mail.ErrEngineRejectedPolicy) {
		t.Fatalf("NewPublicDNSVerifier(nil) error = %v, want ErrEngineRejectedPolicy", err)
	}
	if _, err := mail.NewNetResolverPublicDNSLookup(nil); !errors.Is(err, mail.ErrEngineRejectedPolicy) {
		t.Fatalf("NewNetResolverPublicDNSLookup(nil) error = %v, want ErrEngineRejectedPolicy", err)
	}
	if _, err := mail.NewNetResolverPublicDNSLookup(&net.Resolver{}); err != nil {
		t.Fatalf("NewNetResolverPublicDNSLookup() error = %v", err)
	}
	verifier, err := mail.NewPublicDNSVerifier(publicDNSLookupFixtureFor(publicDNSExpectedRecords()))
	if err != nil {
		t.Fatalf("NewPublicDNSVerifier() error = %v", err)
	}
	if _, err := verifier.Verify(context.Background(), nil); !errors.Is(err, mail.ErrEngineRejectedPolicy) {
		t.Fatalf("Verify(nil) error = %v, want ErrEngineRejectedPolicy", err)
	}
	withoutDKIM := publicDNSExpectedRecords()[:3]
	if _, err := verifier.Verify(context.Background(), withoutDKIM); !errors.Is(err, mail.ErrEngineRejectedPolicy) {
		t.Fatalf("Verify(without DKIM) error = %v, want ErrEngineRejectedPolicy", err)
	}
}

type publicDNSPanicLookup struct {
	fallback *publicDNSLookupFixture
}

func (publicDNSPanicLookup) LookupTXT(context.Context, string) ([]string, error) {
	panic("dns lookup panic")
}

func (lookup publicDNSPanicLookup) LookupMX(ctx context.Context, name string) ([]mail.PublicDNSMX, error) {
	return lookup.fallback.LookupMX(ctx, name)
}

func (lookup publicDNSPanicLookup) LookupCNAME(ctx context.Context, name string) (string, error) {
	return lookup.fallback.LookupCNAME(ctx, name)
}

type publicDNSTXTReply struct {
	values []string
	err    error
}

type publicDNSMXReply struct {
	values []mail.PublicDNSMX
	err    error
}

type publicDNSCNAMEReply struct {
	value string
	err   error
}

type publicDNSLookupFixture struct {
	mu                 sync.Mutex
	txt                map[string]publicDNSTXTReply
	mx                 map[string]publicDNSMXReply
	cname              map[string]publicDNSCNAMEReply
	names              []string
	txtFn              func(context.Context, string) ([]string, error)
	block              chan struct{}
	ignoreContextBlock chan struct{}
	started            atomic.Int32
	active             atomic.Int32
	maximum            atomic.Int32
}

func (fixture *publicDNSLookupFixture) LookupTXT(ctx context.Context, name string) ([]string, error) {
	if fixture.txtFn != nil {
		return fixture.txtFn(ctx, name)
	}
	fixture.enter(ctx, name)
	fixture.mu.Lock()
	reply := fixture.txt[name]
	fixture.mu.Unlock()
	return append([]string(nil), reply.values...), reply.err
}

func (fixture *publicDNSLookupFixture) LookupMX(ctx context.Context, name string) ([]mail.PublicDNSMX, error) {
	fixture.enter(ctx, name)
	fixture.mu.Lock()
	reply := fixture.mx[name]
	fixture.mu.Unlock()
	return append([]mail.PublicDNSMX(nil), reply.values...), reply.err
}

func (fixture *publicDNSLookupFixture) LookupCNAME(ctx context.Context, name string) (string, error) {
	fixture.enter(ctx, name)
	fixture.mu.Lock()
	reply := fixture.cname[name]
	fixture.mu.Unlock()
	return reply.value, reply.err
}

func (fixture *publicDNSLookupFixture) enter(ctx context.Context, name string) {
	fixture.mu.Lock()
	fixture.names = append(fixture.names, name)
	fixture.mu.Unlock()
	fixture.started.Add(1)
	active := fixture.active.Add(1)
	for {
		maximum := fixture.maximum.Load()
		if active <= maximum || fixture.maximum.CompareAndSwap(maximum, active) {
			break
		}
	}
	defer fixture.active.Add(-1)
	if fixture.ignoreContextBlock != nil {
		<-fixture.ignoreContextBlock
		return
	}
	if fixture.block == nil {
		return
	}
	select {
	case <-fixture.block:
	case <-ctx.Done():
	}
}

func (fixture *publicDNSLookupFixture) queries() []string {
	fixture.mu.Lock()
	defer fixture.mu.Unlock()
	return append([]string(nil), fixture.names...)
}

func publicDNSExpectedRecords() []mail.EngineDNSRecord {
	return []mail.EngineDNSRecord{
		{Type: mail.EngineDNSRecordMX, Name: "example.test", Value: "10 mail.example.test", TTL: 300},
		{Type: mail.EngineDNSRecordTXT, Name: "example.test", Value: "v=spf1 -all", TTL: 300},
		{Type: mail.EngineDNSRecordTXT, Name: "_dmarc.example.test", Value: "v=DMARC1; p=none", TTL: 300},
		{Type: mail.EngineDNSRecordTXT, Name: "selector._domainkey.example.test", Value: "v=DKIM1; p=key", TTL: 300},
		{Type: mail.EngineDNSRecordCNAME, Name: "alias.example.test", Value: "target.example.test", TTL: 300},
	}
}

func publicDNSLookupFixtureFor(records []mail.EngineDNSRecord) *publicDNSLookupFixture {
	fixture := &publicDNSLookupFixture{
		txt:   make(map[string]publicDNSTXTReply),
		mx:    make(map[string]publicDNSMXReply),
		cname: make(map[string]publicDNSCNAMEReply),
	}
	for _, record := range records {
		name := record.Name + "."
		switch record.Type {
		case mail.EngineDNSRecordTXT:
			reply := fixture.txt[name]
			reply.values = append(reply.values, record.Value)
			fixture.txt[name] = reply
		case mail.EngineDNSRecordMX:
			fixture.mx[name] = publicDNSMXReply{values: []mail.PublicDNSMX{{Preference: 10, Host: "mail.example.test."}}}
		case mail.EngineDNSRecordCNAME:
			fixture.cname[name] = publicDNSCNAMEReply{value: record.Value + "."}
		}
	}
	return fixture
}
