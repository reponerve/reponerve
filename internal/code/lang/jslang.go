package lang

import (
	"strings"

	gts "github.com/odvcencio/gotreesitter"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
)

func extractJSLike(language string, root *gts.Node, lang *gts.Language, src []byte, filePath string) *FileIndex {
	pkg := packagePathForFile(filePath)
	out := &FileIndex{Language: language}

	gts.Walk(root, func(node *gts.Node, depth int) gts.WalkAction {
		if node == nil {
			return gts.WalkSkipChildren
		}
		switch node.Type(lang) {
		case "import_statement":
			path := extractJSImportPath(node, lang, src)
			if path != "" {
				start, _ := nodeLines(node)
				out.Imports = appendUniqueImport(out.Imports, ImportRef{Path: path, StartLine: start})
			}
			return gts.WalkSkipChildren
		case "export_statement":
			decl := namedChildByType(node, lang, "function_declaration")
			if decl != nil {
				out.Symbols = append(out.Symbols, jsFunctionSymbol(decl, lang, src, pkg)...)
			}
			decl = namedChildByType(node, lang, "class_declaration")
			if decl != nil {
				out.Symbols = append(out.Symbols, jsClassSymbols(decl, lang, src, pkg)...)
			}
			if language == TypeScript {
				decl = namedChildByType(node, lang, "interface_declaration")
				if decl != nil {
					out.Symbols = append(out.Symbols, jsInterfaceSymbol(decl, lang, src, pkg))
				}
			}
			return gts.WalkSkipChildren
		case "function_declaration":
			out.Symbols = append(out.Symbols, jsFunctionSymbol(node, lang, src, pkg)...)
			return gts.WalkSkipChildren
		case "class_declaration":
			out.Symbols = append(out.Symbols, jsClassSymbols(node, lang, src, pkg)...)
			return gts.WalkSkipChildren
		case "interface_declaration":
			if language == TypeScript {
				out.Symbols = append(out.Symbols, jsInterfaceSymbol(node, lang, src, pkg))
			}
			return gts.WalkSkipChildren
		}
		return gts.WalkContinue
	})

	return out
}

func extractTypeScript(root *gts.Node, lang *gts.Language, src []byte, filePath string) *FileIndex {
	return extractJSLike(TypeScript, root, lang, src, filePath)
}

func extractJavaScript(root *gts.Node, lang *gts.Language, src []byte, filePath string) *FileIndex {
	return extractJSLike(JavaScript, root, lang, src, filePath)
}

func extractJSImportPath(node *gts.Node, lang *gts.Language, src []byte) string {
	for i := 0; i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child == nil {
			continue
		}
		if child.Type(lang) == "string" {
			return strings.Trim(nodeText(child, src), `"'`)
		}
	}
	return ""
}

func jsFunctionSymbol(node *gts.Node, lang *gts.Language, src []byte, pkg string) []Symbol {
	name := nodeName(node, lang, src)
	if name == "" {
		return nil
	}
	start, end := nodeLines(node)
	return []Symbol{{
		EntityType:    codemodels.EntityTypeFunction,
		Name:          name,
		QualifiedName: symbolQualifiedName(pkg, name),
		StartLine:     start,
		EndLine:       end,
		Signature:     "function " + name + "(...)",
	}}
}

func jsClassSymbols(node *gts.Node, lang *gts.Language, src []byte, pkg string) []Symbol {
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

	body := node.ChildByFieldName("body", lang)
	if body == nil {
		return symbols
	}
	for i := 0; i < body.NamedChildCount(); i++ {
		child := body.NamedChild(i)
		if child == nil || child.Type(lang) != "method_definition" {
			continue
		}
		methodName := nodeName(child, lang, src)
		if methodName == "" {
			continue
		}
		mStart, mEnd := nodeLines(child)
		symbols = append(symbols, Symbol{
			EntityType:    codemodels.EntityTypeMethod,
			Name:          methodName,
			QualifiedName: methodQualifiedName(pkg, name, methodName),
			StartLine:     mStart,
			EndLine:       mEnd,
			Signature:     name + "." + methodName + "(...)",
			Receiver:      name,
		})
	}
	return symbols
}

func jsInterfaceSymbol(node *gts.Node, lang *gts.Language, src []byte, pkg string) Symbol {
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
