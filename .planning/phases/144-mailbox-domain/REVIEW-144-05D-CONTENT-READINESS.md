# 144-05D Content — plain-only source-authority readiness

> **Acceptance lifecycle:** `ACCEPTANCE_SYNC_CANDIDATE / ATOMIC_EXACT8_COMMIT_REQUIRED / IMMUTABLE_REVIEW_REQUIRED / RUNTIME_NOT_RUN`. This dirty docs-only exact8 acceptance sync is non-self-activating and grants no current source, test, schema, persistence, `go.mod`, Ent, API/UI, testframework, proxy, sanitizer dependency, or runtime authority. Before its activation the lower 05D authority/precommit/current-next routes remain current.

## Exact8 acceptance-sync manifest

The only acceptance-sync documentation manifest is ordered as follows; the acceptance commit diff MUST
be set-equal to these eight paths, with no additions, omissions, renames, source, schema,
OpenAPI, persistence, `go.mod`, testframework, Ent, API/UI, or runtime files.

1. `.planning/STATE.md`
2. `.planning/decisions/specs/MAIL-SPEC.md`
3. `.planning/phases/144-mailbox-domain/PLAN.md`
4. `.planning/phases/144-mailbox-domain/DISPATCH-INDEX.md`
5. `.planning/phases/144-mailbox-domain/DISPATCH-REMAINING.md`
6. `.planning/phases/144-mailbox-domain/REVIEW-144-05D-CONTENT-READINESS.md`
7. `docs/mail/CONTRACT-CHANGELOG.md`
8. `docs/mail/QA-HANDOFF.md`

This receipt records immutable candidate `d77fbf90b1b3a93dec488fa70094a5713c9e1e85` (sole parent
`00dabf6beb77ab5e37fe156d1b50de9f45e2c75f`, tree `51964d4e81d713f527bd717c49e14d0a14ed18a7`,
subject `docs(mail): propose plain content authority`) with ordered blobs: STATE
`24f7de6ac4a138bf6d97df429433a499460c1e25`; MAIL-SPEC
`ea5b04397ab112939fc559f2be7b210cb6090cb8`; PLAN
`5dec55604aee67d19561bb58083a191325569491`; INDEX
`621a058bc12b3f1f089c41f4dc5cd2d8089ddc45`; REMAINING
`28a457bfb7885473dfc544a3fd6da9cc9f4d4d41`; REVIEW
`6f38851f206d37ad1e3d5e7c64ca0521cc36aa3f`; CHANGELOG
`1d14fba90d39960c08e3bbb950f78e2fc0dfd908`; QA
`a6b6bce55375cd894a6c1c818a9ec275b6eadf7f`.

Candidate Contract, Security, and Lifecycle precommit plus immutable exact-SHA reviews PASS
C/I/M `0/0/0`; provenance is Contract Terra/high, Security Terra/xhigh, Lifecycle Terra/high,
dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Only after one atomic manifest-equal exact8
acceptance commit and its immutable acceptance Contract/Security/Lifecycle triple PASS is the
candidate `SOURCE_AUTHORITY_ACCEPTED / EXACT_SHA_PASS`. Then, and only then, lower 05D
authority/precommit/current-next routes are `CONSUMED/HISTORICAL` and root may dispatch exactly
the following authored-first source packet, and nothing else:

1. NEW `backend/tests/mail/mailbox_content_test.go`
2. NEW `backend/internal/mail/mailbox_content.go`

That exact2 source remains `UNIMPLEMENTED / UNCOMMITTED / UNACCEPTED` until its own tests,
review, atomic source commit, immutable exact-SHA review, and status receipt. This acceptance sync
selects no next Mail packet.

## Frozen public-in-package ABI

`mailbox_content.go` owns only the pure plain-content normalization ABI. It imports no sanitizer
library and starts no I/O. The source packet MUST NOT edit `backend/go.mod`: `golang.org/x/net`
`v0.51.0` is already direct and `golang.org/x/text` `v0.35.0` is already present.

```go
type MailboxContentFormat uint8

const (
	MailboxContentFormatPlain MailboxContentFormat = iota + 1
	MailboxContentFormatRich
)

const (
	MailboxContentMaxInputBytes = 8_388_608
	MailboxContentMaxPlainBytes = 8_388_608
	MailboxContentMaxRunes      = 2_000_000
	MailboxContentMaxTokens     = 100_000
	MailboxContentMaxDepth      = 64
	// Maximum accepted raw tokenizer-token bytes, including delimiters, attributes, and text.
	MaxMailboxContentHTMLTokenBytes = 65_536
)

type MailboxContentInput struct {
	SourceFormat MailboxContentFormat
	Body         string
}

type MailboxContentResult struct {
	Format              MailboxContentFormat
	PlainBody           string
	RemoteImagesBlocked bool
}

func NormalizeMailboxContent(input MailboxContentInput) (MailboxContentResult, error)

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

func (e *MailboxContentError) Error() string
func (e *MailboxContentError) Unwrap() error
```

