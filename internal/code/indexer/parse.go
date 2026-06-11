package indexer

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/reponerve/reponerve/internal/code"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
)

const (
	relModuleContainsPackage = "MODULE_CONTAINS_PACKAGE"
	relBelongsToModule       = "BELONGS_TO_MODULE"
	relBelongsToPackage      = "BELONGS_TO_PACKAGE"
	relDefinedInFile         = "DEFINED_IN_FILE"
	relImports               = "IMPORTS"
)

type builder struct {
	repositoryID string
	modulePath   string
	repoPath     string
	indexedAt    time.Time

	entities    []*codemodels.CodeEntity
	rels        []*codemodels.CodeRelationship
	packageIDs  map[string]string
	fileIDs     map[string]string
	relKeys     map[string]struct{}
}

func newBuilder(repositoryID, modulePath, repoPath string, at time.Time) *builder {
	return &builder{
		repositoryID: repositoryID,
		modulePath:   modulePath,
		repoPath:     repoPath,
		indexedAt:    at,
		packageIDs:   make(map[string]string),
		fileIDs:      make(map[string]string),
		relKeys:      make(map[string]struct{}),
	}
}

func (b *builder) addModuleEntity(modulePath, evidenceFile string) string {
	qualified := modulePath
	id := code.EntityID(b.repositoryID, codemodels.EntityTypeModule, qualified)
	b.entities = append(b.entities, &codemodels.CodeEntity{
		ID:            id,
		RepositoryID:  b.repositoryID,
		EntityType:    codemodels.EntityTypeModule,
		Name:          modulePath,
		QualifiedName: qualified,
		FilePath:      evidenceFile,
		ModulePath:    modulePath,
		Language:      "go",
		EvidenceJSON:  marshalEntityEvidence(evidenceFile, 1, 1),
		IndexedAt:     b.indexedAt,
	})
	return id
}

func (b *builder) ensurePackageEntity(packagePath string) string {
	if id, ok := b.packageIDs[packagePath]; ok {
		return id
	}

	qualified := packagePath
	if qualified == "." {
		qualified = b.modulePath
	}

	id := code.EntityID(b.repositoryID, codemodels.EntityTypePackage, qualified)
	entity := &codemodels.CodeEntity{
		ID:            id,
		RepositoryID:  b.repositoryID,
		EntityType:    codemodels.EntityTypePackage,
		Name:          filepath.Base(packagePath),
		QualifiedName: qualified,
		PackagePath:   packagePath,
		ModulePath:    b.modulePath,
		Language:      "go",
		EvidenceJSON:  marshalEntityEvidence(packagePath, 0, 0),
		IndexedAt:     b.indexedAt,
	}
	if packagePath == "." {
		entity.Name = "main"
		entity.FilePath = "."
	} else {
		entity.FilePath = packagePath
	}

	b.entities = append(b.entities, entity)
	b.packageIDs[packagePath] = id
	return id
}

func (b *builder) parseFile(filePath, moduleID string) error {
	absPath := filepath.Join(b.repoPath, filepath.FromSlash(filePath))
	src, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", filePath, err)
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, src, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse %s: %w", filePath, err)
	}

	packagePath := filepath.ToSlash(filepath.Dir(filePath))
	if packagePath == "." {
		packagePath = "."
	}

	packageID := b.ensurePackageEntity(packagePath)
	b.link(relModuleContainsPackage, moduleID, packageID, filePath, "", "", 1)
	b.link(relBelongsToModule, packageID, moduleID, filePath, "", "", 1)

	fileQualified := filePath
	fileID := code.EntityID(b.repositoryID, codemodels.EntityTypeFile, fileQualified)
	startLine := fset.Position(file.Pos()).Line
	endLine := fset.Position(file.End()).Line
	b.entities = append(b.entities, &codemodels.CodeEntity{
		ID:            fileID,
		RepositoryID:  b.repositoryID,
		EntityType:    codemodels.EntityTypeFile,
		Name:          filepath.Base(filePath),
		QualifiedName: fileQualified,
		FilePath:      filePath,
		PackagePath:   packagePath,
		ModulePath:    b.modulePath,
		Language:      "go",
		StartLine:     startLine,
		EndLine:       endLine,
		EvidenceJSON:  marshalEntityEvidence(filePath, startLine, endLine),
		IndexedAt:     b.indexedAt,
	})
	b.fileIDs[filePath] = fileID
	b.link(relBelongsToPackage, fileID, packageID, filePath, "", "", startLine)

	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			if d.Tok != token.TYPE {
				continue
			}
			for _, spec := range d.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				b.addTypeSymbol(file, fset, filePath, packagePath, packageID, fileID, typeSpec)
			}
		case *ast.FuncDecl:
			b.addFuncSymbol(file, fset, filePath, packagePath, packageID, fileID, d)
		}
	}

	for _, imp := range file.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)
		targetPackagePath := resolveImportToPackagePath(b.modulePath, importPath)
		if targetPackagePath == "" {
			continue
		}
		targetID := b.ensurePackageEntity(targetPackagePath)
		line := fset.Position(imp.Pos()).Line
		b.link(relImports, fileID, targetID, filePath, filepath.Base(filePath), targetPackagePath, line)
	}

	return nil
}

