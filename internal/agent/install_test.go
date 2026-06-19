package agent

import (
	"testing"
)

func TestResolveInstalls_NilDefaultsToClaude(t *testing.T) {
	result, err := ResolveInstalls(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0].Name != "claude" {
		names := make([]string, len(result))
		for i, r := range result {
			names[i] = r.Name
		}
		t.Fatalf("expected [claude], got %v", names)
	}
}

func TestResolveInstalls_EmptyMeansNone(t *testing.T) {
	result, err := ResolveInstalls(map[string]bool{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 0 {
		t.Fatalf("expected no installs, got %d", len(result))
	}
}

func TestResolveInstalls_ExplicitAll(t *testing.T) {
	all := map[string]bool{"claude": true, "codex": true, "gemini": true, "opencode": true, "pi": true}
	result, err := ResolveInstalls(all, []string{"node"})
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 5 {
		t.Fatalf("expected 5 installs, got %d", len(result))
	}
}

func TestResolveInstalls_SpecificSelection(t *testing.T) {
	sel := map[string]bool{"gemini": true}
	result, err := ResolveInstalls(sel, []string{"node"})
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0].Name != "gemini" {
		t.Fatalf("expected [gemini], got %v", result)
	}
}

func TestResolveInstalls_UnknownAgent(t *testing.T) {
	sel := map[string]bool{"unknown": true}
	_, err := ResolveInstalls(sel, nil)
	if err == nil {
		t.Fatal("expected error for unknown agent")
	}
}

func TestAllInstallNames(t *testing.T) {
	names := AllInstallNames()
	if len(names) != 6 {
		t.Fatalf("expected 6 agent installs, got %d: %v", len(names), names)
	}
	expected := []string{"claude", "codex", "copilot", "gemini", "opencode", "pi"}
	for i, name := range expected {
		if names[i] != name {
			t.Errorf("names[%d] = %q, want %q", i, names[i], name)
		}
	}
}
