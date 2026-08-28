package stalwart

import (
	"net/netip"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	localDomainDNSDefaultTTL         = 300
	localDomainDNSMinimumTTL         = 60
	localDomainDNSMaximumTTL         = 86400
	localDomainDNSMaximumRecords     = 32
	localDomainDNSMaximumParentheses = 16
	localDomainDNSMaximumTXTBytes    = 2048
)

// LocalDomainDNSRecordType identifies a mail-critical DNS record.
type LocalDomainDNSRecordType string

const (
	// LocalDomainDNSRecordTXT identifies a TXT record.
	LocalDomainDNSRecordTXT LocalDomainDNSRecordType = "TXT"
	// LocalDomainDNSRecordMX identifies an MX record.
	LocalDomainDNSRecordMX LocalDomainDNSRecordType = "MX"
	// LocalDomainDNSRecordCNAME identifies a CNAME record.
	LocalDomainDNSRecordCNAME LocalDomainDNSRecordType = "CNAME"
)

// LocalDomainDNSRecord is one canonical mail-critical DNS record.
type LocalDomainDNSRecord struct {
	Type  LocalDomainDNSRecordType
	Name  string
	Value string
	TTL   int
}

// LocalDomainDNSProjection is an immutable structured projection of a local
// domain's mail-critical manual DNS records.
type LocalDomainDNSProjection struct {
	domainID   string
	domainName string
	records    []LocalDomainDNSRecord
}

// DomainID returns the opaque Registry domain identifier.
func (projection LocalDomainDNSProjection) DomainID() string {
	return strings.Clone(projection.domainID)
}

// DomainName returns the canonical local domain name.
func (projection LocalDomainDNSProjection) DomainName() string {
	return strings.Clone(projection.domainName)
}

// Records returns an independent, non-nil clone of the projected records.
func (projection LocalDomainDNSProjection) Records() []LocalDomainDNSRecord {
	records := make([]LocalDomainDNSRecord, len(projection.records))
	copy(records, projection.records)
	return records
}

// ProjectMailCriticalDNS parses the private Stalwart manual zone into a small,
// canonical mail-critical projection. It never returns the raw provider zone.
func (result LocalDomainDNSZoneFile) ProjectMailCriticalDNS() (LocalDomainDNSProjection, error) {
	if !validOpaque(result.domainID) {
		return LocalDomainDNSProjection{}, ErrStateMismatch
	}
	domainName, validDomainName := canonicalRegistryDomainName(result.domainName)
	if !validDomainName || domainName != result.domainName {
		return LocalDomainDNSProjection{}, ErrStateMismatch
	}
	if !validRegistryDomainDNSZoneFile(result.zoneFile) {
		return LocalDomainDNSProjection{}, ErrProtocol
	}

	logicalRecords, validLogicalRecords := localDomainDNSLogicalRecords(result.zoneFile)
	if !validLogicalRecords {
		return LocalDomainDNSProjection{}, ErrProtocol
	}
	projected := make([]LocalDomainDNSRecord, 0, localDomainDNSMaximumRecords)
	seenRecords := make(map[string]struct{}, len(logicalRecords))
	owners := make(map[string]localDomainDNSOwnerState, len(logicalRecords))
	for _, logicalRecord := range logicalRecords {
		tokens, validTokens := localDomainDNSTokens(logicalRecord)
		if !validTokens {
			return LocalDomainDNSProjection{}, ErrProtocol
		}
		record, validRecord := localDomainDNSRecord(tokens, domainName)
		if !validRecord {
			return LocalDomainDNSProjection{}, ErrProtocol
		}
		if _, duplicate := seenRecords[record.key]; duplicate {
			return LocalDomainDNSProjection{}, ErrProtocol
		}
		seenRecords[record.key] = struct{}{}

		ownerState := owners[record.record.Name]
		if record.record.Type == LocalDomainDNSRecordCNAME {
			if ownerState.hasCNAME || len(ownerState.types) != 0 {
				return LocalDomainDNSProjection{}, ErrProtocol
			}
			ownerState.hasCNAME = true
		} else {
			if ownerState.hasCNAME {
				return LocalDomainDNSProjection{}, ErrProtocol
			}
			if ownerState.types == nil {
				ownerState.types = make(map[LocalDomainDNSRecordType]struct{})
			}
			ownerState.types[record.record.Type] = struct{}{}
		}
		owners[record.record.Name] = ownerState

		if !record.project {
			continue
		}
		if len(projected) == localDomainDNSMaximumRecords {
			return LocalDomainDNSProjection{}, ErrProtocol
		}
		projected = append(projected, record.record)
	}
	if !localDomainDNSComplete(projected, domainName) {
		return LocalDomainDNSProjection{}, ErrStateMismatch
	}
	sort.Slice(projected, func(left, right int) bool {
		if projected[left].Type != projected[right].Type {
			return projected[left].Type < projected[right].Type
		}
		if projected[left].Name != projected[right].Name {
			return projected[left].Name < projected[right].Name
		}
		if projected[left].Value != projected[right].Value {
			return projected[left].Value < projected[right].Value
		}
		return projected[left].TTL < projected[right].TTL
	})
	return LocalDomainDNSProjection{
		domainID:   strings.Clone(result.domainID),
		domainName: strings.Clone(domainName),
		records:    cloneLocalDomainDNSRecords(projected),
	}, nil
}

