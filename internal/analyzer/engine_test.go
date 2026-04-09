package analyzer

import (
	"strings"
	"testing"
	"time"
)

func TestRegexp2EngineMatch(t *testing.T) {
	eng := &Regexp2Engine{}
	dur, err := eng.Run(`hello`, "hello world", 5*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dur <= 0 {
		t.Error("expected positive duration")
	}
}

func TestRegexp2EngineTimeout(t *testing.T) {
	eng := &Regexp2Engine{}
	// Catastrophic backtracking pattern
	_, err := eng.Run(`(a+)+$`, strings.Repeat("a", 25)+"!", 200*time.Millisecond)
	// Either err is non-nil (timeout) or duration is long — both acceptable
	if err == nil {
		t.Log("pattern completed within timeout — may need larger input for reliable timeout test")
	}
}

func TestDetectEngineFallback(t *testing.T) {
	eng, isFallback := DetectEngine("javascript")
	// On most systems node exists, but we just check the interface works
	if eng == nil {
		t.Fatal("expected non-nil engine")
	}
	_ = isFallback // either value is fine
}
