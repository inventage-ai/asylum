package image

import (
	"strings"
	"testing"

	"github.com/inventage-ai/asylum/assets"
	"github.com/inventage-ai/asylum/internal/agent"
	"github.com/inventage-ai/asylum/internal/kit"
	"github.com/inventage-ai/asylum/internal/versions"
)

func allAgentInstalls(t *testing.T) []*agent.AgentInstall {
	t.Helper()
	all := map[string]bool{"claude": true, "codex": true, "gemini": true, "opencode": true, "copilot": true, "pi": true}
	installs, err := agent.ResolveInstalls(all, []string{"node"})
	if err != nil {
		t.Fatal(err)
	}
	return installs
}

func claudeOnlyInstalls(t *testing.T) []*agent.AgentInstall {
	t.Helper()
	installs, err := agent.ResolveInstalls(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	return installs
}

// testOrderedDockerfile is a test helper that computes kit source order and
// assembles the Dockerfile (kits ordered + agent block last), returning the
// result and the ordered kit IDs.
func testOrderedDockerfile(profiles []*kit.Kit, agents []*agent.AgentInstall) ([]byte, []string) {
	sources := collectSources(profiles, nil)
	orderedIDs := computeSourceOrder(sources, nil)
	snippetOf := map[string]string{}
	for _, s := range sources {
		snippetOf[s.ID] = s.Snippet
	}
	agentBlock := agent.AssembleVersionedAgentSnippets(agents, nil)
	return assembleDockerfile(orderedIDs, snippetOf, "", agentBlock), orderedIDs
}

func TestAssembleDockerfile_AllProfiles(t *testing.T) {
	profiles, err := kit.Resolve(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	df, _ := testOrderedDockerfile(profiles, allAgentInstalls(t))
	s := string(df)

	if !strings.HasPrefix(s, string(assets.DockerfileCore)) {
		t.Error("assembled Dockerfile should start with core")
	}
	if !strings.HasSuffix(s, string(assets.DockerfileTail)) {
		t.Error("assembled Dockerfile should end with tail")
	}
	if !strings.Contains(s, "mise install java") {
		t.Error("missing java profile snippet")
	}
	if !strings.Contains(s, "uv tool install black") {
		t.Error("missing python profile snippet")
	}
	if !strings.Contains(s, "npm install -g") {
		t.Error("missing node profile snippet")
	}
}

func TestAssembleDockerfile_NoProfiles(t *testing.T) {
	noKits := []string{}
	profiles, err := kit.Resolve(noKits, nil)
	if err != nil {
		t.Fatal(err)
	}
	df, _ := testOrderedDockerfile(profiles, claudeOnlyInstalls(t))
	s := string(df)

	if !strings.Contains(s, string(assets.DockerfileCore)) {
		t.Error("should contain core")
	}
	if !strings.Contains(s, string(assets.DockerfileTail)) {
		t.Error("should contain tail")
	}
	if strings.Contains(s, "mise install java") {
		t.Error("should not contain java snippet")
	}
}

func TestAssembleDockerfile_AgentSnippets(t *testing.T) {
	profiles, _ := kit.Resolve(nil, nil)

	t.Run("all agents", func(t *testing.T) {
		df, _ := testOrderedDockerfile(profiles, allAgentInstalls(t))
		s := string(df)
		if !strings.Contains(s, "claude.ai/install.sh") {
			t.Error("missing claude install snippet")
		}
		if !strings.Contains(s, "gemini-cli") {
			t.Error("missing gemini install snippet")
		}
		if !strings.Contains(s, "@openai/codex") {
			t.Error("missing codex install snippet")
		}
		if !strings.Contains(s, "opencode") {
			t.Error("missing opencode install snippet")
		}
	})

	t.Run("claude only (default)", func(t *testing.T) {
		df, _ := testOrderedDockerfile(profiles, claudeOnlyInstalls(t))
		s := string(df)
		if !strings.Contains(s, "claude.ai/install.sh") {
			t.Error("missing claude install snippet")
		}
		if strings.Contains(s, "gemini-cli") {
			t.Error("should not contain gemini snippet")
		}
		if strings.Contains(s, "@openai/codex") {
			t.Error("should not contain codex snippet")
		}
	})

	t.Run("no agents", func(t *testing.T) {
		empty := map[string]bool{}
		noAgents, _ := agent.ResolveInstalls(empty, nil)
		df, _ := testOrderedDockerfile(profiles, noAgents)
		s := string(df)
		if strings.Contains(s, "claude.ai/install.sh") {
			t.Error("should not contain claude snippet")
		}
	})
}

func TestAssembleEntrypoint_AllProfiles(t *testing.T) {
	profiles, err := kit.Resolve(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	ep := assembleEntrypoint(profiles, allAgentInstalls(t))
	s := string(ep)

	if !strings.Contains(s, "mise activate bash") {
		t.Error("missing java entrypoint snippet")
	}
	if !strings.Contains(s, "has_python_marker") {
		t.Error("missing python/uv entrypoint snippet")
	}
}

func TestAssembleEntrypoint_NoProfiles(t *testing.T) {
	noKits := []string{}
	profiles, err := kit.Resolve(noKits, nil)
	if err != nil {
		t.Fatal(err)
	}
	ep := assembleEntrypoint(profiles, nil)
	s := string(ep)

	if strings.Contains(s, "ASYLUM_JAVA_VERSION") {
		t.Error("should not contain java version selection snippet")
	}
	if !strings.Contains(s, "Asylum entrypoint") {
		t.Error("should contain core")
	}
	if !strings.Contains(s, "exec \"$@\"") {
		t.Error("should contain tail")
	}
}

func TestAssembleEntrypoint_BannerLines(t *testing.T) {
	profiles, _ := kit.Resolve(nil, nil)

	t.Run("all profiles and agents", func(t *testing.T) {
		ep := string(assembleEntrypoint(profiles, allAgentInstalls(t)))
		if !strings.Contains(ep, "Python:") {
			t.Error("banner should contain Python version line")
		}
		if !strings.Contains(ep, "Java:") {
			t.Error("banner should contain Java version line")
		}
		if !strings.Contains(ep, "Claude:") {
			t.Error("banner should contain Claude version line")
		}
		if !strings.Contains(ep, "Gemini:") {
			t.Error("banner should contain Gemini version line")
		}
	})

	t.Run("claude only agent", func(t *testing.T) {
		ep := string(assembleEntrypoint(profiles, claudeOnlyInstalls(t)))
		if !strings.Contains(ep, "Claude:") {
			t.Error("banner should contain Claude version line")
		}
		if strings.Contains(ep, "Gemini:") {
			t.Error("banner should NOT contain Gemini when not installed")
		}
		if strings.Contains(ep, "Codex:") {
			t.Error("banner should NOT contain Codex when not installed")
		}
	})

	t.Run("no agents", func(t *testing.T) {
		empty := map[string]bool{}
		noAgents, _ := agent.ResolveInstalls(empty, nil)
		ep := string(assembleEntrypoint(profiles, noAgents))
		if strings.Contains(ep, "Claude:") {
			t.Error("banner should NOT contain Claude when no agents")
		}
	})
}

func TestAssembleDockerfile_PackageBlockPlacement(t *testing.T) {
	profiles, _ := kit.Resolve(nil, nil)
	sources := collectSources(profiles, nil)
	orderedIDs := computeSourceOrder(sources, nil)
	snippetOf := map[string]string{}
	for _, s := range sources {
		snippetOf[s.ID] = s.Snippet
	}
	packageBlock, err := basePackageBlock(map[string][]string{"npm": {"@mermaid-js/mermaid-cli"}})
	if err != nil {
		t.Fatal(err)
	}
	agentBlock := agent.AssembleVersionedAgentSnippets(claudeOnlyInstalls(t), nil)

	df := string(assembleDockerfile(orderedIDs, snippetOf, packageBlock, agentBlock))

	kitIdx := strings.Index(df, "# Install Node.js global packages") // node kit snippet
	pkgIdx := strings.Index(df, "@mermaid-js/mermaid-cli")            // global package block
	agentIdx := strings.Index(df, "claude.ai/install.sh")            // agent block
	if kitIdx < 0 || pkgIdx < 0 || agentIdx < 0 {
		t.Fatalf("missing markers: kit=%d pkg=%d agent=%d", kitIdx, pkgIdx, agentIdx)
	}
	if !(kitIdx < pkgIdx && pkgIdx < agentIdx) {
		t.Errorf("package block must sit after kit snippets and before agent block: kit=%d pkg=%d agent=%d", kitIdx, pkgIdx, agentIdx)
	}
}

func TestBaseHash_ChangesWithGlobalPackages(t *testing.T) {
	profiles, _ := kit.Resolve(nil, nil)
	agents := claudeOnlyInstalls(t)
	_, order := testOrderedDockerfile(profiles, agents)
	sources := collectSources(profiles, nil)
	snippetOf := map[string]string{}
	for _, s := range sources {
		snippetOf[s.ID] = s.Snippet
	}
	agentBlock := agent.AssembleVersionedAgentSnippets(agents, nil)

	none := baseHash(order, snippetOf, "", agentBlock, profiles, agents)
	block, err := basePackageBlock(map[string][]string{"npm": {"turbo"}})
	if err != nil {
		t.Fatal(err)
	}
	with := baseHash(order, snippetOf, block, agentBlock, profiles, agents)
	if none == with {
		t.Error("adding a global package should change the base hash")
	}
}

func TestBaseHash_DeterministicAndChanges(t *testing.T) {
	profiles, _ := kit.Resolve(nil, nil)
	agents1 := allAgentInstalls(t)

	_, order1 := testOrderedDockerfile(profiles, agents1)
	sources1 := collectSources(profiles, nil)
	snippetOf1 := map[string]string{}
	for _, s := range sources1 {
		snippetOf1[s.ID] = s.Snippet
	}
	block1 := agent.AssembleVersionedAgentSnippets(agents1, nil)

	h1 := baseHash(order1, snippetOf1, "", block1, profiles, agents1)
	h2 := baseHash(order1, snippetOf1, "", block1, profiles, agents1)
	if h1 != h2 {
		t.Error("baseHash should be deterministic")
	}

	// Different agents → different hash
	agents2 := claudeOnlyInstalls(t)
	_, order2 := testOrderedDockerfile(profiles, agents2)
	sources2 := collectSources(profiles, nil)
	snippetOf2 := map[string]string{}
	for _, s := range sources2 {
		snippetOf2[s.ID] = s.Snippet
	}
	block2 := agent.AssembleVersionedAgentSnippets(agents2, nil)
	h3 := baseHash(order2, snippetOf2, "", block2, profiles, agents2)
	if h1 == h3 {
		t.Error("different agents should produce different hash")
	}

	// Different profiles → different hash
	java := []string{"java"}
	javaOnly, _ := kit.Resolve(java, nil)
	_, order3 := testOrderedDockerfile(javaOnly, agents1)
	sources3 := collectSources(javaOnly, nil)
	snippetOf3 := map[string]string{}
	for _, s := range sources3 {
		snippetOf3[s.ID] = s.Snippet
	}
	h4 := baseHash(order3, snippetOf3, "", block1, javaOnly, agents1)
	if h1 == h4 {
		t.Error("different profiles should produce different hash")
	}
}

func TestGenerateProjectDockerfile_WithProfileSnippets(t *testing.T) {
	snippet := "RUN echo 'from-profile'\n"
	df, err := generateProjectDockerfile(snippet, nil, "", "testuser", false)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(df, "from-profile") {
		t.Error("project dockerfile should contain profile snippet")
	}
	if !strings.HasPrefix(df, "FROM asylum:latest") {
		t.Error("should start with FROM asylum:latest")
	}
}

func TestGenerateProjectDockerfile_EmptyReturnsMinimal(t *testing.T) {
	df, err := generateProjectDockerfile("", nil, "", "testuser", false)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(df, "FROM asylum:latest") {
		t.Error("should start with FROM asylum:latest")
	}
	if !strings.HasSuffix(strings.TrimSpace(df), "USER testuser") {
		t.Error("should end with USER testuser")
	}
}

func TestAssembleProjectEntrypoint(t *testing.T) {
	t.Run("kits with entrypoint snippets", func(t *testing.T) {
		kits := []*kit.Kit{
			{Name: "a", EntrypointSnippet: "echo setup-a\n"},
			{Name: "b", EntrypointSnippet: "echo setup-b\n"},
		}
		ep := assembleProjectEntrypoint(kits)
		if ep == nil {
			t.Fatal("expected non-nil project entrypoint")
		}
		s := string(ep)
		if !strings.HasPrefix(s, "#!/bin/bash\nset -e\n") {
			t.Error("should start with shebang and set -e")
		}
		if !strings.Contains(s, "echo setup-a") {
			t.Error("missing snippet from kit a")
		}
		if !strings.Contains(s, "echo setup-b") {
			t.Error("missing snippet from kit b")
		}
	})

	t.Run("no snippets returns nil", func(t *testing.T) {
		kits := []*kit.Kit{
			{Name: "a"},
			{Name: "b"},
		}
		if ep := assembleProjectEntrypoint(kits); ep != nil {
			t.Error("expected nil when no kits have snippets")
		}
	})

	t.Run("nil kits returns nil", func(t *testing.T) {
		if ep := assembleProjectEntrypoint(nil); ep != nil {
			t.Error("expected nil for nil kits")
		}
	})

	t.Run("banner lines exported", func(t *testing.T) {
		kits := []*kit.Kit{
			{Name: "a", BannerLines: "    echo \"a: v1\"\n"},
		}
		ep := assembleProjectEntrypoint(kits)
		if ep == nil {
			t.Fatal("expected non-nil project entrypoint")
		}
		s := string(ep)
		if !strings.Contains(s, "PROJECT_BANNER") {
			t.Error("should export PROJECT_BANNER")
		}
		if !strings.Contains(s, "a: v1") {
			t.Error("PROJECT_BANNER should contain banner line content")
		}
	})

	t.Run("mixed snippets and banner lines", func(t *testing.T) {
		kits := []*kit.Kit{
			{Name: "a", EntrypointSnippet: "echo setup-a\n"},
			{Name: "b", BannerLines: "    echo \"b: v2\"\n"},
		}
		ep := assembleProjectEntrypoint(kits)
		if ep == nil {
			t.Fatal("expected non-nil project entrypoint")
		}
		s := string(ep)
		if !strings.Contains(s, "echo setup-a") {
			t.Error("missing entrypoint snippet")
		}
		if !strings.Contains(s, "PROJECT_BANNER") {
			t.Error("missing PROJECT_BANNER export")
		}
	})
}

func TestGenerateProjectDockerfile_WithProjectEntrypoint(t *testing.T) {
	df, err := generateProjectDockerfile("", nil, "", "testuser", true)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(df, "COPY --chmod=755 project-entrypoint.sh /usr/local/bin/project-entrypoint.sh") {
		t.Error("should contain COPY for project-entrypoint.sh")
	}
}

func TestGenerateProjectDockerfile_WithoutProjectEntrypoint(t *testing.T) {
	df, err := generateProjectDockerfile("", nil, "", "testuser", false)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(df, "project-entrypoint.sh") {
		t.Error("should not reference project-entrypoint.sh when not present")
	}
}

func TestVersionedAgentSnippets(t *testing.T) {
	profiles, err := kit.Resolve(nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	vm := versions.VersionMap{
		"claude":   "0.2.32",
		"gemini":   "0.8.0",
		"codex":    "0.1.0",
		"copilot":  "0.9.0",
		"opencode": "1.0.0",
		"pi":       "0.5.0",
	}

	sources := collectSources(profiles, nil)
	orderedIDs := computeSourceOrder(sources, nil)
	snippetOf := map[string]string{}
	for _, s := range sources {
		snippetOf[s.ID] = s.Snippet
	}
	agentBlock := agent.AssembleVersionedAgentSnippets(allAgentInstalls(t), vm)
	df := assembleDockerfile(orderedIDs, snippetOf, "", agentBlock)
	s := string(df)

	// Check that version ARGs are present in the Dockerfile
	if !strings.Contains(s, "ARG CLAUDE_VERSION=0.2.32") {
		t.Error("missing Claude version ARG")
	}
	if !strings.Contains(s, "CLAUDE_VERSION") {
		t.Error("missing CLAUDE_VERSION reference in snippet")
	}
	if !strings.Contains(s, "ARG GEMINI_VERSION=0.8.0") {
		t.Error("missing Gemini version ARG")
	}
	if !strings.Contains(s, "@${GEMINI_VERSION}") {
		t.Error("missing Gemini version reference")
	}
	if !strings.Contains(s, "ARG COPILOT_VERSION=0.9.0") {
		t.Error("missing Copilot version ARG")
	}
	if !strings.Contains(s, "VERSION=${COPILOT_VERSION}") {
		t.Error("missing Copilot version reference")
	}
	if !strings.Contains(s, "ARG OPENCODE_VERSION=1.0.0") {
		t.Error("missing Opencode version ARG")
	}

	// Verify version-specific RUN commands don't contain "latest"
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		if strings.Contains(line, "gemini-cli") && strings.Contains(line, "latest") {
			t.Errorf("Gemini install should not use 'latest': %s", line)
		}
	}
}

func TestVersionedSnippetCache(t *testing.T) {
	// Passing a nil version map falls back to unversioned agent snippets.
	block := agent.AssembleVersionedAgentSnippets(allAgentInstalls(t), nil)

	if !strings.Contains(block, "claude.ai/install.sh") {
		t.Errorf("expected Claude install snippet, got: %s", block)
	}
	if strings.Contains(block, "ARG") {
		t.Errorf("without version map, agent block should not contain ARG: %s", block)
	}
}
