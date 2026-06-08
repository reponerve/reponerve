package expertise

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
	"github.com/reponerve/reponerve/pkg/models"
)

// DomainKeywords defines the domains and the keywords used to detect contributor expertise.
var DomainKeywords = map[string][]string{
	"Authentication":     {"auth", "jwt", "login", "password", "token", "session", "oauth", "credential"},
	"Storage":            {"storage", "db", "database", "sqlite", "postgres", "sql", "persistence", "query"},
	"Context Engine":     {"context", "generator", "briefing", "render", "template"},
	"MCP":                {"mcp", "protocol", "stdio", "transport", "tool"},
	"Agent Intelligence": {"agent", "onboarding", "guidance", "impact", "qa", "question", "answer"},
	"Infrastructure":     {"infra", "docker", "pipeline", "ci/cd", "actions", "deploy", "setup"},
}

// Evidence holds the counts of matches and recent activity indicator.
type Evidence struct {
	CommitCount    int  `json:"commit_count"`
	DecisionCount  int  `json:"decision_count"`
	FactCount      int  `json:"fact_count"`
	EventCount     int  `json:"event_count"`
	RecentActivity bool `json:"recent_activity"`
}

// Detector identifies domain-based expertise from repository memory.
type Detector struct{}

// NewDetector creates a new Detector instance.
func NewDetector() *Detector {
	return &Detector{}
}

// authorRegex matches conventional Git author format "Name <email>"
var authorRegex = regexp.MustCompile(`^([^<]+)\s*<([^>]+)>$`)

type contributorDomainMetrics struct {
	commits          int
	decisions        int
	facts            int
	events           int
	lastActivityTime time.Time
}

