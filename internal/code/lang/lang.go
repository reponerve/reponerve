package lang

import "strings"

const (
	Go         = "go"
	JavaScript = "javascript"
	TypeScript = "typescript"
	Python     = "python"
	Rust       = "rust"
	Java       = "java"
	CSharp     = "csharp"
	Ruby       = "ruby"
	Kotlin     = "kotlin"
	Swift      = "swift"
	PHP        = "php"
	C          = "c"
	Cpp        = "cpp"
	Scala      = "scala"
	Lua        = "lua"
	Bash       = "bash"
	SQL        = "sql"
	Dart       = "dart"
	Elixir     = "elixir"
	Zig        = "zig"
)

// SupportedTreeSitterLanguages lists languages indexed via Tree-sitter (excluding Go).
var SupportedTreeSitterLanguages = []string{
	JavaScript, TypeScript, Python, Rust, Java, CSharp, Ruby, Kotlin,
	Swift, PHP, C, Cpp, Scala, Lua, Bash, SQL, Dart, Elixir, Zig,
}

// Detect returns the language for a repository-relative file path, or "" if unsupported.
func Detect(filePath string) string {
	lower := strings.ToLower(filePath)
	switch {
	case strings.HasSuffix(lower, ".go"):
		return Go
	case strings.HasSuffix(lower, ".ts"), strings.HasSuffix(lower, ".tsx"):
		return TypeScript
	case strings.HasSuffix(lower, ".js"), strings.HasSuffix(lower, ".jsx"),
		strings.HasSuffix(lower, ".mjs"), strings.HasSuffix(lower, ".cjs"):
		return JavaScript
	case strings.HasSuffix(lower, ".py"):
		return Python
	case strings.HasSuffix(lower, ".rs"):
		return Rust
	case strings.HasSuffix(lower, ".java"):
		return Java
	case strings.HasSuffix(lower, ".cs"):
		return CSharp
	case strings.HasSuffix(lower, ".rb"):
		return Ruby
	case strings.HasSuffix(lower, ".kt"), strings.HasSuffix(lower, ".kts"):
		return Kotlin
	case strings.HasSuffix(lower, ".swift"):
		return Swift
	case strings.HasSuffix(lower, ".php"):
		return PHP
	case strings.HasSuffix(lower, ".scala"), strings.HasSuffix(lower, ".sc"):
		return Scala
	case strings.HasSuffix(lower, ".cpp"), strings.HasSuffix(lower, ".cc"),
		strings.HasSuffix(lower, ".cxx"), strings.HasSuffix(lower, ".hpp"),
		strings.HasSuffix(lower, ".hh"), strings.HasSuffix(lower, ".hxx"):
		return Cpp
	case strings.HasSuffix(lower, ".c"), strings.HasSuffix(lower, ".h"):
		return C
	case strings.HasSuffix(lower, ".lua"):
		return Lua
	case strings.HasSuffix(lower, ".sh"), strings.HasSuffix(lower, ".bash"):
		return Bash
	case strings.HasSuffix(lower, ".sql"):
		return SQL
	case strings.HasSuffix(lower, ".dart"):
		return Dart
	case strings.HasSuffix(lower, ".ex"), strings.HasSuffix(lower, ".exs"):
		return Elixir
	case strings.HasSuffix(lower, ".zig"):
		return Zig
	default:
		return ""
	}
}

// IsTreeSitterLanguage reports whether the language is indexed with Tree-sitter.
func IsTreeSitterLanguage(language string) bool {
	switch language {
	case JavaScript, TypeScript, Python, Rust, Java, CSharp, Ruby, Kotlin,
		Swift, PHP, C, Cpp, Scala, Lua, Bash, SQL, Dart, Elixir, Zig:
		return true
	default:
		return false
	}
}

// IsIndexable reports whether the file should be indexed.
func IsIndexable(filePath string) bool {
	return Detect(filePath) != "" && !isSkippedFile(filePath)
}

func isSkippedFile(filePath string) bool {
	base := filePath
	if i := lastSlash(filePath); i >= 0 {
		base = filePath[i+1:]
	}
	lower := strings.ToLower(base)
	if strings.HasSuffix(lower, "_test.go") {
		return true
	}
	if strings.HasPrefix(lower, "test_") && strings.HasSuffix(lower, ".py") {
		return true
	}
	return false
}
