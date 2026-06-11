package development

import (
	"fmt"
	"strings"
)

// FormatExplanation renders a DevelopmentExplanation for CLI output.
func FormatExplanation(out *DevelopmentExplanation) string {
	if out == nil {
		return ""
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Topic: %s\n\n", out.Topic)

	if out.CodeContext != nil && hasCodeContent(out.CodeContext) {
		b.WriteString("CODE CONTEXT\n")
		writeEntitySection(&b, "Modules", out.CodeContext.Modules)
		writeEntitySection(&b, "Files", out.CodeContext.Files)
		writeEntitySection(&b, "Packages", out.CodeContext.Packages)
		writeEntitySection(&b, "Structs", out.CodeContext.Structs)
		writeEntitySection(&b, "Interfaces", out.CodeContext.Interfaces)
		writeEntitySection(&b, "Type Aliases", out.CodeContext.TypeAliases)
		writeEntitySection(&b, "Functions", out.CodeContext.Functions)
		writeEntitySection(&b, "Methods", out.CodeContext.Methods)
		writeEntitySection(&b, "Endpoints", out.CodeContext.Endpoints)
		if out.CodeContext.CallGraph != nil && len(out.CodeContext.CallGraph.Edges) > 0 {
			b.WriteString("Call Graph:\n")
			for _, edge := range out.CodeContext.CallGraph.Edges {
				fmt.Fprintf(&b, "  - %s -> %s\n", edge.FromEntityID, edge.ToEntityID)
			}
		}
		if len(out.CodeContext.Dependencies) > 0 {
			b.WriteString("Dependencies:\n")
			for _, dep := range out.CodeContext.Dependencies {
				fmt.Fprintf(&b, "  - %s %s -> %s\n", dep.RelationshipType, dep.FromEntityID, dep.ToEntityID)
			}
		}
		b.WriteString("\n")
	}

	if out.RepositoryContext != nil {
		b.WriteString("REPOSITORY CONTEXT\n")
		writeEntitySection(&b, "Decisions", out.RepositoryContext.Decisions)
		writeEntitySection(&b, "Facts", out.RepositoryContext.Facts)
		writeEntitySection(&b, "Events", out.RepositoryContext.Events)
		writeEntitySection(&b, "Owners", out.RepositoryContext.Owners)
		writeEntitySection(&b, "Expertise", out.RepositoryContext.Expertise)
		b.WriteString("\n")
	}

	if len(out.RepositoryCodeLinks) > 0 {
		b.WriteString("REPOSITORY-CODE LINKS\n")
		for _, link := range out.RepositoryCodeLinks {
			fmt.Fprintf(&b, "  - %s: %s (%s) -> %s (%s)\n",
				link.RelationshipType,
				link.RepositoryEntityRef.EntityType, link.RepositoryEntityRef.Label,
				link.CodeEntityRef.EntityType, link.CodeEntityRef.Label,
			)
		}
		b.WriteString("\n")
	}

	if len(out.Evidence) > 0 {
		b.WriteString("Evidence:\n")
		for _, ev := range out.Evidence {
			fmt.Fprintf(&b, "  - source: %s type: %s\n", ev.Source, ev.Type)
		}
		b.WriteString("\n")
	}

	if len(out.SourceServices) > 0 {
		b.WriteString("Source Services:\n")
		for _, svc := range out.SourceServices {
			fmt.Fprintf(&b, "  - %s\n", svc)
		}
	}

	return strings.TrimRight(b.String(), "\n") + "\n"
}

// FormatAnswer renders a DevelopmentAnswer for CLI output.
func FormatAnswer(out *DevelopmentAnswer) string {
	if out == nil {
		return ""
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Question: %s\n\n", out.Question)
	fmt.Fprintf(&b, "Answer Type: %s\n\n", out.AnswerType)

	if out.Summary != "" {
		b.WriteString("Summary:\n")
		for _, line := range strings.Split(strings.TrimSpace(out.Summary), "\n") {
			fmt.Fprintf(&b, "  %s\n", line)
		}
		b.WriteString("\n")
	}

	if len(out.Related) > 0 {
		b.WriteString("Related Entities:\n")
		for _, ref := range out.Related {
			fmt.Fprintf(&b, "  - %s / %s / %s\n", ref.EntityType, ref.EntityID, ref.Label)
		}
		b.WriteString("\n")
	}

	if len(out.Evidence) > 0 {
		b.WriteString("Evidence:\n")
		for _, ev := range out.Evidence {
			fmt.Fprintf(&b, "  - source: %s\n    type: %s\n", ev.Source, ev.Type)
		}
		b.WriteString("\n")
	}

	if len(out.SourceServices) > 0 {
		b.WriteString("Source Services:\n")
		for _, svc := range out.SourceServices {
			fmt.Fprintf(&b, "  - %s\n", svc)
		}
	}

	if out.AnswerType == answerTypeSearchSummary && len(out.Related) == 0 &&
		strings.Contains(out.Summary, "No deterministic answer pattern matched") {
		b.WriteString("\nTry one of these formats:\n")
		b.WriteString(`  reponerve ask "What is this repository?"` + "\n")
		b.WriteString(`  reponerve ask "Why was decision <decision_id> made?"` + "\n")
		b.WriteString(`  reponerve ask "Who owns <domain or component>?"` + "\n")
		b.WriteString(`  reponerve ask "Why are we using Redis?"` + "\n")
	}

	return strings.TrimRight(b.String(), "\n") + "\n"
}

func writeEntitySection(b *strings.Builder, title string, refs []EntityRef) {
	if len(refs) == 0 {
		return
	}
	fmt.Fprintf(b, "%s:\n", title)
	for _, ref := range refs {
		fmt.Fprintf(b, "  - %s / %s / %s\n", ref.EntityType, ref.EntityID, ref.Label)
	}
}
