package explorecmd

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/config"
	exploreui "github.com/reponerve/reponerve/internal/ui/explore"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
)

// NewCommand creates the explore subcommand.
func NewCommand() *cobra.Command {
	var outputPath string
	var serve bool
	var host string
	var port int

	cmd := &cobra.Command{
		Use:   "explore",
		Short: "Export or browse the knowledge graph",
		Long:  `Load the repository knowledge graph and export HTML or run a local explore UI on 127.0.0.1.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			workspaceDir := config.GetWorkspaceDir()
			cfg, err := config.Load(workspaceDir)
			if err != nil {
				return fmt.Errorf("%s", config.FormatLoadError(workspaceDir, err))
			}

			db, err := sqlite.Open(cfg.Storage.SQLitePath)
			if err != nil {
				return fmt.Errorf("failed to open database: %w", err)
			}
			defer db.Close()

			loader := &exploreui.Loader{DB: db, RepoPath: cfg.Repository.Path}
			payload, err := loader.Load(cmd.Context())
			if err != nil {
				return err
			}

			if serve {
				ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
				defer stop()
				srv := &exploreui.Server{Host: host, Port: port, Payload: payload}
				return srv.ListenAndServe(ctx)
			}

			html, err := exploreui.RenderHTML(payload)
			if err != nil {
				return fmt.Errorf("failed to render HTML: %w", err)
			}

			out := outputPath
			if !filepath.IsAbs(out) {
				out = filepath.Join(cfg.Repository.Path, out)
			}
			if err := os.WriteFile(out, []byte(html), 0o644); err != nil {
				return fmt.Errorf("failed to write HTML: %w", err)
			}

			cmd.Printf("✓ Graph exported to %s\n", out)
			cmd.Printf("  nodes=%d/%d edges=%d/%d communities=%d surprises=%d\n",
				len(payload.Nodes), payload.TotalNodes,
				len(payload.Edges), payload.TotalEdges,
				payload.Stats.Communities, payload.Stats.Surprises)
			cmd.Println("  Tip: reponerve explore --serve for interactive UI")
			return nil
		},
	}
	cmd.Flags().StringVarP(&outputPath, "output", "o", "reponerve-graph.html", "HTML output path (export mode)")
	cmd.Flags().BoolVar(&serve, "serve", false, "Run local explore UI on 127.0.0.1")
	cmd.Flags().StringVar(&host, "host", "127.0.0.1", "Bind host (127.0.0.1 or localhost only)")
	cmd.Flags().IntVar(&port, "port", 8765, "Bind port for --serve")
	return cmd
}
