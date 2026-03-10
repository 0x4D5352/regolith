package output

import (
	"encoding/json"
	"testing"

	"github.com/0x4d5352/regolith/internal/ast"
)

// unmarshal is a test helper that converts a JSON string to map[string]any.
func unmarshal(t *testing.T, s string) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}
	return m
}

// mustRender calls RenderJSON and fails the test on error.
func mustRender(t *testing.T, root *ast.Regexp, pattern, flavor string) string {
	t.Helper()
	out, err := RenderJSON(root, pattern, flavor)
	if err != nil {
		t.Fatalf("RenderJSON failed: %v", err)
	}
	if !json.Valid([]byte(out)) {
		t.Fatalf("RenderJSON produced invalid JSON:\n%s", out)
	}
	return out
}

func TestDocumentEnvelope(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Literal{Text: "a"}},
			}},
		},
	}
	out := mustRender(t, root, "a", "javascript")
	doc := unmarshal(t, out)

	if doc["pattern"] != "a" {
		t.Errorf("pattern = %v, want %q", doc["pattern"], "a")
	}
	if doc["flavor"] != "javascript" {
		t.Errorf("flavor = %v, want %q", doc["flavor"], "javascript")
	}
	if doc["root"] == nil {
		t.Fatal("root is nil")
	}
}

func TestLiteral(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Literal{Text: "hello"}},
			}},
		},
	}
	out := mustRender(t, root, "hello", "javascript")
	doc := unmarshal(t, out)
	seq := doc["root"].(map[string]any)
	if seq["type"] != "sequence" {
		t.Errorf("root type = %v, want sequence", seq["type"])
	}
	elements := seq["elements"].([]any)
	lit := elements[0].(map[string]any)
	if lit["type"] != "literal" {
		t.Errorf("element type = %v, want literal", lit["type"])
	}
	if lit["value"] != "hello" {
		t.Errorf("value = %v, want hello", lit["value"])
	}
}

func TestEmptyLiteral(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Literal{Text: ""}},
			}},
		},
	}
	out := mustRender(t, root, "", "javascript")
	doc := unmarshal(t, out)
	seq := doc["root"].(map[string]any)
	elements := seq["elements"].([]any)
	lit := elements[0].(map[string]any)
	if lit["value"] != "" {
		t.Errorf("value = %v, want empty string", lit["value"])
	}
}

func TestAnyCharacter(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.AnyCharacter{}},
			}},
		},
	}
	out := mustRender(t, root, ".", "javascript")
	doc := unmarshal(t, out)
	seq := doc["root"].(map[string]any)
	elements := seq["elements"].([]any)
	elem := elements[0].(map[string]any)
	if elem["type"] != "anyCharacter" {
		t.Errorf("type = %v, want anyCharacter", elem["type"])
	}
}

func TestAnchor(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Anchor{AnchorType: ast.AnchorWordBoundary}},
			}},
		},
	}
	out := mustRender(t, root, `\b`, "javascript")
	doc := unmarshal(t, out)
	seq := doc["root"].(map[string]any)
	elements := seq["elements"].([]any)
	elem := elements[0].(map[string]any)
	if elem["type"] != "anchor" {
		t.Errorf("type = %v, want anchor", elem["type"])
	}
	if elem["anchorType"] != "word_boundary" {
		t.Errorf("anchorType = %v, want word_boundary", elem["anchorType"])
	}
}

func TestEscape(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Escape{EscapeType: "digit", Code: "d", Value: "\\d"}},
			}},
		},
	}
	out := mustRender(t, root, `\d`, "javascript")
	doc := unmarshal(t, out)
	seq := doc["root"].(map[string]any)
	elements := seq["elements"].([]any)
	elem := elements[0].(map[string]any)
	if elem["type"] != "escape" {
		t.Errorf("type = %v, want escape", elem["type"])
	}
	if elem["escapeType"] != "digit" {
		t.Errorf("escapeType = %v, want digit", elem["escapeType"])
	}
	if elem["code"] != "d" {
		t.Errorf("code = %v, want d", elem["code"])
	}
	if elem["value"] != `\d` {
		t.Errorf("value = %v, want \\d", elem["value"])
	}
}

func TestCharsetBasic(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Charset{
					Inverted: true,
					Items: []ast.CharsetItem{
						&ast.CharsetRange{First: "a", Last: "z"},
						&ast.CharsetLiteral{Text: "0"},
					},
				}},
			}},
		},
	}
	out := mustRender(t, root, "[^a-z0]", "javascript")
	doc := unmarshal(t, out)
	seq := doc["root"].(map[string]any)
	elements := seq["elements"].([]any)
	cc := elements[0].(map[string]any)
	if cc["type"] != "characterClass" {
		t.Errorf("type = %v, want characterClass", cc["type"])
	}
	if cc["negated"] != true {
		t.Errorf("negated = %v, want true", cc["negated"])
	}
	members := cc["members"].([]any)
	if len(members) != 2 {
		t.Fatalf("members len = %d, want 2", len(members))
	}
	rng := members[0].(map[string]any)
	if rng["type"] != "range" || rng["from"] != "a" || rng["to"] != "z" {
		t.Errorf("range = %v, want range a-z", rng)
	}
	lit := members[1].(map[string]any)
	if lit["type"] != "literal" || lit["value"] != "0" {
		t.Errorf("charset literal = %v, want literal 0", lit)
	}
}

