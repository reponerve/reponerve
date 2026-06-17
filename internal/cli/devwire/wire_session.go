package devwire

import (
	"context"
	"fmt"

	agentcontext "github.com/reponerve/reponerve/internal/agent/context"
	agentsearch "github.com/reponerve/reponerve/internal/agent/search"
	agentsession "github.com/reponerve/reponerve/internal/agent/session"
	"github.com/reponerve/reponerve/internal/agent/sessionmemory"
	"github.com/reponerve/reponerve/internal/agent/workflow"
	ctxengine "github.com/reponerve/reponerve/internal/context"
	"github.com/reponerve/reponerve/internal/graph/impact"
	"github.com/reponerve/reponerve/internal/graph/relationships"
	"github.com/reponerve/reponerve/internal/graph/traversal"
	"github.com/reponerve/reponerve/internal/intelligence/changeplan"
	"github.com/reponerve/reponerve/internal/intelligence/discovery"
	"github.com/reponerve/reponerve/internal/intelligence/learning"
	"github.com/reponerve/reponerve/internal/intelligence/reviewers"
	memorystorage "github.com/reponerve/reponerve/internal/memory/storage"
	"github.com/reponerve/reponerve/internal/query/storage"
	"github.com/reponerve/reponerve/internal/scanner/repository"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
)

// WireSessionMemoryService constructs session memory from an open database.
func WireSessionMemoryService(db *sqlite.Database, workspaceDir string) *sessionmemory.Service {
	return sessionmemory.NewService(
		memorystorage.NewSQLiteFactStore(db),
		sqlite.NewSourceStore(db),
		storage.NewSQLiteFactReader(db),
		storage.NewSQLiteEventReader(db),
		storage.NewSQLiteDecisionReader(db),
		sqlite.NewMemorySearchStore(db),
		storage.NewSQLiteSourceReader(db),
		workspaceDir,
	)
}

// WireWorkflowService constructs workflow templates from an open database.
func WireWorkflowService(ctx context.Context, db *sqlite.Database, repositoryPath string) (string, *workflow.Service, error) {
	discoverySvc := repository.NewGitDiscovery()
	repo, err := discoverySvc.Discover(ctx, repositoryPath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to discover repository: %w", err)
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
	learningSvc := learning.NewService(
		discoveryService, decisionReader, factReader, eventReader,
		contributorReader, expertiseReader, sourceReader, relEngine,
	)
	reviewerSvc := reviewers.NewService(
		discoveryService, decisionReader, factReader, eventReader,
		contributorReader, expertiseReader, sourceReader, graphImpactSvc,
	)
	changePlanSvc := changeplan.NewService(graphImpactSvc)
	contextSvc := agentcontext.NewService(
		discoveryService, learningSvc, reviewerSvc, changePlanSvc, generator,
	)
	sessionSvc := agentsession.NewService(contextSvc, searchSvc)

	wf := workflow.NewService(
		discoveryService, learningSvc, reviewerSvc, changePlanSvc,
		contextSvc, searchSvc, sessionSvc,
	)
	return repo.ID, wf, nil
}
