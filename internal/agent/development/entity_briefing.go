package development

import (
	"context"
	"fmt"
	"sort"
	"strings"

	agentsearch "github.com/reponerve/reponerve/internal/agent/search"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
)

const maxEntityBriefings = 8
const maxBriefingRefs = 6

var packageLayerRoles = map[string]string{
	"cmd":      "CLI entry and command wiring",
	"pkg":      "public API surface",
	"internal": "core implementation",
	"api":      "HTTP/API layer",
	"web":      "web/UI layer",
}

func roleForEntity(entity *codemodels.CodeEntity) string {
	if hint := inferPackageLayer(entity.PackagePath); hint != "" {
		return fmt.Sprintf("%s %s — %s", entity.EntityType, entity.Name, hint)
	}
	return fmt.Sprintf("%s %s in package %s", entity.EntityType, entity.Name, entity.PackagePath)
}

func inferPackageLayer(packagePath string) string {
	if hint, ok := packageLayerRoles[packagePath]; ok {
		return hint
	}
	parts := strings.Split(strings.Trim(packagePath, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return ""
	}
	if layer, ok := packageLayerRoles[parts[0]]; ok {
		if len(parts) > 1 {
			return parts[1] + " (" + layer + ")"
		}
		return layer
	}
	if len(parts) >= 2 {
		return parts[len(parts)-1] + " package"
	}
	return ""
}

func (s *Service) buildEntityBriefings(
	ctx context.Context,
	repositoryID string,
	topic string,
	entities []*codemodels.CodeEntity,
	links []RepositoryCodeLinkRef,
) ([]EntityBriefing, error) {
	if len(entities) == 0 {
		return nil, nil
	}

	roots := pickBriefingRoots(entities)
	roots = preferTypeDefinitionsForTopic(roots, topic)
	if len(roots) == 0 {
		return nil, nil
	}

	allEntities, err := s.codeEntityReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	var briefings []EntityBriefing
	for _, root := range roots {
		if len(briefings) >= maxEntityBriefings {
			break
		}
		briefings = append(briefings, s.briefEntity(ctx, repositoryID, root, allEntities, links))
	}
	return briefings, nil
}

func pickBriefingRoots(entities []*codemodels.CodeEntity) []*codemodels.CodeEntity {
	priority := map[string]int{
		codemodels.EntityTypeStruct:     1,
		codemodels.EntityTypeInterface:  2,
		codemodels.EntityTypeTypeAlias:  3,
		codemodels.EntityTypeFunction:   4,
		codemodels.EntityTypeMethod:     5,
		codemodels.EntityTypeFile:       6,
		codemodels.EntityTypePackage:    7,
		codemodels.EntityTypeModule:     8,
	}

	seen := make(map[string]struct{})
	var roots []*codemodels.CodeEntity
	for _, e := range entities {
		if !isBriefingRootType(e.EntityType) {
			continue
		}
		if _, ok := seen[e.ID]; ok {
			continue
		}
		seen[e.ID] = struct{}{}
		roots = append(roots, e)
	}

	sort.Slice(roots, func(i, j int) bool {
		pi, pj := priority[roots[i].EntityType], priority[roots[j].EntityType]
		if pi != pj {
			return pi < pj
		}
		return roots[i].QualifiedName < roots[j].QualifiedName
	})
	return roots
}

func isBriefingRootType(entityType string) bool {
	switch entityType {
	case codemodels.EntityTypeStruct, codemodels.EntityTypeInterface, codemodels.EntityTypeTypeAlias,
		codemodels.EntityTypeFunction, codemodels.EntityTypeMethod, codemodels.EntityTypeFile:
		return true
	default:
		return false
	}
}

func (s *Service) briefEntity(
	ctx context.Context,
	repositoryID string,
	entity *codemodels.CodeEntity,
	allEntities []*codemodels.CodeEntity,
	links []RepositoryCodeLinkRef,
) EntityBriefing {
	brief := EntityBriefing{
		QualifiedName: entity.QualifiedName,
		EntityType:      entity.EntityType,
		Layer:           entity.PackagePath,
		Role:            roleForEntity(entity),
		DefinedIn:       formatDefinedIn(entity),
		Signature:       strings.TrimSpace(entity.Signature),
		Fields:          fieldsFromSignature(entity),
	}

	if entity.EntityType == codemodels.EntityTypeStruct {
		brief.Members = s.structMembers(entity, allEntities)
	}

	if s.relReader != nil {
		inbound, _ := s.relReader.ListByToEntity(ctx, entity.ID)
		outbound, _ := s.relReader.ListByFromEntity(ctx, entity.ID)
		brief.Producers = refsFromRelationships(ctx, s.codeEntityReader, inbound, true, maxBriefingRefs)
		brief.Consumers = refsFromRelationships(ctx, s.codeEntityReader, outbound, false, maxBriefingRefs)
	}

	brief.RelatedDecisions = relatedDecisionsForEntity(entity, links)
	return brief
}

func fieldsFromSignature(entity *codemodels.CodeEntity) []string {
	sig := strings.TrimSpace(entity.Signature)
	if sig == "" {
		return nil
	}
	switch entity.EntityType {
	case codemodels.EntityTypeStruct, codemodels.EntityTypeInterface:
		parts := strings.Split(sig, "; ")
		fields := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				fields = append(fields, p)
			}
		}
		return fields
	default:
		return nil
	}
}

