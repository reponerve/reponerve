package code

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	codemodels "github.com/reponerve/reponerve/internal/code/models"
	"github.com/reponerve/reponerve/internal/query/storage"
)

// Service is the authoritative Code Intelligence API.
type Service struct {
	entityReader   storage.CodeEntityReader
	relReader      storage.CodeRelationshipReader
	repoCodeReader storage.RepositoryCodeRelationshipReader
}

// NewService creates a Code Intelligence service.
func NewService(
	entityReader storage.CodeEntityReader,
	relReader storage.CodeRelationshipReader,
	repoCodeReader storage.RepositoryCodeRelationshipReader,
) *Service {
	return &Service{
		entityReader:   entityReader,
		relReader:      relReader,
		repoCodeReader: repoCodeReader,
	}
}

// ResolveFile resolves a repository file path to code explanation context.
func (s *Service) ResolveFile(ctx context.Context, repositoryID, filePath string) (*codemodels.CodeExplanationContext, error) {
	if repositoryID == "" {
		return nil, fmt.Errorf("repository ID cannot be empty")
	}
	normalized := normalizeFilePath(filePath)
	if normalized == "" {
		return nil, fmt.Errorf("file path cannot be empty")
	}

	entities, err := s.entityReader.ListByFilePath(ctx, repositoryID, normalized)
	if err != nil {
		return nil, err
	}
	if len(entities) == 0 {
		return nil, fmt.Errorf("file not indexed: %s", normalized)
	}

	ctxOut := &codemodels.CodeExplanationContext{
		Subject:  normalized,
		Evidence: []codemodels.EvidenceItem{{Source: "code_index", Detail: normalized}},
	}
	groupEntities(ctxOut, entities)

	var fileEntity *codemodels.CodeEntity
	for _, e := range entities {
		if e.EntityType == codemodels.EntityTypeFile {
			fileEntity = e
			break
		}
	}
	if fileEntity != nil {
		outbound, err := s.relReader.ListByFromEntity(ctx, fileEntity.ID)
		if err != nil {
			return nil, err
		}
		ctxOut.Dependencies = CollectSymbolDependencies(fileEntity.ID, outbound)
		if err := s.attachRelatedEntities(ctx, ctxOut, outbound); err != nil {
			return nil, err
		}
	}

	return ctxOut, nil
}

// ResolveSymbol resolves a qualified or short symbol name to code explanation context.
func (s *Service) ResolveSymbol(ctx context.Context, repositoryID, symbol string) (*codemodels.CodeExplanationContext, error) {
	if repositoryID == "" {
		return nil, fmt.Errorf("repository ID cannot be empty")
	}
	symbol = strings.TrimSpace(symbol)
	if symbol == "" {
		return nil, fmt.Errorf("symbol cannot be empty")
	}

	matches, err := s.resolveSymbolEntities(ctx, repositoryID, symbol)
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("symbol not found: %s", symbol)
	}

	root := matches[0]
	ctxOut := &codemodels.CodeExplanationContext{
		Subject:  root.QualifiedName,
		Evidence: []codemodels.EvidenceItem{{Source: "code_index", Detail: root.QualifiedName}},
	}
	groupEntities(ctxOut, matches)

	fileEntities, err := s.entityReader.ListByFilePath(ctx, repositoryID, root.FilePath)
	if err != nil {
		return nil, err
	}
	groupEntities(ctxOut, fileEntities)

	outbound, err := s.relReader.ListByFromEntity(ctx, root.ID)
	if err != nil {
		return nil, err
	}
	ctxOut.Dependencies = CollectSymbolDependencies(root.ID, outbound)

	allRels, err := s.relReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, err
	}
	ctxOut.CallGraph = BuildCallGraphFromRelationships(root.ID, allRels, 10)

	if err := s.attachRelatedEntities(ctx, ctxOut, outbound); err != nil {
		return nil, err
	}

	return ctxOut, nil
}

// BuildCallGraph builds a call graph from a root code entity ID.
func (s *Service) BuildCallGraph(ctx context.Context, repositoryID, entityID string) (*codemodels.CallGraph, error) {
	entity, err := s.entityReader.GetByID(ctx, entityID)
	if err != nil {
		return nil, err
	}
	if entity.RepositoryID != repositoryID {
		return nil, fmt.Errorf("entity does not belong to repository %s", repositoryID)
	}
	rels, err := s.relReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, err
	}
	return BuildCallGraphFromRelationships(entity.ID, rels, 10), nil
}

// AnalyzeSymbolDependencies analyzes outbound dependencies for a symbol entity.
func (s *Service) AnalyzeSymbolDependencies(ctx context.Context, repositoryID, entityID string) (*codemodels.SymbolDependencyReport, error) {
	entity, err := s.entityReader.GetByID(ctx, entityID)
	if err != nil {
		return nil, err
	}
	if entity.RepositoryID != repositoryID {
		return nil, fmt.Errorf("entity does not belong to repository %s", repositoryID)
	}
	outbound, err := s.relReader.ListByFromEntity(ctx, entityID)
	if err != nil {
		return nil, err
	}
	return &codemodels.SymbolDependencyReport{
		RootEntity:   entity,
		Dependencies: CollectSymbolDependencies(entityID, outbound),
	}, nil
}