func (b *builder) addTypeSymbol(file *ast.File, fset *token.FileSet, filePath, packagePath, packageID, fileID string, spec *ast.TypeSpec) {
	entityType := codemodels.EntityTypeTypeAlias
	switch spec.Type.(type) {
	case *ast.StructType:
		entityType = codemodels.EntityTypeStruct
	case *ast.InterfaceType:
		entityType = codemodels.EntityTypeInterface
	default:
		if spec.Assign.IsValid() {
			entityType = codemodels.EntityTypeTypeAlias
		}
	}

	qualified := symbolQualifiedName(packagePath, spec.Name.Name)
	startLine := fset.Position(spec.Pos()).Line
	endLine := fset.Position(spec.End()).Line

	id := code.EntityID(b.repositoryID, entityType, qualified)
	b.entities = append(b.entities, &codemodels.CodeEntity{
		ID:            id,
		RepositoryID:  b.repositoryID,
		EntityType:    entityType,
		Name:          spec.Name.Name,
		QualifiedName: qualified,
		FilePath:      filePath,
		PackagePath:   packagePath,
		ModulePath:    b.modulePath,
		Language:      "go",
		StartLine:     startLine,
		EndLine:       endLine,
		EvidenceJSON:  marshalEntityEvidence(filePath, startLine, endLine),
		IndexedAt:     b.indexedAt,
	})
	b.link(relBelongsToPackage, id, packageID, filePath, spec.Name.Name, "", startLine)
	b.link(relDefinedInFile, id, fileID, filePath, spec.Name.Name, filepath.Base(filePath), startLine)
}

func (b *builder) addFuncSymbol(file *ast.File, fset *token.FileSet, filePath, packagePath, packageID, fileID string, decl *ast.FuncDecl) {
	startLine := fset.Position(decl.Pos()).Line
	endLine := fset.Position(decl.End()).Line

	entityType := codemodels.EntityTypeFunction
	var qualified, signature string
	if decl.Recv != nil && len(decl.Recv.List) > 0 {
		entityType = codemodels.EntityTypeMethod
		recv := receiverName(decl.Recv.List[0].Type)
		qualified = methodQualifiedName(packagePath, recv, decl.Name.Name)
		signature = formatMethodSignature(recv, decl)
	} else {
		qualified = symbolQualifiedName(packagePath, decl.Name.Name)
		signature = formatFuncSignature(decl)
	}

	id := code.EntityID(b.repositoryID, entityType, qualified)
	b.entities = append(b.entities, &codemodels.CodeEntity{
		ID:            id,
		RepositoryID:  b.repositoryID,
		EntityType:    entityType,
		Name:          decl.Name.Name,
		QualifiedName: qualified,
		FilePath:      filePath,
		PackagePath:   packagePath,
		ModulePath:    b.modulePath,
		Language:      "go",
		StartLine:     startLine,
		EndLine:       endLine,
		Signature:     signature,
		EvidenceJSON:  marshalEntityEvidence(filePath, startLine, endLine),
		IndexedAt:     b.indexedAt,
	})
	b.link(relBelongsToPackage, id, packageID, filePath, decl.Name.Name, "", startLine)
	b.link(relDefinedInFile, id, fileID, filePath, decl.Name.Name, filepath.Base(filePath), startLine)
}

func (b *builder) link(relType, fromID, toID, filePath, fromSymbol, toSymbol string, line int) {
	key := relType + ":" + fromID + ":" + toID
	if _, exists := b.relKeys[key]; exists {
		return
	}
	b.relKeys[key] = struct{}{}

	b.rels = append(b.rels, &codemodels.CodeRelationship{
		ID:               code.RelationshipID(b.repositoryID, relType, fromID, toID),
		RepositoryID:     b.repositoryID,
		FromEntityID:     fromID,
		ToEntityID:       toID,
		RelationshipType: relType,
		EvidenceJSON:     marshalRelationshipEvidence(filePath, relType, fromSymbol, toSymbol, line),
		IndexedAt:        b.indexedAt,
	})
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

func resolveImportToPackagePath(modulePath, importPath string) string {
	if importPath == modulePath {
		return "."
	}
	prefix := modulePath + "/"
	if strings.HasPrefix(importPath, prefix) {
		return strings.TrimPrefix(importPath, prefix)
	}
	return ""
}

func receiverName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		return receiverName(t.X)
	case *ast.Ident:
		return t.Name
	default:
		return "?"
	}
}

func formatFuncSignature(decl *ast.FuncDecl) string {
	return "func " + decl.Name.Name + "(...)"
}

func formatMethodSignature(recv string, decl *ast.FuncDecl) string {
	return fmt.Sprintf("func (%s) %s(...)", recv, decl.Name.Name)
}
