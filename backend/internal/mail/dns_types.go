package mail

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
