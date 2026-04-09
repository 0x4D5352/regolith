package analyzer

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
)

// GenerateCorpus produces a deterministic test string of at least the
// given size for the specified corpus type. Each type uses a fixed seed
// so results are reproducible across runs.
func GenerateCorpus(corpusType string, size int) string {
	switch corpusType {
	case "prose":
		return generateProse(size)
	case "json":
		return generateJSON(size)
	case "yaml":
		return generateYAML(size)
	case "repeated":
		return generateRepeated(size)
	case "random":
		return generateRandom(size)
	default:
		return ""
	}
}

// CorpusTypes returns the list of valid built-in corpus type names.
func CorpusTypes() []string {
	return []string{"prose", "json", "yaml", "repeated", "random"}
}

var proseSentences = []string{
	"The quick brown fox jumps over the lazy dog. ",
	"Pack my box with five dozen liquor jugs. ",
	"How vexingly quick daft zebras jump. ",
	"Sphinx of black quartz judge my vow. ",
	"Two driven jocks help fax my big quiz. ",
	"The five boxing wizards jump quickly. ",
	"Jackdaws love my big sphinx of quartz. ",
	"A wizard that can conjure lightning is powerful. ",
	"Regular expressions are a notation for describing sets of character strings. ",
	"Performance testing reveals issues before production deployment. ",
}

func generateProse(size int) string {
	rng := rand.New(rand.NewSource(42))
	var sb strings.Builder
	for sb.Len() < size {
		sb.WriteString(proseSentences[rng.Intn(len(proseSentences))])
	}
	return sb.String()[:size]
}

func generateJSON(size int) string {
	rng := rand.New(rand.NewSource(43))
	var sb strings.Builder
	keys := []string{"name", "value", "count", "enabled", "label", "data", "items", "type"}
	vals := []string{`"hello"`, `"world"`, `42`, `true`, `false`, `null`, `"test"`, `3.14`}

	sb.WriteString("{")
	for sb.Len() < size {
		key := keys[rng.Intn(len(keys))]
		val := vals[rng.Intn(len(vals))]
		if sb.Len() > 1 {
			sb.WriteString(", ")
		}
		sb.WriteString(`"` + key + `": ` + val)
	}
	sb.WriteString("}")
	return sb.String()[:size]
}

func generateYAML(size int) string {
	rng := rand.New(rand.NewSource(44))
	var sb strings.Builder
	keys := []string{"name", "version", "debug", "port", "host", "timeout", "retries", "level"}
	vals := []string{"hello", "1.0.0", "true", "8080", "localhost", "30s", "3", "info"}

	for sb.Len() < size {
		key := keys[rng.Intn(len(keys))]
		val := vals[rng.Intn(len(vals))]
		sb.WriteString(key + ": " + val + "\n")
	}
	return sb.String()[:size]
}

func generateRepeated(size int) string {
	return strings.Repeat("a", size)
}

func generateRandom(size int) string {
	rng := rand.New(rand.NewSource(45))
	const printable = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 !@#$%^&*()-_=+[]{}|;:',.<>?/\"\\\t\n"
	sb := make([]byte, size)
	for i := range size {
		sb[i] = printable[rng.Intn(len(printable))]
	}
	return string(sb)
}

// LoadCorpusFile reads a custom corpus file and returns its content
// truncated or repeated to the requested size.
func LoadCorpusFile(path string, size int) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading corpus file: %w", err)
	}
	content := string(data)
	if len(content) == 0 {
		return "", fmt.Errorf("corpus file is empty: %s", path)
	}
	for len(content) < size {
		content += string(data)
	}
	return content[:size], nil
}
