package mail_test

import (
	"context"
	"errors"
	"net"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/valerii/hub/backend/internal/mail"
)

func TestPublicDNSDomainProvisionerGetDomainProjectsVerifiedPublicDNS(t *testing.T) {
	records := publicDNSDomainProvisionerRecords()
	upstream := &publicDNSDomainProvisionerUpstream{getState: publicDNSDomainProvisionerDNSRequired(records)}
	lookup := publicDNSDomainLookupFor(records)
	provisioner := newPublicDNSDomainProvisioner(t, upstream, lookup)

	state, err := provisioner.GetDomain(context.Background(), "example.test")
	if err != nil {
		t.Fatalf("GetDomain() error = %v", err)
	}
	want := publicDNSDomainProvisionerDNSRequired(records)
	want.State = mail.EngineDomainVerified
	want.Verification = mail.EngineVerificationVerified
	want.DKIM = mail.EngineVerificationVerified
	if !reflect.DeepEqual(state, want) {
		t.Fatalf("GetDomain() = %#v, want %#v", state, want)
	}
	if upstream.getCalls.Load() != 1 || lookup.calls.Load() != int32(len(records)) {
		t.Fatalf("GetDomain() calls = upstream %d DNS %d, want 1 and %d", upstream.getCalls.Load(), lookup.calls.Load(), len(records))
	}
}

func TestPublicDNSDomainProvisionerEnsureDomainProjectsClosedDNSOutcomes(t *testing.T) {
	records := publicDNSDomainProvisionerRecords()
	for _, testCase := range []struct {
		name   string
		adjust func(*publicDNSDomainLookup)
		state  mail.EngineDomainState
	}{
		{
			name:   "verified",
			adjust: func(*publicDNSDomainLookup) {},
			state: func() mail.EngineDomainState {
				state := publicDNSDomainProvisionerDNSRequired(records)
				state.State = mail.EngineDomainVerified
				state.Verification = mail.EngineVerificationVerified
				state.DKIM = mail.EngineVerificationVerified
				return state
			}(),
		},
		{
			name: "missing is pending",
			adjust: func(lookup *publicDNSDomainLookup) {
				lookup.txt["_dmarc.example.test."] = publicDNSDomainTXTReply{}
			},
			state: func() mail.EngineDomainState {
				state := publicDNSDomainProvisionerDNSRequired(records)
				state.Verification = mail.EngineVerificationPending
				state.DKIM = mail.EngineVerificationVerified
				return state
			}(),
		},
		{
			name: "positive mismatch is failed",
			adjust: func(lookup *publicDNSDomainLookup) {
				lookup.txt["example.test."] = publicDNSDomainTXTReply{values: []string{"v=spf1 +all"}}
			},
			state: func() mail.EngineDomainState {
				state := publicDNSDomainProvisionerDNSRequired(records)
				state.Verification = mail.EngineVerificationFailed
				state.DKIM = mail.EngineVerificationVerified
				return state
			}(),
		},
		{
			name: "dkim mismatch is failed",
			adjust: func(lookup *publicDNSDomainLookup) {
				lookup.txt["selector._domainkey.example.test."] = publicDNSDomainTXTReply{values: []string{"v=DKIM1; p=other"}}
			},
			state: func() mail.EngineDomainState {
				state := publicDNSDomainProvisionerDNSRequired(records)
				state.Verification = mail.EngineVerificationFailed
				state.DKIM = mail.EngineVerificationFailed
				return state
			}(),
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			upstream := &publicDNSDomainProvisionerUpstream{ensureState: publicDNSDomainProvisionerDNSRequired(records)}
			lookup := publicDNSDomainLookupFor(records)
			testCase.adjust(lookup)
			provisioner := newPublicDNSDomainProvisioner(t, upstream, lookup)

			state, err := provisioner.EnsureDomain(context.Background(), publicDNSDomainProvisionerSpec())
			if err != nil {
				t.Fatalf("EnsureDomain() error = %v", err)
			}
			if !reflect.DeepEqual(state, testCase.state) {
				t.Fatalf("EnsureDomain() = %#v, want %#v", state, testCase.state)
			}
			if state.State == mail.EngineDomainDNSRequired && state.Verification == mail.EngineVerificationVerified {
				t.Fatalf("DNS-required state trusted global verification: %#v", state)
			}
			if upstream.ensureCalls.Load() != 1 || lookup.calls.Load() != int32(len(records)) {
				t.Fatalf("EnsureDomain() calls = upstream %d DNS %d, want 1 and %d", upstream.ensureCalls.Load(), lookup.calls.Load(), len(records))
			}
		})
	}
}

