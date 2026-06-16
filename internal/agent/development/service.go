package development

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	agentimpact "github.com/reponerve/reponerve/internal/agent/impact"
	agentsearch "github.com/reponerve/reponerve/internal/agent/search"
	"github.com/reponerve/reponerve/internal/agent/qa"
	"github.com/reponerve/reponerve/internal/code"
	graphimpact "github.com/reponerve/reponerve/internal/graph/impact"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
	"github.com/reponerve/reponerve/internal/intelligence/changeplan"
	"github.com/reponerve/reponerve/internal/intelligence/learning"
	"github.com/reponerve/reponerve/internal/intelligence/reviewers"
	"github.com/reponerve/reponerve/internal/query/storage"
)

const (
	sourceCodeIntelligence      = "code_intelligence"
	sourceRepositorySearch      = "repository_search"
	sourceRepositoryCodeLinks   = "repository_code_links"
	sourceOwnershipIntelligence = "ownership_intelligence"
)

// Service orchestrates Code Intelligence and Repository Intelligence for Development Experience.
type Service struct {
	codeService       *code.Service
	router            *Router
	searchService     *agentsearch.Service
	qaService         *qa.Service
	codeEntityReader  storage.CodeEntityReader
	relReader         storage.CodeRelationshipReader
	repoCodeReader    storage.RepositoryCodeRelationshipReader
	decisionReader    storage.DecisionReader
	factReader        storage.FactReader
	eventReader       storage.EventReader
	expertiseReader   storage.ExpertiseReader
	contributorReader storage.ContributorReader
	sourceReader      storage.SourceReader
	repositoryPath    string
	learningService     *learning.Service
	reviewerService     *reviewers.Service
	changePlanService   *changeplan.Service
	graphImpactService  *graphimpact.Service
	agentImpactService  *agentimpact.Service
}

// NewService creates a Development Experience service.
func NewService(
	codeService *code.Service,
	searchService *agentsearch.Service,
	codeEntityReader storage.CodeEntityReader,
	relReader storage.CodeRelationshipReader,
	repoCodeReader storage.RepositoryCodeRelationshipReader,
	decisionReader storage.DecisionReader,
	factReader storage.FactReader,
	eventReader storage.EventReader,
	expertiseReader storage.ExpertiseReader,
	qaService *qa.Service,
	contributorReader storage.ContributorReader,
	sourceReader storage.SourceReader,
	repositoryPath string,
	learningService *learning.Service,
	reviewerService *reviewers.Service,
	changePlanService *changeplan.Service,
	graphImpactService *graphimpact.Service,
	agentImpactService *agentimpact.Service,
) *Service {
	return &Service{
		codeService:       codeService,
		router:            NewRouter(searchService, codeEntityReader, repoCodeReader),
		searchService:     searchService,
		qaService:         qaService,
		codeEntityReader:  codeEntityReader,
		relReader:         relReader,
		repoCodeReader:    repoCodeReader,
		decisionReader:    decisionReader,
		factReader:        factReader,
		eventReader:       eventReader,
		expertiseReader:   expertiseReader,
		contributorReader: contributorReader,
		sourceReader:      sourceReader,
		repositoryPath:    repositoryPath,
		learningService:    learningService,
		reviewerService:    reviewerService,
		changePlanService:  changePlanService,
		graphImpactService: graphImpactService,
		agentImpactService: agentImpactService,
	}
}

// Explain provides combined repository and code understanding for a natural-language topic.
func (s *Service) Explain(ctx context.Context, req DevelopmentRequest) (*DevelopmentExplanation, error) {
	if strings.TrimSpace(req.Topic) == "" {
		return nil, fmt.Errorf("topic cannot be empty")
	}
	topic, err := s.router.ResolveTopic(ctx, req.RepositoryID, req.Topic)
	if err != nil {
		return nil, err
	}
	return s.assembleExplanation(ctx, req.RepositoryID, req.Topic, topic)
}