func formatDefinedIn(entity *codemodels.CodeEntity) string {
	if entity.FilePath == "" {
		return ""
	}
	if entity.StartLine > 0 && entity.EndLine > 0 {
		return fmt.Sprintf("%s:%d-%d", entity.FilePath, entity.StartLine, entity.EndLine)
	}
	return entity.FilePath
}

func (s *Service) structMembers(structEntity *codemodels.CodeEntity, allEntities []*codemodels.CodeEntity) []EntityRef {
	prefix := structEntity.PackagePath + "." + structEntity.Name + "."
	var members []EntityRef
	for _, e := range allEntities {
		if e.EntityType != codemodels.EntityTypeMethod {
			continue
		}
		if !strings.HasPrefix(e.QualifiedName, prefix) {
			continue
		}
		members = append(members, codeEntityRef(e))
		if len(members) >= maxBriefingRefs {
			break
		}
	}
	sortEntityRefs(members)
	return members
}

func refsFromRelationships(
	ctx context.Context,
	reader interface {
		GetByID(context.Context, string) (*codemodels.CodeEntity, error)
	},
	rels []*codemodels.CodeRelationship,
	fromSide bool,
	limit int,
) []EntityRef {
	seen := make(map[string]struct{})
	var refs []EntityRef
	for _, rel := range rels {
		if rel.RelationshipType != "CALLS" && rel.RelationshipType != "REFERENCES" && rel.RelationshipType != "DEPENDS_ON" {
			continue
		}
		targetID := rel.ToEntityID
		if fromSide {
			targetID = rel.FromEntityID
		}
		if _, ok := seen[targetID]; ok {
			continue
		}
		seen[targetID] = struct{}{}
		entity, err := reader.GetByID(ctx, targetID)
		if err != nil || entity == nil {
			continue
		}
		refs = append(refs, codeEntityRelationshipRef(entity, rel.RelationshipType))
		if len(refs) >= limit {
			break
		}
	}
	sortEntityRefs(refs)
	return refs
}

func relatedDecisionsForEntity(entity *codemodels.CodeEntity, links []RepositoryCodeLinkRef) []EntityRef {
	var refs []EntityRef
	seen := make(map[string]struct{})
	for _, link := range links {
		if link.CodeEntityRef.EntityID != entity.ID && link.CodeEntityRef.Label != entity.QualifiedName {
			continue
		}
		if link.RepositoryEntityRef.EntityType != agentsearch.EntityTypeDecision {
			continue
		}
		if _, ok := seen[link.RepositoryEntityRef.EntityID]; ok {
			continue
		}
		seen[link.RepositoryEntityRef.EntityID] = struct{}{}
		refs = append(refs, link.RepositoryEntityRef)
		if len(refs) >= maxBriefingRefs {
			break
		}
	}
	sortEntityRefs(refs)
	return refs
}

