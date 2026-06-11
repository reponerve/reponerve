package code_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/reponerve/reponerve/internal/code"
	"github.com/reponerve/reponerve/internal/code/indexer"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
	"github.com/reponerve/reponerve/internal/query/storage"
	"github.com/reponerve/reponerve/internal/storage/migrations"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
)

func setupIndexedSampleRepo(t *testing.T) (*sqlite.Database, string, *code.Service) {
	t.Helper()

	repoPath := filepath.Join("indexer", "testdata", "samplemodule")
	absRepo, err := filepath.Abs(repoPath)
	if err != nil {
		t.Fatalf("abs path: %v", err)
	}

	tempDir, err := os.MkdirTemp("", "reponerve-code-service-*")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(tempDir) })

	db, err := sqlite.Open(filepath.Join(tempDir, "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	if err := migrations.RunUp(db); err != nil {
		t.Fatalf("migrations: %v", err)
	}

	repoID := "repo_service"
	_, err = db.Exec(`INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at) VALUES (?, ?, ?, ?, datetime(), datetime())`,
		repoID, "sample", absRepo, "main")
	if err != nil {
		t.Fatalf("insert repository: %v", err)
	}

	entityStore := sqlite.NewSQLiteCodeEntityStore(db)
	relStore := sqlite.NewSQLiteCodeRelationshipStore(db)
	repoCodeStore := sqlite.NewSQLiteRepositoryCodeRelationshipStore(db)
	stateStore := sqlite.NewSQLiteCodeIndexStateStore(db)

	if err := indexer.New(entityStore, relStore, repoCodeStore, stateStore).Index(context.Background(), repoID, absRepo); err != nil {
		t.Fatalf("index failed: %v", err)
	}

	svc := code.NewService(
		storage.NewSQLiteCodeEntityReader(db),
		storage.NewSQLiteCodeRelationshipReader(db),
		storage.NewSQLiteRepositoryCodeRelationshipReader(db),
	)
	return db, repoID, svc
}

func TestService_ResolveFile(t *testing.T) {
	_, repoID, svc := setupIndexedSampleRepo(t)

	ctx, err := svc.ResolveFile(context.Background(), repoID, "internal/auth/service.go")
	if err != nil {
		t.Fatalf("ResolveFile failed: %v", err)
	}
	if ctx.Subject != "internal/auth/service.go" {
		t.Fatalf("unexpected subject: %q", ctx.Subject)
	}
	if len(ctx.Files) != 1 {
		t.Fatalf("expected 1 file entity, got %d", len(ctx.Files))
	}
	if len(ctx.Structs) == 0 || len(ctx.Methods) == 0 {
		t.Fatalf("expected struct and method entities in file context")
	}
}

func TestService_ResolveSymbol(t *testing.T) {
	_, repoID, svc := setupIndexedSampleRepo(t)

	ctx, err := svc.ResolveSymbol(context.Background(), repoID, "internal/auth.Service.Login")
	if err != nil {
		t.Fatalf("ResolveSymbol failed: %v", err)
	}
	if len(ctx.Methods) == 0 {
		t.Fatalf("expected method in symbol context")
	}
	if ctx.CallGraph == nil {
		t.Fatalf("expected call graph object")
	}
}

func TestService_AnalyzeSymbolDependencies(t *testing.T) {
	_, repoID, svc := setupIndexedSampleRepo(t)

	fileCtx, err := svc.ResolveFile(context.Background(), repoID, "internal/auth/service.go")
	if err != nil {
		t.Fatalf("ResolveFile failed: %v", err)
	}
	if len(fileCtx.Methods) == 0 {
		t.Fatal("expected methods")
	}

	report, err := svc.AnalyzeSymbolDependencies(context.Background(), repoID, fileCtx.Methods[0].ID)
	if err != nil {
		t.Fatalf("AnalyzeSymbolDependencies failed: %v", err)
	}
	if report.RootEntity == nil {
		t.Fatal("expected root entity")
	}
	_ = codemodels.EntityTypeMethod
}

func TestService_BuildCallGraph(t *testing.T) {
	_, repoID, svc := setupIndexedSampleRepo(t)

	ctx, err := svc.ResolveSymbol(context.Background(), repoID, "NewService")
	if err != nil {
		t.Fatalf("ResolveSymbol failed: %v", err)
	}
	if len(ctx.Functions) == 0 {
		t.Fatalf("expected function NewService")
	}

	graph, err := svc.BuildCallGraph(context.Background(), repoID, ctx.Functions[0].ID)
	if err != nil {
		t.Fatalf("BuildCallGraph failed: %v", err)
	}
	if graph.RootEntityID != ctx.Functions[0].ID {
		t.Fatalf("unexpected root entity")
	}
}
