package server

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/reponerve/reponerve/internal/mcp"
)

func TestDevelopmentTool_NotConfigured(t *testing.T) {
	registry := mcp.NewRegistry()
	svc := mcp.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	srv := NewServer(registry, svc, strings.NewReader(""), &strings.Builder{})

	req := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"ask","arguments":{"question":"why sqlite?"}}}` + "\n"
	output := runServerTest(t, registry, svc, req)

	var resp JSONRPCResponse
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("unexpected rpc error: %+v", resp.Error)
	}

	result := parseToolResult(t, resp.Result)
	if !result.IsError {
		t.Fatal("expected tool error when development service is not configured")
	}
	if !strings.Contains(result.Content[0].Text, "not configured") {
		t.Fatalf("expected not configured error, got: %q", result.Content[0].Text)
	}

	_ = srv
}

func TestDevelopmentTool_RegisteredInList(t *testing.T) {
	registry := mcp.NewRegistry()
	names := make(map[string]bool)
	for _, tool := range registry.List() {
		names[tool.Name] = true
	}
	for _, name := range []string{
		"ask", "explain", "explain_file", "plan", "review", "analyze_topic_impact",
	} {
		if !names[name] {
			t.Errorf("expected tool %q to be registered", name)
		}
	}
}
