package lang

import (
	"strings"

	gts "github.com/odvcencio/gotreesitter"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
)

func extractPHP(root *gts.Node, lang *gts.Language, src []byte, filePath string) *FileIndex {
	pkg := packagePathForFile(filePath)
	out := &FileIndex{Language: PHP}

	gts.Walk(root, func(node *gts.Node, depth int) gts.WalkAction {
		if node == nil {
			return gts.WalkSkipChildren
		}
		switch node.Type(lang) {
		case "namespace_use_declaration", "include_expression", "include_once_expression",
			"require_expression", "require_once_expression":
			path := strings.TrimSpace(nodeText(node, src))
			if path != "" {
				start, _ := nodeLines(node)
				out.Imports = appendUniqueImport(out.Imports, ImportRef{Path: path, StartLine: start})
			}
			return gts.WalkSkipChildren
		case "class_declaration":
			out.Symbols = append(out.Symbols, phpClassSymbols(node, lang, src, pkg)...)
			return gts.WalkSkipChildren
		case "interface_declaration":
			out.Symbols = append(out.Symbols, phpInterfaceSymbol(node, lang, src, pkg))
			return gts.WalkSkipChildren
		case "function_definition":
			if sym := phpFunctionSymbol(node, lang, src, pkg, ""); sym != nil {
				out.Symbols = append(out.Symbols, *sym)
			}
			return gts.WalkSkipChildren
		}
		return gts.WalkContinue
	})

	return out
}

func phpClassSymbols(node *gts.Node, lang *gts.Language, src []byte, pkg string) []Symbol {
	name := nodeName(node, lang, src)
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

	body := namedChildByType(node, lang, "declaration_list")
	if body == nil {
		return symbols
	}
	for i := 0; i < body.NamedChildCount(); i++ {
		child := body.NamedChild(i)
		if child == nil || child.Type(lang) != "method_declaration" {
			continue
		}
		if sym := phpFunctionSymbol(child, lang, src, pkg, name); sym != nil {
			symbols = append(symbols, *sym)
		}
	}
	return symbols
}

func phpInterfaceSymbol(node *gts.Node, lang *gts.Language, src []byte, pkg string) Symbol {
	name := nodeName(node, lang, src)
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

func phpFunctionSymbol(node *gts.Node, lang *gts.Language, src []byte, pkg, receiver string) *Symbol {
	name := nodeName(node, lang, src)
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
			Signature:     "function " + name + "(...)",
		}
	}
	return &Symbol{
		EntityType:    codemodels.EntityTypeMethod,
		Name:          name,
		QualifiedName: methodQualifiedName(pkg, receiver, name),
		StartLine:     start,
		EndLine:       end,
		Signature:     receiver + "->" + name + "(...)",
		Receiver:      receiver,
	}
}
