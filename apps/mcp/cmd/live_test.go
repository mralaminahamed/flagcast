//go:build live

// Live MCP check: spawns the built mcp binary and drives it with the SDK client
// against a running gateway. Run with:
//
//	go build -o /tmp/flagcast-mcp ./apps/mcp/cmd
//	FLAGCAST_MCP_BIN=/tmp/flagcast-mcp FLAGCAST_API=http://127.0.0.1:8201 \
//	  go test -tags=live -run TestMCPLive -v ./apps/mcp/...
package main

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMCPLive(t *testing.T) {
	bin := os.Getenv("FLAGCAST_MCP_BIN")
	if bin == "" {
		t.Skip("set FLAGCAST_MCP_BIN to the built mcp binary")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := exec.Command(bin)
	cmd.Env = append(os.Environ(), "FLAGCAST_API="+envOr("FLAGCAST_API", "http://127.0.0.1:8201"))
	client := mcp.NewClient(&mcp.Implementation{Name: "probe", Version: "1"}, nil)
	session, err := client.Connect(ctx, &mcp.CommandTransport{Command: cmd}, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer session.Close()

	tools, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}
	names := map[string]bool{}
	for _, tl := range tools.Tools {
		names[tl.Name] = true
	}
	for _, want := range []string{"list_flags", "get_flag", "create_flag", "set_enabled", "set_rollout", "delete_flag"} {
		if !names[want] {
			t.Errorf("missing tool %q", want)
		}
	}
	t.Logf("tools: %v", names)

	res, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "list_flags", Arguments: map[string]any{}})
	if err != nil {
		t.Fatalf("call list_flags: %v", err)
	}
	if len(res.Content) == 0 {
		t.Fatal("list_flags returned no content")
	}
	if tc, ok := res.Content[0].(*mcp.TextContent); ok {
		t.Logf("list_flags -> %.200s", tc.Text)
	}
}
