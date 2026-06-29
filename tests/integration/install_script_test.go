package integration

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestInstallScriptVerifiesDownloadedArchive(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("install.sh is for POSIX shells")
	}
	requireCommand(t, "sh")
	requireCommand(t, "tar")

	tmp := t.TempDir()
	archivePath := filepath.Join(tmp, "release.tar.gz")
	digest := writeReponerveArchive(t, archivePath)
	fakeBin := writeInstallScriptFakes(t, tmp)
	scriptPath := filepath.Join("..", "..", "scripts", "install.sh")
	archiveName := "reponerve_9.9.9_linux_amd64.tar.gz"

	t.Run("accepts matching checksum", func(t *testing.T) {
		checksumsPath := filepath.Join(t.TempDir(), "checksums.txt")
		writeChecksumFile(t, checksumsPath, digest, archiveName)

		installDir := t.TempDir()
		output, err := runInstallScript(t, scriptPath, fakeBin, archivePath, checksumsPath, installDir)
		if err != nil {
			t.Fatalf("install.sh failed with matching checksum: %v\n%s", err, output)
		}
		if !strings.Contains(output, "Checksum verified") {
			t.Fatalf("expected checksum success output, got:\n%s", output)
		}
		installed, err := os.ReadFile(filepath.Join(installDir, "reponerve"))
		if err != nil {
			t.Fatalf("expected installed binary: %v\n%s", err, output)
		}
		if string(installed) != "fake reponerve\n" {
			t.Fatalf("unexpected installed binary content %q", installed)
		}
	})

	t.Run("rejects mismatched checksum", func(t *testing.T) {
		checksumsPath := filepath.Join(t.TempDir(), "checksums.txt")
		writeChecksumFile(t, checksumsPath, strings.Repeat("0", 64), archiveName)

		installDir := t.TempDir()
		output, err := runInstallScript(t, scriptPath, fakeBin, archivePath, checksumsPath, installDir)
		if err == nil {
			t.Fatalf("install.sh succeeded with mismatched checksum:\n%s", output)
		}
		if _, statErr := os.Stat(filepath.Join(installDir, "reponerve")); !os.IsNotExist(statErr) {
			t.Fatalf("expected no installed binary after checksum mismatch, statErr=%v\n%s", statErr, output)
		}
	})
}

func requireCommand(t *testing.T, name string) {
	t.Helper()
	if _, err := exec.LookPath(name); err != nil {
		t.Skipf("%s not available: %v", name, err)
	}
}

func writeReponerveArchive(t *testing.T, path string) string {
	t.Helper()

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create archive: %v", err)
	}
	gzipWriter := gzip.NewWriter(file)
	tarWriter := tar.NewWriter(gzipWriter)
	content := []byte("fake reponerve\n")
	header := &tar.Header{
		Name: "reponerve",
		Mode: 0o755,
		Size: int64(len(content)),
	}
	if err := tarWriter.WriteHeader(header); err != nil {
		t.Fatalf("write tar header: %v", err)
	}
	if _, err := tarWriter.Write(content); err != nil {
		t.Fatalf("write tar content: %v", err)
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatalf("close tar writer: %v", err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatalf("close gzip writer: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close archive: %v", err)
	}

	archive, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read archive: %v", err)
	}
	sum := sha256.Sum256(archive)
	return fmt.Sprintf("%x", sum[:])
}

func writeChecksumFile(t *testing.T, path, digest, archiveName string) {
	t.Helper()
	content := fmt.Sprintf("%s  %s\n", digest, archiveName)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write checksums: %v", err)
	}
}

func writeInstallScriptFakes(t *testing.T, dir string) string {
	t.Helper()

	fakeBin := filepath.Join(dir, "bin")
	if err := os.MkdirAll(fakeBin, 0o755); err != nil {
		t.Fatalf("mkdir fake bin: %v", err)
	}
	fakeCurl := `#!/usr/bin/env sh
set -eu
url=""
out=""
while [ "$#" -gt 0 ]; do
	case "$1" in
		-o)
			out="$2"
			shift 2
			;;
		-*)
			shift
			;;
		*)
			url="$1"
			shift
			;;
	esac
done
[ -n "$out" ] || exit 2
case "$url" in
	*checksums.txt) cp "$FAKE_CHECKSUMS" "$out" ;;
	*.tar.gz) cp "$FAKE_ARCHIVE" "$out" ;;
	*) exit 3 ;;
esac
`
	if err := os.WriteFile(filepath.Join(fakeBin, "curl"), []byte(fakeCurl), 0o755); err != nil {
		t.Fatalf("write fake curl: %v", err)
	}
	fakeUname := `#!/usr/bin/env sh
set -eu
case "${1:-}" in
	-s) printf 'Linux\n' ;;
	-m) printf 'x86_64\n' ;;
	*) exit 2 ;;
esac
`
	if err := os.WriteFile(filepath.Join(fakeBin, "uname"), []byte(fakeUname), 0o755); err != nil {
		t.Fatalf("write fake uname: %v", err)
	}
	return fakeBin
}

func runInstallScript(t *testing.T, scriptPath, fakeBin, archivePath, checksumsPath, installDir string) (string, error) {
	t.Helper()

	cmd := exec.Command("sh", scriptPath)
	cmd.Env = append(os.Environ(),
		"PATH="+fakeBin+string(os.PathListSeparator)+os.Getenv("PATH"),
		"REPONERVE_REPO=example/reponerve",
		"REPONERVE_VERSION=v9.9.9",
		"REPONERVE_VERIFY=1",
		"REPONERVE_INSTALL_DIR="+installDir,
		"FAKE_ARCHIVE="+archivePath,
		"FAKE_CHECKSUMS="+checksumsPath,
	)
	output, err := cmd.CombinedOutput()
	return string(output), err
}
