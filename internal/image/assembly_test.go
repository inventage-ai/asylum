package image

import (
	"strings"
	"testing"

	"github.com/inventage-ai/asylum/assets"
	"github.com/inventage-ai/asylum/internal/profile"
)

func TestAssembleDockerfile_AllProfiles(t *testing.T) {
	profiles, err := profile.Resolve(nil)
	if err != nil {
		t.Fatal(err)
	}
	df := assembleDockerfile(profiles)
	s := string(df)

	if !strings.HasPrefix(s, string(assets.DockerfileCore)) {
		t.Error("assembled Dockerfile should start with core")
	}
	if !strings.HasSuffix(s, string(assets.DockerfileTail)) {
		t.Error("assembled Dockerfile should end with tail")
	}
	// All profile snippets should be present
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
	empty := []string{}
	profiles, err := profile.Resolve(&empty)
	if err != nil {
		t.Fatal(err)
	}
	df := assembleDockerfile(profiles)
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
	if strings.Contains(s, "uv tool install") {
		t.Error("should not contain python snippet")
	}
}

func TestAssembleEntrypoint_AllProfiles(t *testing.T) {
	profiles, err := profile.Resolve(nil)
	if err != nil {
		t.Fatal(err)
	}
	ep := assembleEntrypoint(profiles)
	s := string(ep)

	if !strings.Contains(s, "mise activate bash") {
		t.Error("missing java entrypoint snippet")
	}
	if !strings.Contains(s, "has_python_marker") {
		t.Error("missing python/uv entrypoint snippet")
	}
}

func TestAssembleEntrypoint_NoProfiles(t *testing.T) {
	empty := []string{}
	profiles, err := profile.Resolve(&empty)
	if err != nil {
		t.Fatal(err)
	}
	ep := assembleEntrypoint(profiles)
	s := string(ep)

	if strings.Contains(s, "mise activate bash") {
		t.Error("should not contain java entrypoint snippet")
	}
	if strings.Contains(s, "has_python_marker") {
		t.Error("should not contain python/uv snippet")
	}
	// Should still contain core and tail
	if !strings.Contains(s, "Asylum entrypoint") {
		t.Error("should contain core")
	}
	if !strings.Contains(s, "exec \"$@\"") {
		t.Error("should contain tail")
	}
}

func TestAssembleEntrypoint_BannerLines(t *testing.T) {
	// With all profiles
	profiles, err := profile.Resolve(nil)
	if err != nil {
		t.Fatal(err)
	}
	ep := string(assembleEntrypoint(profiles))
	if !strings.Contains(ep, "Python:") {
		t.Error("banner should contain Python version line")
	}
	if !strings.Contains(ep, "Node.js:") {
		t.Error("banner should contain Node.js version line")
	}
	if !strings.Contains(ep, "Java:") {
		t.Error("banner should contain Java version line")
	}

	// With only java
	java := []string{"java"}
	profiles, err = profile.Resolve(&java)
	if err != nil {
		t.Fatal(err)
	}
	ep = string(assembleEntrypoint(profiles))
	if !strings.Contains(ep, "Java:") {
		t.Error("banner should contain Java version line")
	}
	if strings.Contains(ep, "Python:") {
		t.Error("banner should NOT contain Python when only java is active")
	}
}
