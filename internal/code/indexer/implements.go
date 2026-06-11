package indexer

import (
	"go/ast"
	"go/token"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
)

const (
	relImplements = "IMPLEMENTS"
	relDependsOn  = "DEPENDS_ON"
)

func (b *builder) extractImplementsAssertions(file *ast.File, fset *token.FileSet, filePath, packagePath string) {
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.VAR {
			continue
		}
		for _, spec := range gen.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok || len(valueSpec.Values) == 0 {
				continue
			}
			interfaceName := typeIdentName(valueSpec.Type)
			structName := structNameFromNilConversion(valueSpec.Values[0])
			if interfaceName == "" || structName == "" {
				continue
			}
			structID := b.lookupTypeEntity(packagePath, structName, codemodels.EntityTypeStruct)
			interfaceID := b.lookupTypeEntity(packagePath, interfaceName, codemodels.EntityTypeInterface)
			if structID == "" || interfaceID == "" {
				continue
			}
			line := fset.Position(valueSpec.Pos()).Line
			b.link(relImplements, structID, interfaceID, filePath, structName, interfaceName, line)
		}
	}
}

func (b *builder) extractStructDependencies(file *ast.File, fset *token.FileSet, filePath, packagePath string, spec *ast.TypeSpec, structID string) {
	st, ok := spec.Type.(*ast.StructType)
	if !ok || st.Fields == nil {
		return
	}
	for _, field := range st.Fields.List {
		typeName := typeIdentName(field.Type)
		if typeName == "" {
			continue
		}
		depID := b.lookupTypeEntity(packagePath, typeName, codemodels.EntityTypeStruct, codemodels.EntityTypeInterface, codemodels.EntityTypeTypeAlias)
		if depID == "" || depID == structID {
			continue
		}
		line := fset.Position(field.Pos()).Line
		b.link(relDependsOn, structID, depID, filePath, spec.Name.Name, typeName, line)
	}
}

func (b *builder) lookupTypeEntity(packagePath, name string, types ...string) string {
	allowed := make(map[string]bool, len(types))
	for _, t := range types {
		allowed[t] = true
	}
	for _, e := range b.entities {
		if e.PackagePath != packagePath || e.Name != name {
			continue
		}
		if len(types) == 0 || allowed[e.EntityType] {
			return e.ID
		}
	}
	return ""
}

func typeIdentName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return typeIdentName(t.X)
	case *ast.SelectorExpr:
		if id, ok := t.X.(*ast.Ident); ok {
			return id.Name + "." + t.Sel.Name
		}
	}
	return ""
}

func structNameFromNilConversion(expr ast.Expr) string {
	call, ok := expr.(*ast.CallExpr)
	if !ok || len(call.Args) != 1 {
		return ""
	}
	if ident, ok := call.Args[0].(*ast.Ident); !ok || ident.Name != "nil" {
		return ""
	}
	paren, ok := call.Fun.(*ast.ParenExpr)
	if !ok {
		return ""
	}
	star, ok := paren.X.(*ast.StarExpr)
	if !ok {
		return ""
	}
	return typeIdentName(star.X)
}
