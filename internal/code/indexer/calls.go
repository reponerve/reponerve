package indexer

import (
	"go/ast"
	"go/token"
	"strings"
)

const relCalls = "CALLS"

type typeRef struct {
	packagePath string
	typeName    string
}

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

	varTypes := b.collectVarTypes(decl, packagePath, importMap)

	ast.Inspect(decl.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		calleeID := b.resolveCallee(call, packagePath, importMap, varTypes)
		if calleeID == "" || calleeID == callerID {
			return true
		}
		line := fset.Position(call.Pos()).Line
		b.link(relCalls, callerID, calleeID, filePath, decl.Name.Name, "", line)
		return true
	})
}

func (b *builder) resolveCallee(call *ast.CallExpr, packagePath string, importMap map[string]string, varTypes map[string]typeRef) string {
	switch fn := call.Fun.(type) {
	case *ast.Ident:
		return b.lookupFunction(packagePath, fn.Name)
	case *ast.SelectorExpr:
		return b.resolveSelectorCallee(fn, packagePath, importMap, varTypes)
	default:
		return ""
	}
}

func (b *builder) resolveSelectorCallee(sel *ast.SelectorExpr, packagePath string, importMap map[string]string, varTypes map[string]typeRef) string {
	methodName := sel.Sel.Name
	switch x := sel.X.(type) {
	case *ast.Ident:
		// 1. Check if x is an imported package name (e.g. paths.AddTrailingSlash)
		if pkgPath, ok := importMap[x.Name]; ok {
			if fn := b.lookupFunction(pkgPath, methodName); fn != "" {
				return fn
			}
		}
		// 2. Check if x is a variable with known type in varTypes
		if tRef, ok := varTypes[x.Name]; ok && tRef.typeName != "" {
			if callee := b.lookupMethod(tRef.packagePath, tRef.typeName, methodName); callee != "" {
				return callee
			}
		}
		// 3. Fallback: receiver variable heuristic in current package
		if callee := b.lookupMethodByReceiverVar(packagePath, x.Name, methodName); callee != "" {
			return callee
		}
		// 4. Check if methodName is unique in imported packages
		if callee := b.lookupUniqueMethodInImports(importMap, methodName); callee != "" {
			return callee
		}
		// 5. Check if methodName is unique in current package
		if callee := b.lookupUniqueMethod(packagePath, methodName); callee != "" {
			return callee
		}
		// 6. Global unique method across the whole repository
		return b.lookupGlobalUniqueMethod(methodName)

	case *ast.SelectorExpr:
		// e.g. s.store.Save or pkg.Sub.Fn or repo.New().Find()
		if innerIdent, ok := x.X.(*ast.Ident); ok {
			if pkgPath, ok := importMap[innerIdent.Name]; ok {
				if fn := b.lookupFunction(pkgPath, methodName); fn != "" {
					return fn
				}
			}
			if tRef, ok := varTypes[innerIdent.Name]; ok {
				fieldKey := tRef.packagePath + "." + tRef.typeName + "." + x.Sel.Name
				if fieldType, ok := b.structFields[fieldKey]; ok && fieldType.typeName != "" {
					if callee := b.lookupMethod(fieldType.packagePath, fieldType.typeName, methodName); callee != "" {
						return callee
					}
				}
			}
		}
		if callee := b.lookupUniqueMethodInImports(importMap, methodName); callee != "" {
			return callee
		}
		return b.lookupGlobalUniqueMethod(methodName)
	}
	return ""
}

