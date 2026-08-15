package mail_test

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/valerii/hub/backend/internal/mail"
)

func TestEvaluateAttachmentPolicy_NormalizesValidCleanAttachment(t *testing.T) {
	input := validAttachmentPolicyInput()
	input.Attachments[0].Filename = "Cafe\u0301.txt"
	input.Attachments[0].MediaType = "TEXT/PLAIN"

	decision, err := mail.EvaluateAttachmentPolicy(input)
	if err != nil {
		t.Fatalf("EvaluateAttachmentPolicy() error = %v", err)
	}
	if got := decision.Attachments[0].Filename; got != "Caf\u00e9.txt" {
		t.Fatalf("normalized filename = %q, want NFC Caf\u00e9.txt", got)
	}
	if got := decision.Attachments[0].MediaType; got != "text/plain" {
		t.Fatalf("normalized media type = %q, want text/plain", got)
	}
}

func TestEvaluateAttachmentPolicy_ConfigurationAndCountBounds(t *testing.T) {
	for _, testCase := range []struct {
		name    string
		mutate  func(*mail.AttachmentPolicyInput)
		wantErr error
	}{
		{name: "policy zero", mutate: func(input *mail.AttachmentPolicyInput) { input.MaxAttachmentBytes = 0 }, wantErr: mail.ErrAttachmentInvalid},
		{name: "policy one", mutate: func(input *mail.AttachmentPolicyInput) { input.MaxAttachmentBytes = 1; input.Attachments[0].SizeBytes = 1 }, wantErr: nil},
		{name: "policy maximum", mutate: func(input *mail.AttachmentPolicyInput) { input.MaxAttachmentBytes = mail.MaxAttachmentPolicyBytes; input.Attachments[0].SizeBytes = mail.MaxAttachmentPolicyBytes }, wantErr: nil},
		{name: "policy above maximum", mutate: func(input *mail.AttachmentPolicyInput) { input.MaxAttachmentBytes = mail.MaxAttachmentPolicyBytes + 1 }, wantErr: mail.ErrAttachmentInvalid},
		{name: "zero attachments", mutate: func(input *mail.AttachmentPolicyInput) { input.Attachments = nil }, wantErr: nil},
		{name: "fifty per item maximum attachments", mutate: func(input *mail.AttachmentPolicyInput) { input.Attachments = attachmentInputs(50, input.MaxAttachmentBytes) }, wantErr: nil},
		{name: "fifty one attachments", mutate: func(input *mail.AttachmentPolicyInput) { input.Attachments = attachmentInputs(51, 1) }, wantErr: mail.ErrAttachmentTooLarge},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := validAttachmentPolicyInput()
			testCase.mutate(&input)
			decision, err := mail.EvaluateAttachmentPolicy(input)
			assertAttachmentOutcome(t, decision, err, testCase.wantErr)
		})
	}
}

