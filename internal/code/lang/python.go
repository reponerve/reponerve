package lang

import (
	"strings"

	gts "github.com/odvcencio/gotreesitter"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
)

func extractPython(root *gts.Node, lang *gts.Language, src []byte, filePath string) *FileIndex {
	pkg := packagePathForFile(filePath)
	out := &FileIndex{Language: Python}

	gts.Walk(root, func(node *gts.Node, depth int) gts.WalkAction {
		if node == nil {
			return gts.WalkSkipChildren
		}
		switch node.Type(lang) {
		case "import_statement":
			path := strings.TrimSpace(nodeText(node, src))
			path = strings.TrimPrefix(path, "import ")
			if path != "" {
				start, _ := nodeLines(node)
				out.Imports = appendUniqueImport(out.Imports, ImportRef{Path: path, StartLine: start})
			}
			return gts.WalkSkipChildren
		case "import_from_statement":
			path := extractPythonFromImport(node, lang, src)
			if path != "" {
				start, _ := nodeLines(node)
				out.Imports = appendUniqueImport(out.Imports, ImportRef{Path: path, StartLine: start})
			}
			return gts.WalkSkipChildren
		case "function_definition":
			out.Symbols = append(out.Symbols, pyFunctionSymbol(node, lang, src, pkg)...)
			return gts.WalkSkipChildren
		case "class_definition":
			out.Symbols = append(out.Symbols, pyClassSymbols(node, lang, src, pkg)...)
			return gts.WalkSkipChildren
		}
		return gts.WalkContinue
	})

	return out
}

func extractPythonFromImport(node *gts.Node, lang *gts.Language, src []byte) string {
	module := fieldText(node, lang, "module_name", src)
	if module == "" {
		for i := 0; i < node.NamedChildCount(); i++ {
			child := node.NamedChild(i)
			if child != nil && child.Type(lang) == "dotted_name" {
				module = nodeText(child, src)
				break
			}
		}
	}
	return module
}

func pyFunctionSymbol(node *gts.Node, lang *gts.Language, src []byte, pkg string) []Symbol {
	name := fieldText(node, lang, "name", src)
	if name == "" {
		return nil
	}
	start, end := nodeLines(node)
	recv := ""
	entityType := codemodels.EntityTypeFunction
	qualified := symbolQualifiedName(pkg, name)

	// Methods inside a class body are visited separately by class_definition walker.
	for parent := node.Parent(); parent != nil; parent = parent.Parent() {
		if parent.Type(lang) == "class_definition" {
			className := fieldText(parent, lang, "name", src)
			if className != "" {
				entityType = codemodels.EntityTypeMethod
				recv = className
				qualified = methodQualifiedName(pkg, className, name)
			}
			break
		}
	}

	return []Symbol{{
		EntityType:    entityType,
		Name:          name,
		QualifiedName: qualified,
		StartLine:     start,
		EndLine:       end,
		Signature:     "def " + name + "(...)",
		Receiver:      recv,
	}}
}

func pyClassSymbols(node *gts.Node, lang *gts.Language, src []byte, pkg string) []Symbol {
	name := fieldText(node, lang, "name", src)
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
		if child == nil || child.Type(lang) != "function_definition" {
			continue
		}
		methodName := fieldText(child, lang, "name", src)
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
