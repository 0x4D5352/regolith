package output

import (
	"fmt"
	"strings"

	"github.com/0x4d5352/regolith/internal/ast"
)

var flavorDisplayNames = map[string]string{
	"javascript":  "JavaScript",
	"java":        "Java",
	"dotnet":      ".NET",
	"pcre":        "PCRE",
	"posix-bre":   "POSIX BRE",
	"posix-ere":   "POSIX ERE",
	"gnugrep-bre": "GNU grep BRE",
	"gnugrep-ere": "GNU grep ERE",
}

func formatFlavorName(name string) string {
	if display, ok := flavorDisplayNames[name]; ok {
		return display
	}
	return name
}

var anchorDescriptions = map[string]string{
	ast.AnchorStart:                   "Start of line",
	ast.AnchorEnd:                     "End of line",
	ast.AnchorWordBoundary:            "Word boundary",
	ast.AnchorNonWordBoundary:         "Non-word boundary",
	ast.AnchorStringStart:             "Start of string",
	ast.AnchorStringEnd:               "End of string",
	ast.AnchorAbsoluteEnd:             "Absolute end of string",
	ast.AnchorWordStart:               "Start of word",
	ast.AnchorWordEnd:                 "End of word",
	ast.AnchorGraphemeClusterBoundary: "Grapheme cluster boundary",
}

type markdownWriter struct {
	buf *strings.Builder
}

func (w *markdownWriter) line(indent int, text string) {
	for range indent {
		w.buf.WriteString("  ")
	}
	w.buf.WriteString("- ")
	w.buf.WriteString(text)
	w.buf.WriteByte('\n')
}

// RenderMarkdown converts a parsed AST into a Markdown outline describing the regex structure.
func RenderMarkdown(root *ast.Regexp, pattern, flavorName string) string {
	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("# Regex: `%s`\n\n**Flavor:** %s\n\n", pattern, formatFlavorName(flavorName)))

	w := &markdownWriter{buf: &buf}

	if len(root.Options) > 0 {
		for _, opt := range root.Options {
			w.renderPatternOption(0, opt)
		}
	}

	w.renderRegexp(0, root, true)

	return buf.String()
}

func (w *markdownWriter) renderRegexp(indent int, r *ast.Regexp, isRoot bool) {
	if r == nil {
		return
	}

	if len(r.Matches) > 1 {
		w.line(indent, fmt.Sprintf("**Alternation** (%d branches)", len(r.Matches)))
		for _, m := range r.Matches {
			w.renderMatch(indent+1, m)
		}
		return
	}

	if len(r.Matches) == 1 {
		w.renderMatch(indent, r.Matches[0])
	}
}

func (w *markdownWriter) renderMatch(indent int, m *ast.Match) {
	if m == nil {
		return
	}
	if len(m.Fragments) > 1 {
		w.line(indent, "**Sequence**")
		for _, f := range m.Fragments {
			w.renderFragment(indent+1, f)
		}
		return
	}
	if len(m.Fragments) == 1 {
		w.renderFragment(indent, m.Fragments[0])
	}
}

func (w *markdownWriter) renderFragment(indent int, f *ast.MatchFragment) {
	w.renderNode(indent, f.Content)
	if f.Repeat != nil {
		w.line(indent, w.formatQuantifier(f.Repeat))
	}
}

func (w *markdownWriter) renderNode(indent int, n ast.Node) {
	if n == nil {
		return
	}
	switch v := n.(type) {
	case *ast.Literal:
		w.line(indent, fmt.Sprintf("Literal `%s`", v.Text))
	case *ast.AnyCharacter:
		w.line(indent, "Any character")
	case *ast.Anchor:
		desc, ok := anchorDescriptions[v.AnchorType]
		if !ok {
			desc = v.AnchorType
		}
		w.line(indent, desc)
	case *ast.Escape:
		w.line(indent, fmt.Sprintf("Escape: %s", v.Value))
	case *ast.Charset:
		w.renderCharset(indent, v)
	case *ast.Subexp:
		w.renderSubexp(indent, v)
	case *ast.BackReference:
		w.renderBackReference(indent, v)
	case *ast.Conditional:
		w.renderConditional(indent, v)
	case *ast.RecursiveRef:
		w.renderRecursiveRef(indent, v)
	case *ast.Comment:
		w.line(indent, fmt.Sprintf("Comment: %q", v.Text))
	case *ast.QuotedLiteral:
		w.line(indent, fmt.Sprintf("Literal `%s`", v.Text))
	case *ast.InlineModifier:
		w.renderInlineModifier(indent, v)
	case *ast.BranchReset:
		w.line(indent, "**Branch reset**")
		w.renderRegexp(indent+1, v.Regexp, false)
	case *ast.BalancedGroup:
		w.line(indent, fmt.Sprintf("**Balanced group** %q (pop %q)", v.Name, v.OtherName))
		w.renderRegexp(indent+1, v.Regexp, false)
	case *ast.BacktrackControl:
		if v.Arg != "" {
			w.line(indent, fmt.Sprintf("Backtrack: %s(%s)", v.Verb, v.Arg))
		} else {
			w.line(indent, fmt.Sprintf("Backtrack: %s", v.Verb))
		}
	case *ast.PatternOption:
		w.renderPatternOption(indent, v)
	case *ast.Callout:
		w.renderCallout(indent, v)
	case *ast.CharsetIntersection:
		w.line(indent, "**Intersection**")
		for _, op := range v.Operands {
			w.renderNode(indent+1, op)
		}
	case *ast.CharsetSubtraction:
		w.line(indent, "**Subtraction**")
		for _, op := range v.Operands {
			w.renderNode(indent+1, op)
		}
	case *ast.CharsetStringDisjunction:
		w.renderStringDisjunction(indent, v)
	case *ast.CharsetLiteral:
		w.line(indent, fmt.Sprintf("`%s`", v.Text))
	case *ast.CharsetRange:
		w.line(indent, fmt.Sprintf("Range `%s` to `%s`", v.First, v.Last))
	case *ast.POSIXClass:
		if v.Negated {
			w.line(indent, fmt.Sprintf("NOT POSIX [:%s:]", v.Name))
		} else {
			w.line(indent, fmt.Sprintf("POSIX [:%s:]", v.Name))
		}
	case *ast.UnicodePropertyEscape:
		if v.Negated {
			w.line(indent, fmt.Sprintf("NOT Unicode property \\P{%s}", v.Property))
		} else {
			w.line(indent, fmt.Sprintf("Unicode property \\p{%s}", v.Property))
		}
	}
}

