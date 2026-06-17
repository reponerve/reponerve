package lang

import (
	"strings"

	gts "github.com/odvcencio/gotreesitter"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
)

func extractZig(root *gts.Node, lang *gts.Language, src []byte, filePath string) *FileIndex {
	pkg := packagePathForFile(filePath)
	out := &FileIndex{Language: Zig}

	gts.Walk(root, func(node *gts.Node, depth int) gts.WalkAction {
		if node == nil {
			return gts.WalkSkipChildren
		}
		switch node.Type(lang) {
		case "variable_declaration":
			if path := zigImportPath(node, lang, src); path != "" {
				start, _ := nodeLines(node)
				out.Imports = appendUniqueImport(out.Imports, ImportRef{Path: path, StartLine: start})
			}
			if sym := zigStructSymbol(node, lang, src, pkg); sym != nil {
				out.Symbols = append(out.Symbols, sym...)
			}
			return gts.WalkSkipChildren
		case "function_declaration":
			if isZigTopLevel(node, lang) {
				if sym := zigFunctionSymbol(node, lang, src, pkg, ""); sym != nil {
					out.Symbols = append(out.Symbols, *sym)
				}
			}
			return gts.WalkSkipChildren
		}
		return gts.WalkContinue
	})

	return out
}

func zigImportPath(node *gts.Node, lang *gts.Language, src []byte) string {
	builtin := namedChildByType(node, lang, "builtin_function")
	if builtin == nil {
		return ""
	}
	id := namedChildByType(builtin, lang, "builtin_identifier")
	if id == nil || nodeText(id, src) != "@import" {
		return ""
	}
	for i := 0; i < builtin.NamedChildCount(); i++ {
		child := builtin.NamedChild(i)
		if child != nil && child.Type(lang) == "string" {
			return strings.Trim(nodeText(child, src), `"`)
		}
	}
	return ""
}

func zigStructSymbol(node *gts.Node, lang *gts.Language, src []byte, pkg string) []Symbol {
	structDecl := namedChildByType(node, lang, "struct_declaration")
	if structDecl == nil {
		return nil
	}
	name := ""
	if id := namedChildByType(node, lang, "identifier"); id != nil {
		name = nodeText(id, src)
	}
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
		Signature:     "struct " + name,
	})
	for i := 0; i < structDecl.NamedChildCount(); i++ {
		child := structDecl.NamedChild(i)
		if child == nil || child.Type(lang) != "function_declaration" {
			continue
		}
		if sym := zigFunctionSymbol(child, lang, src, pkg, name); sym != nil {
			symbols = append(symbols, *sym)
		}
	}
	return symbols
}

func zigFunctionSymbol(node *gts.Node, lang *gts.Language, src []byte, pkg, receiver string) *Symbol {
	name := zigFunctionName(node, lang, src)
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
			Signature:     "fn " + name + "(...)",
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

func zigFunctionName(node *gts.Node, lang *gts.Language, src []byte) string {
	if child := namedChildByType(node, lang, "identifier"); child != nil {
		return nodeText(child, src)
	}
	return nodeName(node, lang, src)
}

func isZigTopLevel(node *gts.Node, lang *gts.Language) bool {
	parent := node.Parent()
	return parent != nil && parent.Type(lang) == "source_file"
}
