package rootfk

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

const (
	expectedLineCount = 82
	expectedByteCount = 4767
	expectedSHA256    = "889c14127f47bdb73108bfef07573ff4b0bde7fd3b78b47122718a1f5b0a36b7"

	maxExpectedBytes = 64 << 10
	maxInspectBytes  = 32 << 20
	maxTokenCount    = 1 << 20
)

var allowedSources = map[string]struct{}{
	"rootProof.mailAccountID":              {},
	"rootProof.mailDomainID":               {},
	"rootProof.ownerUserID":                {},
	"calendarProof.calendarEventID":        {},
	"calendarProof.inboundObjectStateID":   {},
	"calendarProof.mutationReceiptID":      {},
	"calendarProof.replyDeliveryID":        {},
	"calendarProof.rsvpAttendeeID":         {},
}

type projection struct {
	canonical string
	count     int
	bytes     int
	sha256    string
}

type tokenKind uint8

const (
	tokenName tokenKind = iota
	tokenPunctuation
	tokenLiteral
)

type token struct {
	kind   tokenKind
	text   string
	quoted bool
}

type Evidence struct { ExpectedCount, ExpectedBytes int; ExpectedSHA256 string; DesiredCount, DesiredBytes int; DesiredSHA256 string; ReplayCount, ReplayBytes int; ReplaySHA256 string }
func (e Evidence) String() string { return fmt.Sprintf("schema-root-fk expected count=%d bytes=%d sha256=%s desired count=%d bytes=%d sha256=%s replay count=%d bytes=%d sha256=%s", e.ExpectedCount,e.ExpectedBytes,e.ExpectedSHA256,e.DesiredCount,e.DesiredBytes,e.DesiredSHA256,e.ReplayCount,e.ReplayBytes,e.ReplaySHA256) }
func VerifyFiles(expectedPath, desiredPath, replayPath string) (Evidence, error) {
	expectedBytes, err := readRegularFile("expected manifest", expectedPath, maxExpectedBytes)
	if err != nil {
		return Evidence{}, err
	}
	expected, err := parseExpectedManifest(expectedBytes)
	if err != nil {
		return Evidence{}, err
	}
	desiredBytes, err := readRegularFile("desired inspect", desiredPath, maxInspectBytes)
	if err != nil {
		return Evidence{}, err
	}
	desired, err := parseInspectProjection(desiredBytes)
	if err != nil {
		return Evidence{}, fmt.Errorf("parse desired inspect: %w", err)
	}
	replayBytes, err := readRegularFile("replay inspect", replayPath, maxInspectBytes)
	if err != nil {
		return Evidence{}, err
	}
	replay, err := parseInspectProjection(replayBytes)
	if err != nil {
		return Evidence{}, fmt.Errorf("parse replay inspect: %w", err)
	}

	if desired.canonical != replay.canonical {
		return Evidence{}, errors.New("desired and replay root-FK projections differ")
	}
	if desired.canonical != expected.canonical {
		return Evidence{}, errors.New("root-FK projection does not match the frozen manifest")
	}
	return Evidence{expectedLineCount,expectedByteCount,expectedSHA256,desired.count,desired.bytes,desired.sha256,replay.count,replay.bytes,replay.sha256}, nil
}
func ParseInspectProjection(data []byte) (string,error) { p,err:=parseInspectProjection(data); if err!=nil{return "",err}; return p.canonical,nil }

func readRegularFile(label, path string, limit int64) ([]byte, error) {
	initial, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", label, err)
	}
	if !initial.Mode().IsRegular() || initial.Size() < 0 || initial.Size() > limit {
		return nil, fmt.Errorf("%s must be a bounded regular file", label)
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", label, err)
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat opened %s: %w", label, err)
	}
	if !opened.Mode().IsRegular() || !os.SameFile(initial, opened) || opened.Size() != initial.Size() {
		return nil, fmt.Errorf("%s changed while being opened", label)
	}
	data, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", label, err)
	}
	if int64(len(data)) != initial.Size() || int64(len(data)) > limit {
		return nil, fmt.Errorf("%s changed while being read", label)
	}
	return data, nil
}

