package mailvalue

import (
	"encoding/json"
	"unicode"
	"unicode/utf8"
)

const (
	// MaxRequiredDNSRecordItems ограничивает публичные DNS-требования домена.
	MaxRequiredDNSRecordItems         = 32
	maxRequiredDNSRecordsJSONBytes = 65536
)

type requiredDNSRecordsValidationError struct{}

func (requiredDNSRecordsValidationError) Error() string {
	return "invalid required DNS records"
}

var errRequiredDNSRecords error = requiredDNSRecordsValidationError{}

// EngineDNSRecordType — тип публичной DNS-записи.
type EngineDNSRecordType string

const (
	// EngineDNSRecordTXT обозначает TXT-запись.
	EngineDNSRecordTXT EngineDNSRecordType = "TXT"
	// EngineDNSRecordMX обозначает MX-запись.
	EngineDNSRecordMX EngineDNSRecordType = "MX"
	// EngineDNSRecordCNAME обозначает CNAME-запись.
	EngineDNSRecordCNAME EngineDNSRecordType = "CNAME"
)

// RequiredDNSRecord хранит публичную DNS-запись, нужную для домена.
type RequiredDNSRecord struct {
	Type  EngineDNSRecordType `json:"type"`
	Name  string              `json:"name"`
	Value string              `json:"value"`
	TTL   int                 `json:"ttl"`
}

// RequiredDNSRecords хранит упорядоченный набор публичных DNS-записей домена.
type RequiredDNSRecords []RequiredDNSRecord

// Validate проверяет ограниченный набор DNS-записей без изменения исходных значений.
func (records RequiredDNSRecords) Validate() error {
	if records == nil || len(records) > MaxRequiredDNSRecordItems {
		return errRequiredDNSRecords
	}
	seen := make(map[[3]string]struct{}, len(records))
	for _, record := range records {
		if !validRequiredDNSRecordType(record.Type) ||
			!validRequiredDNSRecordText(record.Name, 253) ||
			!validRequiredDNSRecordText(record.Value, 2048) ||
			record.TTL < 60 || record.TTL > 86400 {
			return errRequiredDNSRecords
		}
		key := [3]string{string(record.Type), record.Name, record.Value}
		if _, exists := seen[key]; exists {
			return errRequiredDNSRecords
		}
		seen[key] = struct{}{}
	}
	encoded, err := json.Marshal(records)
	if err != nil || len(encoded) > maxRequiredDNSRecordsJSONBytes {
		return errRequiredDNSRecords
	}
	return nil
}

func validRequiredDNSRecordType(recordType EngineDNSRecordType) bool {
	return recordType == EngineDNSRecordMX || recordType == EngineDNSRecordTXT || recordType == EngineDNSRecordCNAME
}

func validRequiredDNSRecordText(value string, maximumBytes int) bool {
	if len(value) > maximumBytes || !utf8.ValidString(value) {
		return false
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return false
		}
	}
	return true
}
