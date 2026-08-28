package mail_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/valerii/hub/backend/internal/mail"
)

func TestMailboxContent_FrozenPublicABI(t *testing.T) {
	if reflect.TypeOf(mail.MailboxContentFormatPlain).Kind() != reflect.Uint8 ||
		mail.MailboxContentFormatPlain != 1 || mail.MailboxContentFormatRich != 2 {
		t.Fatalf("MailboxContentFormat ABI changed: plain=%d rich=%d", mail.MailboxContentFormatPlain, mail.MailboxContentFormatRich)
	}
	if mail.MailboxContentMaxInputBytes != 8_388_608 ||
		mail.MailboxContentMaxPlainBytes != 8_388_608 ||
		mail.MailboxContentMaxRunes != 2_000_000 ||
		mail.MailboxContentMaxTokens != 100_000 ||
		mail.MailboxContentMaxDepth != 64 ||
		mail.MaxMailboxContentHTMLTokenBytes != 65_536 {
		t.Fatal("mailbox content limits changed")
	}
	assertMailboxContentFields(t, reflect.TypeOf(mail.MailboxContentInput{}), []string{"SourceFormat", "Body"})
	assertMailboxContentFieldTypes(t, reflect.TypeOf(mail.MailboxContentInput{}), []reflect.Type{reflect.TypeOf(mail.MailboxContentFormatPlain), reflect.TypeOf("")})
	assertMailboxContentFields(t, reflect.TypeOf(mail.MailboxContentResult{}), []string{"SourceFormat", "SanitizerVersion", "PlainBody", "SanitizedHTML", "RemoteImagesBlocked"})
	assertMailboxContentFieldTypes(t, reflect.TypeOf(mail.MailboxContentResult{}), []reflect.Type{reflect.TypeOf(mail.MailboxContentFormatPlain), reflect.TypeOf(mail.MailboxContentSanitizerVersionV1), reflect.TypeOf(""), reflect.TypeOf(""), reflect.TypeOf(false)})
	if reflect.TypeOf(mail.MailboxContentErrorKind("")).Kind() != reflect.String ||
		string(mail.Invalid) != "invalid" || string(mail.TooLarge) != "too_large" ||
		mail.ErrMailboxContentInvalid.Error() != "mailbox content invalid" ||
		mail.ErrMailboxContentTooLarge.Error() != "mailbox content exceeds bounds" {
		t.Fatal("mailbox content error ABI changed")
	}
	assertMailboxContentFields(t, reflect.TypeOf(mail.MailboxContentError{}), []string{"Kind"})
	assertMailboxContentFieldTypes(t, reflect.TypeOf(mail.MailboxContentError{}), []reflect.Type{reflect.TypeOf(mail.MailboxContentErrorKind(""))})
	assertMailboxContentExactType(t, reflect.TypeOf(mail.NormalizeMailboxContent), reflect.TypeOf(func(mail.MailboxContentInput) (mail.MailboxContentResult, error) { panic("oracle") }))
	errorType := reflect.TypeOf((*mail.MailboxContentError)(nil))
	assertMailboxContentMethod(t, errorType, "Error", reflect.TypeOf(func(*mail.MailboxContentError) string { panic("oracle") }))
	assertMailboxContentMethod(t, errorType, "Unwrap", reflect.TypeOf(func(*mail.MailboxContentError) error { panic("oracle") }))
}

func TestNormalizeMailboxContent_PlainCanonicalizesAndIsIdempotent(t *testing.T) {
	result := requireMailboxContentSuccess(t, mail.MailboxContentInput{SourceFormat: mail.MailboxContentFormatPlain, Body: " \tA\u00a0\u2003B \r\n C \rD\n\n\nE\u2028F\u2029G"})
	if result.PlainBody != "A B\nC\nD\n\nE\nF\nG" {
		t.Fatalf("plain canonical body = %q", result.PlainBody)
	}
	if got := requireMailboxContentSuccess(t, mail.MailboxContentInput{SourceFormat: mail.MailboxContentFormatPlain, Body: "Cafe\u0301"}).PlainBody; got != "Caf\u00E9" {
		t.Fatalf("NFC plain body = %q", got)
	}
	if got := requireMailboxContentSuccess(t, mail.MailboxContentInput{SourceFormat: mail.MailboxContentFormatPlain, Body: ""}); got.PlainBody != "" {
		t.Fatalf("empty plain body = %q", got.PlainBody)
	}
	second := requireMailboxContentSuccess(t, mail.MailboxContentInput{SourceFormat: mail.MailboxContentFormatPlain, Body: result.PlainBody})
	if second != result {
		t.Fatalf("canonical second pass = %#v, want %#v", second, result)
	}
	allowed := requireMailboxContentSuccess(t, mail.MailboxContentInput{SourceFormat: mail.MailboxContentFormatPlain, Body: "a\tb\nc\rd"})
	if allowed.PlainBody != "a b\nc\nd" {
		t.Fatalf("allowed C0 canonical body = %q", allowed.PlainBody)
	}
}

