package mail

import (
	"errors"
	"io"
	"strings"
	"unicode/utf8"

	"golang.org/x/net/html"
	"golang.org/x/text/unicode/norm"
)

type MailboxContentFormat uint8

const (
	MailboxContentFormatPlain MailboxContentFormat = iota + 1
	MailboxContentFormatRich
)

type MailboxContentSanitizerVersion uint16

const MailboxContentSanitizerVersionV1 MailboxContentSanitizerVersion = 1

const (
	MailboxContentMaxInputBytes         = 8_388_608
	MailboxContentMaxPlainBytes         = 8_388_608
	MailboxContentMaxSanitizedHTMLBytes = 8_388_608
	MailboxContentMaxRunes              = 2_000_000
	MailboxContentMaxTokens             = 100_000
	MailboxContentMaxDepth              = 64
	// Maximum accepted raw tokenizer-token bytes, including delimiters, attributes, and text.
	MaxMailboxContentHTMLTokenBytes = 65_536
)

type MailboxContentInput struct {
	SourceFormat MailboxContentFormat
	Body         string
}

type MailboxContentResult struct {
	SourceFormat        MailboxContentFormat
	SanitizerVersion    MailboxContentSanitizerVersion
	PlainBody           string
	SanitizedHTML       string
	RemoteImagesBlocked bool
}

type MailboxContentErrorKind string

const (
	Invalid  MailboxContentErrorKind = "invalid"
	TooLarge MailboxContentErrorKind = "too_large"
)

var (
	ErrMailboxContentInvalid  = errors.New("mailbox content invalid")
	ErrMailboxContentTooLarge = errors.New("mailbox content exceeds bounds")
)

type MailboxContentError struct {
	Kind MailboxContentErrorKind
}

func (e *MailboxContentError) Error() string {
	if e == nil {
		return "mailbox content error"
	}
	switch e.Kind {
	case Invalid:
		return ErrMailboxContentInvalid.Error()
	case TooLarge:
		return ErrMailboxContentTooLarge.Error()
	default:
		return "mailbox content error"
	}
}

func (e *MailboxContentError) Unwrap() error {
	if e == nil {
		return nil
	}
	switch e.Kind {
	case Invalid:
		return ErrMailboxContentInvalid
	case TooLarge:
		return ErrMailboxContentTooLarge
	default:
		return nil
	}
}

// NormalizeMailboxContent deterministically reduces untrusted plain or rich
// body input to canonical plain text and, for rich input only, safe HTML.
// It intentionally returns no URL, image intent, proxy token, or input-derived
// error detail.
func NormalizeMailboxContent(input MailboxContentInput) (MailboxContentResult, error) {
	if err := mailboxContentPreflight(input.Body); err != nil {
		return MailboxContentResult{}, err
	}

	var retainedText, sanitizedHTML string
	var err error
	switch input.SourceFormat {
	case MailboxContentFormatPlain:
		retainedText = input.Body
	case MailboxContentFormatRich:
		retainedText, sanitizedHTML, err = mailboxContentSanitizeRich(input.Body)
		if err != nil {
			return MailboxContentResult{}, err
		}
	default:
		return MailboxContentResult{}, mailboxContentError(Invalid)
	}

	plain, err := mailboxContentCanonicalize(retainedText)
	if err != nil {
		return MailboxContentResult{}, err
	}
	return MailboxContentResult{
		SourceFormat:        input.SourceFormat,
		SanitizerVersion:    MailboxContentSanitizerVersionV1,
		PlainBody:           plain,
		SanitizedHTML:       sanitizedHTML,
		RemoteImagesBlocked: true,
	}, nil
}

func mailboxContentPreflight(body string) error {
	if len(body) > MailboxContentMaxInputBytes {
		return mailboxContentError(TooLarge)
	}
	if !utf8.ValidString(body) {
		return mailboxContentError(Invalid)
	}
	if utf8.RuneCountInString(body) > MailboxContentMaxRunes {
		return mailboxContentError(TooLarge)
	}
	return mailboxContentRejectForbiddenRunes(body)
}

