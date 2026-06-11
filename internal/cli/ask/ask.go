package askcmd

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/agent/guidance"
	"github.com/reponerve/reponerve/internal/agent/impact"
	"github.com/reponerve/reponerve/internal/agent/onboarding"
	"github.com/reponerve/reponerve/internal/agent/qa"
	"github.com/reponerve/reponerve/internal/config"
	ctxengine "github.com/reponerve/reponerve/internal/context"
	"github.com/reponerve/reponerve/internal/query/storage"
	"github.com/reponerve/reponerve/internal/scanner/repository"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
)

var rxWhoCreated = regexp.MustCompile(`(?i)^who created\s+(.+?)\??$`)
var rxWhoWorkedOn = regexp.MustCompile(`(?i)^who worked on\s+(.+?)\??$`)
var rxWhoTouched = regexp.MustCompile(`(?i)^who touched\s+(.+?)\??$`)
var rxWhoOwns = regexp.MustCompile(`(?i)^who owns\s+(.+?)\??$`)

type authorHit struct {
	author string
	count  int
}

var stopwords = map[string]struct{}{
	"a": {}, "an": {}, "and": {}, "are": {}, "as": {}, "at": {}, "be": {},
	"by": {}, "for": {}, "from": {}, "in": {}, "is": {}, "it": {}, "of": {},
	"on": {}, "or": {}, "that": {}, "the": {}, "this": {}, "to": {}, "was": {},
	"were": {}, "who": {}, "with": {}, "what": {}, "created": {}, "worked": {},
	"touched": {}, "owns": {}, "own": {}, "component": {}, "feature": {}, "server": {},
}