func TestEvaluateAttachmentPolicy_SizeBoundsAreOriginAware(t *testing.T) {
	for _, testCase := range []struct {
		name    string
		mutate  func(*mail.AttachmentPolicyInput)
		wantErr error
	}{
		{name: "upload zero", mutate: func(input *mail.AttachmentPolicyInput) { input.Attachments[0].Origin = mail.AttachmentByteOriginUpload; input.Attachments[0].SizeBytes = 0 }, wantErr: mail.ErrAttachmentInvalid},
		{name: "upload negative", mutate: func(input *mail.AttachmentPolicyInput) { input.Attachments[0].Origin = mail.AttachmentByteOriginUpload; input.Attachments[0].SizeBytes = -1 }, wantErr: mail.ErrAttachmentInvalid},
		{name: "upload one", mutate: func(input *mail.AttachmentPolicyInput) { input.Attachments[0].Origin = mail.AttachmentByteOriginUpload; input.Attachments[0].SizeBytes = 1 }, wantErr: nil},
		{name: "upload maximum", mutate: func(input *mail.AttachmentPolicyInput) { input.Attachments[0].Origin = mail.AttachmentByteOriginUpload; input.Attachments[0].SizeBytes = input.MaxAttachmentBytes }, wantErr: nil},
		{name: "upload maximum plus one", mutate: func(input *mail.AttachmentPolicyInput) { input.Attachments[0].Origin = mail.AttachmentByteOriginUpload; input.Attachments[0].SizeBytes = input.MaxAttachmentBytes + 1 }, wantErr: mail.ErrAttachmentTooLarge},
		{name: "engine mime zero", mutate: func(input *mail.AttachmentPolicyInput) { input.Attachments[0].Origin = mail.AttachmentByteOriginEngineMIME; input.Attachments[0].SizeBytes = 0 }, wantErr: nil},
		{name: "engine mime negative", mutate: func(input *mail.AttachmentPolicyInput) { input.Attachments[0].Origin = mail.AttachmentByteOriginEngineMIME; input.Attachments[0].SizeBytes = -1 }, wantErr: mail.ErrAttachmentInvalid},
		{name: "engine mime one", mutate: func(input *mail.AttachmentPolicyInput) { input.Attachments[0].Origin = mail.AttachmentByteOriginEngineMIME; input.Attachments[0].SizeBytes = 1 }, wantErr: nil},
		{name: "engine mime maximum", mutate: func(input *mail.AttachmentPolicyInput) { input.Attachments[0].Origin = mail.AttachmentByteOriginEngineMIME; input.Attachments[0].SizeBytes = input.MaxAttachmentBytes }, wantErr: nil},
		{name: "engine mime maximum plus one", mutate: func(input *mail.AttachmentPolicyInput) { input.Attachments[0].Origin = mail.AttachmentByteOriginEngineMIME; input.Attachments[0].SizeBytes = input.MaxAttachmentBytes + 1 }, wantErr: mail.ErrAttachmentTooLarge},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := validAttachmentPolicyInput()
			testCase.mutate(&input)
			decision, err := mail.EvaluateAttachmentPolicy(input)
			assertAttachmentOutcome(t, decision, err, testCase.wantErr)
		})
	}
}

func TestEvaluateAttachmentPolicy_RejectsEveryFilenamePathCharacter(t *testing.T) {
	for _, character := range []string{"/", "\\", "<", ">", ":", "\"", "|", "?", "*"} {
		t.Run(character, func(t *testing.T) {
			input := validAttachmentPolicyInput()
			input.Attachments[0].Filename = "safe" + character + "name.txt"
			decision, err := mail.EvaluateAttachmentPolicy(input)
			assertAttachmentOutcome(t, decision, err, mail.ErrAttachmentInvalid)
		})
	}
}

func TestEvaluateAttachmentPolicy_RejectsUnknownEnums(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		mutate func(*mail.AttachmentPolicyInput)
	}{
		{name: "operation", mutate: func(input *mail.AttachmentPolicyInput) { input.Operation = mail.AttachmentPolicyOperation("unknown") }},
		{name: "origin", mutate: func(input *mail.AttachmentPolicyInput) { input.Attachments[0].Origin = mail.AttachmentByteOrigin("unknown") }},
		{name: "disposition", mutate: func(input *mail.AttachmentPolicyInput) { input.Attachments[0].Disposition = mail.EngineAttachmentDisposition("unknown") }},
		{name: "scan", mutate: func(input *mail.AttachmentPolicyInput) { input.Attachments[0].ScanStatus = mail.AttachmentScanStatus("unknown") }},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := validAttachmentPolicyInput()
			testCase.mutate(&input)
			decision, err := mail.EvaluateAttachmentPolicy(input)
			assertAttachmentOutcome(t, decision, err, mail.ErrAttachmentInvalid)
		})
	}
}

