package agentcontext

import (
	"encoding/json"
	"fmt"
)

// Supported source values for ContextSection.
const (
	SourceDiscovery  = "discovery"
	SourceLearning   = "learning"
	SourceReviewers  = "reviewers"
	SourceChangePlan = "changeplan"
	SourceContext    = "context"
)

// validSources is the set of accepted source values.
var validSources = map[string]bool{
	SourceDiscovery:  true,
	SourceLearning:   true,
	SourceReviewers:  true,
	SourceChangePlan: true,
	SourceContext:    true,
}

// ContextSection holds one named, sourced slice of packaged intelligence.
type ContextSection struct {
	Name   string          `json:"name"`
	Source string          `json:"source"`
	Data   json.RawMessage `json:"data"`
}

// AgentContextPackage is a structured, deterministic repository context bundle
// produced for AI agent consumption. It contains a fixed, ordered set of
// ContextSections derived from existing Repository Intelligence services.
type AgentContextPackage struct {
	RepositoryID string            `json:"repository_id"`
	Sections     []*ContextSection `json:"sections"`
}

// ValidatePackage ensures that a package is structurally valid before it is
// returned to callers. It does not validate the contents of individual section
// payloads — that responsibility remains with each upstream service.
func ValidatePackage(pkg *AgentContextPackage) error {
	if pkg == nil {
		return fmt.Errorf("package is nil")
	}
	if pkg.RepositoryID == "" {
		return fmt.Errorf("missing repository ID")
	}
	if len(pkg.Sections) == 0 {
		return fmt.Errorf("package has no sections")
	}
	for i, section := range pkg.Sections {
		if section == nil {
			return fmt.Errorf("section %d is nil", i)
		}
		if section.Name == "" {
			return fmt.Errorf("section %d: missing name", i)
		}
		if section.Source == "" {
			return fmt.Errorf("section %d (%q): missing source", i, section.Name)
		}
		if !validSources[section.Source] {
			return fmt.Errorf("section %d (%q): unsupported source %q (must be one of: discovery, learning, reviewers, changeplan, context)", i, section.Name, section.Source)
		}
		if len(section.Data) == 0 {
			return fmt.Errorf("section %d (%q): missing data", i, section.Name)
		}
	}
	return nil
}
