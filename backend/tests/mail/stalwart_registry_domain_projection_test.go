package mail_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/valerii/hub/backend/internal/mail/stalwart"
)

func TestLocalDomainDNSZoneFileProjectMailCriticalDNS_ProjectsStalwartZone(t *testing.T) {
	zone := strings.ReplaceAll(`; Stalwart v0.16.18 manual DNS output
Example.TEST. 3600 IN MX 10 Mail.Example.TEST. ; comment outside quotes
Example.TEST. IN TXT "v=spf1 " "include:mail.example.test ~all"
_dmarc.Example.TEST. IN TXT "v=DMARC1; p=none"
selector._domainkey.Example.TEST. 600 IN TXT (
  "v=DKIM1; k=rsa; p=MIIB"
  "AQEFAAOCAQ8A"
)
mail.Example.TEST. 60 IN CNAME Target.Example.TEST.
`, "\n", "\r\n")

	projection, err := projectMailCriticalDNS(t, zone)
	if err != nil {
		t.Fatalf("ProjectMailCriticalDNS() error = %v", err)
	}
	if got, want := projection.DomainID(), "domain-1"; got != want {
		t.Fatalf("DomainID() = %q, want %q", got, want)
	}
	if got, want := projection.DomainName(), "example.test"; got != want {
		t.Fatalf("DomainName() = %q, want %q", got, want)
	}
	want := []stalwart.LocalDomainDNSRecord{
		{Type: stalwart.LocalDomainDNSRecordCNAME, Name: "mail.example.test", Value: "target.example.test", TTL: 60},
		{Type: stalwart.LocalDomainDNSRecordMX, Name: "example.test", Value: "10 mail.example.test", TTL: 3600},
		{Type: stalwart.LocalDomainDNSRecordTXT, Name: "_dmarc.example.test", Value: "v=DMARC1; p=none", TTL: 300},
		{Type: stalwart.LocalDomainDNSRecordTXT, Name: "example.test", Value: "v=spf1 include:mail.example.test ~all", TTL: 300},
		{Type: stalwart.LocalDomainDNSRecordTXT, Name: "selector._domainkey.example.test", Value: "v=DKIM1; k=rsa; p=MIIBAQEFAAOCAQ8A", TTL: 600},
	}
	if got := projection.Records(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Records() = %#v, want %#v", got, want)
	}
}

func TestLocalDomainDNSZoneFileProjectMailCriticalDNS_IgnoresKnownNonMailClasses(t *testing.T) {
	zone := completeMailCriticalZone() + `
example.test. IN A 192.0.2.1
example.test. IN AAAA 2001:db8::1
example.test. IN CAA 0 issue "ca.example.test"
_submission._tcp.example.test. IN SRV 0 1 587 mail.example.test.
_25._tcp.example.test. IN TLSA 3 1 1 0123456789abcdef
`

	projection, err := projectMailCriticalDNS(t, zone)
	if err != nil {
		t.Fatalf("ProjectMailCriticalDNS() error = %v", err)
	}
	if got, want := len(projection.Records()), 4; got != want {
		t.Fatalf("projected records = %d, want %d", got, want)
	}
}

func TestLocalDomainDNSZoneFileProjectMailCriticalDNS_SortsAndClonesRecords(t *testing.T) {
	projection, err := projectMailCriticalDNS(t, completeMailCriticalZone()+`
mail.example.test. 60 IN CNAME Target.Example.TEST.
`)
	if err != nil {
		t.Fatalf("ProjectMailCriticalDNS() error = %v", err)
	}

	first := projection.Records()
	if len(first) == 0 {
		t.Fatal("Records() returned no records")
	}
	if got, want := first[0].Type, stalwart.LocalDomainDNSRecordCNAME; got != want {
		t.Fatalf("first sorted record type = %q, want %q", got, want)
	}
	first[0].Value = "mutated.example.test"
	second := projection.Records()
	if got, want := second[0].Value, "target.example.test"; got != want {
		t.Fatalf("Records() reused mutable slice: got %q, want %q", got, want)
	}
	var zero stalwart.LocalDomainDNSProjection
	if records := zero.Records(); records == nil {
		t.Fatal("zero Records() = nil, want independent non-nil clone")
	}
}