func TestPublicDNSDomainProvisionerGetDomainReturnsExactAbsentWithoutDNS(t *testing.T) {
	upstream := &publicDNSDomainProvisionerUpstream{getState: mail.EngineDomainState{
		State:        mail.EngineDomainAbsent,
		DNSRecords:   []mail.EngineDNSRecord{},
		Verification: mail.EngineVerificationUnknown,
		DKIM:         mail.EngineVerificationUnknown,
	}}
	lookup := publicDNSDomainLookupFor(publicDNSDomainProvisionerRecords())
	provisioner := newPublicDNSDomainProvisioner(t, upstream, lookup)

	state, err := provisioner.GetDomain(context.Background(), "example.test")
	if err != nil {
		t.Fatalf("GetDomain() error = %v", err)
	}
	if !reflect.DeepEqual(state, upstream.getState) || state.DNSRecords == nil {
		t.Fatalf("GetDomain() = %#v, want independent exact absence %#v", state, upstream.getState)
	}
	if upstream.getCalls.Load() != 1 || lookup.calls.Load() != 0 {
		t.Fatalf("absent GetDomain() calls = upstream %d DNS %d, want 1 and 0", upstream.getCalls.Load(), lookup.calls.Load())
	}
}

func TestPublicDNSDomainProvisionerGetDomainRejectsMalformedAndPreverifiedStates(t *testing.T) {
	records := publicDNSDomainProvisionerRecords()
	for _, testCase := range []struct {
		name  string
		state mail.EngineDomainState
	}{
		{name: "nil records", state: mail.EngineDomainState{DomainID: "domain-opaque", State: mail.EngineDomainDNSRequired, Applied: true, Verification: mail.EngineVerificationPending, DKIM: mail.EngineVerificationPending}},
		{name: "empty records", state: mail.EngineDomainState{DomainID: "domain-opaque", State: mail.EngineDomainDNSRequired, Applied: true, DNSRecords: []mail.EngineDNSRecord{}, Verification: mail.EngineVerificationPending, DKIM: mail.EngineVerificationPending}},
		{name: "uuid domain id", state: func() mail.EngineDomainState {
			state := publicDNSDomainProvisionerDNSRequired(records)
			state.DomainID = "00000000-0000-0000-0000-000000000001"
			return state
		}()},
		{name: "not applied", state: func() mail.EngineDomainState {
			state := publicDNSDomainProvisionerDNSRequired(records)
			state.Applied = false
			return state
		}()},
		{name: "preverified verification", state: func() mail.EngineDomainState {
			state := publicDNSDomainProvisionerDNSRequired(records)
			state.Verification = mail.EngineVerificationVerified
			return state
		}()},
		{name: "preverified dkim", state: func() mail.EngineDomainState {
			state := publicDNSDomainProvisionerDNSRequired(records)
			state.DKIM = mail.EngineVerificationVerified
			return state
		}()},
		{name: "provider verified", state: func() mail.EngineDomainState {
			state := publicDNSDomainProvisionerDNSRequired(records)
			state.State = mail.EngineDomainVerified
			return state
		}()},
		{name: "absent nil slice", state: mail.EngineDomainState{State: mail.EngineDomainAbsent, Verification: mail.EngineVerificationUnknown, DKIM: mail.EngineVerificationUnknown}},
		{name: "absent stray field", state: mail.EngineDomainState{DomainID: "domain-opaque", State: mail.EngineDomainAbsent, DNSRecords: []mail.EngineDNSRecord{}, Verification: mail.EngineVerificationUnknown, DKIM: mail.EngineVerificationUnknown}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			upstream := &publicDNSDomainProvisionerUpstream{getState: testCase.state}
			lookup := publicDNSDomainLookupFor(records)
			provisioner := newPublicDNSDomainProvisioner(t, upstream, lookup)

			state, err := provisioner.GetDomain(context.Background(), "example.test")
			if !errors.Is(err, mail.ErrEngineUnavailable) || !reflect.DeepEqual(state, mail.EngineDomainState{}) {
				t.Fatalf("GetDomain() = %#v, %v; want zero unavailable result", state, err)
			}
			if lookup.calls.Load() != 0 {
				t.Fatalf("malformed state caused %d DNS calls", lookup.calls.Load())
			}
		})
	}
}

