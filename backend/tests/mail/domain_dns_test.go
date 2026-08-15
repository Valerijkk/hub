package mail_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/valerii/hub/backend/internal/mail"
)

func TestRequiredDNSRecordsValidate_AcceptsConcreteEmptyAndValidJSON(t *testing.T) {
	empty := mail.RequiredDNSRecords{}
	if err := empty.Validate(); err != nil {
		t.Fatalf("RequiredDNSRecords{}.Validate() error = %v", err)
	}
	encodedEmpty, err := json.Marshal(empty)
	if err != nil {
		t.Fatalf("marshal empty records = %v", err)
	}
	if got := string(encodedEmpty); got != "[]" {
		t.Fatalf("marshal empty records = %q, want []", got)
	}

	records := mail.RequiredDNSRecords{
		{Type: mail.EngineDNSRecordMX, Name: "mail.example.test", Value: "10 mx.example.test", TTL: 60},
		{Type: mail.EngineDNSRecordTXT, Name: "_dmarc.example.test", Value: "v=DMARC1; p=none", TTL: 3600},
		{Type: mail.EngineDNSRecordCNAME, Name: "selector._domainkey.example.test", Value: "selector.example.test", TTL: 86400},
	}
	if err := records.Validate(); err != nil {
		t.Fatalf("valid records Validate() error = %v", err)
	}
	encoded, err := json.Marshal(records)
	if err != nil {
		t.Fatalf("marshal valid records = %v", err)
	}
	want := `[{"type":"MX","name":"mail.example.test","value":"10 mx.example.test","ttl":60},{"type":"TXT","name":"_dmarc.example.test","value":"v=DMARC1; p=none","ttl":3600},{"type":"CNAME","name":"selector._domainkey.example.test","value":"selector.example.test","ttl":86400}]`
	if got := string(encoded); got != want {
		t.Fatalf("marshal valid records = %s, want %s", got, want)
	}
}

func TestRequiredDNSRecordsValidate_RejectsNil(t *testing.T) {
	var records mail.RequiredDNSRecords
	requireInvalidDNSRecords(t, records)
}

func TestRequiredDNSRecordsValidate_EnforcesCountAndTupleUniqueness(t *testing.T) {
	maximum := make(mail.RequiredDNSRecords, 32)
	for index := range maximum {
		maximum[index] = mail.RequiredDNSRecord{
			Type:  mail.EngineDNSRecordTXT,
			Name:  "record-" + string(rune('a'+index/26)) + string(rune('a'+index%26)) + ".example.test",
			Value: "value",
			TTL:   60,
		}
	}
	if err := maximum.Validate(); err != nil {
		t.Fatalf("32 records Validate() error = %v", err)
	}

	overMaximum := append(append(mail.RequiredDNSRecords{}, maximum...), mail.RequiredDNSRecord{
		Type: mail.EngineDNSRecordTXT, Name: "record.example.test", Value: "value-over", TTL: 60,
	})
	requireInvalidDNSRecords(t, overMaximum)

	duplicate := mail.RequiredDNSRecords{
		{Type: mail.EngineDNSRecordTXT, Name: "_dmarc.example.test", Value: "v=DMARC1; p=none", TTL: 60},
		{Type: mail.EngineDNSRecordTXT, Name: "_dmarc.example.test", Value: "v=DMARC1; p=none", TTL: 86400},
	}
	requireInvalidDNSRecords(t, duplicate)

	distinct := mail.RequiredDNSRecords{
		{Type: mail.EngineDNSRecordTXT, Name: "_dmarc.example.test", Value: "v=DMARC1; p=none", TTL: 60},
		{Type: mail.EngineDNSRecordTXT, Name: "_dmarc.example.test", Value: "v=DMARC1; p=reject", TTL: 60},
	}
	if err := distinct.Validate(); err != nil {
		t.Fatalf("distinct triples Validate() error = %v", err)
	}
}

func TestRequiredDNSRecordsValidate_EnforcesTypeTTLAndTextBounds(t *testing.T) {
	validName253 := strings.Repeat("n", 253)
	validValue2048 := strings.Repeat("v", 2048)
	for _, record := range []mail.RequiredDNSRecord{
		{Type: mail.EngineDNSRecordMX, Name: "", Value: "v", TTL: 60},
		{Type: mail.EngineDNSRecordTXT, Name: "n", Value: "", TTL: 60},
		{Type: mail.EngineDNSRecordMX, Name: "n", Value: "v", TTL: 60},
		{Type: mail.EngineDNSRecordCNAME, Name: validName253, Value: validValue2048, TTL: 86400},
	} {
		if err := (mail.RequiredDNSRecords{record}).Validate(); err != nil {
			t.Fatalf("boundary record %#v Validate() error = %v", record, err)
		}
	}

	for _, record := range []mail.RequiredDNSRecord{
		{Type: "", Name: "n", Value: "v", TTL: 60},
		{Type: mail.EngineDNSRecordType("AAAA"), Name: "n", Value: "v", TTL: 60},
		{Type: mail.EngineDNSRecordMX, Name: strings.Repeat("n", 254), Value: "v", TTL: 60},
		{Type: mail.EngineDNSRecordMX, Name: "n", Value: strings.Repeat("v", 2049), TTL: 60},
		{Type: mail.EngineDNSRecordMX, Name: "n", Value: "v", TTL: 59},
		{Type: mail.EngineDNSRecordMX, Name: "n", Value: "v", TTL: 86401},
	} {
		requireInvalidDNSRecords(t, mail.RequiredDNSRecords{record})
	}
}