func TestNormalizeMailboxContent_PlainRejectsRawControlsAndNeverLeaks(t *testing.T) {
	for value := rune(0); value <= 0x1f; value++ {
		if value == '\t' || value == '\n' || value == '\r' {
			continue
		}
		requireMailboxContentFailure(t, mail.MailboxContentInput{SourceFormat: mail.MailboxContentFormatPlain, Body: "secret-control-" + string(value) + "-body"}, mail.Invalid, mail.ErrMailboxContentInvalid)
	}
	for value := rune(0x7f); value <= 0x9f; value++ {
		requireMailboxContentFailure(t, mail.MailboxContentInput{SourceFormat: mail.MailboxContentFormatPlain, Body: "secret-control-" + string(value) + "-body"}, mail.Invalid, mail.ErrMailboxContentInvalid)
	}
	for _, value := range []rune{0x061c, 0x200e, 0x200f, 0x202a, 0x202b, 0x202c, 0x202d, 0x202e, 0x2066, 0x2067, 0x2068, 0x2069} {
		requireMailboxContentFailure(t, mail.MailboxContentInput{SourceFormat: mail.MailboxContentFormatPlain, Body: "secret-bidi-" + string(value) + "-body"}, mail.Invalid, mail.ErrMailboxContentInvalid)
	}
	requireMailboxContentFailure(t, mail.MailboxContentInput{SourceFormat: mail.MailboxContentFormatPlain, Body: string([]byte{'s', 'e', 'c', 0xff, 'r', 'e', 't'})}, mail.Invalid, mail.ErrMailboxContentInvalid)
	for _, format := range []mail.MailboxContentFormat{0, 3, 255} {
		requireMailboxContentFailure(t, mail.MailboxContentInput{SourceFormat: format, Body: "secret-unsupported-format"}, mail.Invalid, mail.ErrMailboxContentInvalid)
	}
}

func TestNormalizeMailboxContent_PreflightAndCanonicalBoundsAreExact(t *testing.T) {
	invalidAtInputLimit := strings.Repeat("x", mail.MailboxContentMaxInputBytes-1) + string([]byte{0xff})
	if len(invalidAtInputLimit) != mail.MailboxContentMaxInputBytes {
		t.Fatalf("invalid input-limit fixture = %d", len(invalidAtInputLimit))
	}
	requireMailboxContentFailure(t, mail.MailboxContentInput{SourceFormat: mail.MailboxContentFormatPlain, Body: invalidAtInputLimit}, mail.Invalid, mail.ErrMailboxContentInvalid)
	tooLarge := invalidAtInputLimit + "x"
	if len(tooLarge) != mail.MailboxContentMaxInputBytes+1 {
		t.Fatalf("too-large input fixture = %d", len(tooLarge))
	}
	requireMailboxContentFailure(t, mail.MailboxContentInput{SourceFormat: mail.MailboxContentFormatPlain, Body: tooLarge}, mail.TooLarge, mail.ErrMailboxContentTooLarge)

	maxRunes := strings.Repeat("a", mail.MailboxContentMaxRunes)
	if utf8.RuneCountInString(maxRunes) != mail.MailboxContentMaxRunes {
		t.Fatalf("max rune fixture = %d", utf8.RuneCountInString(maxRunes))
	}
	if result := requireMailboxContentSuccess(t, mail.MailboxContentInput{SourceFormat: mail.MailboxContentFormatPlain, Body: maxRunes}); utf8.RuneCountInString(result.PlainBody) != mail.MailboxContentMaxRunes {
		t.Fatalf("max rune output = %d", utf8.RuneCountInString(result.PlainBody))
	}
	requireMailboxContentFailure(t, mail.MailboxContentInput{SourceFormat: mail.MailboxContentFormatPlain, Body: maxRunes + "a"}, mail.TooLarge, mail.ErrMailboxContentTooLarge)
	requireMailboxContentFailure(t, mail.MailboxContentInput{SourceFormat: mail.MailboxContentFormatPlain, Body: strings.Repeat("x", mail.MailboxContentMaxInputBytes+1) + string([]byte{0xff})}, mail.TooLarge, mail.ErrMailboxContentTooLarge)
}

