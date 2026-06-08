package learning

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/reponerve/reponerve/internal/graph/relationships"
	"github.com/reponerve/reponerve/internal/intelligence/discovery"
	"github.com/reponerve/reponerve/internal/ownership/expertise"
	"github.com/reponerve/reponerve/internal/query/storage"
	models "github.com/reponerve/reponerve/pkg/models"
)

var authorRegex = regexp.MustCompile(`^([^<]+)\s*<([^>]+)>$`)

// Service generates structured learning paths from repository intelligence.
type Service struct {
	discoveryService   *discovery.Service
	decisionReader     storage.DecisionReader
	factReader         storage.FactReader
	eventReader        storage.EventReader
	contribReader      storage.ContributorReader
	expertiseReader    storage.ExpertiseReader
	sourceReader       storage.SourceReader
	relationshipEngine *relationships.Engine
}

// NewService constructs a new learning Service.
func NewService(
	discoverySvc *discovery.Service,
	dr storage.DecisionReader,
	fr storage.FactReader,
	er storage.EventReader,
	cr storage.ContributorReader,
	expr storage.ExpertiseReader,
	sr storage.SourceReader,
	relEngine *relationships.Engine,
) *Service {
	return &Service{
		discoveryService:   discoverySvc,
		decisionReader:     dr,
		factReader:         fr,
		eventReader:        er,
		contribReader:      cr,
		expertiseReader:    expr,
		sourceReader:       sr,
		relationshipEngine: relEngine,
	}
}

// GenerateRepositoryPath builds a sequenced repository overview learning path.
func (s *Service) GenerateRepositoryPath(ctx context.Context, repositoryID string) (*LearningPath, error) {
	if repositoryID == "" {
		return nil, fmt.Errorf("repository ID cannot be empty")
	}

	report, err := s.discoveryService.Discover(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to discover knowledge: %w", err)
	}

	// Bucketing categories
	var decisions []*discovery.DiscoveryItem
	var facts []*discovery.DiscoveryItem
	var events []*discovery.DiscoveryItem
	var contributors []*discovery.DiscoveryItem

	for _, item := range report.Items {
		switch item.EntityType {
		case discovery.EntityTypeDecision:
			decisions = append(decisions, item)
		case discovery.EntityTypeFact:
			facts = append(facts, item)
		case discovery.EntityTypeEvent:
			events = append(events, item)
		case discovery.EntityTypeContributor:
			contributors = append(contributors, item)
		}
	}

	// Sequenced order: Decisions -> Facts -> Events -> Contributors
	var orderedItems []*discovery.DiscoveryItem
	orderedItems = append(orderedItems, decisions...)
	orderedItems = append(orderedItems, facts...)
	orderedItems = append(orderedItems, events...)
	orderedItems = append(orderedItems, contributors...)

	steps := make([]*LearningStep, 0, len(orderedItems))
	for i, item := range orderedItems {
		var evidenceMap map[string]interface{}
		if err := json.Unmarshal([]byte(item.EvidenceJSON), &evidenceMap); err != nil {
			evidenceMap = make(map[string]interface{})
		}
		evidenceMap["discovery_score"] = item.Score
		evidenceMap["position_reason"] = "repository_foundation"

		evBytes, _ := json.Marshal(evidenceMap)

		explanation := fmt.Sprintf("This %s appears early because it is highly ranked repository knowledge and provides foundational context.", strings.ToLower(item.EntityType))

		step := &LearningStep{
			EntityType:   item.EntityType,
			EntityID:     item.EntityID,
			Position:     i + 1,
			EvidenceJSON: string(evBytes),
			Explanation:  explanation,
		}

		if err := ValidateStep(step); err != nil {
			return nil, fmt.Errorf("failed to validate learning step: %w", err)
		}
		steps = append(steps, step)
	}

	return &LearningPath{Steps: steps}, nil
}