func parseExpectedManifest(data []byte) (projection, error) {
	if len(data) != expectedByteCount || !strings.HasSuffix(string(data), "\n") || strings.Contains(string(data), "\r") {
		return projection{}, errors.New("expected manifest is not canonical LF-only frozen content")
	}
	if sha256Hex(data) != expectedSHA256 {
		return projection{}, errors.New("expected manifest SHA-256 does not match the frozen value")
	}
	lines := strings.Split(string(data[:len(data)-1]), "\n")
	if len(lines) != expectedLineCount {
		return projection{}, errors.New("expected manifest line count does not match the frozen value")
	}
	triples := make([]string, 0, len(lines))
	previous := ""
	for index, line := range lines {
		fields := strings.Split(line, "|")
		if len(fields) != 6 {
			return projection{}, fmt.Errorf("expected manifest line %d must have six fields", index+1)
		}
		if err := validateManifestFields(fields); err != nil {
			return projection{}, fmt.Errorf("expected manifest line %d: %w", index+1, err)
		}
		triple := strings.Join(fields[:3], "|")
		if previous != "" && triple <= previous {
			return projection{}, fmt.Errorf("expected manifest line %d is not bytewise sorted and unique", index+1)
		}
		previous = triple
		triples = append(triples, triple)
	}
	return makeProjection(triples)
}

func validateManifestFields(fields []string) error {
	if fields[0] != "Organization" && fields[0] != "User" {
		return errors.New("root must be Organization or User")
	}
	if !isStorageIdentifier(fields[1]) || !isStorageIdentifier(fields[2]) {
		return errors.New("table and column must use canonical storage identifier grammar")
	}
	switch fields[3] {
	case "forbidden":
		if fields[4] != "none" || fields[5] != "0" {
			return errors.New("forbidden slot must use none and count 0")
		}
	case "allowlisted":
		if _, ok := allowedSources[fields[4]]; !ok || fields[5] != "1" {
			return errors.New("allowlisted slot has an invalid source or count")
		}
	default:
		return errors.New("mode must be forbidden or allowlisted")
	}
	return nil
}

func isStorageIdentifier(value string) bool {
	if value == "" || value[0] < 'a' || value[0] > 'z' || value[len(value)-1] == '_' {
		return false
	}
	for index := 0; index < len(value); index++ {
		character := value[index]
		if (character < 'a' || character > 'z') && (character < '0' || character > '9') && character != '_' {
			return false
		}
	}
	return true
}

func parseInspectProjection(data []byte) (projection, error) {
	tokens, err := lexSQL(data)
	if err != nil {
		return projection{}, err
	}
	triples := make([]string, 0, expectedLineCount)
	tableScopeCount := 0
	for index := 0; index < len(tokens); {
		if isKeyword(tokens[index], "create") && index+1 < len(tokens) && isKeyword(tokens[index+1], "table") {
			table, bodyStart, bodyEnd, next, err := parseCreateTable(tokens, index)
			if err != nil {
				return projection{}, err
			}
			tableScopeCount++
			entries, err := parseTableRootForeignKeys(table, tokens[bodyStart+1:bodyEnd])
			if err != nil {
				return projection{}, err
			}
			triples = append(triples, entries...)
			index = next
			continue
		}
		if isKeyword(tokens[index], "alter") && index+1 < len(tokens) && isKeyword(tokens[index+1], "table") {
			table, actionsStart, actionsEnd, next, err := parseAlterTable(tokens, index)
			if err != nil {
				return projection{}, err
			}
			tableScopeCount++
			entries, err := parseAlterTableRootForeignKeys(table, tokens[actionsStart:actionsEnd])
			if err != nil {
				return projection{}, err
			}
			triples = append(triples, entries...)
			index = next
			continue
		}
		index++
	}
	if tableScopeCount == 0 {
		return projection{}, errors.New("inspect SQL contains no CREATE TABLE or ALTER TABLE scope")
	}
	return makeProjection(triples)
}

