package indexer

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	codemodels "github.com/reponerve/reponerve/internal/code/models"
	"github.com/reponerve/reponerve/internal/storage"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
)

// Indexer performs deterministic Go code indexing for a repository.
type Indexer struct {
	db            *sqlite.Database
	entityStore   storage.CodeEntityStore
	relStore      storage.CodeRelationshipStore
	repoCodeStore storage.RepositoryCodeRelationshipStore
	stateStore    storage.CodeIndexStateStore
}

// New creates a code indexer backed by storage stores.
func New(
	db *sqlite.Database,
	entityStore storage.CodeEntityStore,
	relStore storage.CodeRelationshipStore,
	repoCodeStore storage.RepositoryCodeRelationshipStore,
	stateStore storage.CodeIndexStateStore,
) *Indexer {
	return &Indexer{
		db:            db,
		entityStore:   entityStore,
		relStore:      relStore,
		repoCodeStore: repoCodeStore,
		stateStore:    stateStore,
	}
}

// Index rebuilds code intelligence for the repository path.
func (idx *Indexer) Index(ctx context.Context, repositoryID, repositoryPath string) error {
	repositoryPath = filepath.Clean(repositoryPath)
	if _, err := os.Stat(filepath.Join(repositoryPath, "go.mod")); err != nil {
		if _, workErr := os.Stat(filepath.Join(repositoryPath, "go.work")); workErr != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil
			}
			return fmt.Errorf("stat go.mod: %w", err)
		}
	}

	skip, err := shouldSkipIndexing(ctx, idx.stateStore, repositoryID, repositoryPath)
	if err != nil {
		return fmt.Errorf("incremental index check: %w", err)
	}
	if skip {
		currentFiles, ferr := listGoFiles(repositoryPath)
		if ferr != nil {
			return fmt.Errorf("go file discovery failed: %w", ferr)
		}
		state, _ := idx.stateStore.GetByRepository(ctx, repositoryID)
		if state != nil && state.FileCount != len(currentFiles) {
			skip = false
		}
	}
	if skip {
		return nil
	}

	moduleRoots, err := discoverModuleRoots(repositoryPath)
	if err != nil {
		return fmt.Errorf("module discovery failed: %w", err)
	}

	now := time.Now().UTC()
	b := newBuilder(repositoryID, moduleRoots[0].modulePath, repositoryPath, now)

	for _, root := range moduleRoots {
		files, err := listGoFiles(root.path)
		if err != nil {
			return fmt.Errorf("go file discovery failed: %w", err)
		}
		sort.Strings(files)

		moduleID := b.addModuleEntity(root.modulePath, root.goModFile)
		prefix := ""
		if root.path != repositoryPath {
			relRoot, err := filepath.Rel(repositoryPath, root.path)
			if err == nil && relRoot != "." {
				prefix = filepath.ToSlash(relRoot)
			}
		}
		for _, filePath := range files {
			if prefix != "" {
				filePath = prefix + "/" + filePath
			}
			if err := b.parseFile(filePath, moduleID); err != nil {
				return err
			}
		}
	}

	sortEntities(b.entities)
	sortRelationships(b.rels)

	if idx.db != nil {
		if err := idx.db.ReplaceCodeIndex(ctx, repositoryID, b.entities, b.rels); err != nil {
			return fmt.Errorf("replace code index: %w", err)
		}
	} else {
		if idx.repoCodeStore != nil {
			if err := idx.repoCodeStore.DeleteByRepository(ctx, repositoryID); err != nil {
				return err
			}
		}
		if err := idx.relStore.DeleteByRepository(ctx, repositoryID); err != nil {
			return err
		}
		if err := idx.entityStore.DeleteByRepository(ctx, repositoryID); err != nil {
			return err
		}
		for _, entity := range b.entities {
			if err := idx.entityStore.UpsertCodeEntity(ctx, entity); err != nil {
				return fmt.Errorf("upsert code entity %s: %w", entity.QualifiedName, err)
			}
		}
		for _, rel := range b.rels {
			if err := idx.relStore.UpsertCodeRelationship(ctx, rel); err != nil {
				return fmt.Errorf("upsert code relationship %s: %w", rel.RelationshipType, err)
			}
		}
	}

	moduleCount := 0
	fileCount := 0
	for _, e := range b.entities {
		switch e.EntityType {
		case codemodels.EntityTypeModule:
			moduleCount++
		case codemodels.EntityTypeFile:
			fileCount++
		}
	}

	state := &codemodels.CodeIndexState{
		RepositoryID:      repositoryID,
		LastIndexedAt:     now,
		ModuleCount:       moduleCount,
		FileCount:         fileCount,
		EntityCount:       len(b.entities),
		RelationshipCount: len(b.rels),
		LinkCount:         0, // updated by repository-code linker after scan
	}
	if err := idx.stateStore.UpsertCodeIndexState(ctx, state); err != nil {
		return fmt.Errorf("update code index state: %w", err)
	}

	return nil
}

func sortEntities(entities []*codemodels.CodeEntity) {
	sort.Slice(entities, func(i, j int) bool {
		if entities[i].EntityType != entities[j].EntityType {
			return entities[i].EntityType < entities[j].EntityType
		}
		if entities[i].ModulePath != entities[j].ModulePath {
			return entities[i].ModulePath < entities[j].ModulePath
		}
		if entities[i].FilePath != entities[j].FilePath {
			return entities[i].FilePath < entities[j].FilePath
		}
		if entities[i].StartLine != entities[j].StartLine {
			return entities[i].StartLine < entities[j].StartLine
		}
		if entities[i].QualifiedName != entities[j].QualifiedName {
			return entities[i].QualifiedName < entities[j].QualifiedName
		}
		return entities[i].ID < entities[j].ID
	})
}

func sortRelationships(rels []*codemodels.CodeRelationship) {
	sort.Slice(rels, func(i, j int) bool {
		if rels[i].RelationshipType != rels[j].RelationshipType {
			return rels[i].RelationshipType < rels[j].RelationshipType
		}
		if rels[i].FromEntityID != rels[j].FromEntityID {
			return rels[i].FromEntityID < rels[j].FromEntityID
		}
		return rels[i].ToEntityID < rels[j].ToEntityID
	})
}
