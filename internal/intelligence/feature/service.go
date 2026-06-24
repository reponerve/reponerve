package feature

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/reponerve/reponerve/internal/extraction/event"
	"github.com/reponerve/reponerve/internal/ownership/expertise"
	"github.com/reponerve/reponerve/internal/query/storage"
)

const sourceFeatureIntelligence = "feature_intelligence"

// Service derives feature summaries from repository memory.
type Service struct {
	eventReader     storage.EventReader
	expertiseReader storage.ExpertiseReader
	decisionReader  storage.DecisionReader
}

// NewService creates a feature intelligence service.
func NewService(
	eventReader storage.EventReader,
	expertiseReader storage.ExpertiseReader,
	decisionReader storage.DecisionReader,
) *Service {
	return &Service{
		eventReader:     eventReader,
		expertiseReader: expertiseReader,
		decisionReader:  decisionReader,
	}
}

// ListFeatures returns derived features for a repository.
func (s *Service) ListFeatures(ctx context.Context, repositoryID string) (*ListResult, error) {
	if repositoryID == "" {
		return nil, fmt.Errorf("repository ID cannot be empty")
	}
	byID := make(map[string]*Summary)

	for domain, keywords := range expertise.DomainKeywords {
		summary := &Summary{
			ID:       featureID(repositoryID, domain),
			Name:     domain,
			Keywords: append([]string(nil), keywords...),
			Sources:  []string{"expertise_domain"},
		}
		byID[summary.ID] = summary
	}

	events, err := s.eventReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, err
	}
	for _, evt := range events {
		if evt == nil || evt.EventType != event.EventTypeFeatureIntroduced {
			continue
		}
		name := normalizeFeatureName(evt.Title)
		if name == "" {
			continue
		}
		id := featureID(repositoryID, name)
		if existing, ok := byID[id]; ok {
			existing.EventCount++
			if !containsSource(existing.Sources, "feature_event") {
				existing.Sources = append(existing.Sources, "feature_event")
			}
			continue
		}
		byID[id] = &Summary{
			ID:         id,
			Name:       name,
			Keywords:   featureKeywords(name),
			Sources:    []string{"feature_event"},
			EventCount: 1,
		}
	}

	decisions, err := s.decisionReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, err
	}
	for _, d := range decisions {
		if d == nil {
			continue
		}
		if match := matchDomainDecision(d.Title); match != "" {
			id := featureID(repositoryID, match)
			if existing, ok := byID[id]; ok {
				if !containsSource(existing.Sources, "decision") {
					existing.Sources = append(existing.Sources, "decision")
				}
				continue
			}
		}
	}

	out := make([]Summary, 0, len(byID))
	for _, f := range byID {
		sort.Strings(f.Sources)
		out = append(out, *f)
	}
	sort.Slice(out, func(i, j int) bool {
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})

	return &ListResult{
		Features:       out,
		SourceServices: []string{sourceFeatureIntelligence},
	}, nil
}

// MatchFeature finds the best feature for a natural-language topic.
func (s *Service) MatchFeature(ctx context.Context, repositoryID, topic string) (*Summary, error) {
	list, err := s.ListFeatures(ctx, repositoryID)
	if err != nil {
		return nil, err
	}
	normalized := strings.ToLower(strings.TrimSpace(topic))
	if normalized == "" {
		return nil, nil
	}

	var best *Summary
	bestScore := 0
	for i := range list.Features {
		f := &list.Features[i]
		score := scoreFeatureMatch(normalized, f)
		if score > bestScore {
			bestScore = score
			best = f
		}
	}
	if bestScore < 2 {
		return nil, nil
	}
	return best, nil
}

// ShouldAutoExplain reports whether the generic Explain path should route to
// feature intelligence. Multi-word topics (e.g. "metadata panel") remain on
// symbol/topic resolution unless they exactly name a derived feature.
func ShouldAutoExplain(topic string, f *Summary) bool {
	if f == nil {
		return false
	}
	t := strings.ToLower(strings.TrimSpace(topic))
	if t == "" {
		return false
	}
	if t == strings.ToLower(f.Name) {
		return true
	}
	if strings.Contains(t, " ") {
		return false
	}
	return scoreFeatureMatch(t, f) >= 8
}

func scoreFeatureMatch(topic string, f *Summary) int {
	score := 0
	name := strings.ToLower(f.Name)
	if topic == name {
		score += 10
	}
	if strings.Contains(topic, name) || strings.Contains(name, topic) {
		score += 6
	}
	for _, kw := range f.Keywords {
		kw = strings.ToLower(kw)
		if kw == "" {
			continue
		}
		if topic == kw {
			score += 5
		}
		if strings.Contains(topic, kw) {
			score += 2
		}
	}
	return score
}

func matchDomainDecision(title string) string {
	lower := strings.ToLower(title)
	for domain := range expertise.DomainKeywords {
		if strings.Contains(lower, strings.ToLower(domain)) {
			return domain
		}
	}
	return ""
}

func normalizeFeatureName(title string) string {
	title = strings.TrimSpace(title)
	if title == "" {
		return ""
	}
	if idx := strings.Index(title, ":"); idx >= 0 {
		title = strings.TrimSpace(title[idx+1:])
	}
	return strings.TrimSpace(title)
}

func featureKeywords(name string) []string {
	parts := strings.FieldsFunc(strings.ToLower(name), func(r rune) bool {
		return !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9')
	})
	seen := make(map[string]struct{})
	var out []string
	for _, p := range parts {
		if len(p) < 3 {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return out
}

func featureID(repositoryID, name string) string {
	h := sha256.Sum256([]byte(repositoryID + "|feature|" + strings.ToLower(name)))
	return "feat_" + hex.EncodeToString(h[:8])
}

func containsSource(sources []string, source string) bool {
	for _, s := range sources {
		if s == source {
			return true
		}
	}
	return false
}

// SearchTerms returns terms for code/repository search for a matched feature.
func (f *Summary) SearchTerms() []string {
	if f == nil {
		return nil
	}
	seen := make(map[string]struct{})
	var terms []string
	add := func(t string) {
		t = strings.ToLower(strings.TrimSpace(t))
		if len(t) < 3 {
			return
		}
		if _, ok := seen[t]; ok {
			return
		}
		seen[t] = struct{}{}
		terms = append(terms, t)
	}
	add(f.Name)
	for _, kw := range f.Keywords {
		add(kw)
	}
	return terms
}
