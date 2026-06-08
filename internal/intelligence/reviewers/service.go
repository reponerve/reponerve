package reviewers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"reponerve/internal/graph/impact"
	"reponerve/internal/intelligence/discovery"
	memorymodels "reponerve/internal/memory/models"
	"reponerve/internal/ownership/expertise"
	"reponerve/internal/query/storage"
	models "reponerve/pkg/models"
)

var authorRegex = regexp.MustCompile(`^([^<]+)\s*<([^>]+)>$`)

// Service recommends reviewers based on repository intelligence, expertise, and impact.
type Service struct {
	discoveryService *discovery.Service
	decisionReader   storage.DecisionReader
	factReader       storage.FactReader
	eventReader      storage.EventReader
	contribReader    storage.ContributorReader
	expertiseReader  storage.ExpertiseReader
	sourceReader     storage.SourceReader
	impactService    *impact.Service
}

// NewService constructs a new reviewer Service.
func NewService(
	discoverySvc *discovery.Service,
	dr storage.DecisionReader,
	fr storage.FactReader,
	er storage.EventReader,
	cr storage.ContributorReader,
	expr storage.ExpertiseReader,
	sr storage.SourceReader,
	impactSvc *impact.Service,
) *Service {
	return &Service{
		discoveryService: discoverySvc,
		decisionReader:   dr,
		factReader:       fr,
		eventReader:      er,
		contribReader:    cr,
		expertiseReader:  expr,
		sourceReader:     sr,
		impactService:    impactSvc,
	}
}

// RecommendRepositoryReviewers lists overall strongest reviewers across the repository.
func (s *Service) RecommendRepositoryReviewers(ctx context.Context, repositoryID string) (*ReviewerRecommendationReport, error) {
	if repositoryID == "" {
		return nil, fmt.Errorf("repository ID cannot be empty")
	}

	report, err := s.discoveryService.Discover(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to discover knowledge: %w", err)
	}

	contribs, err := s.contribReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to list contributors: %w", err)
	}

	exps, err := s.expertiseReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to list expertise: %w", err)
	}

	// Discovery scores map by ContributorID
	discoveryScores := make(map[string]float64)
	for _, item := range report.Items {
		if item.EntityType == discovery.EntityTypeContributor {
			discoveryScores[item.EntityID] = item.Score
		}
	}

	// Group expertise by ContributorID
	contribExps := make(map[string][]string)
	for _, exp := range exps {
		contribExps[exp.ContributorID] = append(contribExps[exp.ContributorID], exp.Domain)
	}

	var recommendations []*ReviewerRecommendation
	for _, c := range contribs {
		expertiseList := contribExps[c.ID]
		expertiseCount := len(expertiseList)

		// Distinct domain count
		domainMap := make(map[string]bool)
		for _, domain := range expertiseList {
			domainMap[domain] = true
		}
		domainCount := len(domainMap)

		discoveryParticipation := discoveryScores[c.ID]

		score := float64(expertiseCount) + float64(domainCount) + discoveryParticipation

		evidence := map[string]interface{}{
			"expertise_count":         expertiseCount,
			"domain_count":            domainCount,
			"discovery_participation": discoveryParticipation,
		}
		evidenceBytes, _ := json.Marshal(evidence)

		explanation := fmt.Sprintf("Contributor is recommended because they participate in %d repository domains and maintain %d expertise areas.", domainCount, expertiseCount)

		rec := &ReviewerRecommendation{
			ContributorID: c.ID,
			Score:         score,
			EvidenceJSON:  string(evidenceBytes),
			Explanation:   explanation,
		}

		if err := ValidateRecommendation(rec); err != nil {
			return nil, fmt.Errorf("invalid recommendation generated: %w", err)
		}

		if score > 0 {
			recommendations = append(recommendations, rec)
		}
	}

	// Sort: Score DESC, ContributorID ASC
	sort.Slice(recommendations, func(i, j int) bool {
		if recommendations[i].Score != recommendations[j].Score {
			return recommendations[i].Score > recommendations[j].Score
		}
		return recommendations[i].ContributorID < recommendations[j].ContributorID
	})

	return &ReviewerRecommendationReport{Recommendations: recommendations}, nil
}

