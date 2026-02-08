# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

regolith is a Go CLI tool that visualizes regular expressions as SVG railroad diagrams. It supports 8 regex flavors: JavaScript, Java, .NET, PCRE, POSIX BRE, POSIX ERE, GNU grep BRE, and GNU grep ERE. Each flavor has its own PEG grammar parsed via [pigeon](https://github.com/mna/pigeon), sharing a common AST and renderer.

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

# Other
make coverage             # Test with coverage report
make lint                 # Requires golangci-lint
make fmt                  # Format code
make install              # Install to GOPATH/bin
```

## Architecture

Parse-then-render pipeline: PEG grammar -> AST -> SVG

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

4. **CLI** (`cmd/regolith/main.go`):
   - Flag parsing, stdin/arg input, `-flavor` flag for flavor selection, error display
   - Blank-imports all flavor packages for side-effect registration

5. **Legacy shim** (`internal/parser/`):
   - `ParseRegex()` delegates to JavaScript flavor for backward compatibility

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

## Gotchas

- **Do not use Go's `regexp` package** - its parser is incompatible with most regex flavors (no negative lookbehind, unordered map packing, etc.).
- **Pigeon PEG quirks**: multi-character string predicates need double quotes (`!"&&"` not `!'&&'`); inside `[...]` only `\]`, `\\`, `\n`, `\r`, `\t` are valid escapes (`\[` is NOT valid).
- **Conditional `|` parsing**: In `(?(cond)yes|no)`, the `|` separates branches, not alternation. Grammar must use `Match` (not `Regexp`) for branches.
- **GNUGrepBRE** has unexported `name` field - use `flavor.Get("gnugrep-bre")` instead of direct struct instantiation.
