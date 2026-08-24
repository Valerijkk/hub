package mail

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"
)

const mailboxDraftMIMEStreamBufferBytes = 32 * 1024

type mailboxDraftMIMEPiece struct {
	text        string
	body        []byte
	sourceIndex int
	source       bool
}

type mailboxDraftMIMEPlan struct {
	draft  EngineDraft
	sources *mailboxDraftAttachmentCapability
	pieces []mailboxDraftMIMEPiece
	size   int64
}

// mailboxDraftMIMEPlanForDraft prepares the complete byte-exact MIME layout
// without opening an object-store stream.
func mailboxDraftMIMEPlanForDraft(ctx context.Context, draft EngineDraft, capability *mailboxDraftAttachmentCapability) (mailboxDraftMIMEPlan, error) {
	if ctx == nil || draft.MaxContentBytes < 1 || draft.MaxContentBytes > MaxEngineMessageBytes || capability == nil && len(draft.Message.Attachments) != 0 || capability != nil && !mailboxDraftAttachmentCapabilityMatches(capability, draft) {
		return mailboxDraftMIMEPlan{}, fakeEngineError(EngineErrorRejectedPolicy)
	}
	base := draft
	base.Message.Attachments = nil
	if err := validateDraftRequest(base); err != nil {
		return mailboxDraftMIMEPlan{}, err
	}
	mixedBoundary := ""
	if capability != nil {
		mixedBoundary = mailboxDraftMIMEBoundary("mixed", draft, capability)
	}
	altBoundary := ""
	if draft.Message.Format == EngineBodyRich {
		altBoundary = mailboxDraftMIMEBoundary("alt", draft, capability)
	}
	headers, err := mailboxDraftMIMERootHeaders(draft, mixedBoundary, altBoundary)
	if err != nil {
		return mailboxDraftMIMEPlan{}, err
	}
	sourceCount := 0
	if capability != nil {
		sourceCount = len(capability.sources)
	}
	plan := mailboxDraftMIMEPlan{draft: draft, sources: capability, pieces: make([]mailboxDraftMIMEPiece, 0, 24+sourceCount*8)}
	plan.addText(headers)
	if capability == nil {
		if draft.Message.Format == EngineBodyRich {
			plan.addText("\r\n")
			plan.addText("--" + altBoundary + "\r\n")
			plan.addText("Content-Type: text/plain; charset=utf-8\r\nContent-Transfer-Encoding: base64\r\n\r\n")
			plan.addBody([]byte(draft.Message.PlainBody))
			plan.addText("--" + altBoundary + "\r\n")
			plan.addText("Content-Type: text/html; charset=utf-8\r\nContent-Transfer-Encoding: base64\r\n\r\n")
			plan.addBody([]byte(draft.Message.RichBody))
			plan.addText("--" + altBoundary + "--\r\n")
		} else {
			plan.addText("Content-Transfer-Encoding: base64\r\n\r\n")
			plan.addBody([]byte(draft.Message.PlainBody))
		}
		if plan.size < 1 || plan.size > draft.MaxContentBytes || plan.size > MaxEngineMessageBytes {
			return mailboxDraftMIMEPlan{}, fakeEngineError(EngineErrorPayloadTooLarge)
		}
		return plan, nil
	}
	if len(capability.sources) == 0 {
		return mailboxDraftMIMEPlan{}, fakeEngineError(EngineErrorRejectedPolicy)
	}

	plan.addText("\r\n")
	plan.addText("--" + mixedBoundary + "\r\n")
	if draft.Message.Format == EngineBodyRich {
		plan.addText("Content-Type: multipart/alternative; boundary=\"" + altBoundary + "\"\r\n\r\n")
		plan.addText("--" + altBoundary + "\r\n")
		plan.addText("Content-Type: text/plain; charset=utf-8\r\nContent-Transfer-Encoding: base64\r\n\r\n")
		plan.addBody([]byte(draft.Message.PlainBody))
		plan.addText("--" + altBoundary + "\r\n")
		plan.addText("Content-Type: text/html; charset=utf-8\r\nContent-Transfer-Encoding: base64\r\n\r\n")
		plan.addBody([]byte(draft.Message.RichBody))
		plan.addText("--" + altBoundary + "--\r\n")
	} else {
		plan.addText("Content-Type: text/plain; charset=utf-8\r\nContent-Transfer-Encoding: base64\r\n\r\n")
		plan.addBody([]byte(draft.Message.PlainBody))
	}
	for index, source := range capability.sources {
		attachmentHeaders, headerErr := mailboxDraftMIMEAttachmentHeaders(source.attachment, mixedBoundary)
		if headerErr != nil {
			return mailboxDraftMIMEPlan{}, headerErr
		}
		plan.addText(attachmentHeaders)
		plan.addSource(index)
	}
	plan.addText("--" + mixedBoundary + "--\r\n")
	if plan.size < 1 || plan.size > draft.MaxContentBytes || plan.size > MaxEngineMessageBytes {
		return mailboxDraftMIMEPlan{}, fakeEngineError(EngineErrorPayloadTooLarge)
	}
	return plan, nil
}