func TestCharsetWithSetExpression(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Charset{
					SetExpression: &ast.CharsetIntersection{
						Operands: []ast.Node{
							&ast.Charset{Items: []ast.CharsetItem{
								&ast.CharsetRange{First: "a", Last: "z"},
							}},
							&ast.UnicodePropertyEscape{Property: "Letter", Negated: false},
						},
					},
				}},
			}},
		},
	}
	out := mustRender(t, root, `[[a-z]&&\p{Letter}]`, "javascript")
	doc := unmarshal(t, out)
	seq := doc["root"].(map[string]any)
	elements := seq["elements"].([]any)
	cc := elements[0].(map[string]any)
	if cc["type"] != "characterClass" {
		t.Fatalf("type = %v, want characterClass", cc["type"])
	}
	members := cc["members"].([]any)
	if len(members) != 1 {
		t.Fatalf("members len = %d, want 1", len(members))
	}
	intersection := members[0].(map[string]any)
	if intersection["type"] != "intersection" {
		t.Errorf("type = %v, want intersection", intersection["type"])
	}
	operands := intersection["operands"].([]any)
	if len(operands) != 2 {
		t.Fatalf("operands len = %d, want 2", len(operands))
	}
}

func TestCharsetSubtraction(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Charset{
					SetExpression: &ast.CharsetSubtraction{
						Operands: []ast.Node{
							&ast.Charset{Items: []ast.CharsetItem{
								&ast.CharsetRange{First: "a", Last: "z"},
							}},
							&ast.Charset{Items: []ast.CharsetItem{
								&ast.CharsetLiteral{Text: "m"},
							}},
						},
					},
				}},
			}},
		},
	}
	out := mustRender(t, root, `[[a-z]--[m]]`, "javascript")
	doc := unmarshal(t, out)
	seq := doc["root"].(map[string]any)
	elements := seq["elements"].([]any)
	cc := elements[0].(map[string]any)
	members := cc["members"].([]any)
	sub := members[0].(map[string]any)
	if sub["type"] != "subtraction" {
		t.Errorf("type = %v, want subtraction", sub["type"])
	}
}

func TestCharsetStringDisjunction(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Charset{
					Items: []ast.CharsetItem{
						&ast.CharsetStringDisjunction{Strings: []string{"abc", "def"}},
					},
				}},
			}},
		},
	}
	out := mustRender(t, root, `[\q{abc|def}]`, "javascript")
	doc := unmarshal(t, out)
	seq := doc["root"].(map[string]any)
	elements := seq["elements"].([]any)
	cc := elements[0].(map[string]any)
	members := cc["members"].([]any)
	sd := members[0].(map[string]any)
	if sd["type"] != "stringDisjunction" {
		t.Errorf("type = %v, want stringDisjunction", sd["type"])
	}
	strs := sd["strings"].([]any)
	if len(strs) != 2 || strs[0] != "abc" || strs[1] != "def" {
		t.Errorf("strings = %v, want [abc def]", strs)
	}
}

func TestPOSIXClass(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Charset{
					Items: []ast.CharsetItem{
						&ast.POSIXClass{Name: "alpha", Negated: false},
					},
				}},
			}},
		},
	}
	out := mustRender(t, root, "[[:alpha:]]", "posix-ere")
	doc := unmarshal(t, out)
	seq := doc["root"].(map[string]any)
	elements := seq["elements"].([]any)
	cc := elements[0].(map[string]any)
	members := cc["members"].([]any)
	pc := members[0].(map[string]any)
	if pc["type"] != "posixClass" {
		t.Errorf("type = %v, want posixClass", pc["type"])
	}
	if pc["name"] != "alpha" {
		t.Errorf("name = %v, want alpha", pc["name"])
	}
	if pc["negated"] != false {
		t.Errorf("negated = %v, want false", pc["negated"])
	}
}

func TestUnicodePropertyEscape(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.UnicodePropertyEscape{Property: "Letter", Negated: true}},
			}},
		},
	}
	out := mustRender(t, root, `\P{Letter}`, "java")
	doc := unmarshal(t, out)
	seq := doc["root"].(map[string]any)
	elements := seq["elements"].([]any)
	elem := elements[0].(map[string]any)
	if elem["type"] != "unicodeProperty" {
		t.Errorf("type = %v, want unicodeProperty", elem["type"])
	}
	if elem["property"] != "Letter" {
		t.Errorf("property = %v, want Letter", elem["property"])
	}
	if elem["negated"] != true {
		t.Errorf("negated = %v, want true", elem["negated"])
	}
}

func TestAlternationMultipleBranches(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Literal{Text: "foo"}},
			}},
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Literal{Text: "bar"}},
			}},
		},
	}
	out := mustRender(t, root, "foo|bar", "javascript")
	doc := unmarshal(t, out)
	alt := doc["root"].(map[string]any)
	if alt["type"] != "alternation" {
		t.Errorf("type = %v, want alternation", alt["type"])
	}
	alts := alt["alternatives"].([]any)
	if len(alts) != 2 {
		t.Fatalf("alternatives len = %d, want 2", len(alts))
	}
}

