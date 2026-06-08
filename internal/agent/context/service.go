package agentcontext

import (
	stdcontext "context"
	"encoding/json"
	"fmt"

	appcontext "reponerve/internal/context"
	"reponerve/internal/intelligence/changeplan"
	"reponerve/internal/intelligence/discovery"
	"reponerve/internal/intelligence/learning"
	"reponerve/internal/intelligence/reviewers"
)

// Service packages Repository Intelligence into deterministic AgentContextPackages.
// It is a composition layer — it never generates intelligence, only assembles it.
type Service struct {
	discoveryService  *discovery.Service
	learningService   *learning.Service
	reviewerService   *reviewers.Service
	changePlanService *changeplan.Service
	contextGenerator  *appcontext.Generator
}

// NewService constructs a new Agent Context Builder Service.
func NewService(
	discoverySvc *discovery.Service,
	learningSvc *learning.Service,
	reviewerSvc *reviewers.Service,
	changePlanSvc *changeplan.Service,
	ctxGenerator *appcontext.Generator,
) *Service {
	return &Service{
		discoveryService:  discoverySvc,
		learningService:   learningSvc,
		reviewerService:   reviewerSvc,
		changePlanService: changePlanSvc,
		contextGenerator:  ctxGenerator,
	}
}

// BuildRepositoryContext produces a 4-section context package answering:
// "What should an agent know about this repository?"
//
// Sections (in fixed order):
//  1. Repository Overview (source: context)
//  2. Discovery           (source: discovery)
//  3. Learning Path       (source: learning)
//  4. Reviewer Recommendations (source: reviewers)
func (s *Service) BuildRepositoryContext(ctx stdcontext.Context, repositoryID string) (*AgentContextPackage, error) {
	if repositoryID == "" {
		return nil, fmt.Errorf("repository ID cannot be empty")
	}

	// Section 1 — Repository Overview
	overviewSection, err := s.buildContextSection("Repository Overview", repositoryID, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build repository overview: %w", err)
	}

	// Section 2 — Discovery
	discoveryReport, err := s.discoveryService.Discover(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to discover knowledge: %w", err)
	}
	discoveryData, err := json.Marshal(discoveryReport)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize discovery report: %w", err)
	}
	discoverySection := &ContextSection{
		Name:   "Discovery",
		Source: SourceDiscovery,
		Data:   discoveryData,
	}

	// Section 3 — Learning Path
	learningPath, err := s.learningService.GenerateRepositoryPath(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate repository learning path: %w", err)
	}
	learningData, err := json.Marshal(learningPath)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize learning path: %w", err)
	}
	learningSection := &ContextSection{
		Name:   "Learning Path",
		Source: SourceLearning,
		Data:   learningData,
	}

	// Section 4 — Reviewer Recommendations
	reviewerReport, err := s.reviewerService.RecommendRepositoryReviewers(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to recommend repository reviewers: %w", err)
	}
	reviewerData, err := json.Marshal(reviewerReport)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize reviewer report: %w", err)
	}
	reviewerSection := &ContextSection{
		Name:   "Reviewer Recommendations",
		Source: SourceReviewers,
		Data:   reviewerData,
	}

	pkg := &AgentContextPackage{
		RepositoryID: repositoryID,
		Sections: []*ContextSection{
			overviewSection,
			discoverySection,
			learningSection,
			reviewerSection,
		},
	}

	if err := ValidatePackage(pkg); err != nil {
		return nil, fmt.Errorf("generated repository context package is invalid: %w", err)
	}

	return pkg, nil
}

