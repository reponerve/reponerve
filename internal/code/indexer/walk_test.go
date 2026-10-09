package indexer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWalkSkipsTestdataAndFixtures(t *testing.T) {
	tempDir := t.TempDir()

	// Create root source file
	if err := os.WriteFile(filepath.Join(tempDir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create testdata subdirectory with files
	testdataDir := filepath.Join(tempDir, "testdata")
	if err := os.MkdirAll(testdataDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(testdataDir, "fixture.go"), []byte("package testdata\nfunc Fake() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create fixtures subdirectory with files
	fixturesDir := filepath.Join(tempDir, "fixtures")
	if err := os.MkdirAll(fixturesDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fixturesDir, "fixture.py"), []byte("def fake(): pass\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create real nested package
	pkgDir := filepath.Join(tempDir, "pkg", "service")
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "svc.go"), []byte("package service\nfunc Svc() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	files, err := listAllIndexableFiles(tempDir)
	if err != nil {
		t.Fatalf("listAllIndexableFiles failed: %v", err)
	}

	for _, file := range files {
		if filepath.HasPrefix(file, "testdata") {
			t.Errorf("expected testdata to be skipped, but found %q", file)
		}
		if filepath.HasPrefix(file, "fixtures") {
			t.Errorf("expected fixtures to be skipped, but found %q", file)
		}
	}

	foundMain := false
	foundSvc := false
	for _, file := range files {
		if file == "main.go" {
			foundMain = true
		}
		if file == filepath.ToSlash(filepath.Join("pkg", "service", "svc.go")) {
			foundSvc = true
		}
	}
	if !foundMain {
		t.Errorf("expected main.go in indexable files")
	}
	if !foundSvc {
		t.Errorf("expected pkg/service/svc.go in indexable files")
	}
}