func mailboxContentSanitizeRich(body string) (string, string, error) {
	tokenizer := html.NewTokenizer(strings.NewReader(body))
	// x/net fails when rawLen >= maxBuf. A 65_536-byte token must pass, so
	// the tokenizer buffer is exactly 65_537 bytes before its first Next call.
	tokenizer.SetMaxBuf(MaxMailboxContentHTMLTokenBytes + 1)

	var retainedText, sanitizedHTML strings.Builder
	stack := make([]mailboxContentHTMLNode, 0, 8)
	tokenCount := 0
	for {
		tokenType := tokenizer.Next()
		if tokenType == html.ErrorToken {
			err := tokenizer.Err()
			switch {
			case err == io.EOF && len(stack) == 0:
				return retainedText.String(), sanitizedHTML.String(), nil
			case errors.Is(err, html.ErrBufferExceeded):
				return "", "", mailboxContentError(TooLarge)
			default:
				return "", "", mailboxContentError(Invalid)
			}
		}
		tokenCount++
		if tokenCount > MailboxContentMaxTokens {
			return "", "", mailboxContentError(TooLarge)
		}

		suppressed := len(stack) != 0 && stack[len(stack)-1].suppressed
		switch tokenType {
		case html.TextToken:
			if suppressed {
				continue
			}
			text, err := mailboxContentNormalizeRichText(string(tokenizer.Text()))
			if err != nil {
				return "", "", err
			}
			if err := mailboxContentAppend(&retainedText, text); err != nil {
				return "", "", err
			}
			if err := mailboxContentAppendEscapedHTML(&sanitizedHTML, text); err != nil {
				return "", "", err
			}
		case html.StartTagToken:
			rawTag, _ := tokenizer.TagName()
			tag := mailboxContentASCIILower(rawTag)
			if mailboxContentVoidTag(tag) {
				if !suppressed && !mailboxContentSuppressionRoot(tag) && mailboxContentBlockTag(tag) {
					if err := mailboxContentAppend(&retainedText, "\n"); err != nil {
						return "", "", err
					}
				}
				if !suppressed && mailboxContentAllowedTag(tag) {
					if err := mailboxContentAppendVoidTag(&sanitizedHTML, tag); err != nil {
						return "", "", err
					}
				}
				continue
			}
			if len(stack) >= MailboxContentMaxDepth {
				return "", "", mailboxContentError(TooLarge)
			}
			node := mailboxContentHTMLNode{
				tag:        tag,
				suppressed: suppressed || mailboxContentSuppressionRoot(tag),
				serialized: !suppressed && mailboxContentAllowedTag(tag),
			}
			if !node.suppressed && mailboxContentBlockTag(tag) {
				if err := mailboxContentAppend(&retainedText, "\n"); err != nil {
					return "", "", err
				}
			}
			if node.serialized {
				if err := mailboxContentAppendOpenTag(&sanitizedHTML, tag); err != nil {
					return "", "", err
				}
			}
			stack = append(stack, node)
		case html.EndTagToken:
			rawTag, _ := tokenizer.TagName()
			tag := mailboxContentASCIILower(rawTag)
			if mailboxContentVoidTag(tag) || len(stack) == 0 || stack[len(stack)-1].tag != tag {
				return "", "", mailboxContentError(Invalid)
			}
			node := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if !node.suppressed && mailboxContentBlockTag(tag) {
				if err := mailboxContentAppend(&retainedText, "\n"); err != nil {
					return "", "", err
				}
			}
			if node.serialized {
				if err := mailboxContentAppendCloseTag(&sanitizedHTML, tag); err != nil {
					return "", "", err
				}
			}
		case html.SelfClosingTagToken:
			rawTag, _ := tokenizer.TagName()
			tag := mailboxContentASCIILower(rawTag)
			if !suppressed && !mailboxContentSuppressionRoot(tag) && mailboxContentBlockTag(tag) {
				if err := mailboxContentAppend(&retainedText, "\n"); err != nil {
					return "", "", err
				}
			}
			if !suppressed && !mailboxContentSuppressionRoot(tag) && mailboxContentAllowedTag(tag) {
				if mailboxContentVoidTag(tag) {
					if err := mailboxContentAppendVoidTag(&sanitizedHTML, tag); err != nil {
						return "", "", err
					}
				} else if err := mailboxContentAppendSelfClosingTag(&sanitizedHTML, tag); err != nil {
					return "", "", err
				}
			}
		case html.CommentToken, html.DoctypeToken:
			// These tokens count against the bounded tokenizer budget but emit nothing.
		}
	}
}

