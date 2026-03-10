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
	ast.AnchorStart:                   "Asserts start of line",
	ast.AnchorEnd:                     "Asserts end of line",
	ast.AnchorWordBoundary:            "Asserts word boundary",
	ast.AnchorNonWordBoundary:         "Asserts non-word boundary",
	ast.AnchorStringStart:             "Asserts start of string",
	ast.AnchorStringEnd:               "Asserts end of string",
	ast.AnchorAbsoluteEnd:             "Asserts absolute end of string",
	ast.AnchorWordStart:               "Asserts start of word",
	ast.AnchorWordEnd:                 "Asserts end of word",
	ast.AnchorGraphemeClusterBoundary: "Asserts grapheme cluster boundary",
}

// escapeInfo maps escape type to [shortName, detail].
// shortName is the concise label (e.g., "any digit").
// detail is the parenthetical expansion (e.g., "0-9").
type escapeInfo struct {
	shortName string
	detail    string
}

var escapeDescriptions = map[string]escapeInfo{
	"digit":                {"any digit", "0-9"},
	"non_digit":            {"any non-digit character", ""},
	"word":                 {"any word character", "a-z, A-Z, 0-9, _"},
	"non_word":             {"any non-word character", ""},
	"whitespace":           {"any whitespace character", "space, tab, newline, etc."},
	"non_whitespace":       {"any non-whitespace character", ""},
	"newline":              {"a newline character", ""},
	"tab":                  {"a tab character", ""},
	"carriage_return":      {"a carriage return", ""},
	"form_feed":            {"a form feed character", ""},
	"null":                 {"a null character", ""},
	"vertical_tab":         {"a vertical tab character", ""},
	"horizontal_space":     {"any horizontal whitespace", ""},
	"non_horizontal_space": {"any non-horizontal whitespace", ""},
	"vertical_space":       {"any vertical whitespace", ""},
	"non_vertical_space":   {"any non-vertical whitespace", ""},
}

var escapeShortCodes = map[string]string{
	"digit":          `\d`,
	"non_digit":      `\D`,
	"word":           `\w`,
	"non_word":       `\W`,
	"whitespace":     `\s`,
	"non_whitespace": `\S`,
}

var groupAnnotations = map[string]string{
	ast.GroupPositiveLookahead:  "asserts what follows matches, without consuming characters",
	ast.GroupNegativeLookahead:  "asserts what follows does NOT match",
	ast.GroupPositiveLookbehind: "asserts what precedes matches, without consuming characters",
	ast.GroupNegativeLookbehind: "asserts what precedes does NOT match",
	ast.GroupAtomic:             "matches without backtracking",
	ast.GroupNonCapture:         "groups without capturing",
}

var rangeDescriptions = map[string]string{
	"a-z": "lowercase letters",
	"A-Z": "uppercase letters",
	"0-9": "digits",
}

var posixClassDescriptions = map[string]string{
	"alnum":  "alphanumeric characters",
	"alpha":  "alphabetic characters",
	"blank":  "space or tab",
	"cntrl":  "control characters",
	"digit":  "digits",
	"graph":  "visible characters",
	"lower":  "lowercase letters",
	"print":  "printable characters",
	"punct":  "punctuation",
	"space":  "whitespace",
	"upper":  "uppercase letters",
	"xdigit": "hex digits",
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
	fmt.Fprintf(&buf, "# Regex: `%s`\n\n**Flavor:** %s\n\n", pattern, formatFlavorName(flavorName))

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
		w.line(indent, fmt.Sprintf("**Alternation** -- matches one of %d branches:", len(r.Matches)))
		for i, m := range r.Matches {
			w.renderBranch(indent+1, m, i+1)
		}
		return
	}

	if len(r.Matches) == 1 {
		w.renderMatch(indent, r.Matches[0])
	}
}

