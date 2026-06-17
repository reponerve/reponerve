package lang

import (
	"strings"

	gts "github.com/odvcencio/gotreesitter"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
)

func extractLua(root *gts.Node, lang *gts.Language, src []byte, filePath string) *FileIndex {
	pkg := packagePathForFile(filePath)
	out := &FileIndex{Language: Lua}

	gts.Walk(root, func(node *gts.Node, depth int) gts.WalkAction {
		if node == nil {
			return gts.WalkSkipChildren
		}
		switch node.Type(lang) {
		case "function_call":
			if path := luaRequirePath(node, lang, src); path != "" {
				start, _ := nodeLines(node)
				out.Imports = appendUniqueImport(out.Imports, ImportRef{Path: path, StartLine: start})
			}
			return gts.WalkSkipChildren
		case "variable_declaration":
			if child := namedChildByType(node, lang, "assignment_statement"); child != nil {
				if sym := luaTableModuleSymbol(child, lang, src, pkg); sym != nil {
					out.Symbols = append(out.Symbols, *sym)
				}
			}
			return gts.WalkSkipChildren
		case "assignment_statement":
			if sym := luaTableModuleSymbol(node, lang, src, pkg); sym != nil {
				out.Symbols = append(out.Symbols, *sym)
			}
			return gts.WalkSkipChildren
		case "function_declaration", "function_definition":
			if sym := luaFunctionSymbol(node, lang, src, pkg); sym != nil {
				out.Symbols = append(out.Symbols, *sym)
			}
			return gts.WalkSkipChildren
		}
		return gts.WalkContinue
	})

	return out
}

func luaRequirePath(node *gts.Node, lang *gts.Language, src []byte) string {
	callee := namedChildByType(node, lang, "identifier")
	if callee == nil || nodeText(callee, src) != "require" {
		return ""
	}
	for i := 0; i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child != nil && child.Type(lang) == "string" {
			return strings.Trim(nodeText(child, src), `"`)
		}
	}
	return ""
}

func luaTableModuleSymbol(node *gts.Node, lang *gts.Language, src []byte, pkg string) *Symbol {
	if !luaHasTableConstructor(node, lang) {
		return nil
	}
	name := luaIdentifierFromList(node, lang, src, "variable_list")
	if name == "" {
		return nil
	}
	start, end := nodeLines(node)
	return &Symbol{
		EntityType:    codemodels.EntityTypeStruct,
		Name:          name,
		QualifiedName: symbolQualifiedName(pkg, name),
		StartLine:     start,
		EndLine:       end,
		Signature:     "table " + name,
	}
}

func luaHasTableConstructor(node *gts.Node, lang *gts.Language) bool {
	if namedChildByType(node, lang, "table_constructor") != nil {
		return true
	}
	if list := namedChildByType(node, lang, "expression_list"); list != nil {
		for i := 0; i < list.NamedChildCount(); i++ {
			child := list.NamedChild(i)
			if child != nil && child.Type(lang) == "table_constructor" {
				return true
			}
		}
	}
	return false
}

func luaIdentifierFromList(node *gts.Node, lang *gts.Language, src []byte, listField string) string {
	list := node.ChildByFieldName(listField, lang)
	if list == nil {
		list = namedChildByType(node, lang, listField)
	}
	if list == nil {
		return ""
	}
	if child := namedChildByType(list, lang, "identifier"); child != nil {
		return nodeText(child, src)
	}
	return ""
}

func luaFunctionSymbol(node *gts.Node, lang *gts.Language, src []byte, pkg string) *Symbol {
	name := luaFunctionName(node, lang, src)
	if name == "" {
		return nil
	}
	start, end := nodeLines(node)
	entityType := codemodels.EntityTypeFunction
	receiver := ""
	signature := "function " + name + "(...)"
	if strings.Contains(name, ".") || strings.Contains(name, ":") {
		entityType = codemodels.EntityTypeMethod
		parts := strings.FieldsFunc(name, func(r rune) bool { return r == '.' || r == ':' })
		if len(parts) >= 2 {
			receiver = parts[0]
			name = parts[len(parts)-1]
			signature = receiver + "." + name + "(...)"
		}
	}
	qualified := symbolQualifiedName(pkg, luaFunctionName(node, lang, src))
	if entityType == codemodels.EntityTypeMethod && receiver != "" {
		qualified = methodQualifiedName(pkg, receiver, name)
	}
	return &Symbol{
		EntityType:    entityType,
		Name:          name,
		QualifiedName: qualified,
		StartLine:     start,
		EndLine:       end,
		Signature:     signature,
		Receiver:      receiver,
	}
}

func luaFunctionName(node *gts.Node, lang *gts.Language, src []byte) string {
	if name := fieldText(node, lang, "name", src); name != "" {
		return name
	}
	if child := node.ChildByFieldName("name", lang); child != nil {
		return nodeText(child, src)
	}
	for i := 0; i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child == nil {
			continue
		}
		switch child.Type(lang) {
		case "identifier", "dot_index_expression", "method_index_expression":
			return nodeText(child, src)
		}
	}
	return ""
}
