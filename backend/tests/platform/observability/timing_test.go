package observability_test
import (
	"github.com/valerii/hub/backend/internal/platform/observability"
	"context"
	"errors"
	"testing"
)

func TestTimed_Success(t *testing.T) {
	ctx := context.Background()
	v, err := observability.Timed(ctx, "test.ok", func(ctx context.Context) (int, error) {
		return 42, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 42 {
		t.Fatalf("expected 42, got %d", v)
	}
}

func TestTimed_Error(t *testing.T) {
	ctx := context.Background()
	want := errors.New("boom")
	_, err := observability.Timed(ctx, "test.fail", func(ctx context.Context) (int, error) {
		return 0, want
	})
	if !errors.Is(err, want) {
		t.Fatalf("expected %v, got %v", want, err)
	}
}

func TestTimedVoid(t *testing.T) {
	ctx := context.Background()
	if err := observability.TimedVoid(ctx, "void.ok", func(ctx context.Context) error { return nil }); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
