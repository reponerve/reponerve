package exploreui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServerNodeDetail(t *testing.T) {
	payload := &Payload{
		Nodes: []NodeView{{ID: "n1", Type: "DECISION", EntityID: "d1"}},
	}
	srv := &Server{Payload: payload}

	t.Run("found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/nodes/n1", nil)
		req.SetPathValue("id", "n1")
		rec := httptest.NewRecorder()
		srv.handleNodeDetail(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "explain_decision") {
			t.Fatal("missing MCP hint")
		}
	})

	t.Run("missing", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/nodes/missing", nil)
		req.SetPathValue("id", "missing")
		rec := httptest.NewRecorder()
		srv.handleNodeDetail(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("status %d", rec.Code)
		}
	})
}