func TestSingleBranchUnwrapsToSequence(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Literal{Text: "abc"}},
			}},
		},
	}
	out := mustRender(t, root, "abc", "javascript")
	doc := unmarshal(t, out)
	rootNode := doc["root"].(map[string]any)
	if rootNode["type"] != "sequence" {
		t.Errorf("type = %v, want sequence (single branch should unwrap)", rootNode["type"])
	}
}

func TestQuantifierFlattening(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{
					Content: &ast.Literal{Text: "a"},
					Repeat:  &ast.Repeat{Min: 1, Max: -1, Greedy: true},
				},
			}},
		},
	}
	out := mustRender(t, root, "a+", "javascript")
	doc := unmarshal(t, out)
	seq := doc["root"].(map[string]any)
	elements := seq["elements"].([]any)
	elem := elements[0].(map[string]any)
	if elem["type"] != "literal" {
		t.Fatalf("type = %v, want literal", elem["type"])
	}
	q, ok := elem["quantifier"].(map[string]any)
	if !ok {
		t.Fatal("quantifier not found on content node")
	}
	if q["min"] != float64(1) {
		t.Errorf("min = %v, want 1", q["min"])
	}
	if q["max"] != nil {
		t.Errorf("max = %v, want nil (unbounded)", q["max"])
	}
	if q["greedy"] != true {
		t.Errorf("greedy = %v, want true", q["greedy"])
	}
}

func TestQuantifierBoundedMax(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{
					Content: &ast.Literal{Text: "x"},
					Repeat:  &ast.Repeat{Min: 2, Max: 5, Greedy: false},
				},
			}},
		},
	}
	out := mustRender(t, root, "x{2,5}?", "javascript")
	doc := unmarshal(t, out)
	seq := doc["root"].(map[string]any)
	elements := seq["elements"].([]any)
	elem := elements[0].(map[string]any)
	q := elem["quantifier"].(map[string]any)
	if q["min"] != float64(2) {
		t.Errorf("min = %v, want 2", q["min"])
	}
	if q["max"] != float64(5) {
		t.Errorf("max = %v, want 5", q["max"])
	}
	if q["greedy"] != false {
		t.Errorf("greedy = %v, want false", q["greedy"])
	}
}

func TestQuantifierPossessive(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{
					Content: &ast.Literal{Text: "a"},
					Repeat:  &ast.Repeat{Min: 0, Max: -1, Greedy: true, Possessive: true},
				},
			}},
		},
	}
	out := mustRender(t, root, "a*+", "pcre")
	doc := unmarshal(t, out)
	seq := doc["root"].(map[string]any)
	elements := seq["elements"].([]any)
	elem := elements[0].(map[string]any)
	q := elem["quantifier"].(map[string]any)
	if q["possessive"] != true {
		t.Errorf("possessive = %v, want true", q["possessive"])
	}
}

func TestQuantifierPossessiveOmitEmpty(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{
					Content: &ast.Literal{Text: "a"},
					Repeat:  &ast.Repeat{Min: 1, Max: -1, Greedy: true, Possessive: false},
				},
			}},
		},
	}
	out := mustRender(t, root, "a+", "javascript")
	doc := unmarshal(t, out)
	seq := doc["root"].(map[string]any)
	elements := seq["elements"].([]any)
	elem := elements[0].(map[string]any)
	q := elem["quantifier"].(map[string]any)
	if _, exists := q["possessive"]; exists {
		t.Errorf("possessive should be omitted when false, got %v", q["possessive"])
	}
}

func TestNoQuantifier(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Literal{Text: "a"}},
			}},
		},
	}
	out := mustRender(t, root, "a", "javascript")
	doc := unmarshal(t, out)
	seq := doc["root"].(map[string]any)
	elements := seq["elements"].([]any)
	elem := elements[0].(map[string]any)
	if _, exists := elem["quantifier"]; exists {
		t.Error("quantifier should not be present when repeat is nil")
	}
}

func TestSubexpCapture(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Subexp{
					GroupType: ast.GroupCapture,
					Number:    1,
					Regexp: &ast.Regexp{
						Matches: []*ast.Match{
							{Fragments: []*ast.MatchFragment{
								{Content: &ast.Literal{Text: "x"}},
							}},
						},
					},
				}},
			}},
		},
	}
	out := mustRender(t, root, "(x)", "javascript")
	doc := unmarshal(t, out)
	seq := doc["root"].(map[string]any)
	elements := seq["elements"].([]any)
	group := elements[0].(map[string]any)
	if group["type"] != "group" {
		t.Errorf("type = %v, want group", group["type"])
	}
	if group["kind"] != "capture" {
		t.Errorf("kind = %v, want capture", group["kind"])
	}
	if group["number"] != float64(1) {
		t.Errorf("number = %v, want 1", group["number"])
	}
	body := group["body"].(map[string]any)
	if body["type"] != "sequence" {
		t.Errorf("body type = %v, want sequence", body["type"])
	}
}

