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
		case "call_expression":
			path := extractJSRequirePath(node, lang, src)
			if path != "" {
				start, _ := nodeLines(node)
				out.Imports = appendUniqueImport(out.Imports, ImportRef{Path: path, StartLine: start})
			}
			return gts.WalkContinue
		case "variable_declarator":
			syms := jsVariableDeclaratorSymbol(node, lang, src, pkg)
			out.Symbols = append(out.Symbols, syms...)
			return gts.WalkContinue
		case "assignment_expression":
			syms := jsAssignmentSymbol(node, lang, src, pkg)
			out.Symbols = append(out.Symbols, syms...)
			return gts.WalkContinue
		case "export_statement":
			return gts.WalkContinue
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
		case "method_definition":
			if parent := node.Parent(); parent != nil && parent.Type(lang) != "class_body" {
				methodName := nodeName(node, lang, src)
				if methodName != "" {
					start, end := nodeLines(node)
					out.Symbols = append(out.Symbols, Symbol{
						EntityType:    codemodels.EntityTypeMethod,
						Name:          methodName,
						QualifiedName: symbolQualifiedName(pkg, methodName),
						StartLine:     start,
						EndLine:       end,
						Signature:     methodName + "(...)",
					})
				}
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

func extractJSRequirePath(node *gts.Node, lang *gts.Language, src []byte) string {
	fn := node.ChildByFieldName("function", lang)
	if fn == nil || nodeText(fn, src) != "require" {
		return ""
	}
	args := node.ChildByFieldName("arguments", lang)
	if args == nil {
		return ""
	}
	for i := 0; i < args.NamedChildCount(); i++ {
		child := args.NamedChild(i)
		if child != nil && (child.Type(lang) == "string" || child.Type(lang) == "template_string") {
			return strings.Trim(nodeText(child, src), `"'`+"`")
		}
	}
	return ""
}

func jsVariableDeclaratorSymbol(node *gts.Node, lang *gts.Language, src []byte, pkg string) []Symbol {
	nameNode := node.ChildByFieldName("name", lang)
	valNode := node.ChildByFieldName("value", lang)
	if nameNode == nil || valNode == nil {
		return nil
	}
	valType := valNode.Type(lang)
	if valType != "arrow_function" && valType != "function" && valType != "function_expression" {
		return nil
	}
	name := nodeText(nameNode, src)
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
		Signature:     "const " + name + " = function(...)",
	}}
}

func jsAssignmentSymbol(node *gts.Node, lang *gts.Language, src []byte, pkg string) []Symbol {
	left := node.ChildByFieldName("left", lang)
	right := node.ChildByFieldName("right", lang)
	if left == nil || right == nil {
		return nil
	}
	rightType := right.Type(lang)
	isFuncRight := rightType == "arrow_function" || rightType == "function" || rightType == "function_expression"

	start, end := nodeLines(node)
	leftType := left.Type(lang)
	if leftType == "member_expression" {
		obj := left.ChildByFieldName("object", lang)
		prop := left.ChildByFieldName("property", lang)
		if obj == nil || prop == nil {
			return nil
		}
		objText := nodeText(obj, src)
		propText := nodeText(prop, src)
		if propText == "" {
			return nil
		}

		if objText == "exports" {
			sig := "exports." + propText
			if isFuncRight {
				sig = "exports." + propText + " = function(...)"
			}
			return []Symbol{{
				EntityType:    codemodels.EntityTypeFunction,
				Name:          propText,
				QualifiedName: symbolQualifiedName(pkg, propText),
				StartLine:     start,
				EndLine:       end,
				Signature:     sig,
			}}
		}
		if objText == "module" && propText == "exports" {
			sig := "module.exports = " + nodeText(right, src)
			if isFuncRight {
				sig = "module.exports = function(...)"
			}
			return []Symbol{{
				EntityType:    codemodels.EntityTypeFunction,
				Name:          "default",
				QualifiedName: symbolQualifiedName(pkg, "default"),
				StartLine:     start,
				EndLine:       end,
				Signature:     sig,
			}}
		}
		if !isFuncRight {
			return nil
		}
		// Method assignment on an object / prototype, e.g. app.use = ... or proto.handle = ...
		receiver := objText
		receiver = strings.TrimSuffix(receiver, ".prototype")
		return []Symbol{{
			EntityType:    codemodels.EntityTypeMethod,
			Name:          propText,
			QualifiedName: methodQualifiedName(pkg, receiver, propText),
			StartLine:     start,
			EndLine:       end,
			Signature:     objText + "." + propText + "(...)",
			Receiver:      receiver,
		}}
	} else if leftType == "identifier" && isFuncRight {
		name := nodeText(left, src)
		if name == "" {
			return nil
		}
		return []Symbol{{
			EntityType:    codemodels.EntityTypeFunction,
			Name:          name,
			QualifiedName: symbolQualifiedName(pkg, name),
			StartLine:     start,
			EndLine:       end,
			Signature:     name + " = function(...)",
		}}
	}
	return nil
}
