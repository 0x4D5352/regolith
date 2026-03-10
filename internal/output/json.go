package output

import (
	"encoding/json"
	"fmt"

	"github.com/0x4d5352/regolith/internal/ast"
)

// Document is the top-level JSON envelope wrapping a parsed regex.
type Document struct {
	Pattern string `json:"pattern"`
	Flavor  string `json:"flavor"`
	Root    any    `json:"root"`
}

// Quantifier represents a repetition attached to its content node.
type Quantifier struct {
	Min        int  `json:"min"`
	Max        *int `json:"max"`
	Greedy     bool `json:"greedy"`
	Possessive bool `json:"possessive,omitempty"`
}

// RenderJSON converts a parsed AST into a pretty-printed JSON string.
func RenderJSON(root *ast.Regexp, pattern, flavorName string) (string, error) {
	rootNode := convertRegexp(root, true)
	doc := Document{
		Pattern: pattern,
		Flavor:  flavorName,
		Root:    rootNode,
	}
	b, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", fmt.Errorf("json marshal: %w", err)
	}
	return string(b), nil
}

// convertRegexp converts an ast.Regexp. When isRoot is true, pattern options
// are attached to the resulting node.
func convertRegexp(r *ast.Regexp, isRoot bool) any {
	if r == nil {
		return nil
	}

	// Single branch: unwrap to the sequence directly.
	if len(r.Matches) == 1 {
		node := convertMatch(r.Matches[0])
		if isRoot && len(r.Options) > 0 {
			if m, ok := node.(map[string]any); ok {
				m["options"] = convertPatternOptions(r.Options)
			}
		}
		if isRoot && r.Flags != "" {
			if m, ok := node.(map[string]any); ok {
				m["flags"] = r.Flags
			}
		}
		return node
	}

	// Multiple branches: alternation.
	alts := make([]any, len(r.Matches))
	for i, m := range r.Matches {
		alts[i] = convertMatch(m)
	}
	result := map[string]any{
		"type":         "alternation",
		"alternatives": alts,
	}
	if r.Flags != "" {
		result["flags"] = r.Flags
	}
	if isRoot && len(r.Options) > 0 {
		result["options"] = convertPatternOptions(r.Options)
	}
	return result
}

func convertPatternOptions(opts []*ast.PatternOption) []any {
	out := make([]any, len(opts))
	for i, o := range opts {
		out[i] = convertNode(o)
	}
	return out
}

func convertMatch(m *ast.Match) any {
	if m == nil {
		return nil
	}
	elements := make([]any, len(m.Fragments))
	for i, f := range m.Fragments {
		elements[i] = convertFragment(f)
	}
	return map[string]any{
		"type":     "sequence",
		"elements": elements,
	}
}

func convertFragment(f *ast.MatchFragment) any {
	node := convertNode(f.Content)
	if f.Repeat != nil && node != nil {
		node["quantifier"] = convertQuantifier(f.Repeat)
	}
	return node
}

func convertQuantifier(r *ast.Repeat) Quantifier {
	q := Quantifier{
		Min:        r.Min,
		Greedy:     r.Greedy,
		Possessive: r.Possessive,
	}
	if r.Max == -1 {
		q.Max = nil
	} else {
		max := r.Max
		q.Max = &max
	}
	return q
}

