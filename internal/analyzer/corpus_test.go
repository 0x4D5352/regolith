package analyzer

import (
	"testing"
)

func TestGenerateCorpusDeterministic(t *testing.T) {
	a := GenerateCorpus("prose", 1000)
	b := GenerateCorpus("prose", 1000)
	if a != b {
		t.Error("GenerateCorpus is not deterministic")
	}
}

func TestGenerateCorpusSize(t *testing.T) {
	types := []string{"prose", "json", "yaml", "repeated", "random"}
	for _, ct := range types {
		t.Run(ct, func(t *testing.T) {
			result := GenerateCorpus(ct, 500)
			if len(result) < 500 {
				t.Errorf("%s: got length %d, want at least 500", ct, len(result))
			}
		})
	}
}

func TestGenerateCorpusUnknownType(t *testing.T) {
	result := GenerateCorpus("nonexistent", 100)
	if result != "" {
		t.Errorf("expected empty string for unknown type, got length %d", len(result))
	}
}