func TestLocalDomainDNSZoneFileProjectMailCriticalDNS_RequiresMailSemanticFamilies(t *testing.T) {
	for _, testCase := range []struct {
		name string
		zone string
	}{
		{
			name: "missing apex mx",
			zone: `example.test. IN TXT "v=spf1 -all"
_dmarc.example.test. IN TXT "v=DMARC1; p=none"
selector._domainkey.example.test. IN TXT "v=DKIM1; p=key"
`,
		},
		{
			name: "missing spf",
			zone: `example.test. IN MX 10 mail.example.test.
_dmarc.example.test. IN TXT "v=DMARC1; p=none"
selector._domainkey.example.test. IN TXT "v=DKIM1; p=key"
`,
		},
		{
			name: "missing dmarc",
			zone: `example.test. IN MX 10 mail.example.test.
example.test. IN TXT "v=spf1 -all"
selector._domainkey.example.test. IN TXT "v=DKIM1; p=key"
`,
		},
		{
			name: "missing dkim",
			zone: `example.test. IN MX 10 mail.example.test.
example.test. IN TXT "v=spf1 -all"
_dmarc.example.test. IN TXT "v=DMARC1; p=none"
`,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := projectMailCriticalDNS(t, testCase.zone)
			if !errors.Is(err, stalwart.ErrStateMismatch) {
				t.Fatalf("ProjectMailCriticalDNS() error = %v, want ErrStateMismatch", err)
			}
		})
	}
}

func TestLocalDomainDNSZoneFileProjectMailCriticalDNS_RejectsForeignDuplicateAndCNAMEConflict(t *testing.T) {
	for _, testCase := range []struct {
		name string
		zone string
	}{
		{
			name: "foreign owner",
			zone: completeMailCriticalZone() + `
other.test. IN TXT "untrusted"
`,
		},
		{
			name: "exact duplicate",
			zone: completeMailCriticalZone() + `
example.test. IN MX 10 mail.example.test.
`,
		},
		{
			name: "duplicate semantic record differing ttl",
			zone: completeMailCriticalZone() + `
example.test. 60 IN MX 10 mail.example.test.
`,
		},
		{
			name: "conflicting cname targets",
			zone: completeMailCriticalZone() + `
mail.example.test. IN CNAME one.example.test.
mail.example.test. IN CNAME two.example.test.
`,
		},
		{
			name: "cname shares owner with another class",
			zone: completeMailCriticalZone() + `
mail.example.test. IN A 192.0.2.10
mail.example.test. IN CNAME target.example.test.
`,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := projectMailCriticalDNS(t, testCase.zone)
			if !errors.Is(err, stalwart.ErrProtocol) {
				t.Fatalf("ProjectMailCriticalDNS() error = %v, want ErrProtocol", err)
			}
		})
	}
}

func TestLocalDomainDNSZoneFileProjectMailCriticalDNS_RejectsEmptyAndOversizeTXT(t *testing.T) {
	for _, testCase := range []struct {
		name string
		zone string
	}{
		{
			name: "empty text",
			zone: completeMailCriticalZone() + `
note.example.test. IN TXT ""
`,
		},
		{
			name: "text exceeds structured record bound",
			zone: completeMailCriticalZone() + "note.example.test. IN TXT \"" + strings.Repeat("x", 2049) + "\"\n",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := projectMailCriticalDNS(t, testCase.zone)
			if !errors.Is(err, stalwart.ErrProtocol) {
				t.Fatalf("ProjectMailCriticalDNS() error = %v, want ErrProtocol", err)
			}
		})
	}
}

