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
	ctxengine "github.com/reponerve/reponerve/internal/context"
	"github.com/reponerve/reponerve/internal/graph/impact"
	"github.com/reponerve/reponerve/internal/graph/relationships"
	"github.com/reponerve/reponerve/internal/graph/traversal"
	"github.com/reponerve/reponerve/internal/intelligence/changeplan"
	"github.com/reponerve/reponerve/internal/intelligence/discovery"
	"github.com/reponerve/reponerve/internal/intelligence/feature"
	"github.com/reponerve/reponerve/internal/intelligence/learning"
	"github.com/reponerve/reponerve/internal/intelligence/reviewers"
	"github.com/reponerve/reponerve/internal/query/storage"
	"github.com/reponerve/reponerve/internal/scanner/repository"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
)

// WireDevelopmentService constructs a Development Experience service from an open database.
func WireDevelopmentService(ctx context.Context, db *sqlite.Database, repositoryPath string) (string, *development.Service, error) {
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
	onboardingService := onboarding.NewService(generator)
	guidanceService := guidance.NewService(decisionReader, intentReader, factReader, eventReader, relationshipReader)
	agentImpactSvc := agentimpact.NewService(decisionReader, intentReader, factReader, eventReader, relationshipReader)
	qaService := qa.NewService(onboardingService, guidanceService, agentImpactSvc)
	changePlanSvc := changeplan.NewService(graphImpactSvc)
	learningSvc := learning.NewService(
		discoveryService, decisionReader, factReader, eventReader,
		contributorReader, expertiseReader, sourceReader, relEngine,
	)
	reviewerSvc := reviewers.NewService(
		discoveryService, decisionReader, factReader, eventReader,
		contributorReader, expertiseReader, sourceReader, graphImpactSvc,
	)

	codeEntityReader := storage.NewSQLiteCodeEntityReader(db)
	relReader := storage.NewSQLiteCodeRelationshipReader(db)
	repoCodeReader := storage.NewSQLiteRepositoryCodeRelationshipReader(db)
	codeSvc := code.NewService(codeEntityReader, relReader, repoCodeReader)
	featureSvc := feature.NewService(eventReader, expertiseReader, decisionReader)

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
		repositoryPath,
		learningSvc,
		reviewerSvc,
		changePlanSvc,
		graphImpactSvc,
		agentImpactSvc,
		featureSvc,
	)

	return repo.ID, devSvc, nil
}