// ExplainFile explains one indexed file path.
func (s *Service) ExplainFile(ctx context.Context, repositoryID, filePath string) (*DevelopmentExplanation, error) {
	codeCtx, err := s.codeService.ResolveFile(ctx, repositoryID, filePath)
	if err != nil {
		return nil, err
	}
	topic := &ResolvedTopic{
		Input:            filePath,
		RepositoryHitIDs: make(map[string]struct{}),
		CodeEntityIDs:    make(map[string]struct{}),
		PrimaryEntityType: "code",
	}
	for _, e := range allCodeEntities(codeCtx) {
		topic.CodeEntityIDs[e.ID] = struct{}{}
	}
	if err := s.router.expandRepositoryCodeLinks(ctx, repositoryID, topic); err != nil {
		return nil, err
	}
	return s.assembleFromCodeContext(ctx, repositoryID, filePath, codeCtx, topic)
}

// ExplainFunction explains a function or method symbol.
func (s *Service) ExplainFunction(ctx context.Context, repositoryID, symbol, packagePath string) (*DevelopmentExplanation, error) {
	return s.explainSymbol(ctx, repositoryID, symbol, packagePath, codemodels.EntityTypeFunction, codemodels.EntityTypeMethod)
}

// ExplainStruct explains a struct symbol.
func (s *Service) ExplainStruct(ctx context.Context, repositoryID, symbol, packagePath string) (*DevelopmentExplanation, error) {
	return s.explainSymbol(ctx, repositoryID, symbol, packagePath, codemodels.EntityTypeStruct)
}

// ExplainInterface explains an interface symbol.
func (s *Service) ExplainInterface(ctx context.Context, repositoryID, symbol, packagePath string) (*DevelopmentExplanation, error) {
	return s.explainSymbol(ctx, repositoryID, symbol, packagePath, codemodels.EntityTypeInterface)
}

// ExplainType explains a type alias symbol.
func (s *Service) ExplainType(ctx context.Context, repositoryID, symbol, packagePath string) (*DevelopmentExplanation, error) {
	return s.explainSymbol(ctx, repositoryID, symbol, packagePath, codemodels.EntityTypeTypeAlias)
}

func (s *Service) explainSymbol(ctx context.Context, repositoryID, symbol, packagePath string, allowedTypes ...string) (*DevelopmentExplanation, error) {
	matches, err := s.codeService.ListSymbolMatches(ctx, repositoryID, symbol, packagePath)
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("symbol not found: %s", symbol)
	}
	matches = filterMatchesByTypes(matches, allowedTypes...)
	if len(matches) == 0 {
		hint := symbolTypeHint(symbol, allowedTypes)
		if hint != "" {
			return nil, fmt.Errorf("symbol %q is not one of requested types (%s); %s", symbol, strings.Join(allowedTypes, ", "), hint)
		}
		return nil, fmt.Errorf("symbol %q is not one of requested types (%s)", symbol, strings.Join(allowedTypes, ", "))
	}
	if isShortNameAmbiguous(symbol, matches) {
		return s.assembleAmbiguousSymbolExplanation(ctx, repositoryID, symbol, matches)
	}

	codeCtx, err := s.codeService.ResolveSymbol(ctx, repositoryID, symbol, packagePath)
	if err != nil {
		return nil, err
	}
	if len(allowedTypes) > 0 {
		if !symbolMatchesTypes(codeCtx, allowedTypes) {
			hint := symbolTypeHint(symbol, allowedTypes)
			if hint != "" {
				return nil, fmt.Errorf("symbol %q is not one of requested types (%s); %s", symbol, strings.Join(allowedTypes, ", "), hint)
			}
			return nil, fmt.Errorf("symbol %q is not one of requested types (%s)", symbol, strings.Join(allowedTypes, ", "))
		}
	}
	topic := &ResolvedTopic{
		Input:            symbol,
		RepositoryHitIDs: make(map[string]struct{}),
		CodeEntityIDs:    make(map[string]struct{}),
		PrimaryEntityType: "code",
	}
	for _, e := range allCodeEntities(codeCtx) {
		topic.CodeEntityIDs[e.ID] = struct{}{}
	}
	if err := s.router.expandRepositoryCodeLinks(ctx, repositoryID, topic); err != nil {
		return nil, err
	}
	return s.assembleFromCodeContext(ctx, repositoryID, symbol, codeCtx, topic)
}