func TestEvaluateAttachmentPolicy_FilenameValidationAndSecuritySkeleton(t *testing.T) {
	for _, testCase := range []struct {
		name    string
		filename string
		wantErr error
	}{
		{name: "empty", filename: "", wantErr: mail.ErrAttachmentInvalid},
		{name: "invalid UTF8", filename: string([]byte{0xff}), wantErr: mail.ErrAttachmentInvalid},
		{name: "control", filename: "bad\x00.txt", wantErr: mail.ErrAttachmentInvalid},
		{name: "format", filename: "bad\u200e.txt", wantErr: mail.ErrAttachmentInvalid},
		{name: "line separator", filename: "bad\u2028.txt", wantErr: mail.ErrAttachmentInvalid},
		{name: "paragraph separator", filename: "bad\u2029.txt", wantErr: mail.ErrAttachmentInvalid},
		{name: "leading whitespace", filename: " bad.txt", wantErr: mail.ErrAttachmentInvalid},
		{name: "trailing whitespace", filename: "bad.txt\u2003", wantErr: mail.ErrAttachmentInvalid},
		{name: "leading dot", filename: ".hidden", wantErr: mail.ErrAttachmentInvalid},
		{name: "trailing dot", filename: "bad.", wantErr: mail.ErrAttachmentInvalid},
		{name: "dot", filename: ".", wantErr: mail.ErrAttachmentInvalid},
		{name: "dot dot", filename: "..", wantErr: mail.ErrAttachmentInvalid},
		{name: "path character", filename: "dir/file.txt", wantErr: mail.ErrAttachmentInvalid},
		{name: "windows character", filename: "bad:name.txt", wantErr: mail.ErrAttachmentInvalid},
		{name: "consecutive dots", filename: "bad..name.txt", wantErr: mail.ErrAttachmentInvalid},
		{name: "safe multi dot", filename: "archive.tar.gz", wantErr: nil},
		{name: "unsafe final extension", filename: "invoice.exe", wantErr: mail.ErrAttachmentUnsafeType},
		{name: "unsafe double extension", filename: "invoice.exe.pdf", wantErr: mail.ErrAttachmentUnsafeType},
		{name: "nfkc skeleton extension", filename: "invoice\uff0eexe", wantErr: mail.ErrAttachmentUnsafeType},
		{name: "nfkc skeleton leading dot", filename: "\uff0ehidden.txt", wantErr: mail.ErrAttachmentInvalid},
		{name: "nfkc skeleton consecutive dots", filename: "safe\uff0e.txt", wantErr: mail.ErrAttachmentInvalid},
		{name: "nfkc skeleton path separator", filename: "dir\uff0ffile.txt", wantErr: mail.ErrAttachmentInvalid},
		{name: "nfkc skeleton windows separator", filename: "safe\uff1aname.txt", wantErr: mail.ErrAttachmentInvalid},
		{name: "maximum runes", filename: strings.Repeat("a", mail.MaxAttachmentTextRunes-4) + ".txt", wantErr: nil},
		{name: "above maximum runes", filename: strings.Repeat("a", mail.MaxAttachmentTextRunes-3) + ".txt", wantErr: mail.ErrAttachmentTooLarge},
		{name: "above maximum bytes", filename: strings.Repeat("\u00e9", 511) + ".txt", wantErr: mail.ErrAttachmentTooLarge},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := validAttachmentPolicyInput()
			input.Attachments[0].Filename = testCase.filename
			decision, err := mail.EvaluateAttachmentPolicy(input)
			assertAttachmentOutcome(t, decision, err, testCase.wantErr)
		})
	}
}