type localDomainDNSOwnerState struct {
	hasCNAME bool
	types    map[LocalDomainDNSRecordType]struct{}
}

type localDomainDNSParsedRecord struct {
	record  LocalDomainDNSRecord
	key     string
	project bool
}

type localDomainDNSToken struct {
	value  string
	quoted bool
}

func cloneLocalDomainDNSRecords(records []LocalDomainDNSRecord) []LocalDomainDNSRecord {
	cloned := make([]LocalDomainDNSRecord, len(records))
	copy(cloned, records)
	return cloned
}

func localDomainDNSLogicalRecords(zoneFile string) ([]string, bool) {
	var (
		records []string
		builder strings.Builder
		depth   int
		quoted  bool
	)
	builder.Grow(len(zoneFile))
	finishRecord := func() {
		record := strings.TrimSpace(builder.String())
		builder.Reset()
		if record != "" {
			records = append(records, record)
		}
	}
	for index := 0; index < len(zoneFile); index++ {
		character := zoneFile[index]
		if quoted && character == '\\' {
			if index+1 >= len(zoneFile) {
				return nil, false
			}
			builder.WriteByte(character)
			index++
			builder.WriteByte(zoneFile[index])
			continue
		}
		if character == '"' {
			quoted = !quoted
			builder.WriteByte(character)
			continue
		}
		if quoted {
			if character == '\r' || character == '\n' {
				return nil, false
			}
			builder.WriteByte(character)
			continue
		}
		switch character {
		case ';':
			for index+1 < len(zoneFile) && zoneFile[index+1] != '\r' && zoneFile[index+1] != '\n' {
				index++
			}
		case '(':
			depth++
			if depth > localDomainDNSMaximumParentheses {
				return nil, false
			}
			builder.WriteByte(' ')
		case ')':
			if depth == 0 {
				return nil, false
			}
			depth--
			builder.WriteByte(' ')
		case '\r':
			if index+1 >= len(zoneFile) || zoneFile[index+1] != '\n' {
				return nil, false
			}
			index++
			fallthrough
		case '\n':
			if depth == 0 {
				finishRecord()
			} else {
				builder.WriteByte(' ')
			}
		default:
			builder.WriteByte(character)
		}
	}
	if quoted || depth != 0 {
		return nil, false
	}
	finishRecord()
	return records, true
}