func TestSubexpNamedCapture(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Subexp{
					GroupType: ast.GroupNamedCapture,
					Number:    1,
					Name:      "word",
					Regexp: &ast.Regexp{
						Matches: []*ast.Match{
							{Fragments: []*ast.MatchFragment{
								{Content: &ast.Escape{EscapeType: "word", Code: "w", Value: "\\w"}},
							}},
						},
					},
				}},
			}},
		},
	}
	out := mustRender(t, root, `(?<word>\w)`, "javascript")
	doc := unmarshal(t, out)
	seq := doc["root"].(map[string]any)
	elements := seq["elements"].([]any)
	group := elements[0].(map[string]any)
	if group["kind"] != "namedCapture" {
		t.Errorf("kind = %v, want namedCapture", group["kind"])
	}
	if group["name"] != "word" {
		t.Errorf("name = %v, want word", group["name"])
	}
	if group["number"] != float64(1) {
		t.Errorf("number = %v, want 1", group["number"])
	}
}

func TestSubexpNonCapture(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Subexp{
					GroupType: ast.GroupNonCapture,
					Regexp: &ast.Regexp{
						Matches: []*ast.Match{
							{Fragments: []*ast.MatchFragment{
								{Content: &ast.Literal{Text: "a"}},
							}},
						},
					},
				}},
			}},
		},
	}
	out := mustRender(t, root, "(?:a)", "javascript")
	doc := unmarshal(t, out)
	seq := doc["root"].(map[string]any)
	elements := seq["elements"].([]any)
	group := elements[0].(map[string]any)
	if group["kind"] != "nonCapture" {
		t.Errorf("kind = %v, want nonCapture", group["kind"])
	}
	if _, exists := group["number"]; exists {
		t.Error("non-capture group should not have number")
	}
}

func TestSubexpAllGroupTypes(t *testing.T) {
	tests := []struct {
		groupType string
		wantKind  string
	}{
		{ast.GroupPositiveLookahead, "positiveLookahead"},
		{ast.GroupNegativeLookahead, "negativeLookahead"},
		{ast.GroupPositiveLookbehind, "positiveLookbehind"},
		{ast.GroupNegativeLookbehind, "negativeLookbehind"},
		{ast.GroupAtomic, "atomic"},
	}
	for _, tt := range tests {
		t.Run(tt.wantKind, func(t *testing.T) {
			root := &ast.Regexp{
				Matches: []*ast.Match{
					{Fragments: []*ast.MatchFragment{
						{Content: &ast.Subexp{
							GroupType: tt.groupType,
							Regexp: &ast.Regexp{
								Matches: []*ast.Match{
									{Fragments: []*ast.MatchFragment{
										{Content: &ast.Literal{Text: "x"}},
									}},
								},
							},
						}},
					}},
				},
			}
			out := mustRender(t, root, "...", "pcre")
			doc := unmarshal(t, out)
			seq := doc["root"].(map[string]any)
			elements := seq["elements"].([]any)
			group := elements[0].(map[string]any)
			if group["kind"] != tt.wantKind {
				t.Errorf("kind = %v, want %v", group["kind"], tt.wantKind)
			}
		})
	}
}

func TestNestedGroups(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Subexp{
					GroupType: ast.GroupCapture,
					Number:    1,
					Regexp: &ast.Regexp{
						Matches: []*ast.Match{
							{Fragments: []*ast.MatchFragment{
								{Content: &ast.Subexp{
									GroupType: ast.GroupCapture,
									Number:    2,
									Regexp: &ast.Regexp{
										Matches: []*ast.Match{
											{Fragments: []*ast.MatchFragment{
												{Content: &ast.Literal{Text: "inner"}},
											}},
										},
									},
								}},
							}},
						},
					},
				}},
			}},
		},
	}
	out := mustRender(t, root, "((inner))", "javascript")
	doc := unmarshal(t, out)
	seq := doc["root"].(map[string]any)
	elements := seq["elements"].([]any)
	outer := elements[0].(map[string]any)
	if outer["number"] != float64(1) {
		t.Errorf("outer number = %v, want 1", outer["number"])
	}
	outerBody := outer["body"].(map[string]any)
	innerElements := outerBody["elements"].([]any)
	inner := innerElements[0].(map[string]any)
	if inner["number"] != float64(2) {
		t.Errorf("inner number = %v, want 2", inner["number"])
	}
}

func TestBackReference(t *testing.T) {
	t.Run("numbered", func(t *testing.T) {
		root := &ast.Regexp{
			Matches: []*ast.Match{
				{Fragments: []*ast.MatchFragment{
					{Content: &ast.BackReference{Number: 1}},
				}},
			},
		}
		out := mustRender(t, root, `\1`, "javascript")
		doc := unmarshal(t, out)
		seq := doc["root"].(map[string]any)
		elements := seq["elements"].([]any)
		br := elements[0].(map[string]any)
		if br["type"] != "backReference" {
			t.Errorf("type = %v, want backReference", br["type"])
		}
		if br["number"] != float64(1) {
			t.Errorf("number = %v, want 1", br["number"])
		}
		if _, exists := br["name"]; exists {
			t.Error("name should not be present for numbered backreference")
		}
	})

	t.Run("named", func(t *testing.T) {
		root := &ast.Regexp{
			Matches: []*ast.Match{
				{Fragments: []*ast.MatchFragment{
					{Content: &ast.BackReference{Name: "word"}},
				}},
			},
		}
		out := mustRender(t, root, `\k<word>`, "pcre")
		doc := unmarshal(t, out)
		seq := doc["root"].(map[string]any)
		elements := seq["elements"].([]any)
		br := elements[0].(map[string]any)
		if br["name"] != "word" {
			t.Errorf("name = %v, want word", br["name"])
		}
		if _, exists := br["number"]; exists {
			t.Error("number should not be present for named backreference")
		}
	})
}