// ListRepositoryCodeLinks returns repository-code links for a repository.
func (s *Service) ListRepositoryCodeLinks(ctx context.Context, repositoryID string) ([]*codemodels.RepositoryCodeRelationship, error) {
	if repositoryID == "" {
		return nil, fmt.Errorf("repository ID cannot be empty")
	}
	return s.repoCodeReader.ListByRepository(ctx, repositoryID)
}

func (s *Service) resolveSymbolEntities(ctx context.Context, repositoryID, symbol string) ([]*codemodels.CodeEntity, error) {
	exact, err := s.entityReader.FindByQualifiedName(ctx, repositoryID, symbol)
	if err != nil {
		return nil, err
	}
	if len(exact) > 0 {
		return filterSymbolEntities(exact), nil
	}

	all, err := s.entityReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	var matches []*codemodels.CodeEntity
	lowerSymbol := strings.ToLower(symbol)
	for _, e := range all {
		if !isSymbolEntityType(e.EntityType) {
			continue
		}
		if strings.EqualFold(e.Name, symbol) || strings.EqualFold(e.QualifiedName, symbol) {
			matches = append(matches, e)
			continue
		}
		if strings.HasSuffix(strings.ToLower(e.QualifiedName), "."+lowerSymbol) {
			matches = append(matches, e)
		}
	}
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].QualifiedName != matches[j].QualifiedName {
			return matches[i].QualifiedName < matches[j].QualifiedName
		}
		return matches[i].EntityType < matches[j].EntityType
	})
	return matches, nil
}

func (s *Service) attachRelatedEntities(ctx context.Context, ctxOut *codemodels.CodeExplanationContext, outbound []*codemodels.CodeRelationship) error {
	seen := make(map[string]struct{})
	for _, rel := range outbound {
		if _, ok := seen[rel.ToEntityID]; ok {
			continue
		}
		seen[rel.ToEntityID] = struct{}{}
		entity, err := s.entityReader.GetByID(ctx, rel.ToEntityID)
		if err != nil {
			continue
		}
		groupEntities(ctxOut, []*codemodels.CodeEntity{entity})
	}
	return nil
}

func groupEntities(ctxOut *codemodels.CodeExplanationContext, entities []*codemodels.CodeEntity) {
	for _, e := range entities {
		switch e.EntityType {
		case codemodels.EntityTypeModule:
			ctxOut.Modules = appendUniqueEntity(ctxOut.Modules, e)
		case codemodels.EntityTypePackage:
			ctxOut.Packages = appendUniqueEntity(ctxOut.Packages, e)
		case codemodels.EntityTypeFile:
			ctxOut.Files = appendUniqueEntity(ctxOut.Files, e)
		case codemodels.EntityTypeStruct:
			ctxOut.Structs = appendUniqueEntity(ctxOut.Structs, e)
		case codemodels.EntityTypeInterface:
			ctxOut.Interfaces = appendUniqueEntity(ctxOut.Interfaces, e)
		case codemodels.EntityTypeTypeAlias:
			ctxOut.TypeAliases = appendUniqueEntity(ctxOut.TypeAliases, e)
		case codemodels.EntityTypeFunction:
			ctxOut.Functions = appendUniqueEntity(ctxOut.Functions, e)
		case codemodels.EntityTypeMethod:
			ctxOut.Methods = appendUniqueEntity(ctxOut.Methods, e)
		case codemodels.EntityTypeEndpoint:
			ctxOut.Endpoints = appendUniqueEntity(ctxOut.Endpoints, e)
		}
	}
}

func appendUniqueEntity(list []*codemodels.CodeEntity, e *codemodels.CodeEntity) []*codemodels.CodeEntity {
	for _, existing := range list {
		if existing.ID == e.ID {
			return list
		}
	}
	return append(list, e)
}

func filterSymbolEntities(entities []*codemodels.CodeEntity) []*codemodels.CodeEntity {
	var out []*codemodels.CodeEntity
	for _, e := range entities {
		if isSymbolEntityType(e.EntityType) {
			out = append(out, e)
		}
	}
	return out
}

func isSymbolEntityType(entityType string) bool {
	switch entityType {
	case codemodels.EntityTypeStruct, codemodels.EntityTypeInterface, codemodels.EntityTypeTypeAlias,
		codemodels.EntityTypeFunction, codemodels.EntityTypeMethod, codemodels.EntityTypeEndpoint:
		return true
	default:
		return false
	}
}

func normalizeFilePath(path string) string {
	path = strings.TrimSpace(path)
	path = filepath.ToSlash(path)
	path = strings.TrimPrefix(path, "./")
	return path
}