func (s *Service) assembleExplanation(ctx context.Context, repositoryID, topic string, resolved *ResolvedTopic) (*DevelopmentExplanation, error) {
	var merged codemodels.CodeExplanationContext
	merged.Subject = topic

	entities, err := s.loadCodeEntities(ctx, repositoryID, resolved.CodeEntityIDs)
	if err != nil {
		return nil, err
	}
	if homonyms := homonymEntitiesForTopic(topic, entities); len(homonyms) > 1 {
		return s.assembleAmbiguousSymbolExplanation(ctx, repositoryID, topic, homonyms)
	}
	groupIntoCodeExplanation(&merged, entities)

	if root := pickCallGraphRoot(entities); root != nil && s.relReader != nil {
		allRels, err := s.relReader.ListByRepository(ctx, repositoryID)
		if err != nil {
			return nil, err
		}
		merged.CallGraph = code.BuildCallGraphFromRelationships(root.ID, allRels, 10)
		outbound, err := s.relReader.ListByFromEntity(ctx, root.ID)
		if err != nil {
			return nil, err
		}
		merged.Dependencies = code.CollectSymbolDependencies(root.ID, outbound)
	}

	topicCopy := *resolved
	return s.assembleFromCodeContext(ctx, repositoryID, topic, &merged, &topicCopy)
}

func (s *Service) assembleFromCodeContext(
	ctx context.Context,
	repositoryID, topic string,
	codeCtx *codemodels.CodeExplanationContext,
	resolved *ResolvedTopic,
) (*DevelopmentExplanation, error) {
	out := &DevelopmentExplanation{
		Topic: topic,
		CodeContext: &CodeContext{
			Modules:     entityRefsFromCode(codeCtx.Modules),
			Files:       entityRefsFromCode(codeCtx.Files),
			Packages:    entityRefsFromCode(codeCtx.Packages),
			Structs:     entityRefsFromCode(codeCtx.Structs),
			Interfaces:  entityRefsFromCode(codeCtx.Interfaces),
			TypeAliases: entityRefsFromCode(codeCtx.TypeAliases),
			Functions:   entityRefsFromCode(codeCtx.Functions),
			Methods:     entityRefsFromCode(codeCtx.Methods),
			Endpoints:   entityRefsFromCode(codeCtx.Endpoints),
			CallGraph:   codeCtx.CallGraph,
			Dependencies: relationshipRefs(codeCtx.Dependencies),
		},
		RepositoryContext: &RepositoryContext{},
		SourceServices:    []string{},
	}

	services := map[string]struct{}{}
	if hasCodeContent(out.CodeContext) {
		services[sourceCodeIntelligence] = struct{}{}
		appendEvidence(&out.Evidence, sourceCodeIntelligence, "code_context", map[string]string{
			"subject": codeCtx.Subject,
		})
	}

	repoCtx, repoEvidence, repoServices, err := s.buildRepositoryContext(ctx, repositoryID, resolved, topic)
	if err != nil {
		return nil, err
	}
	out.RepositoryContext = repoCtx
	out.Evidence = append(out.Evidence, repoEvidence...)
	for _, svc := range repoServices {
		services[svc] = struct{}{}
	}

	links, linkEvidence, err := s.buildRepositoryCodeLinkRefs(ctx, repositoryID, resolved)
	if err != nil {
		return nil, err
	}
	out.RepositoryCodeLinks = links
	if len(links) > 0 {
		services[sourceRepositoryCodeLinks] = struct{}{}
		out.Evidence = append(out.Evidence, linkEvidence...)
	}

	out.SourceServices = sortServices(services)
	sortEvidence(out.Evidence)
	sortRepositoryCodeLinks(out.RepositoryCodeLinks)
	sortEntityRefsAll(out)

	entities, err := s.loadCodeEntities(ctx, repositoryID, resolved.CodeEntityIDs)
	if err != nil {
		return nil, err
	}
	if len(entities) == 0 && codeCtx != nil {
		entities = allCodeEntities(codeCtx)
	}
	briefings, err := s.buildEntityBriefings(ctx, repositoryID, topic, entities, out.RepositoryCodeLinks)
	if err != nil {
		return nil, err
	}
	out.EntityBriefings = briefings

	return out, nil
}

