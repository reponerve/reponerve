package models

import "time"

// Entity type constants for code intelligence.
const (
	EntityTypeModule      = "module"
	EntityTypePackage     = "package"
	EntityTypeFile        = "file"
	EntityTypeStruct      = "struct"
	EntityTypeInterface   = "interface"
	EntityTypeTypeAlias   = "type_alias"
	EntityTypeFunction    = "function"
	EntityTypeMethod      = "method"
	EntityTypeEndpoint    = "endpoint"
)

// CodeEntity represents a deterministic code intelligence entity.
type CodeEntity struct {
	ID            string
	RepositoryID  string
	EntityType    string
	Name          string
	QualifiedName string
	FilePath      string
	PackagePath   string
	ModulePath    string
	Language      string
	StartLine     int
	EndLine       int
	Signature     string
	EndpointType  string
	EvidenceJSON  string
	IndexedAt     time.Time
}

// CodeRelationship represents a deterministic relationship between code entities.
type CodeRelationship struct {
	ID               string
	RepositoryID     string
	FromEntityID     string
	ToEntityID       string
	RelationshipType string
	EvidenceJSON     string
	IndexedAt        time.Time
}

// RepositoryCodeRelationship links repository memory entities to code entities.
type RepositoryCodeRelationship struct {
	ID                   string
	RepositoryID         string
	RepositoryEntityID   string
	RepositoryEntityType string
	CodeEntityID         string
	CodeEntityType       string
	RelationshipType     string
	EvidenceJSON         string
	IndexedAt            time.Time
}

// CodeIndexState tracks code indexing progress for a repository.
type CodeIndexState struct {
	RepositoryID      string
	LastIndexedAt     time.Time
	ModuleCount       int
	FileCount         int
	EntityCount       int
	RelationshipCount int
	LinkCount         int
}