func TestNormalizeMailboxContent_RichExtractsPlainTextAndCanonicalizes(t *testing.T) {
	input := `<DIV> alpha <a href="https://secret.example/path" src="https://secret.example/image" srcset="https://secret.example/set" style="display:block" alt="secret-alt"> label </a><br><p>beta &amp; gamma</p><img src="https://remote.example/image">after</DIV><!--secret-comment--><!DOCTYPE html>`
	result := requireMailboxContentSuccess(t, mail.MailboxContentInput{SourceFormat: mail.MailboxContentFormatRich, Body: input})
	if result.PlainBody != "alpha label\n\nbeta & gamma\nafter" {
		t.Fatalf("rich extracted plain body = %q", result.PlainBody)
	}
	for _, forbidden := range []string{"https://", "secret-alt", "secret-comment", "<", ">", "img", "href", "srcset"} {
		if strings.Contains(result.PlainBody, forbidden) {
			t.Fatalf("rich plain body leaked %q: %q", forbidden, result.PlainBody)
		}
	}
	second := requireMailboxContentSuccess(t, mail.MailboxContentInput{SourceFormat: mail.MailboxContentFormatRich, Body: input})
	if second != result {
		t.Fatalf("rich normalization is not deterministic: first=%#v second=%#v", result, second)
	}
	if got := requireMailboxContentSuccess(t, mail.MailboxContentInput{SourceFormat: mail.MailboxContentFormatRich, Body: `<A DATA-X="small">anchor label</A>`}); got.PlainBody != "anchor label" {
		t.Fatalf("small quoted attribute output = %q", got.PlainBody)
	}
	if got := requireMailboxContentSuccess(t, mail.MailboxContentInput{SourceFormat: mail.MailboxContentFormatRich, Body: `before<img src="https://secret.example/x" alt="secret-alt">after`}); got.PlainBody != "beforeafter" || strings.Contains(got.PlainBody, "secret") {
		t.Fatalf("actual image token output = %q", got.PlainBody)
	}
	if got := requireMailboxContentSuccess(t, mail.MailboxContentInput{SourceFormat: mail.MailboxContentFormatRich, Body: `<p>&#0;</p>`}); got.PlainBody != "\uFFFD" {
		t.Fatalf("numeric null output = %q", got.PlainBody)
	}
	for _, body := range []string{`<p>&#x80;</p>`, `<p>&#128;</p>`} {
		if got := requireMailboxContentSuccess(t, mail.MailboxContentInput{SourceFormat: mail.MailboxContentFormatRich, Body: body}); got.PlainBody != "\u20AC" {
			t.Fatalf("numeric C1 %q output = %q", body, got.PlainBody)
		}
	}
	if got := requireMailboxContentSuccess(t, mail.MailboxContentInput{SourceFormat: mail.MailboxContentFormatRich, Body: `<p>&lt;img src="https://secret.example/x" alt="secret-alt"&gt;</p>`}); got.PlainBody != `<img src="https://secret.example/x" alt="secret-alt">` {
		t.Fatalf("escaped markup plain output = %q", got.PlainBody)
	}
	if got := requireMailboxContentSuccess(t, mail.MailboxContentInput{SourceFormat: mail.MailboxContentFormatRich, Body: `before<?secret processing-instruction?>after`}); got.PlainBody != "beforeafter" || strings.Contains(got.PlainBody, "secret") {
		t.Fatalf("processing instruction output = %q", got.PlainBody)
	}
}