func TestRequiredDNSRecordsValidate_RejectsInvalidUTF8AndControls(t *testing.T) {
	for _, record := range []mail.RequiredDNSRecord{
		{Type: mail.EngineDNSRecordTXT, Name: string([]byte{0xff}), Value: "v", TTL: 60},
		{Type: mail.EngineDNSRecordTXT, Name: "n", Value: string([]byte{0xff}), TTL: 60},
		{Type: mail.EngineDNSRecordTXT, Name: "n\x00", Value: "v", TTL: 60},
		{Type: mail.EngineDNSRecordTXT, Name: "n", Value: "v\x00", TTL: 60},
		{Type: mail.EngineDNSRecordTXT, Name: "n\u0085", Value: "v", TTL: 60},
		{Type: mail.EngineDNSRecordTXT, Name: "n", Value: "v\r\nother", TTL: 60},
	} {
		requireInvalidDNSRecords(t, mail.RequiredDNSRecords{record})
	}
}

func TestRequiredDNSRecordsValidate_EnforcesCanonicalJSONLimit(t *testing.T) {
	withinLimit := recordsWithSerializedSize(t, 65536)
	if err := withinLimit.Validate(); err != nil {
		t.Fatalf("65,536-byte records Validate() error = %v", err)
	}

	overLimit := recordsWithSerializedSize(t, 65537)
	requireInvalidDNSRecords(t, overLimit)
}

func TestRequiredDNSRecordsValidate_DoesNotMutateOrNormalizeRecords(t *testing.T) {
	valid := mail.RequiredDNSRecords{{
		Type: mail.EngineDNSRecordTXT, Name: "Name\u00a0", Value: "Value\u0301", TTL: 60,
	}}
	assertRecordsUnchangedAfterValidation(t, valid)

	records := mail.RequiredDNSRecords{
		{Type: mail.EngineDNSRecordTXT, Name: "Name\u00a0", Value: "Value\u0301", TTL: 60},
		{Type: mail.EngineDNSRecordMX, Name: "bad\x00", Value: "secret-value", TTL: 60},
	}
	assertRecordsUnchangedAfterValidation(t, records)
	if err := records.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want stable validation error")
	} else if strings.Contains(err.Error(), "secret-value") {
		t.Fatal("Validate() error leaked record value")
	}
}

func assertRecordsUnchangedAfterValidation(t *testing.T, records mail.RequiredDNSRecords) {
	t.Helper()
	before, err := json.Marshal(records)
	if err != nil {
		t.Fatalf("marshal source records = %v", err)
	}
	_ = records.Validate()
	after, err := json.Marshal(records)
	if err != nil {
		t.Fatalf("marshal records after Validate = %v", err)
	}
	if got := string(after); got != string(before) {
		t.Fatalf("Validate() mutated records: got %s, want %s", got, before)
	}
}

func requireInvalidDNSRecords(t *testing.T, records mail.RequiredDNSRecords) {
	t.Helper()
	err := records.Validate()
	if err == nil {
		t.Fatal("Validate() error = nil, want stable validation error")
	}
	if got := err.Error(); got != "invalid required DNS records" {
		t.Fatalf("Validate() error = %q, want stable content-free text", got)
	}
}

func recordsWithSerializedSize(t *testing.T, want int) mail.RequiredDNSRecords {
	t.Helper()
	records := make(mail.RequiredDNSRecords, 32)
	for index := range records {
		records[index] = mail.RequiredDNSRecord{
			Type:  mail.EngineDNSRecordTXT,
			Name:  "record-" + string(rune('a'+index/26)) + string(rune('a'+index%26)),
			Value: "v",
			TTL:   60,
		}
	}
	encoded, err := json.Marshal(records)
	if err != nil {
		t.Fatalf("marshal minimum records = %v", err)
	}
	remaining := want - len(encoded)
	for index := range records {
		if remaining == 0 {
			break
		}
		addition := min(remaining, 2047)
		records[index].Value += strings.Repeat("v", addition)
		remaining -= addition
	}
	if remaining != 0 {
		t.Fatalf("cannot construct %d-byte DNS JSON within record bounds", want)
	}
	encoded, err = json.Marshal(records)
	if err != nil {
		t.Fatalf("marshal sized records = %v", err)
	}
	if got := len(encoded); got != want {
		t.Fatalf("sized records JSON length = %d, want %d", got, want)
	}
	return records
}
