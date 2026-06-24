package development

import (
	"fmt"
	"strings"

	"github.com/reponerve/reponerve/internal/intelligence/feature"
)

// FormatExplanation renders a DevelopmentExplanation for CLI output.
func FormatExplanation(out *DevelopmentExplanation) string {
	if out == nil {
		return ""
	}

	var b strings.Builder
	if out.Feature != nil {
		fmt.Fprintf(&b, "Feature: %s\n", out.Feature.Name)
		if len(out.Feature.Keywords) > 0 {
			fmt.Fprintf(&b, "  Keywords: %s\n", strings.Join(out.Feature.Keywords, ", "))
		}
		if len(out.Feature.Sources) > 0 {
			fmt.Fprintf(&b, "  Sources: %s\n", strings.Join(out.Feature.Sources, ", "))
		}
		b.WriteString("\n")
	}
	fmt.Fprintf(&b, "Topic: %s\n\n", out.Topic)

	if len(out.EntityBriefings) > 0 {
		b.WriteString("ENTITY BRIEFINGS\n")
		for _, brief := range out.EntityBriefings {
			fmt.Fprintf(&b, "  %s [%s]\n", brief.QualifiedName, brief.EntityType)
			fmt.Fprintf(&b, "    Layer: %s\n", brief.Layer)
			fmt.Fprintf(&b, "    Role: %s\n", brief.Role)
			if brief.DefinedIn != "" {
				fmt.Fprintf(&b, "    Defined in: %s\n", brief.DefinedIn)
			}
			if len(brief.Fields) > 0 {
				fmt.Fprintf(&b, "    Fields: %s\n", strings.Join(brief.Fields, "; "))
			} else if brief.Signature != "" {
				fmt.Fprintf(&b, "    Signature: %s\n", brief.Signature)
			}
			writeEntitySectionIndented(&b, "    ", "Members", brief.Members)
			writeEntitySectionIndented(&b, "    ", "Called by", brief.Producers)
			writeEntitySectionIndented(&b, "    ", "Calls/uses", brief.Consumers)
			writeEntitySectionIndented(&b, "    ", "Related decisions", brief.RelatedDecisions)
		}
		b.WriteString("\n")
	}

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
				label := dep.Label
				if label == "" {
					label = fmt.Sprintf("%s -> %s", dep.FromEntityID, dep.ToEntityID)
				}
				fmt.Fprintf(&b, "  - %s\n", label)
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

	if out.Plan != nil && len(out.Plan.SuggestedSteps) > 0 {
		b.WriteString("Suggested Steps:\n")
		for _, step := range out.Plan.SuggestedSteps {
			fmt.Fprintf(&b, "  %s\n", step)
		}
		b.WriteString("\n")
	}

	if len(out.EntityBriefings) > 0 {
		b.WriteString("Entity Briefings:\n")
		for _, brief := range out.EntityBriefings {
			fmt.Fprintf(&b, "  %s [%s]\n", brief.QualifiedName, brief.EntityType)
			fmt.Fprintf(&b, "    Role: %s\n", brief.Role)
			if brief.DefinedIn != "" {
				fmt.Fprintf(&b, "    Defined in: %s\n", brief.DefinedIn)
			}
		}
		b.WriteString("\n")
	}

	if len(out.Related) > 0 {
		writeEntitySectionCapped(&b, "Related Entities", out.Related, MaxRelatedRefs)
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

// FormatPlan renders a DevelopmentPlan for CLI output.
func FormatPlan(out *DevelopmentPlan) string {
	if out == nil {
		return ""
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Task: %s\n\n", out.Task)

	if len(out.SuggestedSteps) > 0 {
		b.WriteString("Suggested Steps:\n")
		for _, step := range out.SuggestedSteps {
			fmt.Fprintf(&b, "  %s\n", step)
		}
		b.WriteString("\n")
	}

	if len(out.EntityBriefings) > 0 {
		b.WriteString("ENTITY BRIEFINGS\n")
		for _, brief := range out.EntityBriefings {
			fmt.Fprintf(&b, "  %s [%s]\n", brief.QualifiedName, brief.EntityType)
			fmt.Fprintf(&b, "    Layer: %s\n", brief.Layer)
			fmt.Fprintf(&b, "    Role: %s\n", brief.Role)
			if brief.DefinedIn != "" {
				fmt.Fprintf(&b, "    Defined in: %s\n", brief.DefinedIn)
			}
			if len(brief.Fields) > 0 {
				fmt.Fprintf(&b, "    Fields: %s\n", strings.Join(brief.Fields, "; "))
			}
		}
		b.WriteString("\n")
	}

	writeEntitySectionCapped(&b, "Impacted Areas", out.ImpactedAreas, MaxImpactedAreas)
	writeEntitySectionCapped(&b, "Relevant Decisions", out.RelevantDecisions, MaxImpactedAreas)
	writeEntitySectionCapped(&b, "Relevant Facts", out.RelevantFacts, MaxImpactedAreas)
	writeEntitySectionCapped(&b, "Owners", out.Owners, MaxRelatedRefs)
	writeEntitySectionCapped(&b, "Reviewers", out.Reviewers, MaxRelatedRefs)

	if out.SuggestedWorkflow != "" {
		fmt.Fprintf(&b, "Suggested Workflow: %s\n\n", out.SuggestedWorkflow)
	}

	writeEntitySectionCapped(&b, "Starting Points", out.StartingPoints, MaxPlanStartingPoints)

	if len(out.RepositoryCodeLinks) > 0 {
		b.WriteString("Repository-Code Links:\n")
		for _, link := range out.RepositoryCodeLinks {
			fmt.Fprintf(&b, "  - %s: %s -> %s\n",
				link.RelationshipType, link.RepositoryEntityRef.Label, link.CodeEntityRef.Label)
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

// FormatFeatureList renders a feature list for CLI output.
func FormatFeatureList(out *feature.ListResult) string {
	if out == nil || len(out.Features) == 0 {
		return "No features derived for this repository.\n"
	}
	var b strings.Builder
	b.WriteString("Features:\n")
	for _, f := range out.Features {
		fmt.Fprintf(&b, "  - %s", f.Name)
		if len(f.Sources) > 0 {
			fmt.Fprintf(&b, " [%s]", strings.Join(f.Sources, ", "))
		}
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n") + "\n"
}

// FormatOnboarding renders a DevelopmentOnboardingGuide for CLI output.
func FormatOnboarding(out *DevelopmentOnboardingGuide) string {
	if out == nil {
		return ""
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Repository: %s\n\n", out.RepositoryID)

	if len(out.SuggestedSteps) > 0 {
		b.WriteString("Suggested Steps:\n")
		for _, step := range out.SuggestedSteps {
			fmt.Fprintf(&b, "  %s\n", step)
		}
		b.WriteString("\n")
	}

	writeEntitySection(&b, "Key Decisions", out.KeyDecisions)

	if out.Orientation != nil && out.Orientation.Summary != "" {
		b.WriteString("Orientation:\n")
		for _, line := range strings.Split(strings.TrimSpace(out.Orientation.Summary), "\n") {
			fmt.Fprintf(&b, "  %s\n", line)
		}
		b.WriteString("\n")
	}

	if out.AssignmentPlan != nil {
		b.WriteString("Assignment Plan:\n")
		b.WriteString(FormatPlan(out.AssignmentPlan))
	} else if len(out.EntityBriefings) > 0 {
		b.WriteString("Entity Briefings:\n")
		for _, brief := range out.EntityBriefings {
			fmt.Fprintf(&b, "  %s [%s] — %s\n", brief.QualifiedName, brief.EntityType, brief.DefinedIn)
		}
	}

	return strings.TrimRight(b.String(), "\n") + "\n"
}

// FormatImpactReport renders a DevelopmentImpactReport for CLI output.
func FormatImpactReport(out *DevelopmentImpactReport) string {
	if out == nil {
		return ""
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Subject: %s\n\n", out.Subject)

	writeEntitySection(&b, "Impacted Decisions", out.ImpactedDecisions)
	writeEntitySection(&b, "Impacted Facts", out.ImpactedFacts)
	writeEntitySection(&b, "Impacted Events", out.ImpactedEvents)

	if len(out.CodeDependencies) > 0 {
		b.WriteString("Code Dependencies:\n")
		for _, dep := range out.CodeDependencies {
			fmt.Fprintf(&b, "  - %s\n", dep.Label)
		}
		b.WriteString("\n")
	}

	writeEntitySection(&b, "Dependent Areas", out.DependentAreas)
	writeEntitySection(&b, "Owners", out.Owners)

	if len(out.RepositoryCodeLinks) > 0 {
		b.WriteString("Repository-Code Links:\n")
		for _, link := range out.RepositoryCodeLinks {
			fmt.Fprintf(&b, "  - %s: %s -> %s\n",
				link.RelationshipType, link.RepositoryEntityRef.Label, link.CodeEntityRef.Label)
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

// FormatReviewGuide renders a DevelopmentReviewGuide for CLI output.
func FormatReviewGuide(out *DevelopmentReviewGuide) string {
	if out == nil {
		return ""
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Topic: %s\n\n", out.Topic)

	writeEntitySection(&b, "Recommended Reviewers", out.RecommendedReviewers)
	writeEntitySection(&b, "Required Expertise", out.RequiredExpertise)
	writeEntitySection(&b, "Affected Areas", out.AffectedAreas)
	writeEntitySection(&b, "Related Knowledge", out.RelatedKnowledge)

	if out.SuggestedWorkflow != "" {
		fmt.Fprintf(&b, "Suggested Workflow: %s\n\n", out.SuggestedWorkflow)
	}

	if len(out.RepositoryCodeLinks) > 0 {
		b.WriteString("Repository-Code Links:\n")
		for _, link := range out.RepositoryCodeLinks {
			fmt.Fprintf(&b, "  - %s: %s -> %s\n",
				link.RelationshipType, link.RepositoryEntityRef.Label, link.CodeEntityRef.Label)
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

func writeEntitySectionCapped(b *strings.Builder, title string, refs []EntityRef, limit int) {
	if len(refs) == 0 {
		return
	}
	capped, omitted := capEntityRefs(refs, limit)
	fmt.Fprintf(b, "%s:\n", title)
	for _, ref := range capped {
		fmt.Fprintf(b, "  - %s\n", formatEntityRefLine(ref))
	}
	if omitted > 0 {
		fmt.Fprintf(b, "  - … +%d more (use explain_function or explain_file)\n", omitted)
	}
}

func writeEntitySection(b *strings.Builder, title string, refs []EntityRef) {
	if len(refs) == 0 {
		return
	}
	fmt.Fprintf(b, "%s:\n", title)
	for _, ref := range refs {
		fmt.Fprintf(b, "  - %s\n", formatEntityRefLine(ref))
	}
}

func writeEntitySectionIndented(b *strings.Builder, prefix, title string, refs []EntityRef) {
	if len(refs) == 0 {
		return
	}
	fmt.Fprintf(b, "%s%s:\n", prefix, title)
	for _, ref := range refs {
		fmt.Fprintf(b, "%s  - %s\n", prefix, formatEntityRefLine(ref))
	}
}

func formatEntityRefLine(ref EntityRef) string {
	label := strings.TrimSpace(ref.Label)
	if label == "" {
		label = ref.EntityID
	}
	if ref.EntityType == "" {
		return label
	}
	return fmt.Sprintf("%s [%s]", label, ref.EntityType)
}