// GenerateDomainPath builds a sequenced domain-specific learning path.
func (s *Service) GenerateDomainPath(ctx context.Context, repositoryID string, domain string) (*LearningPath, error) {
	if repositoryID == "" {
		return nil, fmt.Errorf("repository ID cannot be empty")
	}
	if domain == "" {
		return nil, fmt.Errorf("domain cannot be empty")
	}

	report, err := s.discoveryService.Discover(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to discover knowledge: %w", err)
	}

	keywords := getDomainKeywords(domain)
	if keywords == nil {
		return &LearningPath{Steps: []*LearningStep{}}, nil
	}

	// 1. Identify domain matched entities
	matchingDecisions := make(map[string]bool)
	decs, err := s.decisionReader.ListByRepository(ctx, repositoryID)
	if err == nil {
		for _, d := range decs {
			if matchesKeywords(d.Title, keywords) {
				matchingDecisions[d.ID] = true
			}
		}
	}

	matchingFacts := make(map[string]bool)
	facts, err := s.factReader.ListByRepository(ctx, repositoryID)
	if err == nil {
		for _, f := range facts {
			if matchesKeywords(f.Subject, keywords) || matchesKeywords(f.Predicate, keywords) || matchesKeywords(f.Object, keywords) {
				matchingFacts[f.ID] = true
			}
		}
	}

	matchingEvents := make(map[string]bool)
	events, err := s.eventReader.ListByRepository(ctx, repositoryID)
	if err == nil {
		for _, ev := range events {
			if matchesKeywords(ev.Title, keywords) || matchesKeywords(ev.Description, keywords) {
				matchingEvents[ev.ID] = true
			}
		}
	}

	matchingContributors := make(map[string]bool)
	exps, err := s.expertiseReader.ListByRepository(ctx, repositoryID)
	if err == nil {
		for _, exp := range exps {
			if strings.EqualFold(exp.Domain, domain) {
				matchingContributors[exp.ContributorID] = true
			}
		}
	}

	// Filter and group DiscoveryItems
	var contributors []*discovery.DiscoveryItem
	var decisions []*discovery.DiscoveryItem
	var factsList []*discovery.DiscoveryItem
	var eventsList []*discovery.DiscoveryItem

	for _, item := range report.Items {
		switch item.EntityType {
		case discovery.EntityTypeContributor:
			if matchingContributors[item.EntityID] {
				contributors = append(contributors, item)
			}
		case discovery.EntityTypeDecision:
			if matchingDecisions[item.EntityID] {
				decisions = append(decisions, item)
			}
		case discovery.EntityTypeFact:
			if matchingFacts[item.EntityID] {
				factsList = append(factsList, item)
			}
		case discovery.EntityTypeEvent:
			if matchingEvents[item.EntityID] {
				eventsList = append(eventsList, item)
			}
		}
	}

	// Domain prioritization: Contributors (expertise domain matches) -> Decisions -> Facts -> Events
	var orderedItems []*discovery.DiscoveryItem
	orderedItems = append(orderedItems, contributors...)
	orderedItems = append(orderedItems, decisions...)
	orderedItems = append(orderedItems, factsList...)
	orderedItems = append(orderedItems, eventsList...)

	steps := make([]*LearningStep, 0, len(orderedItems))
	for i, item := range orderedItems {
		var evidenceMap map[string]interface{}
		if err := json.Unmarshal([]byte(item.EvidenceJSON), &evidenceMap); err != nil {
			evidenceMap = make(map[string]interface{})
		}
		evidenceMap["discovery_score"] = item.Score

		reason := "domain_" + strings.ToLower(item.EntityType)
		if item.EntityType == discovery.EntityTypeContributor {
			reason = "domain_expertise"
		}
		evidenceMap["position_reason"] = reason

		evBytes, _ := json.Marshal(evidenceMap)

		explanation := fmt.Sprintf("This %s appears early because it is relevant to the selected repository domain.", strings.ToLower(item.EntityType))

		step := &LearningStep{
			EntityType:   item.EntityType,
			EntityID:     item.EntityID,
			Position:     i + 1,
			EvidenceJSON: string(evBytes),
			Explanation:  explanation,
		}

		if err := ValidateStep(step); err != nil {
			return nil, fmt.Errorf("failed to validate learning step: %w", err)
		}
		steps = append(steps, step)
	}

	return &LearningPath{Steps: steps}, nil
}

