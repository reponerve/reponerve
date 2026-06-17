package lang

// Symbol is a deterministic code symbol extracted from a source file.
type Symbol struct {
	EntityType    string
	Name          string
	QualifiedName string
	StartLine     int
	EndLine       int
	Signature     string
	Receiver      string // methods only
}

// ImportRef is a deterministic import extracted from a source file.
type ImportRef struct {
	Path      string
	Symbol    string
	StartLine int
}

// FileIndex is the deterministic parse result for one source file.
type FileIndex struct {
	Language string
	Imports  []ImportRef
	Symbols  []Symbol
}