func (s *Service) assembleAmbiguousSymbolExplanation(
	ctx context.Context,
	repositoryID, symbol string,
	matches []*codemodels.CodeEntity,
) (*DevelopmentExplanation, error) {
	topic := &ResolvedTopic{
		Input:             symbol,
		RepositoryHitIDs:  make(map[string]struct{}),
		CodeEntityIDs:     make(map[string]struct{}),
		PrimaryEntityType: "code",
		MatchEvidence:     "ambiguous_symbol",
	}
	for _, m := range matches {
		topic.CodeEntityIDs[m.ID] = struct{}{}
	}
	if err := s.router.expandRepositoryCodeLinks(ctx, repositoryID, topic); err != nil {
		return nil, err
	}

	merged := &codemodels.CodeExplanationContext{Subject: symbol}
	groupIntoCodeExplanation(merged, matches)
	return s.assembleFromCodeContext(ctx, repositoryID, symbol, merged, topic)
}

func (s *Service) buildRepositoryContext(
	ctx context.Context,
	repositoryID string,
	resolved *ResolvedTopic,
	topic string,
) (*RepositoryContext, []EvidenceItem, []string, error) {
	repoCtx := &RepositoryContext{}
	var evidence []EvidenceItem
	services := map[string]struct{}{}

	for entityID := range resolved.RepositoryHitIDs {
		ref, ev, svc, err := s.resolveRepositoryEntity(ctx, repositoryID, entityID)
		if err != nil {
			continue
		}
		if ref == nil {
			continue
		}
		switch ref.EntityType {
		case agentsearch.EntityTypeDecision:
			repoCtx.Decisions = append(repoCtx.Decisions, *ref)
		case agentsearch.EntityTypeFact:
			repoCtx.Facts = append(repoCtx.Facts, *ref)
		case agentsearch.EntityTypeEvent:
			repoCtx.Events = append(repoCtx.Events, *ref)
		case agentsearch.EntityTypeExpertise:
			repoCtx.Expertise = append(repoCtx.Expertise, *ref)
		case agentsearch.EntityTypeContributor:
			repoCtx.Owners = append(repoCtx.Owners, *ref)
		default:
			continue
		}
		evidence = append(evidence, ev...)
		for _, sname := range svc {
			services[sname] = struct{}{}
		}
	}

	expertise, owners, ev, err := s.matchExpertise(ctx, repositoryID, topic)
	if err != nil {
		return nil, nil, nil, err
	}
	repoCtx.Expertise = appendUniqueRefs(repoCtx.Expertise, expertise)
	repoCtx.Owners = appendUniqueRefs(repoCtx.Owners, owners)
	evidence = append(evidence, ev...)
	if len(expertise) > 0 || len(owners) > 0 {
		services[sourceOwnershipIntelligence] = struct{}{}
	}

	sortEntityRefs(repoCtx.Decisions)
	sortEntityRefs(repoCtx.Facts)
	sortEntityRefs(repoCtx.Events)
	sortEntityRefs(repoCtx.Owners)
	sortEntityRefs(repoCtx.Expertise)

	svcList := sortServices(services)
	if len(resolved.RepositoryHitIDs) > 0 {
		services[sourceRepositorySearch] = struct{}{}
		svcList = sortServices(services)
		appendEvidence(&evidence, sourceRepositorySearch, "topic_resolution", map[string]string{
			"match_evidence": resolved.MatchEvidence,
		})
	}

	return repoCtx, evidence, svcList, nil
}

