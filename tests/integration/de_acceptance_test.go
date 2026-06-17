package integration

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/reponerve/reponerve/internal/cli"
)

// Acceptance criteria: docs/examples/development-experience.md (ISSUE-057 checklist).
func TestDevelopmentExperienceAcceptance(t *testing.T) {
	repoDir := setupDEAcceptanceRepo(t)
	execute := initAndScanDEAcceptance(t, repoDir)

	t.Run("ask ownership", func(t *testing.T) {
		out, err := execute("ask", "Who owns authentication?")
		if err != nil {
			t.Fatalf("ask ownership failed: %v\n%s", err, out)
		}
		assertContainsAll(t, out,
			"Question: Who owns authentication?",
			"Answer Type:",
			"Source Services:",
		)
		assertHasEvidenceOrSummary(t, out)
	})

	t.Run("ask decision rationale use verb", func(t *testing.T) {
		out, err := execute("ask", "Why do we use Redis?")
		if err != nil {
			t.Fatalf("ask use-verb failed: %v\n%s", err, out)
		}
		if !strings.Contains(out, "decision_rationale") {
			t.Fatalf("expected decision_rationale, got:\n%s", out)
		}
	})

	t.Run("ask decision rationale", func(t *testing.T) {
		out, err := execute("ask", "Why are we using Redis?")
		if err != nil {
			t.Fatalf("ask redis failed: %v\n%s", err, out)
		}
		assertContainsAll(t, out,
			"Question: Why are we using Redis?",
			"Answer Type:",
			"decision_rationale",
		)
		if !strings.Contains(strings.ToLower(out), "caching") && !strings.Contains(strings.ToLower(out), "latency") {
			t.Fatalf("expected ADR rationale content in output:\n%s", out)
		}
	})

	t.Run("ask task intake", func(t *testing.T) {
		out, err := execute("ask", "PROJ-99: Add audit logging")
		if err != nil {
			t.Fatalf("ask task intake failed: %v\n%s", err, out)
		}
		assertContainsAll(t, out,
			"Answer Type:",
			"task_plan",
		)
	})

	t.Run("ask authorship", func(t *testing.T) {
		out, err := execute("ask", "Who created metadata panel?")
		if err != nil {
			t.Fatalf("ask authorship failed: %v\n%s", err, out)
		}
		assertContainsAll(t, out,
			"Question: Who created metadata panel?",
			"Answer Type:",
		)
		assertHasEvidenceOrSummary(t, out)
	})

	t.Run("explain topic", func(t *testing.T) {
		out, err := execute("explain", "metadata panel")
		if err != nil {
			t.Fatalf("explain topic failed: %v\n%s", err, out)
		}
		assertContainsAll(t, out,
			"Topic: metadata panel",
			"Evidence:",
			"Source Services:",
		)
		if !strings.Contains(out, "CODE CONTEXT") && !strings.Contains(out, "REPOSITORY CONTEXT") {
			t.Fatalf("expected code or repository context in explain output:\n%s", out)
		}
		assertNoNarrativeFields(t, out)
	})

	t.Run("explain-file", func(t *testing.T) {
		out, err := execute("explain-file", "internal/agent/search/service.go")
		if err != nil {
			t.Fatalf("explain-file failed: %v\n%s", err, out)
		}
		assertContainsAll(t, out,
			"Topic: internal/agent/search/service.go",
			"CODE CONTEXT",
			"Files:",
			"Source Services:",
		)
	})

	t.Run("explain-function", func(t *testing.T) {
		out, err := execute("explain-function", "Search", "--package", "internal/agent/search")
		if err != nil {
			t.Fatalf("explain-function failed: %v\n%s", err, out)
		}
		assertContainsAll(t, out,
			"Topic: Search",
			"CODE CONTEXT",
			"Source Services:",
		)
	})

	t.Run("explain-function homonym", func(t *testing.T) {
		out, err := execute("explain-function", "Search")
		if err != nil {
			t.Fatalf("explain-function homonym failed: %v\n%s", err, out)
		}
		if !strings.Contains(out, "ENTITY BRIEFINGS") && !strings.Contains(strings.ToLower(out), "disambiguat") {
			t.Fatalf("expected ambiguous symbol handling in output:\n%s", out)
		}
	})

	t.Run("explain-struct", func(t *testing.T) {
		out, err := execute("explain-struct", "MetadataPanel")
		if err != nil {
			t.Fatalf("explain-struct failed: %v\n%s", err, out)
		}
		assertContainsAll(t, out,
			"Topic: MetadataPanel",
			"CODE CONTEXT",
			"Structs:",
		)
	})

	t.Run("explain-interface", func(t *testing.T) {
		out, err := execute("explain-interface", "PanelRenderer")
		if err != nil {
			t.Fatalf("explain-interface failed: %v\n%s", err, out)
		}
		assertContainsAll(t, out,
			"Topic: PanelRenderer",
			"CODE CONTEXT",
			"Interfaces:",
		)
	})

	t.Run("explain-type", func(t *testing.T) {
		out, err := execute("explain-type", "MetadataID")
		if err != nil {
			t.Fatalf("explain-type failed: %v\n%s", err, out)
		}
		assertContainsAll(t, out,
			"Topic: MetadataID",
			"CODE CONTEXT",
			"Type Aliases:",
		)
	})

	t.Run("plan", func(t *testing.T) {
		out, err := execute("plan", "Add OAuth login")
		if err != nil {
			t.Fatalf("plan failed: %v\n%s", err, out)
		}
		assertContainsAll(t, out,
			"Task: Add OAuth login",
			"Suggested Workflow: change_preparation",
			"Evidence:",
			"Source Services:",
		)
	})

	t.Run("impact", func(t *testing.T) {
		out, err := execute("impact", "user-service")
		if err != nil {
			t.Fatalf("impact failed: %v\n%s", err, out)
		}
		assertContainsAll(t, out,
			"Subject: user-service",
			"Evidence:",
			"Source Services:",
		)
	})

	t.Run("review", func(t *testing.T) {
		out, err := execute("review", "metadata panel")
		if err != nil {
			t.Fatalf("review failed: %v\n%s", err, out)
		}
		assertContainsAll(t, out,
			"Topic: metadata panel",
			"Suggested Workflow: review_preparation",
			"Evidence:",
			"Source Services:",
		)
	})

	t.Run("second scan preserves ownership", func(t *testing.T) {
		first, err := execute("ask", "Who owns authentication?")
		if err != nil {
			t.Fatalf("first ownership ask failed: %v\n%s", err, first)
		}

		middlewarePath := filepath.Join(repoDir, "internal/auth/middleware.go")
		if err := os.WriteFile(middlewarePath, []byte(`package auth

// Middleware applies authentication checks.
func Middleware() {}
`), 0644); err != nil {
			t.Fatalf("write middleware file: %v", err)
		}
		runGitCommitAs(t, repoDir, "alice@example.com", "Alice Example",
			"feat: extend authentication middleware", "internal/auth/middleware.go")

		buf := new(bytes.Buffer)
		cmd := cli.NewRootCmd()
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"scan"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("second scan failed: %v\n%s", err, buf.String())
		}

		second, err := execute("ask", "Who owns authentication?")
		if err != nil {
			t.Fatalf("second ownership ask failed: %v\n%s", err, second)
		}
		if !strings.Contains(second, "Answer Type:") {
			t.Fatalf("expected ownership answer after rescan:\n%s", second)
		}
		if strings.Contains(second, "commit_count=0") {
			t.Fatalf("ownership data corrupted after rescan:\n%s", second)
		}
	})

	t.Run("determinism", func(t *testing.T) {
		first, err := execute("explain", "metadata panel")
		if err != nil {
			t.Fatalf("first explain failed: %v", err)
		}
		second, err := execute("explain", "metadata panel")
		if err != nil {
			t.Fatalf("second explain failed: %v", err)
		}
		if first != second {
			t.Fatalf("deterministic explain output mismatch:\nfirst:\n%s\nsecond:\n%s", first, second)
		}
	})
}

