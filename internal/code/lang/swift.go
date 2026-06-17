package lang

import (
	"strings"

	gts "github.com/odvcencio/gotreesitter"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
)

func extractSwift(root *gts.Node, lang *gts.Language, src []byte, filePath string) *FileIndex {
	pkg := packagePathForFile(filePath)
	out := &FileIndex{Language: Swift}

	gts.Walk(root, func(node *gts.Node, depth int) gts.WalkAction {
		if node == nil {
			return gts.WalkSkipChildren
		}
		switch node.Type(lang) {
		case "import_declaration":
			path := strings.TrimSpace(nodeText(node, src))
			path = strings.TrimPrefix(path, "import ")
			if path != "" {
				start, _ := nodeLines(node)
				out.Imports = appendUniqueImport(out.Imports, ImportRef{Path: path, StartLine: start})
			}
			return gts.WalkSkipChildren
		case "class_declaration":
			out.Symbols = append(out.Symbols, swiftClassSymbols(node, lang, src, pkg)...)
			return gts.WalkSkipChildren
		case "protocol_declaration":
			out.Symbols = append(out.Symbols, swiftProtocolSymbol(node, lang, src, pkg))
			return gts.WalkSkipChildren
		case "function_declaration":
			if sym := swiftFunctionSymbol(node, lang, src, pkg, ""); sym != nil {
				out.Symbols = append(out.Symbols, *sym)
			}
			return gts.WalkSkipChildren
		}
		return gts.WalkContinue
	})

	return out
}

func swiftTypeName(node *gts.Node, lang *gts.Language, src []byte) string {
	if child := namedChildByType(node, lang, "type_identifier"); child != nil {
		return nodeText(child, src)
	}
	return nodeName(node, lang, src)
}

func swiftFunctionName(node *gts.Node, lang *gts.Language, src []byte) string {
	if child := namedChildByType(node, lang, "simple_identifier"); child != nil {
		return nodeText(child, src)
	}
	return nodeName(node, lang, src)
}

func swiftClassSymbols(node *gts.Node, lang *gts.Language, src []byte, pkg string) []Symbol {
	name := swiftTypeName(node, lang, src)
	if name == "" {
		return nil
	}
	start, end := nodeLines(node)
	var symbols []Symbol
	symbols = append(symbols, Symbol{
		EntityType:    codemodels.EntityTypeStruct,
		Name:          name,
		QualifiedName: symbolQualifiedName(pkg, name),
		StartLine:     start,
		EndLine:       end,
		Signature:     "class " + name,
	})

	body := namedChildByType(node, lang, "class_body")
	if body == nil {
		return symbols
	}
	for i := 0; i < body.NamedChildCount(); i++ {
		child := body.NamedChild(i)
		if child == nil || child.Type(lang) != "function_declaration" {
			continue
		}
		if sym := swiftFunctionSymbol(child, lang, src, pkg, name); sym != nil {
			symbols = append(symbols, *sym)
		}
	}
	return symbols
}

func swiftProtocolSymbol(node *gts.Node, lang *gts.Language, src []byte, pkg string) Symbol {
	name := swiftTypeName(node, lang, src)
	start, end := nodeLines(node)
	return Symbol{
		EntityType:    codemodels.EntityTypeInterface,
		Name:          name,
		QualifiedName: symbolQualifiedName(pkg, name),
		StartLine:     start,
		EndLine:       end,
		Signature:     "protocol " + name,
	}
}

func swiftFunctionSymbol(node *gts.Node, lang *gts.Language, src []byte, pkg, receiver string) *Symbol {
	name := swiftFunctionName(node, lang, src)
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
			Signature:     "func " + name + "(...)",
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