There are exactly two typed failure classes. `Invalid` is literal `invalid` and `TooLarge` is
literal `too_large`; the two sentinels' Error text is
exactly `mailbox content invalid` and `mailbox content exceeds bounds`. For a non-nil known kind,
the pointer receiver `Error()` returns the matching sentinel text and pointer receiver `Unwrap()`
returns the matching sentinel. A nil receiver or unknown kind has Error text exactly
`mailbox content error` and an `Unwrap()` result of nil. `errors.As` distinguishes the typed error
and `errors.Is` reaches only the matching sentinel. Error or unwrap chains never expose body
content, normalized text, HTML, attributes, URLs, image source, token, account, or message
identity. Every failure returns the zero `MailboxContentResult`.

The ABI has no rich output, sanitized HTML, image intent, image URL, token, proxy request, callback,
clock, context, account, message, or storage argument. Successful results always have
`Format == MailboxContentFormatPlain` and `RemoteImagesBlocked == true`.

## Normalization behavior

### Mandatory preflight and shared canonicalizer

Both formats use this exact order before any format-specific allocation: first reject an input
byte length above `8_388_608`, before UTF-8/rune scanning or allocation; then require valid UTF-8;
then count raw decoded runes at most `2_000_000`; then reject forbidden raw runes. The only allowed
C0 runes are HT `U+0009`, LF `U+000A`, and CR `U+000D`; every other `U+0000..U+001F`, DEL/C1
`U+007F..U+009F`, and exactly `U+061C`, `U+200E`, `U+200F`, `U+202A..U+202E`,
`U+2066..U+2069` is `invalid`. Empty valid input is allowed.

The shared canonicalizer is exact and runs in this order for plain input and the fully decoded rich
text: NFC using existing `x/text/unicode/norm`; CRLF first, then CR, LF, `U+2028`, and `U+2029` to
one LF; every run of HT, `U+0020`, `U+00A0`, `U+1680`, `U+2000..U+200A`, `U+202F`, `U+205F`, or
`U+3000` to one ASCII space; trim ASCII spaces at each line edge; collapse three-or-more LFs to
two; trim outer LFs; finally enforce canonical output at most `8_388_608` bytes and
`2_000_000` runes. Each exact cap passes and the plus-one value is `too_large`; no partial result
is returned. The decoded text emitted by rich extraction re-runs the control/bidi rejection before
it can enter this canonicalizer, so an entity-encoded forbidden control is `invalid`.

### Rich input

Rich input is parsed only with the existing `golang.org/x/net/html` tokenizer. No sanitizer,
renderer, CSS parser, URL parser, dependency, `go.mod` mutation, raw HTML output, or proxy token
is permitted. `MaxMailboxContentHTMLTokenBytes = 65_536` is the maximum accepted raw
tokenizer-token byte length, including delimiters, attributes, and text. Pinned `x/net` fails when
`rawLen >= maxBuf`, so before its first `Next` the source MUST call
`tokenizer.SetMaxBuf(MaxMailboxContentHTMLTokenBytes + 1)`: raw length exactly `65_536` passes and
`65_537` reaches `html.ErrBufferExceeded`. That error is `too_large` with a zero result; any
other non-EOF tokenizer error is `invalid`. Count every non-Error token --
Text, Start, End, SelfClosing, Comment and Doctype --
including discarded and suppressed tokens; token `100_000` passes, and `100_001` is `too_large`
before that token is processed. ErrorToken with `io.EOF` is normal only when the strict stack is
empty; EOF with a non-empty stack is `invalid`.

ASCII-lowercase each tag name. The strict stack contains every non-void StartTag: increment depth
before push, so depth `64` passes and `65` is `too_large`; every EndTag must exactly match its top.
SelfClosing tags and void starts add no depth, while an end tag for a void element is `invalid`.
The exact void set is `area,base,br,col,embed,hr,img,input,link,meta,param,source,track,wbr`.

Suppression is inherited through that stack. The exact subtree roots are
`script,style,template,noscript,iframe,frame,object,embed,applet,svg,math,form,img,picture,source,video,audio,track,canvas`;
their text is never emitted. Suppression has precedence over block behavior: if the current stack
is suppressed, or a token opens or closes a suppression-root subtree, that token emits no logical
LF. No descendant Start, End, SelfClosing, or void token inside a suppressed subtree emits text or
a break; self-closing and void suppression roots likewise emit nothing. Discard every attribute,
comment, Doctype, URL, `href`, `src`, `srcset`, `style`, and `alt`; the tokenizer has no accepted
processing-instruction output. Anchor text is retained only outside suppression. The exact
block-break set is
`address,article,aside,blockquote,br,div,dl,dt,dd,fieldset,figcaption,figure,footer,h1..h6,header,hgroup,hr,li,main,nav,ol,p,pre,section,table,tbody,thead,tfoot,tr,td,th,ul`.
Only non-suppressed members of that set add breaks: their start and end each add one logical LF,
except void `br` and `hr`, whose start alone adds one. Thus
`<span>a</span><form><p>x</p></form><span>b</span>` produces `ab`, not a newline-separated value.
Before every append, check the extracted buffer cap without a proportional allocation. No tag,
attribute, URL, comment, raw markup, active primitive, or suppressed-node text can survive.
Remote images are never rewritten or tokenized: they are dropped and `RemoteImagesBlocked` remains
true.