func (s *Service) resolveRepositoryEntity(ctx context.Context, repositoryID, entityID string) (*EntityRef, []EvidenceItem, []string, error) {
	if d, err := s.decisionReader.GetByID(ctx, entityID); err == nil && d != nil {
		ref := &EntityRef{EntityType: agentsearch.EntityTypeDecision, EntityID: d.ID, Label: d.Title}
		var ev []EvidenceItem
		appendEvidence(&ev, sourceRepositorySearch, "decision", map[string]string{
			"id": d.ID, "title": d.Title, "source_id": d.SourceID,
		})
		return ref, ev, []string{sourceRepositorySearch}, nil
	}
	if f, err := s.factReader.GetByID(ctx, entityID); err == nil && f != nil {
		label := strings.TrimSpace(strings.Join([]string{f.Subject, f.Predicate, f.Object}, " "))
		ref := &EntityRef{EntityType: agentsearch.EntityTypeFact, EntityID: f.ID, Label: label}
		var ev []EvidenceItem
		appendEvidence(&ev, sourceRepositorySearch, "fact", map[string]string{
			"id": f.ID, "subject": f.Subject, "predicate": f.Predicate, "object": f.Object,
		})
		return ref, ev, []string{sourceRepositorySearch}, nil
	}
	if ev, err := s.eventReader.GetByID(ctx, entityID); err == nil && ev != nil {
		ref := &EntityRef{EntityType: agentsearch.EntityTypeEvent, EntityID: ev.ID, Label: ev.Title}
		var evItems []EvidenceItem
		appendEvidence(&evItems, sourceRepositorySearch, "event", map[string]string{
			"id": ev.ID, "title": ev.Title, "event_type": ev.EventType,
		})
		return ref, evItems, []string{sourceRepositorySearch}, nil
	}
	if exp, err := s.expertiseReader.ListByRepository(ctx, repositoryID); err == nil {
		for _, e := range exp {
			if e.ID == entityID {
				ref := &EntityRef{EntityType: agentsearch.EntityTypeExpertise, EntityID: e.ID, Label: e.Domain}
				ev := evidenceFromJSON(sourceOwnershipIntelligence, "expertise", e.EvidenceJSON)
				return ref, ev, []string{sourceOwnershipIntelligence}, nil
			}
		}
	}
	return nil, nil, nil, nil
}

func (s *Service) matchExpertise(ctx context.Context, repositoryID, topic string) ([]EntityRef, []EntityRef, []EvidenceItem, error) {
	terms := topicTerms(topic)
	if len(terms) == 0 {
		return nil, nil, nil, nil
	}
	all, err := s.expertiseReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, nil, nil, err
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].Score != all[j].Score {
			return all[i].Score > all[j].Score
		}
		return all[i].Domain < all[j].Domain
	})

	var expertise []EntityRef
	var owners []EntityRef
	var evidence []EvidenceItem
	seenOwners := map[string]struct{}{}

	for _, e := range all {
		domain := strings.ToLower(e.Domain)
		matched := false
		for _, term := range terms {
			if strings.Contains(domain, term) {
				matched = true
				break
			}
		}
		if !matched {
			continue
		}
		expertise = append(expertise, EntityRef{
			EntityType: agentsearch.EntityTypeExpertise,
			EntityID:   e.ID,
			Label:      e.Domain,
		})
		appendEvidence(&evidence, sourceOwnershipIntelligence, "expertise_match", map[string]any{
			"domain": e.Domain,
			"score":  e.Score,
		})
		if _, ok := seenOwners[e.ContributorID]; !ok {
			seenOwners[e.ContributorID] = struct{}{}
			owners = append(owners, EntityRef{
				EntityType: agentsearch.EntityTypeContributor,
				EntityID:   e.ContributorID,
				Label:      e.ContributorID,
			})
		}
	}
	return expertise, owners, evidence, nil
}

