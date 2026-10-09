package development

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/reponerve/reponerve/internal/agent/qa"
	agentsearch "github.com/reponerve/reponerve/internal/agent/search"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
)

const (
	answerTypeOwnership          = "ownership"
	answerTypeAuthorship         = "authorship"
	answerTypeDecisionRationale  = "decision_rationale"
	answerTypeDependency         = "dependency"
	answerTypeOverview           = "overview"
	answerTypeConceptExplanation = "concept_explanation"
	answerTypeTaskPlan           = "task_plan"
	answerTypeSearchSummary      = "search_summary"

	sourceRepositoryQA          = "repository_qa"
	sourceArchitecturalGuidance = "architectural_guidance"
	sourceAgentImpact           = "agent_impact"
	sourceRepositoryOverview    = "repository_overview"
)

var (
	rxWhoCreated  = regexp.MustCompile(`(?i)^who created\s+(.+?)\??$`)
	rxWhoWorkedOn = regexp.MustCompile(`(?i)^who worked on\s+(.+?)\??$`)
	rxWhoTouched  = regexp.MustCompile(`(?i)^who touched\s+(.+?)\??$`)
	rxWhoOwns     = regexp.MustCompile(`(?i)^who owns\s+(.+?)\??$`)
	rxWhyUse      = regexp.MustCompile(`(?i)^why (?:are we |do we )?(?:use|using|uses|used)\s+(.+?)\??$`)
	rxWhatDepends = regexp.MustCompile(`(?i)^what depends on\s+(.+?)\??$`)
	rxWhatIs      = regexp.MustCompile(`(?i)^what is (?:the )?(?:struct |type |function )?(.+?)\??$`)
	rxWhatDoes    = regexp.MustCompile(`(?i)^what does (?:the )?(.+?) do\??$`)
	rxHowDoesWork = regexp.MustCompile(`(?i)^how does (?:the )?(.+?) work\??$`)
	rxTellMeAbout = regexp.MustCompile(`(?i)^tell me about (?:the )?(.+?)\??$`)

	rxRepoOverviewQuestion = regexp.MustCompile(`(?i)^(?:what (?:does|is)|tell me about|explain|describe)\s+(?:the\s+|this\s+)?(?:repo|repository|project|codebase|application|app)(?:\s+do|\s+about)?\??$`)
	rxOverviewDirect       = regexp.MustCompile(`(?i)^(?:project|repository|codebase)\s+overview\??$`)
	rxPurposeQuestion      = regexp.MustCompile(`(?i)^what is the purpose of (?:the\s+|this\s+)?(?:repo|repository|project|codebase)\??$`)
)

func isRepositoryOverviewSubject(subject string) bool {
	s := strings.Trim(strings.ToLower(strings.TrimSpace(subject)), "?")
	switch s {
	case "repo", "repository", "this repo", "this repository", "the repo", "the repository",
		"project", "this project", "the project",
		"codebase", "this codebase", "the codebase",
		"app", "this app", "the app", "application", "this application", "the application":
		return true
	}
	return false
}

func (s *Service) isRepoNameSubject(subj string) bool {
	if s.repositoryPath == "" {
		return false
	}
	base := filepath.Base(s.repositoryPath)
	return strings.EqualFold(subj, base)
}

