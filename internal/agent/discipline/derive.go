package discipline

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	codemodels "github.com/reponerve/reponerve/internal/code/models"
	"github.com/reponerve/reponerve/internal/scanner/adr"
)

const sourceDisciplinePolicy = "discipline_policy"

// DeriveInput is evidence for generating a repository discipline policy.
type DeriveInput struct {
	RepositoryID   string
	RepositoryPath string
	ADRsIndexed    int
	CodeEntities   []*codemodels.CodeEntity
	DocumentPaths  []adr.DocumentPath
}

// Derive builds a deterministic discipline policy from repository layout and scan evidence.
func Derive(_ context.Context, input DeriveInput) *Policy {
	repoPath := filepath.Clean(input.RepositoryPath)
	policy := &Policy{
		RepositoryID:   input.RepositoryID,
		GeneratedAt:    time.Now().UTC(),
		SourceServices: []string{sourceDisciplinePolicy},
	}

	docPaths := input.DocumentPaths
	if len(docPaths) == 0 {
		docPaths = adr.ResolveDocumentPaths(nil)
	}
	if adrDir := adr.PrimaryADRDirectory(repoPath, docPaths); adrDir != "" {
		policy.ADRDirectory = adrDir
		policy.RequireADROnArchitecture = true
		policy.ShipCheckHints = append(policy.ShipCheckHints,
			"Offer a new ADR in "+adrDir+" for significant architecture changes",
		)
	}

	policy.CIWorkflowFiles = detectCIWorkflowFiles(repoPath)
	if len(policy.CIWorkflowFiles) > 0 {
		policy.ShipCheckHints = append(policy.ShipCheckHints,
			"Verify CI workflows pass before ship: "+strings.Join(policy.CIWorkflowFiles, ", "),
		)
	}

	policy.DominantLanguage = detectDominantLanguage(repoPath, input.CodeEntities)
	policy.LayerConventions = detectLayerConventions(repoPath)

	if policy.DominantLanguage == "go" && len(policy.LayerConventions) > 0 {
		policy.ShipCheckHints = append(policy.ShipCheckHints,
			"Respect repository layer prefixes when adding or moving packages",
		)
	}

	sort.Strings(policy.CIWorkflowFiles)
	sort.Strings(policy.ShipCheckHints)
	return policy
}

func detectCIWorkflowFiles(repoPath string) []string {
	var files []string
	checks := []string{
		".github/workflows",
	}
	for _, rel := range checks {
		dir := filepath.Join(repoPath, rel)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := strings.ToLower(e.Name())
			if strings.HasSuffix(name, ".yml") || strings.HasSuffix(name, ".yaml") {
				files = append(files, filepath.Join(rel, e.Name()))
			}
		}
	}
	for _, rel := range []string{".gitlab-ci.yml", "Jenkinsfile", "azure-pipelines.yml", "bitbucket-pipelines.yml"} {
		if fileExists(filepath.Join(repoPath, rel)) {
			files = append(files, rel)
		}
	}
	sort.Strings(files)
	return files
}

func detectDominantLanguage(repoPath string, entities []*codemodels.CodeEntity) string {
	if fileExists(filepath.Join(repoPath, "go.mod")) {
		return "go"
	}
	if fileExists(filepath.Join(repoPath, "package.json")) {
		return "javascript"
	}
	if fileExists(filepath.Join(repoPath, "Cargo.toml")) {
		return "rust"
	}
	if fileExists(filepath.Join(repoPath, "pyproject.toml")) || fileExists(filepath.Join(repoPath, "setup.py")) {
		return "python"
	}

	counts := map[string]int{}
	for _, e := range entities {
		if e == nil || e.FilePath == "" {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.FilePath))
		if ext == "" {
			continue
		}
		counts[ext]++
	}
	bestExt := ""
	best := 0
	for ext, n := range counts {
		if n > best {
			best = n
			bestExt = ext
		}
	}
	switch bestExt {
	case ".go":
		return "go"
	case ".ts", ".tsx":
		return "typescript"
	case ".js", ".jsx":
		return "javascript"
	case ".py":
		return "python"
	case ".rs":
		return "rust"
	default:
		return ""
	}
}

var layerCandidates = []LayerConvention{
	{Prefix: "cmd/", Role: "CLI entry and command wiring"},
	{Prefix: "internal/", Role: "core implementation"},
	{Prefix: "pkg/", Role: "public API surface"},
	{Prefix: "api/", Role: "HTTP/API layer"},
}

func detectLayerConventions(repoPath string) []LayerConvention {
	var out []LayerConvention
	for _, layer := range layerCandidates {
		if dirExists(filepath.Join(repoPath, strings.TrimSuffix(layer.Prefix, "/"))) {
			out = append(out, layer)
		}
	}
	return out
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