func TestPublicDNSDomainProvisionerEnsureDomainMapsMalformedStateToIndeterminate(t *testing.T) {
	records := publicDNSDomainProvisionerRecords()
	state := publicDNSDomainProvisionerDNSRequired(records)
	state.Verification = mail.EngineVerificationVerified
	upstream := &publicDNSDomainProvisionerUpstream{ensureState: state}
	lookup := publicDNSDomainLookupFor(records)
	provisioner := newPublicDNSDomainProvisioner(t, upstream, lookup)

	result, err := provisioner.EnsureDomain(context.Background(), publicDNSDomainProvisionerSpec())
	if !errors.Is(err, mail.ErrEngineIndeterminate) || !reflect.DeepEqual(result, mail.EngineDomainState{}) {
		t.Fatalf("EnsureDomain() = %#v, %v; want zero indeterminate result", result, err)
	}
	if lookup.calls.Load() != 0 {
		t.Fatalf("malformed ensure state caused %d DNS calls", lookup.calls.Load())
	}
}

func TestPublicDNSDomainProvisionerMapsDNSNXDOMAINAndTimeoutFailClosed(t *testing.T) {
	records := publicDNSDomainProvisionerRecords()
	t.Run("nxdomain is pending", func(t *testing.T) {
		upstream := &publicDNSDomainProvisionerUpstream{getState: publicDNSDomainProvisionerDNSRequired(records)}
		lookup := publicDNSDomainLookupFor(records)
		lookup.txt["_dmarc.example.test."] = publicDNSDomainTXTReply{err: &net.DNSError{IsNotFound: true}}
		provisioner := newPublicDNSDomainProvisioner(t, upstream, lookup)

		state, err := provisioner.GetDomain(context.Background(), "example.test")
		if err != nil || state.State != mail.EngineDomainDNSRequired || state.Verification != mail.EngineVerificationPending || state.DKIM != mail.EngineVerificationVerified {
			t.Fatalf("GetDomain() = %#v, %v; want pending DNS-required result", state, err)
		}
	})
	t.Run("timeout is unavailable", func(t *testing.T) {
		upstream := &publicDNSDomainProvisionerUpstream{getState: publicDNSDomainProvisionerDNSRequired(records)}
		lookup := publicDNSDomainLookupFor(records)
		lookup.txt["_dmarc.example.test."] = publicDNSDomainTXTReply{err: &net.DNSError{IsTimeout: true}}
		provisioner := newPublicDNSDomainProvisioner(t, upstream, lookup)

		state, err := provisioner.GetDomain(context.Background(), "example.test")
		if !errors.Is(err, mail.ErrEngineUnavailable) || !reflect.DeepEqual(state, mail.EngineDomainState{}) {
			t.Fatalf("GetDomain() = %#v, %v; want zero unavailable result", state, err)
		}
	})
}

