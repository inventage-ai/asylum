package firstrun

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// BuildConfig output is consumed by config.Load via yaml.Unmarshal, so the
// most important invariant is "parses as valid YAML". The structural checks
// below assert the shape we actually care about for callers.

func TestBuildConfig_ParsesAsYAML(t *testing.T) {
	out := BuildConfig(Choices{
		DefaultAgent: "claude",
		Agents:       map[string]bool{"claude": true, "gemini": true},
		Kits:         map[string]bool{"java": true, "docker": false, "github": true},
	})

	var parsed map[string]any
	if err := yaml.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("generated config is not valid YAML: %v\n---\n%s", err, out)
	}

	if parsed["agent"] != "claude" {
		t.Errorf("agent = %v, want claude", parsed["agent"])
	}
	if parsed["release-channel"] != "stable" {
		t.Errorf("release-channel = %v, want stable", parsed["release-channel"])
	}
}

func TestBuildConfig_DefaultAgentReflectsChoice(t *testing.T) {
	out := BuildConfig(Choices{
		DefaultAgent: "gemini",
		Agents:       map[string]bool{"gemini": true},
		Kits:         map[string]bool{},
	})

	var parsed struct {
		Agent string `yaml:"agent"`
	}
	if err := yaml.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed.Agent != "gemini" {
		t.Errorf("agent = %q, want gemini", parsed.Agent)
	}
}

func TestBuildConfig_AgentsBlockReflectsSelection(t *testing.T) {
	out := BuildConfig(Choices{
		DefaultAgent: "claude",
		Agents:       map[string]bool{"claude": true, "gemini": true},
		Kits:         map[string]bool{},
	})

	// Active agents must be uncommented at the right indent.
	for _, name := range []string{"claude", "gemini"} {
		if !strings.Contains(out, "  "+name+":\n") {
			t.Errorf("active agent %q not present as `  %s:`\n---\n%s", name, name, out)
		}
	}

	// Unselected real agents must appear commented; echo must not appear at all.
	if strings.Contains(out, "echo:") || strings.Contains(out, "# echo:") {
		t.Errorf("echo should not appear in generated config:\n%s", out)
	}
	for _, banned := range []string{"codex", "copilot", "opencode", "pi"} {
		if !strings.Contains(out, "# "+banned+":\n") {
			t.Errorf("unselected agent %q should appear as commented entry\n---\n%s", banned, out)
		}
		// Negative: the unselected entry must not be active.
		if strings.Contains(out, "\n  "+banned+":\n") {
			t.Errorf("unselected agent %q appears active\n---\n%s", banned, out)
		}
	}
}

func TestBuildConfig_KitsBlockReflectsSelection(t *testing.T) {
	out := BuildConfig(Choices{
		DefaultAgent: "claude",
		Agents:       map[string]bool{"claude": true},
		Kits:         map[string]bool{"java": true, "docker": false},
	})

	// java is TierDefault, snippet authored active — selected and kept active.
	if !strings.Contains(out, "\n  java:\n") {
		t.Errorf("java should be active in output:\n%s", out)
	}
	// docker is TierDefault, snippet authored active — deselected and must be commented.
	if !strings.Contains(out, "\n  # docker:") {
		t.Errorf("docker should appear commented in output:\n%s", out)
	}
	if strings.Contains(out, "\n  docker:") {
		t.Errorf("docker should not be active in output:\n%s", out)
	}
}

func TestBuildConfig_NonSelectableKitsVerbatim(t *testing.T) {
	// node (TierAlwaysOn) and ports (TierAlwaysOn) sit outside the wizard's
	// selection set; their authored snippets must reach the generated file
	// without any transformation.
	out := BuildConfig(Choices{
		DefaultAgent: "claude",
		Agents:       map[string]bool{"claude": true},
		Kits:         map[string]bool{},
	})

	// node's authored snippet starts active (multiple inner keys including
	// "shadow-node-modules: true").
	if !strings.Contains(out, "  node:\n    shadow-node-modules: true") {
		t.Errorf("node snippet not present verbatim:\n%s", out)
	}
	// ports's authored snippet is commented in the registry.
	if !strings.Contains(out, "  # ports:") || !strings.Contains(out, "  #   count: 5") {
		t.Errorf("ports commented snippet not preserved:\n%s", out)
	}
}

