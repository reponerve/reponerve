package lang_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/reponerve/reponerve/internal/code/lang"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
)

func TestIndexSource_TypeScript(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("..", "indexer", "testdata", "multilang", "frontend", "src", "api.ts"))
	if err != nil {
		t.Fatal(err)
	}
	idx, err := lang.IndexSource(lang.TypeScript, "frontend/src/api.ts", src)
	if err != nil {
		t.Fatal(err)
	}
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeFunction, "frontend/src.getApiBase")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeStruct, "frontend/src.ApiClient")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeMethod, "frontend/src.ApiClient.getUser")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeInterface, "frontend/src.UserStore")
	if len(idx.Imports) == 0 {
		t.Fatal("expected imports")
	}
}

func TestIndexSource_Python(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("..", "indexer", "testdata", "multilang", "services", "app", "handler.py"))
	if err != nil {
		t.Fatal(err)
	}
	idx, err := lang.IndexSource(lang.Python, "services/app/handler.py", src)
	if err != nil {
		t.Fatal(err)
	}
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeFunction, "services/app.health_check")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeStruct, "services/app.Handler")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeMethod, "services/app.Handler.handle")
}

func TestIndexSource_Rust(t *testing.T) {
	src := readTestFile(t, "crates", "api", "src", "lib.rs")
	idx, err := lang.IndexSource(lang.Rust, "crates/api/src/lib.rs", src)
	if err != nil {
		t.Fatal(err)
	}
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeFunction, "crates/api/src.run")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeStruct, "crates/api/src.Service")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeMethod, "crates/api/src.Service.name")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeInterface, "crates/api/src.Store")
}

func TestIndexSource_JavaScript(t *testing.T) {
	src := readTestFile(t, "frontend", "src", "greet.js")
	idx, err := lang.IndexSource(lang.JavaScript, "frontend/src/greet.js", src)
	if err != nil {
		t.Fatal(err)
	}
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeFunction, "frontend/src.greet")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeStruct, "frontend/src.Greeter")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeMethod, "frontend/src.Greeter.say")
}

func TestIndexSource_Java(t *testing.T) {
	src := readTestFile(t, "java", "src", "main", "java", "com", "example", "api", "Handler.java")
	idx, err := lang.IndexSource(lang.Java, "java/src/main/java/com/example/api/Handler.java", src)
	if err != nil {
		t.Fatal(err)
	}
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeStruct, "java/src/main/java/com/example/api.Handler")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeMethod, "java/src/main/java/com/example/api.Handler.health")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeMethod, "java/src/main/java/com/example/api.Handler.names")
}

func TestIndexSource_CSharp(t *testing.T) {
	src := readTestFile(t, "dotnet", "Api", "Handler.cs")
	idx, err := lang.IndexSource(lang.CSharp, "dotnet/Api/Handler.cs", src)
	if err != nil {
		t.Fatal(err)
	}
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeStruct, "dotnet/Api.Handler")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeMethod, "dotnet/Api.Handler.Health")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeMethod, "dotnet/Api.Handler.Run")
}

func TestIndexSource_Ruby(t *testing.T) {
	src := readTestFile(t, "ruby", "lib", "handler.rb")
	idx, err := lang.IndexSource(lang.Ruby, "ruby/lib/handler.rb", src)
	if err != nil {
		t.Fatal(err)
	}
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeStruct, "ruby/lib.Handler")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeMethod, "ruby/lib.Handler.health")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeFunction, "ruby/lib.bootstrap")
}

func TestIndexSource_Kotlin(t *testing.T) {
	src := readTestFile(t, "kotlin", "src", "main", "kotlin", "com", "example", "Handler.kt")
	idx, err := lang.IndexSource(lang.Kotlin, "kotlin/src/main/kotlin/com/example/Handler.kt", src)
	if err != nil {
		t.Fatal(err)
	}
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeStruct, "kotlin/src/main/kotlin/com/example.Handler")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeMethod, "kotlin/src/main/kotlin/com/example.Handler.health")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeInterface, "kotlin/src/main/kotlin/com/example.Store")
}

func TestIndexSource_Swift(t *testing.T) {
	src := readTestFile(t, "swift", "Sources", "Handler.swift")
	idx, err := lang.IndexSource(lang.Swift, "swift/Sources/Handler.swift", src)
	if err != nil {
		t.Fatal(err)
	}
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeStruct, "swift/Sources.Handler")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeMethod, "swift/Sources.Handler.health")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeInterface, "swift/Sources.Store")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeFunction, "swift/Sources.bootstrap")
}

