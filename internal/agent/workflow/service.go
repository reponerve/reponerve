package workflow

import (
	stdcontext "context"
	"encoding/json"
	"fmt"
	"strings"

	agentcontext "github.com/reponerve/reponerve/internal/agent/context"
	agentsearch "github.com/reponerve/reponerve/internal/agent/search"
	agentsession "github.com/reponerve/reponerve/internal/agent/session"
	"github.com/reponerve/reponerve/internal/intelligence/changeplan"
	"github.com/reponerve/reponerve/internal/intelligence/discovery"
	"github.com/reponerve/reponerve/internal/intelligence/learning"
	"github.com/reponerve/reponerve/internal/intelligence/reviewers"
)

type Service struct {
	discoveryService  *discovery.Service
	learningService   *learning.Service
	reviewerService   *reviewers.Service
	changePlanService *changeplan.Service
	contextService    *agentcontext.Service
	searchService     *agentsearch.Service
	sessionService    *agentsession.Service
}

func NewService(
	discoveryService *discovery.Service,
	learningService *learning.Service,
	reviewerService *reviewers.Service,
	changePlanService *changeplan.Service,
	contextService *agentcontext.Service,
	searchService *agentsearch.Service,
	sessionService *agentsession.Service,
) *Service {
	return &Service{
		discoveryService:  discoveryService,
		learningService:   learningService,
		reviewerService:   reviewerService,
		changePlanService: changePlanService,
		contextService:    contextService,
		searchService:     searchService,
		sessionService:    sessionService,
	}
}

func (s *Service) BuildOnboardingWorkflow(ctx stdcontext.Context, repositoryID string) (*WorkflowPackage, error) {
	if repositoryID == "" {
		return nil, fmt.Errorf("repository ID cannot be empty")
	}
	if s.sessionService == nil {
		return nil, fmt.Errorf("session service is required")
	}
	if s.discoveryService == nil {
		return nil, fmt.Errorf("discovery service is required")
	}
	if s.learningService == nil {
		return nil, fmt.Errorf("learning service is required")
	}

	session, err := s.sessionService.CreateRepositorySession(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository session: %w", err)
	}
	report, err := s.discoveryService.Discover(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to build discovery report: %w", err)
	}
	path, err := s.learningService.GenerateRepositoryPath(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to build learning path: %w", err)
	}

	sessionArtifact, err := newArtifact(ArtifactTypeSession, SourceSession, session)
	if err != nil {
		return nil, fmt.Errorf("failed to package repository session: %w", err)
	}
	discoveryArtifact, err := newArtifact(ArtifactTypeDiscoveryReport, SourceDiscovery, report)
	if err != nil {
		return nil, fmt.Errorf("failed to package discovery report: %w", err)
	}
	learningArtifact, err := newArtifact(ArtifactTypeLearningPath, SourceLearning, path)
	if err != nil {
		return nil, fmt.Errorf("failed to package learning path: %w", err)
	}

	workflow := &WorkflowPackage{
		WorkflowID:   buildWorkflowID(repositoryID, WorkflowTypeOnboarding, defaultWorkflowIdentifier),
		WorkflowType: WorkflowTypeOnboarding,
		RepositoryID: repositoryID,
		Version:      VersionV1,
		Artifacts: []*WorkflowArtifact{
			sessionArtifact,
			discoveryArtifact,
			learningArtifact,
		},
	}

	if err := ValidateWorkflow(workflow); err != nil {
		return nil, fmt.Errorf("generated onboarding workflow is invalid: %w", err)
	}

	return workflow, nil
}

