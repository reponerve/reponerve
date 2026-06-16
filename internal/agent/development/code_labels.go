package development

import (
	"fmt"
	"strings"

	codemodels "github.com/reponerve/reponerve/internal/code/models"
)

// HumanCodeEntityLabel formats a code entity for call-graph and relationship briefings.
func HumanCodeEntityLabel(e *codemodels.CodeEntity, relationshipType string) string {
	if e == nil {
		return ""
	}
	name := strings.TrimSpace(e.Name)
	pkg := strings.TrimSpace(e.PackagePath)
	file := strings.TrimSpace(e.FilePath)

	switch e.EntityType {
	case codemodels.EntityTypeMethod:
		return fmt.Sprintf("method %s (%s)", name, pkg)
	case codemodels.EntityTypeFunction:
		return fmt.Sprintf("function %s (%s)", name, pkg)
	case codemodels.EntityTypeStruct:
		return fmt.Sprintf("struct %s (%s)", name, pkg)
	case codemodels.EntityTypeInterface:
		return fmt.Sprintf("interface %s (%s)", name, pkg)
	case codemodels.EntityTypeTypeAlias:
		return fmt.Sprintf("type %s (%s)", name, pkg)
	case codemodels.EntityTypeFile:
		if file != "" {
			return fmt.Sprintf("file %s", file)
		}
		return fmt.Sprintf("file %s", name)
	case codemodels.EntityTypePackage:
		return fmt.Sprintf("package %s", pkg)
	default:
		if relationshipType == "CALLS" && file != "" {
			return fmt.Sprintf("%s — %s", e.QualifiedName, file)
		}
		return e.QualifiedName
	}
}

func codeEntityRelationshipRef(e *codemodels.CodeEntity, relationshipType string) EntityRef {
	return EntityRef{
		EntityType: strings.ToUpper(e.EntityType),
		EntityID:   e.ID,
		Label:      HumanCodeEntityLabel(e, relationshipType),
	}
}
