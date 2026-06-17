package lang

import (
	"strings"

	gts "github.com/odvcencio/gotreesitter"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
)

func extractCpp(root *gts.Node, lang *gts.Language, src []byte, filePath string) *FileIndex {
	pkg := packagePathForFile(filePath)
	out := &FileIndex{Language: Cpp}

	gts.Walk(root, func(node *gts.Node, depth int) gts.WalkAction {
		if node == nil {
			return gts.WalkSkipChildren
		}
		switch node.Type(lang) {
		case "preproc_include":
			path := clangIncludePath(node, lang, src)
			if path != "" {
				start, _ := nodeLines(node)
				out.Imports = appendUniqueImport(out.Imports, ImportRef{Path: path, StartLine: start})
			}
			return gts.WalkSkipChildren
		case "class_specifier", "struct_specifier":
			out.Symbols = append(out.Symbols, clangTypeSymbols(node, lang, src, pkg)...)
			return gts.WalkSkipChildren
		case "function_definition":
			if sym := clangFunctionSymbol(node, lang, src, pkg, ""); sym != nil {
				out.Symbols = append(out.Symbols, *sym)
			}
			return gts.WalkSkipChildren
		}
		return gts.WalkContinue
	})

	return out
}

func extractC(root *gts.Node, lang *gts.Language, src []byte, filePath string) *FileIndex {
	pkg := packagePathForFile(filePath)
	out := &FileIndex{Language: C}

	gts.Walk(root, func(node *gts.Node, depth int) gts.WalkAction {
		if node == nil {
			return gts.WalkSkipChildren
		}
		switch node.Type(lang) {
		case "preproc_include":
			path := clangIncludePath(node, lang, src)
			if path != "" {
				start, _ := nodeLines(node)
				out.Imports = appendUniqueImport(out.Imports, ImportRef{Path: path, StartLine: start})
			}
			return gts.WalkSkipChildren
		case "struct_specifier":
			out.Symbols = append(out.Symbols, clangTypeSymbols(node, lang, src, pkg)...)
			return gts.WalkSkipChildren
		case "type_definition":
			if child := namedChildByType(node, lang, "struct_specifier"); child != nil {
				name := clangTypeName(child, lang, src)
				if alias := namedChildByType(node, lang, "type_identifier"); alias != nil {
					name = nodeText(alias, src)
				}
				if name != "" {
					start, end := nodeLines(node)
					out.Symbols = append(out.Symbols, Symbol{
						EntityType:    codemodels.EntityTypeStruct,
						Name:          name,
						QualifiedName: symbolQualifiedName(pkg, name),
						StartLine:     start,
						EndLine:       end,
						Signature:     "struct " + name,
					})
				}
			}
			return gts.WalkSkipChildren
		case "function_definition":
			if sym := clangFunctionSymbol(node, lang, src, pkg, ""); sym != nil {
				out.Symbols = append(out.Symbols, *sym)
			}
			return gts.WalkSkipChildren
		}
		return gts.WalkContinue
	})

	return out
}

func clangIncludePath(node *gts.Node, lang *gts.Language, src []byte) string {
	for i := 0; i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child == nil {
			continue
		}
		typ := child.Type(lang)
		if typ == "string_literal" || typ == "system_lib_string" || typ == "path" {
			return strings.Trim(nodeText(child, src), `"<>`)
		}
	}
	text := strings.TrimSpace(nodeText(node, src))
	text = strings.TrimPrefix(text, "#include ")
	return strings.Trim(text, `"<>`)
}

func clangTypeName(node *gts.Node, lang *gts.Language, src []byte) string {
	if child := namedChildByType(node, lang, "type_identifier"); child != nil {
		return nodeText(child, src)
	}
	return nodeName(node, lang, src)
}

func clangTypeSymbols(node *gts.Node, lang *gts.Language, src []byte, pkg string) []Symbol {
	name := clangTypeName(node, lang, src)
	if name == "" {
		return nil
	}
	start, end := nodeLines(node)
	kind := "class"
	if node.Type(lang) == "struct_specifier" {
		kind = "struct"
	}
	var symbols []Symbol
	symbols = append(symbols, Symbol{
		EntityType:    codemodels.EntityTypeStruct,
		Name:          name,
		QualifiedName: symbolQualifiedName(pkg, name),
		StartLine:     start,
		EndLine:       end,
		Signature:     kind + " " + name,
	})

	body := namedChildByType(node, lang, "field_declaration_list")
	if body == nil {
		return symbols
	}
	for i := 0; i < body.NamedChildCount(); i++ {
		child := body.NamedChild(i)
		if child == nil || child.Type(lang) != "function_definition" {
			continue
		}
		if sym := clangFunctionSymbol(child, lang, src, pkg, name); sym != nil {
			symbols = append(symbols, *sym)
		}
	}
	return symbols
}

func clangFunctionSymbol(node *gts.Node, lang *gts.Language, src []byte, pkg, receiver string) *Symbol {
	name := clangFunctionName(node, lang, src)
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
		Signature:     receiver + "::" + name + "(...)",
		Receiver:      receiver,
	}
}

func clangFunctionName(node *gts.Node, lang *gts.Language, src []byte) string {
	decl := namedChildByType(node, lang, "function_declarator")
	if decl == nil {
		decl = node.ChildByFieldName("declarator", lang)
	}
	if decl == nil {
		return ""
	}
	for _, typ := range []string{"identifier", "field_identifier", "destructor_name"} {
		if child := namedChildByType(decl, lang, typ); child != nil {
			return nodeText(child, src)
		}
	}
	if qi := namedChildByType(decl, lang, "qualified_identifier"); qi != nil {
		if child := namedChildByType(qi, lang, "identifier"); child != nil {
			return nodeText(child, src)
		}
	}
	return ""
}
