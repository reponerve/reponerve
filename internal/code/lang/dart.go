package lang

import (
	"strings"

	gts "github.com/odvcencio/gotreesitter"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
)

func extractDart(root *gts.Node, lang *gts.Language, src []byte, filePath string) *FileIndex {
	pkg := packagePathForFile(filePath)
	out := &FileIndex{Language: Dart}

	gts.Walk(root, func(node *gts.Node, depth int) gts.WalkAction {
		if node == nil {
			return gts.WalkSkipChildren
		}
		switch node.Type(lang) {
		case "import_or_export":
			if path := dartImportPath(node, lang, src); path != "" {
				start, _ := nodeLines(node)
				out.Imports = appendUniqueImport(out.Imports, ImportRef{Path: path, StartLine: start})
			}
			return gts.WalkSkipChildren
		case "class_definition":
			out.Symbols = append(out.Symbols, dartClassSymbols(node, lang, src, pkg)...)
			return gts.WalkSkipChildren
		case "function_signature":
			parent := node.Parent()
			if parent != nil && parent.Type(lang) == "method_signature" {
				return gts.WalkSkipChildren
			}
			if sym := dartFunctionSymbol(node, lang, src, pkg, ""); sym != nil {
				out.Symbols = append(out.Symbols, *sym)
			}
			return gts.WalkSkipChildren
		}
		return gts.WalkContinue
	})

	return out
}

func dartImportPath(node *gts.Node, lang *gts.Language, src []byte) string {
	for i := 0; i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child == nil {
			continue
		}
		if child.Type(lang) == "string_literal" {
			return strings.Trim(nodeText(child, src), `'`)
		}
		if uri := namedChildByType(child, lang, "uri"); uri != nil {
			if lit := namedChildByType(uri, lang, "string_literal"); lit != nil {
				return strings.Trim(nodeText(lit, src), `'`)
			}
		}
	}
	return ""
}

func dartClassSymbols(node *gts.Node, lang *gts.Language, src []byte, pkg string) []Symbol {
	name := dartTypeName(node, lang, src)
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
		if child == nil || child.Type(lang) != "method_signature" {
			continue
		}
		sig := namedChildByType(child, lang, "function_signature")
		if sig == nil {
			continue
		}
		if sym := dartFunctionSymbol(sig, lang, src, pkg, name); sym != nil {
			symbols = append(symbols, *sym)
		}
	}
	return symbols
}

func dartTypeName(node *gts.Node, lang *gts.Language, src []byte) string {
	if child := namedChildByType(node, lang, "identifier"); child != nil {
		return nodeText(child, src)
	}
	return nodeName(node, lang, src)
}

func dartFunctionSymbol(node *gts.Node, lang *gts.Language, src []byte, pkg, receiver string) *Symbol {
	name := dartFunctionName(node, lang, src)
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

func dartFunctionName(node *gts.Node, lang *gts.Language, src []byte) string {
	if child := namedChildByType(node, lang, "identifier"); child != nil {
		return nodeText(child, src)
	}
	return nodeName(node, lang, src)
}
