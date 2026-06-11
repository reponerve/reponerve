package devwire

import (
	"context"
	"fmt"

	agentimpact "github.com/reponerve/reponerve/internal/agent/impact"
	"github.com/reponerve/reponerve/internal/agent/development"
	agentsearch "github.com/reponerve/reponerve/internal/agent/search"
	"github.com/reponerve/reponerve/internal/agent/guidance"
	"github.com/reponerve/reponerve/internal/agent/onboarding"
	"github.com/reponerve/reponerve/internal/agent/qa"
	"github.com/reponerve/reponerve/internal/code"
	"github.com/reponerve/reponerve/internal/config"
	ctxengine "github.com/reponerve/reponerve/internal/context"
	"github.com/reponerve/reponerve/internal/graph/impact"
	"github.com/reponerve/reponerve/internal/graph/relationships"
	"github.com/reponerve/reponerve/internal/graph/traversal"
	"github.com/reponerve/reponerve/internal/intelligence/discovery"
	"github.com/reponerve/reponerve/internal/query/storage"
	"github.com/reponerve/reponerve/internal/scanner/repository"
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
		return nil, fmt.Errorf("workspace not initialized; run 'reponerve init' first")
	}

	db, err := sqlite.Open(cfg.Storage.SQLitePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	discoverySvc := repository.NewGitDiscovery()
	repo, err := discoverySvc.Discover(ctx, cfg.Repository.Path)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to discover repository: %w", err)
	}

	decisionReader := storage.NewSQLiteDecisionReader(db)
	factReader := storage.NewSQLiteFactReader(db)
	eventReader := storage.NewSQLiteEventReader(db)
	intentReader := storage.NewSQLiteIntentReader(db)
	relationshipReader := storage.NewSQLiteRelationshipReader(db)
	contributorReader := storage.NewSQLiteContributorReader(db)
	expertiseReader := storage.NewSQLiteExpertiseReader(db)
	sourceReader := storage.NewSQLiteSourceReader(db)

	relEngine := relationships.NewEngine(
		decisionReader, intentReader, factReader, eventReader,
		relationshipReader, contributorReader, expertiseReader, sourceReader,
	)
	travEngine := traversal.NewEngine(relEngine)
	graphImpactSvc := impact.NewService(travEngine)
	discoveryService := discovery.NewService(
		decisionReader, factReader, eventReader, contributorReader, expertiseReader,
		relationshipReader, relEngine, travEngine, graphImpactSvc,
	)

	memorySearchStore := sqlite.NewMemorySearchStore(db)
	searchSvc := agentsearch.NewService(
		decisionReader, factReader, eventReader, relationshipReader,
		contributorReader, expertiseReader, discoveryService, memorySearchStore,
	)

	ctxReader := ctxengine.NewMemoryContextReader(eventReader, decisionReader, intentReader, factReader)
	generator := ctxengine.NewGenerator(ctxReader)
	onboardingService := onboarding.NewService(generator)
	guidanceService := guidance.NewService(decisionReader, intentReader, factReader, eventReader, relationshipReader)
	agentImpactSvc := agentimpact.NewService(decisionReader, intentReader, factReader, eventReader, relationshipReader)
	qaService := qa.NewService(onboardingService, guidanceService, agentImpactSvc)

	codeEntityReader := storage.NewSQLiteCodeEntityReader(db)
	relReader := storage.NewSQLiteCodeRelationshipReader(db)
	repoCodeReader := storage.NewSQLiteRepositoryCodeRelationshipReader(db)
	codeSvc := code.NewService(codeEntityReader, relReader, repoCodeReader)

	devSvc := development.NewService(
		codeSvc,
		searchSvc,
		codeEntityReader,
		relReader,
		repoCodeReader,
		decisionReader,
		factReader,
		eventReader,
		expertiseReader,
		qaService,
		contributorReader,
		sourceReader,
		cfg.Repository.Path,
	)

	return &Handle{
		RepositoryID: repo.ID,
		Service:      devSvc,
		closeDB:      func() { db.Close() },
	}, nil
}
