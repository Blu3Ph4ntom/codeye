// Package scanner provides parallel LoC scanning over git-tracked files.
package scanner

import "strings"

// LangDef defines a programming language and its comment syntax.
type LangDef struct {
	Name           string
	LineComment    []string // e.g. ["//", "#"]
	BlockStart     []string // e.g. ["/*"]
	BlockEnd       []string // e.g. ["*/"]
	DocstringStart []string // e.g. [`"""`] for Python
	DocstringEnd   []string
}

// extensionMap maps lowercase file extension (with dot) to language name.
var extensionMap = map[string]string{
	// Go
	".go": "Go",
	// Rust
	".rs": "Rust",
	// C / C++
	".c": "C", ".h": "C", ".cc": "C++", ".cpp": "C++", ".cxx": "C++",
	".hh": "C++", ".hpp": "C++", ".hxx": "C++",
	// C#
	".cs": "C#",
	// D
	".d": "D",
	// Java
	".java": "Java",
	// Kotlin
	".kt": "Kotlin", ".kts": "Kotlin",
	// Scala
	".scala": "Scala", ".sc": "Scala",
	// JavaScript
	".js": "JavaScript", ".mjs": "JavaScript", ".cjs": "JavaScript",
	".jsx": "JSX",
	// TypeScript
	".ts": "TypeScript", ".mts": "TypeScript", ".cts": "TypeScript",
	".tsx": "TSX",
	// Python
	".py": "Python", ".pyw": "Python", ".pyi": "Python",
	// Ruby
	".rb": "Ruby", ".rake": "Ruby", ".gemspec": "Ruby",
	// PHP
	".php": "PHP", ".phtml": "PHP",
	// Swift
	".swift": "Swift",
	// Objective-C (.m takes priority over MATLAB)
	".mm": "Objective-C",
	// Dart
	".dart": "Dart",
	// Zig
	".zig": "Zig",
	// Elixir
	".ex": "Elixir", ".exs": "Elixir",
	// Erlang
	".erl": "Erlang", ".hrl": "Erlang",
	// Haskell
	".hs": "Haskell", ".lhs": "Haskell",
	// Clojure
	".clj": "Clojure", ".cljs": "Clojure", ".cljc": "Clojure", ".edn": "Clojure",
	// F#
	".fs": "F#", ".fsi": "F#", ".fsx": "F#",
	// OCaml
	".ml": "OCaml", ".mli": "OCaml",
	// Nim
	".nim": "Nim",
	// Crystal
	".cr": "Crystal",
	// Julia
	".jl": "Julia",
	// R
	".r": "R", ".R": "R",
	// Lua
	".lua": "Lua",
	// Perl
	".pl": "Perl", ".pm": "Perl",
	// Shell
	".sh": "Shell", ".bash": "Shell", ".zsh": "Shell", ".fish": "Fish",
	".ps1": "PowerShell", ".psm1": "PowerShell", ".psd1": "PowerShell",
	".bat": "Batch", ".cmd": "Batch",
	// Web
	".html": "HTML", ".htm": "HTML", ".xhtml": "HTML",
	".css":  "CSS",
	".scss": "SCSS", ".sass": "SASS",
	".less":   "Less",
	".svelte": "Svelte",
	".vue":    "Vue",
	".astro":  "Astro",
	// Config
	".yaml": "YAML", ".yml": "YAML",
	".toml":  "TOML",
	".json":  "JSON",
	".json5": "JSON5",
	".xml":   "XML", ".xsd": "XML", ".xsl": "XML",
	".ini": "INI", ".cfg": "INI", ".conf": "INI",
	".hcl": "HCL", ".tf": "HCL", ".tfvars": "HCL",
	".dhall": "Dhall",
	".nix":   "Nix",
	".env":   "Dotenv",
	// Query
	".sql":     "SQL",
	".graphql": "GraphQL", ".gql": "GraphQL",
	// Docs
	".md": "Markdown", ".mdx": "Markdown",
	".rst":  "RST",
	".adoc": "AsciiDoc", ".asciidoc": "AsciiDoc",
	".org": "Org",
	".tex": "LaTeX", ".sty": "LaTeX", ".cls": "LaTeX",
	// Notebooks
	".ipynb": "Jupyter",
	// Build
	".cmake":  "CMake",
	".gradle": "Gradle",
	".bazel":  "Bazel", ".bzl": "Bazel",
	// Misc PL
	".asm": "Assembly", ".s": "Assembly",
	".v": "Verilog", ".sv": "SystemVerilog",
	".vhd": "VHDL", ".vhdl": "VHDL",
	".proto":  "Protobuf",
	".thrift": "Thrift",
	".wasm":   "WebAssembly", ".wat": "WebAssembly",
	".lisp": "Lisp", ".el": "Emacs Lisp", ".scm": "Scheme",
	".rkt":    "Racket",
	".coffee": "CoffeeScript",
	".odin":   "Odin",
}