func parseCreateTable(tokens []token, create int) (string, int, int, int, error) {
	index := create + 2
	if index+2 < len(tokens) && isKeyword(tokens[index], "if") && isKeyword(tokens[index+1], "not") && isKeyword(tokens[index+2], "exists") {
		index += 3
	}
	table, next, err := parseQualifiedName(tokens, index, len(tokens))
	if err != nil {
		return "", 0, 0, 0, fmt.Errorf("parse CREATE TABLE name: %w", err)
	}
	if next >= len(tokens) || tokens[next].text != "(" {
		return "", 0, 0, 0, errors.New("CREATE TABLE must have a parenthesized definition")
	}
	bodyStart := next
	depth := 0
	for index = bodyStart; index < len(tokens); index++ {
		switch tokens[index].text {
		case "(":
			depth++
		case ")":
			depth--
			if depth == 0 {
				if index+1 >= len(tokens) || tokens[index+1].text != ";" {
					return "", 0, 0, 0, errors.New("CREATE TABLE scope must end with a semicolon")
				}
				return table, bodyStart, index, index + 2, nil
			}
			if depth < 0 {
				return "", 0, 0, 0, errors.New("CREATE TABLE has an unbalanced definition")
			}
		}
	}
	return "", 0, 0, 0, errors.New("CREATE TABLE has an unterminated definition")
}

func parseAlterTable(tokens []token, alter int) (string, int, int, int, error) {
	index := alter + 2
	if index < len(tokens) && isKeyword(tokens[index], "if") {
		if index+1 >= len(tokens) || !isKeyword(tokens[index+1], "exists") {
			return "", 0, 0, 0, errors.New("ALTER TABLE IF must be followed by EXISTS")
		}
		index += 2
	}
	if index < len(tokens) && isKeyword(tokens[index], "only") {
		index++
	}
	table, actionsStart, err := parseQualifiedName(tokens, index, len(tokens))
	if err != nil {
		return "", 0, 0, 0, fmt.Errorf("parse ALTER TABLE name: %w", err)
	}
	depth := 0
	for index = actionsStart; index < len(tokens); index++ {
		switch tokens[index].text {
		case "(":
			depth++
		case ")":
			depth--
			if depth < 0 {
				return "", 0, 0, 0, errors.New("ALTER TABLE has unbalanced actions")
			}
		case ";":
			if depth == 0 {
				if index == actionsStart {
					return "", 0, 0, 0, errors.New("ALTER TABLE must include at least one action")
				}
				return table, actionsStart, index, index + 1, nil
			}
		}
	}
	return "", 0, 0, 0, errors.New("ALTER TABLE has an unterminated action list")
}

func parseAlterTableRootForeignKeys(table string, actions []token) ([]string, error) {
	entries := make([]string, 0)
	start := 0
	depth := 0
	for index := 0; index <= len(actions); index++ {
		atEnd := index == len(actions)
		if !atEnd {
			switch actions[index].text {
			case "(":
				depth++
			case ")":
				depth--
				if depth < 0 {
					return nil, errors.New("ALTER TABLE action has unbalanced parentheses")
				}
			}
		}
		if !atEnd && (actions[index].text != "," || depth != 0) {
			continue
		}
		if index == start {
			return nil, errors.New("ALTER TABLE has an empty top-level action")
		}
		entry, err := parseAlterRootForeignKeyAction(table, actions[start:index])
		if err != nil {
			return nil, err
		}
		if entry != "" {
			entries = append(entries, entry)
		}
		start = index + 1
	}
	if depth != 0 {
		return nil, errors.New("ALTER TABLE actions have unbalanced parentheses")
	}
	return entries, nil
}