func summarizeBriefings(briefings []EntityBriefing) string {
	if len(briefings) == 0 {
		return ""
	}
	var lines []string
	if len(briefings) > 1 {
		lines = append(lines, fmt.Sprintf("Found %d matching code entities:", len(briefings)))
	}
	for _, b := range briefings {
		lines = append(lines, fmt.Sprintf("  - %s [%s]", b.QualifiedName, b.EntityType))
		lines = append(lines, fmt.Sprintf("    Layer: %s", b.Layer))
		lines = append(lines, fmt.Sprintf("    Role: %s", b.Role))
		if b.DefinedIn != "" {
			lines = append(lines, fmt.Sprintf("    Defined in: %s", b.DefinedIn))
		}
		if len(b.Fields) > 0 {
			lines = append(lines, fmt.Sprintf("    Fields: %s", strings.Join(b.Fields, "; ")))
		} else if b.Signature != "" {
			lines = append(lines, fmt.Sprintf("    Signature: %s", b.Signature))
		}
		if len(b.Members) > 0 {
			lines = append(lines, fmt.Sprintf("    Members: %d method(s)", len(b.Members)))
		}
		if len(b.Producers) > 0 {
			lines = append(lines, fmt.Sprintf("    Called by: %s", joinRefLabels(b.Producers)))
		}
		if len(b.Consumers) > 0 {
			lines = append(lines, fmt.Sprintf("    Calls/uses: %s", joinRefLabels(b.Consumers)))
		}
		if len(b.RelatedDecisions) > 0 {
			lines = append(lines, fmt.Sprintf("    Related decisions: %s", joinRefLabels(b.RelatedDecisions)))
		}
	}
	return strings.Join(lines, "\n")
}

func joinRefLabels(refs []EntityRef) string {
	labels := make([]string, 0, len(refs))
	for _, ref := range refs {
		label := strings.TrimSpace(ref.Label)
		if label == "" {
			label = ref.EntityID
		}
		labels = append(labels, label)
	}
	return strings.Join(labels, ", ")
}

func homonymEntitiesForTopic(topic string, entities []*codemodels.CodeEntity) []*codemodels.CodeEntity {
	if isShortNameAmbiguous(topic, entities) {
		return entities
	}
	return nil
}

func isShortNameAmbiguous(symbol string, matches []*codemodels.CodeEntity) bool {
	if len(matches) <= 1 {
		return false
	}
	if strings.Contains(symbol, ".") || strings.Contains(symbol, "/") {
		return false
	}
	for _, m := range matches {
		if strings.EqualFold(m.Name, symbol) {
			return true
		}
	}
	return false
}

func prioritizeAndCapRelated(refs []EntityRef, limit int) []EntityRef {
	if len(refs) <= limit {
		return refs
	}
	scored := make([]EntityRef, len(refs))
	copy(scored, refs)
	sort.SliceStable(scored, func(i, j int) bool {
		return relatedPriority(scored[i].EntityType) < relatedPriority(scored[j].EntityType)
	})
	return scored[:limit]
}

func relatedPriority(entityType string) int {
	switch strings.ToUpper(entityType) {
	case agentsearch.EntityTypeDecision:
		return 1
	case agentsearch.EntityTypeFact:
		return 2
	case "STRUCT":
		return 3
	case "INTERFACE":
		return 4
	case "FILE":
		return 5
	case "FUNCTION", "METHOD":
		return 6
	case agentsearch.EntityTypeEvent:
		return 7
	case agentsearch.EntityTypeExpertise, agentsearch.EntityTypeContributor:
		return 8
	default:
		return 9
	}
}

func preferTypeDefinitionsForTopic(roots []*codemodels.CodeEntity, topic string) []*codemodels.CodeEntity {
	topic = strings.TrimSpace(topic)
	if topic == "" {
		return roots
	}
	short := topic
	if i := strings.LastIndex(topic, "."); i >= 0 {
		short = topic[i+1:]
	}

	var typeDefs []*codemodels.CodeEntity
	for _, r := range roots {
		if !strings.EqualFold(r.Name, short) {
			continue
		}
		switch r.EntityType {
		case codemodels.EntityTypeStruct, codemodels.EntityTypeInterface, codemodels.EntityTypeTypeAlias:
			typeDefs = append(typeDefs, r)
		}
	}
	if len(typeDefs) > 0 {
		return typeDefs
	}
	return roots
}

func filterMatchesByTypes(matches []*codemodels.CodeEntity, allowedTypes ...string) []*codemodels.CodeEntity {
	if len(allowedTypes) == 0 {
		return matches
	}
	allowed := make(map[string]bool, len(allowedTypes))
	for _, t := range allowedTypes {
		allowed[t] = true
	}
	var out []*codemodels.CodeEntity
	for _, m := range matches {
		if allowed[m.EntityType] {
			out = append(out, m)
		}
	}
	return out
}
