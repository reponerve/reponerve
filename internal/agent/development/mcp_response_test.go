package development_test

import (
	"encoding/json"
	"testing"

	"github.com/reponerve/reponerve/internal/agent/development"
)

func TestMCPResult_JSONShape(t *testing.T) {
	answer := &development.DevelopmentAnswer{
		Question:   "What is Foo?",
		AnswerType: "concept_explanation",
		EntityBriefings: []development.EntityBriefing{
			{
				QualifiedName: "internal/foo.Foo",
				EntityType:    "struct",
				Fields:        []string{"Bar string"},
			},
		},
	}

	payload := development.NewMCPResult(development.FormatAnswer(answer), answer)

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := decoded["formatted"]; !ok {
		t.Fatal("missing formatted")
	}
	if _, ok := decoded["structured"]; !ok {
		t.Fatal("missing structured")
	}
	if _, ok := decoded["agent"]; !ok {
		t.Fatal("missing agent metadata")
	}

	var briefings struct {
		EntityBriefings []struct {
			Fields []string `json:"fields"`
		} `json:"entity_briefings"`
	}
	if err := json.Unmarshal(decoded["structured"], &briefings); err != nil {
		t.Fatalf("structured decode: %v", err)
	}
	if len(briefings.EntityBriefings) != 1 || len(briefings.EntityBriefings[0].Fields) != 1 {
		t.Fatalf("expected fields in structured briefings")
	}
}