func TestEvaluateAttachmentPolicy_RejectsEveryUnsafeExtension(t *testing.T) {
	for _, extension := range []string{"svg", "svgz", "exe", "com", "bat", "cmd", "msi", "msp", "mst", "scr", "pif", "cpl", "jar", "js", "jse", "vbs", "vbe", "wsf", "wsh", "ps1", "psm1", "psd1", "hta", "chm", "reg", "lnk", "url", "scf", "sh", "bash", "zsh", "fish", "app", "apk", "dmg", "pkg", "deb", "rpm", "iso", "img", "appimage"} {
		t.Run(extension, func(t *testing.T) {
			input := validAttachmentPolicyInput()
			input.Attachments[0].Filename = "safe." + strings.ToUpper(extension) + ".txt"
			decision, err := mail.EvaluateAttachmentPolicy(input)
			assertAttachmentOutcome(t, decision, err, mail.ErrAttachmentUnsafeType)
		})
	}
}

func TestEvaluateAttachmentPolicy_MediaTypeGrammarAndUnsafeTable(t *testing.T) {
	for _, testCase := range []struct {
		name      string
		mediaType string
		wantErr   error
	}{
		{name: "empty", mediaType: "", wantErr: mail.ErrAttachmentInvalid},
		{name: "lowercases", mediaType: "APPLICATION/OCTET-STREAM", wantErr: nil},
		{name: "missing slash", mediaType: "textplain", wantErr: mail.ErrAttachmentInvalid},
		{name: "extra slash", mediaType: "text/plain/extra", wantErr: mail.ErrAttachmentInvalid},
		{name: "parameter", mediaType: "text/plain; charset=utf-8", wantErr: mail.ErrAttachmentInvalid},
		{name: "whitespace", mediaType: "text /plain", wantErr: mail.ErrAttachmentInvalid},
		{name: "wildcard", mediaType: "text/*", wantErr: mail.ErrAttachmentInvalid},
		{name: "invalid token", mediaType: "text/(plain)", wantErr: mail.ErrAttachmentInvalid},
		{name: "exact character maximum", mediaType: strings.Repeat("a", mail.MaxAttachmentTextRunes-2) + "/b", wantErr: nil},
		{name: "above character maximum", mediaType: strings.Repeat("a", mail.MaxAttachmentTextRunes-1) + "/b", wantErr: mail.ErrAttachmentTooLarge},
		{name: "above maximum bytes", mediaType: strings.Repeat("a", mail.MaxAttachmentTextBytes) + "/b", wantErr: mail.ErrAttachmentTooLarge},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := validAttachmentPolicyInput()
			input.Attachments[0].MediaType = testCase.mediaType
			decision, err := mail.EvaluateAttachmentPolicy(input)
			assertAttachmentOutcome(t, decision, err, testCase.wantErr)
		})
	}

	for _, mediaType := range []string{"image/svg+xml", "application/x-msdownload", "application/x-msdos-program", "application/vnd.microsoft.portable-executable", "application/x-dosexec", "application/x-msi", "application/java-archive", "application/x-java-archive", "application/javascript", "text/javascript", "application/ecmascript", "text/ecmascript", "application/x-sh", "application/x-csh", "application/x-powershell", "application/vnd.microsoft.windows-registry", "application/vnd.android.package-archive", "application/x-apple-diskimage", "application/x-iso9660-image", "application/wasm"} {
		t.Run(mediaType, func(t *testing.T) {
			input := validAttachmentPolicyInput()
			input.Attachments[0].MediaType = strings.ToUpper(mediaType)
			decision, err := mail.EvaluateAttachmentPolicy(input)
			assertAttachmentOutcome(t, decision, err, mail.ErrAttachmentUnsafeType)
		})
	}
}