// GenerateContributorPath builds a sequenced contributor-specific learning path.
func (s *Service) GenerateContributorPath(ctx context.Context, repositoryID string, contributorID string) (*LearningPath, error) {
	if repositoryID == "" {
		return nil, fmt.Errorf("repository ID cannot be empty")
	}
	if contributorID == "" {
		return nil, fmt.Errorf("contributor ID cannot be empty")
	}

	report, err := s.discoveryService.Discover(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to discover knowledge: %w", err)
	}

	// 1. Resolve contributor mapping to decisions, facts, events
	sources, err := s.sourceReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to list sources: %w", err)
	}

	sourceToContributor := make(map[string]string)
	for _, src := range sources {
		cID := contributorIDForSource(src)
		if cID != "" {
			sourceToContributor[src.ID] = cID
		}
	}

	matchingDecisions := make(map[string]bool)
	decs, err := s.decisionReader.ListByRepository(ctx, repositoryID)
	if err == nil {
		for _, d := range decs {
			if sourceToContributor[d.SourceID] == contributorID {
				matchingDecisions[d.ID] = true
			}
		}
	}

	matchingFacts := make(map[string]bool)
	facts, err := s.factReader.ListByRepository(ctx, repositoryID)
	if err == nil {
		for _, f := range facts {
			if sourceToContributor[f.SourceID] == contributorID {
				matchingFacts[f.ID] = true
			}
		}
	}

	matchingEvents := make(map[string]bool)
	events, err := s.eventReader.ListByRepository(ctx, repositoryID)
	if err == nil {
		for _, ev := range events {
			if sourceToContributor[ev.SourceID] == contributorID {
				matchingEvents[ev.ID] = true
			}
		}
	}

	// Filter and group DiscoveryItems
	var contributors []*discovery.DiscoveryItem
	var decisions []*discovery.DiscoveryItem
	var factsList []*discovery.DiscoveryItem
	var eventsList []*discovery.DiscoveryItem

	for _, item := range report.Items {
		switch item.EntityType {
		case discovery.EntityTypeContributor:
			if item.EntityID == contributorID {
				contributors = append(contributors, item)
			}
		case discovery.EntityTypeDecision:
			if matchingDecisions[item.EntityID] {
				decisions = append(decisions, item)
			}
		case discovery.EntityTypeFact:
			if matchingFacts[item.EntityID] {
				factsList = append(factsList, item)
			}
		case discovery.EntityTypeEvent:
			if matchingEvents[item.EntityID] {
				eventsList = append(eventsList, item)
			}
		}
	}

	// Contributor prioritization: Contributor profile -> Decisions -> Facts -> Events
	var orderedItems []*discovery.DiscoveryItem
	orderedItems = append(orderedItems, contributors...)
	orderedItems = append(orderedItems, decisions...)
	orderedItems = append(orderedItems, factsList...)
	orderedItems = append(orderedItems, eventsList...)

	steps := make([]*LearningStep, 0, len(orderedItems))
	for i, item := range orderedItems {
		var evidenceMap map[string]interface{}
		if err := json.Unmarshal([]byte(item.EvidenceJSON), &evidenceMap); err != nil {
			evidenceMap = make(map[string]interface{})
		}
		evidenceMap["discovery_score"] = item.Score

		reason := "contributor_" + strings.ToLower(item.EntityType)
		if item.EntityType == discovery.EntityTypeContributor {
			reason = "contributor_profile"
		}
		evidenceMap["position_reason"] = reason

		evBytes, _ := json.Marshal(evidenceMap)

		explanation := fmt.Sprintf("This %s appears early because it is associated with the selected contributor.", strings.ToLower(item.EntityType))
		if item.EntityType == discovery.EntityTypeContributor {
			explanation = "This expertise area appears early because it is associated with the selected contributor."
		}

		step := &LearningStep{
			EntityType:   item.EntityType,
			EntityID:     item.EntityID,
			Position:     i + 1,
			EvidenceJSON: string(evBytes),
			Explanation:  explanation,
		}

		if err := ValidateStep(step); err != nil {
			return nil, fmt.Errorf("failed to validate learning step: %w", err)
		}
		steps = append(steps, step)
	}

	return &LearningPath{Steps: steps}, nil
}

// Helpers

func getDomainKeywords(domain string) []string {
	for k, v := range expertise.DomainKeywords {
		if strings.EqualFold(k, domain) {
			return v
		}
	}
	return nil
}

func matchesKeywords(text string, keywords []string) bool {
	lowerText := strings.ToLower(text)
	for _, kw := range keywords {
		if strings.Contains(lowerText, kw) {
			return true
		}
	}
	return false
}

func contributorIDForSource(src *models.Source) string {
	name := strings.TrimSpace(src.Author)
	email := ""
	matches := authorRegex.FindStringSubmatch(src.Author)
	if len(matches) == 3 {
		name = strings.TrimSpace(matches[1])
		email = strings.TrimSpace(matches[2])
	} else if strings.Contains(src.Author, "@") && !strings.Contains(src.Author, " ") {
		email = strings.TrimSpace(src.Author)
		name = ""
	}
	if name == "" && email == "" {
		return ""
	}
	return contributorID(src.RepositoryID, name, email)
}

func contributorID(repositoryID, name, email string) string {
	var input string
	if email != "" {
		input = repositoryID + ":" + email
	} else {
		input = repositoryID + ":" + name
	}
	h := sha256.Sum256([]byte(input))
	return "ctr_" + hex.EncodeToString(h[:])
}