// filenameMap maps exact filenames (case-sensitive on Linux) to language name.
var filenameMap = map[string]string{
	"Makefile":         "Makefile",
	"makefile":         "Makefile",
	"GNUmakefile":      "Makefile",
	"Dockerfile":       "Dockerfile",
	"Containerfile":    "Dockerfile",
	".dockerfile":      "Dockerfile",
	"Jenkinsfile":      "Groovy",
	"Vagrantfile":      "Ruby",
	"Gemfile":          "Ruby",
	"Rakefile":         "Ruby",
	"Guardfile":        "Ruby",
	"Podfile":          "Ruby",
	"Fastfile":         "Ruby",
	"BUILD":            "Bazel",
	"WORKSPACE":        "Bazel",
	"CMakeLists.txt":   "CMake",
	"meson.build":      "Meson",
	".htaccess":        "Apache Config",
	"nginx.conf":       "Nginx Config",
	"go.mod":           "Go",
	"go.sum":           "Go",
	"Cargo.toml":       "TOML",
	"package.json":     "JSON",
	"tsconfig.json":    "JSON",
	".babelrc":         "JSON",
	".eslintrc":        "JSON",
	"requirements.txt": "Pip Requirements",
	"Pipfile":          "TOML",
	"pyproject.toml":   "TOML",
	"setup.py":         "Python",
	"setup.cfg":        "INI",
	".gitignore":       "Gitignore",
	".gitattributes":   "Gitattributes",
	".editorconfig":    "EditorConfig",
	"LICENSE":          "Text",
	"LICENSE.md":       "Markdown",
	"README":           "Text",
	"CHANGELOG":        "Text",
}

// langDefs maps language name to its comment definition.
var langDefs = map[string]LangDef{
	"Go": {Name: "Go", LineComment: []string{"//"},
		BlockStart: []string{"/*"}, BlockEnd: []string{"*/"}},
	"Rust": {Name: "Rust", LineComment: []string{"//"},
		BlockStart: []string{"/*"}, BlockEnd: []string{"*/"}},
	"C": {Name: "C", LineComment: []string{"//"},
		BlockStart: []string{"/*"}, BlockEnd: []string{"*/"}},
	"C++": {Name: "C++", LineComment: []string{"//"},
		BlockStart: []string{"/*"}, BlockEnd: []string{"*/"}},
	"C#": {Name: "C#", LineComment: []string{"//"},
		BlockStart: []string{"/*"}, BlockEnd: []string{"*/"}},
	"Java": {Name: "Java", LineComment: []string{"//"},
		BlockStart: []string{"/*"}, BlockEnd: []string{"*/"}},
	"Kotlin": {Name: "Kotlin", LineComment: []string{"//"},
		BlockStart: []string{"/*"}, BlockEnd: []string{"*/"}},
	"Scala": {Name: "Scala", LineComment: []string{"//"},
		BlockStart: []string{"/*"}, BlockEnd: []string{"*/"}},
	"JavaScript": {Name: "JavaScript", LineComment: []string{"//"},
		BlockStart: []string{"/*"}, BlockEnd: []string{"*/"}},
	"JSX": {Name: "JSX", LineComment: []string{"//"},
		BlockStart: []string{"/*"}, BlockEnd: []string{"*/"}},
	"TypeScript": {Name: "TypeScript", LineComment: []string{"//"},
		BlockStart: []string{"/*"}, BlockEnd: []string{"*/"}},
	"TSX": {Name: "TSX", LineComment: []string{"//"},
		BlockStart: []string{"/*"}, BlockEnd: []string{"*/"}},
	"Swift": {Name: "Swift", LineComment: []string{"//"},
		BlockStart: []string{"/*"}, BlockEnd: []string{"*/"}},
	"Dart": {Name: "Dart", LineComment: []string{"//"},
		BlockStart: []string{"/*"}, BlockEnd: []string{"*/"}},
	"Zig": {Name: "Zig", LineComment: []string{"//"}},
	"D": {Name: "D", LineComment: []string{"//"},
		BlockStart: []string{"/*"}, BlockEnd: []string{"*/"}},
	"Python": {Name: "Python", LineComment: []string{"#"},
		DocstringStart: []string{`"""`, "'''"}, DocstringEnd: []string{`"""`, "'''"}},
	"Ruby": {Name: "Ruby", LineComment: []string{"#"},
		DocstringStart: []string{"=begin"}, DocstringEnd: []string{"=end"}},
	"PHP": {Name: "PHP", LineComment: []string{"//", "#"},
		BlockStart: []string{"/*"}, BlockEnd: []string{"*/"}},
	"Shell": {Name: "Shell", LineComment: []string{"#"}},
	"Fish":  {Name: "Fish", LineComment: []string{"#"}},
	"PowerShell": {Name: "PowerShell", LineComment: []string{"#"},
		BlockStart: []string{"<#"}, BlockEnd: []string{"#>"}},
	"Batch": {Name: "Batch", LineComment: []string{"REM ", "::"}},
	"Elixir": {Name: "Elixir", LineComment: []string{"#"},
		DocstringStart: []string{`"""`}, DocstringEnd: []string{`"""`}},
	"Erlang": {Name: "Erlang", LineComment: []string{"%"}},
	"Haskell": {Name: "Haskell", LineComment: []string{"--"},
		BlockStart: []string{"{-"}, BlockEnd: []string{"-}"}},
	"Clojure": {Name: "Clojure", LineComment: []string{";"}},
	"F#": {Name: "F#", LineComment: []string{"//"},
		BlockStart: []string{"(*"}, BlockEnd: []string{"*)"}},
	"OCaml": {Name: "OCaml",
		BlockStart: []string{"(*"}, BlockEnd: []string{"*)"}},
	"Lua": {Name: "Lua", LineComment: []string{"--"},
		BlockStart: []string{"--[["}, BlockEnd: []string{"]]"}},
	"Perl": {Name: "Perl", LineComment: []string{"#"}},
	"R":    {Name: "R", LineComment: []string{"#"}},
	"Julia": {Name: "Julia", LineComment: []string{"#"},
		BlockStart: []string{"#="}, BlockEnd: []string{"=#"}},
	"SQL":        {Name: "SQL", LineComment: []string{"--"}, BlockStart: []string{"/*"}, BlockEnd: []string{"*/"}},
	"YAML":       {Name: "YAML", LineComment: []string{"#"}},
	"TOML":       {Name: "TOML", LineComment: []string{"#"}},
	"INI":        {Name: "INI", LineComment: []string{";", "#"}},
	"Makefile":   {Name: "Makefile", LineComment: []string{"#"}},
	"Dockerfile": {Name: "Dockerfile", LineComment: []string{"#"}},
	"HTML": {Name: "HTML",
		BlockStart: []string{"<!--"}, BlockEnd: []string{"-->"}},
	"CSS": {Name: "CSS",
		BlockStart: []string{"/*"}, BlockEnd: []string{"*/"}},
	"SCSS": {Name: "SCSS", LineComment: []string{"//"},
		BlockStart: []string{"/*"}, BlockEnd: []string{"*/"}},
	"GraphQL": {Name: "GraphQL", LineComment: []string{"#"}},
	"Protobuf": {Name: "Protobuf", LineComment: []string{"//"},
		BlockStart: []string{"/*"}, BlockEnd: []string{"*/"}},
	"Markdown": {Name: "Markdown"},
	"RST":      {Name: "RST"},
	"Nim":      {Name: "Nim", LineComment: []string{"#"}},
	"Crystal":  {Name: "Crystal", LineComment: []string{"#"}},
	"Odin":     {Name: "Odin", LineComment: []string{"//"}},
	"Objective-C": {Name: "Objective-C", LineComment: []string{"//"},
		BlockStart: []string{"/*"}, BlockEnd: []string{"*/"}},
	"Assembly":  {Name: "Assembly", LineComment: []string{";", "#"}},
	"HCL":       {Name: "HCL", LineComment: []string{"#", "//"}},
	"Gitignore": {Name: "Gitignore", LineComment: []string{"#"}},
}

