package feature

import (
	"context"
	"testing"

	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
	"github.com/reponerve/reponerve/pkg/models"
)

type stubEventReader struct {
	events []*models.Event
}

func (s *stubEventReader) GetByID(ctx context.Context, id string) (*models.Event, error) {
	return nil, nil
}
func (s *stubEventReader) ListByRepository(ctx context.Context, repoID string) ([]*models.Event, error) {
	return s.events, nil
}
func (s *stubEventReader) ListAll(ctx context.Context) ([]*models.Event, error) {
	return nil, nil
}

type stubExpertiseReader struct{}

func (s *stubExpertiseReader) GetByID(ctx context.Context, id string) (*models.Expertise, error) {
	return nil, nil
}
func (s *stubExpertiseReader) ListByRepository(ctx context.Context, repoID string) ([]*models.Expertise, error) {
	return nil, nil
}
func (s *stubExpertiseReader) ListByContributor(ctx context.Context, repositoryID, contributorID string) ([]*models.Expertise, error) {
	return nil, nil
}
func (s *stubExpertiseReader) ListAll(ctx context.Context) ([]*models.Expertise, error) {
	return nil, nil
}

type stubDecisionReader struct {
	decisions []*memorymodels.Decision
}

func (s *stubDecisionReader) GetByID(ctx context.Context, id string) (*memorymodels.Decision, error) {
	return nil, nil
}
func (s *stubDecisionReader) ListByRepository(ctx context.Context, repoID string) ([]*memorymodels.Decision, error) {
	return s.decisions, nil
}
func (s *stubDecisionReader) ListAll(ctx context.Context) ([]*memorymodels.Decision, error) {
	return nil, nil
}

func TestListFeatures_IncludesDomainsAndEvents(t *testing.T) {
	svc := NewService(
		&stubEventReader{events: []*models.Event{
			{RepositoryID: "repo1", EventType: "FEATURE_INTRODUCED", Title: "feat: billing webhooks"},
		}},
		&stubExpertiseReader{},
		&stubDecisionReader{},
	)
	out, err := svc.ListFeatures(context.Background(), "repo1")
	if err != nil {
		t.Fatalf("ListFeatures: %v", err)
	}
	if len(out.Features) < 2 {
		t.Fatalf("expected domains + event feature, got %d", len(out.Features))
	}
}

func TestMatchFeature_Authentication(t *testing.T) {
	svc := NewService(&stubEventReader{}, &stubExpertiseReader{}, &stubDecisionReader{})
	match, err := svc.MatchFeature(context.Background(), "repo1", "authentication")
	if err != nil {
		t.Fatalf("MatchFeature: %v", err)
	}
	if match == nil || match.Name != "Authentication" {
		t.Fatalf("match=%+v", match)
	}
	if !ShouldAutoExplain("authentication", match) {
		t.Fatal("expected authentication to auto-explain as feature")
	}
}

func TestShouldAutoExplain_MultiWordTopicStaysOffFeaturePath(t *testing.T) {
	svc := NewService(&stubEventReader{events: []*models.Event{
		{RepositoryID: "repo1", EventType: "FEATURE_INTRODUCED", Title: "feat: introduce metadata panel UI"},
	}}, &stubExpertiseReader{}, &stubDecisionReader{})
	match, err := svc.MatchFeature(context.Background(), "repo1", "metadata panel")
	if err != nil {
		t.Fatalf("MatchFeature: %v", err)
	}
	if match == nil {
		t.Fatal("expected feature match for metadata panel event")
	}
	if ShouldAutoExplain("metadata panel", match) {
		t.Fatal("multi-word symbol topics must not auto-route to feature explain")
	}
}