func isRepositoryOverviewQuestion(q string) bool {
	q = strings.TrimSpace(q)
	return rxRepoOverviewQuestion.MatchString(q) || rxOverviewDirect.MatchString(q) || rxPurposeQuestion.MatchString(q)
}

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

	if LooksLikeTaskDescription(question) {
		return s.answerTaskPlan(ctx, req.RepositoryID, question)
	}

	if isRepositoryOverviewQuestion(question) {
		return s.answerRepositoryOverview(ctx, req.RepositoryID, question)
	}

	if matches := rxWhatIs.FindStringSubmatch(question); len(matches) > 1 {
		subj := strings.TrimSpace(matches[1])
		if isRepositoryOverviewSubject(subj) || s.isRepoNameSubject(subj) {
			return s.answerRepositoryOverview(ctx, req.RepositoryID, question)
		}
		return s.answerConcept(ctx, req.RepositoryID, question, subj)
	}
	if matches := rxWhatDoes.FindStringSubmatch(question); len(matches) > 1 {
		subj := strings.TrimSpace(matches[1])
		if isRepositoryOverviewSubject(subj) || s.isRepoNameSubject(subj) {
			return s.answerRepositoryOverview(ctx, req.RepositoryID, question)
		}
		return s.answerConcept(ctx, req.RepositoryID, question, subj)
	}
	if matches := rxHowDoesWork.FindStringSubmatch(question); len(matches) > 1 {
		subj := strings.TrimSpace(matches[1])
		if isRepositoryOverviewSubject(subj) || s.isRepoNameSubject(subj) {
			return s.answerRepositoryOverview(ctx, req.RepositoryID, question)
		}
		return s.answerConcept(ctx, req.RepositoryID, question, subj)
	}
	if matches := rxTellMeAbout.FindStringSubmatch(question); len(matches) > 1 {
		subj := strings.TrimSpace(matches[1])
		if isRepositoryOverviewSubject(subj) || s.isRepoNameSubject(subj) {
			return s.answerRepositoryOverview(ctx, req.RepositoryID, question)
		}
		return s.answerConcept(ctx, req.RepositoryID, question, subj)
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
	if matches := rxWhyUse.FindStringSubmatch(question); len(matches) > 1 {
		return s.answerDecisionRationale(ctx, req.RepositoryID, question, matches[1])
	}
	if matches := rxWhatDepends.FindStringSubmatch(question); len(matches) > 1 {
		return s.answerDependency(ctx, req.RepositoryID, question, matches[1])
	}

	if s.qaService != nil {
		if answer, err := s.qaService.Answer(ctx, req.RepositoryID, qa.Question{Text: question}); err == nil {
			return s.fromQAAnswer(ctx, req.RepositoryID, answer)
		}
	}

	return s.answerSearchSummary(ctx, req.RepositoryID, question)
}

func (s *Service) answerConcept(ctx context.Context, repositoryID, question, subject string) (*DevelopmentAnswer, error) {
	subject = strings.TrimSpace(subject)
	if subject == "" {
		return s.answerSearchSummary(ctx, repositoryID, question)
	}

	expl, err := s.Explain(ctx, DevelopmentRequest{
		RepositoryID: repositoryID,
		Topic:        subject,
	})
	if err != nil {
		return s.answerSearchSummary(ctx, repositoryID, question)
	}

	out := conceptAnswerFromExplanation(question, expl)
	sortEntityRefs(out.Related)
	sortEvidence(out.Evidence)
	return out, nil
}

func conceptAnswerFromExplanation(question string, expl *DevelopmentExplanation) *DevelopmentAnswer {
	out := &DevelopmentAnswer{
		Question:        question,
		AnswerType:      answerTypeConceptExplanation,
		EntityBriefings: expl.EntityBriefings,
		Evidence:        expl.Evidence,
		SourceServices:  expl.SourceServices,
		Related:         relatedFromExplanation(expl),
	}

	var summaryParts []string
	if briefSummary := summarizeBriefings(expl.EntityBriefings); briefSummary != "" {
		summaryParts = append(summaryParts, briefSummary)
	}
	if expl.RepositoryContext != nil {
		if len(expl.RepositoryContext.Decisions) > 0 {
			labels := make([]string, 0, len(expl.RepositoryContext.Decisions))
			for _, d := range expl.RepositoryContext.Decisions {
				labels = append(labels, d.Label)
			}
			summaryParts = append(summaryParts, fmt.Sprintf("Related decisions:\n  - %s", strings.Join(labels, "\n  - ")))
		}
	}
	if len(summaryParts) == 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("Concept briefing for %q assembled from code and repository intelligence.", expl.Topic))
	}
	out.Summary = strings.Join(summaryParts, "\n\n")
	return out
}