func (s *Service) BuildReviewPreparationWorkflow(ctx stdcontext.Context, repositoryID string, query string) (*WorkflowPackage, error) {
	if repositoryID == "" {
		return nil, fmt.Errorf("repository ID cannot be empty")
	}
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}
	if s.sessionService == nil {
		return nil, fmt.Errorf("session service is required")
	}
	if s.reviewerService == nil {
		return nil, fmt.Errorf("reviewer service is required")
	}
	if s.searchService == nil {
		return nil, fmt.Errorf("search service is required")
	}
	if s.contextService == nil {
		return nil, fmt.Errorf("context service is required")
	}

	session, err := s.sessionService.CreateRepositorySession(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository session: %w", err)
	}
	report, err := s.buildReviewerReport(ctx, repositoryID, query)
	if err != nil {
		return nil, fmt.Errorf("failed to build reviewer report: %w", err)
	}
	result, err := s.searchService.Search(ctx, repositoryID, query)
	if err != nil {
		return nil, fmt.Errorf("failed to build search result: %w", err)
	}
	contextPackage, err := s.contextService.BuildRepositoryContext(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to build repository context package: %w", err)
	}

	sessionArtifact, err := newArtifact(ArtifactTypeSession, SourceSession, session)
	if err != nil {
		return nil, fmt.Errorf("failed to package repository session: %w", err)
	}
	reviewerArtifact, err := newArtifact(ArtifactTypeReviewerReport, SourceReviewers, report)
	if err != nil {
		return nil, fmt.Errorf("failed to package reviewer report: %w", err)
	}
	searchArtifact, err := newArtifact(ArtifactTypeSearchResult, SourceSearch, result)
	if err != nil {
		return nil, fmt.Errorf("failed to package search result: %w", err)
	}
	contextArtifact, err := newArtifact(ArtifactTypeContextPackage, SourceContext, contextPackage)
	if err != nil {
		return nil, fmt.Errorf("failed to package repository context: %w", err)
	}

	workflow := &WorkflowPackage{
		WorkflowID:   buildWorkflowID(repositoryID, WorkflowTypeReviewPreparation, query),
		WorkflowType: WorkflowTypeReviewPreparation,
		RepositoryID: repositoryID,
		Version:      VersionV1,
		Artifacts: []*WorkflowArtifact{
			sessionArtifact,
			reviewerArtifact,
			searchArtifact,
			contextArtifact,
		},
	}

	if err := ValidateWorkflow(workflow); err != nil {
		return nil, fmt.Errorf("generated review preparation workflow is invalid: %w", err)
	}

	return workflow, nil
}

func (s *Service) BuildChangePreparationWorkflow(ctx stdcontext.Context, repositoryID string, entityID string) (*WorkflowPackage, error) {
	if repositoryID == "" {
		return nil, fmt.Errorf("repository ID cannot be empty")
	}
	if entityID == "" {
		return nil, fmt.Errorf("entity ID cannot be empty")
	}
	if s.sessionService == nil {
		return nil, fmt.Errorf("session service is required")
	}
	if s.changePlanService == nil {
		return nil, fmt.Errorf("change plan service is required")
	}
	if s.searchService == nil {
		return nil, fmt.Errorf("search service is required")
	}
	if s.contextService == nil {
		return nil, fmt.Errorf("context service is required")
	}

	session, err := s.sessionService.CreateRepositorySession(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository session: %w", err)
	}
	plan, err := s.buildChangePlan(ctx, repositoryID, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to build change plan: %w", err)
	}
	result, err := s.searchService.Search(ctx, repositoryID, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to build search result: %w", err)
	}
	contextPackage, err := s.contextService.BuildRepositoryContext(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to build repository context package: %w", err)
	}

	sessionArtifact, err := newArtifact(ArtifactTypeSession, SourceSession, session)
	if err != nil {
		return nil, fmt.Errorf("failed to package repository session: %w", err)
	}
	changePlanArtifact, err := newArtifact(ArtifactTypeChangePlan, SourceChangePlan, plan)
	if err != nil {
		return nil, fmt.Errorf("failed to package change plan: %w", err)
	}
	searchArtifact, err := newArtifact(ArtifactTypeSearchResult, SourceSearch, result)
	if err != nil {
		return nil, fmt.Errorf("failed to package search result: %w", err)
	}
	contextArtifact, err := newArtifact(ArtifactTypeContextPackage, SourceContext, contextPackage)
	if err != nil {
		return nil, fmt.Errorf("failed to package repository context: %w", err)
	}

	workflow := &WorkflowPackage{
		WorkflowID:   buildWorkflowID(repositoryID, WorkflowTypeChangePreparation, entityID),
		WorkflowType: WorkflowTypeChangePreparation,
		RepositoryID: repositoryID,
		Version:      VersionV1,
		Artifacts: []*WorkflowArtifact{
			sessionArtifact,
			changePlanArtifact,
			searchArtifact,
			contextArtifact,
		},
	}

	if err := ValidateWorkflow(workflow); err != nil {
		return nil, fmt.Errorf("generated change preparation workflow is invalid: %w", err)
	}

	return workflow, nil
}

