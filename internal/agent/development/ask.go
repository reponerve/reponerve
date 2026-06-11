package development

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"

	agentsearch "github.com/reponerve/reponerve/internal/agent/search"
	"github.com/reponerve/reponerve/internal/agent/qa"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
)

const (
	answerTypeOwnership         = "ownership"
	answerTypeAuthorship        = "authorship"
	answerTypeDecisionRationale = "decision_rationale"
	answerTypeDependency        = "dependency"
	answerTypeOverview          = "overview"
	answerTypeSearchSummary     = "search_summary"

	sourceRepositoryQA        = "repository_qa"
	sourceArchitecturalGuidance = "architectural_guidance"
	sourceAgentImpact         = "agent_impact"
)

var (
	rxWhoCreated  = regexp.MustCompile(`(?i)^who created\s+(.+?)\??$`)
	rxWhoWorkedOn = regexp.MustCompile(`(?i)^who worked on\s+(.+?)\??$`)
	rxWhoTouched  = regexp.MustCompile(`(?i)^who touched\s+(.+?)\??$`)
	rxWhoOwns     = regexp.MustCompile(`(?i)^who owns\s+(.+?)\??$`)
	rxWhyUsing    = regexp.MustCompile(`(?i)^why (?:are we |do we )?using\s+(.+?)\??$`)
	rxWhatDepends = regexp.MustCompile(`(?i)^what depends on\s+(.+?)\??$`)
)

var askStopwords = map[string]struct{}{
	"a": {}, "an": {}, "and": {}, "are": {}, "as": {}, "at": {}, "be": {},
	"by": {}, "for": {}, "from": {}, "in": {}, "is": {}, "it": {}, "of": {},
	"on": {}, "or": {}, "that": {}, "the": {}, "this": {}, "to": {}, "was": {},
	"were": {}, "who": {}, "with": {}, "what": {}, "created": {}, "worked": {},
	"touched": {}, "owns": {}, "own": {}, "component": {}, "feature": {}, "server": {},
}

type authorScore struct {
	name  string
	score int
}

// Ask answers repository and development questions using upstream authorities.
func (s *Service) Ask(ctx context.Context, req DevelopmentRequest) (*DevelopmentAnswer, error) {
	question := strings.TrimSpace(req.Topic)
	if question == "" {
		return nil, fmt.Errorf("question cannot be empty")
	}

	if s.qaService != nil {
		if answer, err := s.qaService.Answer(ctx, req.RepositoryID, qa.Question{Text: question}); err == nil {
			return s.fromQAAnswer(ctx, req.RepositoryID, answer)
		}
	}

	if matches := rxWhoOwns.FindStringSubmatch(question); len(matches) > 1 {
		return s.answerOwnership(ctx, req.RepositoryID, question, matches[1])
	}
	if matches := rxWhoCreated.FindStringSubmatch(question); len(matches) > 1 {
		return s.answerAuthorship(ctx, req.RepositoryID, question, matches[1], "creators")
	}
	if matches := rxWhoWorkedOn.FindStringSubmatch(question); len(matches) > 1 {
		return s.answerAuthorship(ctx, req.RepositoryID, question, matches[1], "contributors")
	}
	if matches := rxWhoTouched.FindStringSubmatch(question); len(matches) > 1 {
		return s.answerAuthorship(ctx, req.RepositoryID, question, matches[1], "contributors")
	}
	if matches := rxWhyUsing.FindStringSubmatch(question); len(matches) > 1 {
		return s.answerDecisionRationale(ctx, req.RepositoryID, question, matches[1])
	}
	if matches := rxWhatDepends.FindStringSubmatch(question); len(matches) > 1 {
		return s.answerDependency(ctx, req.RepositoryID, question, matches[1])
	}

	return s.answerSearchSummary(ctx, req.RepositoryID, question)
}