func relatedFromExplanation(expl *DevelopmentExplanation) []EntityRef {
	if expl == nil {
		return nil
	}
	seen := make(map[string]struct{})
	var refs []EntityRef
	appendUnique := func(list []EntityRef) {
		for _, ref := range list {
			key := ref.EntityType + ":" + ref.EntityID
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			refs = append(refs, ref)
		}
	}
	if expl.CodeContext != nil {
		appendUnique(expl.CodeContext.Structs)
		appendUnique(expl.CodeContext.Interfaces)
		appendUnique(expl.CodeContext.TypeAliases)
		appendUnique(expl.CodeContext.Functions)
		appendUnique(expl.CodeContext.Files)
	}
	if expl.RepositoryContext != nil {
		appendUnique(expl.RepositoryContext.Decisions)
		appendUnique(expl.RepositoryContext.Facts)
		appendUnique(expl.RepositoryContext.Events)
	}
	return refs
}

func (s *Service) answerRepositoryOverview(ctx context.Context, repositoryID, question string) (*DevelopmentAnswer, error) {
	out := &DevelopmentAnswer{
		Question:       question,
		AnswerType:     answerTypeOverview,
		SourceServices: []string{sourceArchitecturalGuidance, sourceRepositoryOverview},
	}

	var overviewDoc string
	var docRef string
	var docTitle string

	// 1. Try finding README or vision source in indexed memory
	if s.sourceReader != nil {
		sources, err := s.sourceReader.ListByRepository(ctx, repositoryID)
		if err == nil {
			for _, src := range sources {
				refLower := strings.ToLower(src.Reference)
				if strings.HasSuffix(refLower, "readme.md") || refLower == "readme.md" {
					if content := extractContentFromMetadata(src.MetadataJSON); content != "" {
						overviewDoc = content
						docRef = src.Reference
						docTitle = src.Title
						out.Related = append(out.Related, EntityRef{
							EntityType: "DOCUMENT",
							EntityID:   src.ID,
							Label:      src.Reference,
						})
						break
					}
				}
			}
			if overviewDoc == "" {
				for _, src := range sources {
					refLower := strings.ToLower(src.Reference)
					if strings.Contains(refLower, "vision.md") || src.SourceType == "architecture_doc" {
						if content := extractContentFromMetadata(src.MetadataJSON); content != "" {
							overviewDoc = content
							docRef = src.Reference
							docTitle = src.Title
							out.Related = append(out.Related, EntityRef{
								EntityType: "DOCUMENT",
								EntityID:   src.ID,
								Label:      src.Reference,
							})
							break
						}
					}
				}
			}
		}
	}

	// 2. Fallback to reading README.md from disk if not found in memory
	if overviewDoc == "" && s.repositoryPath != "" {
		readmeCandidates := []string{"README.md", "readme.md", "README", "docs/vision/vision.md"}
		for _, name := range readmeCandidates {
			path := filepath.Join(s.repositoryPath, name)
			if data, err := os.ReadFile(path); err == nil && len(data) > 0 {
				overviewDoc = string(data)
				docRef = name
				docTitle = parseMarkdownTitle(overviewDoc)
				break
			}
		}
	}

	if overviewDoc != "" {
		appendEvidence(&out.Evidence, sourceArchitecturalGuidance, "repository_overview", map[string]string{
			"reference": docRef,
			"title":     docTitle,
		})
	}

	// 3. Extract summary from markdown
	summaryText := extractOverviewSummary(overviewDoc, docTitle)

	// 4. Discover key capabilities / features
	var capabilities []string
	if s.featureService != nil {
		if list, err := s.featureService.ListFeatures(ctx, repositoryID); err == nil && list != nil {
			for i, f := range list.Features {
				if i >= 5 {
					break
				}
				if len(f.Keywords) > 0 {
					capabilities = append(capabilities, fmt.Sprintf("%s (%s)", f.Name, strings.Join(f.Keywords, ", ")))
				} else {
					capabilities = append(capabilities, f.Name)
				}
				out.Related = append(out.Related, EntityRef{
					EntityType: "FEATURE",
					EntityID:   f.ID,
					Label:      f.Name,
				})
			}
		}
	}

	var sb strings.Builder
	if summaryText != "" {
		sb.WriteString(summaryText)
	} else {
		repoName := filepath.Base(s.repositoryPath)
		if repoName == "" || repoName == "." {
			repoName = "Repository"
		}
		sb.WriteString(fmt.Sprintf("%s: software repository overview.", repoName))
	}

	if len(capabilities) > 0 {
		sb.WriteString("\n\nPrimary capabilities and components:\n")
		for _, cap := range capabilities {
			sb.WriteString(fmt.Sprintf("  - %s\n", cap))
		}
	}

	out.Summary = strings.TrimSpace(sb.String())
	sortEntityRefs(out.Related)
	sortEvidence(out.Evidence)
	return out, nil
}

