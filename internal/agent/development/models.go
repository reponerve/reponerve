package development

import (
	"encoding/json"

	codemodels "github.com/reponerve/reponerve/internal/code/models"
)

// DevelopmentRequest is input for Development Experience workflows.
type DevelopmentRequest struct {
	RepositoryID string
	Topic        string
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

// DevelopmentAnswer is the structured ask output contract.
type DevelopmentAnswer struct {
	Question       string         `json:"question"`
	AnswerType     string         `json:"answer_type"`
	Summary        string         `json:"summary"`
	Related        []EntityRef    `json:"related"`
	Evidence       []EvidenceItem `json:"evidence"`
	SourceServices []string       `json:"source_services"`
}

// DevelopmentExplanation is the unified explain output contract.
type DevelopmentExplanation struct {
	Topic               string                  `json:"topic"`
	CodeContext         *CodeContext            `json:"code_context"`
	RepositoryContext   *RepositoryContext      `json:"repository_context"`
	RepositoryCodeLinks []RepositoryCodeLinkRef `json:"repository_code_links"`
	Evidence            []EvidenceItem          `json:"evidence"`
	SourceServices      []string                `json:"source_services"`
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