// shebangs maps shebang interpreter patterns to language names.
var shebangs = map[string]string{
	"python":  "Python",
	"python3": "Python",
	"ruby":    "Ruby",
	"node":    "JavaScript",
	"nodejs":  "JavaScript",
	"bash":    "Shell",
	"sh":      "Shell",
	"zsh":     "Shell",
	"fish":    "Fish",
	"perl":    "Perl",
	"php":     "PHP",
	"lua":     "Lua",
	"rscript": "R",
	"Rscript": "R",
	"julia":   "Julia",
	"deno":    "TypeScript",
	"ts-node": "TypeScript",
	"awk":     "AWK",
	"gawk":    "AWK",
	"groovy":  "Groovy",
}

// DetectLanguage returns the Language name for a given file path and optional
// first-line content (for shebang detection). Returns "Unknown" if undetected.
func DetectLanguage(path string, firstBytes []byte) string {
	// 1. Exact filename match
	base := fileBase(path)
	if lang, ok := filenameMap[base]; ok {
		return lang
	}

	// 2. Extension match (case-insensitive)
	ext := fileExt(path)
	if ext != "" {
		if lang, ok := extensionMap[strings.ToLower(ext)]; ok {
			return lang
		}
	}

	// 3. Shebang detection
	if len(firstBytes) > 2 && firstBytes[0] == '#' && firstBytes[1] == '!' {
		line := strings.Fields(string(firstBytes))
		if len(line) > 0 {
			interp := line[0][2:] // strip #!
			// handle /usr/bin/env python3
			if strings.HasSuffix(interp, "env") && len(line) > 1 {
				interp = line[1]
			}
			interp = fileBase(interp)
			interp = strings.ToLower(interp)
			if lang, ok := shebangs[interp]; ok {
				return lang
			}
		}
	}

	return "Unknown"
}

// GetLangDef returns the LangDef for a language name (or a generic one for unknown).
func GetLangDef(name string) LangDef {
	if def, ok := langDefs[name]; ok {
		return def
	}
	return LangDef{Name: name}
}

func fileBase(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[i+1:]
		}
	}
	return path
}

func fileExt(path string) string {
	base := fileBase(path)
	for i := len(base) - 1; i >= 0; i-- {
		if base[i] == '.' {
			if i == 0 {
				return "" // dotfiles have no extension
			}
			return base[i:]
		}
	}
	return ""
}