func TestBuildConfig_PreservesAuthoredFormatWhenChoiceMatches(t *testing.T) {
	// A TierDefault kit the user accepts (java) and a TierOptIn kit the user
	// leaves alone (cx) should both come through with their authored snippet
	// formatting intact — no transformation reflow.
	out := BuildConfig(Choices{
		DefaultAgent: "claude",
		Agents:       map[string]bool{"claude": true},
		Kits:         map[string]bool{"java": true /* cx omitted = false */},
	})

	expected := []string{
		// java authored snippet (verbatim multi-line block)
		"  java:\n    versions:\n      - 17\n      - 21\n      - 25\n    default-version: 21\n",
		// cx authored commented snippet (verbatim block with original 2-space + "# " indent)
		"  # cx:                 # Semantic code navigation\n  #   packages:        # tree-sitter language grammars\n  #     - python\n  #     - typescript\n  #     - go\n",
	}
	for _, want := range expected {
		if !strings.Contains(out, want) {
			t.Errorf("expected block not found verbatim:\n--- want ---\n%s\n--- got ---\n%s", want, out)
		}
	}
}

func TestBuildConfig_GroupsActiveBeforeCommented(t *testing.T) {
	out := BuildConfig(Choices{
		DefaultAgent: "claude",
		Agents:       map[string]bool{"claude": true},
		Kits:         map[string]bool{"github": true /* others omitted */},
	})

	activeIdx := strings.Index(out, "  github:")
	commentedIdx := strings.Index(out, "  # docker:")
	if activeIdx == -1 || commentedIdx == -1 {
		t.Fatalf("expected markers missing:\n%s", out)
	}
	if activeIdx >= commentedIdx {
		t.Errorf("active kits should appear before commented ones (github at %d, docker at %d)", activeIdx, commentedIdx)
	}
}

func TestBuildConfig_DeselectedKitWithInnerCommentsHasNoDoubleComment(t *testing.T) {
	// python's authored ConfigSnippet is active at the outer level but
	// carries inner `#` hint lines. Deselecting it must not produce
	// `# # versions:` style double-commented lines — those are valid YAML
	// but unsightly. The fix is for commentSnippet to leave already-
	// commented lines alone; this asserts the resulting block has no
	// `# # ` sequence.
	out := BuildConfig(Choices{
		DefaultAgent: "claude",
		Agents:       map[string]bool{"claude": true},
		// python omitted → false → must be commented in output
		Kits: map[string]bool{},
	})

	if strings.Contains(out, "# # ") {
		t.Errorf("generated config contains double-comment sequence `# # `:\n%s", out)
	}

	// The outer `python:` key must still appear commented exactly once.
	if !strings.Contains(out, "  # python:\n") {
		t.Errorf("expected `  # python:` (commented kit header) in output:\n%s", out)
	}

	// Inner authored hint lines must survive unchanged at their original
	// indent and original single-`#` prefix.
	for _, expectedLine := range []string{
		"    # versions:\n",
		"    #   - 3.14\n",
		"    # packages:",
	} {
		if !strings.Contains(out, expectedLine) {
			t.Errorf("expected authored hint line %q to survive verbatim:\n%s", expectedLine, out)
		}
	}
}

func TestWriteConfig_RoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	choices := Choices{
		DefaultAgent: "claude",
		Agents:       map[string]bool{"claude": true},
		Kits:         map[string]bool{"java": true},
	}
	if err := WriteConfig(path, choices); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(data); got != BuildConfig(choices) {
		t.Errorf("WriteConfig output diverges from BuildConfig")
	}
}
