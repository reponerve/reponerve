package lang

import (
	"fmt"

	gts "github.com/odvcencio/gotreesitter"
	"github.com/odvcencio/gotreesitter/grammars"
)

// IndexSource parses source and extracts deterministic symbols and imports.
func IndexSource(language, filePath string, src []byte) (*FileIndex, error) {
	langObj, err := languageFor(language)
	if err != nil {
		return nil, err
	}
	parser := gts.NewParser(langObj)
	tree, err := parser.Parse(src)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", filePath, err)
	}
	if tree == nil || tree.RootNode() == nil {
		return &FileIndex{Language: language}, nil
	}

	root := tree.RootNode()
	switch language {
	case TypeScript:
		return extractTypeScript(root, langObj, src, filePath), nil
	case JavaScript:
		return extractJavaScript(root, langObj, src, filePath), nil
	case Python:
		return extractPython(root, langObj, src, filePath), nil
	case Rust:
		return extractRust(root, langObj, src, filePath), nil
	case Java:
		return extractJava(root, langObj, src, filePath), nil
	case CSharp:
		return extractCSharp(root, langObj, src, filePath), nil
	case Ruby:
		return extractRuby(root, langObj, src, filePath), nil
	case Kotlin:
		return extractKotlin(root, langObj, src, filePath), nil
	case Swift:
		return extractSwift(root, langObj, src, filePath), nil
	case PHP:
		return extractPHP(root, langObj, src, filePath), nil
	case Cpp:
		return extractCpp(root, langObj, src, filePath), nil
	case C:
		return extractC(root, langObj, src, filePath), nil
	case Scala:
		return extractScala(root, langObj, src, filePath), nil
	case Lua:
		return extractLua(root, langObj, src, filePath), nil
	case Bash:
		return extractBash(root, langObj, src, filePath), nil
	case SQL:
		return extractSQL(root, langObj, src, filePath), nil
	case Dart:
		return extractDart(root, langObj, src, filePath), nil
	case Elixir:
		return extractElixir(root, langObj, src, filePath), nil
	case Zig:
		return extractZig(root, langObj, src, filePath), nil
	default:
		return nil, fmt.Errorf("unsupported language %q", language)
	}
}

func languageFor(language string) (*gts.Language, error) {
	switch language {
	case TypeScript:
		return grammars.TypescriptLanguage(), nil
	case JavaScript:
		return grammars.JavascriptLanguage(), nil
	case Python:
		return grammars.PythonLanguage(), nil
	case Rust:
		return grammars.RustLanguage(), nil
	case Java:
		return grammars.JavaLanguage(), nil
	case CSharp:
		return grammars.CSharpLanguage(), nil
	case Ruby:
		return grammars.RubyLanguage(), nil
	case Kotlin:
		return grammars.KotlinLanguage(), nil
	case Swift:
		return grammars.SwiftLanguage(), nil
	case PHP:
		return grammars.PhpLanguage(), nil
	case Cpp:
		return grammars.CppLanguage(), nil
	case C:
		return grammars.CLanguage(), nil
	case Scala:
		return grammars.ScalaLanguage(), nil
	case Lua:
		return grammars.LuaLanguage(), nil
	case Bash:
		return grammars.BashLanguage(), nil
	case SQL:
		return grammars.SqlLanguage(), nil
	case Dart:
		return grammars.DartLanguage(), nil
	case Elixir:
		return grammars.ElixirLanguage(), nil
	case Zig:
		return grammars.ZigLanguage(), nil
	default:
		return nil, fmt.Errorf("unsupported language %q", language)
	}
}

func packagePathForFile(filePath string) string {
	// Use the full file path (including extension) as the symbol namespace.
	// This guarantees each file is its own namespace, preventing qualified name
	// collisions between any two files — whether they share a directory, a stem
	// (e.g. main.js vs main.css), or any other prefix.
	if filePath == "" {
		return "."
	}
	return filePath
}

func symbolQualifiedName(packagePath, name string) string {
	if packagePath == "." || packagePath == "" {
		return name
	}
	return packagePath + "." + name
}

func methodQualifiedName(packagePath, receiver, name string) string {
	if packagePath == "." || packagePath == "" {
		return receiver + "." + name
	}
	return packagePath + "." + receiver + "." + name
}

func lastSlash(path string) int {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return i
		}
	}
	return -1
}

func line1(point gts.Point) int {
	return int(point.Row) + 1
}

func nodeLines(node *gts.Node) (start, end int) {
	if node == nil {
		return 0, 0
	}
	return line1(node.StartPoint()), line1(node.EndPoint())
}

func nodeText(node *gts.Node, src []byte) string {
	if node == nil {
		return ""
	}
	return node.Text(src)
}

func fieldText(node *gts.Node, lang *gts.Language, field string, src []byte) string {
	child := node.ChildByFieldName(field, lang)
	return nodeText(child, src)
}

func nodeName(node *gts.Node, lang *gts.Language, src []byte) string {
	if name := fieldText(node, lang, "name", src); name != "" {
		return name
	}
	for _, typ := range []string{"identifier", "type_identifier", "property_identifier", "simple_identifier", "constant"} {
		if child := namedChildByType(node, lang, typ); child != nil {
			return nodeText(child, src)
		}
	}
	return ""
}

func namedChildByType(node *gts.Node, lang *gts.Language, typ string) *gts.Node {
	if node == nil {
		return nil
	}
	for i := 0; i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child != nil && child.Type(lang) == typ {
			return child
		}
	}
	return nil
}

func appendUniqueImport(imports []ImportRef, imp ImportRef) []ImportRef {
	for _, existing := range imports {
		if existing.Path == imp.Path && existing.Symbol == imp.Symbol {
			return imports
		}
	}
	return append(imports, imp)
}