func (s *Service) BuildKnowledgeExplorationWorkflow(ctx stdcontext.Context, repositoryID string, query string) (*WorkflowPackage, error) {
	if repositoryID == "" {
		return nil, fmt.Errorf("repository ID cannot be empty")
	}
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}
	if s.sessionService == nil {
		return nil, fmt.Errorf("session service is required")
	}
	if s.searchService == nil {
		return nil, fmt.Errorf("search service is required")
	}

	session, err := s.sessionService.CreateRepositorySession(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository session: %w", err)
	}
	result, err := s.searchService.Search(ctx, repositoryID, query)
	if err != nil {
		return nil, fmt.Errorf("failed to build search result: %w", err)
	}

	sessionArtifact, err := newArtifact(ArtifactTypeSession, SourceSession, session)
	if err != nil {
		return nil, fmt.Errorf("failed to package repository session: %w", err)
	}
	searchArtifact, err := newArtifact(ArtifactTypeSearchResult, SourceSearch, result)
	if err != nil {
		return nil, fmt.Errorf("failed to package search result: %w", err)
	}

	workflow := &WorkflowPackage{
		WorkflowID:   buildWorkflowID(repositoryID, WorkflowTypeKnowledgeExploration, query),
		WorkflowType: WorkflowTypeKnowledgeExploration,
		RepositoryID: repositoryID,
		Version:      VersionV1,
		Artifacts: []*WorkflowArtifact{
			sessionArtifact,
			searchArtifact,
		},
	}

	if err := ValidateWorkflow(workflow); err != nil {
		return nil, fmt.Errorf("generated knowledge exploration workflow is invalid: %w", err)
	}

	return workflow, nil
}

func (s *Service) buildReviewerReport(ctx stdcontext.Context, repositoryID string, query string) (*reviewers.ReviewerRecommendationReport, error) {
	if domain, ok := parseDomainQuery(query); ok {
		return s.reviewerService.RecommendDomainReviewers(ctx, repositoryID, domain)
	}
	return s.reviewerService.RecommendRepositoryReviewers(ctx, repositoryID)
}

func (s *Service) buildChangePlan(ctx stdcontext.Context, repositoryID string, entityID string) (*changeplan.ChangePlan, error) {
	result, err := s.searchService.Search(ctx, repositoryID, entityID)
	if err != nil {
		return nil, err
	}

	for _, hit := range result.Hits {
		if hit.EntityID != entityID {
			continue
		}

		switch hit.EntityType {
		case agentsearch.EntityTypeDecision:
			return s.changePlanService.GenerateDecisionPlan(ctx, repositoryID, entityID)
		case agentsearch.EntityTypeFact:
			return s.changePlanService.GenerateFactPlan(ctx, repositoryID, entityID)
		case agentsearch.EntityTypeEvent:
			return s.changePlanService.GenerateEventPlan(ctx, repositoryID, entityID)
		case agentsearch.EntityTypeContributor:
			return s.changePlanService.GenerateContributorPlan(ctx, repositoryID, entityID)
		}
	}

	return nil, fmt.Errorf("unsupported change entity %q", entityID)
}

func parseDomainQuery(query string) (string, bool) {
	query = strings.TrimSpace(query)
	if !strings.HasPrefix(strings.ToLower(query), "domain:") {
		return "", false
	}

	domain := strings.TrimSpace(query[len("domain:"):])
	if domain == "" {
		return "", false
	}

	return domain, true
}

func newArtifact(artifactType string, source string, data any) (*WorkflowArtifact, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &WorkflowArtifact{
		ArtifactType: artifactType,
		Source:       source,
		Data:         json.RawMessage(payload),
	}, nil
}
