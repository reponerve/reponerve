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
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeFunction, "frontend/src/api.ts.getApiBase")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeStruct, "frontend/src/api.ts.ApiClient")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeMethod, "frontend/src/api.ts.ApiClient.getUser")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeInterface, "frontend/src/api.ts.UserStore")
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
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeFunction, "services/app/handler.py.health_check")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeStruct, "services/app/handler.py.Handler")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeMethod, "services/app/handler.py.Handler.handle")
}

func TestIndexSource_Rust(t *testing.T) {
	src := readTestFile(t, "crates", "api", "src", "lib.rs")
	idx, err := lang.IndexSource(lang.Rust, "crates/api/src/lib.rs", src)
	if err != nil {
		t.Fatal(err)
	}
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeFunction, "crates/api/src/lib.rs.run")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeStruct, "crates/api/src/lib.rs.Service")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeMethod, "crates/api/src/lib.rs.Service.name")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeInterface, "crates/api/src/lib.rs.Store")
}

func TestIndexSource_JavaScript(t *testing.T) {
	src := readTestFile(t, "frontend", "src", "greet.js")
	idx, err := lang.IndexSource(lang.JavaScript, "frontend/src/greet.js", src)
	if err != nil {
		t.Fatal(err)
	}
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeFunction, "frontend/src/greet.js.greet")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeStruct, "frontend/src/greet.js.Greeter")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeMethod, "frontend/src/greet.js.Greeter.say")
}

func TestIndexSource_Java(t *testing.T) {
	src := readTestFile(t, "java", "src", "main", "java", "com", "example", "api", "Handler.java")
	idx, err := lang.IndexSource(lang.Java, "java/src/main/java/com/example/api/Handler.java", src)
	if err != nil {
		t.Fatal(err)
	}
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeStruct, "java/src/main/java/com/example/api/Handler.java.Handler")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeMethod, "java/src/main/java/com/example/api/Handler.java.Handler.health")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeMethod, "java/src/main/java/com/example/api/Handler.java.Handler.names")
}

func TestIndexSource_CSharp(t *testing.T) {
	src := readTestFile(t, "dotnet", "Api", "Handler.cs")
	idx, err := lang.IndexSource(lang.CSharp, "dotnet/Api/Handler.cs", src)
	if err != nil {
		t.Fatal(err)
	}
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeStruct, "dotnet/Api/Handler.cs.Handler")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeMethod, "dotnet/Api/Handler.cs.Handler.Health")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeMethod, "dotnet/Api/Handler.cs.Handler.Run")
}

func TestIndexSource_Ruby(t *testing.T) {
	src := readTestFile(t, "ruby", "lib", "handler.rb")
	idx, err := lang.IndexSource(lang.Ruby, "ruby/lib/handler.rb", src)
	if err != nil {
		t.Fatal(err)
	}
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeStruct, "ruby/lib/handler.rb.Handler")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeMethod, "ruby/lib/handler.rb.Handler.health")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeFunction, "ruby/lib/handler.rb.bootstrap")
}

func TestIndexSource_Kotlin(t *testing.T) {
	src := readTestFile(t, "kotlin", "src", "main", "kotlin", "com", "example", "Handler.kt")
	idx, err := lang.IndexSource(lang.Kotlin, "kotlin/src/main/kotlin/com/example/Handler.kt", src)
	if err != nil {
		t.Fatal(err)
	}
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeStruct, "kotlin/src/main/kotlin/com/example/Handler.kt.Handler")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeMethod, "kotlin/src/main/kotlin/com/example/Handler.kt.Handler.health")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeInterface, "kotlin/src/main/kotlin/com/example/Handler.kt.Store")
}

func TestIndexSource_Swift(t *testing.T) {
	src := readTestFile(t, "swift", "Sources", "Handler.swift")
	idx, err := lang.IndexSource(lang.Swift, "swift/Sources/Handler.swift", src)
	if err != nil {
		t.Fatal(err)
	}
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeStruct, "swift/Sources/Handler.swift.Handler")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeMethod, "swift/Sources/Handler.swift.Handler.health")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeInterface, "swift/Sources/Handler.swift.Store")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeFunction, "swift/Sources/Handler.swift.bootstrap")
}

func TestIndexSource_PHP(t *testing.T) {
	src := readTestFile(t, "php", "src", "Handler.php")
	idx, err := lang.IndexSource(lang.PHP, "php/src/Handler.php", src)
	if err != nil {
		t.Fatal(err)
	}
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeStruct, "php/src/Handler.php.Handler")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeMethod, "php/src/Handler.php.Handler.health")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeFunction, "php/src/Handler.php.bootstrap")
}

