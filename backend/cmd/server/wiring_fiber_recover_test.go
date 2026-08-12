package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/valerii/hub/backend/internal/platform/config"
)

// TC-PLATFORM-01 / A14-007: recover config EnableStackTrace follows DevMode.
// Mutation: hardcode EnableStackTrace:true in fiberRecoverConfig → false case FAIL.
func TestFiberRecoverConfig_EnableStackTraceFollowsDevMode_A14_007(t *testing.T) {
	cases := []struct {
		name    string
		devMode bool
		want    bool
	}{
		{name: "dev_on", devMode: true, want: true},
		{name: "dev_off", devMode: false, want: false},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := fiberRecoverConfig(config.ServerConfig{DevMode: tt.devMode})
			if got.EnableStackTrace != tt.want {
				t.Fatalf("EnableStackTrace = %v, want %v (DevMode=%v)", got.EnableStackTrace, tt.want, tt.devMode)
			}
		})
	}
}

// Mutation: replace fiberRecoverConfig(...) with Config{EnableStackTrace: true} in
// wiring_fiber.go → this FAIL (source assert). Builder unit alone is not enough.
func TestWiringFiber_UsesFiberRecoverConfig_A14_007(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	srcPath := filepath.Join(filepath.Dir(thisFile), "wiring_fiber.go")
	src, err := os.ReadFile(srcPath)
	if err != nil {
		t.Fatalf("read wiring_fiber.go: %v", err)
	}
	body := string(src)
	if !strings.Contains(body, "fiberrecover.New(fiberRecoverConfig(") {
		t.Fatal("wiring_fiber.go must pass fiberRecoverConfig(...) into fiberrecover.New")
	}
	if strings.Contains(body, "EnableStackTrace: true") {
		t.Fatal("wiring_fiber.go must not hardcode EnableStackTrace: true")
	}
}