func TestPublicDNSDomainProvisionerSanitizesUpstreamErrors(t *testing.T) {
	records := publicDNSDomainProvisionerRecords()
	t.Run("get known engine kind is cloned", func(t *testing.T) {
		upstream := &publicDNSDomainProvisionerUpstream{getErr: &mail.EngineError{Kind: mail.EngineErrorRejectedPolicy}}
		provisioner := newPublicDNSDomainProvisioner(t, upstream, publicDNSDomainLookupFor(records))

		_, err := provisioner.GetDomain(context.Background(), "example.test")
		if !errors.Is(err, mail.ErrEngineRejectedPolicy) || err.Error() != mail.ErrEngineRejectedPolicy.Error() {
			t.Fatalf("GetDomain() error = %v, want cloned rejected-policy error", err)
		}
	})
	t.Run("get unknown error is redacted", func(t *testing.T) {
		upstream := &publicDNSDomainProvisionerUpstream{getErr: errors.New("provider-secret@example.test")}
		provisioner := newPublicDNSDomainProvisioner(t, upstream, publicDNSDomainLookupFor(records))

		state, err := provisioner.GetDomain(context.Background(), "example.test")
		if !errors.Is(err, mail.ErrEngineUnavailable) || !reflect.DeepEqual(state, mail.EngineDomainState{}) || strings.Contains(err.Error(), "provider-secret") {
			t.Fatalf("GetDomain() = %#v, %v; want redacted unavailable", state, err)
		}
	})
	t.Run("ensure errors are always indeterminate after dispatch", func(t *testing.T) {
		upstream := &publicDNSDomainProvisionerUpstream{ensureErr: errors.New("provider-secret@example.test")}
		provisioner := newPublicDNSDomainProvisioner(t, upstream, publicDNSDomainLookupFor(records))

		state, err := provisioner.EnsureDomain(context.Background(), publicDNSDomainProvisionerSpec())
		if !errors.Is(err, mail.ErrEngineIndeterminate) || !reflect.DeepEqual(state, mail.EngineDomainState{}) || strings.Contains(err.Error(), "provider-secret") {
			t.Fatalf("EnsureDomain() = %#v, %v; want redacted indeterminate", state, err)
		}
	})
}

func TestPublicDNSDomainProvisionerHandlesContextLossBeforeAndAfterDispatch(t *testing.T) {
	records := publicDNSDomainProvisionerRecords()
	t.Run("get before dispatch", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		upstream := &publicDNSDomainProvisionerUpstream{getState: publicDNSDomainProvisionerDNSRequired(records)}
		provisioner := newPublicDNSDomainProvisioner(t, upstream, publicDNSDomainLookupFor(records))

		state, err := provisioner.GetDomain(ctx, "example.test")
		if !errors.Is(err, mail.ErrEngineUnavailable) || !reflect.DeepEqual(state, mail.EngineDomainState{}) || upstream.getCalls.Load() != 0 {
			t.Fatalf("GetDomain() = %#v, %v; want pre-dispatch unavailable without I/O", state, err)
		}
	})
	t.Run("ensure before dispatch", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		upstream := &publicDNSDomainProvisionerUpstream{ensureState: publicDNSDomainProvisionerDNSRequired(records)}
		provisioner := newPublicDNSDomainProvisioner(t, upstream, publicDNSDomainLookupFor(records))

		state, err := provisioner.EnsureDomain(ctx, publicDNSDomainProvisionerSpec())
		if !errors.Is(err, mail.ErrEngineUnavailable) || !reflect.DeepEqual(state, mail.EngineDomainState{}) || upstream.ensureCalls.Load() != 0 {
			t.Fatalf("EnsureDomain() = %#v, %v; want pre-dispatch unavailable without I/O", state, err)
		}
	})
	t.Run("get after dispatch", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		upstream := &publicDNSDomainProvisionerUpstream{getFn: func(context.Context, string) (mail.EngineDomainState, error) {
			cancel()
			return publicDNSDomainProvisionerDNSRequired(records), nil
		}}
		provisioner := newPublicDNSDomainProvisioner(t, upstream, publicDNSDomainLookupFor(records))

		state, err := provisioner.GetDomain(ctx, "example.test")
		if !errors.Is(err, mail.ErrEngineUnavailable) || !reflect.DeepEqual(state, mail.EngineDomainState{}) {
			t.Fatalf("GetDomain() = %#v, %v; want post-dispatch unavailable", state, err)
		}
	})
	t.Run("ensure after dispatch", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		upstream := &publicDNSDomainProvisionerUpstream{ensureFn: func(context.Context, mail.DomainSpec) (mail.EngineDomainState, error) {
			cancel()
			return publicDNSDomainProvisionerDNSRequired(records), nil
		}}
		provisioner := newPublicDNSDomainProvisioner(t, upstream, publicDNSDomainLookupFor(records))

		state, err := provisioner.EnsureDomain(ctx, publicDNSDomainProvisionerSpec())
		if !errors.Is(err, mail.ErrEngineIndeterminate) || !reflect.DeepEqual(state, mail.EngineDomainState{}) {
			t.Fatalf("EnsureDomain() = %#v, %v; want post-dispatch indeterminate", state, err)
		}
	})
}

