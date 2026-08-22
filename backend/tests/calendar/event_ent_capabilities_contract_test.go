package calendar_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"runtime"
	"testing"
)

func TestEventEntCapabilities_ExposeOnlyFrozenOptionalBackends(t *testing.T) {
	testCases := []struct {
		file       string
		interfaces []string
		methods    []string
	}{
		{
			file: "event_ent_update.go",
			interfaces: []string{
				"CalendarEventUpdateTransactionBackend",
				"CalendarEventOccurrenceCancelTransactionBackend",
			},
			methods: []string{
				"ResolveCalendarEventUpdate",
				"PersistCalendarEventUpdate",
				"ResolveCalendarEventOccurrenceCancel",
				"PersistCalendarEventOccurrenceCancel",
			},
		},
		{
			file: "event_ent_rsvp.go",
			interfaces: []string{
				"CalendarEventSeriesRSVPTransactionBackend",
				"CalendarEventOccurrenceRSVPTransactionBackend",
			},
			methods: []string{
				"ResolveCalendarEventSeriesRSVP",
				"PersistCalendarEventSeriesRSVP",
				"ResolveCalendarEventOccurrenceRSVP",
				"PersistCalendarEventOccurrenceRSVP",
			},
		},
		{
			file:       "event_ent_inbound_reply.go",
			interfaces: []string{"CalendarInboundReplyTransactionBackend"},
			methods: []string{
				"ResolveCalendarInboundReplyOwner",
				"ResolveCalendarInboundReply",
				"PersistCalendarInboundReply",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.file, func(t *testing.T) {
			file := parseCalendarCapabilitySource(t, testCase.file)
			for _, interfaceName := range testCase.interfaces {
				if !calendarCapabilityHasBackendAssertion(file, interfaceName) {
					t.Fatalf("missing eventEntTransactionBackend assertion for %s", interfaceName)
				}
			}
			for _, methodName := range testCase.methods {
				method, found := calendarCapabilityMethod(file, methodName)
				if !found {
					t.Fatalf("missing eventEntTransactionBackend method %s", methodName)
				}
				if !calendarCapabilityCalls(method, "ForUpdate") {
					t.Fatalf("%s must lock authoritative Ent state", methodName)
				}
			}
		})
	}
}

func parseCalendarCapabilitySource(t *testing.T, name string) *ast.File {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate capability contract test")
	}
	path := filepath.Join(filepath.Dir(thisFile), "..", "..", "internal", "calendar", name)
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}
	return file
}

func calendarCapabilityHasBackendAssertion(file *ast.File, interfaceName string) bool {
	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok || general.Tok != token.VAR {
			continue
		}
		for _, specification := range general.Specs {
			value, ok := specification.(*ast.ValueSpec)
			if !ok || len(value.Values) != 1 {
				continue
			}
			interfaceIdentifier, ok := value.Type.(*ast.Ident)
			if !ok || interfaceIdentifier.Name != interfaceName {
				continue
			}
			conversion, ok := value.Values[0].(*ast.CallExpr)
			if !ok {
				continue
			}
			pointer, ok := conversion.Fun.(*ast.ParenExpr)
			if !ok {
				continue
			}
			star, ok := pointer.X.(*ast.StarExpr)
			if !ok {
				continue
			}
			backend, ok := star.X.(*ast.Ident)
			if ok && backend.Name == "eventEntTransactionBackend" {
				return true
			}
		}
	}
	return false
}

func calendarCapabilityMethod(file *ast.File, name string) (*ast.FuncDecl, bool) {
	for _, declaration := range file.Decls {
		method, ok := declaration.(*ast.FuncDecl)
		if !ok || method.Recv == nil || method.Name.Name != name || len(method.Recv.List) != 1 {
			continue
		}
		receiver, ok := method.Recv.List[0].Type.(*ast.StarExpr)
		if !ok {
			continue
		}
		backend, ok := receiver.X.(*ast.Ident)
		if ok && backend.Name == "eventEntTransactionBackend" {
			return method, true
		}
	}
	return nil, false
}

func calendarCapabilityCalls(method *ast.FuncDecl, selector string) bool {
	called := false
	ast.Inspect(method.Body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		selected, ok := call.Fun.(*ast.SelectorExpr)
		if ok && selected.Sel.Name == selector {
			called = true
		}
		return !called
	})
	return called
}
