package server

import (
	"bytes"
	stdcontext "context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/reponerve/reponerve/internal/context"
	"github.com/reponerve/reponerve/internal/context/render"
	"github.com/reponerve/reponerve/internal/graph/impact"
	"github.com/reponerve/reponerve/internal/graph/model"
	"github.com/reponerve/reponerve/internal/graph/relationships"
	"github.com/reponerve/reponerve/internal/graph/traversal"
	"github.com/reponerve/reponerve/internal/mcp"
	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
	ownershipquery "github.com/reponerve/reponerve/internal/ownership/query"
	"github.com/reponerve/reponerve/internal/query/storage"
	"github.com/reponerve/reponerve/internal/storage/migrations"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
	models "github.com/reponerve/reponerve/pkg/models"
)

func testContributorID(repositoryID, email string) string {
	h := sha256.Sum256([]byte(repositoryID + ":" + email))
	return "ctr_" + hex.EncodeToString(h[:])
}

func runServerTest(t *testing.T, registry *mcp.Registry, service *mcp.Service, input string) string {
	inReader, inWriter := io.Pipe()
	outReader, outWriter := io.Pipe()

	ctx, cancel := stdcontext.WithCancel(stdcontext.Background())
	defer cancel()

	srv := NewServer(registry, service, inReader, outWriter)

	errChan := make(chan error, 1)
	go func() {
		errChan <- srv.Start(ctx)
	}()

	// Write input in a separate goroutine
	go func() {
		_, _ = inWriter.Write([]byte(input))
		_ = inWriter.Close()
	}()

	// Read all responses until EOF
	var outBuf bytes.Buffer
	doneChan := make(chan struct{})
	go func() {
		_, _ = io.Copy(&outBuf, outReader)
		close(doneChan)
	}()

	// Wait for server to finish
	select {
	case err := <-errChan:
		if err != nil && err != io.EOF {
			t.Logf("server returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for server to exit")
	}

	// Close pipes
	_ = outReader.Close()
	_ = outWriter.Close()
	<-doneChan

	return outBuf.String()
}

func parseToolResult(t *testing.T, result interface{}) CallToolResult {
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal result: %v", err)
	}
	var callResult CallToolResult
	if err := json.Unmarshal(data, &callResult); err != nil {
		t.Fatalf("failed to unmarshal CallToolResult: %v", err)
	}
	return callResult
}

func TestServer_JSONRPC(t *testing.T) {
	registry := mcp.NewRegistry()

	t.Run("Initialize handshake", func(t *testing.T) {
		req := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05"}}` + "\n"
		output := runServerTest(t, registry, nil, req)

		var resp JSONRPCResponse
		if err := json.Unmarshal([]byte(output), &resp); err != nil {
			t.Fatalf("failed to parse response: %v, output was: %q", err, output)
		}

		if resp.JSONRPC != "2.0" {
			t.Errorf("expected jsonrpc '2.0', got %q", resp.JSONRPC)
		}
		if resp.Error != nil {
			t.Fatalf("unexpected error response: %+v", resp.Error)
		}

		var id int
		if err := json.Unmarshal(*resp.ID, &id); err != nil {
			t.Fatalf("failed to parse response ID: %v", err)
		}
		if id != 1 {
			t.Errorf("expected ID 1, got %d", id)
		}

		// Verify result
		resultData, err := json.Marshal(resp.Result)
		if err != nil {
			t.Fatalf("failed to marshal result: %v", err)
		}
		var result InitializeResult
		if err := json.Unmarshal(resultData, &result); err != nil {
			t.Fatalf("failed to unmarshal initialize result: %v", err)
		}

		if result.ProtocolVersion != "2024-11-05" {
			t.Errorf("expected protocolVersion '2024-11-05', got %q", result.ProtocolVersion)
		}
		if result.ServerInfo.Name != "reponerve" {
			t.Errorf("expected server name 'reponerve', got %q", result.ServerInfo.Name)
		}
	})

	t.Run("Tools list discovery", func(t *testing.T) {
		req := `{"jsonrpc":"2.0","id":"req-2","method":"tools/list"}` + "\n"
		output := runServerTest(t, registry, nil, req)

		var resp JSONRPCResponse
		if err := json.Unmarshal([]byte(output), &resp); err != nil {
			t.Fatalf("failed to parse response: %v, output was: %q", err, output)
		}

		if resp.Error != nil {
			t.Fatalf("unexpected error response: %+v", resp.Error)
		}

		var id string
		if err := json.Unmarshal(*resp.ID, &id); err != nil {
			t.Fatalf("failed to parse response ID: %v", err)
		}
		if id != "req-2" {
			t.Errorf("expected ID 'req-2', got %q", id)
		}

		// Verify result
		resultData, err := json.Marshal(resp.Result)
		if err != nil {
			t.Fatalf("failed to marshal result: %v", err)
		}
		var result ListToolsResult
		if err := json.Unmarshal(resultData, &result); err != nil {
			t.Fatalf("failed to unmarshal tools list: %v", err)
		}

		if len(result.Tools) != 49 {
			t.Errorf("expected 49 tools, got %d", len(result.Tools))
		}

		expectedTools := map[string]bool{
			"analyze_topic_impact":   true,
			"ask":                    true,
			"explain":                true,
			"explain_decision":       true,
			"explain_feature":        true,
			"explain_file":           true,
			"explain_function":       true,
			"explain_interface":      true,
			"explain_struct":         true,
			"explain_type":           true,
			"onboard":                true,
			"plan":                   true,
			"review":                 true,
			"explain_event":          true,
			"export_context":         true,
			"generate_context":       true,
			"get_contributor":        true,
			"get_decision":           true,
			"get_event":              true,
			"get_fact":               true,
			"get_intent":             true,
			"list_contributors":      true,
			"list_decisions":         true,
			"list_events":            true,
			"list_expertise":         true,
			"list_facts":             true,
			"list_features":          true,
			"list_intents":           true,
			"recommend_reviewers":    true,
			"trace_contributor":      true,
			"trace_decision":         true,
			"trace_event":            true,
			"trace_graph":            true,
			"trace_path":             true,
			"analyze_impact":         true,
			"find_dependencies":      true,
			"find_dependents":        true,
			"discover_knowledge":     true,
			"discover_surprises":     true,
			"doctor":                 true,
			"forget":                 true,
			"query_graph":            true,
			"remember":               true,
			"reuse_check":            true,
			"ship_check":             true,
			"pr_context":             true,
			"suggest_questions":      true,
			"generate_learning_path": true,
			"generate_change_plan":   true,
		}

		for _, tool := range result.Tools {
			if !expectedTools[tool.Name] {
				t.Errorf("unexpected tool %q returned", tool.Name)
			}
			if tool.InputSchema.Type != "object" {
				t.Errorf("expected tool input schema type 'object', got %q", tool.InputSchema.Type)
			}
		}
	})

	t.Run("Unknown method returns method not found", func(t *testing.T) {
		req := `{"jsonrpc":"2.0","id":100,"method":"some_unknown_method"}` + "\n"
		output := runServerTest(t, registry, nil, req)

		var resp JSONRPCResponse
		if err := json.Unmarshal([]byte(output), &resp); err != nil {
			t.Fatalf("failed to parse response: %v, output was: %q", err, output)
		}

		if resp.Error == nil {
			t.Fatal("expected error response, got nil")
		}
		if resp.Error.Code != -32601 {
			t.Errorf("expected error code -32601 (Method not found), got %d", resp.Error.Code)
		}
		if !strings.Contains(resp.Error.Message, "Method not found") {
			t.Errorf("expected error message to contain 'Method not found', got %q", resp.Error.Message)
		}
	})

	t.Run("Notification is ignored (no response)", func(t *testing.T) {
		req := `{"jsonrpc":"2.0","method":"notifications/initialized"}` + "\n"
		output := runServerTest(t, registry, nil, req)

		if len(strings.TrimSpace(output)) != 0 {
			t.Errorf("expected no response output for notification, got %q", output)
		}
	})

	t.Run("Invalid JSON returns parse error", func(t *testing.T) {
		req := `invalid json` + "\n"
		output := runServerTest(t, registry, nil, req)

		var resp JSONRPCResponse
		if err := json.Unmarshal([]byte(output), &resp); err != nil {
			t.Fatalf("failed to parse response: %v, output was: %q", err, output)
		}

		if resp.Error == nil {
			t.Fatal("expected error response, got nil")
		}
		if resp.Error.Code != -32700 {
			t.Errorf("expected error code -32700 (Parse error), got %d", resp.Error.Code)
		}
	})
}

