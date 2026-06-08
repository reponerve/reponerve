package query

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"sort"
	"strings"

	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
	"github.com/reponerve/reponerve/internal/query/storage"
	models "github.com/reponerve/reponerve/pkg/models"
)

var authorRegex = regexp.MustCompile(`^([^<]+)\s*<([^>]+)>$`)

// Reader provides deterministic query capabilities for ownership intelligence.
type Reader struct {
	contribReader   storage.ContributorReader
	expertiseReader storage.ExpertiseReader
	sourceReader    storage.SourceReader
	decisionReader  storage.DecisionReader
	factReader      storage.FactReader
	eventReader     storage.EventReader
}

// NewReader creates a new high-level ownership Reader.
func NewReader(
	contribReader storage.ContributorReader,
	expertiseReader storage.ExpertiseReader,
	sourceReader storage.SourceReader,
	decisionReader storage.DecisionReader,
	factReader storage.FactReader,
	eventReader storage.EventReader,
) *Reader {
	return &Reader{
		contribReader:   contribReader,
		expertiseReader: expertiseReader,
		sourceReader:    sourceReader,
		decisionReader:  decisionReader,
		factReader:      factReader,
		eventReader:     eventReader,
	}
}

// ListContributors lists contributors for a repository, sorted by Name (with ID fallback).
func (r *Reader) ListContributors(ctx context.Context, repositoryID string) ([]*models.Contributor, error) {
	list, err := r.contribReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to list contributors: %w", err)
	}

	sort.Slice(list, func(i, j int) bool {
		if list[i].Name == list[j].Name {
			return list[i].ID < list[j].ID
		}
		return list[i].Name < list[j].Name
	})

	if list == nil {
		list = make([]*models.Contributor, 0)
	}
	return list, nil
}

// GetContributor retrieves a specific contributor.
func (r *Reader) GetContributor(ctx context.Context, repositoryID string, contributorID string) (*models.Contributor, error) {
	c, err := r.contribReader.GetByID(ctx, repositoryID, contributorID)
	if err != nil {
		return nil, err // Preserve underlying errors such as sql.ErrNoRows
	}
	return c, nil
}

// ListExpertise lists all expertise records for a repository, sorted by Domain (with ID fallback).
func (r *Reader) ListExpertise(ctx context.Context, repositoryID string) ([]*models.Expertise, error) {
	list, err := r.expertiseReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to list expertise: %w", err)
	}

	sort.Slice(list, func(i, j int) bool {
		if list[i].Domain == list[j].Domain {
			return list[i].ID < list[j].ID
		}
		return list[i].Domain < list[j].Domain
	})

	if list == nil {
		list = make([]*models.Expertise, 0)
	}
	return list, nil
}

// ListContributorExpertise lists expertise for a contributor, sorted by Domain (with ID fallback).
func (r *Reader) ListContributorExpertise(ctx context.Context, repositoryID string, contributorID string) ([]*models.Expertise, error) {
	list, err := r.expertiseReader.ListByContributor(ctx, repositoryID, contributorID)
	if err != nil {
		return nil, fmt.Errorf("failed to list contributor expertise: %w", err)
	}

	sort.Slice(list, func(i, j int) bool {
		if list[i].Domain == list[j].Domain {
			return list[i].ID < list[j].ID
		}
		return list[i].Domain < list[j].Domain
	})

	if list == nil {
		list = make([]*models.Expertise, 0)
	}
	return list, nil
}

// TraceContributor generates a complete ownership view for a contributor.
func (r *Reader) TraceContributor(ctx context.Context, repositoryID string, contributorID string) (*ContributorTrace, error) {
	c, err := r.GetContributor(ctx, repositoryID, contributorID)
	if err != nil {
		return nil, err
	}

	exps, err := r.ListContributorExpertise(ctx, repositoryID, contributorID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch contributor expertise: %w", err)
	}

	sources, err := r.sourceReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repository sources: %w", err)
	}

	// Map SourceID -> ContributorID
	sourceToContributor := make(map[string]string)
	for _, src := range sources {
		cID := contributorIDForSource(src)
		if cID != "" {
			sourceToContributor[src.ID] = cID
		}
	}

	// Fetch & filter Decisions
	allDecs, err := r.decisionReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch decisions: %w", err)
	}
	filteredDecs := make([]*memorymodels.Decision, 0)
	for _, d := range allDecs {
		if sourceToContributor[d.SourceID] == contributorID {
			filteredDecs = append(filteredDecs, d)
		}
	}
	sort.Slice(filteredDecs, func(i, j int) bool {
		if filteredDecs[i].CreatedAt.Equal(filteredDecs[j].CreatedAt) {
			return filteredDecs[i].ID > filteredDecs[j].ID
		}
		return filteredDecs[i].CreatedAt.After(filteredDecs[j].CreatedAt)
	})

	// Fetch & filter Facts
	allFacts, err := r.factReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch facts: %w", err)
	}
	filteredFacts := make([]*memorymodels.Fact, 0)
	for _, f := range allFacts {
		if sourceToContributor[f.SourceID] == contributorID {
			filteredFacts = append(filteredFacts, f)
		}
	}
	sort.Slice(filteredFacts, func(i, j int) bool {
		if filteredFacts[i].Subject == filteredFacts[j].Subject {
			return filteredFacts[i].ID < filteredFacts[j].ID
		}
		return filteredFacts[i].Subject < filteredFacts[j].Subject
	})

	// Fetch & filter Events
	allEvents, err := r.eventReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch events: %w", err)
	}
	filteredEvents := make([]*models.Event, 0)
	for _, e := range allEvents {
		if sourceToContributor[e.SourceID] == contributorID {
			filteredEvents = append(filteredEvents, e)
		}
	}
	sort.Slice(filteredEvents, func(i, j int) bool {
		if filteredEvents[i].Timestamp.Equal(filteredEvents[j].Timestamp) {
			return filteredEvents[i].ID > filteredEvents[j].ID
		}
		return filteredEvents[i].Timestamp.After(filteredEvents[j].Timestamp)
	})

	return &ContributorTrace{
		Contributor: c,
		Expertise:   exps,
		Decisions:   filteredDecs,
		Facts:       filteredFacts,
		Events:      filteredEvents,
	}, nil
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
