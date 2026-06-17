package lang

import (
	"strings"

	gts "github.com/odvcencio/gotreesitter"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
)

func extractRust(root *gts.Node, lang *gts.Language, src []byte, filePath string) *FileIndex {
	pkg := packagePathForFile(filePath)
	out := &FileIndex{Language: Rust}

	gts.Walk(root, func(node *gts.Node, depth int) gts.WalkAction {
		if node == nil {
			return gts.WalkSkipChildren
		}
		switch node.Type(lang) {
		case "use_declaration":
			path := extractRustUsePath(node, lang, src)
			if path != "" {
				start, _ := nodeLines(node)
				out.Imports = appendUniqueImport(out.Imports, ImportRef{Path: path, StartLine: start})
			}
			return gts.WalkSkipChildren
		case "function_item":
			out.Symbols = append(out.Symbols, rsFunctionSymbol(node, lang, src, pkg))
			return gts.WalkSkipChildren
		case "struct_item":
			out.Symbols = append(out.Symbols, rsStructSymbol(node, lang, src, pkg))
			return gts.WalkSkipChildren
		case "trait_item":
			out.Symbols = append(out.Symbols, rsTraitSymbol(node, lang, src, pkg))
			return gts.WalkSkipChildren
		case "impl_item":
			out.Symbols = append(out.Symbols, rsImplMethods(node, lang, src, pkg)...)
			return gts.WalkSkipChildren
		}
		return gts.WalkContinue
	})

	return out
}

func extractRustUsePath(node *gts.Node, lang *gts.Language, src []byte) string {
	for i := 0; i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child == nil {
			continue
		}
		typ := child.Type(lang)
		if typ == "scoped_identifier" || typ == "identifier" || typ == "scoped_use_list" || typ == "use_list" {
			text := strings.TrimSpace(nodeText(child, src))
			if text != "" {
				return text
			}
		}
	}
	text := strings.TrimSpace(strings.TrimPrefix(nodeText(node, src), "use "))
	text = strings.TrimSuffix(text, ";")
	return text
}

func rsFunctionSymbol(node *gts.Node, lang *gts.Language, src []byte, pkg string) Symbol {
	name := fieldText(node, lang, "name", src)
	if name == "" {
		name = rustItemName(node, lang, src)
	}
	start, end := nodeLines(node)
	return Symbol{
		EntityType:    codemodels.EntityTypeFunction,
		Name:          name,
		QualifiedName: symbolQualifiedName(pkg, name),
		StartLine:     start,
		EndLine:       end,
		Signature:     "fn " + name + "(...)",
	}
}

func rsStructSymbol(node *gts.Node, lang *gts.Language, src []byte, pkg string) Symbol {
	name := fieldText(node, lang, "name", src)
	if name == "" {
		name = rustItemName(node, lang, src)
	}
	start, end := nodeLines(node)
	return Symbol{
		EntityType:    codemodels.EntityTypeStruct,
		Name:          name,
		QualifiedName: symbolQualifiedName(pkg, name),
		StartLine:     start,
		EndLine:       end,
		Signature:     "struct " + name,
	}
}

func rsTraitSymbol(node *gts.Node, lang *gts.Language, src []byte, pkg string) Symbol {
	name := fieldText(node, lang, "name", src)
	if name == "" {
		name = rustItemName(node, lang, src)
	}
	start, end := nodeLines(node)
	return Symbol{
		EntityType:    codemodels.EntityTypeInterface,
		Name:          name,
		QualifiedName: symbolQualifiedName(pkg, name),
		StartLine:     start,
		EndLine:       end,
		Signature:     "trait " + name,
	}
}

func rsImplMethods(node *gts.Node, lang *gts.Language, src []byte, pkg string) []Symbol {
	typeNode := node.ChildByFieldName("type", lang)
	receiver := rustTypeName(typeNode, lang, src)
	if receiver == "" {
		return nil
	}

	body := node.ChildByFieldName("body", lang)
	if body == nil {
		return nil
	}

	var symbols []Symbol
	for i := 0; i < body.NamedChildCount(); i++ {
		child := body.NamedChild(i)
		if child == nil || child.Type(lang) != "function_item" {
			continue
		}
		name := fieldText(child, lang, "name", src)
		if name == "" {
			continue
		}
		start, end := nodeLines(child)
		symbols = append(symbols, Symbol{
			EntityType:    codemodels.EntityTypeMethod,
			Name:          name,
			QualifiedName: methodQualifiedName(pkg, receiver, name),
			StartLine:     start,
			EndLine:       end,
			Signature:     receiver + "." + name + "(...)",
			Receiver:      receiver,
		})
	}
	return symbols
}

func rustItemName(node *gts.Node, lang *gts.Language, src []byte) string {
	for i := 0; i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child != nil && child.Type(lang) == "type_identifier" {
			return nodeText(child, src)
		}
		if child != nil && child.Type(lang) == "identifier" {
			return nodeText(child, src)
		}
	}
	return ""
}

func rustTypeName(node *gts.Node, lang *gts.Language, src []byte) string {
	if node == nil {
		return ""
	}
	switch node.Type(lang) {
	case "type_identifier", "identifier":
		return nodeText(node, src)
	case "generic_type":
		return fieldText(node, lang, "type", src)
	case "scoped_type_identifier":
		parts := strings.Split(nodeText(node, src), "::")
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
	}
	return strings.TrimSpace(nodeText(node, src))
}