func TestEvaluateAttachmentPolicy_ContentIDDispositionAndTextBounds(t *testing.T) {
	for _, testCase := range []struct {
		name    string
		mutate  func(*mail.AttachmentPolicyInput)
		wantErr error
	}{
		{name: "inline needs content id", mutate: func(input *mail.AttachmentPolicyInput) { input.Attachments[0].Disposition = mail.EngineAttachmentDispositionInline }, wantErr: mail.ErrAttachmentInvalid},
		{name: "inline preserves content id", mutate: func(input *mail.AttachmentPolicyInput) { input.Attachments[0].Disposition = mail.EngineAttachmentDispositionInline; input.Attachments[0].ContentID = "MiXeD-Content-ID@example.test" }, wantErr: nil},
		{name: "attachment may carry content id", mutate: func(input *mail.AttachmentPolicyInput) { input.Attachments[0].ContentID = "MiXeD-Content-ID@example.test" }, wantErr: nil},
		{name: "invalid content id UTF8", mutate: func(input *mail.AttachmentPolicyInput) { input.Attachments[0].ContentID = string([]byte{0xff}) }, wantErr: mail.ErrAttachmentInvalid},
		{name: "control content id", mutate: func(input *mail.AttachmentPolicyInput) { input.Attachments[0].ContentID = "bad\r\n" }, wantErr: mail.ErrAttachmentInvalid},
		{name: "format content id", mutate: func(input *mail.AttachmentPolicyInput) { input.Attachments[0].ContentID = "bad\u200e" }, wantErr: mail.ErrAttachmentInvalid},
		{name: "line separator content id", mutate: func(input *mail.AttachmentPolicyInput) { input.Attachments[0].ContentID = "bad\u2028" }, wantErr: mail.ErrAttachmentInvalid},
		{name: "paragraph separator content id", mutate: func(input *mail.AttachmentPolicyInput) { input.Attachments[0].ContentID = "bad\u2029" }, wantErr: mail.ErrAttachmentInvalid},
		{name: "content id maximum", mutate: func(input *mail.AttachmentPolicyInput) { input.Attachments[0].ContentID = strings.Repeat("a", mail.MaxAttachmentTextRunes) }, wantErr: nil},
		{name: "content id above rune maximum", mutate: func(input *mail.AttachmentPolicyInput) { input.Attachments[0].ContentID = strings.Repeat("a", mail.MaxAttachmentTextRunes+1) }, wantErr: mail.ErrAttachmentTooLarge},
		{name: "content id above byte maximum", mutate: func(input *mail.AttachmentPolicyInput) { input.Attachments[0].ContentID = strings.Repeat("\u00e9", 511) }, wantErr: mail.ErrAttachmentTooLarge},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := validAttachmentPolicyInput()
			testCase.mutate(&input)
			decision, err := mail.EvaluateAttachmentPolicy(input)
			assertAttachmentOutcome(t, decision, err, testCase.wantErr)
			if testCase.name == "inline preserves content id" && decision.Attachments[0].ContentID != "MiXeD-Content-ID@example.test" {
				t.Fatalf("ContentID = %q, want exact preservation", decision.Attachments[0].ContentID)
			}
		})
	}
}

