package mail

import "github.com/valerii/hub/backend/internal/mailvalue"

// EngineDNSRecordType — тип публичной DNS-записи.
type EngineDNSRecordType = mailvalue.EngineDNSRecordType

const (
	// EngineDNSRecordTXT обозначает TXT-запись.
	EngineDNSRecordTXT = mailvalue.EngineDNSRecordTXT
	// EngineDNSRecordMX обозначает MX-запись.
	EngineDNSRecordMX = mailvalue.EngineDNSRecordMX
	// EngineDNSRecordCNAME обозначает CNAME-запись.
	EngineDNSRecordCNAME = mailvalue.EngineDNSRecordCNAME
	// MaxEngineDNSRecordItems ограничивает публичные DNS-требования домена.
	MaxEngineDNSRecordItems = mailvalue.MaxRequiredDNSRecordItems
)

// RequiredDNSRecord хранит публичную DNS-запись, нужную для домена.
type RequiredDNSRecord = mailvalue.RequiredDNSRecord

// RequiredDNSRecords хранит упорядоченный набор публичных DNS-записей домена.
type RequiredDNSRecords = mailvalue.RequiredDNSRecords
