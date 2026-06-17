package hookcmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestHookInstallUninstall(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	install := newInstallCommand()
	install.SetOut(os.Stdout)
	install.SetErr(os.Stderr)
	if err := install.RunE(install, nil); err != nil {
		t.Fatalf("hook install: %v", err)
	}

	hookPath := filepath.Join(dir, ".git", "hooks", "post-commit")
	data, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("read hook: %v", err)
	}
	if !strings.Contains(string(data), hookMarker) {
		t.Fatalf("missing marker in hook: %s", data)
	}

	status := newStatusCommand()
	status.SetOut(os.Stdout)
	if err := status.RunE(status, nil); err != nil {
		t.Fatalf("hook status: %v", err)
	}

	uninstall := newUninstallCommand()
	uninstall.SetOut(os.Stdout)
	if err := uninstall.RunE(uninstall, nil); err != nil {
		t.Fatalf("hook uninstall: %v", err)
	}

	data, err = os.ReadFile(hookPath)
	if err == nil && strings.Contains(string(data), hookMarker) {
		t.Fatalf("marker still present after uninstall: %s", data)
	}
}

func TestPostCommitHookPath(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	path, err := postCommitHookPath()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, ".git", "hooks", "post-commit")
	want, _ = filepath.EvalSymlinks(want)
	got, _ := filepath.EvalSymlinks(path)
	if got != want {
		t.Fatalf("got %q want %q", path, want)
	}
}

func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	run := func(args ...string) {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %s", err, out)
		}
	}
	run("git", "init")
	run("git", "config", "user.email", "test@example.com")
	run("git", "config", "user.name", "Test")
	_ = os.WriteFile(filepath.Join(dir, "README.md"), []byte("x"), 0o644)
	run("git", "add", "README.md")
	run("git", "commit", "-m", "init")
}