func TestNormalizeMailboxContent_RichBlockBreakSetAndSuppressionAreExact(t *testing.T) {
	blockTags := []string{"address", "article", "aside", "blockquote", "div", "dl", "dt", "dd", "fieldset", "figcaption", "figure", "footer", "header", "hgroup", "li", "main", "nav", "ol", "p", "pre", "section", "table", "tbody", "thead", "tfoot", "tr", "td", "th", "ul", "h1", "h2", "h3", "h4", "h5", "h6"}
	for _, tag := range blockTags {
		t.Run(tag, func(t *testing.T) {
			result := requireMailboxContentSuccess(t, mail.MailboxContentInput{SourceFormat: mail.MailboxContentFormatRich, Body: `<span>a</span><` + tag + `>b</` + tag + `><span>c</span>`})
			if result.PlainBody != "a\nb\nc" {
				t.Fatalf("block %q output = %q", tag, result.PlainBody)
			}
		})
	}
	for _, tag := range []string{"br", "hr"} {
		result := requireMailboxContentSuccess(t, mail.MailboxContentInput{SourceFormat: mail.MailboxContentFormatRich, Body: `<span>a</span><` + tag + `><span>c</span>`})
		if result.PlainBody != "a\nc" {
			t.Fatalf("void block %q output = %q", tag, result.PlainBody)
		}
	}
	if got := requireMailboxContentSuccess(t, mail.MailboxContentInput{SourceFormat: mail.MailboxContentFormatRich, Body: `<span>a</span><form><p>x</p></form><span>b</span>`}); got.PlainBody != "ab" {
		t.Fatalf("suppression precedence output = %q", got.PlainBody)
	}

	nonVoidSuppressed := []string{"script", "style", "template", "noscript", "iframe", "frame", "object", "applet", "svg", "math", "form", "picture", "video", "audio", "canvas"}
	for _, tag := range nonVoidSuppressed {
		t.Run("suppressed-"+tag, func(t *testing.T) {
			result := requireMailboxContentSuccess(t, mail.MailboxContentInput{SourceFormat: mail.MailboxContentFormatRich, Body: `before<` + tag + `>secret-<a href="https://secret.example">label</a><p>hidden</p></` + tag + `>after`})
			if result.PlainBody != "beforeafter" || strings.Contains(result.PlainBody, "secret") || strings.Contains(result.PlainBody, "label") || strings.Contains(result.PlainBody, "hidden") {
				t.Fatalf("suppressed %q output = %q", tag, result.PlainBody)
			}
		})
	}
	for _, tag := range []string{"embed", "img", "source", "track"} {
		result := requireMailboxContentSuccess(t, mail.MailboxContentInput{SourceFormat: mail.MailboxContentFormatRich, Body: `before<` + tag + `>after`})
		if result.PlainBody != "beforeafter" {
			t.Fatalf("void suppressed %q output = %q", tag, result.PlainBody)
		}
	}
}

func TestNormalizeMailboxContent_RichStrictStackAndDecodedControlFailures(t *testing.T) {
	for _, tag := range []string{"area", "base", "br", "col", "embed", "hr", "img", "input", "link", "meta", "param", "source", "track", "wbr"} {
		t.Run("void-"+tag, func(t *testing.T) {
			result := requireMailboxContentSuccess(t, mail.MailboxContentInput{SourceFormat: mail.MailboxContentFormatRich, Body: `before<` + tag + `>after`})
			want := "beforeafter"
			if tag == "br" || tag == "hr" {
				want = "before\nafter"
			}
			if result.PlainBody != want {
				t.Fatalf("void %q output = %q, want %q", tag, result.PlainBody, want)
			}
			requireMailboxContentFailure(t, mail.MailboxContentInput{SourceFormat: mail.MailboxContentFormatRich, Body: `<` + tag + `></` + tag + `>`}, mail.Invalid, mail.ErrMailboxContentInvalid)
		})
	}
	if got := requireMailboxContentSuccess(t, mail.MailboxContentInput{SourceFormat: mail.MailboxContentFormatRich, Body: `<span>a</span><span/><img/><span>b</span>`}); got.PlainBody != "ab" {
		t.Fatalf("self-closing stack output = %q", got.PlainBody)
	}
	for _, body := range []string{
		`<br></br>`,
		`<div></p>`,
		`<div>unclosed`,
		`<a>safe&#x202E;secret</a>`,
	} {
		requireMailboxContentFailure(t, mail.MailboxContentInput{SourceFormat: mail.MailboxContentFormatRich, Body: body}, mail.Invalid, mail.ErrMailboxContentInvalid)
	}
	quotedOver := `<a data-x="` + strings.Repeat("x", mail.MaxMailboxContentHTMLTokenBytes)
	requireMailboxContentFailure(t, mail.MailboxContentInput{SourceFormat: mail.MailboxContentFormatRich, Body: quotedOver}, mail.TooLarge, mail.ErrMailboxContentTooLarge)
}

