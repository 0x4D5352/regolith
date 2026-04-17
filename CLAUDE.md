# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

For human-facing contributor setup and workflow details, see
[`CONTRIBUTING.md`](CONTRIBUTING.md). This file focuses on context
that isn't obvious from reading the code.

## Project Overview

regolith is a Go CLI tool that visualizes regular expressions as SVG railroad diagrams, JSON AST dumps, and Markdown outlines. It supports 8 regex flavors: JavaScript, Java, .NET, PCRE, POSIX BRE, POSIX ERE, GNU grep BRE, and GNU grep ERE. Each flavor has its own PEG grammar parsed via [pigeon](https://github.com/mna/pigeon), sharing a common AST and renderer.

## Common Commands

```bash
# Build
make build                # Build for current platform
go build ./cmd/regolith   # Alternative direct build

# Test
make test                 # Run all tests with verbose output
go test ./...             # Run tests directly

# Generate parsers from grammars (required after any grammar.peg change)
make generate             # Regenerates ALL flavor parsers
make generate-javascript  # Regenerate a single flavor's parser

# Update golden test files (when intentionally changing SVG output)
make golden               # Shortcut
GOLDEN_UPDATE=1 go test ./internal/renderer/...

# Update output golden files (when intentionally changing JSON/Markdown output)
GOLDEN_UPDATE=1 go test ./internal/output/...

# Update analyzer + renderer goldens together
make golden-analysis

# Other
make coverage             # Test with coverage report
make lint                 # Requires golangci-lint
make fmt                  # Format code
make install              # Install to GOPATH/bin
```

Run the analyzer subcommand:

```bash
regolith analyze 'a.*b.*c'                          # text report (default)
regolith analyze --format json 'a.*b.*c'            # JSON output
regolith analyze --format svg -o out.svg 'a.*b.*c'  # annotated SVG
regolith analyze --benchmark 'a.*b.*c'              # run runtime engines
```

## Architecture

Parse-then-render pipeline: PEG grammar -> AST -> SVG / JSON / Markdown

1. **AST** (`internal/ast/ast.go`):
   - Shared node types used by all flavors: `Regexp`, `Match`, `MatchFragment`, `Literal`, `AnyCharacter`, `Escape`, `Anchor`, `Charset`, `Subexp`, `Repeat`, `BackReference`, `Conditional`, `PatternOption`, `Callout`, `CharsetIntersection`, `CharsetSubtraction`, `CharsetStringDisjunction`, `UnicodePropertyEscape`
   - All nodes implement `Node` interface with `Type() string`

2. **Flavor system** (`internal/flavor/`):
   - `flavor.go` - `Flavor` interface (`Name`, `Description`, `Parse`, `SupportedFlags`, `SupportedFeatures`) + registry (`Register`, `Get`, `List`)
   - Each flavor in its own package (`internal/flavor/{name}/`), each containing:
     - `grammar.peg` - PEG grammar (do NOT edit `parser.go` directly; run `make generate`)
     - `parser.go` - Generated parser (auto-generated, do not edit)
     - `flavor.go` - Flavor struct + `init()` for registry registration
     - `helpers.go` - Parser action helper functions
     - `flavor_test.go` - Parser tests
   - Flavors: `javascript`, `java`, `dotnet`, `pcre`, `posix_bre`, `posix_ere`, `gnugrep_bre`, `gnugrep_ere`

3. **Renderer** (`internal/renderer/`):
   - `renderer.go` - Dispatches AST nodes to specialized render methods
   - `svg.go` - SVG element types with `Render()` methods
   - `layout.go` - Bounding box calculations, `SpaceHorizontally()`, `SpaceVertically()`
   - `styles.go` - Configuration struct with colors, dimensions, fonts

4. **Output formats** (`internal/output/`):
   - `json.go` - AST-to-JSON translation with stable consumer-friendly schema (discriminated union via `type` field)
   - `markdown.go` - AST-to-Markdown nested bullet list for human-readable output
   - `text.go` - Plain-text AST summary (default format in v0.2.0+)
   - `color.go` - Terminal color profile resolution (`--color auto|always|never`) via termenv
   - `analysis_text.go` / `analysis_json.go` - Analyzer report formatters
   - Golden tests in `internal/output/testdata/golden/{json,markdown}/`

5. **CLI** (`cmd/regolith/`):
   - `main.go` - Top-level dispatcher; routes `regolith analyze ...` to `runAnalyze` and everything else to `runRender` **before** pflag parsing, because the two subcommands own separate FlagSets with different defaults
   - `flags.go` - Shared `commonFlags` (`--flavor`, `--format`, `--output`, `--color`, `--theme`, `--padding`, `--font-size`, `--line-width`) and `svgStyleFlags` (color overrides)
   - `render.go` - Default subcommand: parse + emit text/json/svg. `--format` defaults to `text`; `--output ""` means stdout
   - `analyze.go` - Analyzer subcommand with its own flags (`--benchmark`, `--timeout`, `--corpus`, `--sizes`, `--severity`)
   - Blank-imports all flavor packages in `main.go` for side-effect registration

6. **Legacy shim** (`internal/parser/`):
   - `ParseRegex()` delegates to JavaScript flavor for backward compatibility
   - Type aliases (`parser.Regexp = ast.Regexp` etc.) — `*ast.Regexp` and `*parser.Regexp` are interchangeable

7. **Analyzer** (`internal/analyzer/`):
   - `analyzer.go` - `Analyze(root, pattern, flavorName, features)` entry point; single group-metadata pre-pass, then global rules, then recursive per-scope walk
   - `rules.go` + `rules_test.go` - Static-analysis rules (missing anchors, adjacent unbounded quantifiers, overlapping alternatives, invalid backrefs, etc.)
   - `engine.go` + `engine_{grep,node,python,regexp2}.go` - External regex engine adapters used by `--benchmark` to measure real-world runtime behaviour
   - `corpus.go` - Deterministic corpus generation (prose, json, yaml, repeated, random)
   - `benchmark.go` - Benchmark orchestration across engines × corpus × sizes
   - `report.go` - `AnalysisReport` / `Finding` result types

8. **Theme system** (`internal/renderer/theme/`):
   - `theme.go` - `Theme` interface + registry (`Register`, `Get`, `List`); themes register themselves via `init()`
   - Palette files (`catppuccin.go`, `gruvbox.go`, `pastels.go`, `colorblind.go`, `high_contrast.go`) register one or more named themes each
   - `Apply(cfg)` rewrites **only** the color-bearing fields of `renderer.Config` — never dimensions, typography, stroke widths, or severity colors — so style flags can layer on top of a theme

9. **Annotated rendering** (`internal/renderer/annotate.go`):
   - Overlays severity-coloured highlights on the rendered SVG for analyzer findings; used by `regolith analyze --format svg`

10. **Unescape** (`internal/unescape/`):
    - Applies string-literal unescaping (`\\` -> `\`, etc.) before parsing; wired to `--unescape`/`-u`

## Key Patterns

- Flavors register via `init()` in their package; accessed via `flavor.Get("name")`
- Flavor packages must be blank-imported (`_ ".../{flavor}"`) for registration
- `RenderedNode` struct pairs `SVGElement` with `BoundingBox`
- Layout uses anchor points (`AnchorLeft`, `AnchorRight`, `AnchorY`) for connecting elements
- SVG paths use builder pattern via `NewPathBuilder()`
- `FeatureSet` on each flavor declares supported capabilities (lookahead, conditionals, etc.)

## Testing

- Per-flavor parser tests in `internal/flavor/{name}/flavor_test.go`
- Golden file tests in `internal/renderer/testdata/golden/{flavor-slug}/`
  - Newer flavors (Java, .NET, PCRE, GNU grep): strict - error if golden file missing, require `GOLDEN_UPDATE=1` to update
  - Older flavors (POSIX BRE/ERE, base JS): lenient - auto-create missing golden files
- Integration tests in `internal/renderer/integration_test.go` using `validateSVG()` helper
- Output format golden tests in `internal/output/testdata/golden/{json,markdown}/` (strict, `GOLDEN_UPDATE=1`)

## Gotchas

- **Do not use Go's `regexp` package** - its parser is incompatible with most regex flavors (no negative lookbehind, unordered map packing, etc.).
- **Pigeon PEG quirks**: multi-character string predicates need double quotes (`!"&&"` not `!'&&'`); inside `[...]` only `\]`, `\\`, `\n`, `\r`, `\t` are valid escapes (`\[` is NOT valid).
- **Conditional `|` parsing**: In `(?(cond)yes|no)`, the `|` separates branches, not alternation. Grammar must use `Match` (not `Regexp`) for branches.
- **GNUGrepBRE** has unexported `name` field - use `flavor.Get("gnugrep-bre")` instead of direct struct instantiation.
- **`-f` shorthand is taken** by `--flavor` — new flags needing `-f` must use long form only.
- **`internal/parser/` is type aliases** — `parser.Regexp` and `ast.Regexp` are the same type. New code should import `internal/ast` directly.
- **Subcommand dispatch runs before pflag** in `cmd/regolith/main.go`. `regolith analyze` owns a separate FlagSet with different defaults, so the inspection of `args[1] == "analyze"` must happen before any flag parsing.
- **Themes never touch shape/dimension fields**. Only color-bearing fields of `renderer.Config` are rewritten by `Theme.Apply`. Dimensions, typography, stroke widths, and severity colors stay stable across themes so readers can transfer visual cues.
- **Analyzer global vs. per-scope rules**. Rules that check pattern-wide properties (missing anchor, invalid backref target) are invoked once from `Analyze`; rules that need per-branch visibility are invoked inside `walkRegexp`/`walkMatch`/`walkFragment`. Putting a global rule in the per-scope walker produces duplicate findings.