func localDomainDNSTokens(record string) ([]localDomainDNSToken, bool) {
	tokens := make([]localDomainDNSToken, 0, 8)
	for index := 0; index < len(record); {
		for index < len(record) && localDomainDNSWhitespace(record[index]) {
			index++
		}
		if index == len(record) {
			break
		}
		if record[index] == '"' {
			value, next, valid := localDomainDNSQuotedToken(record, index)
			if !valid || next < len(record) && !localDomainDNSWhitespace(record[next]) {
				return nil, false
			}
			tokens = append(tokens, localDomainDNSToken{value: value, quoted: true})
			index = next
			continue
		}
		start := index
		for index < len(record) && !localDomainDNSWhitespace(record[index]) {
			if record[index] == '"' || record[index] == '\\' {
				return nil, false
			}
			index++
		}
		if start == index {
			return nil, false
		}
		tokens = append(tokens, localDomainDNSToken{value: record[start:index]})
	}
	return tokens, len(tokens) != 0
}

func localDomainDNSQuotedToken(record string, index int) (string, int, bool) {
	var builder strings.Builder
	builder.Grow(len(record) - index)
	for index++; index < len(record); index++ {
		character := record[index]
		switch character {
		case '"':
			value := builder.String()
			return value, index + 1, localDomainDNSRecordText(value)
		case '\\':
			if index+1 >= len(record) {
				return "", 0, false
			}
			next := record[index+1]
			switch next {
			case '"', '\\':
				builder.WriteByte(next)
				index++
			default:
				if index+3 >= len(record) || !localDomainDNSDecimal(record[index+1]) || !localDomainDNSDecimal(record[index+2]) || !localDomainDNSDecimal(record[index+3]) {
					return "", 0, false
				}
				value := int(record[index+1]-'0')*100 + int(record[index+2]-'0')*10 + int(record[index+3]-'0')
				if value > 255 {
					return "", 0, false
				}
				builder.WriteByte(byte(value))
				index += 3
			}
		default:
			builder.WriteByte(character)
		}
	}
	return "", 0, false
}