func convertNode(n ast.Node) map[string]any {
	if n == nil {
		return nil
	}
	switch v := n.(type) {
	case *ast.Literal:
		return map[string]any{
			"type":  "literal",
			"value": v.Text,
		}
	case *ast.AnyCharacter:
		return map[string]any{
			"type": "anyCharacter",
		}
	case *ast.Anchor:
		return map[string]any{
			"type":       "anchor",
			"anchorType": v.AnchorType,
		}
	case *ast.Escape:
		return map[string]any{
			"type":       "escape",
			"escapeType": v.EscapeType,
			"code":       v.Code,
			"value":      v.Value,
		}
	case *ast.Charset:
		return convertCharset(v)
	case *ast.CharsetLiteral:
		return map[string]any{
			"type":  "literal",
			"value": v.Text,
		}
	case *ast.CharsetRange:
		return map[string]any{
			"type": "range",
			"from": v.First,
			"to":   v.Last,
		}
	case *ast.POSIXClass:
		return map[string]any{
			"type":    "posixClass",
			"name":    v.Name,
			"negated": v.Negated,
		}
	case *ast.UnicodePropertyEscape:
		return map[string]any{
			"type":     "unicodeProperty",
			"property": v.Property,
			"negated":  v.Negated,
		}
	case *ast.Subexp:
		return convertSubexp(v)
	case *ast.BackReference:
		return convertBackReference(v)
	case *ast.Conditional:
		return convertConditional(v)
	case *ast.RecursiveRef:
		return map[string]any{
			"type":   "recursiveReference",
			"target": v.Target,
		}
	case *ast.Comment:
		return map[string]any{
			"type": "comment",
			"text": v.Text,
		}
	case *ast.QuotedLiteral:
		return map[string]any{
			"type": "quotedLiteral",
			"text": v.Text,
		}
	case *ast.InlineModifier:
		return convertInlineModifier(v)
	case *ast.BranchReset:
		return map[string]any{
			"type": "branchReset",
			"body": convertRegexp(v.Regexp, false),
		}
	case *ast.BalancedGroup:
		return map[string]any{
			"type":      "balancedGroup",
			"name":      v.Name,
			"otherName": v.OtherName,
			"body":      convertRegexp(v.Regexp, false),
		}
	case *ast.BacktrackControl:
		result := map[string]any{
			"type": "backtrackControl",
			"verb": v.Verb,
		}
		if v.Arg != "" {
			result["arg"] = v.Arg
		}
		return result
	case *ast.PatternOption:
		result := map[string]any{
			"type": "patternOption",
			"name": v.Name,
		}
		if v.Value != "" {
			result["value"] = v.Value
		}
		return result
	case *ast.Callout:
		return convertCallout(v)
	case *ast.CharsetIntersection:
		operands := make([]any, len(v.Operands))
		for i, op := range v.Operands {
			operands[i] = convertNode(op)
		}
		return map[string]any{
			"type":     "intersection",
			"operands": operands,
		}
	case *ast.CharsetSubtraction:
		operands := make([]any, len(v.Operands))
		for i, op := range v.Operands {
			operands[i] = convertNode(op)
		}
		return map[string]any{
			"type":     "subtraction",
			"operands": operands,
		}
	case *ast.CharsetStringDisjunction:
		return map[string]any{
			"type":    "stringDisjunction",
			"strings": v.Strings,
		}
	default:
		return map[string]any{
			"type": "unknown",
		}
	}
}

func convertCharset(c *ast.Charset) map[string]any {
	result := map[string]any{
		"type":    "characterClass",
		"negated": c.Inverted,
	}
	if c.SetExpression != nil {
		result["members"] = []any{convertNode(c.SetExpression)}
	} else {
		members := make([]any, len(c.Items))
		for i, item := range c.Items {
			members[i] = convertNode(item)
		}
		result["members"] = members
	}
	return result
}

var groupTypeToKind = map[string]string{
	ast.GroupCapture:            "capture",
	ast.GroupNonCapture:         "nonCapture",
	ast.GroupPositiveLookahead:  "positiveLookahead",
	ast.GroupNegativeLookahead:  "negativeLookahead",
	ast.GroupPositiveLookbehind: "positiveLookbehind",
	ast.GroupNegativeLookbehind: "negativeLookbehind",
	ast.GroupNamedCapture:       "namedCapture",
	ast.GroupAtomic:             "atomic",
}

func convertSubexp(s *ast.Subexp) map[string]any {
	kind, ok := groupTypeToKind[s.GroupType]
	if !ok {
		kind = s.GroupType
	}
	result := map[string]any{
		"type": "group",
		"kind": kind,
		"body": convertRegexp(s.Regexp, false),
	}
	if s.GroupType == ast.GroupCapture || s.GroupType == ast.GroupNamedCapture {
		result["number"] = s.Number
	}
	if s.Name != "" {
		result["name"] = s.Name
	}
	return result
}

func convertBackReference(br *ast.BackReference) map[string]any {
	result := map[string]any{
		"type": "backReference",
	}
	if br.Number != 0 {
		result["number"] = br.Number
	}
	if br.Name != "" {
		result["name"] = br.Name
	}
	return result
}

func convertConditional(c *ast.Conditional) map[string]any {
	result := map[string]any{
		"type":      "conditional",
		"condition": convertNode(c.Condition),
		"ifTrue":    convertRegexp(c.TrueMatch, false),
	}
	if c.FalseMatch != nil {
		result["ifFalse"] = convertRegexp(c.FalseMatch, false)
	}
	return result
}

func convertInlineModifier(im *ast.InlineModifier) map[string]any {
	result := map[string]any{
		"type": "inlineModifier",
	}
	if im.Enable != "" {
		result["enable"] = im.Enable
	}
	if im.Disable != "" {
		result["disable"] = im.Disable
	}
	if im.Regexp != nil {
		result["body"] = convertRegexp(im.Regexp, false)
	}
	return result
}

func convertCallout(co *ast.Callout) map[string]any {
	result := map[string]any{
		"type": "callout",
	}
	if co.Number >= 0 {
		n := co.Number
		result["number"] = &n
	}
	if co.Text != "" {
		result["text"] = co.Text
	}
	return result
}