func mailboxContentNormalizeRichText(value string) (string, error) {
	value = norm.NFC.String(value)
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	value = strings.ReplaceAll(value, "\u2028", "\n")
	value = strings.ReplaceAll(value, "\u2029", "\n")
	if err := mailboxContentRejectForbiddenRunes(value); err != nil {
		return "", err
	}
	return value, nil
}

func mailboxContentCanonicalize(value string) (string, error) {
	value = norm.NFC.String(value)
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	value = strings.ReplaceAll(value, "\u2028", "\n")
	value = strings.ReplaceAll(value, "\u2029", "\n")

	var whitespaceNormalized strings.Builder
	for _, character := range value {
		if mailboxContentHorizontalSpace(character) {
			whitespaceNormalized.WriteByte(' ')
			continue
		}
		whitespaceNormalized.WriteRune(character)
	}

	lines := strings.Split(whitespaceNormalized.String(), "\n")
	var trimmed strings.Builder
	for index, line := range lines {
		if index != 0 {
			trimmed.WriteByte('\n')
		}
		trimmed.WriteString(strings.Trim(line, " "))
	}

	var collapsed strings.Builder
	newlines := 0
	for _, character := range trimmed.String() {
		if character == '\n' {
			newlines++
			if newlines <= 2 {
				collapsed.WriteByte('\n')
			}
			continue
		}
		newlines = 0
		collapsed.WriteRune(character)
	}
	result := strings.Trim(collapsed.String(), "\n")
	if len(result) > MailboxContentMaxPlainBytes || utf8.RuneCountInString(result) > MailboxContentMaxRunes {
		return "", mailboxContentError(TooLarge)
	}
	return result, nil
}

func mailboxContentAppend(builder *strings.Builder, value string) error {
	if len(value) > MailboxContentMaxPlainBytes-builder.Len() {
		return mailboxContentError(TooLarge)
	}
	builder.WriteString(value)
	return nil
}

func mailboxContentAppendSanitizedHTML(builder *strings.Builder, value string) error {
	if len(value) > MailboxContentMaxSanitizedHTMLBytes-builder.Len() {
		return mailboxContentError(TooLarge)
	}
	builder.WriteString(value)
	return nil
}

func mailboxContentAppendEscapedHTML(builder *strings.Builder, value string) error {
	start := 0
	for index, character := range value {
		var replacement string
		switch character {
		case '&':
			replacement = "&amp;"
		case '\'':
			replacement = "&#39;"
		case '<':
			replacement = "&lt;"
		case '>':
			replacement = "&gt;"
		case '"':
			replacement = "&#34;"
		default:
			continue
		}
		if err := mailboxContentAppendSanitizedHTML(builder, value[start:index]); err != nil {
			return err
		}
		if err := mailboxContentAppendSanitizedHTML(builder, replacement); err != nil {
			return err
		}
		start = index + utf8.RuneLen(character)
	}
	return mailboxContentAppendSanitizedHTML(builder, value[start:])
}

func mailboxContentAppendOpenTag(builder *strings.Builder, tag string) error {
	if err := mailboxContentAppendSanitizedHTML(builder, "<"); err != nil {
		return err
	}
	if err := mailboxContentAppendSanitizedHTML(builder, tag); err != nil {
		return err
	}
	return mailboxContentAppendSanitizedHTML(builder, ">")
}

func mailboxContentAppendCloseTag(builder *strings.Builder, tag string) error {
	if err := mailboxContentAppendSanitizedHTML(builder, "</"); err != nil {
		return err
	}
	if err := mailboxContentAppendSanitizedHTML(builder, tag); err != nil {
		return err
	}
	return mailboxContentAppendSanitizedHTML(builder, ">")
}

