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
	empty := []string{}
	result, err := ResolveInstalls(&empty, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 0 {
		t.Fatalf("expected no installs, got %d", len(result))
	}
}

func TestResolveInstalls_ExplicitAll(t *testing.T) {
	all := []string{"claude", "codex", "gemini", "opencode"}
	result, err := ResolveInstalls(&all, []string{"node"})
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 4 {
		t.Fatalf("expected 4 installs, got %d", len(result))
	}
}

func TestResolveInstalls_SpecificSelection(t *testing.T) {
	sel := []string{"gemini"}
	result, err := ResolveInstalls(&sel, []string{"node"})
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0].Name != "gemini" {
		t.Fatalf("expected [gemini], got %v", result)
	}
}

func TestResolveInstalls_UnknownAgent(t *testing.T) {
	sel := []string{"unknown"}
	_, err := ResolveInstalls(&sel, nil)
	if err == nil {
		t.Fatal("expected error for unknown agent")
	}
}

func TestResolveInstalls_Deduplication(t *testing.T) {
	sel := []string{"claude", "claude"}
	result, err := ResolveInstalls(&sel, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 install after dedup, got %d", len(result))
	}
}

func TestAllInstallNames(t *testing.T) {
	names := AllInstallNames()
	if len(names) != 4 {
		t.Fatalf("expected 4 agent installs, got %d: %v", len(names), names)
	}
	expected := []string{"claude", "codex", "gemini", "opencode"}
	for i, name := range expected {
		if names[i] != name {
			t.Errorf("names[%d] = %q, want %q", i, names[i], name)
		}
	}
}
