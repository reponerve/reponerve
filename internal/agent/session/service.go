package agentsession

import (
	stdcontext "context"
	"encoding/json"
	"fmt"

	agentcontext "github.com/reponerve/reponerve/internal/agent/context"
	agentsearch "github.com/reponerve/reponerve/internal/agent/search"
)

type Service struct {
	contextService *agentcontext.Service
	searchService  *agentsearch.Service
}

func NewService(contextService *agentcontext.Service, searchService *agentsearch.Service) *Service {
	return &Service{
		contextService: contextService,
		searchService:  searchService,
	}
}

func (s *Service) CreateRepositorySession(ctx stdcontext.Context, repositoryID string) (*AgentSession, error) {
	if repositoryID == "" {
		return nil, fmt.Errorf("repository ID cannot be empty")
	}
	if s.contextService == nil {
		return nil, fmt.Errorf("context service is required")
	}

	contextPackage, err := s.contextService.BuildRepositoryContext(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to build repository context package: %w", err)
	}

	contextArtifact, err := newArtifact(ArtifactTypeContextPackage, SourceContext, contextPackage)
	if err != nil {
		return nil, fmt.Errorf("failed to package repository context: %w", err)
	}

	session := &AgentSession{
		SessionID:    buildSessionID(repositoryID, SessionTypeRepository, defaultRepositoryIdentifier),
		SessionType:  SessionTypeRepository,
		RepositoryID: repositoryID,
		Artifacts: []*SessionArtifact{
			contextArtifact,
		},
	}

	if err := ValidateSession(session); err != nil {
		return nil, fmt.Errorf("generated repository session is invalid: %w", err)
	}

	return session, nil
}

func (s *Service) CreateDomainSession(ctx stdcontext.Context, repositoryID string, domain string) (*AgentSession, error) {
	if repositoryID == "" {
		return nil, fmt.Errorf("repository ID cannot be empty")
	}
	if domain == "" {
		return nil, fmt.Errorf("domain cannot be empty")
	}
	if s.contextService == nil {
		return nil, fmt.Errorf("context service is required")
	}
	if s.searchService == nil {
		return nil, fmt.Errorf("search service is required")
	}

	contextPackage, err := s.contextService.BuildDomainContext(ctx, repositoryID, domain)
	if err != nil {
		return nil, fmt.Errorf("failed to build domain context package: %w", err)
	}
	searchResult, err := s.searchService.Search(ctx, repositoryID, "domain:"+domain)
	if err != nil {
		return nil, fmt.Errorf("failed to build domain search result: %w", err)
	}

	contextArtifact, err := newArtifact(ArtifactTypeContextPackage, SourceContext, contextPackage)
	if err != nil {
		return nil, fmt.Errorf("failed to package domain context: %w", err)
	}
	searchArtifact, err := newArtifact(ArtifactTypeSearchResult, SourceSearch, searchResult)
	if err != nil {
		return nil, fmt.Errorf("failed to package domain search result: %w", err)
	}

	session := &AgentSession{
		SessionID:    buildSessionID(repositoryID, SessionTypeDomain, domain),
		SessionType:  SessionTypeDomain,
		RepositoryID: repositoryID,
		Artifacts: []*SessionArtifact{
			contextArtifact,
			searchArtifact,
		},
	}

	if err := ValidateSession(session); err != nil {
		return nil, fmt.Errorf("generated domain session is invalid: %w", err)
	}

	return session, nil
}

func (s *Service) CreateContributorSession(ctx stdcontext.Context, repositoryID string, contributorID string) (*AgentSession, error) {
	if repositoryID == "" {
		return nil, fmt.Errorf("repository ID cannot be empty")
	}
	if contributorID == "" {
		return nil, fmt.Errorf("contributor ID cannot be empty")
	}
	if s.contextService == nil {
		return nil, fmt.Errorf("context service is required")
	}
	if s.searchService == nil {
		return nil, fmt.Errorf("search service is required")
	}

	contextPackage, err := s.contextService.BuildContributorContext(ctx, repositoryID, contributorID)
	if err != nil {
		return nil, fmt.Errorf("failed to build contributor context package: %w", err)
	}
	searchResult, err := s.searchService.Search(ctx, repositoryID, contributorID)
	if err != nil {
		return nil, fmt.Errorf("failed to build contributor search result: %w", err)
	}

	contextArtifact, err := newArtifact(ArtifactTypeContextPackage, SourceContext, contextPackage)
	if err != nil {
		return nil, fmt.Errorf("failed to package contributor context: %w", err)
	}
	searchArtifact, err := newArtifact(ArtifactTypeSearchResult, SourceSearch, searchResult)
	if err != nil {
		return nil, fmt.Errorf("failed to package contributor search result: %w", err)
	}

	session := &AgentSession{
		SessionID:    buildSessionID(repositoryID, SessionTypeContributor, contributorID),
		SessionType:  SessionTypeContributor,
		RepositoryID: repositoryID,
		Artifacts: []*SessionArtifact{
			contextArtifact,
			searchArtifact,
		},
	}

	if err := ValidateSession(session); err != nil {
		return nil, fmt.Errorf("generated contributor session is invalid: %w", err)
	}

	return session, nil
}

func newArtifact(artifactType string, source string, data any) (*SessionArtifact, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &SessionArtifact{
		ArtifactType: artifactType,
		Source:       source,
		Data:         json.RawMessage(payload),
	}, nil
}
