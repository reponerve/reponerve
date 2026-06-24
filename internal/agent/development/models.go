package development

import (
	"encoding/json"

	codemodels "github.com/reponerve/reponerve/internal/code/models"
	"github.com/reponerve/reponerve/internal/intelligence/feature"
)

// DevelopmentRequest is input for Development Experience workflows.
type DevelopmentRequest struct {
	RepositoryID string
	Topic        string
	// PackagePath disambiguates short symbol names (e.g. internal/context).
	PackagePath string
}

// EntityRef references one repository or code entity in DE output.
type EntityRef struct {
	EntityType string `json:"entity_type"`
	EntityID   string `json:"entity_id"`
	Label      string `json:"label"`
}

// RelationshipRef references a code relationship with evidence.
type RelationshipRef struct {
	RelationshipType string `json:"relationship_type"`
	FromEntityID     string `json:"from_entity_id"`
	ToEntityID       string `json:"to_entity_id"`
	Label            string `json:"label"`
	EvidenceJSON     string `json:"evidence_json"`
}

// EvidenceItem traces upstream authority output.
type EvidenceItem struct {
	Source  string          `json:"source"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// RepositoryCodeLinkRef connects repository memory to code entities.
type RepositoryCodeLinkRef struct {
	RelationshipType    string    `json:"relationship_type"`
	RepositoryEntityRef EntityRef `json:"repository_entity_ref"`
	CodeEntityRef       EntityRef `json:"code_entity_ref"`
	EvidenceJSON        string    `json:"evidence_json"`
}

// CodeContext is authoritative code understanding assembled from Code Intelligence.
type CodeContext struct {
	Modules      []EntityRef         `json:"modules"`
	Files        []EntityRef         `json:"files"`
	Packages     []EntityRef         `json:"packages"`
	Structs      []EntityRef         `json:"structs"`
	Interfaces   []EntityRef         `json:"interfaces"`
	TypeAliases  []EntityRef         `json:"type_aliases"`
	Functions    []EntityRef         `json:"functions"`
	Methods      []EntityRef         `json:"methods"`
	Endpoints    []EntityRef         `json:"endpoints"`
	CallGraph    *codemodels.CallGraph `json:"call_graph,omitempty"`
	Dependencies []RelationshipRef   `json:"dependencies"`
}

// RepositoryContext is repository intelligence assembled from upstream authorities.
type RepositoryContext struct {
	Decisions   []EntityRef `json:"decisions"`
	Facts       []EntityRef `json:"facts"`
	Events      []EntityRef `json:"events"`
	Owners      []EntityRef `json:"owners"`
	Expertise   []EntityRef `json:"expertise"`
	Reviewers   []EntityRef `json:"reviewers"`
	Impact      []EntityRef `json:"impact"`
	ChangePlans []EntityRef `json:"change_plans"`
}

// DevelopmentImpactReport is the structured impact analysis output contract.
type DevelopmentImpactReport struct {
	Subject             string                  `json:"subject"`
	ImpactedDecisions   []EntityRef             `json:"impacted_decisions"`
	ImpactedFacts       []EntityRef             `json:"impacted_facts"`
	ImpactedEvents      []EntityRef             `json:"impacted_events"`
	CodeDependencies    []RelationshipRef       `json:"code_dependencies"`
	DependentAreas      []EntityRef             `json:"dependent_areas"`
	Owners              []EntityRef             `json:"owners"`
	RepositoryCodeLinks []RepositoryCodeLinkRef `json:"repository_code_links"`
	Evidence            []EvidenceItem          `json:"evidence"`
	SourceServices      []string                `json:"source_services"`
}

// DevelopmentReviewGuide is the structured review preparation output contract.
type DevelopmentReviewGuide struct {
	Topic                string                  `json:"topic"`
	RecommendedReviewers []EntityRef             `json:"recommended_reviewers"`
	RequiredExpertise    []EntityRef             `json:"required_expertise"`
	AffectedAreas        []EntityRef             `json:"affected_areas"`
	RelatedKnowledge     []EntityRef             `json:"related_knowledge"`
	DisciplineChecks     []DisciplineCheck       `json:"discipline_checks,omitempty"`
	RecommendedNextTools []string                `json:"recommended_next_tools,omitempty"`
	SuggestedWorkflow    string                  `json:"suggested_workflow"`
	RepositoryCodeLinks  []RepositoryCodeLinkRef `json:"repository_code_links"`
	Evidence             []EvidenceItem          `json:"evidence"`
	SourceServices       []string                `json:"source_services"`
}

// DevelopmentPlan is the structured plan output contract.
type DevelopmentPlan struct {
	Task                string                  `json:"task"`
	EntityBriefings     []EntityBriefing        `json:"entity_briefings,omitempty"`
	SuggestedSteps      []string                `json:"suggested_steps,omitempty"`
	ImpactedAreas       []EntityRef             `json:"impacted_areas"`
	RelevantDecisions   []EntityRef             `json:"relevant_decisions"`
	RelevantFacts       []EntityRef             `json:"relevant_facts"`
	Owners              []EntityRef             `json:"owners"`
	Reviewers           []EntityRef             `json:"reviewers"`
	SuggestedWorkflow   string                  `json:"suggested_workflow"`
	StartingPoints      []EntityRef             `json:"starting_points"`
	RepositoryCodeLinks []RepositoryCodeLinkRef `json:"repository_code_links"`
	Evidence            []EvidenceItem          `json:"evidence"`
	SourceServices      []string                `json:"source_services"`
}

// EntityBriefing is a structured agent-ready summary of one code entity.
type EntityBriefing struct {
	QualifiedName    string      `json:"qualified_name"`
	EntityType       string      `json:"entity_type"`
	Layer            string      `json:"layer"`
	Role             string      `json:"role"`
	DefinedIn        string      `json:"defined_in"`
	Signature        string      `json:"signature,omitempty"`
	Fields           []string    `json:"fields,omitempty"`
	Members          []EntityRef `json:"members,omitempty"`
	Producers        []EntityRef `json:"producers"`
	Consumers        []EntityRef `json:"consumers"`
	RelatedDecisions []EntityRef `json:"related_decisions"`
}

// DevelopmentAnswer is the structured ask output contract.
type DevelopmentAnswer struct {
	Question        string           `json:"question"`
	AnswerType      string           `json:"answer_type"`
	Summary         string           `json:"summary"`
	Plan            *DevelopmentPlan `json:"plan,omitempty"`
	EntityBriefings []EntityBriefing `json:"entity_briefings,omitempty"`
	Related         []EntityRef      `json:"related"`
	Evidence        []EvidenceItem   `json:"evidence"`
	SourceServices  []string         `json:"source_services"`
}

// DevelopmentExplanation is the unified explain output contract.
type DevelopmentExplanation struct {
	Topic               string                  `json:"topic"`
	Feature             *feature.Summary        `json:"feature,omitempty"`
	EntityBriefings     []EntityBriefing        `json:"entity_briefings,omitempty"`
	CodeContext         *CodeContext            `json:"code_context"`
	RepositoryContext   *RepositoryContext      `json:"repository_context"`
	RepositoryCodeLinks []RepositoryCodeLinkRef `json:"repository_code_links"`
	Evidence            []EvidenceItem          `json:"evidence"`
	SourceServices      []string                `json:"source_services"`
}

// ReuseCandidate is an existing symbol or knowledge artifact to prefer before new code.
type ReuseCandidate struct {
	QualifiedName string `json:"qualified_name"`
	EntityType    string `json:"entity_type"`
	DefinedIn     string `json:"defined_in,omitempty"`
	Role          string `json:"role,omitempty"`
	Rank          int    `json:"rank"`
}

// ReuseCheckResult is the structured Reuse Protocol output contract.
type ReuseCheckResult struct {
	Intent               string           `json:"intent"`
	ReuseCandidates      []ReuseCandidate `json:"reuse_candidates"`
	RelatedDecisions     []EntityRef      `json:"related_decisions"`
	RecommendedNextTools []string         `json:"recommended_next_tools"`
	Evidence             []EvidenceItem   `json:"evidence"`
	SourceServices       []string         `json:"source_services"`
}

// ShipCheckItem is one ship readiness finding.
type ShipCheckItem struct {
	Severity string      `json:"severity"`
	Category string      `json:"category"`
	Message  string      `json:"message"`
	Related  []EntityRef `json:"related,omitempty"`
}

// ShipCheckResult is the structured Ship Readiness output contract.
type ShipCheckResult struct {
	Topic                string          `json:"topic"`
	ImpactedAreas        []EntityRef     `json:"impacted_areas"`
	RelatedKnowledge     []EntityRef     `json:"related_knowledge"`
	RecommendedReviewers []EntityRef     `json:"recommended_reviewers"`
	ShipBlockers         []ShipCheckItem `json:"ship_blockers"`
	Advisories           []ShipCheckItem `json:"advisories"`
	RecommendedNextTools []string        `json:"recommended_next_tools"`
	Evidence             []EvidenceItem  `json:"evidence"`
	SourceServices       []string        `json:"source_services"`
}

// ResolvedTopic is the result of topic resolution across authorities.
type ResolvedTopic struct {
	Input               string
	RepositoryHitIDs    map[string]struct{}
	CodeEntityIDs       map[string]struct{}
	RepositoryCodeLinks []*codemodels.RepositoryCodeRelationship
	PrimaryEntityType   string // code | repository | mixed
	MatchEvidence       string
}