func (w *markdownWriter) renderBranch(indent int, m *ast.Match, branchNum int) {
	if m == nil {
		return
	}
	// Single-fragment branches get inlined: "**Branch N:** description"
	if len(m.Fragments) == 1 {
		desc := w.describeFragment(m.Fragments[0])
		w.line(indent, fmt.Sprintf("**Branch %d:** %s", branchNum, desc))
		// If the fragment has children (subexp, charset, etc.), render them
		w.renderFragmentChildren(indent+1, m.Fragments[0])
		return
	}
	// Multi-fragment branches get a sequence under the branch header
	w.line(indent, fmt.Sprintf("**Branch %d:**", branchNum))
	for _, f := range m.Fragments {
		w.renderFragment(indent+1, f)
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
	switch v := f.Content.(type) {
	case *ast.Literal, *ast.AnyCharacter, *ast.Escape, *ast.Anchor,
		*ast.BackReference, *ast.QuotedLiteral, *ast.Comment,
		*ast.RecursiveRef, *ast.BacktrackControl, *ast.PatternOption,
		*ast.Callout, *ast.InlineModifier, *ast.UnicodePropertyEscape:
		// Simple nodes: merge quantifier into single line
		desc := w.describeNode(v)
		desc += w.formatQuantifierSuffix(f.Repeat)
		w.line(indent, desc)
		// InlineModifier may have children
		if im, ok := v.(*ast.InlineModifier); ok && im.Regexp != nil {
			w.renderRegexp(indent+1, im.Regexp, false)
		}
	case *ast.Charset:
		w.renderCharset(indent, v, f.Repeat)
	case *ast.Subexp:
		w.renderSubexp(indent, v, f.Repeat)
	case *ast.Conditional:
		w.renderConditional(indent, v)
	case *ast.BranchReset:
		w.line(indent, "**Branch reset** -- resets group numbering for each branch")
		w.renderRegexp(indent+1, v.Regexp, false)
	case *ast.BalancedGroup:
		w.line(indent, fmt.Sprintf("**Balanced group** %q (pop %q)", v.Name, v.OtherName))
		w.renderRegexp(indent+1, v.Regexp, false)
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
		w.line(indent, w.describeRange(v))
	case *ast.POSIXClass:
		w.line(indent, w.describePOSIXClass(v))
	default:
		if v != nil {
			w.line(indent, fmt.Sprintf("Unknown node: %T", v))
		}
	}
}

// describeFragment returns a single-line description for a fragment (used in branch inlining).
func (w *markdownWriter) describeFragment(f *ast.MatchFragment) string {
	switch f.Content.(type) {
	case *ast.Charset, *ast.Subexp, *ast.Conditional, *ast.BranchReset,
		*ast.BalancedGroup, *ast.CharsetIntersection, *ast.CharsetSubtraction:
		// Complex nodes — the header is returned, children rendered separately
		return w.describeComplexHeader(f)
	default:
		desc := w.describeNode(f.Content)
		desc += w.formatQuantifierSuffix(f.Repeat)
		return desc
	}
}

// describeComplexHeader returns the header line for complex nodes in branch context.
func (w *markdownWriter) describeComplexHeader(f *ast.MatchFragment) string {
	switch v := f.Content.(type) {
	case *ast.Charset:
		return w.charsetHeader(v, f.Repeat)
	case *ast.Subexp:
		return w.subexpHeader(v, f.Repeat)
	case *ast.Conditional:
		return "**Conditional** -- matches based on a condition"
	case *ast.BranchReset:
		return "**Branch reset** -- resets group numbering for each branch"
	case *ast.BalancedGroup:
		return fmt.Sprintf("**Balanced group** %q (pop %q)", v.Name, v.OtherName)
	default:
		return w.describeNode(f.Content)
	}
}

// renderFragmentChildren renders child nodes for complex fragments (used after branch inlining).
func (w *markdownWriter) renderFragmentChildren(indent int, f *ast.MatchFragment) {
	switch v := f.Content.(type) {
	case *ast.Charset:
		w.renderCharsetItems(indent, v)
	case *ast.Subexp:
		w.renderRegexp(indent, v.Regexp, false)
	case *ast.Conditional:
		w.renderNode(indent, v.Condition)
		if v.TrueMatch != nil {
			w.renderRegexp(indent, v.TrueMatch, false)
		}
		if v.FalseMatch != nil {
			w.renderRegexp(indent, v.FalseMatch, false)
		}
	case *ast.BranchReset:
		w.renderRegexp(indent, v.Regexp, false)
	case *ast.BalancedGroup:
		w.renderRegexp(indent, v.Regexp, false)
	case *ast.InlineModifier:
		if v.Regexp != nil {
			w.renderRegexp(indent, v.Regexp, false)
		}
	}
}

func (w *markdownWriter) describeNode(n ast.Node) string {
	if n == nil {
		return ""
	}
	switch v := n.(type) {
	case *ast.Literal:
		return fmt.Sprintf("Matches `%s` literally", v.Text)
	case *ast.AnyCharacter:
		return "Matches any character"
	case *ast.Anchor:
		desc, ok := anchorDescriptions[v.AnchorType]
		if !ok {
			desc = "Asserts " + v.AnchorType
		}
		return desc
	case *ast.Escape:
		return describeEscape(v)
	case *ast.BackReference:
		return describeBackReference(v)
	case *ast.QuotedLiteral:
		return fmt.Sprintf("Matches `%s` literally", v.Text)
	case *ast.Comment:
		return fmt.Sprintf("Comment: %q", v.Text)
	case *ast.RecursiveRef:
		if v.Target == "R" || v.Target == "0" {
			return "Recurse whole pattern"
		}
		return fmt.Sprintf("Recurse group %s", v.Target)
	case *ast.BacktrackControl:
		if v.Arg != "" {
			return fmt.Sprintf("Backtrack: %s(%s)", v.Verb, v.Arg)
		}
		return fmt.Sprintf("Backtrack: %s", v.Verb)
	case *ast.PatternOption:
		if v.Value != "" {
			return fmt.Sprintf("Option: %s=%s", v.Name, v.Value)
		}
		return fmt.Sprintf("Option: %s", v.Name)
	case *ast.Callout:
		if v.Number == -1 {
			return fmt.Sprintf("Callout %q", v.Text)
		}
		return fmt.Sprintf("Callout #%d", v.Number)
	case *ast.InlineModifier:
		return describeInlineModifier(v)
	case *ast.UnicodePropertyEscape:
		if v.Negated {
			return fmt.Sprintf("NOT Unicode property `\\P{%s}`", v.Property)
		}
		return fmt.Sprintf("Unicode property `\\p{%s}`", v.Property)
	}
	return ""
}

func describeEscape(e *ast.Escape) string {
	if info, ok := escapeDescriptions[e.EscapeType]; ok {
		if code, ok2 := escapeShortCodes[e.EscapeType]; ok2 {
			if info.detail != "" {
				return fmt.Sprintf("Matches %s `%s` (%s)", info.shortName, code, info.detail)
			}
			return fmt.Sprintf("Matches %s `%s`", info.shortName, code)
		}
		if info.detail != "" {
			return fmt.Sprintf("Matches %s (%s)", info.shortName, info.detail)
		}
		return fmt.Sprintf("Matches %s", info.shortName)
	}
	return fmt.Sprintf("Matches escape `\\%s`", e.Code)
}

func describeBackReference(br *ast.BackReference) string {
	if br.Name != "" {
		return fmt.Sprintf("Matches the same text previously captured by group %q", br.Name)
	}
	return fmt.Sprintf("Matches the same text previously captured by group #%d", br.Number)
}

func describeInlineModifier(im *ast.InlineModifier) string {
	var parts []string
	if im.Enable != "" {
		parts = append(parts, "+"+im.Enable)
	}
	if im.Disable != "" {
		parts = append(parts, "-"+im.Disable)
	}
	return fmt.Sprintf("Flags: %s -- modifies matching behavior", strings.Join(parts, " "))
}

func (w *markdownWriter) formatQuantifierSuffix(r *ast.Repeat) string {
	if r == nil {
		return ""
	}
	switch {
	case r.Min == 0 && r.Max == -1:
		return fmt.Sprintf(", 0 or more times (%s)", quantifierModifier(r))
	case r.Min == 1 && r.Max == -1:
		return fmt.Sprintf(", 1 or more times (%s)", quantifierModifier(r))
	case r.Min == 0 && r.Max == 1:
		if r.Greedy && !r.Possessive {
			return ", optionally"
		}
		return fmt.Sprintf(", optionally (%s)", quantifierModifier(r))
	case r.Min == r.Max:
		return fmt.Sprintf(", exactly %d times", r.Min)
	case r.Max == -1:
		return fmt.Sprintf(", %d or more times (%s)", r.Min, quantifierModifier(r))
	default:
		return fmt.Sprintf(", %d to %d times (%s)", r.Min, r.Max, quantifierModifier(r))
	}
}

func quantifierModifier(r *ast.Repeat) string {
	if r.Possessive {
		return "possessive"
	}
	if !r.Greedy {
		return "lazy"
	}
	return "greedy"
}

func (w *markdownWriter) renderCharset(indent int, c *ast.Charset, repeat *ast.Repeat) {
	header := w.charsetHeader(c, repeat)
	w.line(indent, header)
	w.renderCharsetItems(indent+1, c)
}

func (w *markdownWriter) charsetHeader(c *ast.Charset, repeat *ast.Repeat) string {
	var header string
	if c.Inverted {
		header = "Matches any character NOT in:"
	} else {
		header = "Matches one of the following:"
	}
	if repeat != nil {
		// Replace trailing colon with quantifier suffix + colon
		header = strings.TrimSuffix(header, ":")
		header += w.formatQuantifierSuffix(repeat) + ":"
	}
	return header
}

func (w *markdownWriter) renderCharsetItems(indent int, c *ast.Charset) {
	if c.SetExpression != nil {
		w.renderNode(indent, c.SetExpression)
	} else {
		// Group consecutive CharsetLiterals together
		i := 0
		for i < len(c.Items) {
			if lit, ok := c.Items[i].(*ast.CharsetLiteral); ok {
				// Collect consecutive literals
				lits := []string{fmt.Sprintf("`%s`", lit.Text)}
				j := i + 1
				for j < len(c.Items) {
					if nextLit, ok := c.Items[j].(*ast.CharsetLiteral); ok {
						lits = append(lits, fmt.Sprintf("`%s`", nextLit.Text))
						j++
					} else {
						break
					}
				}
				if len(lits) > 1 {
					w.line(indent, strings.Join(lits, ", ")+" (literal characters)")
				} else {
					w.line(indent, lits[0])
				}
				i = j
			} else {
				w.renderNode(indent, c.Items[i])
				i++
			}
		}
	}
}

// renderNode renders a node within a charset or other container context.
func (w *markdownWriter) renderNode(indent int, n ast.Node) {
	if n == nil {
		return
	}
	switch v := n.(type) {
	case *ast.Charset:
		w.renderCharset(indent, v, nil)
	case *ast.CharsetLiteral:
		w.line(indent, fmt.Sprintf("`%s`", v.Text))
	case *ast.CharsetRange:
		w.line(indent, w.describeRange(v))
	case *ast.Escape:
		w.line(indent, describeEscapeInCharset(v))
	case *ast.POSIXClass:
		w.line(indent, w.describePOSIXClass(v))
	case *ast.UnicodePropertyEscape:
		if v.Negated {
			w.line(indent, fmt.Sprintf("NOT Unicode property `\\P{%s}`", v.Property))
		} else {
			w.line(indent, fmt.Sprintf("Unicode property `\\p{%s}`", v.Property))
		}
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
	default:
		// Fallback for any other node types in charset context
		w.line(indent, w.describeNode(n))
	}
}

func describeEscapeInCharset(e *ast.Escape) string {
	if info, ok := escapeDescriptions[e.EscapeType]; ok {
		if code, ok2 := escapeShortCodes[e.EscapeType]; ok2 {
			if info.detail != "" {
				return fmt.Sprintf("%s `%s` (%s)", info.shortName, code, info.detail)
			}
			return fmt.Sprintf("%s `%s`", info.shortName, code)
		}
		if info.detail != "" {
			return fmt.Sprintf("%s (%s)", info.shortName, info.detail)
		}
		return info.shortName
	}
	return fmt.Sprintf("escape `\\%s`", e.Code)
}

func (w *markdownWriter) describeRange(r *ast.CharsetRange) string {
	key := r.First + "-" + r.Last
	if desc, ok := rangeDescriptions[key]; ok {
		return fmt.Sprintf("`%s` to `%s` (%s)", r.First, r.Last, desc)
	}
	return fmt.Sprintf("`%s` to `%s`", r.First, r.Last)
}

func (w *markdownWriter) describePOSIXClass(pc *ast.POSIXClass) string {
	desc := ""
	if d, ok := posixClassDescriptions[pc.Name]; ok {
		desc = " (" + d + ")"
	}
	if pc.Negated {
		return fmt.Sprintf("NOT POSIX `[:%s:]`%s", pc.Name, desc)
	}
	return fmt.Sprintf("POSIX `[:%s:]`%s", pc.Name, desc)
}

func (w *markdownWriter) renderSubexp(indent int, s *ast.Subexp, repeat *ast.Repeat) {
	header := w.subexpHeader(s, repeat)
	w.line(indent, header)
	w.renderRegexp(indent+1, s.Regexp, false)
}

func (w *markdownWriter) subexpHeader(s *ast.Subexp, repeat *ast.Repeat) string {
	var header string
	suffix := w.formatQuantifierSuffix(repeat)

	switch s.GroupType {
	case ast.GroupCapture:
		header = fmt.Sprintf("**Capture group #%d** -- captures matched text for back-reference as `\\%d`%s", s.Number, s.Number, suffix)
	case ast.GroupNamedCapture:
		header = fmt.Sprintf("**Named capture group #%d %q** -- captures matched text for back-reference as `\\%d` or by name%s", s.Number, s.Name, s.Number, suffix)
	default:
		annotation := ""
		if ann, ok := groupAnnotations[s.GroupType]; ok {
			annotation = " -- " + ann
		}
		label := groupLabel(s.GroupType)
		header = fmt.Sprintf("**%s**%s%s", label, annotation, suffix)
	}
	return header
}

func groupLabel(groupType string) string {
	switch groupType {
	case ast.GroupNonCapture:
		return "Non-capturing group"
	case ast.GroupPositiveLookahead:
		return "Positive lookahead"
	case ast.GroupNegativeLookahead:
		return "Negative lookahead"
	case ast.GroupPositiveLookbehind:
		return "Positive lookbehind"
	case ast.GroupNegativeLookbehind:
		return "Negative lookbehind"
	case ast.GroupAtomic:
		return "Atomic group"
	default:
		return fmt.Sprintf("Group (%s)", groupType)
	}
}

func (w *markdownWriter) renderConditional(indent int, c *ast.Conditional) {
	w.line(indent, "**Conditional** -- matches based on a condition")
	w.renderNode(indent+1, c.Condition)
	if c.TrueMatch != nil {
		w.renderRegexp(indent+1, c.TrueMatch, false)
	}
	if c.FalseMatch != nil {
		w.renderRegexp(indent+1, c.FalseMatch, false)
	}
}

func (w *markdownWriter) renderPatternOption(indent int, po *ast.PatternOption) {
	if po.Value != "" {
		w.line(indent, fmt.Sprintf("Option: %s=%s", po.Name, po.Value))
	} else {
		w.line(indent, fmt.Sprintf("Option: %s", po.Name))
	}
}

func (w *markdownWriter) renderStringDisjunction(indent int, csd *ast.CharsetStringDisjunction) {
	quoted := make([]string, len(csd.Strings))
	for i, s := range csd.Strings {
		quoted[i] = fmt.Sprintf("%q", s)
	}
	w.line(indent, fmt.Sprintf("String disjunction: %s", strings.Join(quoted, ", ")))
}