func TestPublicDNSDomainProvisionerRejectsInvalidInputBeforeUpstream(t *testing.T) {
	records := publicDNSDomainProvisionerRecords()
	upstream := &publicDNSDomainProvisionerUpstream{
		getState:    publicDNSDomainProvisionerDNSRequired(records),
		ensureState: publicDNSDomainProvisionerDNSRequired(records),
	}
	provisioner := newPublicDNSDomainProvisioner(t, upstream, publicDNSDomainLookupFor(records))

	state, err := provisioner.GetDomain(context.Background(), " Example.TEST ")
	if !errors.Is(err, mail.ErrEngineRejectedPolicy) || !reflect.DeepEqual(state, mail.EngineDomainState{}) || upstream.getCalls.Load() != 0 {
		t.Fatalf("GetDomain() = %#v, %v; want pre-dispatch rejected-policy result", state, err)
	}
	spec := publicDNSDomainProvisionerSpec()
	spec.Enabled = false
	state, err = provisioner.EnsureDomain(context.Background(), spec)
	if !errors.Is(err, mail.ErrEngineRejectedPolicy) || !reflect.DeepEqual(state, mail.EngineDomainState{}) || upstream.ensureCalls.Load() != 0 {
		t.Fatalf("EnsureDomain() = %#v, %v; want pre-dispatch rejected-policy result", state, err)
	}
}

func TestPublicDNSDomainProvisionerRecoversPanicsFailClosed(t *testing.T) {
	records := publicDNSDomainProvisionerRecords()
	t.Run("upstream panics", func(t *testing.T) {
		upstream := &publicDNSDomainProvisionerUpstream{getPanic: true, ensurePanic: true}
		provisioner := newPublicDNSDomainProvisioner(t, upstream, publicDNSDomainLookupFor(records))

		state, err := provisioner.GetDomain(context.Background(), "example.test")
		if !errors.Is(err, mail.ErrEngineUnavailable) || !reflect.DeepEqual(state, mail.EngineDomainState{}) {
			t.Fatalf("GetDomain() = %#v, %v; want panic unavailable", state, err)
		}
		state, err = provisioner.EnsureDomain(context.Background(), publicDNSDomainProvisionerSpec())
		if !errors.Is(err, mail.ErrEngineIndeterminate) || !reflect.DeepEqual(state, mail.EngineDomainState{}) {
			t.Fatalf("EnsureDomain() = %#v, %v; want panic indeterminate", state, err)
		}
	})
	t.Run("verifier lookup panics", func(t *testing.T) {
		upstream := &publicDNSDomainProvisionerUpstream{
			getState:    publicDNSDomainProvisionerDNSRequired(records),
			ensureState: publicDNSDomainProvisionerDNSRequired(records),
		}
		provisioner := newPublicDNSDomainProvisioner(t, upstream, publicDNSDomainPanicLookup{})

		state, err := provisioner.GetDomain(context.Background(), "example.test")
		if !errors.Is(err, mail.ErrEngineUnavailable) || !reflect.DeepEqual(state, mail.EngineDomainState{}) {
			t.Fatalf("GetDomain() = %#v, %v; want verifier unavailable", state, err)
		}
		state, err = provisioner.EnsureDomain(context.Background(), publicDNSDomainProvisionerSpec())
		if !errors.Is(err, mail.ErrEngineIndeterminate) || !reflect.DeepEqual(state, mail.EngineDomainState{}) {
			t.Fatalf("EnsureDomain() = %#v, %v; want verifier indeterminate", state, err)
		}
	})
}