func TestNormalizeMailboxContent_RichTokenAndDepthBoundsAreExact(t *testing.T) {
	exactComment := "<!--" + strings.Repeat("x", mail.MaxMailboxContentHTMLTokenBytes-7) + "-->"
	if len(exactComment) != mail.MaxMailboxContentHTMLTokenBytes {
		t.Fatalf("exact raw comment token length = %d", len(exactComment))
	}
	if got := requireMailboxContentSuccess(t, mail.MailboxContentInput{SourceFormat: mail.MailboxContentFormatRich, Body: exactComment}); got.PlainBody != "" {
		t.Fatalf("exact raw comment output = %q", got.PlainBody)
	}
	overComment := "<!--" + strings.Repeat("x", mail.MaxMailboxContentHTMLTokenBytes-6) + "-->"
	if len(overComment) != mail.MaxMailboxContentHTMLTokenBytes+1 {
		t.Fatalf("over raw comment token length = %d", len(overComment))
	}
	requireMailboxContentFailure(t, mail.MailboxContentInput{SourceFormat: mail.MailboxContentFormatRich, Body: overComment}, mail.TooLarge, mail.ErrMailboxContentTooLarge)

	exactTokens := strings.Repeat("<!---->", mail.MailboxContentMaxTokens)
	if got := requireMailboxContentSuccess(t, mail.MailboxContentInput{SourceFormat: mail.MailboxContentFormatRich, Body: exactTokens}); got.PlainBody != "" {
		t.Fatalf("exact token count output = %q", got.PlainBody)
	}
	requireMailboxContentFailure(t, mail.MailboxContentInput{SourceFormat: mail.MailboxContentFormatRich, Body: strings.Repeat("<!---->", mail.MailboxContentMaxTokens+1)}, mail.TooLarge, mail.ErrMailboxContentTooLarge)

	exactDepth := strings.Repeat("<div>", mail.MailboxContentMaxDepth) + "x" + strings.Repeat("</div>", mail.MailboxContentMaxDepth)
	if got := requireMailboxContentSuccess(t, mail.MailboxContentInput{SourceFormat: mail.MailboxContentFormatRich, Body: exactDepth}); got.PlainBody != "x" {
		t.Fatalf("exact depth output = %q", got.PlainBody)
	}
	overDepth := strings.Repeat("<div>", mail.MailboxContentMaxDepth+1) + "x" + strings.Repeat("</div>", mail.MailboxContentMaxDepth+1)
	requireMailboxContentFailure(t, mail.MailboxContentInput{SourceFormat: mail.MailboxContentFormatRich, Body: overDepth}, mail.TooLarge, mail.ErrMailboxContentTooLarge)
}

func TestMailboxContentError_TypedSentinelsNilUnknownAndNoLeak(t *testing.T) {
	for _, assertion := range []struct {
		kind     mail.MailboxContentErrorKind
		sentinel error
		text     string
	}{
		{kind: mail.Invalid, sentinel: mail.ErrMailboxContentInvalid, text: "mailbox content invalid"},
		{kind: mail.TooLarge, sentinel: mail.ErrMailboxContentTooLarge, text: "mailbox content exceeds bounds"},
	} {
		err := &mail.MailboxContentError{Kind: assertion.kind}
		if err.Error() != assertion.text || err.Unwrap() != assertion.sentinel || !errors.Is(err, assertion.sentinel) {
			t.Fatalf("known error %q = error:%q unwrap:%v", assertion.kind, err.Error(), err.Unwrap())
		}
		other := mail.ErrMailboxContentInvalid
		if assertion.sentinel == other {
			other = mail.ErrMailboxContentTooLarge
		}
		if errors.Is(err, other) {
			t.Fatalf("known error %q matched unrelated sentinel", assertion.kind)
		}
		var typed *mail.MailboxContentError
		if !errors.As(err, &typed) || typed.Kind != assertion.kind {
			t.Fatalf("known error %q errors.As = %#v", assertion.kind, typed)
		}
	}
	unknown := &mail.MailboxContentError{Kind: "unknown-secret-body"}
	if unknown.Error() != "mailbox content error" || unknown.Unwrap() != nil || errors.Unwrap(unknown) != nil {
		t.Fatalf("unknown error is unsafe: error:%q unwrap:%v", unknown.Error(), unknown.Unwrap())
	}
	var nilError *mail.MailboxContentError
	if nilError.Error() != "mailbox content error" || nilError.Unwrap() != nil {
		t.Fatalf("nil error is unsafe: error:%q unwrap:%v", nilError.Error(), nilError.Unwrap())
	}
}

