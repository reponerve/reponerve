package lang

import (
	"strings"

	gts "github.com/odvcencio/gotreesitter"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
)

func extractCSharp(root *gts.Node, lang *gts.Language, src []byte, filePath string) *FileIndex {
	pkg := packagePathForFile(filePath)
	out := &FileIndex{Language: CSharp}

	gts.Walk(root, func(node *gts.Node, depth int) gts.WalkAction {
		if node == nil {
			return gts.WalkSkipChildren
		}
		switch node.Type(lang) {
		case "using_directive":
			path := strings.TrimSpace(nodeText(node, src))
			path = strings.TrimPrefix(path, "using ")
			path = strings.TrimSuffix(path, ";")
			if path != "" {
				start, _ := nodeLines(node)
				out.Imports = appendUniqueImport(out.Imports, ImportRef{Path: path, StartLine: start})
			}
			return gts.WalkSkipChildren
		case "class_declaration", "struct_declaration", "record_declaration":
			out.Symbols = append(out.Symbols, csharpClassSymbols(node, lang, src, pkg)...)
			return gts.WalkSkipChildren
		case "interface_declaration":
			out.Symbols = append(out.Symbols, csharpInterfaceSymbol(node, lang, src, pkg))
			return gts.WalkSkipChildren
		case "method_declaration", "local_function_statement":
			if sym := csharpMethodSymbol(node, lang, src, pkg, ""); sym != nil {
				out.Symbols = append(out.Symbols, *sym)
			}
			return gts.WalkSkipChildren
		}
		return gts.WalkContinue
	})

	return out
}

func csharpClassSymbols(node *gts.Node, lang *gts.Language, src []byte, pkg string) []Symbol {
	name := nodeName(node, lang, src)
	if name == "" {
		return nil
	}
	start, end := nodeLines(node)
	kind := node.Type(lang)
	var symbols []Symbol
	symbols = append(symbols, Symbol{
		EntityType:    codemodels.EntityTypeStruct,
		Name:          name,
		QualifiedName: symbolQualifiedName(pkg, name),
		StartLine:     start,
		EndLine:       end,
		Signature:     kind + " " + name,
	})

	body := node.ChildByFieldName("body", lang)
	if body == nil {
		return symbols
	}
	for i := 0; i < body.NamedChildCount(); i++ {
		child := body.NamedChild(i)
		if child == nil {
			continue
		}
		typ := child.Type(lang)
		if typ != "method_declaration" && typ != "local_function_statement" {
			continue
		}
		if sym := csharpMethodSymbol(child, lang, src, pkg, name); sym != nil {
			symbols = append(symbols, *sym)
		}
	}
	return symbols
}

func csharpInterfaceSymbol(node *gts.Node, lang *gts.Language, src []byte, pkg string) Symbol {
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

func csharpMethodSymbol(node *gts.Node, lang *gts.Language, src []byte, pkg, receiver string) *Symbol {
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
			Signature:     name + "(...)",
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