func TestPublicDNSDomainProvisionerReturnsIndependentRecords(t *testing.T) {
	records := publicDNSDomainProvisionerRecords()
	upstream := &publicDNSDomainProvisionerUpstream{ensureState: publicDNSDomainProvisionerDNSRequired(records)}
	lookup := publicDNSDomainLookupFor(records)
	provisioner := newPublicDNSDomainProvisioner(t, upstream, lookup)

	first, err := provisioner.EnsureDomain(context.Background(), publicDNSDomainProvisionerSpec())
	if err != nil {
		t.Fatalf("first EnsureDomain() error = %v", err)
	}
	first.DNSRecords[0].Value = "mutated"
	if upstream.ensureState.DNSRecords[0].Value != records[0].Value {
		t.Fatalf("EnsureDomain() leaked returned record mutation into upstream state: %#v", upstream.ensureState.DNSRecords)
	}
	second, err := provisioner.EnsureDomain(context.Background(), publicDNSDomainProvisionerSpec())
	if err != nil {
		t.Fatalf("second EnsureDomain() error = %v", err)
	}
	want := publicDNSDomainProvisionerDNSRequired(records)
	want.State = mail.EngineDomainVerified
	want.Verification = mail.EngineVerificationVerified
	want.DKIM = mail.EngineVerificationVerified
	if !reflect.DeepEqual(second, want) {
		t.Fatalf("second EnsureDomain() = %#v, want independent %#v", second, want)
	}
}

func TestNewPublicDNSDomainProvisionerRejectsNilDependencies(t *testing.T) {
	lookup := publicDNSDomainLookupFor(publicDNSDomainProvisionerRecords())
	verifier, err := mail.NewPublicDNSVerifier(lookup)
	if err != nil {
		t.Fatalf("NewPublicDNSVerifier() error = %v", err)
	}
	if _, err := mail.NewPublicDNSDomainProvisioner(nil, verifier); !errors.Is(err, mail.ErrEngineRejectedPolicy) {
		t.Fatalf("NewPublicDNSDomainProvisioner(nil) error = %v", err)
	}
	var upstream *publicDNSDomainProvisionerUpstream
	var typedNil mail.MailboxDomainProvisioner = upstream
	if _, err := mail.NewPublicDNSDomainProvisioner(typedNil, verifier); !errors.Is(err, mail.ErrEngineRejectedPolicy) {
		t.Fatalf("NewPublicDNSDomainProvisioner(typed nil) error = %v", err)
	}
	if _, err := mail.NewPublicDNSDomainProvisioner(&publicDNSDomainProvisionerUpstream{}, nil); !errors.Is(err, mail.ErrEngineRejectedPolicy) {
		t.Fatalf("NewPublicDNSDomainProvisioner(nil verifier) error = %v", err)
	}
}

type publicDNSDomainProvisionerUpstream struct {
	getState    mail.EngineDomainState
	getErr      error
	getFn       func(context.Context, string) (mail.EngineDomainState, error)
	getPanic    bool
	ensureState mail.EngineDomainState
	ensureErr   error
	ensureFn    func(context.Context, mail.DomainSpec) (mail.EngineDomainState, error)
	ensurePanic bool
	getCalls    atomic.Int32
	ensureCalls atomic.Int32
}

var _ mail.MailboxDomainProvisioner = (*publicDNSDomainProvisionerUpstream)(nil)

func (upstream *publicDNSDomainProvisionerUpstream) GetDomain(ctx context.Context, domain string) (mail.EngineDomainState, error) {
	upstream.getCalls.Add(1)
	if upstream.getPanic {
		panic("upstream get panic")
	}
	if upstream.getFn != nil {
		return upstream.getFn(ctx, domain)
	}
	return upstream.getState, upstream.getErr
}

func (upstream *publicDNSDomainProvisionerUpstream) EnsureDomain(ctx context.Context, spec mail.DomainSpec) (mail.EngineDomainState, error) {
	upstream.ensureCalls.Add(1)
	if upstream.ensurePanic {
		panic("upstream ensure panic")
	}
	if upstream.ensureFn != nil {
		return upstream.ensureFn(ctx, spec)
	}
	return upstream.ensureState, upstream.ensureErr
}

type publicDNSDomainTXTReply struct {
	values []string
	err    error
}

type publicDNSDomainMXReply struct {
	values []mail.PublicDNSMX
	err    error
}

type publicDNSDomainCNAMEReply struct {
	value string
	err   error
}

type publicDNSDomainLookup struct {
	txt   map[string]publicDNSDomainTXTReply
	mx    map[string]publicDNSDomainMXReply
	cname map[string]publicDNSDomainCNAMEReply
	calls atomic.Int32
}

func (lookup *publicDNSDomainLookup) LookupTXT(_ context.Context, name string) ([]string, error) {
	lookup.calls.Add(1)
	reply := lookup.txt[name]
	return append([]string(nil), reply.values...), reply.err
}

