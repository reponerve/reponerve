package health

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/reponerve/reponerve/internal/agent/discipline"
	"github.com/reponerve/reponerve/internal/code/indexer"
	"github.com/reponerve/reponerve/internal/config"
	"github.com/reponerve/reponerve/internal/scanner/repository"
	"github.com/reponerve/reponerve/internal/storage"
)

const (
	StatusOK   = "ok"
	StatusWarn = "warn"
	StatusFail = "fail"
)

// Check describes one doctor finding.
type Check struct {
	Name     string            `json:"name"`
	Status   string            `json:"status"`
	Message  string            `json:"message"`
	Evidence map[string]string `json:"evidence,omitempty"`
}

// DoctorResult is the structured doctor output.
type DoctorResult struct {
	OK              bool     `json:"ok"`
	Checks          []Check  `json:"checks"`
	Recommendations []string `json:"recommendations"`
}

// Checker runs freshness diagnostics.
type Checker struct {
	ScanStateStore    storage.ScanStateStore
	CodeIndexStateStore storage.CodeIndexStateStore
	Discovery         repository.Discovery
}

// NewChecker creates a doctor checker.
func NewChecker(
	scanState storage.ScanStateStore,
	codeIndexState storage.CodeIndexStateStore,
	discovery repository.Discovery,
) *Checker {
	if discovery == nil {
		discovery = repository.NewGitDiscovery()
	}
	return &Checker{
		ScanStateStore:      scanState,
		CodeIndexStateStore: codeIndexState,
		Discovery:           discovery,
	}
}

// CheckInput configures a doctor run.
type CheckInput struct {
	WorkspaceDir   string
	RepositoryPath string
}