func TestServer_ToolsExecution(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "reponerve-server-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := migrations.RunUp(db); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	repoID := "repo_xxx"

	// Insert test data
	_, err = db.Exec("INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at) VALUES (?, ?, ?, ?, datetime(), datetime())", repoID, "Repo Test", tempDir, "main")
	if err != nil {
		t.Fatalf("failed to insert repository: %v", err)
	}

	_, err = db.Exec("INSERT INTO sources (id, repository_id, source_type, reference, title, author, timestamp, created_at) VALUES (?, ?, ?, ?, ?, ?, datetime(), datetime())", "src_1", repoID, "adr", "docs/adr/0001.md", "Author", "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("failed to insert source: %v", err)
	}

	// Insert decision
	_, err = db.Exec("INSERT INTO memory_decisions (id, repository_id, source_id, title, status, created_at) VALUES (?, ?, ?, ?, ?, datetime())", "dec_1", repoID, "src_1", "Use Redis Cache", "Accepted")
	if err != nil {
		t.Fatalf("failed to insert decision: %v", err)
	}

	// Insert intent
	_, err = db.Exec("INSERT INTO memory_intents (id, repository_id, source_id, description, created_at) VALUES (?, ?, ?, ?, datetime())", "int_1", repoID, "src_1", "Reduce Database Latency")
	if err != nil {
		t.Fatalf("failed to insert intent: %v", err)
	}

	// Insert fact
	_, err = db.Exec("INSERT INTO memory_facts (id, repository_id, source_id, subject, predicate, object, created_at) VALUES (?, ?, ?, ?, ?, ?, datetime())", "fact_1", repoID, "src_1", "Authentication Service", "USES", "Redis")
	if err != nil {
		t.Fatalf("failed to insert fact: %v", err)
	}

	// Insert event
	_, err = db.Exec("INSERT INTO memory_events (id, repository_id, event_type, title, description, source_id, timestamp, created_at) VALUES (?, ?, ?, ?, ?, ?, datetime(), datetime())", "evt_1", repoID, "FEATURE_INTRODUCED", "Introduce Redis Cache", "Added cache capability", "src_1")
	if err != nil {
		t.Fatalf("failed to insert event: %v", err)
	}

	// Insert relationships
	_, err = db.Exec("INSERT INTO memory_relationships (id, repository_id, from_id, to_id, relationship_type, created_at) VALUES (?, ?, ?, ?, ?, datetime())", "rel_1", repoID, "int_1", "dec_1", "INTENT_DRIVES_DECISION")
	if err != nil {
		t.Fatalf("failed to insert relationship 1: %v", err)
	}
	_, err = db.Exec("INSERT INTO memory_relationships (id, repository_id, from_id, to_id, relationship_type, created_at) VALUES (?, ?, ?, ?, ?, datetime())", "rel_2", repoID, "fact_1", "dec_1", "FACT_SUPPORTS_DECISION")
	if err != nil {
		t.Fatalf("failed to insert relationship 2: %v", err)
	}
	_, err = db.Exec("INSERT INTO memory_relationships (id, repository_id, from_id, to_id, relationship_type, created_at) VALUES (?, ?, ?, ?, ?, datetime())", "rel_3", repoID, "dec_1", "evt_1", "DECISION_RESULTS_IN_EVENT")
	if err != nil {
		t.Fatalf("failed to insert relationship 3: %v", err)
	}

	// Set up Service
	er := storage.NewSQLiteEventReader(db)
	dr := storage.NewSQLiteDecisionReader(db)
	ir := storage.NewSQLiteIntentReader(db)
	fr := storage.NewSQLiteFactReader(db)
	rr := storage.NewSQLiteRelationshipReader(db)
	cr := storage.NewSQLiteContributorReader(db)
	expr := storage.NewSQLiteExpertiseReader(db)
	sr := storage.NewSQLiteSourceReader(db)

	ctxReader := context.NewMemoryContextReader(er, dr, ir, fr)
	generator := context.NewGenerator(ctxReader)
	renderer := render.NewRenderer()
	ownershipReader := ownershipquery.NewReader(cr, expr, sr, dr, fr, er)

	service := mcp.NewService(dr, ir, fr, er, rr, generator, renderer, ownershipReader, nil, nil, nil, nil, nil, nil, nil)
	registry := mcp.NewRegistry()

	// 1. Test list_decisions (all)
	t.Run("list_decisions all", func(t *testing.T) {
		req := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"list_decisions"}}` + "\n"
		output := runServerTest(t, registry, service, req)

		var resp JSONRPCResponse
		if err := json.Unmarshal([]byte(output), &resp); err != nil {
			t.Fatalf("failed to parse response: %v, output: %s", err, output)
		}
		if resp.Error != nil {
			t.Fatalf("unexpected error response: %+v", resp.Error)
		}

		result := parseToolResult(t, resp.Result)
		if result.IsError {
			t.Fatalf("unexpected tool error: %s", result.Content[0].Text)
		}

		var decisions []*memorymodels.Decision
		if err := json.Unmarshal([]byte(result.Content[0].Text), &decisions); err != nil {
			t.Fatalf("failed to parse decisions: %v", err)
		}
		if len(decisions) != 1 {
			t.Errorf("expected 1 decision, got %d", len(decisions))
		}
		if decisions[0].Title != "Use Redis Cache" {
			t.Errorf("expected decision title 'Use Redis Cache', got %q", decisions[0].Title)
		}
	})

	// 2. Test get_decision
	t.Run("get_decision success", func(t *testing.T) {
		req := `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"get_decision","arguments":{"decision_id":"dec_1"}}}` + "\n"
		output := runServerTest(t, registry, service, req)

		var resp JSONRPCResponse
		if err := json.Unmarshal([]byte(output), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		result := parseToolResult(t, resp.Result)
		if result.IsError {
			t.Fatalf("unexpected tool error: %s", result.Content[0].Text)
		}

		var dec memorymodels.Decision
		if err := json.Unmarshal([]byte(result.Content[0].Text), &dec); err != nil {
			t.Fatalf("failed to parse decision: %v", err)
		}
		if dec.ID != "dec_1" || dec.Title != "Use Redis Cache" {
			t.Errorf("unexpected decision properties: %+v", dec)
		}
	})

	t.Run("get_decision not found", func(t *testing.T) {
		req := `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"get_decision","arguments":{"decision_id":"nonexistent"}}}` + "\n"
		output := runServerTest(t, registry, service, req)

		var resp JSONRPCResponse
		if err := json.Unmarshal([]byte(output), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		result := parseToolResult(t, resp.Result)
		if !result.IsError {
			t.Fatalf("expected tool error but got success")
		}

		text := result.Content[0].Text
		if !strings.Contains(text, "not found") {
			t.Errorf("expected error message to contain 'not found', got %q", text)
		}
	})

	// 3. Test list_events with repository_id filter
	t.Run("list_events with repository_id", func(t *testing.T) {
		req := `{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"list_events","arguments":{"repository_id":"repo_xxx"}}}` + "\n"
		output := runServerTest(t, registry, service, req)

		var resp JSONRPCResponse
		if err := json.Unmarshal([]byte(output), &resp); err != nil {
			t.Fatalf("failed to parse: %v", err)
		}

		result := parseToolResult(t, resp.Result)
		if result.IsError {
			t.Fatalf("unexpected tool error: %s", result.Content[0].Text)
		}

		var events []*models.Event
		if err := json.Unmarshal([]byte(result.Content[0].Text), &events); err != nil {
			t.Fatalf("failed to parse events: %v", err)
		}
		if len(events) != 1 {
			t.Errorf("expected 1 event, got %d", len(events))
		}
		if events[0].Title != "Introduce Redis Cache" {
			t.Errorf("unexpected event title: %q", events[0].Title)
		}
	})

	// 4. Test trace_decision
	t.Run("trace_decision success", func(t *testing.T) {
		req := `{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"trace_decision","arguments":{"decision_id":"dec_1"}}}` + "\n"
		output := runServerTest(t, registry, service, req)

		var resp JSONRPCResponse
		if err := json.Unmarshal([]byte(output), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		result := parseToolResult(t, resp.Result)
		if result.IsError {
			t.Fatalf("unexpected tool error: %s", result.Content[0].Text)
		}

		var trace DecisionTraceResult
		if err := json.Unmarshal([]byte(result.Content[0].Text), &trace); err != nil {
			t.Fatalf("failed to parse decision trace: %v", err)
		}

		if trace.Decision.ID != "dec_1" {
			t.Errorf("expected decision ID 'dec_1', got %q", trace.Decision.ID)
		}
		if len(trace.Intents) != 1 || trace.Intents[0].ID != "int_1" {
			t.Errorf("expected 1 intent 'int_1', got: %+v", trace.Intents)
		}
		if len(trace.Facts) != 1 || trace.Facts[0].ID != "fact_1" {
			t.Errorf("expected 1 fact 'fact_1', got: %+v", trace.Facts)
		}
		if len(trace.Events) != 1 || trace.Events[0].ID != "evt_1" {
			t.Errorf("expected 1 event 'evt_1', got: %+v", trace.Events)
		}
	})

	// 5. Test trace_event
	t.Run("trace_event success", func(t *testing.T) {
		req := `{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"trace_event","arguments":{"event_id":"evt_1"}}}` + "\n"
		output := runServerTest(t, registry, service, req)

		var resp JSONRPCResponse
		if err := json.Unmarshal([]byte(output), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		result := parseToolResult(t, resp.Result)
		if result.IsError {
			t.Fatalf("unexpected tool error: %s", result.Content[0].Text)
		}

		var trace EventTraceResult
		if err := json.Unmarshal([]byte(result.Content[0].Text), &trace); err != nil {
			t.Fatalf("failed to parse event trace: %v", err)
		}

		if trace.Event.ID != "evt_1" {
			t.Errorf("expected event ID 'evt_1', got %q", trace.Event.ID)
		}
		if len(trace.Decisions) != 1 || trace.Decisions[0].ID != "dec_1" {
			t.Errorf("expected 1 decision 'dec_1', got: %+v", trace.Decisions)
		}
		if len(trace.Intents) != 1 || trace.Intents[0].ID != "int_1" {
			t.Errorf("expected 1 intent 'int_1', got: %+v", trace.Intents)
		}
	})

	// 6. Test explain_decision
	t.Run("explain_decision success", func(t *testing.T) {
		req := `{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"explain_decision","arguments":{"decision_id":"dec_1"}}}` + "\n"
		output := runServerTest(t, registry, service, req)

		var resp JSONRPCResponse
		if err := json.Unmarshal([]byte(output), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		result := parseToolResult(t, resp.Result)
		if result.IsError {
			t.Fatalf("unexpected tool error: %s", result.Content[0].Text)
		}

		var explanation DecisionExplanationResult
		if err := json.Unmarshal([]byte(result.Content[0].Text), &explanation); err != nil {
			t.Fatalf("failed to parse decision explanation: %v", err)
		}

		if explanation.Decision.ID != "dec_1" {
			t.Errorf("expected decision ID 'dec_1', got %q", explanation.Decision.ID)
		}
		if len(explanation.Reason) != 1 || explanation.Reason[0] != "Reduce Database Latency" {
			t.Errorf("expected reason 'Reduce Database Latency', got: %v", explanation.Reason)
		}
		if len(explanation.SupportingFacts) != 1 || explanation.SupportingFacts[0] != "Authentication Service USES Redis" {
			t.Errorf("expected fact 'Authentication Service USES Redis', got: %v", explanation.SupportingFacts)
		}
		if len(explanation.ResultingEvents) != 1 || explanation.ResultingEvents[0] != "Introduce Redis Cache" {
			t.Errorf("expected event 'Introduce Redis Cache', got: %v", explanation.ResultingEvents)
		}
	})

	// 7. Test explain_event
	t.Run("explain_event success", func(t *testing.T) {
		req := `{"jsonrpc":"2.0","id":8,"method":"tools/call","params":{"name":"explain_event","arguments":{"event_id":"evt_1"}}}` + "\n"
		output := runServerTest(t, registry, service, req)

		var resp JSONRPCResponse
		if err := json.Unmarshal([]byte(output), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		result := parseToolResult(t, resp.Result)
		if result.IsError {
			t.Fatalf("unexpected tool error: %s", result.Content[0].Text)
		}

		var explanation EventExplanationResult
		if err := json.Unmarshal([]byte(result.Content[0].Text), &explanation); err != nil {
			t.Fatalf("failed to parse event explanation: %v", err)
		}

		if explanation.Event.ID != "evt_1" {
			t.Errorf("expected event ID 'evt_1', got %q", explanation.Event.ID)
		}
		if len(explanation.CausedBy) != 1 || explanation.CausedBy[0] != "Use Redis Cache" {
			t.Errorf("expected causedBy 'Use Redis Cache', got: %v", explanation.CausedBy)
		}
		if len(explanation.Reason) != 1 || explanation.Reason[0] != "Reduce Database Latency" {
			t.Errorf("expected reason 'Reduce Database Latency', got: %v", explanation.Reason)
		}
	})

	// 8. Test list_intents, list_facts, get_intent, get_fact
	t.Run("list_intents, list_facts, get_intent, get_fact success", func(t *testing.T) {
		// list_intents
		req := `{"jsonrpc":"2.0","id":9,"method":"tools/call","params":{"name":"list_intents"}}` + "\n"
		output := runServerTest(t, registry, service, req)
		var resp JSONRPCResponse
		_ = json.Unmarshal([]byte(output), &resp)
		result := parseToolResult(t, resp.Result)
		text := result.Content[0].Text
		var intents []*memorymodels.Intent
		_ = json.Unmarshal([]byte(text), &intents)
		if len(intents) != 1 || intents[0].ID != "int_1" {
			t.Errorf("unexpected intents: %+v", intents)
		}

		// get_intent
		req = `{"jsonrpc":"2.0","id":10,"method":"tools/call","params":{"name":"get_intent","arguments":{"intent_id":"int_1"}}}` + "\n"
		output = runServerTest(t, registry, service, req)
		_ = json.Unmarshal([]byte(output), &resp)
		result = parseToolResult(t, resp.Result)
		text = result.Content[0].Text
		var intent memorymodels.Intent
		_ = json.Unmarshal([]byte(text), &intent)
		if intent.ID != "int_1" {
			t.Errorf("unexpected intent: %+v", intent)
		}

		// list_facts
		req = `{"jsonrpc":"2.0","id":11,"method":"tools/call","params":{"name":"list_facts"}}` + "\n"
		output = runServerTest(t, registry, service, req)
		_ = json.Unmarshal([]byte(output), &resp)
		result = parseToolResult(t, resp.Result)
		text = result.Content[0].Text
		var facts []*memorymodels.Fact
		_ = json.Unmarshal([]byte(text), &facts)
		if len(facts) != 1 || facts[0].ID != "fact_1" {
			t.Errorf("unexpected facts: %+v", facts)
		}

		// get_fact
		req = `{"jsonrpc":"2.0","id":12,"method":"tools/call","params":{"name":"get_fact","arguments":{"fact_id":"fact_1"}}}` + "\n"
		output = runServerTest(t, registry, service, req)
		_ = json.Unmarshal([]byte(output), &resp)
		result = parseToolResult(t, resp.Result)
		text = result.Content[0].Text
		var fact memorymodels.Fact
		_ = json.Unmarshal([]byte(text), &fact)
		if fact.ID != "fact_1" {
			t.Errorf("unexpected fact: %+v", fact)
		}
	})

	// 9. Test missing required argument
	t.Run("missing required argument error", func(t *testing.T) {
		req := `{"jsonrpc":"2.0","id":13,"method":"tools/call","params":{"name":"get_decision","arguments":{}}}` + "\n"
		output := runServerTest(t, registry, service, req)

		var resp JSONRPCResponse
		if err := json.Unmarshal([]byte(output), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		result := parseToolResult(t, resp.Result)
		if !result.IsError {
			t.Fatalf("expected error for missing required argument")
		}

		text := result.Content[0].Text
		if !strings.Contains(text, "missing required argument") {
			t.Errorf("expected error message to contain 'missing required argument', got %q", text)
		}
	})

	// 10. Test generate_context success
	t.Run("generate_context success", func(t *testing.T) {
		req := `{"jsonrpc":"2.0","id":14,"method":"tools/call","params":{"name":"generate_context","arguments":{"repository_id":"repo_xxx"}}}` + "\n"
		output := runServerTest(t, registry, service, req)

		var resp JSONRPCResponse
		if err := json.Unmarshal([]byte(output), &resp); err != nil {
			t.Fatalf("failed to parse response: %v, output: %s", err, output)
		}

		result := parseToolResult(t, resp.Result)
		if result.IsError {
			t.Fatalf("unexpected tool error: %s", result.Content[0].Text)
		}

		var rc context.RepositoryContext
		if err := json.Unmarshal([]byte(result.Content[0].Text), &rc); err != nil {
			t.Fatalf("failed to parse RepositoryContext: %v", err)
		}

		if rc.RepositoryID != "repo_xxx" {
			t.Errorf("expected repository ID 'repo_xxx', got %q", rc.RepositoryID)
		}
		if len(rc.Decisions) != 1 || rc.Decisions[0].ID != "dec_1" {
			t.Errorf("expected 1 decision 'dec_1'")
		}
		if len(rc.Intents) != 1 || rc.Intents[0].ID != "int_1" {
			t.Errorf("expected 1 intent 'int_1'")
		}
		if len(rc.Facts) != 1 || rc.Facts[0].ID != "fact_1" {
			t.Errorf("expected 1 fact 'fact_1'")
		}
		if len(rc.Events) != 1 || rc.Events[0].ID != "evt_1" {
			t.Errorf("expected 1 event 'evt_1'")
		}
	})

	// 11. Test export_context success
	t.Run("export_context success", func(t *testing.T) {
		req := `{"jsonrpc":"2.0","id":15,"method":"tools/call","params":{"name":"export_context","arguments":{"repository_id":"repo_xxx"}}}` + "\n"
		output := runServerTest(t, registry, service, req)

		var resp JSONRPCResponse
		if err := json.Unmarshal([]byte(output), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		result := parseToolResult(t, resp.Result)
		if result.IsError {
			t.Fatalf("unexpected tool error: %s", result.Content[0].Text)
		}

		text := result.Content[0].Text
		if !strings.Contains(text, "# Repository Context") {
			t.Errorf("expected header '# Repository Context' in markdown, got %q", text)
		}
		if !strings.Contains(text, "Repository: repo_xxx") {
			t.Errorf("expected repository ID in markdown, got %q", text)
		}
		if !strings.Contains(text, "## Key Decisions") || !strings.Contains(text, "* Use Redis Cache") {
			t.Errorf("expected Decisions section, got %q", text)
		}
	})

	// 12. Test generate_context empty repo
	t.Run("generate_context empty repo", func(t *testing.T) {
		req := `{"jsonrpc":"2.0","id":16,"method":"tools/call","params":{"name":"generate_context","arguments":{"repository_id":"empty_repo"}}}` + "\n"
		output := runServerTest(t, registry, service, req)

		var resp JSONRPCResponse
		if err := json.Unmarshal([]byte(output), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		result := parseToolResult(t, resp.Result)
		if result.IsError {
			t.Fatalf("unexpected tool error: %s", result.Content[0].Text)
		}

		var rc context.RepositoryContext
		if err := json.Unmarshal([]byte(result.Content[0].Text), &rc); err != nil {
			t.Fatalf("failed to parse RepositoryContext: %v", err)
		}

		if len(rc.Decisions) != 0 || len(rc.Intents) != 0 || len(rc.Facts) != 0 || len(rc.Events) != 0 {
			t.Errorf("expected empty context lists, got: %+v", rc)
		}
	})

	// 13. Test export_context empty repo
	t.Run("export_context empty repo", func(t *testing.T) {
		req := `{"jsonrpc":"2.0","id":17,"method":"tools/call","params":{"name":"export_context","arguments":{"repository_id":"empty_repo"}}}` + "\n"
		output := runServerTest(t, registry, service, req)

		var resp JSONRPCResponse
		if err := json.Unmarshal([]byte(output), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		result := parseToolResult(t, resp.Result)
		if result.IsError {
			t.Fatalf("unexpected tool error: %s", result.Content[0].Text)
		}

		text := result.Content[0].Text
		if text != "No repository context available." {
			t.Errorf("expected empty context text, got %q", text)
		}
	})
}

func TestServer_OwnershipToolsExecution(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "reponerve-ownership-server-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := migrations.RunUp(db); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	repoID := "repo_xxx"

	// Insert test data
	_, err = db.Exec("INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at) VALUES (?, ?, ?, ?, datetime(), datetime())", repoID, "Repo Test", tempDir, "main")
	if err != nil {
		t.Fatalf("failed to insert repository: %v", err)
	}

	janeID := testContributorID(repoID, "jane@example.com")
	johnID := testContributorID(repoID, "john@example.com")
	aliceID := testContributorID(repoID, "alice@example.com")

	// Insert contributors
	_, err = db.Exec(`INSERT INTO contributors (id, repository_id, name, email, first_seen, last_seen, commit_count) 
		VALUES (?, ?, ?, ?, datetime(), datetime(), ?)`, janeID, repoID, "Jane Doe", "jane@example.com", 5)
	if err != nil {
		t.Fatalf("failed to insert contributor: %v", err)
	}
	_, err = db.Exec(`INSERT INTO contributors (id, repository_id, name, email, first_seen, last_seen, commit_count) 
		VALUES (?, ?, ?, ?, datetime(), datetime(), ?)`, johnID, repoID, "John Smith", "john@example.com", 10)
	if err != nil {
		t.Fatalf("failed to insert contributor 2: %v", err)
	}
	_, err = db.Exec(`INSERT INTO contributors (id, repository_id, name, email, first_seen, last_seen, commit_count) 
		VALUES (?, ?, ?, ?, datetime(), datetime(), ?)`, aliceID, repoID, "Alice", "alice@example.com", 3)
	if err != nil {
		t.Fatalf("failed to insert contributor 3: %v", err)
	}

	// Insert expertise
	_, err = db.Exec(`INSERT INTO expertise (id, repository_id, contributor_id, domain, score, evidence_json) 
		VALUES (?, ?, ?, ?, ?, ?)`, "exp_1", repoID, janeID, "Storage", 0.95, `{"recent_activity": true}`)
	if err != nil {
		t.Fatalf("failed to insert expertise: %v", err)
	}
	_, err = db.Exec(`INSERT INTO expertise (id, repository_id, contributor_id, domain, score, evidence_json) 
		VALUES (?, ?, ?, ?, ?, ?)`, "exp_2", repoID, johnID, "Storage", 0.95, `{"recent_activity": false}`)
	if err != nil {
		t.Fatalf("failed to insert expertise 2: %v", err)
	}
	_, err = db.Exec(`INSERT INTO expertise (id, repository_id, contributor_id, domain, score, evidence_json) 
		VALUES (?, ?, ?, ?, ?, ?)`, "exp_3", repoID, aliceID, "Storage", 0.90, `{"recent_activity": true}`)
	if err != nil {
		t.Fatalf("failed to insert expertise 3: %v", err)
	}

	// Insert source, decision, fact, event for Jane Doe to test trace_contributor
	_, err = db.Exec("INSERT INTO sources (id, repository_id, source_type, reference, title, author, timestamp, created_at) VALUES (?, ?, ?, ?, ?, ?, datetime(), datetime())", "src_jane", repoID, "git", "commit_jane", "Title Jane", "Jane Doe <jane@example.com>")
	if err != nil {
		t.Fatalf("failed to insert source jane: %v", err)
	}
	_, err = db.Exec("INSERT INTO memory_decisions (id, repository_id, source_id, title, status, created_at) VALUES (?, ?, ?, ?, ?, datetime())", "dec_jane", repoID, "src_jane", "Jane's Storage Decision", "Accepted")
	if err != nil {
		t.Fatalf("failed to insert decision jane: %v", err)
	}
	_, err = db.Exec("INSERT INTO memory_facts (id, repository_id, source_id, subject, predicate, object, created_at) VALUES (?, ?, ?, ?, ?, ?, datetime())", "fact_jane", repoID, "src_jane", "Jane Storage", "WORKS", "Fine")
	if err != nil {
		t.Fatalf("failed to insert fact jane: %v", err)
	}
	_, err = db.Exec("INSERT INTO memory_events (id, repository_id, event_type, title, description, source_id, timestamp, created_at) VALUES (?, ?, ?, ?, ?, ?, datetime(), datetime())", "evt_jane", repoID, "COMMIT_COMMITTED", "Jane's Event", "Description Jane", "src_jane")
	if err != nil {
		t.Fatalf("failed to insert event jane: %v", err)
	}

	// Set up Service
	er := storage.NewSQLiteEventReader(db)
	dr := storage.NewSQLiteDecisionReader(db)
	ir := storage.NewSQLiteIntentReader(db)
	fr := storage.NewSQLiteFactReader(db)
	rr := storage.NewSQLiteRelationshipReader(db)
	cr := storage.NewSQLiteContributorReader(db)
	expr := storage.NewSQLiteExpertiseReader(db)
	sr := storage.NewSQLiteSourceReader(db)

	ctxReader := context.NewMemoryContextReader(er, dr, ir, fr)
	generator := context.NewGenerator(ctxReader)
	renderer := render.NewRenderer()
	ownershipReader := ownershipquery.NewReader(cr, expr, sr, dr, fr, er)

	service := mcp.NewService(dr, ir, fr, er, rr, generator, renderer, ownershipReader, nil, nil, nil, nil, nil, nil, nil)
	registry := mcp.NewRegistry()

	// 1. Test list_contributors
	t.Run("list_contributors success", func(t *testing.T) {
		req := `{"jsonrpc":"2.0","id":18,"method":"tools/call","params":{"name":"list_contributors","arguments":{"repository_id":"repo_xxx"}}}` + "\n"
		output := runServerTest(t, registry, service, req)

		var resp JSONRPCResponse
		if err := json.Unmarshal([]byte(output), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		result := parseToolResult(t, resp.Result)
		if result.IsError {
			t.Fatalf("unexpected tool error: %s", result.Content[0].Text)
		}

		var contributors []*models.Contributor
		if err := json.Unmarshal([]byte(result.Content[0].Text), &contributors); err != nil {
			t.Fatalf("failed to parse contributors: %v", err)
		}
		// Expect 3 contributors (Alice, Jane Doe, John Smith) sorted by Name: Alice, Jane Doe, John Smith
		if len(contributors) != 3 {
			t.Errorf("expected 3 contributors, got %d", len(contributors))
		}
		if contributors[0].Name != "Alice" || contributors[1].Name != "Jane Doe" || contributors[2].Name != "John Smith" {
			t.Errorf("incorrect sort order: %+v", contributors)
		}
	})

	// 2. Test get_contributor
	t.Run("get_contributor success", func(t *testing.T) {
		req := fmt.Sprintf(`{"jsonrpc":"2.0","id":19,"method":"tools/call","params":{"name":"get_contributor","arguments":{"contributor_id":%q,"repository_id":"repo_xxx"}}}`, janeID) + "\n"
		output := runServerTest(t, registry, service, req)

		var resp JSONRPCResponse
		if err := json.Unmarshal([]byte(output), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		result := parseToolResult(t, resp.Result)
		if result.IsError {
			t.Fatalf("unexpected tool error: %s", result.Content[0].Text)
		}

		var c models.Contributor
		if err := json.Unmarshal([]byte(result.Content[0].Text), &c); err != nil {
			t.Fatalf("failed to parse contributor: %v", err)
		}
		if c.Name != "Jane Doe" {
			t.Errorf("expected contributor 'Jane Doe', got %q", c.Name)
		}
	})

	// 3. Test list_expertise
	t.Run("list_expertise success", func(t *testing.T) {
		req := `{"jsonrpc":"2.0","id":20,"method":"tools/call","params":{"name":"list_expertise","arguments":{"repository_id":"repo_xxx"}}}` + "\n"
		output := runServerTest(t, registry, service, req)

		var resp JSONRPCResponse
		if err := json.Unmarshal([]byte(output), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		result := parseToolResult(t, resp.Result)
		if result.IsError {
			t.Fatalf("unexpected tool error: %s", result.Content[0].Text)
		}

		var expertise []*models.Expertise
		if err := json.Unmarshal([]byte(result.Content[0].Text), &expertise); err != nil {
			t.Fatalf("failed to parse expertise: %v", err)
		}
		if len(expertise) != 3 {
			t.Errorf("expected 3 expertise records, got %d", len(expertise))
		}
	})

	// 4. Test trace_contributor
	t.Run("trace_contributor success", func(t *testing.T) {
		req := fmt.Sprintf(`{"jsonrpc":"2.0","id":21,"method":"tools/call","params":{"name":"trace_contributor","arguments":{"contributor_id":%q,"repository_id":"repo_xxx"}}}`, janeID) + "\n"
		output := runServerTest(t, registry, service, req)

		var resp JSONRPCResponse
		if err := json.Unmarshal([]byte(output), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		result := parseToolResult(t, resp.Result)
		if result.IsError {
			t.Fatalf("unexpected tool error: %s", result.Content[0].Text)
		}

		var trace ownershipquery.ContributorTrace
		if err := json.Unmarshal([]byte(result.Content[0].Text), &trace); err != nil {
			t.Fatalf("failed to parse trace: %v", err)
		}
		if trace.Contributor.Name != "Jane Doe" {
			t.Errorf("expected name 'Jane Doe', got %q", trace.Contributor.Name)
		}
		if len(trace.Expertise) != 1 || trace.Expertise[0].Domain != "Storage" {
			t.Errorf("expected 1 expertise 'Storage'")
		}
		if len(trace.Decisions) != 1 || trace.Decisions[0].Title != "Jane's Storage Decision" {
			t.Errorf("expected 1 decision 'Jane's Storage Decision'")
		}
		if len(trace.Facts) != 1 || trace.Facts[0].Subject != "Jane Storage" {
			t.Errorf("expected 1 fact 'Jane Storage'")
		}
		if len(trace.Events) != 1 || trace.Events[0].Title != "Jane's Event" {
			t.Errorf("expected 1 event 'Jane's Event'")
		}
	})

	// 5. Test recommend_reviewers (with nil ReviewerService, expect service-not-configured error)
	t.Run("recommend_reviewers success", func(t *testing.T) {
		req := `{"jsonrpc":"2.0","id":22,"method":"tools/call","params":{"name":"recommend_reviewers","arguments":{"repository_id":"repo_xxx","recommendation_type":"domain","domain":"Storage"}}}` + "\n"
		output := runServerTest(t, registry, service, req)

		var resp JSONRPCResponse
		if err := json.Unmarshal([]byte(output), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		result := parseToolResult(t, resp.Result)
		// ReviewerService is nil in this test context, so expect a tool error.
		if !result.IsError {
			t.Fatal("expected tool error when reviewer service is not configured")
		}
		if !strings.Contains(result.Content[0].Text, "reviewer service is not configured") {
			t.Errorf("expected 'reviewer service is not configured' error, got: %s", result.Content[0].Text)
		}
	})
}

func TestServer_GraphToolsExecution(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "reponerve-graph-server-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := migrations.RunUp(db); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	repoID := "repo_graph_test"

	// Insert repository
	_, err = db.Exec("INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at) VALUES (?, ?, ?, ?, datetime(), datetime())", repoID, "Graph Test Repo", tempDir, "main")
	if err != nil {
		t.Fatalf("failed to insert repository: %v", err)
	}

	// Insert source
	_, err = db.Exec("INSERT INTO sources (id, repository_id, source_type, reference, title, author, timestamp, created_at) VALUES (?, ?, ?, ?, ?, ?, datetime(), datetime())", "src_g1", repoID, "adr", "docs/adr/0001.md", "Author", "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("failed to insert source: %v", err)
	}

	// Insert two decisions with an explicit reference in content
	_, err = db.Exec("INSERT INTO memory_decisions (id, repository_id, source_id, title, status, created_at) VALUES (?, ?, ?, ?, ?, datetime())", "dec_g1", repoID, "src_g1", "Use PostgreSQL", "Accepted")
	if err != nil {
		t.Fatalf("failed to insert decision g1: %v", err)
	}
	_, err = db.Exec("INSERT INTO memory_decisions (id, repository_id, source_id, title, status, created_at) VALUES (?, ?, ?, ?, ?, datetime())", "dec_g2", repoID, "src_g1", "Add Connection Pooling", "Accepted")
	if err != nil {
		t.Fatalf("failed to insert decision g2: %v", err)
	}

	// Insert fact (subject chain: PostgreSQL -> SUPPORTS -> JSONB)
	_, err = db.Exec("INSERT INTO memory_facts (id, repository_id, source_id, subject, predicate, object, created_at) VALUES (?, ?, ?, ?, ?, ?, datetime())", "fact_g1", repoID, "src_g1", "PostgreSQL", "SUPPORTS", "JSONB")
	if err != nil {
		t.Fatalf("failed to insert fact g1: %v", err)
	}
	_, err = db.Exec("INSERT INTO memory_facts (id, repository_id, source_id, subject, predicate, object, created_at) VALUES (?, ?, ?, ?, ?, ?, datetime())", "fact_g2", repoID, "src_g1", "JSONB", "ENABLES", "SchemalessStorage")
	if err != nil {
		t.Fatalf("failed to insert fact g2: %v", err)
	}

	// Insert event
	_, err = db.Exec("INSERT INTO memory_events (id, repository_id, event_type, title, description, source_id, timestamp, created_at) VALUES (?, ?, ?, ?, ?, ?, datetime(), datetime())", "evt_g1", repoID, "FEATURE_INTRODUCED", "Introduce Pooling", "Connection pool added", "src_g1")
	if err != nil {
		t.Fatalf("failed to insert event: %v", err)
	}

	// Insert stored relationships
	_, err = db.Exec("INSERT INTO memory_relationships (id, repository_id, from_id, to_id, relationship_type, created_at) VALUES (?, ?, ?, ?, ?, datetime())", "rel_g1", repoID, "fact_g1", "dec_g1", "FACT_SUPPORTS_DECISION")
	if err != nil {
		t.Fatalf("failed to insert rel g1: %v", err)
	}
	_, err = db.Exec("INSERT INTO memory_relationships (id, repository_id, from_id, to_id, relationship_type, created_at) VALUES (?, ?, ?, ?, ?, datetime())", "rel_g2", repoID, "dec_g1", "evt_g1", "DECISION_RESULTS_IN_EVENT")
	if err != nil {
		t.Fatalf("failed to insert rel g2: %v", err)
	}

	// Insert contributor and expertise
	_, err = db.Exec("INSERT INTO contributors (id, repository_id, email, name, first_seen, last_seen, commit_count) VALUES (?, ?, ?, ?, datetime(), datetime(), ?)", "contrib_g1", repoID, "dev@example.com", "Dev User", 5)
	if err != nil {
		t.Fatalf("failed to insert contributor: %v", err)
	}
	_, err = db.Exec("INSERT INTO expertise (id, repository_id, contributor_id, domain, score, evidence_json) VALUES (?, ?, ?, ?, ?, ?)", "exp_g1", repoID, "contrib_g1", "storage", 0.9, `{"commits":5,"recent_activity":true}`)
	if err != nil {
		t.Fatalf("failed to insert expertise: %v", err)
	}

	// Set up graph stack
	er := storage.NewSQLiteEventReader(db)
	dr := storage.NewSQLiteDecisionReader(db)
	ir := storage.NewSQLiteIntentReader(db)
	fr := storage.NewSQLiteFactReader(db)
	rr := storage.NewSQLiteRelationshipReader(db)
	cr := storage.NewSQLiteContributorReader(db)
	expr := storage.NewSQLiteExpertiseReader(db)
	sr := storage.NewSQLiteSourceReader(db)

	ctxReader := context.NewMemoryContextReader(er, dr, ir, fr)
	generator := context.NewGenerator(ctxReader)
	renderer := render.NewRenderer()
	ownershipReader := ownershipquery.NewReader(cr, expr, sr, dr, fr, er)

	// Build graph services using the test helper to avoid import cycle
	graphService := buildGraphService(dr, ir, fr, er, rr, cr, expr, sr)

	service := mcp.NewService(dr, ir, fr, er, rr, generator, renderer, ownershipReader,
		graphService.travEngine, graphService.impactSvc, nil, nil, nil, nil, nil)
	registry := mcp.NewRegistry()

	// --- Tool discovery: verify graph schemas ---
	t.Run("tools/list includes graph tools", func(t *testing.T) {
		req := `{"jsonrpc":"2.0","id":1,"method":"tools/list"}` + "\n"
		output := runServerTest(t, registry, service, req)

		var resp JSONRPCResponse
		if err := json.Unmarshal([]byte(output), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
		resultData, _ := json.Marshal(resp.Result)
		var listResult ListToolsResult
		_ = json.Unmarshal(resultData, &listResult)

		graphToolNames := map[string]bool{
			"trace_graph": false, "trace_path": false,
			"find_dependencies": false, "find_dependents": false,
			"analyze_impact": false,
		}
		for _, tool := range listResult.Tools {
			if _, ok := graphToolNames[tool.Name]; ok {
				graphToolNames[tool.Name] = true
			}
		}
		for name, found := range graphToolNames {
			if !found {
				t.Errorf("graph tool %q not found in tools/list", name)
			}
		}

		// Verify schema for analyze_impact has required fields
		for _, tool := range listResult.Tools {
			if tool.Name == "analyze_impact" {
				if len(tool.InputSchema.Required) == 0 {
					t.Error("analyze_impact schema missing required fields")
				}
				found := false
				for _, req := range tool.InputSchema.Required {
					if req == "node_type" {
						found = true
					}
				}
				if !found {
					t.Error("analyze_impact schema required does not include node_type")
				}
			}
		}
	})

	// --- trace_graph: empty start node (no outbound edges) ---
	t.Run("trace_graph no paths", func(t *testing.T) {
		req := fmt.Sprintf(`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"trace_graph","arguments":{"node_id":%q,"repository_id":%q}}}`+"\n",
			"nod_nonexistent", repoID)
		output := runServerTest(t, registry, service, req)

		var resp JSONRPCResponse
		_ = json.Unmarshal([]byte(output), &resp)
		result := parseToolResult(t, resp.Result)
		if result.IsError {
			t.Fatalf("unexpected tool error: %s", result.Content[0].Text)
		}
		// Should return a traversal result with empty paths
		var travResult map[string]interface{}
		_ = json.Unmarshal([]byte(result.Content[0].Text), &travResult)
		if paths, ok := travResult["paths"]; ok {
			if paths != nil {
				pathsSlice, ok := paths.([]interface{})
				if ok && len(pathsSlice) > 0 {
					t.Errorf("expected no paths for unknown node, got %d", len(pathsSlice))
				}
			}
		}
	})

	// --- trace_graph: decision node with outbound stored edges ---
	t.Run("trace_graph from decision node", func(t *testing.T) {
		// Compute the graph node ID for dec_g1
		h := fmt.Sprintf("%s:%s:%s", repoID, "DECISION", "dec_g1")
		_ = h // node ID is sha256 of this
		// Use the entity ID to compute node ID via model.NodeID logic: sha256(repoID:nodeType:entityID)[:16]
		nodeID := computeNodeID(repoID, "DECISION", "dec_g1")
		req := fmt.Sprintf(`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"trace_graph","arguments":{"node_id":%q,"repository_id":%q}}}`+"\n",
			nodeID, repoID)
		output := runServerTest(t, registry, service, req)

		var resp JSONRPCResponse
		_ = json.Unmarshal([]byte(output), &resp)
		result := parseToolResult(t, resp.Result)
		if result.IsError {
			t.Fatalf("unexpected tool error: %s", result.Content[0].Text)
		}
		// Traversal result should have at least one path (decision→event)
		var travResult map[string]interface{}
		_ = json.Unmarshal([]byte(result.Content[0].Text), &travResult)
		paths, ok := travResult["paths"]
		if !ok {
			t.Fatal("expected 'paths' key in traversal result")
		}
		pathsSlice, _ := paths.([]interface{})
		if len(pathsSlice) == 0 {
			t.Error("expected at least one traversal path from decision node, got 0")
		}
		// Verify the path has nodes and edges
		if len(pathsSlice) > 0 {
			path0, _ := pathsSlice[0].(map[string]interface{})
			if path0 == nil {
				t.Fatal("expected path[0] to be an object")
			}
			if _, ok := path0["nodes"]; !ok {
				t.Error("expected 'nodes' in path")
			}
			if _, ok := path0["edges"]; !ok {
				t.Error("expected 'edges' in path")
			}
		}
	})

	// --- trace_path: no matching path ---
	t.Run("trace_path no match", func(t *testing.T) {
		req := fmt.Sprintf(`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"trace_path","arguments":{"start_node_id":%q,"end_node_id":%q,"repository_id":%q}}}`+"\n",
			"nod_start", "nod_end", repoID)
		output := runServerTest(t, registry, service, req)

		var resp JSONRPCResponse
		_ = json.Unmarshal([]byte(output), &resp)
		result := parseToolResult(t, resp.Result)
		if result.IsError {
			t.Fatalf("unexpected tool error: %s", result.Content[0].Text)
		}
		// Should return empty array (not error)
		var paths []interface{}
		_ = json.Unmarshal([]byte(result.Content[0].Text), &paths)
		if len(paths) != 0 {
			t.Errorf("expected 0 paths for no match, got %d", len(paths))
		}
	})

	// --- find_dependencies: empty result ---
	t.Run("find_dependencies empty", func(t *testing.T) {
		req := fmt.Sprintf(`{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"find_dependencies","arguments":{"node_id":%q,"repository_id":%q}}}`+"\n",
			"nod_unknown", repoID)
		output := runServerTest(t, registry, service, req)

		var resp JSONRPCResponse
		_ = json.Unmarshal([]byte(output), &resp)
		result := parseToolResult(t, resp.Result)
		if result.IsError {
			t.Fatalf("unexpected tool error: %s", result.Content[0].Text)
		}
	})

	// --- find_dependents: empty result ---
	t.Run("find_dependents empty", func(t *testing.T) {
		req := fmt.Sprintf(`{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"find_dependents","arguments":{"node_id":%q,"repository_id":%q}}}`+"\n",
			"nod_unknown", repoID)
		output := runServerTest(t, registry, service, req)

		var resp JSONRPCResponse
		_ = json.Unmarshal([]byte(output), &resp)
		result := parseToolResult(t, resp.Result)
		if result.IsError {
			t.Fatalf("unexpected tool error: %s", result.Content[0].Text)
		}
	})

	// --- analyze_impact DECISION ---
	t.Run("analyze_impact decision", func(t *testing.T) {
		req := fmt.Sprintf(`{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"analyze_impact","arguments":{"node_id":%q,"node_type":"DECISION","repository_id":%q}}}`+"\n",
			"dec_g1", repoID)
		output := runServerTest(t, registry, service, req)

		var resp JSONRPCResponse
		_ = json.Unmarshal([]byte(output), &resp)
		result := parseToolResult(t, resp.Result)
		if result.IsError {
			t.Fatalf("unexpected tool error: %s", result.Content[0].Text)
		}
		// Verify impact report structure
		var report map[string]interface{}
		if err := json.Unmarshal([]byte(result.Content[0].Text), &report); err != nil {
			t.Fatalf("failed to parse impact report: %v", err)
		}
		paths, ok := report["impact_paths"]
		if !ok {
			t.Fatal("expected 'impact_paths' in impact report")
		}
		pathsSlice, _ := paths.([]interface{})
		if len(pathsSlice) == 0 {
			t.Error("expected at least one impact path for dec_g1, got 0")
		}
		// Verify reason and evidence preservation
		if len(pathsSlice) > 0 {
			ip0, _ := pathsSlice[0].(map[string]interface{})
			if ip0 == nil {
				t.Fatal("expected impact_paths[0] to be an object")
			}
			if reason, ok := ip0["reason"]; !ok || reason == "" {
				t.Error("expected non-empty reason in impact path")
			}
			pathObj, _ := ip0["path"].(map[string]interface{})
			if pathObj == nil {
				t.Fatal("expected 'path' object in impact path")
			}
			if _, hasEdges := pathObj["edges"]; !hasEdges {
				t.Error("expected 'edges' in impact path.path (evidence preservation)")
			}
		}
	})

	// --- analyze_impact FACT ---
	t.Run("analyze_impact fact", func(t *testing.T) {
		req := fmt.Sprintf(`{"jsonrpc":"2.0","id":8,"method":"tools/call","params":{"name":"analyze_impact","arguments":{"node_id":%q,"node_type":"FACT","repository_id":%q}}}`+"\n",
			"fact_g1", repoID)
		output := runServerTest(t, registry, service, req)

		var resp JSONRPCResponse
		_ = json.Unmarshal([]byte(output), &resp)
		result := parseToolResult(t, resp.Result)
		if result.IsError {
			t.Fatalf("unexpected tool error: %s", result.Content[0].Text)
		}
		var report map[string]interface{}
		_ = json.Unmarshal([]byte(result.Content[0].Text), &report)
		if _, ok := report["impact_paths"]; !ok {
			t.Error("expected 'impact_paths' key in fact impact report")
		}
	})

	// --- analyze_impact CONTRIBUTOR ---
	t.Run("analyze_impact contributor", func(t *testing.T) {
		req := fmt.Sprintf(`{"jsonrpc":"2.0","id":9,"method":"tools/call","params":{"name":"analyze_impact","arguments":{"node_id":%q,"node_type":"CONTRIBUTOR","repository_id":%q}}}`+"\n",
			"contrib_g1", repoID)
		output := runServerTest(t, registry, service, req)

		var resp JSONRPCResponse
		_ = json.Unmarshal([]byte(output), &resp)
		result := parseToolResult(t, resp.Result)
		if result.IsError {
			t.Fatalf("unexpected tool error: %s", result.Content[0].Text)
		}
		var report map[string]interface{}
		_ = json.Unmarshal([]byte(result.Content[0].Text), &report)
		if _, ok := report["impact_paths"]; !ok {
			t.Error("expected 'impact_paths' key in contributor impact report")
		}
	})

	// --- analyze_impact unsupported type ---
	t.Run("analyze_impact unsupported type", func(t *testing.T) {
		req := fmt.Sprintf(`{"jsonrpc":"2.0","id":10,"method":"tools/call","params":{"name":"analyze_impact","arguments":{"node_id":%q,"node_type":"UNKNOWN_TYPE","repository_id":%q}}}`+"\n",
			"dec_g1", repoID)
		output := runServerTest(t, registry, service, req)

		var resp JSONRPCResponse
		_ = json.Unmarshal([]byte(output), &resp)
		result := parseToolResult(t, resp.Result)
		if !result.IsError {
			t.Fatal("expected isError=true for unsupported node_type")
		}
		if !strings.Contains(result.Content[0].Text, "unsupported node_type") {
			t.Errorf("expected error to mention 'unsupported node_type', got: %s", result.Content[0].Text)
		}
	})

	// --- Error: missing node_id ---
	t.Run("trace_graph missing node_id", func(t *testing.T) {
		req := `{"jsonrpc":"2.0","id":11,"method":"tools/call","params":{"name":"trace_graph","arguments":{}}}` + "\n"
		output := runServerTest(t, registry, service, req)

		var resp JSONRPCResponse
		_ = json.Unmarshal([]byte(output), &resp)
		result := parseToolResult(t, resp.Result)
		if !result.IsError {
			t.Fatal("expected isError=true for missing node_id")
		}
	})

	// --- Error: missing node_type for analyze_impact ---
	t.Run("analyze_impact missing node_type", func(t *testing.T) {
		req := fmt.Sprintf(`{"jsonrpc":"2.0","id":12,"method":"tools/call","params":{"name":"analyze_impact","arguments":{"node_id":%q}}}`+"\n",
			"dec_g1")
		output := runServerTest(t, registry, service, req)

		var resp JSONRPCResponse
		_ = json.Unmarshal([]byte(output), &resp)
		result := parseToolResult(t, resp.Result)
		if !result.IsError {
			t.Fatal("expected isError=true for missing node_type")
		}
	})
}

// graphServiceDeps holds graph service dependencies for tests.
type graphServiceDeps struct {
	travEngine *traversal.Engine
	impactSvc  *impact.Service
}

func buildGraphService(
	dr storage.DecisionReader,
	ir storage.IntentReader,
	fr storage.FactReader,
	er storage.EventReader,
	rr storage.RelationshipReader,
	cr storage.ContributorReader,
	expr storage.ExpertiseReader,
	sr storage.SourceReader,
) graphServiceDeps {
	relEngine := relationships.NewEngine(dr, ir, fr, er, rr, cr, expr, sr)
	travEngine := traversal.NewEngine(relEngine)
	impactSvc := impact.NewService(travEngine)
	return graphServiceDeps{
		travEngine: travEngine,
		impactSvc:  impactSvc,
	}
}

func computeNodeID(repoID string, nodeType string, entityID string) string {
	return model.NodeID(repoID, model.NodeType(nodeType), entityID)
}
