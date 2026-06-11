package indexer

import (
	"go/ast"
	"go/token"
	"strings"
)

const relCalls = "CALLS"

func (b *builder) extractCalls(
	file *ast.File,
	fset *token.FileSet,
	filePath, packagePath string,
	callerID string,
	decl *ast.FuncDecl,
	importMap map[string]string,
) {
	if decl.Body == nil {
		return
	}

	ast.Inspect(decl.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		calleeID := b.resolveCallee(call, packagePath, importMap)
		if calleeID == "" || calleeID == callerID {
			return true
		}
		line := fset.Position(call.Pos()).Line
		b.link(relCalls, callerID, calleeID, filePath, decl.Name.Name, "", line)
		return true
	})
}

func (b *builder) resolveCallee(call *ast.CallExpr, packagePath string, importMap map[string]string) string {
	switch fn := call.Fun.(type) {
	case *ast.Ident:
		return b.lookupFunctionInPackage(packagePath, fn.Name)
	case *ast.SelectorExpr:
		return b.resolveSelectorCallee(fn, packagePath, importMap)
	default:
		return ""
	}
}

func (b *builder) resolveSelectorCallee(sel *ast.SelectorExpr, packagePath string, importMap map[string]string) string {
	switch x := sel.X.(type) {
	case *ast.Ident:
		if pkgPath, ok := importMap[x.Name]; ok {
			return b.lookupFunctionInPackage(pkgPath, sel.Sel.Name)
		}
		if x.Name == "" {
			return ""
		}
		// Same-package selector on local identifier (e.g. store.New).
		return b.lookupFunctionInPackage(packagePath, sel.Sel.Name)
	case *ast.SelectorExpr:
		// Chained selector — resolve only the final call name in the import package when possible.
		if id, ok := x.X.(*ast.Ident); ok {
			if pkgPath, ok := importMap[id.Name]; ok {
				return b.lookupFunctionInPackage(pkgPath, sel.Sel.Name)
			}
		}
	}
	return ""
}

func (b *builder) lookupFunctionInPackage(packagePath, name string) string {
	if id, ok := b.funcIndex[funcIndexKey(packagePath, name)]; ok {
		return id
	}
	return ""
}

func funcIndexKey(packagePath, name string) string {
	return packagePath + "\x00" + name
}

func buildImportMap(file *ast.File, modulePath string) map[string]string {
	imports := make(map[string]string)
	for _, imp := range file.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)
		localName := ""
		if imp.Name != nil {
			localName = imp.Name.Name
		} else {
			parts := strings.Split(importPath, "/")
			localName = parts[len(parts)-1]
		}
		pkgPath := resolveImportToPackagePath(modulePath, importPath)
		if pkgPath == "" {
			continue
		}
		imports[localName] = pkgPath
	}
	return imports
}