func (b *builder) collectVarTypes(decl *ast.FuncDecl, packagePath string, importMap map[string]string) map[string]typeRef {
	varTypes := make(map[string]typeRef)

	// 1. Receiver
	if decl.Recv != nil && len(decl.Recv.List) > 0 {
		recv := decl.Recv.List[0]
		typeName := receiverName(recv.Type)
		for _, name := range recv.Names {
			if name != nil {
				varTypes[name.Name] = typeRef{packagePath: packagePath, typeName: typeName}
			}
		}
	}

	// 2. Parameters
	if decl.Type.Params != nil {
		for _, param := range decl.Type.Params.List {
			tRef := resolveTypeExpr(param.Type, packagePath, importMap)
			if tRef.typeName != "" {
				for _, name := range param.Names {
					if name != nil {
						varTypes[name.Name] = tRef
					}
				}
			}
		}
	}

	// 3. Local declarations and assignments in body
	if decl.Body != nil {
		ast.Inspect(decl.Body, func(n ast.Node) bool {
			switch stmt := n.(type) {
			case *ast.DeclStmt:
				if genDecl, ok := stmt.Decl.(*ast.GenDecl); ok {
					for _, spec := range genDecl.Specs {
						if valSpec, ok := spec.(*ast.ValueSpec); ok && valSpec.Type != nil {
							tRef := resolveTypeExpr(valSpec.Type, packagePath, importMap)
							if tRef.typeName != "" {
								for _, name := range valSpec.Names {
									if name != nil {
										varTypes[name.Name] = tRef
									}
								}
							}
						}
					}
				}
			case *ast.AssignStmt:
				for i, lhs := range stmt.Lhs {
					ident, ok := lhs.(*ast.Ident)
					if !ok || ident == nil {
						continue
					}
					if i < len(stmt.Rhs) {
						tRef := inferExprType(stmt.Rhs[i], packagePath, importMap)
						if tRef.typeName != "" {
							varTypes[ident.Name] = tRef
						}
					}
				}
			}
			return true
		})
	}

	return varTypes
}

func resolveTypeExpr(expr ast.Expr, packagePath string, importMap map[string]string) typeRef {
	switch t := expr.(type) {
	case *ast.StarExpr:
		return resolveTypeExpr(t.X, packagePath, importMap)
	case *ast.Ident:
		return typeRef{packagePath: packagePath, typeName: t.Name}
	case *ast.SelectorExpr:
		if pkgIdent, ok := t.X.(*ast.Ident); ok {
			if pkgPath, ok := importMap[pkgIdent.Name]; ok {
				return typeRef{packagePath: pkgPath, typeName: t.Sel.Name}
			}
		}
	}
	return typeRef{}
}

func inferExprType(expr ast.Expr, packagePath string, importMap map[string]string) typeRef {
	switch e := expr.(type) {
	case *ast.UnaryExpr:
		return inferExprType(e.X, packagePath, importMap)
	case *ast.CompositeLit:
		return resolveTypeExpr(e.Type, packagePath, importMap)
	case *ast.TypeAssertExpr:
		return resolveTypeExpr(e.Type, packagePath, importMap)
	case *ast.CallExpr:
		if ident, ok := e.Fun.(*ast.Ident); ok && ident.Name == "new" && len(e.Args) == 1 {
			return resolveTypeExpr(e.Args[0], packagePath, importMap)
		}
		switch fun := e.Fun.(type) {
		case *ast.Ident:
			if strings.HasPrefix(fun.Name, "New") && len(fun.Name) > 3 {
				return typeRef{packagePath: packagePath, typeName: strings.TrimPrefix(fun.Name, "New")}
			}
		case *ast.SelectorExpr:
			if pkgIdent, ok := fun.X.(*ast.Ident); ok {
				pkgPath := importMap[pkgIdent.Name]
				if strings.HasPrefix(fun.Sel.Name, "New") {
					typeName := strings.TrimPrefix(fun.Sel.Name, "New")
					if typeName == "" {
						typeName = guessReceiverType(pkgIdent.Name)
					}
					return typeRef{packagePath: pkgPath, typeName: typeName}
				}
			}
		}
	}
	return typeRef{}
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

func (b *builder) lookupUniqueMethodInImports(importMap map[string]string, methodName string) string {
	suffix := "\x00" + methodName
	var match string
	for _, pkgPath := range importMap {
		prefix := pkgPath + "\x00"
		for key, id := range b.methodIndex {
			if strings.HasPrefix(key, prefix) && strings.HasSuffix(key, suffix) {
				if match != "" && match != id {
					return "" // ambiguous
				}
				match = id
			}
		}
	}
	return match
}

func (b *builder) lookupGlobalUniqueMethod(methodName string) string {
	suffix := "\x00" + methodName
	var match string
	for key, id := range b.methodIndex {
		if strings.HasSuffix(key, suffix) {
			if match != "" && match != id {
				return "" // ambiguous
			}
			match = id
		}
	}
	return match
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