func extractOverviewSummary(markdown, title string) string {
	if markdown == "" {
		return ""
	}
	lines := strings.Split(markdown, "\n")
	var paragraphs []string
	var current strings.Builder
	headerCount := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") {
			if title == "" {
				title = strings.TrimSpace(strings.TrimPrefix(trimmed, "# "))
			}
			continue
		}
		// Stop at major secondary sections like Installation, Getting Started, Contributing, License
		if strings.HasPrefix(trimmed, "## ") {
			secTitle := strings.ToLower(trimmed)
			if strings.Contains(secTitle, "install") || strings.Contains(secTitle, "getting started") ||
				strings.Contains(secTitle, "usage") || strings.Contains(secTitle, "contributing") ||
				strings.Contains(secTitle, "license") || strings.Contains(secTitle, "table of contents") {
				break
			}
			headerCount++
			if headerCount > 2 {
				break
			}
			continue
		}
		if strings.HasPrefix(trimmed, "---") || strings.HasPrefix(trimmed, "***") {
			if len(paragraphs) > 0 {
				break
			}
			continue
		}
		// Skip badges and images
		if strings.HasPrefix(trimmed, "[![") || strings.HasPrefix(trimmed, "![") || strings.HasPrefix(trimmed, "<!--") {
			continue
		}
		if trimmed == "" {
			if current.Len() > 0 {
				paragraphs = append(paragraphs, current.String())
				current.Reset()
				if len(paragraphs) >= 3 {
					break
				}
			}
			continue
		}
		if current.Len() > 0 {
			current.WriteString(" ")
		}
		current.WriteString(trimmed)
	}
	if current.Len() > 0 && len(paragraphs) < 3 {
		paragraphs = append(paragraphs, current.String())
	}

	var out strings.Builder
	if title != "" {
		out.WriteString(fmt.Sprintf("# %s\n\n", title))
	}
	out.WriteString(strings.Join(paragraphs, "\n\n"))
	return strings.TrimSpace(out.String())
}

func parseMarkdownTitle(content string) string {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "# "))
		}
	}
	return ""
}

func extractContentFromMetadata(metadataJSON string) string {
	if metadataJSON == "" {
		return ""
	}
	var m struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal([]byte(metadataJSON), &m); err == nil {
		return m.Content
	}
	return ""
}