func TestIndexSource_Cpp(t *testing.T) {
	src := readTestFile(t, "cpp", "src", "handler.cpp")
	idx, err := lang.IndexSource(lang.Cpp, "cpp/src/handler.cpp", src)
	if err != nil {
		t.Fatal(err)
	}
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeStruct, "cpp/src/handler.cpp.Handler")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeMethod, "cpp/src/handler.cpp.Handler.health")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeFunction, "cpp/src/handler.cpp.bootstrap")
}

func TestIndexSource_C(t *testing.T) {
	src := readTestFile(t, "c", "src", "handler.c")
	idx, err := lang.IndexSource(lang.C, "c/src/handler.c", src)
	if err != nil {
		t.Fatal(err)
	}
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeStruct, "c/src/handler.c.Service")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeFunction, "c/src/handler.c.run")
}

func TestIndexSource_Scala(t *testing.T) {
	src := readTestFile(t, "scala", "src", "main", "scala", "com", "example", "Handler.scala")
	idx, err := lang.IndexSource(lang.Scala, "scala/src/main/scala/com/example/Handler.scala", src)
	if err != nil {
		t.Fatal(err)
	}
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeStruct, "scala/src/main/scala/com/example/Handler.scala.Handler")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeMethod, "scala/src/main/scala/com/example/Handler.scala.Handler.health")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeInterface, "scala/src/main/scala/com/example/Handler.scala.Store")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeStruct, "scala/src/main/scala/com/example/Handler.scala.Bootstrap")
}

func TestIndexSource_Lua(t *testing.T) {
	src := readTestFile(t, "lua", "lib", "handler.lua")
	idx, err := lang.IndexSource(lang.Lua, "lua/lib/handler.lua", src)
	if err != nil {
		t.Fatal(err)
	}
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeStruct, "lua/lib/handler.lua.Handler")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeMethod, "lua/lib/handler.lua.Handler.new")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeMethod, "lua/lib/handler.lua.Handler.health")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeFunction, "lua/lib/handler.lua.bootstrap")
}

func TestIndexSource_Bash(t *testing.T) {
	src := readTestFile(t, "bash", "scripts", "handler.sh")
	idx, err := lang.IndexSource(lang.Bash, "bash/scripts/handler.sh", src)
	if err != nil {
		t.Fatal(err)
	}
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeFunction, "bash/scripts/handler.sh.health_check")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeFunction, "bash/scripts/handler.sh.run_handler")
}

func TestIndexSource_SQL(t *testing.T) {
	src := readTestFile(t, "sql", "schema.sql")
	idx, err := lang.IndexSource(lang.SQL, "sql/schema.sql", src)
	if err != nil {
		t.Fatal(err)
	}
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeStruct, "sql/schema.sql.handlers")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeTypeAlias, "sql/schema.sql.handler_names")
}

func TestIndexSource_Dart(t *testing.T) {
	src := readTestFile(t, "dart", "lib", "handler.dart")
	idx, err := lang.IndexSource(lang.Dart, "dart/lib/handler.dart", src)
	if err != nil {
		t.Fatal(err)
	}
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeStruct, "dart/lib/handler.dart.Handler")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeMethod, "dart/lib/handler.dart.Handler.health")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeMethod, "dart/lib/handler.dart.Handler.run")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeFunction, "dart/lib/handler.dart.bootstrap")
}

func TestIndexSource_Elixir(t *testing.T) {
	src := readTestFile(t, "elixir", "lib", "handler.ex")
	idx, err := lang.IndexSource(lang.Elixir, "elixir/lib/handler.ex", src)
	if err != nil {
		t.Fatal(err)
	}
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeStruct, "elixir/lib/handler.ex.App.Handler")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeFunction, "elixir/lib/handler.ex.health")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeStruct, "elixir/lib/handler.ex.App.Store")
}

func TestIndexSource_Zig(t *testing.T) {
	src := readTestFile(t, "zig", "src", "handler.zig")
	idx, err := lang.IndexSource(lang.Zig, "zig/src/handler.zig", src)
	if err != nil {
		t.Fatal(err)
	}
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeStruct, "zig/src/handler.zig.Handler")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeMethod, "zig/src/handler.zig.Handler.health")
	assertSymbol(t, idx.Symbols, codemodels.EntityTypeFunction, "zig/src/handler.zig.bootstrap")
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