func localDomainDNSRecord(tokens []localDomainDNSToken, domainName string) (localDomainDNSParsedRecord, bool) {
	if len(tokens) < 4 || tokens[0].quoted {
		return localDomainDNSParsedRecord{}, false
	}
	owner, validOwner := canonicalLocalDomainDNSOwner(tokens[0].value)
	if !validOwner || !localDomainDNSOwnerWithinDomain(owner, domainName) {
		return localDomainDNSParsedRecord{}, false
	}
	index := 1
	ttl := localDomainDNSDefaultTTL
	if index < len(tokens) && !tokens[index].quoted && localDomainDNSAllDecimal(tokens[index].value) {
		parsedTTL, validTTL := localDomainDNSTTL(tokens[index].value)
		if !validTTL {
			return localDomainDNSParsedRecord{}, false
		}
		ttl = parsedTTL
		index++
	}
	if index+2 > len(tokens) || tokens[index].quoted || !strings.EqualFold(tokens[index].value, "IN") || tokens[index+1].quoted {
		return localDomainDNSParsedRecord{}, false
	}
	recordType := strings.ToUpper(tokens[index+1].value)
	rdata := tokens[index+2:]
	var (
		value   string
		project bool
	)
	switch recordType {
	case string(LocalDomainDNSRecordTXT):
		if len(rdata) == 0 {
			return localDomainDNSParsedRecord{}, false
		}
		var text strings.Builder
		for _, token := range rdata {
			if !token.quoted {
				return localDomainDNSParsedRecord{}, false
			}
			text.WriteString(token.value)
		}
		value = text.String()
		if len(value) == 0 || len(value) > localDomainDNSMaximumTXTBytes || !localDomainDNSRecordText(value) {
			return localDomainDNSParsedRecord{}, false
		}
		project = true
	case string(LocalDomainDNSRecordMX):
		if len(rdata) != 2 || rdata[0].quoted || rdata[1].quoted {
			return localDomainDNSParsedRecord{}, false
		}
		priority, validPriority := localDomainDNSUInt16(rdata[0].value)
		target, validTarget := canonicalLocalDomainDNSTarget(rdata[1].value)
		if !validPriority || !validTarget {
			return localDomainDNSParsedRecord{}, false
		}
		value = strconv.Itoa(priority) + " " + target
		project = true
	case string(LocalDomainDNSRecordCNAME):
		if len(rdata) != 1 || rdata[0].quoted {
			return localDomainDNSParsedRecord{}, false
		}
		target, validTarget := canonicalLocalDomainDNSTarget(rdata[0].value)
		if !validTarget {
			return localDomainDNSParsedRecord{}, false
		}
		value = target
		project = true
	case "A":
		if len(rdata) != 1 || rdata[0].quoted {
			return localDomainDNSParsedRecord{}, false
		}
		address, parseErr := netip.ParseAddr(rdata[0].value)
		if parseErr != nil || !address.Is4() {
			return localDomainDNSParsedRecord{}, false
		}
		value = address.String()
	case "AAAA":
		if len(rdata) != 1 || rdata[0].quoted {
			return localDomainDNSParsedRecord{}, false
		}
		address, parseErr := netip.ParseAddr(rdata[0].value)
		if parseErr != nil || !address.Is6() || address.Is4In6() {
			return localDomainDNSParsedRecord{}, false
		}
		value = address.String()
	case "CAA":
		if len(rdata) != 3 || rdata[0].quoted || rdata[1].quoted || !rdata[2].quoted {
			return localDomainDNSParsedRecord{}, false
		}
		flags, validFlags := localDomainDNSUInt8(rdata[0].value)
		if !validFlags || !localDomainDNSCAATag(rdata[1].value) || !localDomainDNSRecordText(rdata[2].value) {
			return localDomainDNSParsedRecord{}, false
		}
		value = strconv.Itoa(flags) + " " + strings.ToLower(rdata[1].value) + " " + rdata[2].value
	case "SRV":
		if len(rdata) != 4 || rdata[0].quoted || rdata[1].quoted || rdata[2].quoted || rdata[3].quoted {
			return localDomainDNSParsedRecord{}, false
		}
		priority, validPriority := localDomainDNSUInt16(rdata[0].value)
		weight, validWeight := localDomainDNSUInt16(rdata[1].value)
		port, validPort := localDomainDNSUInt16(rdata[2].value)
		target, validTarget := canonicalLocalDomainDNSTarget(rdata[3].value)
		if !validPriority || !validWeight || !validPort || !validTarget {
			return localDomainDNSParsedRecord{}, false
		}
		value = strconv.Itoa(priority) + " " + strconv.Itoa(weight) + " " + strconv.Itoa(port) + " " + target
	case "TLSA":
		if len(rdata) != 4 || rdata[0].quoted || rdata[1].quoted || rdata[2].quoted || rdata[3].quoted {
			return localDomainDNSParsedRecord{}, false
		}
		usage, validUsage := localDomainDNSUInt8(rdata[0].value)
		selector, validSelector := localDomainDNSUInt8(rdata[1].value)
		matchingType, validMatchingType := localDomainDNSUInt8(rdata[2].value)
		if !validUsage || usage > 3 || !validSelector || selector > 1 || !validMatchingType || matchingType > 2 || !localDomainDNSHex(rdata[3].value) {
			return localDomainDNSParsedRecord{}, false
		}
		value = strconv.Itoa(usage) + " " + strconv.Itoa(selector) + " " + strconv.Itoa(matchingType) + " " + strings.ToLower(rdata[3].value)
	default:
		return localDomainDNSParsedRecord{}, false
	}
	record := LocalDomainDNSRecord{Type: LocalDomainDNSRecordType(recordType), Name: owner, Value: value, TTL: ttl}
	return localDomainDNSParsedRecord{
		record:  record,
		key:     string(record.Type) + "\x00" + record.Name + "\x00" + record.Value,
		project: project,
	}, true
}

