# Contributing to regolith

Thanks for your interest in contributing! regolith is a command-line
regex visualizer written in Go. This document describes how to get a
working development environment, how the codebase is organized, and
the conventions for submitting changes.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Ways to Contribute](#ways-to-contribute)
- [Reporting Bugs](#reporting-bugs)
- [Suggesting Enhancements](#suggesting-enhancements)
- [Development Setup](#development-setup)
  - [Prerequisites](#prerequisites)
  - [Building and Testing](#building-and-testing)
  - [Parser Generation](#parser-generation)
  - [Updating Golden Tests](#updating-golden-tests)
- [Project Structure](#project-structure)
- [Coding Conventions](#coding-conventions)
- [Commit Messages](#commit-messages)
- [Pull Request Process](#pull-request-process)

## Code of Conduct

Be respectful, act in good faith, and assume the same of others. This
project does not yet ship a formal Code of Conduct document; until it
does, the [Contributor Covenant](https://www.contributor-covenant.org/)
is the de facto standard to follow.

## Ways to Contribute

- **Report a bug** you encountered while using regolith
- **Suggest an enhancement** — a new flavor, a new analyzer check, a
  new theme, an output format tweak
- **Improve documentation** — README, CONTRIBUTING, inline godoc,
  example patterns
- **Submit a pull request** for a bug fix or feature

Small, focused changes are easier to review than large ones. If you
are planning a significant change, please open an issue first to
discuss the approach before writing code.

## Reporting Bugs

Before filing a bug report, search the existing
[issues](https://github.com/0x4d5352/regolith/issues) to see if
someone has already reported it. When filing a new report, please
include:

- A clear, descriptive title
- The regex pattern that triggered the issue
- The `--flavor` you were using
- The full command line you ran
- What you expected to happen and what actually happened
- The output of `regolith --version`
- Your operating system and Go version
- Any stderr output or stack traces

Reduced test cases are especially helpful — the smallest pattern that
still reproduces the bug.

## Suggesting Enhancements

Enhancement proposals are welcome as GitHub issues. Please include:

- A clear description of the problem the enhancement solves
- A sketch of the proposed solution
- Any alternatives you considered
- Whether you are willing to implement it yourself

## Development Setup

### Prerequisites

- **Go 1.25 or later** — the project tracks the latest stable Go
  release and generally supports the two most recent minor versions
- **[pigeon](https://github.com/mna/pigeon)** — PEG parser generator,
  installed automatically by `make generate` if missing
- **[golangci-lint](https://golangci-lint.run/)** — required for
  `make lint`; not required for day-to-day development but expected
  to pass before a pull request is merged

Clone the repository and build:

```bash
git clone https://github.com/0x4d5352/regolith.git
cd regolith
make build
```

### Building and Testing

```bash
make build                # Build for current platform
make test                 # Run all tests with verbose output
make coverage             # Run tests with coverage report
make lint                 # Run golangci-lint
make fmt                  # Format code
```

Before committing, please run:

```bash
go vet ./...
go fmt ./...
go test ./...
```

If `golangci-lint` is installed, `make lint` should also pass.

### Parser Generation

Each regex flavor is parsed by a
[pigeon](https://github.com/mna/pigeon)-generated PEG parser. The
grammar files live at `internal/flavor/<name>/grammar.peg`. After
modifying any grammar, regenerate the parser:

```bash
make generate             # Regenerate ALL flavor parsers
make generate-javascript  # Regenerate a single flavor's parser
```

Do **not** edit the `parser.go` files directly — they are
auto-generated and will be overwritten on the next run of
`make generate`. All parser logic should live in the `grammar.peg`
file and in `helpers.go` for complex action functions.

Pigeon has a few quirks worth knowing about:
- Multi-character string predicates need double quotes:
  `!"&&"` is valid, `!'&&'` is not
- Inside character classes `[...]`, only `\]`, `\\`, `\n`, `\r`,
  and `\t` are valid escapes — `\[` is **not** valid
- To exclude `[` from a character class, use a negative predicate:
  `!'[' [^\]\\]`

### Updating Golden Tests

regolith uses golden file tests to pin SVG, JSON, and Markdown output.
When you intentionally change the rendered output, regenerate the
fixtures and review the diff carefully before committing — a golden
file update is a form of approval testing.

```bash
# Update SVG golden files after a renderer change
make golden
# or equivalently:
GOLDEN_UPDATE=1 go test ./internal/renderer/...

# Update JSON / Markdown / analysis golden files
GOLDEN_UPDATE=1 go test ./internal/output/...

# Update both in one go
make golden-analysis
```

Golden file tests fall into two categories:
- **Strict** (newer flavors: Java, .NET, PCRE, GNU grep; all analysis
  output): tests fail if the golden file is missing; `GOLDEN_UPDATE=1`
  is required to create or update them
- **Lenient** (older flavors: POSIX BRE/ERE, base JavaScript): missing
  golden files are auto-created on first run

## Project Structure

```
regolith/
├── cmd/regolith/              # CLI entry point
│   ├── main.go                #   main() + subcommand dispatch
│   ├── flags.go               #   Shared commonFlags / svgStyleFlags structs
│   ├── render.go              #   Main render command body
│   └── analyze.go             #   `regolith analyze` subcommand body
├── internal/
│   ├── ast/                   # Shared AST node types
│   │   └── ast.go
│   ├── flavor/                # Flavor interface and registry
│   │   ├── flavor.go
│   │   ├── javascript/        # Each flavor has its own package:
│   │   │   ├── grammar.peg    #   PEG grammar definition
│   │   │   ├── parser.go      #   Generated parser (do not edit)
│   │   │   ├── flavor.go      #   Flavor registration
│   │   │   ├── helpers.go     #   Parser action helpers
│   │   │   └── flavor_test.go #   Parser tests
│   │   ├── java/
│   │   ├── dotnet/
│   │   ├── pcre/
│   │   ├── posix_bre/
│   │   ├── posix_ere/
│   │   ├── gnugrep_bre/
│   │   └── gnugrep_ere/
│   ├── analyzer/              # Static analysis and runtime benchmarking
│   │   ├── analyzer.go        #   Finding detection (catastrophic backtracking, etc.)
│   │   ├── benchmark.go       #   Corpus-driven benchmark harness
│   │   └── engine.go          #   Runtime engine detection per flavor
│   ├── output/                # Text output formats
│   │   ├── json.go            #   AST-to-JSON translation
│   │   ├── markdown.go        #   AST-to-Markdown outline (delegates to text.go)
│   │   ├── text.go            #   Dual-mode RenderText (ANSI or Markdown)
│   │   ├── analysis_text.go   #   Analysis report → ANSI / Markdown
│   │   ├── analysis_json.go   #   Analysis report → JSON
│   │   ├── color.go           #   --color profile resolution
│   │   └── testdata/golden/   #   Golden test files (JSON + Markdown + analysis)
│   ├── renderer/              # SVG rendering
│   │   ├── renderer.go        #   AST-to-SVG dispatch
│   │   ├── svg.go             #   SVG element types
│   │   ├── layout.go          #   Bounding box and positioning
│   │   ├── styles.go          #   Color/dimension configuration
│   │   ├── theme/             #   Theming subpackage
│   │   │   ├── theme.go       #   Theme interface and registry
│   │   │   ├── catppuccin.go  #   Catppuccin palette variants (mocha/macchiato/frappe/latte)
│   │   │   ├── gruvbox.go     #   Gruvbox dark/light
│   │   │   ├── pastels.go     #   Pastels dark/light
│   │   │   ├── high_contrast.go
│   │   │   └── colorblind.go
│   │   └── testdata/golden/   #   Golden test SVGs per flavor
│   ├── parser/                # Legacy shim (delegates to JS flavor)
│   └── unescape/              # String literal unescaping
├── assets/                    # Example SVGs referenced from README.md
├── CLAUDE.md                  # AI-agent instructions (not required reading)
├── CONTRIBUTING.md            # This file
└── README.md                  # User-facing documentation
```

## Coding Conventions

- **Follow idiomatic Go.** Run `go fmt ./...` and `go vet ./...`
  before committing. When in doubt, the
  [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
  and [Effective Go](https://go.dev/doc/effective_go) are the
  authoritative style references.
- **Prefer modern standard-library features**: `any` over
  `interface{}`, `slog` over `log`, `slices`/`maps` over bespoke
  helpers, `errors.Is`/`errors.As` over type assertions.
- **No Go `regexp` package in flavor parsing.** It is incompatible
  with most regex flavors we support (no negative lookbehind,
  unordered alternations, etc.). Stick to the PEG parser for any
  flavor-related matching.
- **Keep changes focused.** Unrelated refactors belong in their own
  commits and pull requests.
- **Comments explain the "why", not the "what".** Only add comments
  when the intent is non-obvious from the code itself. Business
  logic, design tradeoffs, and historical decisions are good
  candidates; plumbing code usually is not.
- **New features need tests.** New flavor support needs parser tests,
  new AST nodes need renderer coverage, new analyzer checks need
  unit tests. Golden files are the right choice for deterministic
  output transformations.
- **Ask before adding dependencies.** The project currently has a
  very small `go.mod`; please open an issue before adding a new
  direct dependency.

## Commit Messages

regolith follows [Conventional Commits](https://www.conventionalcommits.org/)
v1.0.0 for commit messages. The subject line takes the form:

```
<type>[(optional scope)][!]: <description>
```

- **Types**: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`,
  `test`, `build`, `ci`, `chore`
- **Scope** (optional): the module, package, or component name —
  e.g., `feat(renderer)`, `fix(parser)`, `refactor(cmd)`
- **Description**: lowercase imperative mood ("add", not "added"),
  no trailing period, 72 characters or less
- **Breaking changes**: append `!` before the colon and include a
  `BREAKING CHANGE:` footer

Scale the body to the change:

| Change size | Body expectation |
|-------------|------------------|
| Trivial (dep bump, typo) | Subject only, no body |
| Small (single-file fix) | 1–3 sentences covering the *why* |
| Medium (multi-file behavior change) | A paragraph or two: rationale + what changed |
| Large (new module, major feature) | Full narrative: design rationale, per-file breakdown, known limitations, verification |

Wrap body and footer lines at 72 characters. Use backticks around
type names and short code snippets; use indented code blocks for
anything longer than one line.

Example of a small commit:

```
fix(parser): handle empty input without panic

The parser assumed non-empty input in `parse_header`, causing a
panic on empty files. Guard with an early return instead.

Fixes #42
```

## Pull Request Process

1. **Fork the repository** and create a branch from `main`. Branch
   names are up to you; `feat/<short-description>` or
   `fix/<short-description>` are reasonable defaults.
2. **Make your changes** in small, focused commits. Run
   `go vet ./...`, `go fmt ./...`, and `go test ./...` before each
   commit. If you touched any grammar file, run `make generate` and
   commit the regenerated parser alongside.
3. **Update documentation.** If you added a flag, feature, or
   behavior change, update `README.md`. If you changed the
   development workflow, update this file.
4. **Update golden tests** when the change intentionally alters
   rendered output. Review the diff before committing — a golden
   file update is an explicit approval of the new output.
5. **Push your branch** to your fork and open a pull request
   against `main`. The PR title should follow the Conventional
   Commits format (the description for a squash merge).
6. **Describe your change** in the PR body: what problem it solves,
   how you approached it, what tradeoffs you made, and how to
   verify it. Link any related issues.
7. **Respond to review feedback.** Additional commits are welcome
   during review; the branch will be squashed on merge, so history
   cleanliness matters less than clarity of discussion.

Thanks for contributing!