func TestEvaluateAttachmentPolicy_ScanStatesAndSHA(t *testing.T) {
	for _, testCase := range []struct {
		name      string
		operation mail.AttachmentPolicyOperation
		status    mail.AttachmentScanStatus
		shaLength int
		wantErr   error
	}{
		{name: "validate pending empty", operation: mail.AttachmentPolicyOperationValidate, status: mail.AttachmentScanStatusPending, shaLength: 0},
		{name: "validate pending sha", operation: mail.AttachmentPolicyOperationValidate, status: mail.AttachmentScanStatusPending, shaLength: 32},
		{name: "validate error empty", operation: mail.AttachmentPolicyOperationValidate, status: mail.AttachmentScanStatusError, shaLength: 0},
		{name: "validate error sha", operation: mail.AttachmentPolicyOperationValidate, status: mail.AttachmentScanStatusError, shaLength: 32},
		{name: "validate rejected sha", operation: mail.AttachmentPolicyOperationValidate, status: mail.AttachmentScanStatusRejected, shaLength: 32},
		{name: "validate clean sha", operation: mail.AttachmentPolicyOperationValidate, status: mail.AttachmentScanStatusClean, shaLength: 32},
		{name: "validate clean missing sha", operation: mail.AttachmentPolicyOperationValidate, status: mail.AttachmentScanStatusClean, shaLength: 0, wantErr: mail.ErrAttachmentInvalid},
		{name: "validate rejected short sha", operation: mail.AttachmentPolicyOperationValidate, status: mail.AttachmentScanStatusRejected, shaLength: 31, wantErr: mail.ErrAttachmentInvalid},
		{name: "validate pending long sha", operation: mail.AttachmentPolicyOperationValidate, status: mail.AttachmentScanStatusPending, shaLength: 33, wantErr: mail.ErrAttachmentInvalid},
		{name: "send pending", operation: mail.AttachmentPolicyOperationSend, status: mail.AttachmentScanStatusPending, shaLength: 0, wantErr: mail.ErrAttachmentNotClean},
		{name: "send rejected", operation: mail.AttachmentPolicyOperationSend, status: mail.AttachmentScanStatusRejected, shaLength: 32, wantErr: mail.ErrAttachmentNotClean},
		{name: "send error", operation: mail.AttachmentPolicyOperationSend, status: mail.AttachmentScanStatusError, shaLength: 32, wantErr: mail.ErrAttachmentNotClean},
		{name: "send clean", operation: mail.AttachmentPolicyOperationSend, status: mail.AttachmentScanStatusClean, shaLength: 32},
		{name: "read pending", operation: mail.AttachmentPolicyOperationRead, status: mail.AttachmentScanStatusPending, shaLength: 0, wantErr: mail.ErrAttachmentNotClean},
		{name: "read rejected", operation: mail.AttachmentPolicyOperationRead, status: mail.AttachmentScanStatusRejected, shaLength: 32, wantErr: mail.ErrAttachmentNotClean},
		{name: "read error", operation: mail.AttachmentPolicyOperationRead, status: mail.AttachmentScanStatusError, shaLength: 32, wantErr: mail.ErrAttachmentNotClean},
		{name: "read clean", operation: mail.AttachmentPolicyOperationRead, status: mail.AttachmentScanStatusClean, shaLength: 32},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := validAttachmentPolicyInput()
			input.Operation = testCase.operation
			input.Attachments[0].ScanStatus = testCase.status
			input.Attachments[0].SHA256 = make([]byte, testCase.shaLength)
			decision, err := mail.EvaluateAttachmentPolicy(input)
			assertAttachmentOutcome(t, decision, err, testCase.wantErr)
		})
	}
}

func TestEvaluateAttachmentPolicy_SHA256LengthRulesForEveryScanState(t *testing.T) {
	for _, testCase := range []struct {
		status mail.AttachmentScanStatus
		length int
		wantErr error
	}{
		{status: mail.AttachmentScanStatusPending, length: 0},
		{status: mail.AttachmentScanStatusPending, length: 31, wantErr: mail.ErrAttachmentInvalid},
		{status: mail.AttachmentScanStatusPending, length: 32},
		{status: mail.AttachmentScanStatusPending, length: 33, wantErr: mail.ErrAttachmentInvalid},
		{status: mail.AttachmentScanStatusClean, length: 0, wantErr: mail.ErrAttachmentInvalid},
		{status: mail.AttachmentScanStatusClean, length: 31, wantErr: mail.ErrAttachmentInvalid},
		{status: mail.AttachmentScanStatusClean, length: 32},
		{status: mail.AttachmentScanStatusClean, length: 33, wantErr: mail.ErrAttachmentInvalid},
		{status: mail.AttachmentScanStatusRejected, length: 0, wantErr: mail.ErrAttachmentInvalid},
		{status: mail.AttachmentScanStatusRejected, length: 31, wantErr: mail.ErrAttachmentInvalid},
		{status: mail.AttachmentScanStatusRejected, length: 32},
		{status: mail.AttachmentScanStatusRejected, length: 33, wantErr: mail.ErrAttachmentInvalid},
		{status: mail.AttachmentScanStatusError, length: 0},
		{status: mail.AttachmentScanStatusError, length: 31, wantErr: mail.ErrAttachmentInvalid},
		{status: mail.AttachmentScanStatusError, length: 32},
		{status: mail.AttachmentScanStatusError, length: 33, wantErr: mail.ErrAttachmentInvalid},
	} {
		t.Run(fmt.Sprintf("%s-%d", testCase.status, testCase.length), func(t *testing.T) {
			input := validAttachmentPolicyInput()
			input.Attachments[0].ScanStatus = testCase.status
			input.Attachments[0].SHA256 = make([]byte, testCase.length)
			decision, err := mail.EvaluateAttachmentPolicy(input)
			assertAttachmentOutcome(t, decision, err, testCase.wantErr)
		})
	}
}

