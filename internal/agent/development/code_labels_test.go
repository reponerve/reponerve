package development_test

import (
	"strings"
	"testing"

	"github.com/reponerve/reponerve/internal/agent/development"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
)

func TestHumanCodeEntityLabel_MethodAndStruct(t *testing.T) {
	method := &codemodels.CodeEntity{
		EntityType:  codemodels.EntityTypeMethod,
		Name:        "Ask",
		PackagePath: "internal/agent/development",
	}
	label := development.HumanCodeEntityLabel(method, "CALLS")
	if !strings.Contains(label, "method Ask") || !strings.Contains(label, "internal/agent/development") {
		t.Fatalf("unexpected method label: %q", label)
	}

	structEnt := &codemodels.CodeEntity{
		EntityType:  codemodels.EntityTypeStruct,
		Name:        "RepositoryContext",
		PackagePath: "internal/context",
	}
	label = development.HumanCodeEntityLabel(structEnt, "CALLS")
	if !strings.Contains(label, "struct RepositoryContext") {
		t.Fatalf("unexpected struct label: %q", label)
	}
}
