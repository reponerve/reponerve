package indexer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/reponerve/reponerve/internal/code"
	"github.com/reponerve/reponerve/internal/code/lang"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
)

func (b *builder) parseTreeSitterFile(filePath, language string) error {
	absPath := filepath.Join(b.repoPath, filepath.FromSlash(filePath))
	src, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", filePath, err)
	}

	index, err := lang.IndexSource(language, filePath, src)
	if err != nil {
		return err
	}

	modulePath := langModulePath(b.repoPath, filePath, language)
	moduleID := b.ensureLangModuleEntity(modulePath, language, filePath)
	packagePath := langPackagePath(filePath)
	packageID := b.ensureLangPackageEntity(packagePath, modulePath, language)

	fileID := code.EntityID(b.repositoryID, codemodels.EntityTypeFile, filePath)
	startLine := 1
	endLine := strings.Count(string(src), "\n") + 1
	b.addEntity(&codemodels.CodeEntity{
		ID:            fileID,
		RepositoryID:  b.repositoryID,
		EntityType:    codemodels.EntityTypeFile,
		Name:          filepath.Base(filePath),
		QualifiedName: filePath,
		FilePath:      filePath,
		PackagePath:   packagePath,
		ModulePath:    modulePath,
		Language:      language,
		StartLine:     startLine,
		EndLine:       endLine,
		EvidenceJSON:  marshalEntityEvidenceWithParser(filePath, startLine, endLine, parserForLanguage(language)),
		IndexedAt:     b.indexedAt,
	})
	b.fileIDs[filePath] = fileID
	b.link(relModuleContainsPackage, moduleID, packageID, filePath, "", "", startLine)
	b.link(relBelongsToModule, packageID, moduleID, filePath, "", "", startLine)
	b.link(relBelongsToPackage, fileID, packageID, filePath, "", "", startLine)

	for _, sym := range index.Symbols {
		id := code.EntityID(b.repositoryID, sym.EntityType, sym.QualifiedName)
		b.addEntity(&codemodels.CodeEntity{
			ID:            id,
			RepositoryID:  b.repositoryID,
			EntityType:    sym.EntityType,
			Name:          sym.Name,
			QualifiedName: sym.QualifiedName,
			FilePath:      filePath,
			PackagePath:   packagePath,
			ModulePath:    modulePath,
			Language:      language,
			StartLine:     sym.StartLine,
			EndLine:       sym.EndLine,
			Signature:     sym.Signature,
			EvidenceJSON:  marshalEntityEvidenceWithParser(filePath, sym.StartLine, sym.EndLine, parserForLanguage(language)),
			IndexedAt:     b.indexedAt,
		})
		b.link(relBelongsToPackage, id, packageID, filePath, sym.Name, "", sym.StartLine)
		b.link(relDefinedInFile, id, fileID, filePath, sym.Name, filepath.Base(filePath), sym.StartLine)
		if sym.EntityType == codemodels.EntityTypeFunction {
			b.funcIndex[sym.QualifiedName] = id
		}
		if sym.EntityType == codemodels.EntityTypeMethod && sym.Receiver != "" {
			b.methodIndex[methodIndexKey(packagePath, sym.Receiver, sym.Name)] = id
		}
	}

	for _, imp := range index.Imports {
		targetPackage := importTargetPackage(imp.Path)
		targetID := b.ensureLangPackageEntity(targetPackage, modulePath, language)
		b.link(relImports, fileID, targetID, filePath, filepath.Base(filePath), targetPackage, imp.StartLine)
	}

	return nil
}

func (b *builder) ensureLangModuleEntity(modulePath, language, evidenceFile string) string {
	qualified := language + ":" + modulePath
	id := code.EntityID(b.repositoryID, codemodels.EntityTypeModule, qualified)
	if _, exists := b.langModuleIDs[qualified]; exists {
		return id
	}
	b.langModuleIDs[qualified] = id
	b.addEntity(&codemodels.CodeEntity{
		ID:            id,
		RepositoryID:  b.repositoryID,
		EntityType:    codemodels.EntityTypeModule,
		Name:          modulePath,
		QualifiedName: qualified,
		FilePath:      evidenceFile,
		ModulePath:    modulePath,
		Language:      language,
		EvidenceJSON:  marshalEntityEvidenceWithParser(evidenceFile, 1, 1, parserForLanguage(language)),
		IndexedAt:     b.indexedAt,
	})
	return id
}