func (s *Service) fromQAAnswer(ctx context.Context, repositoryID string, answer *qa.Answer) (*DevelopmentAnswer, error) {
	out := &DevelopmentAnswer{
		Question:   answer.Question,
		AnswerType: answerTypeSearchSummary,
		SourceServices: []string{sourceRepositoryQA},
	}
	appendEvidence(&out.Evidence, sourceRepositoryQA, "qa_result", answer.Result)

	switch answer.Result.(type) {
	default:
		if strings.Contains(strings.ToLower(answer.Question), "repository") {
			out.AnswerType = answerTypeOverview
			out.SourceServices = []string{sourceRepositoryQA}
		}
	}

	raw, err := json.Marshal(answer.Result)
	if err != nil {
		out.Summary = "Structured answer available from repository Q&A."
	} else {
		out.Summary = string(raw)
	}

	topic, err := s.router.ResolveTopic(ctx, repositoryID, answer.Question)
	if err == nil {
		refs, ev, _ := s.relatedFromTopic(ctx, repositoryID, topic)
		out.Related = refs
		out.Evidence = append(out.Evidence, ev...)
	}
	sortEntityRefs(out.Related)
	sortEvidence(out.Evidence)
	return out, nil
}

func (s *Service) answerOwnership(ctx context.Context, repositoryID, question, subject string) (*DevelopmentAnswer, error) {
	terms := extractAskTerms(subject)
	out := &DevelopmentAnswer{
		Question:       question,
		AnswerType:     answerTypeOwnership,
		SourceServices: []string{sourceOwnershipIntelligence, sourceRepositorySearch},
	}

	topic, _ := s.router.ResolveTopic(ctx, repositoryID, subject)
	if topic != nil {
		refs, ev, _ := s.relatedFromTopic(ctx, repositoryID, topic)
		out.Related = refs
		out.Evidence = append(out.Evidence, ev...)
	}

	expertise, owners, ev, err := s.matchExpertise(ctx, repositoryID, subject)
	if err != nil {
		return nil, err
	}
	out.Related = appendUniqueRefs(out.Related, expertise)
	out.Related = appendUniqueRefs(out.Related, owners)
	out.Evidence = append(out.Evidence, ev...)

	if len(owners) == 0 && s.contributorReader != nil && s.sourceReader != nil {
		auth, err := s.authorshipFromSources(ctx, repositoryID, subject, terms)
		if err != nil {
			return nil, err
		}
		if auth != nil {
			out.Summary = auth.summary
			out.Related = appendUniqueRefs(out.Related, auth.related)
			out.Evidence = append(out.Evidence, auth.evidence...)
		}
	}

	if out.Summary == "" && len(owners) > 0 {
		var lines []string
		lines = append(lines, fmt.Sprintf("Primary owner candidates for %q:", subject))
		for i, o := range owners {
			if i >= 3 {
				break
			}
			lines = append(lines, fmt.Sprintf("  - %s", o.Label))
		}
		out.Summary = strings.Join(lines, "\n")
	}
	if out.Summary == "" {
		out.Summary = fmt.Sprintf("No ownership evidence found for %q.", subject)
	}

	sortEntityRefs(out.Related)
	sortEvidence(out.Evidence)
	return out, nil
}

func (s *Service) answerAuthorship(ctx context.Context, repositoryID, question, subject, label string) (*DevelopmentAnswer, error) {
	terms := extractAskTerms(subject)
	out := &DevelopmentAnswer{
		Question:       question,
		AnswerType:     answerTypeAuthorship,
		SourceServices: []string{sourceOwnershipIntelligence},
	}

	auth, err := s.authorshipFromSources(ctx, repositoryID, subject, terms)
	if err != nil {
		return nil, err
	}
	if auth != nil {
		out.Summary = strings.Replace(auth.summary, "owners", label, 1)
		out.Summary = strings.Replace(out.Summary, "creators", label, 1)
		out.Summary = strings.Replace(out.Summary, "contributors", label, 1)
		out.Related = auth.related
		out.Evidence = auth.evidence
	} else {
		out.Summary = fmt.Sprintf("No evidence found for %q in indexed sources or code history.", subject)
	}

	topic, _ := s.router.ResolveTopic(ctx, repositoryID, subject)
	if topic != nil {
		refs, ev, _ := s.relatedFromTopic(ctx, repositoryID, topic)
		out.Related = appendUniqueRefs(out.Related, refs)
		out.Evidence = append(out.Evidence, ev...)
	}

	sortEntityRefs(out.Related)
	sortEvidence(out.Evidence)
	return out, nil
}

type authorshipResult struct {
	summary string
	related []EntityRef
	evidence []EvidenceItem
}