// RecommendDomainReviewers lists domain-specific reviewer recommendations.
func (s *Service) RecommendDomainReviewers(ctx context.Context, repositoryID string, domain string) (*ReviewerRecommendationReport, error) {
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

	contribs, err := s.contribReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to list contributors: %w", err)
	}

	exps, err := s.expertiseReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to list expertise: %w", err)
	}

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

	keywords := getDomainKeywords(domain)

	// Fetch all decisions, facts, and events to check domain relevance
	decisions, _ := s.decisionReader.ListByRepository(ctx, repositoryID)
	decMap := make(map[string]*memorymodels.Decision)
	for _, d := range decisions {
		decMap[d.ID] = d
	}

	facts, _ := s.factReader.ListByRepository(ctx, repositoryID)
	factMap := make(map[string]*memorymodels.Fact)
	for _, f := range facts {
		factMap[f.ID] = f
	}

	events, _ := s.eventReader.ListByRepository(ctx, repositoryID)
	eventMap := make(map[string]*models.Event)
	for _, ev := range events {
		eventMap[ev.ID] = ev
	}

	// For each item in report, check if relevant to keywords.
	// If yes, map to contributor and add score to discovery_participation.
	contribDiscoveryPart := make(map[string]float64)
	if len(keywords) > 0 {
		for _, item := range report.Items {
			var isRelevant bool
			var sourceID string

			switch item.EntityType {
			case discovery.EntityTypeDecision:
				if d, ok := decMap[item.EntityID]; ok {
					isRelevant = matchesKeywords(d.Title, keywords)
					sourceID = d.SourceID
				}
			case discovery.EntityTypeFact:
				if f, ok := factMap[item.EntityID]; ok {
					isRelevant = matchesKeywords(f.Subject, keywords) || matchesKeywords(f.Predicate, keywords) || matchesKeywords(f.Object, keywords)
					sourceID = f.SourceID
				}
			case discovery.EntityTypeEvent:
				if ev, ok := eventMap[item.EntityID]; ok {
					isRelevant = matchesKeywords(ev.Title, keywords) || matchesKeywords(ev.Description, keywords)
					sourceID = ev.SourceID
				}
			}

			if isRelevant && sourceID != "" {
				if cID, ok := sourceToContributor[sourceID]; ok {
					contribDiscoveryPart[cID] += item.Score
				}
			}
		}
	}

	// Group matching expertise by ContributorID
	matchingExps := make(map[string]float64)
	for _, exp := range exps {
		if strings.EqualFold(exp.Domain, domain) {
			matchingExps[exp.ContributorID] += exp.Score
		}
	}

	var recommendations []*ReviewerRecommendation
	for _, c := range contribs {
		matchingExpertise := matchingExps[c.ID]
		discoveryPart := contribDiscoveryPart[c.ID]

		score := matchingExpertise + discoveryPart

		evidence := map[string]interface{}{
			"domain":                  domain,
			"matching_expertise":      matchingExpertise,
			"discovery_participation": discoveryPart,
		}
		evidenceBytes, _ := json.Marshal(evidence)

		explanation := "Contributor is recommended because their expertise matches the selected repository domain."

		rec := &ReviewerRecommendation{
			ContributorID: c.ID,
			Score:         score,
			EvidenceJSON:  string(evidenceBytes),
			Explanation:   explanation,
		}

		if err := ValidateRecommendation(rec); err != nil {
			return nil, fmt.Errorf("invalid recommendation generated: %w", err)
		}

		if score > 0 {
			recommendations = append(recommendations, rec)
		}
	}

	// Sort: Score DESC, ContributorID ASC
	sort.Slice(recommendations, func(i, j int) bool {
		if recommendations[i].Score != recommendations[j].Score {
			return recommendations[i].Score > recommendations[j].Score
		}
		return recommendations[i].ContributorID < recommendations[j].ContributorID
	})

	return &ReviewerRecommendationReport{Recommendations: recommendations}, nil
}

