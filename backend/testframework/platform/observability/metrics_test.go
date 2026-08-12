package observability_test
import (
	"github.com/valerii/hub/backend/internal/platform/observability"
	"context"
	"testing"
)

func TestRecordOperationAndSnapshot(t *testing.T) {
	observability.ResetMetrics()
	t.Cleanup(observability.ResetMetrics)

	observability.ExportRecordOperation("test.op", 100, false)
	observability.ExportRecordOperation("test.op", 200, true)
	observability.ExportRecordOperation("audit.op", 50, false)

	snap := observability.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2 ops, got %d", len(snap))
	}
	if snap[0].Operation != "audit.op" || snap[0].Count != 1 {
		t.Fatalf("unexpected first metric: %+v", snap[0])
	}
	if snap[1].Operation != "test.op" {
		t.Fatalf("expected test.op second, got %s", snap[1].Operation)
	}
	if snap[1].Count != 2 || snap[1].ErrorCount != 1 || snap[1].TotalMs != 300 {
		t.Fatalf("unexpected test.op stats: %+v", snap[1])
	}
	if snap[1].AvgMs != 150 {
		t.Fatalf("expected avg 150, got %f", snap[1].AvgMs)
	}
}

func TestTimed_RecordsMetrics(t *testing.T) {
	observability.ResetMetrics()
	t.Cleanup(observability.ResetMetrics)

	ctx := context.Background()
	_, _ = observability.Timed(ctx, "timed.record", func(ctx context.Context) (int, error) {
		return 1, nil
	})

	snap := observability.Snapshot()
	if len(snap) != 1 || snap[0].Operation != "timed.record" || snap[0].Count != 1 {
		t.Fatalf("expected timed.record metric, got %+v", snap)
	}
}
