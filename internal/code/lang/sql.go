package lang

import (
	gts "github.com/odvcencio/gotreesitter"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
)

func extractSQL(root *gts.Node, lang *gts.Language, src []byte, filePath string) *FileIndex {
	pkg := packagePathForFile(filePath)
	out := &FileIndex{Language: SQL}

	gts.Walk(root, func(node *gts.Node, depth int) gts.WalkAction {
		if node == nil {
			return gts.WalkSkipChildren
		}
		switch node.Type(lang) {
		case "create_table_statement":
			if sym := sqlObjectSymbol(node, lang, src, pkg, "table"); sym != nil {
				out.Symbols = append(out.Symbols, *sym)
			}
			return gts.WalkSkipChildren
		case "create_view_statement":
			if sym := sqlObjectSymbol(node, lang, src, pkg, "view"); sym != nil {
				out.Symbols = append(out.Symbols, *sym)
			}
			return gts.WalkSkipChildren
		case "relation":
			if sym := sqlRelationImport(node, lang, src); sym != nil {
				out.Imports = appendUniqueImport(out.Imports, *sym)
			}
			return gts.WalkSkipChildren
		}
		return gts.WalkContinue
	})

	return out
}

func sqlObjectSymbol(node *gts.Node, lang *gts.Language, src []byte, pkg, kind string) *Symbol {
	name := sqlFirstIdentifier(node, lang, src)
	if name == "" {
		return nil
	}
	start, end := nodeLines(node)
	entityType := codemodels.EntityTypeStruct
	if kind == "view" {
		entityType = codemodels.EntityTypeTypeAlias
	}
	return &Symbol{
		EntityType:    entityType,
		Name:          name,
		QualifiedName: symbolQualifiedName(pkg, name),
		StartLine:     start,
		EndLine:       end,
		Signature:     kind + " " + name,
	}
}

func sqlRelationImport(node *gts.Node, lang *gts.Language, src []byte) *ImportRef {
	name := nodeText(node, src)
	if name == "" {
		return nil
	}
	start, _ := nodeLines(node)
	return &ImportRef{Path: name, StartLine: start}
}

func sqlFirstIdentifier(node *gts.Node, lang *gts.Language, src []byte) string {
	if child := namedChildByType(node, lang, "identifier"); child != nil {
		return nodeText(child, src)
	}
	for i := 0; i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child != nil && child.Type(lang) == "identifier" {
			return nodeText(child, src)
		}
	}
	return ""
}