func TestEvaluateAttachmentPolicy_PrecedenceAndDeepCopies(t *testing.T) {
	input := validAttachmentPolicyInput()
	input.Attachments[0].Filename = "unsafe.exe"
	input.Attachments[0].SizeBytes = 0
	decision, err := mail.EvaluateAttachmentPolicy(input)
	assertAttachmentOutcome(t, decision, err, mail.ErrAttachmentInvalid)

	input = validAttachmentPolicyInput()
	input.Attachments[0].Filename = "unsafe.exe"
	input.Attachments[0].ScanStatus = mail.AttachmentScanStatusPending
	decision, err = mail.EvaluateAttachmentPolicy(input)
	assertAttachmentOutcome(t, decision, err, mail.ErrAttachmentUnsafeType)

	input = validAttachmentPolicyInput()
	input.Attachments[0].ScanStatus = mail.AttachmentScanStatusPending
	input.Attachments[0].SizeBytes = input.MaxAttachmentBytes + 1
	decision, err = mail.EvaluateAttachmentPolicy(input)
	assertAttachmentOutcome(t, decision, err, mail.ErrAttachmentTooLarge)

	input = validAttachmentPolicyInput()
	original := cloneAttachmentInput(input)
	decision, err = mail.EvaluateAttachmentPolicy(input)
	if err != nil {
		t.Fatalf("valid attachment error = %v", err)
	}
	if !reflect.DeepEqual(input, original) {
		t.Fatalf("input mutated: %#v", input)
	}
	input.Attachments[0].SHA256[0] = 77
	if decision.Attachments[0].SHA256[0] == 77 {
		t.Fatalf("decision SHA aliases input: %#v", decision)
	}
	decision.Attachments[0].SHA256[1] = 88
	if input.Attachments[0].SHA256[1] == 88 {
		t.Fatalf("input SHA aliases decision: %#v", input)
	}
	second, err := mail.EvaluateAttachmentPolicy(original)
	if err != nil || !reflect.DeepEqual(second, mail.AttachmentPolicyDecision{Attachments: original.Attachments}) {
		t.Fatalf("repeated deterministic evaluation = %#v / %v", second, err)
	}
}