func parseAlterRootForeignKeyAction(table string, action []token) (string, error) {
	hasRootReference, err := hasExactlyOneRootReference(action)
	if err != nil || !hasRootReference {
		return "", err
	}
	if len(action) == 0 || !isKeyword(action[0], "add") {
		return "", errors.New("direct root foreign-key ALTER action must start with ADD")
	}
	foreign := 1
	if foreign < len(action) && isKeyword(action[foreign], "constraint") {
		if foreign+1 >= len(action) || action[foreign+1].kind != tokenName {
			return "", errors.New("ALTER TABLE root foreign-key CONSTRAINT requires a name")
		}
		foreign += 2
	}
	if foreign >= len(action) || !isKeyword(action[foreign], "foreign") {
		return "", errors.New("direct root foreign-key ALTER action must use FOREIGN KEY")
	}
	for index := foreign + 1; index < len(action); index++ {
		if isKeyword(action[index], "foreign") {
			return "", errors.New("direct root foreign-key ALTER action has multiple FOREIGN clauses")
		}
	}
	root, column, _, err := parseDirectRootForeignKey(action, foreign, len(action))
	if err != nil {
		return "", err
	}
	if !isStorageIdentifier(table) || !isStorageIdentifier(column) {
		return "", errors.New("direct root foreign-key table or column is not canonical storage grammar")
	}
	return root + "|" + table + "|" + column, nil
}

func hasExactlyOneRootReference(action []token) (bool, error) {
	referenceCount := 0
	rootReferenceCount := 0
	for index := 0; index < len(action); index++ {
		if !isKeyword(action[index], "references") {
			continue
		}
		referenceCount++
		_, _, isRoot, err := parseRootReference(action, index+1, len(action))
		if err != nil {
			return false, err
		}
		if isRoot {
			rootReferenceCount++
		}
	}
	if rootReferenceCount == 0 {
		return false, nil
	}
	if rootReferenceCount != 1 || referenceCount != 1 {
		return false, errors.New("direct root foreign-key ALTER action has multiple REFERENCES clauses")
	}
	return true, nil
}

func parseTableRootForeignKeys(table string, body []token) ([]string, error) {
	entries := make([]string, 0)
	start := 0
	depth := 0
	for index := 0; index <= len(body); index++ {
		atEnd := index == len(body)
		if !atEnd {
			switch body[index].text {
			case "(":
				depth++
			case ")":
				depth--
				if depth < 0 {
					return nil, errors.New("CREATE TABLE entry has unbalanced parentheses")
				}
			}
		}
		if !atEnd && (body[index].text != "," || depth != 0) {
			continue
		}
		if index == start {
			return nil, errors.New("CREATE TABLE has an empty entry")
		}
		entry, err := parseRootForeignKeyEntry(table, body[start:index])
		if err != nil {
			return nil, err
		}
		if entry != "" {
			entries = append(entries, entry)
		}
		start = index + 1
	}
	if depth != 0 {
		return nil, errors.New("CREATE TABLE entries have unbalanced parentheses")
	}
	return entries, nil
}

func parseRootForeignKeyEntry(table string, entry []token) (string, error) {
	found := ""
	for reference := 0; reference < len(entry); reference++ {
		if !isKeyword(entry[reference], "references") {
			continue
		}
		root, afterReference, isRoot, err := parseRootReference(entry, reference+1, len(entry))
		if err != nil {
			return "", err
		}
		if !isRoot {
			continue
		}
		foreign := -1
		for index := 0; index < reference; index++ {
			if isKeyword(entry[index], "foreign") {
				if foreign != -1 {
					return "", errors.New("direct root foreign-key entry has multiple FOREIGN clauses")
				}
				foreign = index
			}
		}
		if foreign == -1 {
			return "", errors.New("direct root reference is not a FOREIGN KEY constraint")
		}
		parsedRoot, column, afterForeign, err := parseDirectRootForeignKey(entry, foreign, len(entry))
		if err != nil {
			return "", err
		}
		if parsedRoot != root || afterForeign != afterReference {
			return "", errors.New("direct root foreign-key constraint is malformed or ambiguous")
		}
		if !isStorageIdentifier(table) || !isStorageIdentifier(column) {
			return "", errors.New("direct root foreign-key table or column is not canonical storage grammar")
		}
		if found != "" {
			return "", errors.New("CREATE TABLE entry has multiple direct root foreign-key references")
		}
		found = root + "|" + table + "|" + column
	}
	return found, nil
}