func (plan *mailboxDraftMIMEPlan) addText(value string) {
	plan.pieces = append(plan.pieces, mailboxDraftMIMEPiece{text: value})
	plan.size = mailboxDraftMIMECheckedAdd(plan.size, int64(len(value)))
}

func (plan *mailboxDraftMIMEPlan) addBody(value []byte) {
	plan.pieces = append(plan.pieces, mailboxDraftMIMEPiece{body: append([]byte(nil), value...)})
	plan.size = mailboxDraftMIMECheckedAdd(plan.size, mailboxDraftMIMEBase64Size(int64(len(value))))
}

func (plan *mailboxDraftMIMEPlan) addSource(index int) {
	plan.pieces = append(plan.pieces, mailboxDraftMIMEPiece{sourceIndex: index, source: true})
	plan.size = mailboxDraftMIMECheckedAdd(plan.size, mailboxDraftMIMEBase64Size(plan.sources.sources[index].sizeBytes))
}

func mailboxDraftMIMECheckedAdd(left, right int64) int64 {
	if left < 0 || right < 0 || left > MaxEngineMessageBytes-right {
		return MaxEngineMessageBytes + 1
	}
	return left + right
}

func mailboxDraftMIMEBase64Size(raw int64) int64 {
	if raw == 0 {
		return 0
	}
	encoded := ((raw + 2) / 3) * 4
	return encoded + ((encoded + 75) / 76 * 2)
}

// writeTo emits the planned MIME stream. It never retains attachment bytes.
func (plan mailboxDraftMIMEPlan) writeTo(ctx context.Context, writer io.Writer) error {
	if ctx == nil || writer == nil || plan.sources != nil && !mailboxDraftAttachmentCapabilityMatches(plan.sources, plan.draft) || plan.sources == nil && len(plan.draft.Message.Attachments) != 0 || plan.size < 1 || plan.size > plan.draft.MaxContentBytes || plan.size > MaxEngineMessageBytes {
		return fakeEngineError(EngineErrorRejectedPolicy)
	}
	counting := &mailboxDraftMIMECountingWriter{writer: writer}
	for _, piece := range plan.pieces {
		if err := ctx.Err(); err != nil {
			return err
		}
		if piece.text != "" {
			if _, err := io.WriteString(counting, piece.text); err != nil {
				return err
			}
			continue
		}
		if piece.source {
			if piece.sourceIndex < 0 || piece.sourceIndex >= len(plan.sources.sources) {
				return fakeEngineError(EngineErrorRejectedPolicy)
			}
			if err := mailboxDraftMIMEWriteSource(ctx, counting, plan.sources.sources[piece.sourceIndex]); err != nil {
				return err
			}
			continue
		}
		if err := mailboxDraftMIMEWriteBytes(ctx, counting, piece.body); err != nil {
			return err
		}
	}
	if counting.size != plan.size {
		return fakeEngineError(EngineErrorConflict)
	}
	return nil
}

