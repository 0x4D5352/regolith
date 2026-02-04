package flavor

import (
	"testing"

	"github.com/0x4d5352/regolith/internal/ast"
)

// mockFlavor is a test implementation of the Flavor interface
type mockFlavor struct {
	name        string
	description string
}

func (m *mockFlavor) Name() string        { return m.name }
func (m *mockFlavor) Description() string { return m.description }
func (m *mockFlavor) Parse(pattern string) (*ast.Regexp, error) {
	return &ast.Regexp{}, nil
}
func (m *mockFlavor) SupportedFlags() []FlagInfo {
	return []FlagInfo{{Char: 'i', Name: "ignoreCase", Description: "Case insensitive"}}
}
func (m *mockFlavor) SupportedFeatures() FeatureSet {
	return FeatureSet{Lookahead: true}
}

func TestRegister(t *testing.T) {
	// Clear registry for test
	registryLock.Lock()
	originalRegistry := registry
	registry = make(map[string]Flavor)
	registryLock.Unlock()
	defer func() {
		registryLock.Lock()
		registry = originalRegistry
		registryLock.Unlock()
	}()

	mock := &mockFlavor{name: "test", description: "Test flavor"}
	Register(mock)

	if Count() != 1 {
		t.Errorf("expected 1 registered flavor, got %d", Count())
	}

	f, ok := Get("test")
	if !ok {
		t.Fatal("expected to find 'test' flavor")
	}
	if f.Name() != "test" {
		t.Errorf("expected name 'test', got '%s'", f.Name())
	}
}

func TestGet(t *testing.T) {
	// Clear registry for test
	registryLock.Lock()
	originalRegistry := registry
	registry = make(map[string]Flavor)
	registryLock.Unlock()
	defer func() {
		registryLock.Lock()
		registry = originalRegistry
		registryLock.Unlock()
	}()

	// Test getting non-existent flavor
	_, ok := Get("nonexistent")
	if ok {
		t.Error("expected Get to return false for non-existent flavor")
	}

	// Register and get
	mock := &mockFlavor{name: "test", description: "Test flavor"}
	Register(mock)

	f, ok := Get("test")
	if !ok {
		t.Fatal("expected to find 'test' flavor")
	}
	if f.Description() != "Test flavor" {
		t.Errorf("expected description 'Test flavor', got '%s'", f.Description())
	}
}

func TestList(t *testing.T) {
	// Clear registry for test
	registryLock.Lock()
	originalRegistry := registry
	registry = make(map[string]Flavor)
	registryLock.Unlock()
	defer func() {
		registryLock.Lock()
		registry = originalRegistry
		registryLock.Unlock()
	}()

	// Empty registry
	if len(List()) != 0 {
		t.Error("expected empty list for empty registry")
	}

	// Add flavors
	Register(&mockFlavor{name: "zebra", description: "Zebra flavor"})
	Register(&mockFlavor{name: "alpha", description: "Alpha flavor"})
	Register(&mockFlavor{name: "beta", description: "Beta flavor"})

	list := List()
	if len(list) != 3 {
		t.Fatalf("expected 3 flavors, got %d", len(list))
	}

	// Should be sorted
	if list[0] != "alpha" || list[1] != "beta" || list[2] != "zebra" {
		t.Errorf("expected sorted list [alpha, beta, zebra], got %v", list)
	}
}

func TestAll(t *testing.T) {
	// Clear registry for test
	registryLock.Lock()
	originalRegistry := registry
	registry = make(map[string]Flavor)
	registryLock.Unlock()
	defer func() {
		registryLock.Lock()
		registry = originalRegistry
		registryLock.Unlock()
	}()

	Register(&mockFlavor{name: "test1", description: "Test 1"})
	Register(&mockFlavor{name: "test2", description: "Test 2"})

	all := All()
	if len(all) != 2 {
		t.Fatalf("expected 2 flavors, got %d", len(all))
	}

	// Verify it's a copy
	delete(all, "test1")
	if Count() != 2 {
		t.Error("modifying returned map should not affect registry")
	}
}

func TestCount(t *testing.T) {
	// Clear registry for test
	registryLock.Lock()
	originalRegistry := registry
	registry = make(map[string]Flavor)
	registryLock.Unlock()
	defer func() {
		registryLock.Lock()
		registry = originalRegistry
		registryLock.Unlock()
	}()

	if Count() != 0 {
		t.Error("expected count 0 for empty registry")
	}

	Register(&mockFlavor{name: "test", description: "Test"})
	if Count() != 1 {
		t.Error("expected count 1 after registering one flavor")
	}
}

func TestRegisterOverwrite(t *testing.T) {
	// Clear registry for test
	registryLock.Lock()
	originalRegistry := registry
	registry = make(map[string]Flavor)
	registryLock.Unlock()
	defer func() {
		registryLock.Lock()
		registry = originalRegistry
		registryLock.Unlock()
	}()

	Register(&mockFlavor{name: "test", description: "Original"})
	Register(&mockFlavor{name: "test", description: "Updated"})

	if Count() != 1 {
		t.Error("expected count 1 after overwriting")
	}

	f, _ := Get("test")
	if f.Description() != "Updated" {
		t.Errorf("expected description 'Updated', got '%s'", f.Description())
	}
}
