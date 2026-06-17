package lang

import (
	"strings"

	gts "github.com/odvcencio/gotreesitter"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
)

func extractBash(root *gts.Node, lang *gts.Language, src []byte, filePath string) *FileIndex {
	pkg := packagePathForFile(filePath)
	out := &FileIndex{Language: Bash}

	gts.Walk(root, func(node *gts.Node, depth int) gts.WalkAction {
		if node == nil {
			return gts.WalkSkipChildren
		}
		switch node.Type(lang) {
		case "command":
			if path := bashSourcePath(node, lang, src); path != "" {
				start, _ := nodeLines(node)
				out.Imports = appendUniqueImport(out.Imports, ImportRef{Path: path, StartLine: start})
			}
			return gts.WalkSkipChildren
		case "function_definition":
			if sym := bashFunctionSymbol(node, lang, src, pkg); sym != nil {
				out.Symbols = append(out.Symbols, *sym)
			}
			return gts.WalkSkipChildren
		}
		return gts.WalkContinue
	})

	return out
}

func bashSourcePath(node *gts.Node, lang *gts.Language, src []byte) string {
	cmdName := namedChildByType(node, lang, "command_name")
	if cmdName == nil {
		return ""
	}
	word := namedChildByType(cmdName, lang, "word")
	if word == nil || nodeText(word, src) != "source" {
		return ""
	}
	for i := 0; i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child != nil && child.Type(lang) == "word" && nodeText(child, src) != "source" {
			return strings.Trim(nodeText(child, src), `"'`)
		}
	}
	return ""
}

func bashFunctionSymbol(node *gts.Node, lang *gts.Language, src []byte, pkg string) *Symbol {
	name := bashFunctionName(node, lang, src)
	if name == "" {
		return nil
	}
	start, end := nodeLines(node)
	return &Symbol{
		EntityType:    codemodels.EntityTypeFunction,
		Name:          name,
		QualifiedName: symbolQualifiedName(pkg, name),
		StartLine:     start,
		EndLine:       end,
		Signature:     name + "()",
	}
}

func bashFunctionName(node *gts.Node, lang *gts.Language, src []byte) string {
	for i := 0; i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child != nil && child.Type(lang) == "word" {
			return nodeText(child, src)
		}
	}
	return nodeName(node, lang, src)
}