// mailboxDraftMIMEMaterializeAttachments is intentionally Fake-only support:
// production callers should pass the plan to a streaming provider transport.
func mailboxDraftMIMEMaterializeAttachments(ctx context.Context, plan mailboxDraftMIMEPlan) (EngineDraft, error) {
	if ctx == nil || plan.sources == nil || !mailboxDraftAttachmentCapabilityMatches(plan.sources, plan.draft) {
		return EngineDraft{}, fakeEngineError(EngineErrorRejectedPolicy)
	}
	materialized := plan.draft
	materialized.Message.Attachments = make([]EngineDraftAttachment, len(plan.draft.Message.Attachments))
	for index, source := range plan.sources.sources {
		bytesValue, err := mailboxDraftMIMEReadSource(ctx, source)
		if err != nil {
			return EngineDraft{}, err
		}
		attachment := cloneMailboxDraftAttachment(source.attachment)
		attachment.Bytes = bytesValue
		materialized.Message.Attachments[index] = attachment
	}
	return materialized, nil
}

func mailboxDraftMIMEWriteBytes(ctx context.Context, writer io.Writer, value []byte) error {
	return mailboxDraftMIMEWriteReader(ctx, writer, bytes.NewReader(value), int64(len(value)), [32]byte{}, false)
}

func mailboxDraftMIMEWriteSource(ctx context.Context, writer io.Writer, source *mailboxDraftAttachmentSource) error {
	reader, err := source.openOnce(ctx)
	if err != nil {
		return err
	}
	writeErr := mailboxDraftMIMEWriteReader(ctx, writer, reader, source.sizeBytes, source.attachment.SHA256, true)
	closeErr := reader.Close()
	if writeErr != nil {
		return writeErr
	}
	if closeErr != nil {
		return fakeEngineError(EngineErrorUnavailable)
	}
	return nil
}

func mailboxDraftMIMEReadSource(ctx context.Context, source *mailboxDraftAttachmentSource) ([]byte, error) {
	reader, err := source.openOnce(ctx)
	if err != nil {
		return nil, err
	}
	value := make([]byte, source.sizeBytes)
	readErr := mailboxDraftMIMEReadExact(ctx, reader, value, source.attachment.SHA256)
	closeErr := reader.Close()
	if readErr != nil {
		return nil, readErr
	}
	if closeErr != nil {
		return nil, fakeEngineError(EngineErrorUnavailable)
	}
	return value, nil
}

func mailboxDraftMIMEWriteReader(ctx context.Context, writer io.Writer, reader io.Reader, expected int64, expectedSHA [32]byte, verifySHA bool) error {
	if ctx == nil || writer == nil || reader == nil || expected < 0 {
		return fakeEngineError(EngineErrorRejectedPolicy)
	}
	var buffer [mailboxDraftMIMEStreamBufferBytes]byte
	var line [57]byte
	lineBytes := 0
	remaining := expected
	hash := sha256.New()
	for remaining > 0 {
		if err := ctx.Err(); err != nil {
			return err
		}
		limit := int64(len(buffer))
		if remaining < limit {
			limit = remaining
		}
		n, readErr := reader.Read(buffer[:int(limit)])
		if n > 0 {
			if int64(n) > remaining {
				return fakeEngineError(EngineErrorConflict)
			}
			chunk := buffer[:n]
			_, _ = hash.Write(chunk)
			remaining -= int64(n)
			for len(chunk) > 0 {
				copied := copy(line[lineBytes:], chunk)
				lineBytes += copied
				chunk = chunk[copied:]
				if lineBytes == len(line) {
					if err := mailboxDraftMIMEWriteBase64Line(ctx, writer, line[:]); err != nil {
						return err
					}
					lineBytes = 0
				}
			}
		}
		if readErr != nil {
			if readErr == io.EOF && remaining == 0 {
				break
			}
			return fakeEngineError(EngineErrorUnavailable)
		}
		if n == 0 {
			return fakeEngineError(EngineErrorUnavailable)
		}
	}
	var probe [1]byte
	probeN, probeErr := reader.Read(probe[:])
	if probeN != 0 {
		return fakeEngineError(EngineErrorConflict)
	}
	if probeErr != io.EOF {
		return fakeEngineError(EngineErrorUnavailable)
	}
	if lineBytes > 0 {
		if err := mailboxDraftMIMEWriteBase64Line(ctx, writer, line[:lineBytes]); err != nil {
			return err
		}
	}
	if verifySHA {
		var actual [32]byte
		copy(actual[:], hash.Sum(nil))
		if actual != expectedSHA {
			return fakeEngineError(EngineErrorConflict)
		}
	}
	return nil
}

