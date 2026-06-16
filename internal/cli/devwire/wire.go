package devwire

import (
	"context"
	"fmt"

	"github.com/reponerve/reponerve/internal/agent/development"
	"github.com/reponerve/reponerve/internal/config"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
)

// Handle holds a wired Development Experience service session.
type Handle struct {
	RepositoryID string
	Service      *development.Service
	closeDB      func()
}

// Close releases database resources.
func (h *Handle) Close() {
	if h != nil && h.closeDB != nil {
		h.closeDB()
	}
}

// Open wires Development Experience dependencies for CLI commands.
func Open(ctx context.Context, workspaceDir string) (*Handle, error) {
	cfg, err := config.Load(workspaceDir)
	if err != nil {
		return nil, fmt.Errorf("%s", config.FormatLoadError(workspaceDir, err))
	}

	db, err := sqlite.Open(cfg.Storage.SQLitePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	repoID, devSvc, err := WireDevelopmentService(ctx, db, cfg.Repository.Path)
	if err != nil {
		db.Close()
		return nil, err
	}

	return &Handle{
		RepositoryID: repoID,
		Service:      devSvc,
		closeDB:      func() { db.Close() },
	}, nil
}
