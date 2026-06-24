package discipline

import (
	"fmt"
	"strings"
)

// FormatPolicy renders a policy for CLI prose output.
func FormatPolicy(p *Policy) string {
	if p == nil {
		return "No discipline policy found. Run reponerve scan.\n"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Repository: %s\n", p.RepositoryID)
	if p.DominantLanguage != "" {
		fmt.Fprintf(&b, "Dominant language: %s\n", p.DominantLanguage)
	}
	if p.ADRDirectory != "" {
		fmt.Fprintf(&b, "ADR directory: %s\n", p.ADRDirectory)
	}
	if len(p.CIWorkflowFiles) > 0 {
		b.WriteString("CI workflows:\n")
		for _, f := range p.CIWorkflowFiles {
			fmt.Fprintf(&b, "  - %s\n", f)
		}
	}
	if len(p.LayerConventions) > 0 {
		b.WriteString("Layer conventions:\n")
		for _, layer := range p.LayerConventions {
			fmt.Fprintf(&b, "  - %s %s\n", layer.Prefix, layer.Role)
		}
	}
	if len(p.ShipCheckHints) > 0 {
		b.WriteString("Ship check hints:\n")
		for _, hint := range p.ShipCheckHints {
			fmt.Fprintf(&b, "  - %s\n", hint)
		}
	}
	return strings.TrimRight(b.String(), "\n") + "\n"
}
