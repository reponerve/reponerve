package linker

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/reponerve/reponerve/internal/code"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
	"github.com/reponerve/reponerve/internal/query/storage"
	platformstorage "github.com/reponerve/reponerve/internal/storage"
	models "github.com/reponerve/reponerve/pkg/models"
)

const (
	repoEntityDecision = "DECISION"
	repoEntityFact     = "FACT"
	repoEntityEvent    = "EVENT"

	relDecisionReferencesCode = "DECISION_REFERENCES_CODE"
	relFactReferencesCode     = "FACT_REFERENCES_CODE"
	relEventReferencesCode    = "EVENT_REFERENCES_CODE"
)

type linkEvidence struct {
	Source             string `json:"source"`
	Match              string `json:"match"`
	Field              string `json:"field"`
	RepositoryEntityID string `json:"repository_entity_id"`
}

// Linker creates deterministic repository-code relationships.
type Linker struct {
	eventReader      storage.EventReader
	decisionReader   storage.DecisionReader
	factReader       storage.FactReader
	sourceReader     storage.SourceReader
	codeEntityReader storage.CodeEntityReader
	repoCodeStore    platformstorage.RepositoryCodeRelationshipStore
	stateStore       platformstorage.CodeIndexStateStore
}

// New creates a repository-code linker.
func New(
	eventReader storage.EventReader,
	decisionReader storage.DecisionReader,
	factReader storage.FactReader,
	sourceReader storage.SourceReader,
	codeEntityReader storage.CodeEntityReader,
	repoCodeStore platformstorage.RepositoryCodeRelationshipStore,
	stateStore platformstorage.CodeIndexStateStore,
) *Linker {
	return &Linker{
		eventReader:      eventReader,
		decisionReader:   decisionReader,
		factReader:       factReader,
		sourceReader:     sourceReader,
		codeEntityReader: codeEntityReader,
		repoCodeStore:    repoCodeStore,
		stateStore:       stateStore,
	}
}

// Link rebuilds repository-code relationships for a repository.
func (l *Linker) Link(ctx context.Context, repositoryID string) error {
	entities, err := l.codeEntityReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return err
	}
	if len(entities) == 0 {
		return nil
	}

	index := buildCodeIndex(entities)
	now := time.Now().UTC()

	var links []*codemodels.RepositoryCodeRelationship
	seen := make(map[string]struct{})

	decisions, err := l.decisionReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return fmt.Errorf("list decisions: %w", err)
	}
	sourceContent := l.loadADRContent(ctx, repositoryID)
	for _, d := range decisions {
		texts := decisionTexts(d, sourceContent[d.SourceID])
		links = append(links, l.matchTexts(repositoryID, repoEntityDecision, d.ID, relDecisionReferencesCode, texts, index, now, seen)...)
	}

	facts, err := l.factReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return fmt.Errorf("list facts: %w", err)
	}
	for _, f := range facts {
		texts := factTexts(f)
		links = append(links, l.matchTexts(repositoryID, repoEntityFact, f.ID, relFactReferencesCode, texts, index, now, seen)...)
	}

	events, err := l.eventReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return fmt.Errorf("list events: %w", err)
	}
	for _, ev := range events {
		texts := eventTexts(ev)
		links = append(links, l.matchTexts(repositoryID, repoEntityEvent, ev.ID, relEventReferencesCode, texts, index, now, seen)...)
	}

	sortRepositoryCodeLinks(links)

	if err := l.repoCodeStore.DeleteByRepository(ctx, repositoryID); err != nil {
		return err
	}
	for _, link := range links {
		if err := l.repoCodeStore.UpsertRepositoryCodeRelationship(ctx, link); err != nil {
			return fmt.Errorf("upsert repository-code link: %w", err)
		}
	}

	return l.stateStore.UpdateLinkCount(ctx, repositoryID, len(links))
}

type indexedText struct {
	Text  string
	Field string
}

type codeIndex struct {
	files       map[string]*codemodels.CodeEntity
	packages    map[string]*codemodels.CodeEntity
	byQualified map[string]*codemodels.CodeEntity
	symbolNames []string
}

func buildCodeIndex(entities []*codemodels.CodeEntity) *codeIndex {
	idx := &codeIndex{
		files:       make(map[string]*codemodels.CodeEntity),
		packages:    make(map[string]*codemodels.CodeEntity),
		byQualified: make(map[string]*codemodels.CodeEntity),
	}
	for _, e := range entities {
		idx.byQualified[e.QualifiedName] = e
		switch e.EntityType {
		case codemodels.EntityTypeFile:
			idx.files[e.FilePath] = e
		case codemodels.EntityTypePackage:
			key := e.PackagePath
			if key == "" {
				key = e.QualifiedName
			}
			idx.packages[key] = e
		default:
			if e.EntityType != codemodels.EntityTypeModule {
				idx.symbolNames = append(idx.symbolNames, e.QualifiedName)
			}
		}
	}
	sort.Strings(idx.symbolNames)
	return idx
}