// RecommendImpactReviewers lists reviewer recommendations based on impact analysis.
func (s *Service) RecommendImpactReviewers(ctx context.Context, repositoryID string, entityID string) (*ReviewerRecommendationReport, error) {
	if repositoryID == "" {
		return nil, fmt.Errorf("repository ID cannot be empty")
	}
	if entityID == "" {
		return nil, fmt.Errorf("entity ID cannot be empty")
	}

	// 1. Determine entity type of entityID
	var entityType string
	var err error

	if _, err = s.decisionReader.GetByID(ctx, entityID); err == nil {
		entityType = "DECISION"
	} else if _, err = s.factReader.GetByID(ctx, entityID); err == nil {
		entityType = "FACT"
	} else if _, err = s.eventReader.GetByID(ctx, entityID); err == nil {
		entityType = "EVENT"
	}

	if entityType == "" {
		return &ReviewerRecommendationReport{Recommendations: []*ReviewerRecommendation{}}, nil
	}

	// 2. Perform impact analysis
	var impReport *impact.ImpactReport
	switch entityType {
	case "DECISION":
		impReport, err = s.impactService.AnalyzeDecisionImpact(ctx, repositoryID, entityID)
	case "FACT":
		impReport, err = s.impactService.AnalyzeFactImpact(ctx, repositoryID, entityID)
	case "EVENT":
		impReport, err = s.impactService.AnalyzeEventImpact(ctx, repositoryID, entityID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to analyze impact: %w", err)
	}

	// Collect all impacted entities
	impactedEntities := make(map[string]string) // EntityID -> NodeType
	impactedEntities[entityID] = entityType

	for _, path := range impReport.ImpactPaths {
		if path.Path != nil {
			for _, n := range path.Path.Nodes {
				if n != nil && n.EntityID != "" {
					impactedEntities[n.EntityID] = string(n.NodeType)
				}
			}
		}
	}

	// 3. Load other readers
	contribs, err := s.contribReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to list contributors: %w", err)
	}

	exps, err := s.expertiseReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to list expertise: %w", err)
	}

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

	// Group contributor expertise by contributor ID
	contribExpertise := make(map[string][]*models.Expertise)
	for _, exp := range exps {
		contribExpertise[exp.ContributorID] = append(contribExpertise[exp.ContributorID], exp)
	}

	// Load discovery report to consume rankings
	discReport, err := s.discoveryService.Discover(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to discover knowledge: %w", err)
	}

	discScores := make(map[string]float64)
	for _, item := range discReport.Items {
		discScores[item.EntityID] = item.Score
	}

	// Load details of decisions, facts, events to map to their source and keywords
	decisions, _ := s.decisionReader.ListByRepository(ctx, repositoryID)
	decMap := make(map[string]*memorymodels.Decision)
	for _, d := range decisions {
		decMap[d.ID] = d
	}

	facts, _ := s.factReader.ListByRepository(ctx, repositoryID)
	factMap := make(map[string]*memorymodels.Fact)
	for _, f := range facts {
		factMap[f.ID] = f
	}

	events, _ := s.eventReader.ListByRepository(ctx, repositoryID)
	eventMap := make(map[string]*models.Event)
	for _, ev := range events {
		eventMap[ev.ID] = ev
	}

	var recommendations []*ReviewerRecommendation
	for _, c := range contribs {
		// Calculate impact_entities: count of impacted entities authored by this contributor
		impactEntitiesCount := 0
		for entID := range impactedEntities {
			var sourceID string
			if d, ok := decMap[entID]; ok {
				sourceID = d.SourceID
			} else if f, ok := factMap[entID]; ok {
				sourceID = f.SourceID
			} else if ev, ok := eventMap[entID]; ok {
				sourceID = ev.SourceID
			}

			if sourceID != "" && sourceToContributor[sourceID] == c.ID {
				impactEntitiesCount++
			}
		}

		// Calculate matching_expertise
		matchingExpertiseScore := 0.0
		cExps := contribExpertise[c.ID]
		for entID := range impactedEntities {
			for _, exp := range cExps {
				keywords := getDomainKeywords(exp.Domain)
				if len(keywords) == 0 {
					continue
				}

				var matches bool
				if d, ok := decMap[entID]; ok {
					matches = matchesKeywords(d.Title, keywords)
				} else if f, ok := factMap[entID]; ok {
					matches = matchesKeywords(f.Subject, keywords) || matchesKeywords(f.Predicate, keywords) || matchesKeywords(f.Object, keywords)
				} else if ev, ok := eventMap[entID]; ok {
					matches = matchesKeywords(ev.Title, keywords) || matchesKeywords(ev.Description, keywords)
				}

				if matches {
					matchingExpertiseScore += exp.Score
				}
			}
		}

		// Calculate discovery_participation
		discoveryPart := 0.0
		for entID := range impactedEntities {
			var sourceID string
			if d, ok := decMap[entID]; ok {
				sourceID = d.SourceID
			} else if f, ok := factMap[entID]; ok {
				sourceID = f.SourceID
			} else if ev, ok := eventMap[entID]; ok {
				sourceID = ev.SourceID
			}

			if sourceID != "" && sourceToContributor[sourceID] == c.ID {
				if score, ok := discScores[entID]; ok {
					discoveryPart += score
				}
			}
		}

		score := float64(impactEntitiesCount) + matchingExpertiseScore + discoveryPart

		evidence := map[string]interface{}{
			"impact_entities":         impactEntitiesCount,
			"matching_expertise":      matchingExpertiseScore,
			"discovery_participation": discoveryPart,
		}
		evidenceBytes, _ := json.Marshal(evidence)

		explanation := "Contributor is recommended because their expertise overlaps with impacted repository knowledge."

		rec := &ReviewerRecommendation{
			ContributorID: c.ID,
			Score:         score,
			EvidenceJSON:  string(evidenceBytes),
			Explanation:   explanation,
		}

		if err := ValidateRecommendation(rec); err != nil {
			return nil, fmt.Errorf("invalid recommendation generated: %w", err)
		}

		if score > 0 {
			recommendations = append(recommendations, rec)
		}
	}

	// Sort: Score DESC, ContributorID ASC
	sort.Slice(recommendations, func(i, j int) bool {
		if recommendations[i].Score != recommendations[j].Score {
			return recommendations[i].Score > recommendations[j].Score
		}
		return recommendations[i].ContributorID < recommendations[j].ContributorID
	})

	return &ReviewerRecommendationReport{Recommendations: recommendations}, nil
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
