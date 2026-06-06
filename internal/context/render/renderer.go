package render

import (
	"strings"
	"time"

	"reponerve/internal/context"
)

type Renderer struct{}

func NewRenderer() *Renderer {
	return &Renderer{}
}

// Render converts a RepositoryContext into formatted markdown.
func (r *Renderer) Render(rc *context.RepositoryContext) (string, error) {
	var sb strings.Builder

	sb.WriteString("# Repository Context\n\n")
	sb.WriteString("Repository: " + rc.RepositoryID + "\n\n")
	sb.WriteString("Generated: " + rc.GeneratedAt.UTC().Format(time.RFC3339) + "\n")

	if len(rc.Decisions) > 0 {
		sb.WriteString("\n## Key Decisions\n\n")
		for _, dec := range rc.Decisions {
			sb.WriteString("* " + dec.Title + "\n")
		}
	}

	if len(rc.Intents) > 0 {
		sb.WriteString("\n## Key Intents\n\n")
		for _, intent := range rc.Intents {
			sb.WriteString("* " + intent.Description + "\n")
		}
	}

	if len(rc.Facts) > 0 {
		sb.WriteString("\n## Key Facts\n\n")
		for _, fact := range rc.Facts {
			sb.WriteString("* " + fact.Subject + " " + fact.Predicate + " " + fact.Object + "\n")
		}
	}

	if len(rc.Events) > 0 {
		sb.WriteString("\n## Recent Events\n\n")
		for _, event := range rc.Events {
			sb.WriteString("* " + event.Title + "\n")
		}
	}

	return sb.String(), nil
}