func localDomainDNSComplete(records []LocalDomainDNSRecord, domainName string) bool {
	var (
		hasMX    bool
		hasSPF   bool
		hasDMARC bool
		hasDKIM  bool
	)
	for _, record := range records {
		switch record.Type {
		case LocalDomainDNSRecordMX:
			if record.Name == domainName {
				hasMX = true
			}
		case LocalDomainDNSRecordTXT:
			if record.Name == domainName && strings.HasPrefix(record.Value, "v=spf1 ") {
				hasSPF = true
			}
			if record.Name == "_dmarc."+domainName && strings.HasPrefix(record.Value, "v=DMARC1;") {
				hasDMARC = true
			}
			if strings.HasPrefix(record.Value, "v=DKIM1;") && localDomainDNSDKIMOwner(record.Name, domainName) {
				hasDKIM = true
			}
		}
	}
	return hasMX && hasSPF && hasDMARC && hasDKIM
}

func localDomainDNSDKIMOwner(owner, domainName string) bool {
	suffix := "._domainkey." + domainName
	if !strings.HasSuffix(owner, suffix) {
		return false
	}
	selector := strings.TrimSuffix(owner, suffix)
	return selector != "" && !strings.Contains(selector, ".")
}

func canonicalLocalDomainDNSOwner(value string) (string, bool) {
	if !strings.HasSuffix(value, ".") || value == "." {
		return "", false
	}
	name := strings.TrimSuffix(value, ".")
	if len(name) == 0 || len(name) > registryDomainMaximumBytes || strings.HasSuffix(name, ".") {
		return "", false
	}
	labels := strings.Split(name, ".")
	if len(labels) < 2 {
		return "", false
	}
	for _, label := range labels {
		if !localDomainDNSOwnerLabel(label) {
			return "", false
		}
	}
	return strings.ToLower(name), true
}

func canonicalLocalDomainDNSTarget(value string) (string, bool) {
	if !strings.HasSuffix(value, ".") || value == "." {
		return "", false
	}
	return canonicalRegistryDomainName(strings.TrimSuffix(value, "."))
}

func localDomainDNSOwnerWithinDomain(owner, domainName string) bool {
	return owner == domainName || strings.HasSuffix(owner, "."+domainName)
}

func localDomainDNSOwnerLabel(label string) bool {
	if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
		return false
	}
	for _, character := range []byte(label) {
		if character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9' || character == '-' || character == '_' {
			continue
		}
		return false
	}
	return true
}

func localDomainDNSTTL(value string) (int, bool) {
	parsed, err := strconv.ParseUint(value, 10, 32)
	if err != nil || parsed < localDomainDNSMinimumTTL || parsed > localDomainDNSMaximumTTL {
		return 0, false
	}
	return int(parsed), true
}

func localDomainDNSUInt16(value string) (int, bool) {
	if !localDomainDNSAllDecimal(value) {
		return 0, false
	}
	parsed, err := strconv.ParseUint(value, 10, 16)
	if err != nil {
		return 0, false
	}
	return int(parsed), true
}

func localDomainDNSUInt8(value string) (int, bool) {
	if !localDomainDNSAllDecimal(value) {
		return 0, false
	}
	parsed, err := strconv.ParseUint(value, 10, 8)
	if err != nil {
		return 0, false
	}
	return int(parsed), true
}

func localDomainDNSAllDecimal(value string) bool {
	if value == "" {
		return false
	}
	for _, character := range []byte(value) {
		if !localDomainDNSDecimal(character) {
			return false
		}
	}
	return true
}

func localDomainDNSDecimal(value byte) bool {
	return value >= '0' && value <= '9'
}

func localDomainDNSHex(value string) bool {
	if len(value) == 0 || len(value)%2 != 0 {
		return false
	}
	for _, character := range []byte(value) {
		if character >= '0' && character <= '9' || character >= 'a' && character <= 'f' || character >= 'A' && character <= 'F' {
			continue
		}
		return false
	}
	return true
}

func localDomainDNSCAATag(value string) bool {
	if len(value) == 0 || len(value) > 15 {
		return false
	}
	for _, character := range []byte(value) {
		if character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9' || character == '-' {
			continue
		}
		return false
	}
	return true
}

func localDomainDNSRecordText(value string) bool {
	if !utf8.ValidString(value) {
		return false
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return false
		}
	}
	return true
}

func localDomainDNSWhitespace(value byte) bool {
	return value == ' ' || value == '\t'
}