func TestConditional(t *testing.T) {
	t.Run("with_false_branch", func(t *testing.T) {
		root := &ast.Regexp{
			Matches: []*ast.Match{
				{Fragments: []*ast.MatchFragment{
					{Content: &ast.Conditional{
						Condition: &ast.BackReference{Number: 1},
						TrueMatch: &ast.Regexp{
							Matches: []*ast.Match{
								{Fragments: []*ast.MatchFragment{
									{Content: &ast.Literal{Text: "yes"}},
								}},
							},
						},
						FalseMatch: &ast.Regexp{
							Matches: []*ast.Match{
								{Fragments: []*ast.MatchFragment{
									{Content: &ast.Literal{Text: "no"}},
								}},
							},
						},
					}},
				}},
			},
		}
		out := mustRender(t, root, "(?(1)yes|no)", "pcre")
		doc := unmarshal(t, out)
		seq := doc["root"].(map[string]any)
		elements := seq["elements"].([]any)
		cond := elements[0].(map[string]any)
		if cond["type"] != "conditional" {
			t.Errorf("type = %v, want conditional", cond["type"])
		}
		if cond["ifTrue"] == nil {
			t.Error("ifTrue is nil")
		}
		if cond["ifFalse"] == nil {
			t.Error("ifFalse is nil")
		}
	})

	t.Run("without_false_branch", func(t *testing.T) {
		root := &ast.Regexp{
			Matches: []*ast.Match{
				{Fragments: []*ast.MatchFragment{
					{Content: &ast.Conditional{
						Condition: &ast.BackReference{Number: 1},
						TrueMatch: &ast.Regexp{
							Matches: []*ast.Match{
								{Fragments: []*ast.MatchFragment{
									{Content: &ast.Literal{Text: "yes"}},
								}},
							},
						},
					}},
				}},
			},
		}
		out := mustRender(t, root, "(?(1)yes)", "pcre")
		doc := unmarshal(t, out)
		seq := doc["root"].(map[string]any)
		elements := seq["elements"].([]any)
		cond := elements[0].(map[string]any)
		if _, exists := cond["ifFalse"]; exists {
			t.Error("ifFalse should not be present when FalseMatch is nil")
		}
	})
}

func TestRecursiveRef(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.RecursiveRef{Target: "R"}},
			}},
		},
	}
	out := mustRender(t, root, "(?R)", "pcre")
	doc := unmarshal(t, out)
	seq := doc["root"].(map[string]any)
	elements := seq["elements"].([]any)
	rr := elements[0].(map[string]any)
	if rr["type"] != "recursiveReference" {
		t.Errorf("type = %v, want recursiveReference", rr["type"])
	}
	if rr["target"] != "R" {
		t.Errorf("target = %v, want R", rr["target"])
	}
}

func TestComment(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Comment{Text: "this is a comment"}},
			}},
		},
	}
	out := mustRender(t, root, "(?#this is a comment)", "pcre")
	doc := unmarshal(t, out)
	seq := doc["root"].(map[string]any)
	elements := seq["elements"].([]any)
	c := elements[0].(map[string]any)
	if c["type"] != "comment" {
		t.Errorf("type = %v, want comment", c["type"])
	}
	if c["text"] != "this is a comment" {
		t.Errorf("text = %v, want 'this is a comment'", c["text"])
	}
}

func TestQuotedLiteral(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.QuotedLiteral{Text: "hello.world"}},
			}},
		},
	}
	out := mustRender(t, root, `\Qhello.world\E`, "pcre")
	doc := unmarshal(t, out)
	seq := doc["root"].(map[string]any)
	elements := seq["elements"].([]any)
	ql := elements[0].(map[string]any)
	if ql["type"] != "quotedLiteral" {
		t.Errorf("type = %v, want quotedLiteral", ql["type"])
	}
	if ql["text"] != "hello.world" {
		t.Errorf("text = %v, want hello.world", ql["text"])
	}
}