func (s *Service) authorshipFromSources(ctx context.Context, repositoryID, subject string, terms []string) (*authorshipResult, error) {
	if s.sourceReader == nil || len(terms) == 0 {
		return nil, nil
	}

	sources, err := s.sourceReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("list sources: %w", err)
	}

	hitMap := make(map[string]int)
	for _, src := range sources {
		evidenceText := src.Title + " " + src.Reference + " " + src.MetadataJSON
		score := scoreAskTextMatch(evidenceText, terms)
		if score > 0 {
			author := strings.TrimSpace(src.Author)
			if author == "" {
				author = "unknown"
			}
			hitMap[author] += score
		}
	}

	if len(hitMap) == 0 && s.repositoryPath != "" {
		codeHits, err := findContributorsFromGit(s.repositoryPath, terms)
		if err != nil {
			return nil, err
		}
		if len(codeHits) > 0 {
			return formatAuthorScores(subject, "contributors", codeHits, sourceOwnershipIntelligence)
		}
		return nil, nil
	}

	if len(hitMap) == 0 {
		return nil, nil
	}

	hits := mapToAuthorScores(hitMap)
	return formatAuthorScores(subject, "contributors", hits, sourceOwnershipIntelligence)
}

func formatAuthorScores(subject, label string, hits []authorScore, source string) (*authorshipResult, error) {
	var lines []string
	lines = append(lines, fmt.Sprintf("Possible %s for %q based on evidence:", label, subject))
	var related []EntityRef
	var evidence []EvidenceItem
	for _, hit := range hits {
		lines = append(lines, fmt.Sprintf("  - %s (evidence score %d)", hit.name, hit.score))
		related = append(related, EntityRef{
			EntityType: agentsearch.EntityTypeContributor,
			EntityID:   hit.name,
			Label:      hit.name,
		})
		appendEvidence(&evidence, source, "contributor_activity", map[string]any{
			"email": hit.name, "score": hit.score,
		})
	}
	return &authorshipResult{
		summary:  strings.Join(lines, "\n"),
		related:  related,
		evidence: evidence,
	}, nil
}

func (s *Service) answerDecisionRationale(ctx context.Context, repositoryID, question, subject string) (*DevelopmentAnswer, error) {
	out := &DevelopmentAnswer{
		Question:       question,
		AnswerType:     answerTypeDecisionRationale,
		SourceServices: []string{sourceRepositorySearch, sourceArchitecturalGuidance},
	}

	topic, err := s.router.ResolveTopic(ctx, repositoryID, subject)
	if err != nil {
		return nil, err
	}
	refs, ev, _ := s.relatedFromTopic(ctx, repositoryID, topic)
	out.Related = refs
	out.Evidence = append(out.Evidence, ev...)

	var decisionLabels []string
	for _, ref := range refs {
		if ref.EntityType == agentsearch.EntityTypeDecision {
			decisionLabels = append(decisionLabels, ref.Label)
		}
	}
	if len(decisionLabels) > 0 {
		out.Summary = fmt.Sprintf("Relevant decisions for %q:\n  - %s", subject, strings.Join(decisionLabels, "\n  - "))
	} else {
		out.Summary = fmt.Sprintf("No decision evidence found for %q. Run `reponerve scan` to refresh repository memory.", subject)
	}

	sortEntityRefs(out.Related)
	sortEvidence(out.Evidence)
	return out, nil
}

func (s *Service) answerDependency(ctx context.Context, repositoryID, question, subject string) (*DevelopmentAnswer, error) {
	out := &DevelopmentAnswer{
		Question:       question,
		AnswerType:     answerTypeDependency,
		SourceServices: []string{sourceRepositorySearch, sourceCodeIntelligence, sourceRepositoryCodeLinks},
	}

	topic, err := s.router.ResolveTopic(ctx, repositoryID, subject)
	if err != nil {
		return nil, err
	}
	refs, ev, links := s.relatedFromTopic(ctx, repositoryID, topic)
	out.Related = refs
	out.Evidence = append(out.Evidence, ev...)

	for _, link := range links {
		appendEvidence(&out.Evidence, sourceRepositoryCodeLinks, "link", json.RawMessage(link.EvidenceJSON))
	}

	if len(refs) == 0 {
		out.Summary = fmt.Sprintf("No dependency evidence found for %q.", subject)
	} else {
		out.Summary = fmt.Sprintf("Entities related to %q (%d matches).", subject, len(refs))
	}

	sortEntityRefs(out.Related)
	sortEvidence(out.Evidence)
	return out, nil
}