func (s *Service) buildRepositoryCodeLinkRefs(
	ctx context.Context,
	repositoryID string,
	resolved *ResolvedTopic,
) ([]RepositoryCodeLinkRef, []EvidenceItem, error) {
	var links []RepositoryCodeLinkRef
	var evidence []EvidenceItem

	for _, link := range resolved.RepositoryCodeLinks {
		repoRef, err := s.repositoryEntityRef(ctx, repositoryID, link)
		if err != nil {
			continue
		}
		codeEntity, err := s.codeEntityReader.GetByID(ctx, link.CodeEntityID)
		if err != nil {
			continue
		}
		links = append(links, RepositoryCodeLinkRef{
			RelationshipType:    link.RelationshipType,
			RepositoryEntityRef: repoRef,
			CodeEntityRef:       codeEntityRef(codeEntity),
			EvidenceJSON:        link.EvidenceJSON,
		})
		if link.EvidenceJSON != "" {
			evidence = append(evidence, EvidenceItem{
				Source:  sourceRepositoryCodeLinks,
				Type:    "link",
				Payload: json.RawMessage(link.EvidenceJSON),
			})
		}
	}
	return links, evidence, nil
}

func (s *Service) repositoryEntityRef(ctx context.Context, repositoryID string, link *codemodels.RepositoryCodeRelationship) (EntityRef, error) {
	ref, _, _, err := s.resolveRepositoryEntity(ctx, repositoryID, link.RepositoryEntityID)
	if err != nil || ref == nil {
		return EntityRef{
			EntityType: link.RepositoryEntityType,
			EntityID:   link.RepositoryEntityID,
			Label:      link.RepositoryEntityID,
		}, nil
	}
	return *ref, nil
}

func (s *Service) loadCodeEntities(ctx context.Context, repositoryID string, ids map[string]struct{}) ([]*codemodels.CodeEntity, error) {
	var entities []*codemodels.CodeEntity
	seen := make(map[string]struct{}, len(ids))
	for id := range ids {
		e, err := s.codeEntityReader.GetByID(ctx, id)
		if err != nil {
			continue
		}
		entities = append(entities, e)
		seen[e.ID] = struct{}{}
		if e.EntityType == codemodels.EntityTypeFile {
			fileEntities, err := s.codeEntityReader.ListByFilePath(ctx, repositoryID, e.FilePath)
			if err != nil {
				return nil, err
			}
			for _, fe := range fileEntities {
				if _, ok := seen[fe.ID]; ok {
					continue
				}
				seen[fe.ID] = struct{}{}
				entities = append(entities, fe)
			}
		}
	}
	sort.Slice(entities, func(i, j int) bool {
		if entities[i].EntityType != entities[j].EntityType {
			return entities[i].EntityType < entities[j].EntityType
		}
		return entities[i].QualifiedName < entities[j].QualifiedName
	})
	return entities, nil
}

func groupIntoCodeExplanation(ctxOut *codemodels.CodeExplanationContext, entities []*codemodels.CodeEntity) {
	for _, e := range entities {
		switch e.EntityType {
		case codemodels.EntityTypeModule:
			ctxOut.Modules = appendUniqueCode(ctxOut.Modules, e)
		case codemodels.EntityTypePackage:
			ctxOut.Packages = appendUniqueCode(ctxOut.Packages, e)
		case codemodels.EntityTypeFile:
			ctxOut.Files = appendUniqueCode(ctxOut.Files, e)
		case codemodels.EntityTypeStruct:
			ctxOut.Structs = appendUniqueCode(ctxOut.Structs, e)
		case codemodels.EntityTypeInterface:
			ctxOut.Interfaces = appendUniqueCode(ctxOut.Interfaces, e)
		case codemodels.EntityTypeTypeAlias:
			ctxOut.TypeAliases = appendUniqueCode(ctxOut.TypeAliases, e)
		case codemodels.EntityTypeFunction:
			ctxOut.Functions = appendUniqueCode(ctxOut.Functions, e)
		case codemodels.EntityTypeMethod:
			ctxOut.Methods = appendUniqueCode(ctxOut.Methods, e)
		case codemodels.EntityTypeEndpoint:
			ctxOut.Endpoints = appendUniqueCode(ctxOut.Endpoints, e)
		}
	}
}