func TestInlineModifier(t *testing.T) {
	t.Run("standalone", func(t *testing.T) {
		root := &ast.Regexp{
			Matches: []*ast.Match{
				{Fragments: []*ast.MatchFragment{
					{Content: &ast.InlineModifier{Enable: "im", Disable: "s"}},
				}},
			},
		}
		out := mustRender(t, root, "(?im-s)", "pcre")
		doc := unmarshal(t, out)
		seq := doc["root"].(map[string]any)
		elements := seq["elements"].([]any)
		im := elements[0].(map[string]any)
		if im["type"] != "inlineModifier" {
			t.Errorf("type = %v, want inlineModifier", im["type"])
		}
		if im["enable"] != "im" {
			t.Errorf("enable = %v, want im", im["enable"])
		}
		if im["disable"] != "s" {
			t.Errorf("disable = %v, want s", im["disable"])
		}
		if _, exists := im["body"]; exists {
			t.Error("body should not be present for standalone modifier")
		}
	})

	t.Run("scoped", func(t *testing.T) {
		root := &ast.Regexp{
			Matches: []*ast.Match{
				{Fragments: []*ast.MatchFragment{
					{Content: &ast.InlineModifier{
						Enable: "i",
						Regexp: &ast.Regexp{
							Matches: []*ast.Match{
								{Fragments: []*ast.MatchFragment{
									{Content: &ast.Literal{Text: "abc"}},
								}},
							},
						},
					}},
				}},
			},
		}
		out := mustRender(t, root, "(?i:abc)", "pcre")
		doc := unmarshal(t, out)
		seq := doc["root"].(map[string]any)
		elements := seq["elements"].([]any)
		im := elements[0].(map[string]any)
		if im["body"] == nil {
			t.Error("body should be present for scoped modifier")
		}
		if _, exists := im["disable"]; exists {
			t.Error("disable should be omitted when empty")
		}
	})
}

func TestBranchReset(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.BranchReset{
					Regexp: &ast.Regexp{
						Matches: []*ast.Match{
							{Fragments: []*ast.MatchFragment{
								{Content: &ast.Literal{Text: "a"}},
							}},
							{Fragments: []*ast.MatchFragment{
								{Content: &ast.Literal{Text: "b"}},
							}},
						},
					},
				}},
			}},
		},
	}
	out := mustRender(t, root, "(?|a|b)", "pcre")
	doc := unmarshal(t, out)
	seq := doc["root"].(map[string]any)
	elements := seq["elements"].([]any)
	br := elements[0].(map[string]any)
	if br["type"] != "branchReset" {
		t.Errorf("type = %v, want branchReset", br["type"])
	}
	body := br["body"].(map[string]any)
	if body["type"] != "alternation" {
		t.Errorf("body type = %v, want alternation", body["type"])
	}
}

func TestBalancedGroup(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.BalancedGroup{
					Name:      "Open",
					OtherName: "Close",
					Regexp: &ast.Regexp{
						Matches: []*ast.Match{
							{Fragments: []*ast.MatchFragment{
								{Content: &ast.Literal{Text: "x"}},
							}},
						},
					},
				}},
			}},
		},
	}
	out := mustRender(t, root, "(?<Open-Close>x)", "dotnet")
	doc := unmarshal(t, out)
	seq := doc["root"].(map[string]any)
	elements := seq["elements"].([]any)
	bg := elements[0].(map[string]any)
	if bg["type"] != "balancedGroup" {
		t.Errorf("type = %v, want balancedGroup", bg["type"])
	}
	if bg["name"] != "Open" {
		t.Errorf("name = %v, want Open", bg["name"])
	}
	if bg["otherName"] != "Close" {
		t.Errorf("otherName = %v, want Close", bg["otherName"])
	}
}

func TestBacktrackControl(t *testing.T) {
	t.Run("without_arg", func(t *testing.T) {
		root := &ast.Regexp{
			Matches: []*ast.Match{
				{Fragments: []*ast.MatchFragment{
					{Content: &ast.BacktrackControl{Verb: "FAIL"}},
				}},
			},
		}
		out := mustRender(t, root, "(*FAIL)", "pcre")
		doc := unmarshal(t, out)
		seq := doc["root"].(map[string]any)
		elements := seq["elements"].([]any)
		bc := elements[0].(map[string]any)
		if bc["type"] != "backtrackControl" {
			t.Errorf("type = %v, want backtrackControl", bc["type"])
		}
		if bc["verb"] != "FAIL" {
			t.Errorf("verb = %v, want FAIL", bc["verb"])
		}
		if _, exists := bc["arg"]; exists {
			t.Error("arg should be omitted when empty")
		}
	})

	t.Run("with_arg", func(t *testing.T) {
		root := &ast.Regexp{
			Matches: []*ast.Match{
				{Fragments: []*ast.MatchFragment{
					{Content: &ast.BacktrackControl{Verb: "MARK", Arg: "name"}},
				}},
			},
		}
		out := mustRender(t, root, "(*MARK:name)", "pcre")
		doc := unmarshal(t, out)
		seq := doc["root"].(map[string]any)
		elements := seq["elements"].([]any)
		bc := elements[0].(map[string]any)
		if bc["arg"] != "name" {
			t.Errorf("arg = %v, want name", bc["arg"])
		}
	})
}