## Required authored-first oracle matrix

`backend/tests/mail/mailbox_content_test.go` MUST cover all of the following before implementation:

| Matrix | Required oracle |
|---|---|
| Reflection ABI | Exact function name/signature; format enum names/order; all six limits, including exported `MaxMailboxContentHTMLTokenBytes = 65_536`; exact Input/Result fields; `MailboxContentErrorKind string`, literal `Invalid`/`TooLarge`, exact sentinels, `MailboxContentError{Kind}`, and pointer `Error`/`Unwrap` signatures. |
| Plain happy/edge | Empty body; valid UTF-8; NFC; exact CRLF/CR/LF/U+2028/U+2029 and horizontal-set ordering; per-line ASCII trim, `>=3` LF collapse to two, outer-LF trim; deterministic/idempotent second normalization. |
| Plain hostile | Invalid UTF-8; allowed HT/LF/CR; every other C0, DEL/C1 and exact U+061C/U+200E/U+200F/U+202A..U+202E/U+2066..U+2069 are `invalid`; no input fragment appears in Result, Error, or unwrap. |
| Rich extraction | Exact non-Error token count; `SetMaxBuf(MaxMailboxContentHTMLTokenBytes + 1)` before first `Next`; ASCII-lower tags; strict void/self-closing/start/end stack; retained anchor label outside suppression; complete block-break set only when non-suppressed; decoded entity controls; same canonicalizer and deterministic repeated normalization. |
| Rich hostile | Every exact suppressed root and inherited subtree; every attr/href/src/srcset/style/alt; comments/Doctype/no PI; entity-encoded controls; void end, mismatch and unclosed stack. The separate unterminated quoted-attribute `"<a data-x=\"" + strings.Repeat("x", MaxMailboxContentHTMLTokenBytes)` vector remains below the body cap yet reaches `html.ErrBufferExceeded` and returns only `too_large`/zero result/no panic/no leak; a small quoted attribute succeeds. None may leak markup, URL, active/fetch primitive, token, or raw input. |
| Token-byte boundary | `MaxMailboxContentHTMLTokenBytes = 65_536` is accepted raw token bytes including delimiters/attributes/text. With pinned `x/net` `rawLen >= maxBuf` behavior, source calls `SetMaxBuf(MaxMailboxContentHTMLTokenBytes + 1)` before first `Next`. `"<!--" + strings.Repeat("x", MaxMailboxContentHTMLTokenBytes-7) + "-->"` is raw token length `65_536`, succeeds and yields empty plain output; replacing payload with `MaxMailboxContentHTMLTokenBytes-6` is raw `65_537`, returns only typed `too_large` and zero Result. |
| Bounds | Raw input bytes/runes before scanning/allocation; canonical output bytes/runes after normalization; literal tokens `100000/100001`; depth `64/65`; HTML raw tokenizer-token bytes `65_536`, including delimiters/attributes/text; extracted buffer checked before append. Exact succeeds; plus-one returns only typed `too_large`, zero Result and no partial output. |
| Error safety | Unsupported/zero format, invalid parser/stack/control conditions are typed `invalid`; `html.ErrBufferExceeded` is typed `too_large`; static source review, rather than an impossible generic external injection oracle, owns mapping every other non-EOF tokenizer error to `invalid`; known Error/Unwrap exactly match their sentinel, nil/unknown Error is `mailbox content error` with nil unwrap; `errors.As`/`errors.Is` distinguish only the matching class. |
| Purity | No network, file, DB, proxy, tokenizer callback I/O, clock, goroutine, global mutable state, or raw rich return. Result depends only on `MailboxContentInput`; no remote-image URL, intent, or token exists. |

## Boundaries and blockers

The only dependencies of this plain-only 05D candidate are accepted 144-03C-Attachment and
accepted 144-04A. It does not depend on 144-04L, does not authorize rich sanitization, and is
strictly severed from remote-image token/proxy behavior and 144-05D-Remote.

144-05D-Remote, rich sanitizer work, token rewriting, proxy/network work, and 144-05E remain
`BLOCKED / NOT_AUTHORIZED`; they need separately selected contracts and lifecycles. Existing
03C-Attachment and 04A acceptance does not lift Ent, generation, migration, persistence,
OpenAPI/API/UI, testframework, launcher, executable, or runtime blockers. Launcher is `NOT_READY`
and all executable/runtime evidence remains `NOT_RUN`.
