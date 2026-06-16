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
		return b.lookupFunction(packagePath, fn.Name)
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
			return b.lookupFunction(pkgPath, sel.Sel.Name)
		}
		if callee := b.lookupMethodByReceiverVar(packagePath, x.Name, sel.Sel.Name); callee != "" {
			return callee
		}
		return b.lookupUniqueMethod(packagePath, sel.Sel.Name)
	case *ast.SelectorExpr:
		if id, ok := x.X.(*ast.Ident); ok {
			if pkgPath, ok := importMap[id.Name]; ok {
				return b.lookupFunction(pkgPath, sel.Sel.Name)
			}
		}
	}
	return ""
}

func (b *builder) lookupFunction(packagePath, name string) string {
	if id, ok := b.funcIndex[symbolQualifiedName(packagePath, name)]; ok {
		return id
	}
	return ""
}

func (b *builder) lookupMethod(packagePath, recvType, methodName string) string {
	if id, ok := b.methodIndex[methodIndexKey(packagePath, recvType, methodName)]; ok {
		return id
	}
	return ""
}

func (b *builder) lookupMethodByReceiverVar(packagePath, varName, methodName string) string {
	if recvType := guessReceiverType(varName); recvType != "" {
		if id := b.lookupMethod(packagePath, recvType, methodName); id != "" {
			return id
		}
	}
	return ""
}

func (b *builder) lookupUniqueMethod(packagePath, methodName string) string {
	suffix := "\x00" + methodName
	var match string
	prefix := packagePath + "\x00"
	for key, id := range b.methodIndex {
		if !strings.HasPrefix(key, prefix) || !strings.HasSuffix(key, suffix) {
			continue
		}
		if match != "" {
			return ""
		}
		match = id
	}
	return match
}

func methodIndexKey(packagePath, recvType, methodName string) string {
	return packagePath + "\x00" + recvType + "\x00" + methodName
}

func guessReceiverType(varName string) string {
	if varName == "" {
		return ""
	}
	runes := []rune(varName)
	runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]
	return string(runes)
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
