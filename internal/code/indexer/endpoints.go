package indexer

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/reponerve/reponerve/internal/code"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
)

const relExposesEndpoint = "EXPOSES_ENDPOINT"

func (b *builder) extractEndpoints(
	file *ast.File,
	fset *token.FileSet,
	filePath, packagePath, packageID, fileID string,
	_ map[string]string,
) {
	if !detectHTTPImports(file) {
		return
	}

	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name == nil {
			continue
		}
		if !looksLikeHTTPEndpoint(fn.Name.Name) {
			continue
		}
		startLine := fset.Position(fn.Pos()).Line
		endLine := fset.Position(fn.End()).Line
		qualified := symbolQualifiedName(packagePath, fn.Name.Name)
		id := code.EntityID(b.repositoryID, codemodels.EntityTypeEndpoint, qualified)
		b.entities = append(b.entities, &codemodels.CodeEntity{
			ID:            id,
			RepositoryID:  b.repositoryID,
			EntityType:    codemodels.EntityTypeEndpoint,
			Name:          fn.Name.Name,
			QualifiedName: qualified,
			FilePath:      filePath,
			PackagePath:   packagePath,
			ModulePath:    b.modulePath,
			Language:      "go",
			EndpointType:  "http_handler",
			StartLine:     startLine,
			EndLine:       endLine,
			EvidenceJSON:  marshalEntityEvidence(filePath, startLine, endLine),
			IndexedAt:     b.indexedAt,
		})
		b.link(relBelongsToPackage, id, packageID, filePath, fn.Name.Name, "", startLine)
		b.link(relDefinedInFile, id, fileID, filePath, fn.Name.Name, filepathBase(filePath), startLine)
		b.link(relExposesEndpoint, packageID, id, filePath, packagePath, fn.Name.Name, startLine)
	}
}

func looksLikeHTTPEndpoint(name string) bool {
	if strings.HasPrefix(name, "Handle") {
		return true
	}
	if strings.HasSuffix(name, "Handler") {
		return true
	}
	return false
}

func filepathBase(path string) string {
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}

// detectHTTPImports checks raw import paths from the file.
func detectHTTPImports(file *ast.File) bool {
	for _, imp := range file.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		if path == "net/http" {
			return true
		}
	}
	return false
}
