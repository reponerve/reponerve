package development

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// PRContextResult is the structured PR evidence pack for CI and team workflows.
type PRContextResult struct {
	Topic              string                  `json:"topic"`
	ChangedFiles       []string                `json:"changed_files"`
	Review             *DevelopmentReviewGuide `json:"review,omitempty"`
	ShipCheck          *ShipCheckResult        `json:"ship_check,omitempty"`
	PRCommentMarkdown  string                  `json:"pr_comment_markdown"`
	Evidence           []EvidenceItem          `json:"evidence"`
	SourceServices     []string                `json:"source_services"`
}

// PRContextRequest is input for PR-scoped evidence assembly.
type PRContextRequest struct {
	RepositoryID string
	Topic        string
	ChangedFiles []string
}

// PreparePRContext assembles review and ship readiness for a pull request diff scope.
func (s *Service) PreparePRContext(ctx context.Context, req PRContextRequest) (*PRContextResult, error) {
	files := normalizeChangedFiles(req.ChangedFiles)
	if len(files) == 0 {
		return nil, fmt.Errorf("at least one changed file is required")
	}

	topic := strings.TrimSpace(req.Topic)
	if topic == "" {
		topic = deriveTopicFromChangedFiles(files)
	}
	if topic == "" {
		topic = "pull request changes"
	}

	devReq := DevelopmentRequest{
		RepositoryID: req.RepositoryID,
		Topic:        topic,
	}

	review, err := s.PrepareReview(ctx, devReq)
	if err != nil {
		return nil, err
	}
	ship, err := s.ShipCheck(ctx, devReq)
	if err != nil {
		return nil, err
	}

	out := &PRContextResult{
		Topic:        topic,
		ChangedFiles: files,
		Review:       review,
		ShipCheck:    ship,
		SourceServices: mergeSourceServices(
			review.SourceServices,
			ship.SourceServices,
			[]string{sourceDevelopmentDiscipline},
		),
	}
	out.Evidence = append(out.Evidence, review.Evidence...)
	out.Evidence = append(out.Evidence, ship.Evidence...)
	sortEvidence(out.Evidence)
	out.PRCommentMarkdown = FormatPRCommentMarkdown(out)
	return out, nil
}

func normalizeChangedFiles(files []string) []string {
	seen := make(map[string]struct{}, len(files))
	out := make([]string, 0, len(files))
	for _, f := range files {
		f = filepath.ToSlash(strings.TrimSpace(f))
		if f == "" || f == "." {
			continue
		}
		if _, ok := seen[f]; ok {
			continue
		}
		seen[f] = struct{}{}
		out = append(out, f)
	}
	sort.Strings(out)
	return out
}

func deriveTopicFromChangedFiles(files []string) string {
	counts := make(map[string]int)
	for _, f := range files {
		parts := strings.Split(filepath.ToSlash(f), "/")
		var key string
		switch {
		case len(parts) >= 2 && parts[0] == "internal":
			key = parts[1]
		case len(parts) >= 2 && parts[0] == "pkg":
			key = parts[1]
		case len(parts) >= 2 && parts[0] == "cmd":
			key = parts[1]
		case len(parts) > 0:
			base := parts[len(parts)-1]
			key = strings.TrimSuffix(base, filepath.Ext(base))
		}
		if key != "" {
			counts[key]++
		}
	}
	bestKey := ""
	bestCount := 0
	for key, count := range counts {
		if count > bestCount || (count == bestCount && key < bestKey) {
			bestKey = key
			bestCount = count
		}
	}
	return bestKey
}

// FormatPRCommentMarkdown renders a bounded GitHub PR comment from PR context.
func FormatPRCommentMarkdown(out *PRContextResult) string {
	if out == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString("## RepoNerve PR Context\n\n")
	fmt.Fprintf(&b, "**Topic:** %s\n\n", out.Topic)

	if len(out.ChangedFiles) > 0 {
		b.WriteString("**Changed files:**\n")
		limit := len(out.ChangedFiles)
		if limit > 15 {
			limit = 15
		}
		for _, f := range out.ChangedFiles[:limit] {
			fmt.Fprintf(&b, "- `%s`\n", f)
		}
		if len(out.ChangedFiles) > limit {
			fmt.Fprintf(&b, "- _…and %d more_\n", len(out.ChangedFiles)-limit)
		}
		b.WriteString("\n")
	}

	if out.Review != nil {
		writeEntitySectionMarkdown(&b, "Recommended reviewers", out.Review.RecommendedReviewers)
		writeEntitySectionMarkdown(&b, "Affected areas", out.Review.AffectedAreas)
		writeEntitySectionMarkdown(&b, "Related knowledge", out.Review.RelatedKnowledge)
		if len(out.Review.DisciplineChecks) > 0 {
			b.WriteString("### Discipline checks\n")
			for _, check := range out.Review.DisciplineChecks {
				fmt.Fprintf(&b, "- [%s] %s\n", check.Category, check.Message)
			}
			b.WriteString("\n")
		}
	}

	if out.ShipCheck != nil {
		if len(out.ShipCheck.ShipBlockers) > 0 {
			b.WriteString("### Ship blockers\n")
			for _, item := range out.ShipCheck.ShipBlockers {
				fmt.Fprintf(&b, "- **%s**: %s\n", item.Category, item.Message)
			}
			b.WriteString("\n")
		}
		if len(out.ShipCheck.Advisories) > 0 {
			b.WriteString("### Advisories\n")
			limit := len(out.ShipCheck.Advisories)
			if limit > 8 {
				limit = 8
			}
			for _, item := range out.ShipCheck.Advisories[:limit] {
				fmt.Fprintf(&b, "- %s\n", item.Message)
			}
			if len(out.ShipCheck.Advisories) > limit {
				fmt.Fprintf(&b, "- _…and %d more advisories_\n", len(out.ShipCheck.Advisories)-limit)
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("---\n_Evidence-backed context from RepoNerve. Run `reponerve pr-context` locally for full JSON._\n")
	return b.String()
}

func writeEntitySectionMarkdown(b *strings.Builder, title string, refs []EntityRef) {
	if len(refs) == 0 {
		return
	}
	fmt.Fprintf(b, "### %s\n", title)
	limit := len(refs)
	if limit > 6 {
		limit = 6
	}
	for _, ref := range refs[:limit] {
		label := ref.Label
		if label == "" {
			label = ref.EntityID
		}
		fmt.Fprintf(b, "- %s\n", label)
	}
	if len(refs) > limit {
		fmt.Fprintf(b, "- _…and %d more_\n", len(refs)-limit)
	}
	b.WriteString("\n")
}

// FormatPRContext renders PR context for CLI prose output.
func FormatPRContext(out *PRContextResult) string {
	if out == nil {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Topic: %s\n", out.Topic)
	if len(out.ChangedFiles) > 0 {
		b.WriteString("Changed files:\n")
		for _, f := range out.ChangedFiles {
			fmt.Fprintf(&b, "  - %s\n", f)
		}
		b.WriteString("\n")
	}
	if out.Review != nil {
		b.WriteString(FormatReviewGuide(out.Review))
	}
	if out.ShipCheck != nil {
		b.WriteString(FormatShipCheck(out.ShipCheck))
	}
	b.WriteString("\n--- PR comment markdown ---\n\n")
	b.WriteString(out.PRCommentMarkdown)
	return strings.TrimRight(b.String(), "\n") + "\n"
}