func parseDirectRootForeignKey(tokens []token, foreign, end int) (string, string, int, error) {
	if foreign+1 >= end || !isKeyword(tokens[foreign+1], "key") {
		return "", "", 0, errors.New("FOREIGN must be followed by KEY")
	}
	columns, next, err := parseNameList(tokens, foreign+2, end)
	if err != nil {
		return "", "", 0, fmt.Errorf("parse direct root foreign-key columns: %w", err)
	}
	if len(columns) != 1 {
		return "", "", 0, errors.New("direct root foreign-key must use exactly one source column")
	}
	if next >= end || !isKeyword(tokens[next], "references") {
		return "", "", 0, errors.New("FOREIGN KEY must be followed by REFERENCES")
	}
	root, next, isRoot, err := parseRootReference(tokens, next+1, end)
	if err != nil {
		return "", "", 0, err
	}
	if !isRoot {
		return "", "", 0, errors.New("FOREIGN KEY root target was not recognized")
	}
	return root, columns[0], next, nil
}

func parseRootReference(tokens []token, index, end int) (string, int, bool, error) {
	table, next, err := parseQualifiedName(tokens, index, end)
	if err != nil {
		return "", 0, false, fmt.Errorf("parse REFERENCES table: %w", err)
	}
	if table != "users" && table != "organizations" {
		return "", next, false, nil
	}
	columns, after, err := parseNameList(tokens, next, end)
	if err != nil {
		return "", 0, false, fmt.Errorf("parse direct root reference columns: %w", err)
	}
	if len(columns) != 1 || columns[0] != "id" {
		return "", 0, false, errors.New("direct root reference must target exactly id")
	}
	if table == "users" {
		return "User", after, true, nil
	}
	return "Organization", after, true, nil
}

func parseQualifiedName(tokens []token, index, end int) (string, int, error) {
	if index >= end || tokens[index].kind != tokenName {
		return "", 0, errors.New("qualified name must start with an identifier")
	}
	name := tokens[index].text
	index++
	parts := 1
	for index+1 < end && tokens[index].text == "." {
		if tokens[index+1].kind != tokenName {
			return "", 0, errors.New("qualified name has an invalid identifier")
		}
		parts++
		if parts > 2 {
			return "", 0, errors.New("qualified name has too many parts")
		}
		name = tokens[index+1].text
		index += 2
	}
	return name, index, nil
}

func parseNameList(tokens []token, index, end int) ([]string, int, error) {
	if index >= end || tokens[index].text != "(" {
		return nil, 0, errors.New("column list must start with an opening parenthesis")
	}
	index++
	values := make([]string, 0, 1)
	for {
		if index >= end || tokens[index].kind != tokenName {
			return nil, 0, errors.New("column list requires an identifier")
		}
		values = append(values, tokens[index].text)
		index++
		if index >= end {
			return nil, 0, errors.New("column list is unterminated")
		}
		if tokens[index].text == ")" {
			return values, index + 1, nil
		}
		if tokens[index].text != "," {
			return nil, 0, errors.New("column list requires a comma or closing parenthesis")
		}
		index++
	}
}

func makeProjection(triples []string) (projection, error) {
	sort.Strings(triples)
	for index := range triples {
		if index > 0 && triples[index-1] == triples[index] {
			return projection{}, errors.New("direct root foreign-key projection has a duplicate triple")
		}
	}
	var output strings.Builder
	for _, triple := range triples {
		output.WriteString(triple)
		output.WriteByte('\n')
	}
	canonical := output.String()
	return projection{
		canonical: canonical,
		count:     len(triples),
		bytes:     len(canonical),
		sha256:    sha256Hex([]byte(canonical)),
	}, nil
}