func (w *markdownWriter) renderCharset(indent int, c *ast.Charset) {
	if c.Inverted {
		w.line(indent, "**Negated character class**: none of")
	} else {
		w.line(indent, "**Character class**: one of")
	}
	if c.SetExpression != nil {
		w.renderNode(indent+1, c.SetExpression)
	} else {
		for _, item := range c.Items {
			w.renderNode(indent+1, item)
		}
	}
}

func (w *markdownWriter) renderSubexp(indent int, s *ast.Subexp) {
	switch s.GroupType {
	case ast.GroupCapture:
		w.line(indent, fmt.Sprintf("**Capture group #%d**", s.Number))
	case ast.GroupNamedCapture:
		w.line(indent, fmt.Sprintf("**Named capture group #%d %q**", s.Number, s.Name))
	case ast.GroupNonCapture:
		w.line(indent, "**Non-capturing group**")
	case ast.GroupPositiveLookahead:
		w.line(indent, "**Positive lookahead**")
	case ast.GroupNegativeLookahead:
		w.line(indent, "**Negative lookahead**")
	case ast.GroupPositiveLookbehind:
		w.line(indent, "**Positive lookbehind**")
	case ast.GroupNegativeLookbehind:
		w.line(indent, "**Negative lookbehind**")
	case ast.GroupAtomic:
		w.line(indent, "**Atomic group**")
	default:
		w.line(indent, fmt.Sprintf("**Group (%s)**", s.GroupType))
	}
	w.renderRegexp(indent+1, s.Regexp, false)
}

func (w *markdownWriter) renderBackReference(indent int, br *ast.BackReference) {
	if br.Name != "" {
		w.line(indent, fmt.Sprintf("Back-reference %q", br.Name))
	} else {
		w.line(indent, fmt.Sprintf("Back-reference #%d", br.Number))
	}
}

func (w *markdownWriter) renderConditional(indent int, c *ast.Conditional) {
	w.line(indent, "**Conditional**")
	w.renderNode(indent+1, c.Condition)
	if c.TrueMatch != nil {
		w.renderRegexp(indent+1, c.TrueMatch, false)
	}
	if c.FalseMatch != nil {
		w.renderRegexp(indent+1, c.FalseMatch, false)
	}
}

func (w *markdownWriter) renderRecursiveRef(indent int, rr *ast.RecursiveRef) {
	if rr.Target == "R" || rr.Target == "0" {
		w.line(indent, "Recurse whole pattern")
	} else {
		w.line(indent, fmt.Sprintf("Recurse group %s", rr.Target))
	}
}

func (w *markdownWriter) renderInlineModifier(indent int, im *ast.InlineModifier) {
	var parts []string
	if im.Enable != "" {
		parts = append(parts, "+"+im.Enable)
	}
	if im.Disable != "" {
		parts = append(parts, "-"+im.Disable)
	}
	w.line(indent, fmt.Sprintf("Flags: %s", strings.Join(parts, " ")))
	if im.Regexp != nil {
		w.renderRegexp(indent+1, im.Regexp, false)
	}
}

func (w *markdownWriter) renderPatternOption(indent int, po *ast.PatternOption) {
	if po.Value != "" {
		w.line(indent, fmt.Sprintf("Option: %s=%s", po.Name, po.Value))
	} else {
		w.line(indent, fmt.Sprintf("Option: %s", po.Name))
	}
}

func (w *markdownWriter) renderCallout(indent int, co *ast.Callout) {
	if co.Number == -1 {
		w.line(indent, fmt.Sprintf("Callout %q", co.Text))
	} else {
		w.line(indent, fmt.Sprintf("Callout #%d", co.Number))
	}
}

func (w *markdownWriter) renderStringDisjunction(indent int, csd *ast.CharsetStringDisjunction) {
	quoted := make([]string, len(csd.Strings))
	for i, s := range csd.Strings {
		quoted[i] = fmt.Sprintf("%q", s)
	}
	w.line(indent, fmt.Sprintf("String disjunction: %s", strings.Join(quoted, ", ")))
}

func (w *markdownWriter) formatQuantifier(r *ast.Repeat) string {
	var desc string
	switch {
	case r.Min == 0 && r.Max == -1:
		desc = "0 or more"
	case r.Min == 1 && r.Max == -1:
		desc = "1 or more"
	case r.Min == 0 && r.Max == 1:
		desc = "optional"
	case r.Min == r.Max:
		return fmt.Sprintf("Quantifier: exactly %d times", r.Min)
	case r.Max == -1:
		desc = fmt.Sprintf("%d or more", r.Min)
	default:
		desc = fmt.Sprintf("%d to %d times", r.Min, r.Max)
	}

	modifier := "greedy"
	if r.Possessive {
		modifier = "possessive"
	} else if !r.Greedy {
		modifier = "lazy"
	}
	return fmt.Sprintf("Quantifier: %s (%s)", desc, modifier)
}