// BuildDomainContext produces a 3-section context package answering:
// "What should an agent know about this repository domain?"
//
// Sections (in fixed order):
//  1. Domain Overview          (source: context)
//  2. Learning Path            (source: learning)
//  3. Reviewer Recommendations (source: reviewers)
func (s *Service) BuildDomainContext(ctx stdcontext.Context, repositoryID string, domain string) (*AgentContextPackage, error) {
	if repositoryID == "" {
		return nil, fmt.Errorf("repository ID cannot be empty")
	}
	if domain == "" {
		return nil, fmt.Errorf("domain cannot be empty")
	}

	// Section 1 — Domain Overview
	overviewSection, err := s.buildContextSection("Domain Overview", repositoryID, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build domain overview: %w", err)
	}

	// Section 2 — Learning Path (domain)
	learningPath, err := s.learningService.GenerateDomainPath(ctx, repositoryID, domain)
	if err != nil {
		return nil, fmt.Errorf("failed to generate domain learning path: %w", err)
	}
	learningData, err := json.Marshal(learningPath)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize domain learning path: %w", err)
	}
	learningSection := &ContextSection{
		Name:   "Learning Path",
		Source: SourceLearning,
		Data:   learningData,
	}

	// Section 3 — Reviewer Recommendations (domain)
	reviewerReport, err := s.reviewerService.RecommendDomainReviewers(ctx, repositoryID, domain)
	if err != nil {
		return nil, fmt.Errorf("failed to recommend domain reviewers: %w", err)
	}
	reviewerData, err := json.Marshal(reviewerReport)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize domain reviewer report: %w", err)
	}
	reviewerSection := &ContextSection{
		Name:   "Reviewer Recommendations",
		Source: SourceReviewers,
		Data:   reviewerData,
	}

	pkg := &AgentContextPackage{
		RepositoryID: repositoryID,
		Sections: []*ContextSection{
			overviewSection,
			learningSection,
			reviewerSection,
		},
	}

	if err := ValidatePackage(pkg); err != nil {
		return nil, fmt.Errorf("generated domain context package is invalid: %w", err)
	}

	return pkg, nil
}

// BuildContributorContext produces a 3-section context package answering:
// "What should an agent know about this contributor area?"
//
// Sections (in fixed order):
//  1. Contributor Overview (source: context)
//  2. Learning Path        (source: learning)
//  3. Change Plan          (source: changeplan)
func (s *Service) BuildContributorContext(ctx stdcontext.Context, repositoryID string, contributorID string) (*AgentContextPackage, error) {
	if repositoryID == "" {
		return nil, fmt.Errorf("repository ID cannot be empty")
	}
	if contributorID == "" {
		return nil, fmt.Errorf("contributor ID cannot be empty")
	}

	// Section 1 — Contributor Overview
	overviewSection, err := s.buildContextSection("Contributor Overview", repositoryID, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build contributor overview: %w", err)
	}

	// Section 2 — Learning Path (contributor)
	learningPath, err := s.learningService.GenerateContributorPath(ctx, repositoryID, contributorID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate contributor learning path: %w", err)
	}
	learningData, err := json.Marshal(learningPath)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize contributor learning path: %w", err)
	}
	learningSection := &ContextSection{
		Name:   "Learning Path",
		Source: SourceLearning,
		Data:   learningData,
	}

	// Section 3 — Change Plan (contributor areas)
	changePlan, err := s.changePlanService.GenerateContributorPlan(ctx, repositoryID, contributorID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate contributor change plan: %w", err)
	}
	changePlanData, err := json.Marshal(changePlan)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize change plan: %w", err)
	}
	changePlanSection := &ContextSection{
		Name:   "Change Plan",
		Source: SourceChangePlan,
		Data:   changePlanData,
	}

	pkg := &AgentContextPackage{
		RepositoryID: repositoryID,
		Sections: []*ContextSection{
			overviewSection,
			learningSection,
			changePlanSection,
		},
	}

	if err := ValidatePackage(pkg); err != nil {
		return nil, fmt.Errorf("generated contributor context package is invalid: %w", err)
	}

	return pkg, nil
}

// buildContextSection calls the context generator and marshals the result
// into a ContextSection with the given name and source "context".
func (s *Service) buildContextSection(name string, repositoryID string, ctx stdcontext.Context) (*ContextSection, error) {
	repoCtx, err := s.contextGenerator.Generate(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate context: %w", err)
	}
	data, err := json.Marshal(repoCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize context: %w", err)
	}
	return &ContextSection{
		Name:   name,
		Source: SourceContext,
		Data:   data,
	}, nil
}
