//go:build live

// Live SDK check against a running evaluator. Run with:
//
//	FLAGCAST_ADDR=localhost:50051 go test -tags=live ./packages/sdk/...
//
// Excluded from the default build/CI (needs a live server + seeded flags).
package flagcast

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"testing"
	"time"
)

func apiBase() string {
	if v := os.Getenv("FLAGCAST_API"); v != "" {
		return v
	}
	return "http://localhost:8080"
}

func apiDo(t *testing.T, method, path, body string) int {
	t.Helper()
	req, err := http.NewRequest(method, apiBase()+path, bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("req: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	resp.Body.Close()
	return resp.StatusCode
}

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

func TestSDKWatchLive(t *testing.T) {
	addr := os.Getenv("FLAGCAST_ADDR")
	if addr == "" {
		addr = "localhost:50051"
	}
	c, err := Dial(addr)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer c.Close()

	// Ensure the watched flag exists and is OFF.
	apiDo(t, http.MethodPost, "/api/flags", `{"key":"watch-me","name":"Watch me","enabled":false,"rollout":100}`)
	apiDo(t, http.MethodPut, "/api/flags/watch-me", `{"name":"Watch me","enabled":false,"rollout":100}`)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	changes := make(chan Change, 32)
	go func() { _ = c.Watch(ctx, Context{Key: "u1"}, func(ch Change) { changes <- ch }) }()

	// Drain the initial snapshot.
	time.Sleep(300 * time.Millisecond)
	for drained := false; !drained; {
		select {
		case <-changes:
		default:
			drained = true
		}
	}

	// Flip it ON and expect a pushed update.
	if code := apiDo(t, http.MethodPut, "/api/flags/watch-me", `{"name":"Watch me","enabled":true,"rollout":100}`); code != http.StatusOK {
		t.Fatalf("enable returned %d", code)
	}
	deadline := time.After(5 * time.Second)
	for {
		select {
		case ch := <-changes:
			if ch.FlagKey == "watch-me" && ch.Value {
				t.Logf("received pushed update: %+v", ch)
				return
			}
		case <-deadline:
			t.Fatal("did not receive watch-me=true update within 5s")
		}
	}
}
