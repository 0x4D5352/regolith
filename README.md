# regolith

A command-line tool for visualizing regular expressions as SVG diagrams.

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)

## Features

- Parse JavaScript regular expression syntax
- Generate clean SVG diagrams
- Customizable colors and dimensions
- Support for all standard regex features:
  - Literals and alternation
  - Character classes and ranges
  - Quantifiers (`*`, `+`, `?`, `{n}`, `{n,}`, `{n,m}`)
  - Capture groups and non-capture groups
  - Lookahead assertions (`(?=)`, `(?!)`)
  - Anchors (`^`, `$`, `\b`, `\B`)
  - Escape sequences (`\d`, `\w`, `\s`, etc.)
  - Back-references

## Installation

### From Source

```bash
go install github.com/0x4d5352/regolith/cmd/regolith@latest
```

### Build from Repository

```bash
git clone https://github.com/0x4d5352/regolith.git
cd regolith
go build ./cmd/regolith
```

## Usage

### Basic Usage

```bash
# Visualize a regex pattern
regolith 'a|b|c'

# Specify output file
regolith -o output.svg '[a-z]+'

# Read pattern from stdin
echo '^hello$' | regolith
```

### Examples

```bash
# Email pattern
regolith -o email.svg '[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}'

# Phone number
regolith -o phone.svg '\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4}'

# Identifier
regolith -o identifier.svg '[a-zA-Z_][a-zA-Z0-9_]*'

# URL pattern
regolith -o url.svg 'https?://[a-zA-Z0-9.-]+(?:/[a-zA-Z0-9./_-]*)?'
```

### Customization

#### Colors

```bash
regolith -literal-fill '#ff6b6b' -escape-fill '#4ecdc4' 'hello\d+'
```

Available color flags:
- `-text-color` - Text color (default: `#000`)
- `-line-color` - Line/stroke color (default: `#000`)
- `-literal-fill` - Literal box fill (default: `#dae9e5`)
- `-charset-fill` - Character set fill (default: `#cbcbba`)
- `-escape-fill` - Escape sequence fill (default: `#bada55`)
- `-anchor-fill` - Anchor box fill (default: `#6b6659`)
- `-subexp-fill` - Subexpression fill (default: `#e5e5e5`)

#### Dimensions

```bash
regolith -padding 20 -font-size 16 -line-width 3 'pattern'
```

Available dimension flags:
- `-padding` - Padding around diagram (default: `10`)
- `-font-size` - Font size in pixels (default: `14`)
- `-line-width` - Stroke width for lines (default: `2`)

### All Flags

```
regolith - Visualize regular expressions as SVG diagrams

Usage:
  regolith [flags] <pattern>
  echo 'pattern' | regolith [flags]

Flags:
  -anchor-fill string      Anchor box fill color (default "#6b6659")
  -charset-fill string     Character set box fill color (default "#cbcbba")
  -escape-fill string      Escape sequence box fill color (default "#bada55")
  -font-size float         Font size in pixels (default 14)
  -h                       Show help
  -line-color string       Line/stroke color (default "#000")
  -line-width float        Stroke width for lines (default 2)
  -literal-fill string     Literal box fill color (default "#dae9e5")
  -o string                Output file path (default "regex.svg")
  -padding float           Padding around diagram (default 10)
  -subexp-fill string      Subexpression box fill color (default "#e5e5e5")
  -text-color string       Text color (default "#000")
  -v                       Show version
```

## Supported Regex Features

| Feature | Syntax | Example |
|---------|--------|---------|
| Literal | `abc` | `hello` |
| Alternation | `a\|b` | `cat\|dog` |
| Character class | `[abc]` | `[aeiou]` |
| Negated class | `[^abc]` | `[^0-9]` |
| Range | `[a-z]` | `[A-Za-z]` |
| Any character | `.` | `a.b` |
| Zero or more | `*` | `a*` |
| One or more | `+` | `a+` |
| Optional | `?` | `a?` |
| Exact count | `{n}` | `a{3}` |
| Min count | `{n,}` | `a{2,}` |
| Range count | `{n,m}` | `a{2,5}` |
| Non-greedy | `*?` `+?` `??` | `a+?` |
| Capture group | `()` | `(abc)` |
| Non-capture | `(?:)` | `(?:abc)` |
| Positive lookahead | `(?=)` | `(?=abc)` |
| Negative lookahead | `(?!)` | `(?!abc)` |
| Start anchor | `^` | `^start` |
| End anchor | `$` | `end$` |
| Word boundary | `\b` | `\bword\b` |
| Digit | `\d` | `\d+` |
| Word character | `\w` | `\w+` |
| Whitespace | `\s` | `\s+` |
| Back-reference | `\1` | `(a)\1` |

## Development

### Prerequisites

- Go 1.21 or later
- [pigeon](https://github.com/mna/pigeon) for parser generation

### Building

```bash
# Install dependencies
go mod download

# Generate parser (if grammar changes)
go install github.com/mna/pigeon@latest
pigeon -o internal/parser/parser.go internal/parser/grammar.peg

# Build
go build ./cmd/regolith

# Run tests
go test ./...
```

### Project Structure

```
regolith/
├── cmd/regolith/          # CLI entry point
├── internal/
│   ├── parser/         # PEG grammar and AST
│   │   ├── grammar.peg # PEG grammar definition
│   │   ├── ast.go      # AST node types
│   │   └── parser.go   # Generated parser
│   └── renderer/       # SVG rendering
│       ├── svg.go      # SVG element types
│       ├── layout.go   # Bounding box and positioning
│       ├── styles.go   # Configuration
│       └── renderer.go # Main rendering logic
├── testdata/           # Test files and golden outputs
└── README.md
```

## License

MIT License - see LICENSE file for details.

## Acknowledgments

Inspired by [regexper.com](https://regexper.com/) by Jeff Avallone.