func pickCallGraphRoot(entities []*codemodels.CodeEntity) *codemodels.CodeEntity {
	priority := map[string]int{
		codemodels.EntityTypeFunction: 1,
		codemodels.EntityTypeMethod:   2,
		codemodels.EntityTypeStruct:   3,
	}
	var best *codemodels.CodeEntity
	bestPri := 999
	for _, e := range entities {
		p, ok := priority[e.EntityType]
		if !ok || p >= bestPri {
			continue
		}
		best = e
		bestPri = p
	}
	return best
}

func symbolTypeHint(symbol string, allowedTypes []string) string {
	for _, t := range allowedTypes {
		if t == codemodels.EntityTypeTypeAlias {
			return fmt.Sprintf("try: reponerve explain-struct %q", symbol)
		}
		if t == codemodels.EntityTypeStruct {
			return fmt.Sprintf("try: reponerve explain-type %q if it is a type alias", symbol)
		}
	}
	return ""
}

func symbolMatchesTypes(ctx *codemodels.CodeExplanationContext, types []string) bool {
	allowed := map[string]bool{}
	for _, t := range types {
		allowed[t] = true
	}
	for _, e := range allCodeEntities(ctx) {
		if isSymbolType(e.EntityType) && allowed[e.EntityType] {
			return true
		}
	}
	return false
}

func isSymbolType(entityType string) bool {
	switch entityType {
	case codemodels.EntityTypeStruct, codemodels.EntityTypeInterface, codemodels.EntityTypeTypeAlias,
		codemodels.EntityTypeFunction, codemodels.EntityTypeMethod:
		return true
	default:
		return false
	}
}

func allCodeEntities(ctx *codemodels.CodeExplanationContext) []*codemodels.CodeEntity {
	var out []*codemodels.CodeEntity
	out = append(out, ctx.Modules...)
	out = append(out, ctx.Packages...)
	out = append(out, ctx.Files...)
	out = append(out, ctx.Structs...)
	out = append(out, ctx.Interfaces...)
	out = append(out, ctx.TypeAliases...)
	out = append(out, ctx.Functions...)
	out = append(out, ctx.Methods...)
	out = append(out, ctx.Endpoints...)
	return out
}

func entityRefsFromCode(entities []*codemodels.CodeEntity) []EntityRef {
	refs := make([]EntityRef, 0, len(entities))
	for _, e := range entities {
		refs = append(refs, codeEntityRef(e))
	}
	sortEntityRefs(refs)
	return refs
}

func codeEntityRef(e *codemodels.CodeEntity) EntityRef {
	return EntityRef{
		EntityType: strings.ToUpper(e.EntityType),
		EntityID:   e.ID,
		Label:      e.QualifiedName,
	}
}

func relationshipRefs(rels []*codemodels.CodeRelationship) []RelationshipRef {
	out := make([]RelationshipRef, 0, len(rels))
	for _, rel := range rels {
		out = append(out, RelationshipRef{
			RelationshipType: rel.RelationshipType,
			FromEntityID:     rel.FromEntityID,
			ToEntityID:       rel.ToEntityID,
			Label:            rel.RelationshipType,
			EvidenceJSON:     rel.EvidenceJSON,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].RelationshipType != out[j].RelationshipType {
			return out[i].RelationshipType < out[j].RelationshipType
		}
		return out[i].ToEntityID < out[j].ToEntityID
	})
	return out
}

func appendUniqueCode(list []*codemodels.CodeEntity, e *codemodels.CodeEntity) []*codemodels.CodeEntity {
	for _, existing := range list {
		if existing.ID == e.ID {
			return list
		}
	}
	return append(list, e)
}

func appendUniqueRefs(list []EntityRef, refs []EntityRef) []EntityRef {
	seen := make(map[string]struct{}, len(list))
	for _, r := range list {
		seen[r.EntityID] = struct{}{}
	}
	for _, r := range refs {
		if _, ok := seen[r.EntityID]; ok {
			continue
		}
		seen[r.EntityID] = struct{}{}
		list = append(list, r)
	}
	return list
}