func mailboxDraftMIMEReadExact(ctx context.Context, reader io.Reader, target []byte, expectedSHA [32]byte) error {
	if ctx == nil || reader == nil || len(target) == 0 {
		return fakeEngineError(EngineErrorRejectedPolicy)
	}
	var buffer [mailboxDraftMIMEStreamBufferBytes]byte
	hash := sha256.New()
	offset := 0
	for offset < len(target) {
		if err := ctx.Err(); err != nil {
			return err
		}
		remaining := len(target) - offset
		limit := len(buffer)
		if remaining < limit {
			limit = remaining
		}
		n, err := reader.Read(buffer[:limit])
		if n > 0 {
			if n > remaining {
				return fakeEngineError(EngineErrorConflict)
			}
			copy(target[offset:offset+n], buffer[:n])
			_, _ = hash.Write(buffer[:n])
			offset += n
		}
		if err != nil {
			if err == io.EOF && offset == len(target) {
				break
			}
			return fakeEngineError(EngineErrorUnavailable)
		}
		if n == 0 {
			return fakeEngineError(EngineErrorUnavailable)
		}
	}
	var probe [1]byte
	n, err := reader.Read(probe[:])
	if n != 0 {
		return fakeEngineError(EngineErrorConflict)
	}
	if err != io.EOF {
		return fakeEngineError(EngineErrorUnavailable)
	}
	var actual [32]byte
	copy(actual[:], hash.Sum(nil))
	if actual != expectedSHA {
		return fakeEngineError(EngineErrorConflict)
	}
	return nil
}

func mailboxDraftMIMEWriteBase64Line(ctx context.Context, writer io.Writer, value []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(value)))
	base64.StdEncoding.Encode(encoded, value)
	if _, err := writer.Write(encoded); err != nil {
		return err
	}
	_, err := io.WriteString(writer, "\r\n")
	return err
}

type mailboxDraftMIMECountingWriter struct {
	writer io.Writer
	size   int64
}

func (writer *mailboxDraftMIMECountingWriter) Write(value []byte) (int, error) {
	n, err := writer.writer.Write(value)
	writer.size += int64(n)
	return n, err
}

