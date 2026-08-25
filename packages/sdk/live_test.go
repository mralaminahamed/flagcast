//go:build live

// Live SDK check against a running evaluator. Run with:
//
//	FLAGCAST_ADDR=localhost:50051 go test -tags=live ./packages/sdk/...
//
// Excluded from the default build/CI (needs a live server + seeded flags).
package flagcast

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestSDKLive(t *testing.T) {
	addr := os.Getenv("FLAGCAST_ADDR")
	if addr == "" {
		addr = "localhost:50051"
	}
	c, err := Dial(addr)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer c.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if !c.BoolValue(ctx, "new-ui", Context{Key: "user-1"}, false) {
		t.Error("new-ui should be on")
	}
	if c.BoolValue(ctx, "dark-mode", Context{Key: "user-1"}, true) {
		t.Error("dark-mode should be off")
	}
	// unknown flag falls back to the default
	if !c.BoolValue(ctx, "does-not-exist", Context{Key: "user-1"}, true) {
		t.Error("unknown flag should return the default (true)")
	}
	all, err := c.AllValues(ctx, Context{Key: "user-1"})
	if err != nil || len(all) == 0 {
		t.Fatalf("AllValues: %v (len=%d)", err, len(all))
	}
	t.Logf("AllValues(user-1) = %v", all)
}