func (b *builder) ensureLangPackageEntity(packagePath, modulePath, language string) string {
	qualified := language + ":" + packagePath
	if id, ok := b.langPackageIDs[qualified]; ok {
		return id
	}
	id := code.EntityID(b.repositoryID, codemodels.EntityTypePackage, qualified)
	b.langPackageIDs[qualified] = id
	b.addEntity(&codemodels.CodeEntity{
		ID:            id,
		RepositoryID:  b.repositoryID,
		EntityType:    codemodels.EntityTypePackage,
		Name:          filepath.Base(packagePath),
		QualifiedName: qualified,
		PackagePath:   packagePath,
		ModulePath:    modulePath,
		Language:      language,
		FilePath:      packagePath,
		EvidenceJSON:  marshalEntityEvidenceWithParser(packagePath, 0, 0, parserForLanguage(language)),
		IndexedAt:     b.indexedAt,
	})
	return id
}

func langPackagePath(filePath string) string {
	dir := filepath.ToSlash(filepath.Dir(filePath))
	if dir == "." {
		return "."
	}
	return dir
}

func langModulePath(repoPath, filePath, language string) string {
	relDir := langPackagePath(filePath)
	absDir := filepath.Join(repoPath, filepath.FromSlash(relDir))
	for {
		if langModuleMarker(absDir, language) {
			modRel, err := filepath.Rel(repoPath, absDir)
			if err != nil || modRel == "." {
				return language
			}
			return filepath.ToSlash(modRel)
		}
		parent := filepath.Dir(absDir)
		if parent == absDir || !strings.HasPrefix(absDir, repoPath) {
			break
		}
		absDir = parent
	}
	if relDir == "." {
		return language
	}
	top := relDir
	if i := strings.Index(top, "/"); i >= 0 {
		top = top[:i]
	}
	return top
}

func langModuleMarker(dir, language string) bool {
	switch language {
	case lang.TypeScript, lang.JavaScript:
		_, err := os.Stat(filepath.Join(dir, "package.json"))
		return err == nil
	case lang.Python:
		for _, name := range []string{"pyproject.toml", "setup.py", "requirements.txt"} {
			if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
				return true
			}
		}
		return false
	case lang.Rust:
		_, err := os.Stat(filepath.Join(dir, "Cargo.toml"))
		return err == nil
	case lang.Java, lang.Kotlin:
		for _, name := range []string{"pom.xml", "build.gradle", "build.gradle.kts"} {
			if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
				return true
			}
		}
		return false
	case lang.CSharp:
		matches, _ := filepath.Glob(filepath.Join(dir, "*.csproj"))
		return len(matches) > 0
	case lang.Ruby:
		_, err := os.Stat(filepath.Join(dir, "Gemfile"))
		return err == nil
	case lang.Swift:
		for _, name := range []string{"Package.swift", "Package.resolved"} {
			if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
				return true
			}
		}
		return false
	case lang.PHP:
		_, err := os.Stat(filepath.Join(dir, "composer.json"))
		return err == nil
	case lang.C, lang.Cpp:
		for _, name := range []string{"CMakeLists.txt", "Makefile", "meson.build"} {
			if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
				return true
			}
		}
		return false
	case lang.Scala:
		for _, name := range []string{"build.sbt", "build.sc"} {
			if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
				return true
			}
		}
		return false
	case lang.Lua:
		for _, name := range []string{".luarc.json", "rockspec"} {
			if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
				return true
			}
		}
		matches, _ := filepath.Glob(filepath.Join(dir, "*.rockspec"))
		return len(matches) > 0
	case lang.Bash:
		return false
	case lang.SQL:
		return false
	case lang.Dart:
		_, err := os.Stat(filepath.Join(dir, "pubspec.yaml"))
		return err == nil
	case lang.Elixir:
		_, err := os.Stat(filepath.Join(dir, "mix.exs"))
		return err == nil
	case lang.Zig:
		for _, name := range []string{"build.zig", "build.zig.zon"} {
			if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func importTargetPackage(importPath string) string {
	importPath = strings.TrimSpace(importPath)
	importPath = strings.Trim(importPath, `"'`)
	importPath = strings.TrimPrefix(importPath, "./")
	if importPath == "" {
		return "external"
	}
	return importPath
}

func parserForLanguage(language string) string {
	return "tree-sitter/" + language
}