func mailboxDraftContentIDValid(value string) bool {
	if value == "" {
		return true
	}
	if len(value) > MaxEngineHeaderValueBytes-len("Content-ID: <>") || strings.Count(value, "@") != 1 || value[0] == '@' || value[len(value)-1] == '@' {
		return false
	}
	for index := 0; index < len(value); index++ {
		character := value[index]
		if character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9' || strings.ContainsRune(".!#$%&'*+-/=?^_`{|}~@", rune(character)) {
			continue
		}
		return false
	}
	return true
}

func mailboxDraftMIMERootHeaders(draft EngineDraft, mixedBoundary, altBoundary string) (string, error) {
	contentType := "text/plain; charset=utf-8"
	if mixedBoundary != "" {
		contentType = "multipart/mixed; boundary=\"" + mixedBoundary + "\""
	} else if draft.Message.Format == EngineBodyRich {
		contentType = "multipart/alternative; boundary=\"" + altBoundary + "\""
	}
	values := []struct{ name, value string }{
		{"From", mailboxDraftMIMEAddress(draft.Message.From)},
		{"To", mailboxDraftMIMEAddresses(draft.Message.To)},
		{"Cc", mailboxDraftMIMEAddresses(draft.Message.CC)},
		{"Bcc", mailboxDraftMIMEAddresses(draft.Message.BCC)},
		{"Reply-To", mailboxDraftMIMEAddresses(draft.Message.ReplyTo)},
		{"Subject", mailboxDraftMIMEEncodedWords(draft.Message.Subject)},
		{"In-Reply-To", draft.Message.InReplyTo},
		{"References", strings.Join(draft.Message.References, " ")},
		{"Importance", string(draft.Message.Priority)},
		{"MIME-Version", "1.0"},
		{"Content-Type", contentType},
	}
	var builder strings.Builder
	for _, value := range values {
		if value.value == "" && (value.name == "To" || value.name == "Cc" || value.name == "Bcc" || value.name == "Reply-To" || value.name == "Subject" || value.name == "In-Reply-To" || value.name == "References") {
			continue
		}
		folded, err := mailboxDraftMIMEFoldHeader(value.name, value.value)
		if err != nil {
			return "", err
		}
		builder.WriteString(folded)
	}
	return builder.String(), nil
}

func mailboxDraftMIMEAttachmentHeaders(attachment EngineDraftAttachment, boundary string) (string, error) {
	if !mailboxDraftContentIDValid(attachment.ContentID) || attachment.Disposition != EngineAttachmentDispositionAttachment && attachment.Disposition != EngineAttachmentDispositionInline {
		return "", fakeEngineError(EngineErrorRejectedPolicy)
	}
	filename, err := mailboxDraftMIMERFC2231Filename(attachment.Filename)
	if err != nil {
		return "", err
	}
	if len(attachment.MediaType) > MaxEngineHeaderValueBytes-len("Content-Type: ") {
		return "", fakeEngineError(EngineErrorRejectedPolicy)
	}
	var builder strings.Builder
	builder.WriteString("--")
	builder.WriteString(boundary)
	builder.WriteString("\r\nContent-Type:")
	if len("Content-Type: ")+len(attachment.MediaType) > 78 {
		builder.WriteString("\r\n ")
	} else {
		builder.WriteString(" ")
	}
	builder.WriteString(attachment.MediaType)
	builder.WriteString(";\r\n ")
	builder.WriteString(strings.ReplaceAll(filename, "; ", ";\r\n "))
	builder.WriteString("\r\nContent-Transfer-Encoding: base64\r\nContent-Disposition: ")
	builder.WriteString(string(attachment.Disposition))
	builder.WriteString(";\r\n ")
	builder.WriteString(strings.ReplaceAll(filename, "; ", ";\r\n "))
	builder.WriteString("\r\n")
	if attachment.ContentID != "" {
		builder.WriteString("Content-ID: <")
		builder.WriteString(attachment.ContentID)
		builder.WriteString(">\r\n")
	}
	builder.WriteString("\r\n")
	return builder.String(), nil
}

func mailboxDraftMIMEAddress(address EngineAddress) string {
	if address.Email == "" {
		return ""
	}
	if address.Name == "" {
		return address.Email
	}
	return mailboxDraftMIMEEncodedWords(address.Name) + " <" + address.Email + ">"
}

func mailboxDraftMIMEAddresses(values []EngineAddress) string {
	addresses := make([]string, 0, len(values))
	for _, address := range values {
		addresses = append(addresses, mailboxDraftMIMEAddress(address))
	}
	return strings.Join(addresses, ", ")
}

func mailboxDraftMIMEEncodedWords(value string) string {
	if value == "" {
		return ""
	}
	chunks := make([]string, 0, 1)
	current := make([]byte, 0, 45)
	for _, runeValue := range value {
		encoded := []byte(string(runeValue))
		if len(current) > 0 && len(current)+len(encoded) > 45 {
			chunks = append(chunks, "=?UTF-8?B?"+base64.StdEncoding.EncodeToString(current)+"?=")
			current = current[:0]
		}
		current = append(current, encoded...)
	}
	if len(current) > 0 {
		chunks = append(chunks, "=?UTF-8?B?"+base64.StdEncoding.EncodeToString(current)+"?=")
	}
	return strings.Join(chunks, " ")
}

func mailboxDraftMIMEFoldHeader(name, value string) (string, error) {
	if name == "" || value == "" || len(name)+1 > 998 || !utf8.ValidString(value) {
		return "", fakeEngineError(EngineErrorRejectedPolicy)
	}
	words := strings.Fields(value)
	if len(words) == 0 {
		return "", fakeEngineError(EngineErrorRejectedPolicy)
	}
	line := name + ":"
	var builder strings.Builder
	for _, word := range words {
		if len(word) > 997 {
			return "", fakeEngineError(EngineErrorRejectedPolicy)
		}
		if len(line)+1+len(word) > 78 {
			builder.WriteString(line)
			builder.WriteString("\r\n")
			line = " " + word
		} else {
			line += " " + word
		}
		if len(line) > 998 {
			return "", fakeEngineError(EngineErrorRejectedPolicy)
		}
	}
	builder.WriteString(line)
	builder.WriteString("\r\n")
	return builder.String(), nil
}

func mailboxDraftMIMERFC2231Filename(value string) (string, error) {
	if !utf8.ValidString(value) || value == "" {
		return "", fakeEngineError(EngineErrorRejectedPolicy)
	}
	units := make([]string, 0, len(value))
	for _, byteValue := range []byte(value) {
		if (byteValue >= 'A' && byteValue <= 'Z') || (byteValue >= 'a' && byteValue <= 'z') || (byteValue >= '0' && byteValue <= '9') || strings.ContainsRune("!#$&+-.^_`|~", rune(byteValue)) {
			units = append(units, string(byteValue))
		} else {
			const hex = "0123456789ABCDEF"
			units = append(units, string([]byte{'%', hex[byteValue>>4], hex[byteValue&0x0f]}))
		}
	}
	parts := make([]string, 0, (len(units)+47)/48)
	var current strings.Builder
	for _, unit := range units {
		if current.Len() > 0 && current.Len()+len(unit) > 48 {
			parts = append(parts, current.String())
			current.Reset()
		}
		current.WriteString(unit)
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	if len(parts) == 0 || len(parts) > 99 {
		return "", fakeEngineError(EngineErrorRejectedPolicy)
	}
	values := make([]string, len(parts))
	for index, part := range parts {
		prefix := "filename*" + strconv.Itoa(index) + "*="
		if index == 0 {
			prefix += "utf-8''"
		}
		values[index] = prefix + part
	}
	return strings.Join(values, "; "), nil
}