func setupDEAcceptanceRepo(t *testing.T) string {
	t.Helper()

	fixtureRoot := filepath.Join("..", "fixtures", "de-acceptance")
	absFixture, err := filepath.Abs(fixtureRoot)
	if err != nil {
		t.Fatalf("resolve fixture path: %v", err)
	}

	tempDir, err := os.MkdirTemp("", "reponerve-de-acceptance-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(tempDir) })

	if err := copyDir(absFixture, tempDir); err != nil {
		t.Fatalf("copy fixture: %v", err)
	}

	runGitCommand(t, tempDir, "init")
	runGitCommand(t, tempDir, "config", "user.name", "Alice Example")
	runGitCommand(t, tempDir, "config", "user.email", "alice@example.com")

	runGitCommitAs(t, tempDir, "alice@example.com", "Alice Example",
		"feat: introduce user service handler", "internal/service/user/handler.go")
	runGitCommitAs(t, tempDir, "alice@example.com", "Alice Example",
		"feat: introduce metadata panel UI", "internal/ui/metadata/panel.go")
	runGitCommitAs(t, tempDir, "bob@example.com", "Bob Example",
		"feat: add authentication service with OAuth login", "internal/auth/service.go", "internal/store/store.go")
	runGitCommitAs(t, tempDir, "carol@example.com", "Carol Example",
		"feat: add deterministic repository search service", "internal/agent/search/service.go", "internal/api/search/api.go")
	runGitCommitAs(t, tempDir, "alice@example.com", "Alice Example",
		"docs: add architecture decision records for authentication metadata and Redis cache",
		"docs/adr/0001-authentication.md", "docs/adr/0002-metadata-ui.md", "docs/adr/0003-redis-cache.md")
	runGitCommand(t, tempDir, "branch", "-M", "main")

	return tempDir
}

func initAndScanDEAcceptance(t *testing.T, repoDir string) func(args ...string) (string, error) {
	t.Helper()

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(repoDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	workspaceDir := filepath.Join(repoDir, ".reponerve")
	os.Setenv("REPONERVE_WORKSPACE", workspaceDir)
	t.Cleanup(func() { os.Unsetenv("REPONERVE_WORKSPACE") })

	for _, args := range [][]string{{"init"}, {"scan"}} {
		buf := new(bytes.Buffer)
		cmd := cli.NewRootCmd()
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs(args)
		if err := cmd.Execute(); err != nil {
			t.Fatalf("reponerve %s failed: %v\n%s", args[0], err, buf.String())
		}
	}

	return func(args ...string) (string, error) {
		cmd := cli.NewRootCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs(args)
		err := cmd.Execute()
		return buf.String(), err
	}
}

func runGitCommitAs(t *testing.T, dir, email, name, message string, paths ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"add"}, paths...)...)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git add failed: %v", err)
	}
	commit := exec.Command("git", "commit", "-m", message)
	commit.Dir = dir
	commit.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME="+name,
		"GIT_AUTHOR_EMAIL="+email,
		"GIT_COMMITTER_NAME="+name,
		"GIT_COMMITTER_EMAIL="+email,
	)
	if err := commit.Run(); err != nil {
		t.Fatalf("git commit failed: %v", err)
	}
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		return copyFile(path, target, info.Mode())
	})
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func assertContainsAll(t *testing.T, output string, parts ...string) {
	t.Helper()
	for _, part := range parts {
		if !strings.Contains(output, part) {
			t.Fatalf("expected output to contain %q, got:\n%s", part, output)
		}
	}
}

func assertHasEvidenceOrSummary(t *testing.T, output string) {
	t.Helper()
	if strings.Contains(output, "Evidence:") || strings.Contains(output, "Summary:") {
		return
	}
	t.Fatalf("expected evidence or summary in output:\n%s", output)
}

func assertNoNarrativeFields(t *testing.T, output string) {
	t.Helper()
	for _, forbidden := range []string{"Purpose:", "History:"} {
		if strings.Contains(output, forbidden) {
			t.Fatalf("narrative field %q must not appear in DE output:\n%s", forbidden, output)
		}
	}
}