// NewCommand creates and returns the ask subcommand.
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "ask [question]",
		Short: "Ask a question about the repository",
		Long:  `Retrieve repository memory and explain historical decisions based on developer queries.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			question := strings.TrimSpace(args[0])
			cmd.Printf("Querying repository memory for: %q...\n", question)

			workspaceDir := config.GetWorkspaceDir()
			cfg, err := config.Load(workspaceDir)
			if err != nil {
				return fmt.Errorf("workspace not initialized; run 'reponerve init' first")
			}

			db, err := sqlite.Open(cfg.Storage.SQLitePath)
			if err != nil {
				return fmt.Errorf("failed to open database: %w", err)
			}
			defer db.Close()

			discovery := repository.NewGitDiscovery()
			repo, err := discovery.Discover(cmd.Context(), cfg.Repository.Path)
			if err != nil {
				return fmt.Errorf("failed to discover repository: %w", err)
			}

			eventReader := storage.NewSQLiteEventReader(db)
			decisionReader := storage.NewSQLiteDecisionReader(db)
			intentReader := storage.NewSQLiteIntentReader(db)
			factReader := storage.NewSQLiteFactReader(db)
			relationshipReader := storage.NewSQLiteRelationshipReader(db)
			sourceReader := storage.NewSQLiteSourceReader(db)
			contributorReader := storage.NewSQLiteContributorReader(db)
			expertiseReader := storage.NewSQLiteExpertiseReader(db)

			ctxReader := ctxengine.NewMemoryContextReader(eventReader, decisionReader, intentReader, factReader)
			generator := ctxengine.NewGenerator(ctxReader)

			onboardingService := onboarding.NewService(generator)
			guidanceService := guidance.NewService(decisionReader, intentReader, factReader, eventReader, relationshipReader)
			impactService := impact.NewService(decisionReader, intentReader, factReader, eventReader, relationshipReader)
			qaService := qa.NewService(onboardingService, guidanceService, impactService)

			answer, err := qaService.Answer(cmd.Context(), repo.ID, qa.Question{Text: question})
			if err == nil {
				return printAnswer(cmd, answer)
			}

			if strings.Contains(strings.ToLower(err.Error()), "unknown question") {
				handled, fallbackErr := handleOwnershipStyleQuestion(
					cmd,
					sourceReader,
					contributorReader,
					expertiseReader,
					repo.ID,
					question,
				)
				if fallbackErr != nil {
					return fallbackErr
				}
				if handled {
					return nil
				}

				cmd.Println("No deterministic answer pattern matched this question.")
				cmd.Println("Try one of these formats:")
				cmd.Println(`  reponerve ask "What is this repository?"`)
				cmd.Println(`  reponerve ask "Why was decision <decision_id> made?"`)
				cmd.Println(`  reponerve ask "What caused event <event_id>?"`)
				cmd.Println(`  reponerve ask "What happens if decision <decision_id> changes?"`)
				cmd.Println(`  reponerve ask "Who created <feature or component>?"`)
				cmd.Println(`  reponerve ask "Who worked on <feature or component>?"`)
				cmd.Println(`  reponerve ask "Who touched <feature or component>?"`)
				cmd.Println(`  reponerve ask "Who owns <domain or component>?"`)
				return nil
			}

			return err
		},
	}
}

func printAnswer(cmd *cobra.Command, answer *qa.Answer) error {
	cmd.Println("Answer:")
	body, err := json.MarshalIndent(answer.Result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format answer: %w", err)
	}
	cmd.Println(string(body))
	return nil
}

func handleOwnershipStyleQuestion(
	cmd *cobra.Command,
	sourceReader storage.SourceReader,
	contributorReader storage.ContributorReader,
	expertiseReader storage.ExpertiseReader,
	repositoryID,
	question string,
) (bool, error) {
	trimmed := strings.TrimSpace(question)

	if matches := rxWhoCreated.FindStringSubmatch(trimmed); len(matches) > 1 {
		return handleSourceEvidenceQuestion(cmd, sourceReader, repositoryID, matches[1], "creators")
	}
	if matches := rxWhoWorkedOn.FindStringSubmatch(trimmed); len(matches) > 1 {
		return handleSourceEvidenceQuestion(cmd, sourceReader, repositoryID, matches[1], "contributors")
	}
	if matches := rxWhoTouched.FindStringSubmatch(trimmed); len(matches) > 1 {
		return handleSourceEvidenceQuestion(cmd, sourceReader, repositoryID, matches[1], "contributors")
	}
	if matches := rxWhoOwns.FindStringSubmatch(trimmed); len(matches) > 1 {
		return handleOwnershipByExpertiseQuestion(cmd, sourceReader, contributorReader, expertiseReader, repositoryID, matches[1])
	}

	return false, nil
}

func handleSourceEvidenceQuestion(
	cmd *cobra.Command,
	sourceReader storage.SourceReader,
	repositoryID,
	queryText,
	label string,
) (bool, error) {
	terms := extractQueryTerms(queryText)
	if len(terms) == 0 {
		return true, nil
	}

	sources, err := sourceReader.ListByRepository(cmd.Context(), repositoryID)
	if err != nil {
		return false, fmt.Errorf("failed to query sources: %w", err)
	}

	hitMap := make(map[string]int)
	for _, src := range sources {
		evidenceText := src.Title + " " + src.Reference + " " + src.MetadataJSON
		score := scoreTextMatch(evidenceText, terms)
		if score > 0 {
			author := strings.TrimSpace(src.Author)
			if author == "" {
				author = "unknown"
			}
			hitMap[author] += score
		}
	}

	if len(hitMap) == 0 {
		workspaceDir := config.GetWorkspaceDir()
		cfg, cfgErr := config.Load(workspaceDir)
		if cfgErr == nil {
			codeHits, codeErr := findContributorsFromCodeEvidence(cfg.Repository.Path, terms)
			if codeErr == nil && len(codeHits) > 0 {
				cmd.Printf("Possible %s for %q based on code history evidence:\n", label, queryText)
				for _, hit := range codeHits {
					cmd.Printf("- %s (evidence score %d)\n", hit.author, hit.count)
				}
				return true, nil
			}
		}

		cmd.Printf("No evidence found for %q in indexed sources or code history search.\n", queryText)
		cmd.Println("Tip: run `reponerve scan` again after the relevant commits/ADRs are present.")
		return true, nil
	}

	hits := make([]authorHit, 0, len(hitMap))
	for author, count := range hitMap {
		hits = append(hits, authorHit{author: author, count: count})
	}
	sort.Slice(hits, func(i, j int) bool {
		if hits[i].count == hits[j].count {
			return hits[i].author < hits[j].author
		}
		return hits[i].count > hits[j].count
	})

	cmd.Printf("Possible %s for %q based on source evidence:\n", label, queryText)
	for _, hit := range hits {
		cmd.Printf("- %s (evidence score %d)\n", hit.author, hit.count)
	}

	return true, nil
}

func findContributorsFromCodeEvidence(repoPath string, terms []string) ([]authorHit, error) {
	if len(terms) == 0 {
		return nil, nil
	}

	patternParts := make([]string, 0, len(terms))
	for _, t := range terms {
		if t == "" {
			continue
		}
		patternParts = append(patternParts, regexp.QuoteMeta(t))
	}
	if len(patternParts) == 0 {
		return nil, nil
	}
	pattern := strings.Join(patternParts, "|")

	grepCmd := exec.Command("git", "-C", repoPath, "grep", "-inE", "--no-color", pattern)
	out, err := grepCmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil, nil
		}
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	authorCount := make(map[string]int)
	const maxMatches = 50
	processed := 0

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 3)
		if len(parts) < 3 {
			continue
		}
		file := parts[0]
		lineNum, convErr := strconv.Atoi(parts[1])
		if convErr != nil {
			continue
		}
		lineText := parts[2]
		lineScore := scoreTextMatch(file+" "+lineText, terms)
		if lineScore == 0 {
			continue
		}

		author := blameAuthor(repoPath, file, lineNum)
		if author == "" {
			author = "unknown"
		}
		authorCount[author] += lineScore
		processed++
		if processed >= maxMatches {
			break
		}
	}

	if len(authorCount) == 0 {
		return nil, nil
	}

	hits := make([]authorHit, 0, len(authorCount))
	for author, count := range authorCount {
		hits = append(hits, authorHit{author: author, count: count})
	}
	sort.Slice(hits, func(i, j int) bool {
		if hits[i].count == hits[j].count {
			return hits[i].author < hits[j].author
		}
		return hits[i].count > hits[j].count
	})

	return hits, nil
}

func blameAuthor(repoPath, file string, lineNum int) string {
	blameCmd := exec.Command(
		"git", "-C", repoPath,
		"blame", "--line-porcelain",
		"-L", fmt.Sprintf("%d,%d", lineNum, lineNum),
		"--", file,
	)
	out, err := blameCmd.Output()
	if err != nil {
		return ""
	}

	for _, l := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(l, "author ") {
			return strings.TrimSpace(strings.TrimPrefix(l, "author "))
		}
	}
	return ""
}

func handleOwnershipByExpertiseQuestion(
	cmd *cobra.Command,
	sourceReader storage.SourceReader,
	contributorReader storage.ContributorReader,
	expertiseReader storage.ExpertiseReader,
	repositoryID,
	queryText string,
) (bool, error) {
	terms := extractQueryTerms(queryText)
	if len(terms) == 0 {
		return true, nil
	}

	contributors, err := contributorReader.ListByRepository(cmd.Context(), repositoryID)
	if err != nil {
		return false, fmt.Errorf("failed to list contributors: %w", err)
	}
	expertise, err := expertiseReader.ListByRepository(cmd.Context(), repositoryID)
	if err != nil {
		return false, fmt.Errorf("failed to list expertise: %w", err)
	}

	contribByID := make(map[string]string)
	for _, c := range contributors {
		name := strings.TrimSpace(c.Name)
		if name == "" {
			name = strings.TrimSpace(c.Email)
		}
		if name == "" {
			name = c.ID
		}
		contribByID[c.ID] = name
	}

	type ownerHit struct {
		name  string
		score float64
	}

	ownerScore := make(map[string]float64)
	for _, exp := range expertise {
		domainScore := scoreTextMatch(exp.Domain, terms)
		if domainScore > 0 {
			ownerScore[exp.ContributorID] += exp.Score * float64(domainScore)
		}
	}

	if len(ownerScore) == 0 {
		cmd.Printf("No ownership evidence found for %q in expertise data. Falling back to contribution evidence.\n", queryText)
		return handleSourceEvidenceQuestion(cmd, sourceReader, repositoryID, queryText, "owners (inferred from contribution evidence)")
	}

	hits := make([]ownerHit, 0, len(ownerScore))
	for contributorID, score := range ownerScore {
		name := contribByID[contributorID]
		if name == "" {
			name = contributorID
		}
		hits = append(hits, ownerHit{name: name, score: score})
	}

	sort.Slice(hits, func(i, j int) bool {
		if hits[i].score == hits[j].score {
			return hits[i].name < hits[j].name
		}
		return hits[i].score > hits[j].score
	})

	cmd.Printf("Possible owners for %q based on expertise evidence:\n", queryText)
	for _, hit := range hits {
		cmd.Printf("- %s (aggregated expertise score %.2f)\n", hit.name, hit.score)
	}

	return true, nil
}

func extractQueryTerms(text string) []string {
	lower := strings.ToLower(strings.TrimSpace(text))
	if lower == "" {
		return nil
	}

	words := strings.FieldsFunc(lower, func(r rune) bool {
		return !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9')
	})
	if len(words) == 0 {
		return nil
	}

	seen := make(map[string]struct{})
	terms := make([]string, 0, len(words))
	for _, w := range words {
		if len(w) < 3 {
			continue
		}
		if _, blocked := stopwords[w]; blocked {
			continue
		}
		if _, ok := seen[w]; ok {
			continue
		}
		seen[w] = struct{}{}
		terms = append(terms, w)
	}
	return terms
}

func scoreTextMatch(text string, terms []string) int {
	hay := strings.ToLower(text)
	if hay == "" || len(terms) == 0 {
		return 0
	}

	score := 0
	for _, t := range terms {
		if t == "" {
			continue
		}
		if strings.Contains(hay, t) {
			score++
		}
	}
	return score
}