func mailboxContentAppendVoidTag(builder *strings.Builder, tag string) error {
	return mailboxContentAppendOpenTag(builder, tag)
}

func mailboxContentAppendSelfClosingTag(builder *strings.Builder, tag string) error {
	if err := mailboxContentAppendOpenTag(builder, tag); err != nil {
		return err
	}
	return mailboxContentAppendCloseTag(builder, tag)
}

func mailboxContentRejectForbiddenRunes(value string) error {
	for _, character := range value {
		if character <= 0x1f && character != '\t' && character != '\n' && character != '\r' ||
			character >= 0x7f && character <= 0x9f ||
			character == 0x061c || character == 0x200e || character == 0x200f ||
			character >= 0x202a && character <= 0x202e ||
			character >= 0x2066 && character <= 0x2069 {
			return mailboxContentError(Invalid)
		}
	}
	return nil
}

func mailboxContentHorizontalSpace(character rune) bool {
	return character == '\t' || character == ' ' || character == 0x00a0 || character == 0x1680 ||
		character >= 0x2000 && character <= 0x200a || character == 0x202f || character == 0x205f || character == 0x3000
}

func mailboxContentASCIILower(value []byte) string {
	result := make([]byte, len(value))
	for index, character := range value {
		if character >= 'A' && character <= 'Z' {
			result[index] = character + ('a' - 'A')
			continue
		}
		result[index] = character
	}
	return string(result)
}

func mailboxContentError(kind MailboxContentErrorKind) error {
	return &MailboxContentError{Kind: kind}
}

type mailboxContentHTMLNode struct {
	tag        string
	suppressed bool
	serialized bool
}

func mailboxContentVoidTag(tag string) bool {
	switch tag {
	case "area", "base", "br", "col", "embed", "hr", "img", "input", "link", "meta", "param", "source", "track", "wbr":
		return true
	default:
		return false
	}
}

func mailboxContentSuppressionRoot(tag string) bool {
	switch tag {
	case "script", "style", "template", "noscript", "iframe", "frame", "object", "embed", "applet", "svg", "math", "form", "img", "picture", "source", "video", "audio", "track", "canvas":
		return true
	default:
		return false
	}
}

func mailboxContentAllowedTag(tag string) bool {
	switch tag {
	case "p", "br", "hr", "strong", "b", "em", "i", "u", "s", "del", "code", "pre", "blockquote", "ul", "ol", "li", "h1", "h2", "h3", "h4", "h5", "h6", "div", "span":
		return true
	default:
		return false
	}
}

func mailboxContentBlockTag(tag string) bool {
	switch tag {
	case "address", "article", "aside", "blockquote", "br", "div", "dl", "dt", "dd", "fieldset", "figcaption", "figure", "footer", "h1", "h2", "h3", "h4", "h5", "h6", "header", "hgroup", "hr", "li", "main", "nav", "ol", "p", "pre", "section", "table", "tbody", "thead", "tfoot", "tr", "td", "th", "ul":
		return true
	default:
		return false
	}
}

func mailboxContentResultValid(result MailboxContentResult) bool {
	if result.SanitizerVersion != MailboxContentSanitizerVersionV1 || !result.RemoteImagesBlocked ||
		len(result.PlainBody) > MailboxContentMaxPlainBytes || utf8.RuneCountInString(result.PlainBody) > MailboxContentMaxRunes ||
		!utf8.ValidString(result.PlainBody) || strings.ContainsRune(result.PlainBody, '\x00') ||
		len(result.SanitizedHTML) > MailboxContentMaxSanitizedHTMLBytes || !utf8.ValidString(result.SanitizedHTML) || strings.ContainsRune(result.SanitizedHTML, '\x00') {
		return false
	}
	switch result.SourceFormat {
	case MailboxContentFormatPlain:
		return result.SanitizedHTML == ""
	case MailboxContentFormatRich:
		return true
	default:
		return false
	}
}
