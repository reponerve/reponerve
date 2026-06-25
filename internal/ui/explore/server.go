package exploreui

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

// Server serves the local explore UI.
type Server struct {
	Host    string
	Port    int
	Payload *Payload
}

// ListenAndServe starts the HTTP server on Host:Port (127.0.0.1 only).
func (s *Server) ListenAndServe(ctx context.Context) error {
	host := s.Host
	if host == "" {
		host = "127.0.0.1"
	}
	if host != "127.0.0.1" && host != "localhost" {
		return fmt.Errorf("explore server only binds to 127.0.0.1 or localhost, got %q", host)
	}
	port := s.Port
	if port == 0 {
		port = 8765
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.handleIndex)
	mux.HandleFunc("GET /nodes", s.handleNodes)
	mux.HandleFunc("GET /nodes/{id}", s.handleNodeDetail)

	addr := fmt.Sprintf("%s:%d", host, port)
	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdown)
	}()

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	fmt.Printf("RepoNerve explore UI at http://%s/\n", addr)
	return srv.Serve(ln)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = ExplorePage(s.Payload).Render(r.Context(), w)
}

func (s *Server) handleNodes(w http.ResponseWriter, r *http.Request) {
	nodeType := r.URL.Query().Get("type")
	query := r.URL.Query().Get("q")
	nodes := FilterNodes(s.Payload.Nodes, nodeType, query)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = NodesTable(nodes).Render(r.Context(), w)
}

func (s *Server) handleNodeDetail(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	detail, err := NodeDetailFor(s.Payload, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = NodeDetailPanel(detail).Render(r.Context(), w)
}
