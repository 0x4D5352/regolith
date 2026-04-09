package analyzer

import (
	"os/exec"
	"time"
)

// Engine can execute a regex pattern against an input string and
// measure how long the match takes.
type Engine interface {
	// Name returns the engine identifier, e.g. "node", "python3", "regexp2".
	Name() string

	// Run executes the pattern against input with a timeout.
	// Returns the measured duration, or an error if the timeout was
	// exceeded or the engine failed.
	Run(pattern, input string, timeout time.Duration) (time.Duration, error)
}

// DetectEngine returns the best available engine for the given flavor.
// It uses exec.LookPath for fast detection and falls back to regexp2.
// The second return value is true when using the fallback engine.
func DetectEngine(flavorName string) (Engine, bool) {
	primary := primaryEngineCmd(flavorName)
	if primary != "" {
		if _, err := exec.LookPath(primary); err == nil {
			eng := newExternalEngine(flavorName, primary)
			if eng != nil {
				return eng, false
			}
		}
	}
	return &Regexp2Engine{}, true
}

func primaryEngineCmd(flavorName string) string {
	switch flavorName {
	case "javascript":
		return "node"
	case "java":
		return "java"
	case "pcre":
		return "python3"
	case "dotnet":
		return "python3"
	case "posix-bre", "posix-ere", "gnugrep-bre", "gnugrep-ere":
		return "grep"
	default:
		return ""
	}
}

// newExternalEngine creates the appropriate external engine for a flavor.
// Returns nil if the flavor has no external engine support yet.
// External engines are added in Tasks 14-15.
func newExternalEngine(flavorName, cmd string) Engine {
	return nil
}
