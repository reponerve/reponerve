package indexer

import "encoding/json"

type entityEvidence struct {
	Source    string `json:"source"`
	File      string `json:"file"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
	Parser    string `json:"parser"`
}

type relationshipEvidence struct {
	Source       string `json:"source"`
	File         string `json:"file"`
	Line         int    `json:"line"`
	Relationship string `json:"relationship"`
	FromSymbol   string `json:"from_symbol,omitempty"`
	ToSymbol     string `json:"to_symbol,omitempty"`
}

func marshalEntityEvidence(file string, startLine, endLine int) string {
	b, _ := json.Marshal(entityEvidence{
		Source:    "go/ast",
		File:      file,
		StartLine: startLine,
		EndLine:   endLine,
		Parser:    "go/parser",
	})
	return string(b)
}

func marshalRelationshipEvidence(file, relType, fromSymbol, toSymbol string, line int) string {
	b, _ := json.Marshal(relationshipEvidence{
		Source:       "go/ast",
		File:         file,
		Line:         line,
		Relationship: relType,
		FromSymbol:   fromSymbol,
		ToSymbol:     toSymbol,
	})
	return string(b)
}