func (l *Linker) loadADRContent(ctx context.Context, repositoryID string) map[string]string {
	out := make(map[string]string)
	sources, err := l.sourceReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return out
	}
	for _, src := range sources {
		if src.SourceType != "adr" || src.MetadataJSON == "" {
			continue
		}
		var meta struct {
			Content string `json:"content"`
		}
		if err := json.Unmarshal([]byte(src.MetadataJSON), &meta); err == nil && meta.Content != "" {
			out[src.ID] = meta.Content
		}
	}
	return out
}

func decisionTexts(d *memorymodels.Decision, adrContent string) []indexedText {
	var texts []indexedText
	if d.Title != "" {
		texts = append(texts, indexedText{Text: d.Title, Field: "title"})
	}
	if adrContent != "" {
		texts = append(texts, indexedText{Text: adrContent, Field: "adr_content"})
	}
	return texts
}

func factTexts(f *memorymodels.Fact) []indexedText {
	return []indexedText{
		{Text: strings.Join([]string{f.Subject, f.Predicate, f.Object}, " "), Field: "fact"},
	}
}

func eventTexts(ev *models.Event) []indexedText {
	var texts []indexedText
	if ev.Title != "" {
		texts = append(texts, indexedText{Text: ev.Title, Field: "title"})
	}
	if ev.Description != "" {
		texts = append(texts, indexedText{Text: ev.Description, Field: "description"})
	}
	return texts
}

func (l *Linker) matchTexts(
	repositoryID, repoEntityType, repoEntityID, relType string,
	texts []indexedText,
	index *codeIndex,
	indexedAt time.Time,
	seen map[string]struct{},
) []*codemodels.RepositoryCodeRelationship {
	var links []*codemodels.RepositoryCodeRelationship
	for _, text := range texts {
		for _, match := range extractGoFilePaths(text.Text, text.Field) {
			entity := index.files[match.Value]
			if entity == nil {
				continue
			}
			if link := l.buildLink(repositoryID, repoEntityType, repoEntityID, relType, entity, "path_reference", match, indexedAt, seen); link != nil {
				links = append(links, link)
			}
		}
		for _, match := range extractPackagePaths(text.Text, text.Field) {
			entity := index.packages[match.Value]
			if entity == nil {
				continue
			}
			if link := l.buildLink(repositoryID, repoEntityType, repoEntityID, relType, entity, "package_reference", match, indexedAt, seen); link != nil {
				links = append(links, link)
			}
		}
		for _, match := range extractQualifiedSymbols(text.Text, text.Field, index.symbolNames) {
			entity := index.byQualified[match.Value]
			if entity == nil {
				continue
			}
			if link := l.buildLink(repositoryID, repoEntityType, repoEntityID, relType, entity, "symbol_reference", match, indexedAt, seen); link != nil {
				links = append(links, link)
			}
		}
	}
	return links
}

func (l *Linker) buildLink(
	repositoryID, repoEntityType, repoEntityID, relType string,
	codeEntity *codemodels.CodeEntity,
	source string,
	match textMatch,
	indexedAt time.Time,
	seen map[string]struct{},
) *codemodels.RepositoryCodeRelationship {
	key := repoEntityID + ":" + codeEntity.ID + ":" + relType
	if _, exists := seen[key]; exists {
		return nil
	}
	seen[key] = struct{}{}

	evidence, _ := json.Marshal(linkEvidence{
		Source:             source,
		Match:              match.Value,
		Field:              match.Field,
		RepositoryEntityID: repoEntityID,
	})

	return &codemodels.RepositoryCodeRelationship{
		ID:                   code.RepositoryCodeLinkID(repositoryID, relType, repoEntityID, codeEntity.ID),
		RepositoryID:         repositoryID,
		RepositoryEntityID:   repoEntityID,
		RepositoryEntityType: repoEntityType,
		CodeEntityID:         codeEntity.ID,
		CodeEntityType:       codeEntity.EntityType,
		RelationshipType:     relType,
		EvidenceJSON:         string(evidence),
		IndexedAt:            indexedAt,
	}
}

func sortRepositoryCodeLinks(links []*codemodels.RepositoryCodeRelationship) {
	sort.Slice(links, func(i, j int) bool {
		if links[i].RelationshipType != links[j].RelationshipType {
			return links[i].RelationshipType < links[j].RelationshipType
		}
		if links[i].RepositoryEntityType != links[j].RepositoryEntityType {
			return links[i].RepositoryEntityType < links[j].RepositoryEntityType
		}
		if links[i].RepositoryEntityID != links[j].RepositoryEntityID {
			return links[i].RepositoryEntityID < links[j].RepositoryEntityID
		}
		if links[i].CodeEntityType != links[j].CodeEntityType {
			return links[i].CodeEntityType < links[j].CodeEntityType
		}
		return links[i].CodeEntityID < links[j].CodeEntityID
	})
}