// Detect processes contributors, events, decisions, facts, and sources to calculate expertise scores and evidence.
func (d *Detector) Detect(
	ctx context.Context,
	contributors []*models.Contributor,
	events []*models.Event,
	decisions []*memorymodels.Decision,
	facts []*memorymodels.Fact,
	sources []*models.Source,
) ([]*models.Expertise, error) {
	if len(contributors) == 0 {
		return nil, nil
	}

	// 1. Calculate latest commit timestamp across all git sources
	var latestRepoTimestamp time.Time
	for _, src := range sources {
		if src.SourceType == "commit" {
			if src.Timestamp.After(latestRepoTimestamp) {
				latestRepoTimestamp = src.Timestamp
			}
		}
	}

	// 2. Index contributors by ID and build SourceID -> ContributorID mapping
	knownContributors := make(map[string]*models.Contributor)
	for _, c := range contributors {
		knownContributors[c.ID] = c
	}

	sourceToContributor := make(map[string]string)
	for _, src := range sources {
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
			continue
		}
		cID := contributorID(src.RepositoryID, name, email)
		if _, exists := knownContributors[cID]; exists {
			sourceToContributor[src.ID] = cID
		}
	}

	// 3. Pre-initialize metrics structure for all known contributors and domains
	metricsMap := make(map[string]map[string]*contributorDomainMetrics)
	for _, c := range contributors {
		metricsMap[c.ID] = make(map[string]*contributorDomainMetrics)
		for domain := range DomainKeywords {
			metricsMap[c.ID][domain] = &contributorDomainMetrics{}
		}
	}

	// Helper to match text against keywords
	matchesDomain := func(text string, keywords []string) bool {
		lowerText := strings.ToLower(text)
		for _, kw := range keywords {
			if strings.Contains(lowerText, kw) {
				return true
			}
		}
		return false
	}

	// Helper to update max activity time
	updateActivityTime := func(metrics *contributorDomainMetrics, t time.Time) {
		if metrics.lastActivityTime.IsZero() || t.After(metrics.lastActivityTime) {
			metrics.lastActivityTime = t
		}
	}

	// 4. Count matches for Commits (sources)
	for _, src := range sources {
		if src.SourceType != "commit" {
			continue
		}
		cID := getContributorIDForSource(src)
		if _, ok := metricsMap[cID]; !ok {
			continue
		}
		for domain, keywords := range DomainKeywords {
			if matchesDomain(src.Title, keywords) {
				metrics := metricsMap[cID][domain]
				metrics.commits++
				updateActivityTime(metrics, src.Timestamp)
			}
		}
	}

	// 5. Count matches for Decisions
	for _, dec := range decisions {
		cID, ok := sourceToContributor[dec.SourceID]
		if !ok {
			continue
		}
		if _, exists := metricsMap[cID]; !exists {
			continue
		}
		for domain, keywords := range DomainKeywords {
			if matchesDomain(dec.Title, keywords) {
				metrics := metricsMap[cID][domain]
				metrics.decisions++
				updateActivityTime(metrics, dec.CreatedAt)
			}
		}
	}

	// 6. Count matches for Facts
	for _, fact := range facts {
		cID, ok := sourceToContributor[fact.SourceID]
		if !ok {
			continue
		}
		if _, exists := metricsMap[cID]; !exists {
			continue
		}
		for domain, keywords := range DomainKeywords {
			if matchesDomain(fact.Subject, keywords) || matchesDomain(fact.Predicate, keywords) || matchesDomain(fact.Object, keywords) {
				metrics := metricsMap[cID][domain]
				metrics.facts++
				updateActivityTime(metrics, fact.CreatedAt)
			}
		}
	}

	// 7. Count matches for Events
	for _, event := range events {
		cID, ok := sourceToContributor[event.SourceID]
		if !ok {
			continue
		}
		if _, exists := metricsMap[cID]; !exists {
			continue
		}
		for domain, keywords := range DomainKeywords {
			if matchesDomain(event.Title, keywords) || matchesDomain(event.Description, keywords) {
				metrics := metricsMap[cID][domain]
				metrics.events++
				updateActivityTime(metrics, event.Timestamp)
			}
		}
	}

	// 8. Find max rawScore in each domain across all contributors
	maxRawScoreInDomain := make(map[string]float64)
	for domain := range DomainKeywords {
		var maxScore float64
		for _, c := range contributors {
			m := metricsMap[c.ID][domain]
			rawScore := (float64(m.commits) * 1.0) + (float64(m.decisions) * 5.0) + (float64(m.facts) * 2.0) + (float64(m.events) * 3.0)
			if rawScore > maxScore {
				maxScore = rawScore
			}
		}
		maxRawScoreInDomain[domain] = maxScore
	}

	// 9. Build Expertise slice
	var expertiseRecords []*models.Expertise
	for _, c := range contributors {
		for domain := range DomainKeywords {
			m := metricsMap[c.ID][domain]
			rawScore := (float64(m.commits) * 1.0) + (float64(m.decisions) * 5.0) + (float64(m.facts) * 2.0) + (float64(m.events) * 3.0)
			if rawScore == 0 {
				continue
			}

			score := 0.0
			if maxRawScoreInDomain[domain] > 0 {
				score = rawScore / maxRawScoreInDomain[domain]
			}

			// Check recency (<= 30 days difference from latestRepoTimestamp)
			recentActivity := false
			if !latestRepoTimestamp.IsZero() && !m.lastActivityTime.IsZero() {
				diff := latestRepoTimestamp.Sub(m.lastActivityTime)
				if diff <= 30*24*time.Hour {
					recentActivity = true
				}
			}

			evidenceObj := Evidence{
				CommitCount:    m.commits,
				DecisionCount:  m.decisions,
				FactCount:      m.facts,
				EventCount:     m.events,
				RecentActivity: recentActivity,
			}

			bytes, err := json.Marshal(evidenceObj)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal evidence for contributor %s, domain %s: %w", c.ID, domain, err)
			}

			expertiseRecords = append(expertiseRecords, &models.Expertise{
				ID:            expertiseID(c.RepositoryID, c.ID, domain),
				RepositoryID:  c.RepositoryID,
				ContributorID: c.ID,
				Domain:        domain,
				Score:         score,
				EvidenceJSON:  string(bytes),
			})
		}
	}

	// 10. Sort returned slice by ID for deterministic order
	sort.Slice(expertiseRecords, func(i, j int) bool {
		return expertiseRecords[i].ID < expertiseRecords[j].ID
	})

	return expertiseRecords, nil
}

func getContributorIDForSource(src *models.Source) string {
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

func expertiseID(repositoryID, contributorID, domain string) string {
	h := sha256.Sum256([]byte(repositoryID + ":" + contributorID + ":" + domain))
	return "exp_" + hex.EncodeToString(h[:])
}
