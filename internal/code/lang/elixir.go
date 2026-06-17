package lang

import (
	"strings"

	gts "github.com/odvcencio/gotreesitter"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
)

func extractElixir(root *gts.Node, lang *gts.Language, src []byte, filePath string) *FileIndex {
	pkg := packagePathForFile(filePath)
	out := &FileIndex{Language: Elixir}

	gts.Walk(root, func(node *gts.Node, depth int) gts.WalkAction {
		if node == nil {
			return gts.WalkSkipChildren
		}
		if node.Type(lang) != "call" {
			return gts.WalkContinue
		}
		callee := elixirCallee(node, lang, src)
		switch callee {
		case "defmodule":
			if sym := elixirModuleSymbol(node, lang, src, pkg); sym != nil {
				out.Symbols = append(out.Symbols, *sym)
			}
			return gts.WalkContinue
		case "def", "defp":
			if sym := elixirFunctionSymbol(node, lang, src, pkg); sym != nil {
				out.Symbols = append(out.Symbols, *sym)
			}
			return gts.WalkSkipChildren
		case "import", "alias", "require", "use":
			if path := elixirImportPath(node, lang, src); path != "" {
				start, _ := nodeLines(node)
				out.Imports = appendUniqueImport(out.Imports, ImportRef{Path: path, StartLine: start})
			}
			return gts.WalkSkipChildren
		}
		return gts.WalkContinue
	})

	return out
}

func elixirCallee(node *gts.Node, lang *gts.Language, src []byte) string {
	if child := namedChildByType(node, lang, "identifier"); child != nil {
		return nodeText(child, src)
	}
	if child := node.ChildByFieldName("function", lang); child != nil {
		return nodeText(child, src)
	}
	return ""
}

func elixirModuleSymbol(node *gts.Node, lang *gts.Language, src []byte, pkg string) *Symbol {
	name := elixirFirstArgumentName(node, lang, src)
	if name == "" {
		return nil
	}
	start, end := nodeLines(node)
	return &Symbol{
		EntityType:    codemodels.EntityTypeStruct,
		Name:          elixirShortName(name),
		QualifiedName: symbolQualifiedName(pkg, name),
		StartLine:     start,
		EndLine:       end,
		Signature:     "defmodule " + name,
	}
}

func elixirFunctionSymbol(node *gts.Node, lang *gts.Language, src []byte, pkg string) *Symbol {
	name := elixirFirstArgumentName(node, lang, src)
	if name == "" {
		return nil
	}
	start, end := nodeLines(node)
	return &Symbol{
		EntityType:    codemodels.EntityTypeFunction,
		Name:          elixirShortName(name),
		QualifiedName: symbolQualifiedName(pkg, name),
		StartLine:     start,
		EndLine:       end,
		Signature:     "def " + name + "(...)",
	}
}

func elixirImportPath(node *gts.Node, lang *gts.Language, src []byte) string {
	return elixirFirstArgumentName(node, lang, src)
}

func elixirFirstArgumentName(node *gts.Node, lang *gts.Language, src []byte) string {
	args := namedChildByType(node, lang, "arguments")
	if args == nil {
		return ""
	}
	for i := 0; i < args.NamedChildCount(); i++ {
		child := args.NamedChild(i)
		if child == nil {
			continue
		}
		switch child.Type(lang) {
		case "alias", "identifier", "atom", "quoted_atom":
			return strings.Trim(nodeText(child, src), `":`)
		}
	}
	return ""
}

func elixirShortName(qualified string) string {
	if i := strings.LastIndex(qualified, "."); i >= 0 {
		return qualified[i+1:]
	}
	return qualified
}