// Check runs all doctor checks.
func (c *Checker) Check(ctx context.Context, in CheckInput) (*DoctorResult, error) {
	workspaceDir := in.WorkspaceDir
	if workspaceDir == "" {
		workspaceDir = config.GetWorkspaceDir()
	}

	result := &DoctorResult{Checks: []Check{}}
	add := func(name, status, msg string, evidence map[string]string) {
		result.Checks = append(result.Checks, Check{
			Name: name, Status: status, Message: msg, Evidence: evidence,
		})
	}

	cfgPath := filepath.Join(workspaceDir, "config.yaml")
	memPath := filepath.Join(workspaceDir, "memory.db")
	if _, err := os.Stat(cfgPath); err != nil {
		add("workspace", StatusFail, "workspace not initialized — run reponerve init", nil)
		result.Recommendations = append(result.Recommendations, "reponerve init")
		result.OK = false
		return result, nil
	}
	if _, err := os.Stat(memPath); err != nil {
		add("memory_db", StatusFail, "memory database missing — run reponerve scan", map[string]string{"path": memPath})
		result.Recommendations = append(result.Recommendations, "reponerve scan")
		result.OK = false
		return result, nil
	}
	add("workspace", StatusOK, "workspace initialized", map[string]string{"path": workspaceDir})

	cfg, err := config.Load(workspaceDir)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	repoPath := in.RepositoryPath
	if repoPath == "" {
		repoPath = cfg.Repository.Path
	}

	repo, err := c.Discovery.Discover(ctx, repoPath)
	if err != nil {
		add("repository", StatusFail, fmt.Sprintf("repository discovery failed: %v", err), nil)
		result.OK = false
		return result, nil
	}
	add("repository", StatusOK, "repository discovered", map[string]string{"id": repo.ID, "path": repo.Path})

	scanState, err := c.ScanStateStore.GetScanState(ctx, repo.ID)
	if err != nil {
		return nil, fmt.Errorf("scan state: %w", err)
	}
	if scanState == nil || scanState.LastScanCommit == "" {
		add("scan_state", StatusWarn, "no scan recorded — run reponerve scan", nil)
		result.Recommendations = append(result.Recommendations, "reponerve scan")
	} else {
		add("scan_state", StatusOK, "scan state present", map[string]string{
			"last_scan_commit": scanState.LastScanCommit,
			"updated_at":       scanState.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}

	head, err := gitHEAD(ctx, repoPath)
	if err != nil {
		add("git_head", StatusWarn, fmt.Sprintf("could not read git HEAD: %v", err), nil)
	} else if scanState == nil || scanState.LastScanCommit == "" {
		add("git_freshness", StatusWarn, "scan never completed — memory may be empty", map[string]string{"head": head})
		result.Recommendations = appendUnique(result.Recommendations, "reponerve scan")
	} else if head != scanState.LastScanCommit {
		add("git_freshness", StatusWarn, "git moved since last scan", map[string]string{
			"head":             head,
			"last_scan_commit": scanState.LastScanCommit,
		})
		result.Recommendations = appendUnique(result.Recommendations, "reponerve scan")
	} else {
		add("git_freshness", StatusOK, "memory matches git HEAD", map[string]string{"head": head})
	}

	codeState, err := c.CodeIndexStateStore.GetByRepository(ctx, repo.ID)
	if err != nil {
		return nil, fmt.Errorf("code index state: %w", err)
	}
	if codeState == nil || codeState.LastIndexedAt.IsZero() {
		add("code_index", StatusWarn, "code index never built — run reponerve scan", nil)
		result.Recommendations = appendUnique(result.Recommendations, "reponerve scan")
	} else {
		stale, serr := indexer.RepositorySourceFilesChangedSince(repoPath, codeState.LastIndexedAt)
		if serr != nil {
			add("code_index", StatusWarn, fmt.Sprintf("code index freshness check failed: %v", serr), nil)
		} else if stale {
			add("code_index", StatusWarn, "source files changed since last code index", map[string]string{
				"last_indexed_at": codeState.LastIndexedAt.UTC().Format(time.RFC3339),
			})
			result.Recommendations = appendUnique(result.Recommendations, "reponerve scan")
		} else {
			add("code_index", StatusOK, "code index is current", map[string]string{
				"last_indexed_at": codeState.LastIndexedAt.UTC().Format(time.RFC3339),
				"entity_count":    fmt.Sprintf("%d", codeState.EntityCount),
			})
		}
	}

	policy, perr := discipline.LoadPolicy(workspaceDir)
	if perr != nil || policy == nil {
		add("discipline_policy", StatusWarn, "discipline policy missing — run reponerve scan", nil)
		result.Recommendations = appendUnique(result.Recommendations, "reponerve scan")
	} else {
		add("discipline_policy", StatusOK, "discipline policy present", map[string]string{
			"generated_at": policy.GeneratedAt.UTC().Format(time.RFC3339),
		})
	}

	if installed, msg := postCommitHookStatus(repoPath); installed {
		add("post_commit_hook", StatusOK, msg, nil)
	} else {
		add("post_commit_hook", StatusOK, msg, map[string]string{"advisory": "true"})
	}

	result.OK = true
	for _, chk := range result.Checks {
		if chk.Status == StatusFail || chk.Status == StatusWarn {
			result.OK = false
			break
		}
	}
	return result, nil
}

// FormatDoctor renders a prose summary.
func FormatDoctor(r *DoctorResult) string {
	if r == nil {
		return "Doctor: no result"
	}
	var b strings.Builder
	if r.OK {
		b.WriteString("RepoNerve doctor: all checks passed.\n")
	} else {
		b.WriteString("RepoNerve doctor: action recommended.\n")
	}
	for _, chk := range r.Checks {
		b.WriteString(fmt.Sprintf("  [%s] %s: %s\n", chk.Status, chk.Name, chk.Message))
	}
	if len(r.Recommendations) > 0 {
		b.WriteString("\nRecommendations:\n")
		for _, rec := range r.Recommendations {
			b.WriteString("  - " + rec + "\n")
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

func gitHEAD(ctx context.Context, repoPath string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func postCommitHookStatus(repoPath string) (bool, string) {
	gitDir, err := gitDir(repoPath)
	if err != nil {
		return false, "not a git repository — hook not applicable"
	}
	hookPath := filepath.Join(gitDir, "hooks", "post-commit")
	data, err := os.ReadFile(hookPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, "post-commit hook not installed (optional)"
		}
		return false, "could not read post-commit hook"
	}
	if strings.Contains(string(data), "# reponerve") {
		return true, "post-commit hook installed"
	}
	return false, "post-commit hook present but reponerve block missing"
}

func gitDir(repoPath string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	p := strings.TrimSpace(string(out))
	if !filepath.IsAbs(p) {
		p = filepath.Join(repoPath, p)
	}
	return filepath.Clean(p), nil
}

func appendUnique(list []string, item string) []string {
	for _, existing := range list {
		if existing == item {
			return list
		}
	}
	return append(list, item)
}