func (s *Service) answerSearchSummary(ctx context.Context, repositoryID, question string) (*DevelopmentAnswer, error) {
	out := &DevelopmentAnswer{
		Question:       question,
		AnswerType:     answerTypeSearchSummary,
		SourceServices: []string{sourceRepositorySearch},
	}

	topic, err := s.router.ResolveTopic(ctx, repositoryID, question)
	if err != nil {
		return nil, err
	}
	refs, ev, _ := s.relatedFromTopic(ctx, repositoryID, topic)
	out.Related = refs
	out.Evidence = append(out.Evidence, ev...)

	if len(refs) == 0 {
		out.Summary = "No deterministic answer pattern matched this question."
	} else {
		out.Summary = fmt.Sprintf("Search found %d related entities for %q.", len(refs), question)
	}

	sortEntityRefs(out.Related)
	sortEvidence(out.Evidence)
	return out, nil
}

func (s *Service) relatedFromTopic(ctx context.Context, repositoryID string, topic *ResolvedTopic) ([]EntityRef, []EvidenceItem, []*codemodels.RepositoryCodeRelationship) {
	if topic == nil {
		return nil, nil, nil
	}
	var refs []EntityRef
	var evidence []EvidenceItem
	for entityID := range topic.RepositoryHitIDs {
		ref, ev, _, err := s.resolveRepositoryEntity(ctx, repositoryID, entityID)
		if err != nil || ref == nil {
			continue
		}
		refs = append(refs, *ref)
		evidence = append(evidence, ev...)
	}
	for entityID := range topic.CodeEntityIDs {
		e, err := s.codeEntityReader.GetByID(ctx, entityID)
		if err != nil {
			continue
		}
		refs = append(refs, codeEntityRef(e))
		appendEvidence(&evidence, sourceCodeIntelligence, "code_entity", map[string]string{
			"id": e.ID, "qualified_name": e.QualifiedName,
		})
	}
	return refs, evidence, topic.RepositoryCodeLinks
}

func extractAskTerms(text string) []string {
	lower := strings.ToLower(strings.TrimSpace(text))
	words := strings.FieldsFunc(lower, func(r rune) bool {
		return !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9')
	})
	seen := make(map[string]struct{})
	var terms []string
	for _, w := range words {
		if len(w) < 3 {
			continue
		}
		if _, blocked := askStopwords[w]; blocked {
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

func scoreAskTextMatch(text string, terms []string) int {
	hay := strings.ToLower(text)
	score := 0
	for _, t := range terms {
		if strings.Contains(hay, t) {
			score++
		}
	}
	return score
}

func mapToAuthorScores(hitMap map[string]int) []authorScore {
	hits := make([]authorScore, 0, len(hitMap))
	for author, count := range hitMap {
		hits = append(hits, authorScore{name: author, score: count})
	}
	sort.Slice(hits, func(i, j int) bool {
		if hits[i].score == hits[j].score {
			return hits[i].name < hits[j].name
		}
		return hits[i].score > hits[j].score
	})
	return hits
}

func findContributorsFromGit(repoPath string, terms []string) ([]authorScore, error) {
	patternParts := make([]string, 0, len(terms))
	for _, t := range terms {
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

	authorCount := make(map[string]int)
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 3)
		if len(parts) < 3 {
			continue
		}
		lineNum, convErr := strconv.Atoi(parts[1])
		if convErr != nil {
			continue
		}
		if scoreAskTextMatch(parts[0]+" "+parts[2], terms) == 0 {
			continue
		}
		author := gitBlameAuthor(repoPath, parts[0], lineNum)
		if author == "" {
			author = "unknown"
		}
		authorCount[author]++
	}
	return mapToAuthorScores(authorCount), nil
}

func gitBlameAuthor(repoPath, file string, lineNum int) string {
	blameCmd := exec.Command(
		"git", "-C", repoPath, "blame", "--line-porcelain",
		"-L", fmt.Sprintf("%d,%d", lineNum, lineNum), "--", file,
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
