package image

import (
	"strings"
	"testing"
)

func TestGenerateProjectDockerfile(t *testing.T) {
	t.Run("unknown key rejected", func(t *testing.T) {
		_, err := generateProjectDockerfile("", map[string][]string{"bad": {"x"}}, "", "testuser", false)
		if err == nil {
			t.Error("expected error for unknown package type")
		}
	})

	t.Run("apt packages", func(t *testing.T) {
		df, err := generateProjectDockerfile("", map[string][]string{
			"apt": {"curl", "jq"},
		}, "", "testuser", false)
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
		if !strings.HasSuffix(strings.TrimSpace(df), "USER testuser") {
			t.Error("should end with USER testuser")
		}
	})

	t.Run("npm packages", func(t *testing.T) {
		df, err := generateProjectDockerfile("", map[string][]string{
			"npm": {"typescript", "eslint"},
		}, "", "testuser", false)
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
		}, "", "testuser", false)
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
		}, "", "testuser", false)
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
		}, "", "testuser", false)
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

	t.Run("always ends with USER testuser", func(t *testing.T) {
		df, err := generateProjectDockerfile("", map[string][]string{
			"apt": {"curl"},
		}, "", "testuser", false)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasSuffix(strings.TrimSpace(df), "USER testuser") {
			t.Errorf("dockerfile does not end with USER testuser:\n%s", df)
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
			_, err := generateProjectDockerfile("", map[string][]string{"apt": {name}}, "", "testuser", false)
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
			_, err := generateProjectDockerfile("", map[string][]string{"npm": {name}}, "", "testuser", false)
			if err == nil {
				t.Errorf("expected error for npm package name %q", name)
			}
		}
	})

	t.Run("scoped npm package accepted", func(t *testing.T) {
		df, err := generateProjectDockerfile("", map[string][]string{"npm": {"@mermaid-js/mermaid-cli"}}, "", "testuser", false)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(df, "@mermaid-js/mermaid-cli") {
			t.Error("missing scoped npm package")
		}
	})

	t.Run("leading dash package rejected", func(t *testing.T) {
		_, err := generateProjectDockerfile("", map[string][]string{"npm": {"--registry=https://evil.example"}}, "", "testuser", false)
		if err == nil {
			t.Error("expected error for leading-dash package name")
		}
	})

	t.Run("kitProjectSnippets inserted", func(t *testing.T) {
		df, err := generateProjectDockerfile("", map[string][]string{}, "RUN echo from-kit\n", "testuser", false)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(df, "echo from-kit") {
			t.Error("missing kit project snippet")
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
			_, err := generateProjectDockerfile("", map[string][]string{"pip": {name}}, "", "testuser", false)
			if err == nil {
				t.Errorf("expected error for pip package name %q", name)
			}
		}
	})

	t.Run("different packages produce different hashes", func(t *testing.T) {
		df1, err := generateProjectDockerfile("", map[string][]string{
			"npm": {"typescript-language-server"},
		}, "", "testuser", false)
		if err != nil {
			t.Fatal(err)
		}
		df2, err := generateProjectDockerfile("", map[string][]string{
			"npm": {"eslint"},
		}, "", "testuser", false)
		if err != nil {
			t.Fatal(err)
		}
		if df1 == df2 {
			t.Error("different packages should produce different Dockerfiles")
		}
	})
}

func TestBasePackageBlock(t *testing.T) {
	t.Run("empty returns nothing", func(t *testing.T) {
		block, err := basePackageBlock(nil)
		if err != nil {
			t.Fatal(err)
		}
		if block != "" {
			t.Errorf("expected empty block, got %q", block)
		}
	})

	t.Run("apt as root, npm as build user, restores user", func(t *testing.T) {
		block, err := basePackageBlock(map[string][]string{
			"apt": {"jq"},
			"npm": {"@mermaid-js/mermaid-cli"},
		})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(block, "USER root") {
			t.Error("apt should run as USER root")
		}
		if !strings.Contains(block, "npm install -g") || !strings.Contains(block, "@mermaid-js/mermaid-cli") {
			t.Error("missing npm install")
		}
		// The build-arg user is used in the base image, and the block must
		// restore it so the agent block that follows runs unprivileged.
		if !strings.Contains(block, "USER ${USERNAME}") {
			t.Error("block should reference the ${USERNAME} build arg")
		}
		if !strings.HasSuffix(strings.TrimSpace(block), "USER ${USERNAME}") {
			t.Error("block must end by restoring USER ${USERNAME}")
		}
	})

	t.Run("invalid name rejected", func(t *testing.T) {
		if _, err := basePackageBlock(map[string][]string{"npm": {"--registry=https://evil"}}); err == nil {
			t.Error("expected error for flag-like package name")
		}
	})
}

// When only global packages are configured, the project-tier map is empty and
// EnsureProject returns the base tag without building a project image.
func TestEnsureProject_EmptyProjectPackagesReturnsBase(t *testing.T) {
	tag, err := EnsureProject(nil, nil, map[string][]string{}, nil, "test", false, false)
	if err != nil {
		t.Fatal(err)
	}
	if tag != baseTag {
		t.Errorf("expected %q, got %q", baseTag, tag)
	}
}