func (s *Service) fromQAAnswer(ctx context.Context, repositoryID string, answer *qa.Answer) (*DevelopmentAnswer, error) {
	out := &DevelopmentAnswer{
		Question:       answer.Question,
		AnswerType:     answerTypeSearchSummary,
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
	summary  string
	related  []EntityRef
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
	out.Evidence = append(out.Evidence, ev...)

	decisions, err := s.decisionReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	sources, err := s.sourceReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, err
	}
	sourceByID := make(map[string]string, len(sources))
	for _, src := range sources {
		sourceByID[src.ID] = src.MetadataJSON
	}

	type scoredDecision struct {
		decision *memorymodels.Decision
		score    int
	}
	var scoredList []scoredDecision
	for _, d := range decisions {
		score := scoreDecisionRelevance(d.Title, subject, question)
		if score > 0 {
			scoredList = append(scoredList, scoredDecision{decision: d, score: score})
		}
	}

	sort.SliceStable(scoredList, func(i, j int) bool {
		if scoredList[i].score != scoredList[j].score {
			return scoredList[i].score > scoredList[j].score
		}
		return scoredList[i].decision.ID < scoredList[j].decision.ID
	})

	var rationaleLines []string
	var decisionRefs []EntityRef
	for _, sd := range scoredList {
		d := sd.decision
		decisionRefs = append(decisionRefs, EntityRef{
			EntityType: agentsearch.EntityTypeDecision,
			EntityID:   d.ID,
			Label:      d.Title,
		})
		snippet := adrRationaleSnippet(sourceByID[d.SourceID])
		if snippet == "" {
			rationaleLines = append(rationaleLines, d.Title)
		} else {
			rationaleLines = append(rationaleLines, fmt.Sprintf("%s — %s", d.Title, snippet))
		}
		appendEvidence(&out.Evidence, sourceArchitecturalGuidance, "decision_rationale", map[string]string{
			"decision_id": d.ID,
			"title":       d.Title,
			"snippet":     snippet,
		})
	}

	// Append non-decision related entities from topic
	for _, ref := range refs {
		if ref.EntityType != agentsearch.EntityTypeDecision {
			decisionRefs = append(decisionRefs, ref)
		}
	}
	out.Related = decisionRefs

	if len(rationaleLines) > 0 {
		out.Summary = fmt.Sprintf("Relevant decisions for %q:\n  - %s", subject, strings.Join(rationaleLines, "\n  - "))
	} else {
		out.Summary = fmt.Sprintf("No decision evidence found for %q. Run `reponerve scan` to refresh repository memory.", subject)
	}

	sortEvidence(out.Evidence)
	return out, nil
}

func scoreDecisionRelevance(title, subject, question string) int {
	titleLower := strings.ToLower(title)
	subjLower := strings.ToLower(strings.TrimSpace(subject))
	score := 0

	// Foundational adoption prefix: e.g. "Use SQLite as ..." or "Use SQLite for ..."
	if strings.HasPrefix(titleLower, "use "+subjLower+" as") || strings.HasPrefix(titleLower, "adopt "+subjLower+" as") {
		score += 300
	} else if strings.HasPrefix(titleLower, "use "+subjLower) || strings.HasPrefix(titleLower, "adopt "+subjLower) {
		score += 200
	}

	if strings.Contains(titleLower, subjLower+" as ") {
		score += 150
	}

	if containsWholeWord(titleLower, subjLower) {
		score += 100
	} else if strings.Contains(titleLower, subjLower) {
		score += 40
	}

	// Question keywords
	qWords := strings.Fields(strings.ToLower(question))
	for _, w := range qWords {
		if len(w) > 3 && containsWholeWord(titleLower, w) {
			score += 20
		}
	}

	// Specialized sub-feature penalty when asking general question about technology
	if subjLower == "sqlite" && (strings.Contains(titleLower, "fts5") || strings.Contains(titleLower, "migration")) {
		score -= 40
	}

	return score
}

func containsWholeWord(text, word string) bool {
	words := strings.FieldsFunc(text, func(r rune) bool {
		return !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9')
	})
	for _, w := range words {
		if w == word {
			return true
		}
	}
	return false
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
	out.Related = prioritizeAndCapRelated(refs, 20)
	out.Evidence = append(out.Evidence, ev...)

	if len(refs) == 0 {
		out.Summary = fmt.Sprintf("No matches found for %q in this repository.", question)
	} else if len(out.Related) < len(refs) {
		out.Summary = fmt.Sprintf("Search found %d related entities for %q (showing top %d).", len(refs), question, len(out.Related))
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

	var repoHitIDs []string
	for entityID := range topic.RepositoryHitIDs {
		repoHitIDs = append(repoHitIDs, entityID)
	}
	sort.Strings(repoHitIDs)

	for _, entityID := range repoHitIDs {
		ref, ev, _, err := s.resolveRepositoryEntity(ctx, repositoryID, entityID)
		if err != nil || ref == nil {
			continue
		}
		refs = append(refs, *ref)
		evidence = append(evidence, ev...)
	}

	var codeHitIDs []string
	for entityID := range topic.CodeEntityIDs {
		codeHitIDs = append(codeHitIDs, entityID)
	}
	sort.Strings(codeHitIDs)

	for _, entityID := range codeHitIDs {
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