func requireMailboxContentSuccess(t *testing.T, input mail.MailboxContentInput) mail.MailboxContentResult {
	t.Helper()
	result, err := mail.NormalizeMailboxContent(input)
	if err != nil {
		t.Fatalf("NormalizeMailboxContent(%#v) error = %v", input, err)
	}
	if result.SourceFormat != input.SourceFormat ||
		result.SanitizerVersion != mail.MailboxContentSanitizerVersionV1 ||
		!result.RemoteImagesBlocked ||
		(input.SourceFormat == mail.MailboxContentFormatPlain && result.SanitizedHTML != "") {
		t.Fatalf("NormalizeMailboxContent(%#v) result ABI = %#v", input, result)
	}
	return result
}

func requireMailboxContentFailure(t *testing.T, input mail.MailboxContentInput, wantKind mail.MailboxContentErrorKind, wantSentinel error) {
	t.Helper()
	result, err := mail.NormalizeMailboxContent(input)
	if result != (mail.MailboxContentResult{}) {
		t.Fatalf("NormalizeMailboxContent(%#v) result = %#v, want zero", input, result)
	}
	if !errors.Is(err, wantSentinel) {
		t.Fatalf("NormalizeMailboxContent(%#v) error = %v, want %v", input, err, wantSentinel)
	}
	var typed *mail.MailboxContentError
	if !errors.As(err, &typed) || typed.Kind != wantKind {
		t.Fatalf("NormalizeMailboxContent(%#v) typed error = %#v, want %q", input, typed, wantKind)
	}
	if err == nil || strings.Contains(err.Error(), input.Body) || strings.Contains(err.Error(), "secret-") {
		t.Fatalf("NormalizeMailboxContent(%#v) leaked body through error = %v", input, err)
	}
	if unwrapped := errors.Unwrap(err); unwrapped != nil && (strings.Contains(unwrapped.Error(), input.Body) || strings.Contains(unwrapped.Error(), "secret-")) {
		t.Fatalf("NormalizeMailboxContent(%#v) leaked body through unwrap = %v", input, unwrapped)
	}
}

func assertMailboxContentFields(t *testing.T, valueType reflect.Type, want []string) {
	t.Helper()
	if valueType.NumField() != len(want) {
		t.Fatalf("%s field count = %d, want %d", valueType.Name(), valueType.NumField(), len(want))
	}
	for index, name := range want {
		field := valueType.Field(index)
		if field.Name != name || !field.IsExported() {
			t.Fatalf("%s field %d = %#v, want exported %q", valueType.Name(), index, field, name)
		}
	}
}

func assertMailboxContentFieldTypes(t *testing.T, valueType reflect.Type, want []reflect.Type) {
	t.Helper()
	if valueType.NumField() != len(want) {
		t.Fatalf("%s field count = %d, want %d", valueType.Name(), valueType.NumField(), len(want))
	}
	for index, wantType := range want {
		field := valueType.Field(index)
		if field.Type != wantType {
			t.Fatalf("%s field %d type = %s, want %s", valueType.Name(), index, field.Type, wantType)
		}
	}
}

func assertMailboxContentExactType(t *testing.T, got, want reflect.Type) {
	t.Helper()
	if got != want {
		t.Fatalf("signature = %s, want %s", got, want)
	}
}

func assertMailboxContentMethod(t *testing.T, valueType reflect.Type, name string, want reflect.Type) {
	t.Helper()
	method, found := valueType.MethodByName(name)
	if !found || method.Type != want {
		t.Fatalf("%s.%s signature = %v, want %s", valueType, name, method.Type, want)
	}
}