func hasCodeContent(ctx *CodeContext) bool {
	if ctx == nil {
		return false
	}
	return len(ctx.Modules) > 0 || len(ctx.Files) > 0 || len(ctx.Packages) > 0 ||
		len(ctx.Structs) > 0 || len(ctx.Interfaces) > 0 || len(ctx.TypeAliases) > 0 ||
		len(ctx.Functions) > 0 || len(ctx.Methods) > 0 || len(ctx.Endpoints) > 0
}

func hasRepositoryContent(ctx *RepositoryContext) bool {
	if ctx == nil {
		return false
	}
	return len(ctx.Decisions) > 0 || len(ctx.Facts) > 0 || len(ctx.Events) > 0 ||
		len(ctx.Owners) > 0 || len(ctx.Expertise) > 0 || len(ctx.Reviewers) > 0 ||
		len(ctx.Impact) > 0 || len(ctx.ChangePlans) > 0
}

func appendEvidence(items *[]EvidenceItem, source, typ string, payload any) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return
	}
	*items = append(*items, EvidenceItem{Source: source, Type: typ, Payload: raw})
}

func evidenceFromJSON(source, typ, evidenceJSON string) []EvidenceItem {
	if evidenceJSON == "" {
		return nil
	}
	return []EvidenceItem{{Source: source, Type: typ, Payload: json.RawMessage(evidenceJSON)}}
}

func sortServices(services map[string]struct{}) []string {
	out := make([]string, 0, len(services))
	for s := range services {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

func sortEvidence(items []EvidenceItem) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Source != items[j].Source {
			return items[i].Source < items[j].Source
		}
		return items[i].Type < items[j].Type
	})
}

func sortEntityRefs(refs []EntityRef) {
	sort.Slice(refs, func(i, j int) bool {
		if refs[i].EntityType != refs[j].EntityType {
			return refs[i].EntityType < refs[j].EntityType
		}
		if refs[i].Label != refs[j].Label {
			return refs[i].Label < refs[j].Label
		}
		return refs[i].EntityID < refs[j].EntityID
	})
}

func sortRepositoryCodeLinks(links []RepositoryCodeLinkRef) {
	sort.Slice(links, func(i, j int) bool {
		if links[i].RelationshipType != links[j].RelationshipType {
			return links[i].RelationshipType < links[j].RelationshipType
		}
		if links[i].RepositoryEntityRef.EntityID != links[j].RepositoryEntityRef.EntityID {
			return links[i].RepositoryEntityRef.EntityID < links[j].RepositoryEntityRef.EntityID
		}
		return links[i].CodeEntityRef.EntityID < links[j].CodeEntityRef.EntityID
	})
}

func sortEntityRefsAll(out *DevelopmentExplanation) {
	if out.CodeContext != nil {
		sortEntityRefs(out.CodeContext.Modules)
		sortEntityRefs(out.CodeContext.Files)
		sortEntityRefs(out.CodeContext.Packages)
		sortEntityRefs(out.CodeContext.Structs)
		sortEntityRefs(out.CodeContext.Interfaces)
		sortEntityRefs(out.CodeContext.TypeAliases)
		sortEntityRefs(out.CodeContext.Functions)
		sortEntityRefs(out.CodeContext.Methods)
		sortEntityRefs(out.CodeContext.Endpoints)
	}
	if out.RepositoryContext != nil {
		sortEntityRefs(out.RepositoryContext.Decisions)
		sortEntityRefs(out.RepositoryContext.Facts)
		sortEntityRefs(out.RepositoryContext.Events)
		sortEntityRefs(out.RepositoryContext.Owners)
		sortEntityRefs(out.RepositoryContext.Expertise)
		sortEntityRefs(out.RepositoryContext.Reviewers)
		sortEntityRefs(out.RepositoryContext.Impact)
		sortEntityRefs(out.RepositoryContext.ChangePlans)
	}
}