func lexSQL(data []byte) ([]token, error) {
	tokens := make([]token, 0, 1024)
	appendToken := func(kind tokenKind, text string, quoted bool) error {
		if len(tokens) == maxTokenCount {
			return errors.New("inspect SQL exceeds the token bound")
		}
		tokens = append(tokens, token{kind: kind, text: text, quoted: quoted})
		return nil
	}
	for index := 0; index < len(data); {
		if isSpace(data[index]) {
			index++
			continue
		}
		if data[index] == '-' && index+1 < len(data) && data[index+1] == '-' {
			index += 2
			for index < len(data) && data[index] != '\n' {
				index++
			}
			continue
		}
		if data[index] == '/' && index+1 < len(data) && data[index+1] == '*' {
			end := index + 2
			for end+1 < len(data) && (data[end] != '*' || data[end+1] != '/') {
				end++
			}
			if end+1 >= len(data) {
				return nil, errors.New("inspect SQL has an unterminated block comment")
			}
			index = end + 2
			continue
		}
		if data[index] == '"' {
			value, next, err := scanQuotedIdentifier(data, index)
			if err != nil {
				return nil, err
			}
			if err := appendToken(tokenName, value, true); err != nil {
				return nil, err
			}
			index = next
			continue
		}
		if data[index] == '\'' {
			next, err := scanSingleQuotedLiteral(data, index)
			if err != nil {
				return nil, err
			}
			if err := appendToken(tokenLiteral, "literal", false); err != nil {
				return nil, err
			}
			index = next
			continue
		}
		if data[index] == '$' {
			if next, ok := scanDollarQuotedLiteral(data, index); ok {
				if err := appendToken(tokenLiteral, "literal", false); err != nil {
					return nil, err
				}
				index = next
				continue
			}
		}
		if isIdentifierStart(data[index]) {
			start := index
			index++
			for index < len(data) && isIdentifierContinue(data[index]) {
				index++
			}
			if err := appendToken(tokenName, strings.ToLower(string(data[start:index])), false); err != nil {
				return nil, err
			}
			continue
		}
		if err := appendToken(tokenPunctuation, string(data[index]), false); err != nil {
			return nil, err
		}
		index++
	}
	return tokens, nil
}

func scanQuotedIdentifier(data []byte, index int) (string, int, error) {
	var value strings.Builder
	for index++; index < len(data); index++ {
		if data[index] != '"' {
			value.WriteByte(data[index])
			continue
		}
		if index+1 < len(data) && data[index+1] == '"' {
			value.WriteByte('"')
			index++
			continue
		}
		if value.Len() == 0 {
			return "", 0, errors.New("inspect SQL has an empty quoted identifier")
		}
		return value.String(), index + 1, nil
	}
	return "", 0, errors.New("inspect SQL has an unterminated quoted identifier")
}

func scanSingleQuotedLiteral(data []byte, index int) (int, error) {
	for index++; index < len(data); index++ {
		if data[index] == '\\' && index+1 < len(data) {
			index++
			continue
		}
		if data[index] != '\'' {
			continue
		}
		if index+1 < len(data) && data[index+1] == '\'' {
			index++
			continue
		}
		return index + 1, nil
	}
	return 0, errors.New("inspect SQL has an unterminated quoted literal")
}

func scanDollarQuotedLiteral(data []byte, index int) (int, bool) {
	end := index + 1
	if end < len(data) && data[end] == '$' {
		end++
	} else {
		if end >= len(data) || !isIdentifierStart(data[end]) {
			return 0, false
		}
		for end < len(data) && isIdentifierContinue(data[end]) {
			end++
		}
		if end >= len(data) || data[end] != '$' {
			return 0, false
		}
	}
	delimiter := string(data[index : end+1])
	for next := end + 1; next+len(delimiter) <= len(data); next++ {
		if string(data[next:next+len(delimiter)]) == delimiter {
			return next + len(delimiter), true
		}
	}
	return 0, false
}

func isKeyword(token token, keyword string) bool {
	return token.kind == tokenName && !token.quoted && token.text == keyword
}

func isSpace(character byte) bool {
	return character == ' ' || character == '\t' || character == '\n' || character == '\r' || character == '\f'
}

func isIdentifierStart(character byte) bool {
	return (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') || character == '_'
}

func isIdentifierContinue(character byte) bool {
	return isIdentifierStart(character) || (character >= '0' && character <= '9') || character == '$'
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum[:])
}
