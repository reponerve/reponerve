package hookcmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/config"
)

const hookMarker = "# reponerve"

const postCommitScript = `#!/bin/sh
# reponerve — refresh repository memory after commit
reponerve scan >/dev/null 2>&1 || true
`

// NewCommand creates the hook subcommand group.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hook",
		Short: "Install git hooks for automatic repository scans",
		Long:  `Install or remove a post-commit hook that runs reponerve scan after each commit.`,
	}
	cmd.AddCommand(newInstallCommand(), newUninstallCommand(), newStatusCommand())
	return cmd
}

func newInstallCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install post-commit hook in the current git repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			hookPath, err := postCommitHookPath()
			if err != nil {
				return err
			}
			if err := os.MkdirAll(filepath.Dir(hookPath), 0o755); err != nil {
				return fmt.Errorf("create hooks directory: %w", err)
			}

			existing := ""
			if data, err := os.ReadFile(hookPath); err == nil {
				existing = string(data)
			} else if !os.IsNotExist(err) {
				return fmt.Errorf("read hook: %w", err)
			}

			if strings.Contains(existing, hookMarker) {
				cmd.Println("✓ post-commit hook already installed")
				return nil
			}

			var content string
			if strings.TrimSpace(existing) == "" {
				content = postCommitScript
			} else {
				content = strings.TrimRight(existing, "\n") + "\n\n" + postCommitScript
			}

			if err := os.WriteFile(hookPath, []byte(content), 0o755); err != nil {
				return fmt.Errorf("write hook: %w", err)
			}
			cmd.Printf("✓ Installed post-commit hook at %s\n", hookPath)
			return nil
		},
	}
}

func newUninstallCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: "Remove reponerve block from post-commit hook",
		RunE: func(cmd *cobra.Command, args []string) error {
			hookPath, err := postCommitHookPath()
			if err != nil {
				return err
			}
			data, err := os.ReadFile(hookPath)
			if os.IsNotExist(err) {
				cmd.Println("✓ post-commit hook not present")
				return nil
			}
			if err != nil {
				return fmt.Errorf("read hook: %w", err)
			}
			text := string(data)
			if !strings.Contains(text, hookMarker) {
				cmd.Println("✓ reponerve block not found in post-commit hook")
				return nil
			}
			idx := strings.Index(text, hookMarker)
			updated := strings.TrimSpace(text[:idx])
			if updated == "" {
				if err := os.Remove(hookPath); err != nil {
					return fmt.Errorf("remove hook: %w", err)
				}
				cmd.Println("✓ Removed post-commit hook")
				return nil
			}
			if !strings.HasSuffix(updated, "\n") {
				updated += "\n"
			}
			if err := os.WriteFile(hookPath, []byte(updated), 0o755); err != nil {
				return fmt.Errorf("write hook: %w", err)
			}
			cmd.Println("✓ Removed reponerve block from post-commit hook")
			return nil
		},
	}
}

func newStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show whether reponerve post-commit hook is installed",
		RunE: func(cmd *cobra.Command, args []string) error {
			hookPath, err := postCommitHookPath()
			if err != nil {
				return err
			}
			data, err := os.ReadFile(hookPath)
			if os.IsNotExist(err) {
				cmd.Println("post-commit: not installed")
				return nil
			}
			if err != nil {
				return fmt.Errorf("read hook: %w", err)
			}
			if strings.Contains(string(data), hookMarker) {
				cmd.Printf("post-commit: installed (%s)\n", hookPath)
				return nil
			}
			cmd.Printf("post-commit: present without reponerve block (%s)\n", hookPath)
			return nil
		},
	}
}

func postCommitHookPath() (string, error) {
	repoPath, err := repositoryPath()
	if err != nil {
		return "", err
	}
	gitDir := filepath.Join(repoPath, ".git")
	if info, err := os.Stat(gitDir); err == nil && !info.IsDir() {
		// worktree: .git is a file
		data, err := os.ReadFile(gitDir)
		if err != nil {
			return "", fmt.Errorf("read .git file: %w", err)
		}
		line := strings.TrimSpace(string(data))
		if strings.HasPrefix(line, "gitdir: ") {
			gitDir = strings.TrimSpace(strings.TrimPrefix(line, "gitdir: "))
		}
	}
	return filepath.Join(gitDir, "hooks", "post-commit"), nil
}

func repositoryPath() (string, error) {
	cfg, err := config.Load(config.GetWorkspaceDir())
	if err == nil && strings.TrimSpace(cfg.Repository.Path) != "" {
		return filepath.Clean(cfg.Repository.Path), nil
	}
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("not a git repository; run reponerve init first")
	}
	return strings.TrimSpace(string(out)), nil
}