func TestLocalDomainDNSZoneFileProjectMailCriticalDNS_RejectsUnknownDirectivesAndInvalidEscapes(t *testing.T) {
	for _, testCase := range []struct {
		name string
		zone string
	}{
		{
			name: "origin directive",
			zone: "$ORIGIN example.test.\n" + completeMailCriticalZone(),
		},
		{
			name: "unknown class",
			zone: completeMailCriticalZone() + `
example.test. IN NS ns1.example.test.
`,
		},
		{
			name: "short escaped decimal",
			zone: completeMailCriticalZone() + `
note.example.test. IN TXT "bad\12"
`,
		},
		{
			name: "unsupported escaped character",
			zone: completeMailCriticalZone() + `
note.example.test. IN TXT "bad\n"
`,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := projectMailCriticalDNS(t, testCase.zone)
			if !errors.Is(err, stalwart.ErrProtocol) {
				t.Fatalf("ProjectMailCriticalDNS() error = %v, want ErrProtocol", err)
			}
		})
	}
}

func TestLocalDomainDNSZoneFileProjectMailCriticalDNS_RejectsControlOversizeAndTooManyRecords(t *testing.T) {
	t.Run("control input", func(t *testing.T) {
		_, err := readProjectionZone(t, completeMailCriticalZone()+"\x00")
		if !errors.Is(err, stalwart.ErrProtocol) {
			t.Fatalf("ReadLocalDomainDNSZoneFile() error = %v, want ErrProtocol", err)
		}
	})

	t.Run("oversize input", func(t *testing.T) {
		_, err := readProjectionZone(t, strings.Repeat("z", 64<<10))
		if !errors.Is(err, stalwart.ErrProtocol) {
			t.Fatalf("ReadLocalDomainDNSZoneFile() error = %v, want ErrProtocol", err)
		}
	})

	t.Run("too many projected records", func(t *testing.T) {
		zone := completeMailCriticalZone()
		for index := 0; index < 29; index++ {
			zone += fmt.Sprintf("record-%d.example.test. IN TXT \"extra\"\n", index)
		}
		_, err := projectMailCriticalDNS(t, zone)
		if !errors.Is(err, stalwart.ErrProtocol) {
			t.Fatalf("ProjectMailCriticalDNS() error = %v, want ErrProtocol", err)
		}
	})
}

func completeMailCriticalZone() string {
	return `example.test. IN MX 10 mail.example.test.
example.test. IN TXT "v=spf1 -all"
_dmarc.example.test. IN TXT "v=DMARC1; p=none"
selector._domainkey.example.test. IN TXT "v=DKIM1; p=key"
`
}

func projectMailCriticalDNS(t *testing.T, zone string) (stalwart.LocalDomainDNSProjection, error) {
	t.Helper()
	result, err := readProjectionZone(t, zone)
	if err != nil {
		return stalwart.LocalDomainDNSProjection{}, err
	}
	return result.ProjectMailCriticalDNS()
}

func readProjectionZone(t *testing.T, zone string) (stalwart.LocalDomainDNSZoneFile, error) {
	t.Helper()
	replies := []registryDomainReply{
		registryDomainQueryReply([]string{"domain-1"}),
		registryDomainDNSZoneFileReply("domain-1", localRegistryDomain, true, false, zone),
	}
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		calls++
		if request.Method != http.MethodPost || request.URL.Path != "/jmap" || calls > len(replies) {
			writer.WriteHeader(http.StatusForbidden)
			return
		}
		var got stalwart.JMAPRequest
		if err := json.NewDecoder(request.Body).Decode(&got); err != nil || !registryDomainRequestEqual(got, replies[calls-1].want) {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(replies[calls-1].response))
	}))
	t.Cleanup(func() {
		server.Close()
		if calls != len(replies) {
			t.Errorf("Registry calls = %d, want %d", calls, len(replies))
		}
	})

	client, err := stalwart.NewClient(stalwart.Config{
		Origin:             server.URL,
		HTTPClient:         server.Client(),
		RegistryAuthorizer: registryDomainTestAuthorizer{},
		MailboxAuthorizer:  registryDomainTestAuthorizer{},
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	controller, err := stalwart.NewRegistryDomainController(client)
	if err != nil {
		t.Fatalf("NewRegistryDomainController() error = %v", err)
	}
	return controller.ReadLocalDomainDNSZoneFile(context.Background(), localRegistryDomain)
}
