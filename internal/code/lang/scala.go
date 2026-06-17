package lang

import (
	"strings"

	gts "github.com/odvcencio/gotreesitter"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
)

func extractScala(root *gts.Node, lang *gts.Language, src []byte, filePath string) *FileIndex {
	pkg := packagePathForFile(filePath)
	out := &FileIndex{Language: Scala}

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
		case "class_definition":
			out.Symbols = append(out.Symbols, scalaTypeSymbols(node, lang, src, pkg, "class")...)
			return gts.WalkSkipChildren
		case "object_definition":
			out.Symbols = append(out.Symbols, scalaTypeSymbols(node, lang, src, pkg, "object")...)
			return gts.WalkSkipChildren
		case "trait_definition":
			out.Symbols = append(out.Symbols, scalaTraitSymbol(node, lang, src, pkg))
			return gts.WalkSkipChildren
		case "function_definition", "function_declaration":
			if sym := scalaFunctionSymbol(node, lang, src, pkg, ""); sym != nil {
				out.Symbols = append(out.Symbols, *sym)
			}
			return gts.WalkSkipChildren
		}
		return gts.WalkContinue
	})

	return out
}

func scalaTypeName(node *gts.Node, lang *gts.Language, src []byte) string {
	if child := namedChildByType(node, lang, "identifier"); child != nil {
		return nodeText(child, src)
	}
	return nodeName(node, lang, src)
}

func scalaFunctionName(node *gts.Node, lang *gts.Language, src []byte) string {
	if child := namedChildByType(node, lang, "identifier"); child != nil {
		return nodeText(child, src)
	}
	return nodeName(node, lang, src)
}

func scalaTypeSymbols(node *gts.Node, lang *gts.Language, src []byte, pkg, kind string) []Symbol {
	name := scalaTypeName(node, lang, src)
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
		Signature:     kind + " " + name,
	})

	body := namedChildByType(node, lang, "template_body")
	if body == nil {
		return symbols
	}
	for i := 0; i < body.NamedChildCount(); i++ {
		child := body.NamedChild(i)
		if child == nil {
			continue
		}
		typ := child.Type(lang)
		if typ != "function_definition" && typ != "function_declaration" {
			continue
		}
		if sym := scalaFunctionSymbol(child, lang, src, pkg, name); sym != nil {
			symbols = append(symbols, *sym)
		}
	}
	return symbols
}

func scalaTraitSymbol(node *gts.Node, lang *gts.Language, src []byte, pkg string) Symbol {
	name := scalaTypeName(node, lang, src)
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

func scalaFunctionSymbol(node *gts.Node, lang *gts.Language, src []byte, pkg, receiver string) *Symbol {
	name := scalaFunctionName(node, lang, src)
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
			Signature:     "def " + name + "(...)",
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