func TestAttachmentPolicyErrors_AreTypedStableAndContentFree(t *testing.T) {
	for _, testCase := range []struct {
		kind     mail.AttachmentErrorKind
		sentinel error
		literal  string
		text     string
	}{
		{kind: mail.AttachmentErrorInvalid, sentinel: mail.ErrAttachmentInvalid, literal: "invalid", text: "mail attachment invalid"},
		{kind: mail.AttachmentErrorTooLarge, sentinel: mail.ErrAttachmentTooLarge, literal: "too_large", text: "mail attachment exceeds bounds"},
		{kind: mail.AttachmentErrorUnsafeType, sentinel: mail.ErrAttachmentUnsafeType, literal: "unsafe_type", text: "mail attachment type rejected"},
		{kind: mail.AttachmentErrorNotClean, sentinel: mail.ErrAttachmentNotClean, literal: "not_clean", text: "mail attachment not clean"},
	} {
		t.Run(testCase.literal, func(t *testing.T) {
			err := &mail.AttachmentPolicyError{Kind: testCase.kind}
			if string(testCase.kind) != testCase.literal || err.Error() != testCase.text || testCase.sentinel.Error() != testCase.text || !errors.Is(err, testCase.sentinel) {
				t.Fatalf("attachment error contract = kind %q text %q sentinel %v", testCase.kind, err.Error(), errors.Is(err, testCase.sentinel))
			}
		})
	}
	for _, err := range []*mail.AttachmentPolicyError{nil, {Kind: mail.AttachmentErrorKind("HOSTILE-ATTACHMENT-MARKER@example.test")}} {
		if err.Error() != "mail attachment policy error" || errors.Unwrap(err) != nil {
			t.Fatalf("generic AttachmentPolicyError = %q / %v", err.Error(), errors.Unwrap(err))
		}
	}

	input := validAttachmentPolicyInput()
	input.Attachments[0].Filename = "HOSTILE-ATTACHMENT-MARKER@example.test.exe"
	_, err := mail.EvaluateAttachmentPolicy(input)
	if !errors.Is(err, mail.ErrAttachmentUnsafeType) || strings.Contains(err.Error(), "HOSTILE-ATTACHMENT-MARKER") {
		t.Fatalf("unsafe attachment error leaks input: %v", err)
	}
}

func assertAttachmentOutcome(t *testing.T, decision mail.AttachmentPolicyDecision, err error, wantErr error) {
	t.Helper()
	if wantErr == nil {
		if err != nil {
			t.Fatalf("EvaluateAttachmentPolicy() error = %v", err)
		}
		return
	}
	if !errors.Is(err, wantErr) {
		t.Fatalf("EvaluateAttachmentPolicy() error = %v, want errors.Is(%v)", err, wantErr)
	}
	if !reflect.DeepEqual(decision, mail.AttachmentPolicyDecision{}) {
		t.Fatalf("failure decision = %#v, want zero value", decision)
	}
}

func validAttachmentPolicyInput() mail.AttachmentPolicyInput {
	return mail.AttachmentPolicyInput{
		Operation:          mail.AttachmentPolicyOperationValidate,
		MaxAttachmentBytes: 100,
		Attachments: []mail.AttachmentPolicyAttachment{{
			Origin:      mail.AttachmentByteOriginUpload,
			Filename:    "safe.txt",
			MediaType:   "text/plain",
			Disposition: mail.EngineAttachmentDispositionAttachment,
			ScanStatus:  mail.AttachmentScanStatusClean,
			SizeBytes:   1,
			SHA256:      sha256Bytes(),
		}},
	}
}

func attachmentInputs(count int, size int64) []mail.AttachmentPolicyAttachment {
	attachments := make([]mail.AttachmentPolicyAttachment, count)
	for index := range attachments {
		attachments[index] = mail.AttachmentPolicyAttachment{Origin: mail.AttachmentByteOriginUpload, Filename: "safe.txt", MediaType: "text/plain", Disposition: mail.EngineAttachmentDispositionAttachment, ScanStatus: mail.AttachmentScanStatusClean, SizeBytes: size, SHA256: sha256Bytes()}
	}
	return attachments
}

func sha256Bytes() []byte {
	return make([]byte, 32)
}

func cloneAttachmentInput(input mail.AttachmentPolicyInput) mail.AttachmentPolicyInput {
	clone := input
	clone.Attachments = append([]mail.AttachmentPolicyAttachment(nil), input.Attachments...)
	for index := range clone.Attachments {
		clone.Attachments[index].SHA256 = append([]byte(nil), input.Attachments[index].SHA256...)
	}
	return clone
}
