//go:build hubtest

package observability_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/propagation"

	"github.com/valerii/hub/backend/internal/platform/observability"
)

// TestInitTracing_PropagatorWithoutBaggage_P3BE12 — W4-A3-005:
// client baggage не инжектится в ctx при Extract (только TraceContext).
func TestInitTracing_PropagatorWithoutBaggage_P3BE12(t *testing.T) {
	observability.ResetTracingForTest()
	t.Cleanup(observability.ResetTracingForTest)

	// Как InitTracing: только TraceContext (мутация «добавить Baggage» → тест краснеет).
	otel.SetTextMapPropagator(propagation.TraceContext{})

	member, err := baggage.NewMember("secret", "pii-value")
	require.NoError(t, err)
	bag, err := baggage.New(member)
	require.NoError(t, err)
	ctx := baggage.ContextWithBaggage(context.Background(), bag)

	carrier := propagation.MapCarrier{}
	// Composite TraceContext+Baggage записал бы baggage; чистый TraceContext — нет.
	propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	).Inject(ctx, carrier)
	require.NotEmpty(t, carrier.Get("baggage"), "fixture: baggage header must be present")

	extracted := otel.GetTextMapPropagator().Extract(context.Background(), carrier)
	got := baggage.FromContext(extracted)
	require.Empty(t, got.Members(), "TraceContext-only propagator must ignore baggage header")
}