func TestPatternOption(t *testing.T) {
	t.Run("without_value", func(t *testing.T) {
		root := &ast.Regexp{
			Matches: []*ast.Match{
				{Fragments: []*ast.MatchFragment{
					{Content: &ast.Literal{Text: "a"}},
				}},
			},
			Options: []*ast.PatternOption{
				{Name: "UTF"},
			},
		}
		out := mustRender(t, root, "(*UTF)a", "pcre")
		doc := unmarshal(t, out)
		rootNode := doc["root"].(map[string]any)
		opts := rootNode["options"].([]any)
		if len(opts) != 1 {
			t.Fatalf("options len = %d, want 1", len(opts))
		}
		opt := opts[0].(map[string]any)
		if opt["type"] != "patternOption" {
			t.Errorf("type = %v, want patternOption", opt["type"])
		}
		if opt["name"] != "UTF" {
			t.Errorf("name = %v, want UTF", opt["name"])
		}
		if _, exists := opt["value"]; exists {
			t.Error("value should be omitted when empty")
		}
	})

	t.Run("with_value", func(t *testing.T) {
		root := &ast.Regexp{
			Matches: []*ast.Match{
				{Fragments: []*ast.MatchFragment{
					{Content: &ast.Literal{Text: "a"}},
				}},
			},
			Options: []*ast.PatternOption{
				{Name: "LIMIT_MATCH", Value: "1000"},
			},
		}
		out := mustRender(t, root, "(*LIMIT_MATCH=1000)a", "pcre")
		doc := unmarshal(t, out)
		rootNode := doc["root"].(map[string]any)
		opts := rootNode["options"].([]any)
		opt := opts[0].(map[string]any)
		if opt["value"] != "1000" {
			t.Errorf("value = %v, want 1000", opt["value"])
		}
	})
}

func TestPatternOptionsOnAlternation(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Literal{Text: "a"}},
			}},
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Literal{Text: "b"}},
			}},
		},
		Options: []*ast.PatternOption{
			{Name: "UTF"},
		},
	}
	out := mustRender(t, root, "(*UTF)a|b", "pcre")
	doc := unmarshal(t, out)
	rootNode := doc["root"].(map[string]any)
	if rootNode["type"] != "alternation" {
		t.Errorf("type = %v, want alternation", rootNode["type"])
	}
	opts := rootNode["options"].([]any)
	if len(opts) != 1 {
		t.Errorf("options len = %d, want 1", len(opts))
	}
}

func TestCallout(t *testing.T) {
	t.Run("numeric", func(t *testing.T) {
		root := &ast.Regexp{
			Matches: []*ast.Match{
				{Fragments: []*ast.MatchFragment{
					{Content: &ast.Callout{Number: 42, Text: ""}},
				}},
			},
		}
		out := mustRender(t, root, "(?C42)", "pcre")
		doc := unmarshal(t, out)
		seq := doc["root"].(map[string]any)
		elements := seq["elements"].([]any)
		co := elements[0].(map[string]any)
		if co["type"] != "callout" {
			t.Errorf("type = %v, want callout", co["type"])
		}
		// *int stored in map[string]any serializes as a JSON number,
		// which unmarshals back as float64.
		num, exists := co["number"]
		if !exists || num == nil {
			t.Fatal("number should be present for numeric callout")
		}
		if num != float64(42) {
			t.Errorf("number = %v, want 42", num)
		}
	})

	t.Run("string", func(t *testing.T) {
		root := &ast.Regexp{
			Matches: []*ast.Match{
				{Fragments: []*ast.MatchFragment{
					{Content: &ast.Callout{Number: -1, Text: "hello"}},
				}},
			},
		}
		out := mustRender(t, root, `(?C"hello")`, "pcre")
		doc := unmarshal(t, out)
		seq := doc["root"].(map[string]any)
		elements := seq["elements"].([]any)
		co := elements[0].(map[string]any)
		if co["type"] != "callout" {
			t.Errorf("type = %v, want callout", co["type"])
		}
		if _, exists := co["number"]; exists {
			t.Error("number should not be present for string callout")
		}
		if co["text"] != "hello" {
			t.Errorf("text = %v, want hello", co["text"])
		}
	})

	t.Run("default_zero", func(t *testing.T) {
		root := &ast.Regexp{
			Matches: []*ast.Match{
				{Fragments: []*ast.MatchFragment{
					{Content: &ast.Callout{Number: 0, Text: ""}},
				}},
			},
		}
		out := mustRender(t, root, "(?C)", "pcre")
		doc := unmarshal(t, out)
		seq := doc["root"].(map[string]any)
		elements := seq["elements"].([]any)
		co := elements[0].(map[string]any)
		// Number 0 is valid and should be present
		if co["number"] == nil {
			t.Error("number should be present for (?C) which defaults to 0")
		}
	})
}

func TestFlagsOnRoot(t *testing.T) {
	t.Run("single_branch", func(t *testing.T) {
		root := &ast.Regexp{
			Matches: []*ast.Match{
				{Fragments: []*ast.MatchFragment{
					{Content: &ast.Literal{Text: "a"}},
				}},
			},
			Flags: "gi",
		}
		out := mustRender(t, root, "/a/gi", "javascript")
		doc := unmarshal(t, out)
		rootNode := doc["root"].(map[string]any)
		if rootNode["flags"] != "gi" {
			t.Errorf("flags = %v, want gi", rootNode["flags"])
		}
	})

	t.Run("multi_branch", func(t *testing.T) {
		root := &ast.Regexp{
			Matches: []*ast.Match{
				{Fragments: []*ast.MatchFragment{
					{Content: &ast.Literal{Text: "a"}},
				}},
				{Fragments: []*ast.MatchFragment{
					{Content: &ast.Literal{Text: "b"}},
				}},
			},
			Flags: "i",
		}
		out := mustRender(t, root, "/a|b/i", "javascript")
		doc := unmarshal(t, out)
		rootNode := doc["root"].(map[string]any)
		if rootNode["flags"] != "i" {
			t.Errorf("flags = %v, want i", rootNode["flags"])
		}
	})
}

