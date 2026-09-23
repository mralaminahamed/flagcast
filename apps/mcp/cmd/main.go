// Command mcp is a stdio MCP server exposing flagcast's flags as tools an agent
// (e.g. Claude) can call. Every tool goes through the gateway REST API, so the
// same auth and validation apply as for a human using the console.
//
// Config: FLAGCAST_API (default http://localhost:8201), FLAGCAST_API_KEY.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/mralaminahamed/flagcast/packages/shared/logger"
	"github.com/mralaminahamed/flagcast/packages/shared/models"
)

type gw struct {
	base   string
	key    string
	client *http.Client
}

func (g *gw) do(ctx context.Context, method, path string, body any) ([]byte, error) {
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, g.base+path, r)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Actor", "mcp")
	if g.key != "" {
		req.Header.Set("X-API-Key", g.key)
	}
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("gateway %d: %s", resp.StatusCode, string(data))
	}
	return data, nil
}

// flagInput is the create/update body the gateway expects.
type flagInput struct {
	Key         string   `json:"key,omitempty"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Enabled     bool     `json:"enabled"`
	Rollout     int      `json:"rollout"`
	Tags        []string `json:"tags,omitempty"`
}

func (g *gw) getFlag(ctx context.Context, key string) (models.Flag, error) {
	data, err := g.do(ctx, http.MethodGet, "/api/flags/"+url.PathEscape(key), nil)
	if err != nil {
		return models.Flag{}, err
	}
	var f models.Flag
	return f, json.Unmarshal(data, &f)
}

// Tool argument types — the `jsonschema` tags become the tool's input schema.
type noArgs struct{}
type keyArgs struct {
	Key string `json:"key" jsonschema:"the flag key"`
}
type createArgs struct {
	Key     string `json:"key" jsonschema:"unique flag key, lowercase with dashes (e.g. new-checkout)"`
	Name    string `json:"name" jsonschema:"human-readable name"`
	Enabled bool   `json:"enabled" jsonschema:"whether the flag starts on"`
	Rollout int    `json:"rollout" jsonschema:"percentage rollout, 0-100"`
}
type enabledArgs struct {
	Key     string `json:"key" jsonschema:"the flag key"`
	Enabled bool   `json:"enabled" jsonschema:"true to turn the flag on, false to turn it off"`
}
type rolloutArgs struct {
	Key     string `json:"key" jsonschema:"the flag key"`
	Rollout int    `json:"rollout" jsonschema:"percentage rollout, 0-100"`
}

func textResult(v any) (*mcp.CallToolResult, any, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}, nil, nil
}

func rawResult(b []byte) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}, nil, nil
}

func main() {
	// stdout carries the MCP protocol — logs must go to stderr.
	logger.InitLogger(logger.LoggerOptions{Level: envOr("LOG_LEVEL", "info"), Stderr: true})
	g := &gw{
		base:   envOr("FLAGCAST_API", "http://localhost:8201"),
		key:    os.Getenv("FLAGCAST_API_KEY"),
		client: &http.Client{Timeout: 15 * time.Second},
	}

	server := mcp.NewServer(&mcp.Implementation{Name: "flagcast", Version: "0.1.0"}, nil)

	mcp.AddTool(server, &mcp.Tool{Name: "list_flags", Description: "List all feature flags with their state and rollout."},
		func(ctx context.Context, _ *mcp.CallToolRequest, _ noArgs) (*mcp.CallToolResult, any, error) {
			data, err := g.do(ctx, http.MethodGet, "/api/flags", nil)
			if err != nil {
				return nil, nil, err
			}
			return rawResult(data)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "get_flag", Description: "Get one feature flag by key."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in keyArgs) (*mcp.CallToolResult, any, error) {
			data, err := g.do(ctx, http.MethodGet, "/api/flags/"+url.PathEscape(in.Key), nil)
			if err != nil {
				return nil, nil, err
			}
			return rawResult(data)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "create_flag", Description: "Create a new feature flag."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in createArgs) (*mcp.CallToolResult, any, error) {
			data, err := g.do(ctx, http.MethodPost, "/api/flags", flagInput{
				Key: in.Key, Name: in.Name, Enabled: in.Enabled, Rollout: in.Rollout,
			})
			if err != nil {
				return nil, nil, err
			}
			return rawResult(data)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "set_enabled", Description: "Turn a feature flag on or off."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in enabledArgs) (*mcp.CallToolResult, any, error) {
			f, err := g.getFlag(ctx, in.Key)
			if err != nil {
				return nil, nil, err
			}
			data, err := g.do(ctx, http.MethodPut, "/api/flags/"+url.PathEscape(in.Key), flagInput{
				Name: f.Name, Description: f.Description, Enabled: in.Enabled, Rollout: f.Rollout, Tags: f.Tags,
			})
			if err != nil {
				return nil, nil, err
			}
			return rawResult(data)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "set_rollout", Description: "Set a feature flag's percentage rollout (0-100)."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in rolloutArgs) (*mcp.CallToolResult, any, error) {
			f, err := g.getFlag(ctx, in.Key)
			if err != nil {
				return nil, nil, err
			}
			data, err := g.do(ctx, http.MethodPut, "/api/flags/"+url.PathEscape(in.Key), flagInput{
				Name: f.Name, Description: f.Description, Enabled: f.Enabled, Rollout: in.Rollout, Tags: f.Tags,
			})
			if err != nil {
				return nil, nil, err
			}
			return rawResult(data)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "delete_flag", Description: "Delete a feature flag."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in keyArgs) (*mcp.CallToolResult, any, error) {
			if _, err := g.do(ctx, http.MethodDelete, "/api/flags/"+url.PathEscape(in.Key), nil); err != nil {
				return nil, nil, err
			}
			return textResult(map[string]string{"status": "deleted", "key": in.Key})
		})

	logger.Log.Info().Str("api", g.base).Msg("flagcast MCP server ready (stdio)")
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		logger.Log.Fatal().Err(err).Msg("mcp server stopped")
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