func TestIndexSource_PHP(t *testing.T) {
	src := readTestFile(t, "php", "src", "Handler.php")
	idx, err := lang.IndexSource(lang.PHP, "php/src/Handler.php", src)
	if err != nil {
		t.Fatal(err)
	}
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeStruct, "php/src.Handler")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeMethod, "php/src.Handler.health")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeFunction, "php/src.bootstrap")
}

func TestIndexSource_Cpp(t *testing.T) {
	src := readTestFile(t, "cpp", "src", "handler.cpp")
	idx, err := lang.IndexSource(lang.Cpp, "cpp/src/handler.cpp", src)
	if err != nil {
		t.Fatal(err)
	}
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeStruct, "cpp/src.Handler")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeMethod, "cpp/src.Handler.health")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeFunction, "cpp/src.bootstrap")
}

func TestIndexSource_C(t *testing.T) {
	src := readTestFile(t, "c", "src", "handler.c")
	idx, err := lang.IndexSource(lang.C, "c/src/handler.c", src)
	if err != nil {
		t.Fatal(err)
	}
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeStruct, "c/src.Service")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeFunction, "c/src.run")
}

func TestIndexSource_Scala(t *testing.T) {
	src := readTestFile(t, "scala", "src", "main", "scala", "com", "example", "Handler.scala")
	idx, err := lang.IndexSource(lang.Scala, "scala/src/main/scala/com/example/Handler.scala", src)
	if err != nil {
		t.Fatal(err)
	}
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeStruct, "scala/src/main/scala/com/example.Handler")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeMethod, "scala/src/main/scala/com/example.Handler.health")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeInterface, "scala/src/main/scala/com/example.Store")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeStruct, "scala/src/main/scala/com/example.Bootstrap")
}

func TestIndexSource_Lua(t *testing.T) {
	src := readTestFile(t, "lua", "lib", "handler.lua")
	idx, err := lang.IndexSource(lang.Lua, "lua/lib/handler.lua", src)
	if err != nil {
		t.Fatal(err)
	}
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeStruct, "lua/lib.Handler")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeMethod, "lua/lib.Handler.new")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeMethod, "lua/lib.Handler.health")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeFunction, "lua/lib.bootstrap")
}

func TestIndexSource_Bash(t *testing.T) {
	src := readTestFile(t, "bash", "scripts", "handler.sh")
	idx, err := lang.IndexSource(lang.Bash, "bash/scripts/handler.sh", src)
	if err != nil {
		t.Fatal(err)
	}
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeFunction, "bash/scripts.health_check")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeFunction, "bash/scripts.run_handler")
}

func TestIndexSource_SQL(t *testing.T) {
	src := readTestFile(t, "sql", "schema.sql")
	idx, err := lang.IndexSource(lang.SQL, "sql/schema.sql", src)
	if err != nil {
		t.Fatal(err)
	}
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeStruct, "sql.handlers")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeTypeAlias, "sql.handler_names")
}

func TestIndexSource_Dart(t *testing.T) {
	src := readTestFile(t, "dart", "lib", "handler.dart")
	idx, err := lang.IndexSource(lang.Dart, "dart/lib/handler.dart", src)
	if err != nil {
		t.Fatal(err)
	}
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeStruct, "dart/lib.Handler")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeMethod, "dart/lib.Handler.health")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeMethod, "dart/lib.Handler.run")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeFunction, "dart/lib.bootstrap")
}

func TestIndexSource_Elixir(t *testing.T) {
	src := readTestFile(t, "elixir", "lib", "handler.ex")
	idx, err := lang.IndexSource(lang.Elixir, "elixir/lib/handler.ex", src)
	if err != nil {
		t.Fatal(err)
	}
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeStruct, "elixir/lib.App.Handler")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeFunction, "elixir/lib.health")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeStruct, "elixir/lib.App.Store")
}

func TestIndexSource_Zig(t *testing.T) {
	src := readTestFile(t, "zig", "src", "handler.zig")
	idx, err := lang.IndexSource(lang.Zig, "zig/src/handler.zig", src)
	if err != nil {
		t.Fatal(err)
	}
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeStruct, "zig/src.Handler")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeMethod, "zig/src.Handler.health")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeFunction, "zig/src.bootstrap")
}

func readTestFile(t *testing.T, parts ...string) []byte {
	t.Helper()
	path := append([]string{"..", "indexer", "testdata", "multilang"}, parts...)
	src, err := os.ReadFile(filepath.Join(path...))
	if err != nil {
		t.Fatal(err)
	}
	return src
}

func assertSymbol(t *testing.T, symbols []lang.Symbol, entityType, qualified string) {
	t.Helper()
	for _, sym := range symbols {
		if sym.EntityType == entityType && sym.QualifiedName == qualified {
			return
		}
	}
	t.Fatalf("expected %s symbol %q, got %#v", entityType, qualified, symbols)
}