func TestFlagsOmittedWhenEmpty(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Literal{Text: "a"}},
			}},
		},
	}
	out := mustRender(t, root, "a", "javascript")
	doc := unmarshal(t, out)
	rootNode := doc["root"].(map[string]any)
	if _, exists := rootNode["flags"]; exists {
		t.Error("flags should be omitted when empty")
	}
}

func TestExampleOutput(t *testing.T) {
	// Matches the example from the spec: foo([a-z]+)
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Literal{Text: "foo"}},
				{Content: &ast.Subexp{
					GroupType: ast.GroupCapture,
					Number:    1,
					Regexp: &ast.Regexp{
						Matches: []*ast.Match{
							{Fragments: []*ast.MatchFragment{
								{
									Content: &ast.Charset{
										Inverted: false,
										Items: []ast.CharsetItem{
											&ast.CharsetRange{First: "a", Last: "z"},
										},
									},
									Repeat: &ast.Repeat{Min: 1, Max: -1, Greedy: true},
								},
							}},
						},
					},
				}},
			}},
		},
	}
	out := mustRender(t, root, "foo([a-z]+)", "javascript")

	// Verify valid JSON
	if !json.Valid([]byte(out)) {
		t.Fatalf("output is not valid JSON:\n%s", out)
	}

	// Verify structure
	doc := unmarshal(t, out)
	if doc["pattern"] != "foo([a-z]+)" {
		t.Errorf("pattern = %v", doc["pattern"])
	}
	if doc["flavor"] != "javascript" {
		t.Errorf("flavor = %v", doc["flavor"])
	}

	rootNode := doc["root"].(map[string]any)
	if rootNode["type"] != "sequence" {
		t.Errorf("root type = %v, want sequence", rootNode["type"])
	}
	elements := rootNode["elements"].([]any)
	if len(elements) != 2 {
		t.Fatalf("elements len = %d, want 2", len(elements))
	}

	lit := elements[0].(map[string]any)
	if lit["value"] != "foo" {
		t.Errorf("first element value = %v, want foo", lit["value"])
	}

	group := elements[1].(map[string]any)
	if group["type"] != "group" || group["kind"] != "capture" {
		t.Errorf("second element = %v", group)
	}

	groupBody := group["body"].(map[string]any)
	groupElements := groupBody["elements"].([]any)
	cc := groupElements[0].(map[string]any)
	if cc["type"] != "characterClass" {
		t.Errorf("charset type = %v, want characterClass", cc["type"])
	}
	q := cc["quantifier"].(map[string]any)
	if q["min"] != float64(1) {
		t.Errorf("quantifier min = %v, want 1", q["min"])
	}
	if q["max"] != nil {
		t.Errorf("quantifier max = %v, want nil", q["max"])
	}
}

func TestEscapeInCharset(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Charset{
					Items: []ast.CharsetItem{
						&ast.Escape{EscapeType: "digit", Code: "d", Value: "\\d"},
					},
				}},
			}},
		},
	}
	out := mustRender(t, root, `[\d]`, "javascript")
	doc := unmarshal(t, out)
	seq := doc["root"].(map[string]any)
	elements := seq["elements"].([]any)
	cc := elements[0].(map[string]any)
	members := cc["members"].([]any)
	esc := members[0].(map[string]any)
	if esc["type"] != "escape" {
		t.Errorf("type = %v, want escape", esc["type"])
	}
}

func TestAllOutputIsValidJSON(t *testing.T) {
	// Build a complex AST with many node types
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Anchor{AnchorType: ast.AnchorStart}},
				{
					Content: &ast.Literal{Text: "hello"},
					Repeat:  &ast.Repeat{Min: 0, Max: 1, Greedy: false},
				},
				{Content: &ast.AnyCharacter{}},
				{Content: &ast.Escape{EscapeType: "digit", Code: "d", Value: "\\d"}},
				{Content: &ast.Charset{
					Inverted: false,
					Items: []ast.CharsetItem{
						&ast.CharsetRange{First: "0", Last: "9"},
						&ast.CharsetLiteral{Text: "_"},
					},
				}},
				{Content: &ast.Subexp{
					GroupType: ast.GroupCapture,
					Number:    1,
					Regexp: &ast.Regexp{
						Matches: []*ast.Match{
							{Fragments: []*ast.MatchFragment{
								{Content: &ast.Literal{Text: "world"}},
							}},
						},
					},
				}},
				{Content: &ast.BackReference{Number: 1}},
				{Content: &ast.Anchor{AnchorType: ast.AnchorEnd}},
			}},
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Literal{Text: "alt"}},
			}},
		},
		Flags: "gi",
	}
	out := mustRender(t, root, `^hello?.[\d][0-9_](world)\1$|alt`, "javascript")
	if !json.Valid([]byte(out)) {
		t.Fatalf("complex AST produced invalid JSON:\n%s", out)
	}
}
