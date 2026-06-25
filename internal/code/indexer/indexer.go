package indexer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/reponerve/reponerve/internal/code/lang"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
	"github.com/reponerve/reponerve/internal/storage"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
)

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

	hasGoMod := fileExists(filepath.Join(repositoryPath, "go.mod"))
	hasGoWork := fileExists(filepath.Join(repositoryPath, "go.work"))
	allFiles, err := listAllIndexableFiles(repositoryPath)
	if err != nil {
		return fmt.Errorf("source file discovery failed: %w", err)
	}
	if len(allFiles) == 0 {
		return nil
	}
	if !hasGoMod && !hasGoWork && !hasNonGoFiles(allFiles) {
		return nil
	}

	skip, err := shouldSkipIndexing(ctx, idx.stateStore, repositoryID, repositoryPath)
	if err != nil {
		return fmt.Errorf("incremental index check: %w", err)
	}
	if skip {
		currentFiles, ferr := listAllIndexableFiles(repositoryPath)
		if ferr != nil {
			return fmt.Errorf("source file discovery failed: %w", ferr)
		}
		state, _ := idx.stateStore.GetByRepository(ctx, repositoryID)
		if state != nil && state.FileCount != len(currentFiles) {
			skip = false
		}
	}
	if skip {
		return nil
	}

	now := time.Now().UTC()
	defaultModule := filepath.Base(repositoryPath)
	if defaultModule == "" || defaultModule == "." {
		defaultModule = "."
	}
	b := newBuilder(repositoryID, defaultModule, repositoryPath, now)

	if hasGoMod || hasGoWork {
		if err := idx.indexGoModules(b, repositoryPath); err != nil {
			return err
		}
	}

	multiLangFiles, err := listMultiLangFiles(repositoryPath)
	if err != nil {
		return fmt.Errorf("multi-language file discovery failed: %w", err)
	}
	for _, filePath := range multiLangFiles {
		language := lang.Detect(filePath)
		if err := b.parseTreeSitterFile(filePath, language); err != nil {
			return err
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

// IndexModules rebuilds code intelligence for specific Go module paths in a monorepo.
func (idx *Indexer) IndexModules(ctx context.Context, repositoryID, repositoryPath string, modulePaths []string) error {
	if len(modulePaths) == 0 {
		return idx.Index(ctx, repositoryID, repositoryPath)
	}

	repositoryPath = filepath.Clean(repositoryPath)
	moduleRoots, err := discoverModuleRoots(repositoryPath)
	if err != nil {
		return fmt.Errorf("module discovery failed: %w", err)
	}
	filtered, err := FilterModuleRoots(moduleRoots, modulePaths)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	defaultModule := filepath.Base(repositoryPath)
	if defaultModule == "" || defaultModule == "." {
		defaultModule = "."
	}
	b := newBuilder(repositoryID, defaultModule, repositoryPath, now)
	b.modulePath = filtered[0].modulePath

	if err := idx.indexGoModulesFiltered(b, repositoryPath, filtered); err != nil {
		return err
	}

	multiLangFiles, err := listMultiLangFiles(repositoryPath)
	if err != nil {
		return fmt.Errorf("multi-language file discovery failed: %w", err)
	}
	for _, filePath := range multiLangFiles {
		if !fileUnderModuleRoots(repositoryPath, filePath, filtered) {
			continue
		}
		language := lang.Detect(filePath)
		if err := b.parseTreeSitterFile(filePath, language); err != nil {
			return err
		}
	}

	sortEntities(b.entities)
	sortRelationships(b.rels)

	scopedPaths := make([]string, len(filtered))
	for i, root := range filtered {
		scopedPaths[i] = root.modulePath
	}

	if idx.db != nil {
		if err := idx.db.MergeCodeIndexModules(ctx, repositoryID, scopedPaths, b.entities, b.rels); err != nil {
			return fmt.Errorf("merge scoped code index: %w", err)
		}
	} else {
		return fmt.Errorf("scoped code index requires database-backed indexer")
	}

	state, err := idx.loadCodeIndexState(ctx, repositoryID)
	if err != nil {
		return err
	}
	state.LastIndexedAt = now
	if err := idx.stateStore.UpsertCodeIndexState(ctx, state); err != nil {
		return fmt.Errorf("update code index state: %w", err)
	}
	return nil
}

func (idx *Indexer) loadCodeIndexState(ctx context.Context, repositoryID string) (*codemodels.CodeIndexState, error) {
	state, err := idx.stateStore.GetByRepository(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("load code index state: %w", err)
	}
	if state == nil {
		state = &codemodels.CodeIndexState{RepositoryID: repositoryID}
	}
	if idx.db == nil {
		return state, nil
	}
	var moduleCount, fileCount, entityCount, relCount int
	row := idx.db.QueryRowContext(ctx, `
		SELECT
			COUNT(DISTINCT CASE WHEN entity_type = 'module' THEN module_path END),
			COUNT(DISTINCT CASE WHEN entity_type = 'file' THEN file_path END),
			COUNT(*),
			(SELECT COUNT(*) FROM code_relationships WHERE repository_id = ?)
		FROM code_entities WHERE repository_id = ?`, repositoryID, repositoryID)
	if err := row.Scan(&moduleCount, &fileCount, &entityCount, &relCount); err != nil {
		return nil, fmt.Errorf("count code index: %w", err)
	}
	state.ModuleCount = moduleCount
	state.FileCount = fileCount
	state.EntityCount = entityCount
	state.RelationshipCount = relCount
	return state, nil
}

func (idx *Indexer) indexGoModulesFiltered(b *builder, repositoryPath string, moduleRoots []moduleRoot) error {
	if len(moduleRoots) == 0 {
		return nil
	}
	for _, root := range moduleRoots {
		files, err := listGoFiles(root.path)
		if err != nil {
			return fmt.Errorf("go file discovery failed: %w", err)
		}

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
	return nil
}

func (idx *Indexer) indexGoModules(b *builder, repositoryPath string) error {
	moduleRoots, err := discoverModuleRoots(repositoryPath)
	if err != nil {
		return fmt.Errorf("module discovery failed: %w", err)
	}
	if len(moduleRoots) == 0 {
		return nil
	}
	b.modulePath = moduleRoots[0].modulePath
	return idx.indexGoModulesFiltered(b, repositoryPath, moduleRoots)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func hasNonGoFiles(files []string) bool {
	for _, f := range files {
		if lang.Detect(f) != lang.Go {
			return true
		}
	}
	return false
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