func (lookup *publicDNSDomainLookup) LookupMX(_ context.Context, name string) ([]mail.PublicDNSMX, error) {
	lookup.calls.Add(1)
	reply := lookup.mx[name]
	return append([]mail.PublicDNSMX(nil), reply.values...), reply.err
}

func (lookup *publicDNSDomainLookup) LookupCNAME(_ context.Context, name string) (string, error) {
	lookup.calls.Add(1)
	reply := lookup.cname[name]
	return reply.value, reply.err
}

type publicDNSDomainPanicLookup struct{}

func (publicDNSDomainPanicLookup) LookupTXT(context.Context, string) ([]string, error) {
	panic("public DNS panic")
}

func (publicDNSDomainPanicLookup) LookupMX(context.Context, string) ([]mail.PublicDNSMX, error) {
	panic("public DNS panic")
}

func (publicDNSDomainPanicLookup) LookupCNAME(context.Context, string) (string, error) {
	panic("public DNS panic")
}

func newPublicDNSDomainProvisioner(t *testing.T, upstream mail.MailboxDomainProvisioner, lookup mail.PublicDNSLookup) *mail.PublicDNSDomainProvisioner {
	t.Helper()
	verifier, err := mail.NewPublicDNSVerifier(lookup)
	if err != nil {
		t.Fatalf("NewPublicDNSVerifier() error = %v", err)
	}
	provisioner, err := mail.NewPublicDNSDomainProvisioner(upstream, verifier)
	if err != nil {
		t.Fatalf("NewPublicDNSDomainProvisioner() error = %v", err)
	}
	return provisioner
}

func publicDNSDomainProvisionerSpec() mail.DomainSpec {
	return mail.DomainSpec{ProvisioningKey: "domain-key", Domain: "example.test", DesiredGeneration: 1, DesiredHash: "desired-hash", Present: true, Enabled: true}
}

func publicDNSDomainProvisionerDNSRequired(records []mail.EngineDNSRecord) mail.EngineDomainState {
	return mail.EngineDomainState{
		DomainID:     "domain-opaque",
		State:        mail.EngineDomainDNSRequired,
		Applied:      true,
		DNSRecords:   append([]mail.EngineDNSRecord(nil), records...),
		Verification: mail.EngineVerificationPending,
		DKIM:         mail.EngineVerificationPending,
	}
}

func publicDNSDomainProvisionerRecords() []mail.EngineDNSRecord {
	return []mail.EngineDNSRecord{
		{Type: mail.EngineDNSRecordMX, Name: "example.test", Value: "10 mail.example.test", TTL: 300},
		{Type: mail.EngineDNSRecordTXT, Name: "example.test", Value: "v=spf1 -all", TTL: 300},
		{Type: mail.EngineDNSRecordTXT, Name: "_dmarc.example.test", Value: "v=DMARC1; p=none", TTL: 300},
		{Type: mail.EngineDNSRecordTXT, Name: "selector._domainkey.example.test", Value: "v=DKIM1; p=key", TTL: 300},
		{Type: mail.EngineDNSRecordCNAME, Name: "alias.example.test", Value: "target.example.test", TTL: 300},
	}
}

func publicDNSDomainLookupFor(records []mail.EngineDNSRecord) *publicDNSDomainLookup {
	lookup := &publicDNSDomainLookup{
		txt:   make(map[string]publicDNSDomainTXTReply),
		mx:    make(map[string]publicDNSDomainMXReply),
		cname: make(map[string]publicDNSDomainCNAMEReply),
	}
	for _, record := range records {
		name := record.Name + "."
		switch record.Type {
		case mail.EngineDNSRecordTXT:
			reply := lookup.txt[name]
			reply.values = append(reply.values, record.Value)
			lookup.txt[name] = reply
		case mail.EngineDNSRecordMX:
			lookup.mx[name] = publicDNSDomainMXReply{values: []mail.PublicDNSMX{{Preference: 10, Host: "mail.example.test."}}}
		case mail.EngineDNSRecordCNAME:
			lookup.cname[name] = publicDNSDomainCNAMEReply{value: record.Value + "."}
		}
	}
	return lookup
}
