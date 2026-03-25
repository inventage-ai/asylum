package image

import (
	"strings"
	"testing"

	"github.com/inventage-ai/asylum/assets"
	"github.com/inventage-ai/asylum/internal/agent"
	"github.com/inventage-ai/asylum/internal/kit"
)

func allAgentInstalls(t *testing.T) []*agent.AgentInstall {
	t.Helper()
	all := map[string]bool{"claude": true, "codex": true, "gemini": true, "opencode": true}
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

func TestAssembleDockerfile_AllProfiles(t *testing.T) {
	profiles, err := kit.Resolve(nil)
	if err != nil {
		t.Fatal(err)
	}
	df := assembleDockerfile(profiles, allAgentInstalls(t))
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
	profiles, err := kit.Resolve(noKits)
	if err != nil {
		t.Fatal(err)
	}
	df := assembleDockerfile(profiles, claudeOnlyInstalls(t))
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
	profiles, _ := kit.Resolve(nil)

	t.Run("all agents", func(t *testing.T) {
		df := string(assembleDockerfile(profiles, allAgentInstalls(t)))
		if !strings.Contains(df, "claude.ai/install.sh") {
			t.Error("missing claude install snippet")
		}
		if !strings.Contains(df, "gemini-cli") {
			t.Error("missing gemini install snippet")
		}
		if !strings.Contains(df, "@openai/codex") {
			t.Error("missing codex install snippet")
		}
		if !strings.Contains(df, "opencode") {
			t.Error("missing opencode install snippet")
		}
	})

	t.Run("claude only (default)", func(t *testing.T) {
		df := string(assembleDockerfile(profiles, claudeOnlyInstalls(t)))
		if !strings.Contains(df, "claude.ai/install.sh") {
			t.Error("missing claude install snippet")
		}
		if strings.Contains(df, "gemini-cli") {
			t.Error("should not contain gemini snippet")
		}
		if strings.Contains(df, "@openai/codex") {
			t.Error("should not contain codex snippet")
		}
	})

	t.Run("no agents", func(t *testing.T) {
		empty := map[string]bool{}
		noAgents, _ := agent.ResolveInstalls(empty, nil)
		df := string(assembleDockerfile(profiles, noAgents))
		if strings.Contains(df, "claude.ai/install.sh") {
			t.Error("should not contain claude snippet")
		}
	})
}

func TestAssembleEntrypoint_AllProfiles(t *testing.T) {
	profiles, err := kit.Resolve(nil)
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
	profiles, err := kit.Resolve(noKits)
	if err != nil {
		t.Fatal(err)
	}
	ep := assembleEntrypoint(profiles, nil)
	s := string(ep)

	if strings.Contains(s, "mise activate bash") {
		t.Error("should not contain java entrypoint snippet")
	}
	if !strings.Contains(s, "Asylum entrypoint") {
		t.Error("should contain core")
	}
	if !strings.Contains(s, "exec \"$@\"") {
		t.Error("should contain tail")
	}
}

func TestAssembleEntrypoint_BannerLines(t *testing.T) {
	profiles, _ := kit.Resolve(nil)

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

func TestBaseHash_DeterministicAndChanges(t *testing.T) {
	profiles, _ := kit.Resolve(nil)
	agents1 := allAgentInstalls(t)

	h1 := baseHash(profiles, agents1)
	h2 := baseHash(profiles, agents1)
	if h1 != h2 {
		t.Error("baseHash should be deterministic")
	}

	// Different agents → different hash
	h3 := baseHash(profiles, claudeOnlyInstalls(t))
	if h1 == h3 {
		t.Error("different agents should produce different hash")
	}

	// Different profiles → different hash
	java := []string{"java"}
	javaOnly, _ := kit.Resolve(java)
	h4 := baseHash(javaOnly, agents1)
	if h1 == h4 {
		t.Error("different profiles should produce different hash")
	}
}

func TestGenerateProjectDockerfile_WithProfileSnippets(t *testing.T) {
	snippet := "RUN echo 'from-profile'\n"
	df, err := generateProjectDockerfile(snippet, nil, "")
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
	df, err := generateProjectDockerfile("", nil, "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(df, "FROM asylum:latest") {
		t.Error("should start with FROM asylum:latest")
	}
	if !strings.HasSuffix(strings.TrimSpace(df), "USER claude") {
		t.Error("should end with USER claude")
	}
}
