package lang

import (
	"strings"

	gts "github.com/odvcencio/gotreesitter"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
)

func extractKotlin(root *gts.Node, lang *gts.Language, src []byte, filePath string) *FileIndex {
	pkg := packagePathForFile(filePath)
	out := &FileIndex{Language: Kotlin}

	gts.Walk(root, func(node *gts.Node, depth int) gts.WalkAction {
		if node == nil {
			return gts.WalkSkipChildren
		}
		switch node.Type(lang) {
		case "import_header", "import_list":
			path := strings.TrimSpace(nodeText(node, src))
			path = strings.TrimPrefix(path, "import ")
			if path != "" {
				start, _ := nodeLines(node)
				out.Imports = appendUniqueImport(out.Imports, ImportRef{Path: path, StartLine: start})
			}
			return gts.WalkSkipChildren
		case "class_declaration", "object_declaration":
			if kotlinIsInterface(node, lang) {
				out.Symbols = append(out.Symbols, kotlinInterfaceSymbol(node, lang, src, pkg))
			} else {
				out.Symbols = append(out.Symbols, kotlinClassSymbols(node, lang, src, pkg)...)
			}
			return gts.WalkSkipChildren
		case "interface_declaration":
			out.Symbols = append(out.Symbols, kotlinInterfaceSymbol(node, lang, src, pkg))
			return gts.WalkSkipChildren
		case "function_declaration":
			if sym := kotlinFunctionSymbol(node, lang, src, pkg, ""); sym != nil {
				out.Symbols = append(out.Symbols, *sym)
			}
			return gts.WalkSkipChildren
		}
		return gts.WalkContinue
	})

	return out
}

func kotlinClassSymbols(node *gts.Node, lang *gts.Language, src []byte, pkg string) []Symbol {
	name := kotlinTypeName(node, lang, src)
	if name == "" {
		return nil
	}
	start, end := nodeLines(node)
	isInterface := kotlinIsInterface(node, lang)
	entityType := codemodels.EntityTypeStruct
	signature := "class " + name
	if isInterface {
		entityType = codemodels.EntityTypeInterface
		signature = "interface " + name
	}

	var symbols []Symbol
	symbols = append(symbols, Symbol{
		EntityType:    entityType,
		Name:          name,
		QualifiedName: symbolQualifiedName(pkg, name),
		StartLine:     start,
		EndLine:       end,
		Signature:     signature,
	})
	if isInterface {
		return symbols
	}

	body := node.ChildByFieldName("body", lang)
	if body == nil {
		body = namedChildByType(node, lang, "class_body")
	}
	if body == nil {
		return symbols
	}
	for i := 0; i < body.NamedChildCount(); i++ {
		child := body.NamedChild(i)
		if child == nil || child.Type(lang) != "function_declaration" {
			continue
		}
		if sym := kotlinFunctionSymbol(child, lang, src, pkg, name); sym != nil {
			symbols = append(symbols, *sym)
		}
	}
	return symbols
}

func kotlinInterfaceSymbol(node *gts.Node, lang *gts.Language, src []byte, pkg string) Symbol {
	name := kotlinTypeName(node, lang, src)
	start, end := nodeLines(node)
	return Symbol{
		EntityType:    codemodels.EntityTypeInterface,
		Name:          name,
		QualifiedName: symbolQualifiedName(pkg, name),
		StartLine:     start,
		EndLine:       end,
		Signature:     "interface " + name,
	}
}

func kotlinFunctionSymbol(node *gts.Node, lang *gts.Language, src []byte, pkg, receiver string) *Symbol {
	name := kotlinFunctionName(node, lang, src)
	if name == "" {
		return nil
	}
	start, end := nodeLines(node)
	if receiver == "" {
		return &Symbol{
			EntityType:    codemodels.EntityTypeFunction,
			Name:          name,
			QualifiedName: symbolQualifiedName(pkg, name),
			StartLine:     start,
			EndLine:       end,
			Signature:     "fun " + name + "(...)",
		}
	}
	return &Symbol{
		EntityType:    codemodels.EntityTypeMethod,
		Name:          name,
		QualifiedName: methodQualifiedName(pkg, receiver, name),
		StartLine:     start,
		EndLine:       end,
		Signature:     receiver + "." + name + "(...)",
		Receiver:      receiver,
	}
}

func kotlinTypeName(node *gts.Node, lang *gts.Language, src []byte) string {
	if child := namedChildByType(node, lang, "type_identifier"); child != nil {
		return kotlinFlattenType(child, lang, src)
	}
	return nodeName(node, lang, src)
}

func kotlinFlattenType(node *gts.Node, lang *gts.Language, src []byte) string {
	if node == nil {
		return ""
	}
	if node.Type(lang) == "type_identifier" {
		var parts []string
		for i := 0; i < node.NamedChildCount(); i++ {
			child := node.NamedChild(i)
			if child != nil && child.Type(lang) == "simple_identifier" {
				parts = append(parts, nodeText(child, src))
			}
		}
		if len(parts) > 0 {
			return strings.Join(parts, ".")
		}
	}
	return nodeText(node, src)
}

func kotlinFunctionName(node *gts.Node, lang *gts.Language, src []byte) string {
	if name := fieldText(node, lang, "name", src); name != "" {
		return name
	}
	if child := namedChildByType(node, lang, "simple_identifier"); child != nil {
		return nodeText(child, src)
	}
	return nodeName(node, lang, src)
}

func kotlinIsInterface(node *gts.Node, lang *gts.Language) bool {
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil && child.Type(lang) == "interface" {
			return true
		}
	}
	return false
}