func mailboxDraftMIMEBoundary(role string, draft EngineDraft, capability *mailboxDraftAttachmentCapability) string {
	hash := sha256.New()
	mailboxDraftMIMEHashField(hash, "hub-mail-draft-mime-v1")
	mailboxDraftMIMEHashField(hash, role)
	mailboxDraftMIMEHashField(hash, draft.AccountID)
	mailboxDraftMIMEHashField(hash, draft.DraftKey)
	mailboxDraftMIMEHashField(hash, draft.MessageID)
	mailboxDraftMIMEHashField(hash, draft.PreviousVersion)
	mailboxDraftMIMEHashField(hash, draft.MutationID)
	mailboxDraftMIMEHashField(hash, mailboxDraftMIMEAddress(draft.Message.From))
	mailboxDraftMIMEHashField(hash, mailboxDraftMIMEAddresses(draft.Message.To))
	mailboxDraftMIMEHashField(hash, mailboxDraftMIMEAddresses(draft.Message.CC))
	mailboxDraftMIMEHashField(hash, mailboxDraftMIMEAddresses(draft.Message.BCC))
	mailboxDraftMIMEHashField(hash, mailboxDraftMIMEAddresses(draft.Message.ReplyTo))
	mailboxDraftMIMEHashField(hash, draft.Message.Subject)
	mailboxDraftMIMEHashField(hash, draft.Message.InReplyTo)
	for _, reference := range draft.Message.References {
		mailboxDraftMIMEHashField(hash, reference)
	}
	mailboxDraftMIMEHashField(hash, string(draft.Message.Format))
	mailboxDraftMIMEHashField(hash, draft.Message.PlainBody)
	mailboxDraftMIMEHashField(hash, draft.Message.RichBody)
	mailboxDraftMIMEHashField(hash, string(draft.Message.Priority))
	if capability != nil {
		for _, source := range capability.sources {
			mailboxDraftMIMEHashField(hash, source.attachment.AttachmentKey)
			mailboxDraftMIMEHashField(hash, source.attachment.ContentID)
			mailboxDraftMIMEHashField(hash, source.attachment.Filename)
			mailboxDraftMIMEHashField(hash, source.attachment.MediaType)
			mailboxDraftMIMEHashField(hash, string(source.attachment.Disposition))
			mailboxDraftMIMEHashField(hash, string(source.attachment.SHA256[:]))
			mailboxDraftMIMEHashField(hash, strconv.FormatInt(source.sizeBytes, 10))
		}
	}
	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(hash.Sum(nil))
	return "hub-" + role + "-v1-" + strings.ToLower(encoded)
}

func mailboxDraftMIMEHashField(hash io.Writer, value string) {
	_, _ = io.WriteString(hash, strconv.Itoa(len(value)))
	_, _ = io.WriteString(hash, ":")
	_, _ = io.WriteString(hash, value)
	_, _ = io.WriteString(hash, ";")
}
