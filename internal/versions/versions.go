package versions

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var httpClient = &http.Client{Timeout: 60 * time.Second}

// VersionMap holds the latest known version for each agent.
// Keys are agent names (e.g. "claude", "gemini"), values are version strings
// as returned by their respective sources (e.g. "v2.1.195", "0.8.0").
type VersionMap map[string]string

// NpmVersion is the structure returned by the npm registry /latest endpoint.
type NpmVersion struct {
	Version string `json:"version"`
}

// fetchNpmVersion queries the npm registry for the latest version of a package.
func fetchNpmVersion(packageName string) (string, error) {
	url := fmt.Sprintf("https://registry.npmjs.org/%s/latest", packageName)
	resp, err := httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("fetch npm version for %s: %w", packageName, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch npm version for %s: HTTP %d", packageName, resp.StatusCode)
	}

	var result NpmVersion
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("parse npm version for %s: %w", packageName, err)
	}
	if result.Version == "" {
		return "", fmt.Errorf("no version returned for %s", packageName)
	}
	return result.Version, nil
}

// fetchGitHubRelease queries the GitHub API for the latest release tag.
func fetchGitHubRelease(repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	resp, err := httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("fetch github release for %s: %w", repo, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch github release for %s: HTTP %d", repo, resp.StatusCode)
	}

	var result struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("parse github release for %s: %w", repo, err)
	}
	if result.TagName == "" {
		return "", fmt.Errorf("no tag name returned for %s", repo)
	}
	return strings.TrimPrefix(result.TagName, "v"), nil
}

// fetchGitHubTags queries the GitHub API for the first non-pre-release tag.
func fetchGitHubTags(repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/tags?per_page=30", repo)
	resp, err := httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("fetch github tags for %s: %w", repo, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch github tags for %s: HTTP %d", repo, resp.StatusCode)
	}

	var tags []struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return "", fmt.Errorf("parse github tags for %s: %w", repo, err)
	}

	// Find the first non-pre-release tag (pre-releases contain "-", "+", "~")
	for _, t := range tags {
		name := strings.TrimPrefix(t.Name, "v")
		if strings.ContainsAny(name, "-+~") {
			continue
		}
		return name, nil
	}

	return "", fmt.Errorf("no non-pre-release tag found for %s", repo)
}

// Fetchers maps agent names to their version fetch function.
var fetchers = map[string]func() (string, error){
	// Claude: GitHub tags (anthropics/claude-code)
	"claude": func() (string, error) { return fetchGitHubTags("anthropics/claude-code") },

	// Gemini: npm registry (@google/gemini-cli)
	"gemini": func() (string, error) { return fetchNpmVersion("@google/gemini-cli") },

	// Codex: npm registry (@openai/codex)
	"codex": func() (string, error) { return fetchNpmVersion("@openai/codex") },

	// Copilot: GitHub releases (github/copilot-cli)
	"copilot": func() (string, error) { return fetchGitHubRelease("github/copilot-cli") },

	// Opencode: GitHub releases (opencode-ai/opencode)
	"opencode": func() (string, error) { return fetchGitHubRelease("opencode-ai/opencode") },

	// Pi: npm registry (@earendil-works/pi-coding-agent)
	"pi": func() (string, error) { return fetchNpmVersion("@earendil-works/pi-coding-agent") },
}

// AgentNames returns the sorted list of all agents tracked by this package.
func AgentNames() []string {
	names := make([]string, 0, len(fetchers))
	for name := range fetchers {
		names = append(names, name)
	}
	return names
}

// AgentSource describes where an agent's version is fetched from.
type AgentSource int

const (
	SourceNpmAgent AgentSource = iota
	SourceGitHubRelease
	SourceGitHubTag
)

// AgentSourceMap maps agent names to their fetch source type.
var AgentSourceMap = map[string]AgentSource{
	"claude":   SourceGitHubTag,
	"gemini":   SourceNpmAgent,
	"codex":    SourceNpmAgent,
	"copilot":  SourceGitHubRelease,
	"opencode": SourceGitHubRelease,
	"pi":       SourceNpmAgent,
}

// FetchAll fetches the latest version for all registered agents.
// Agents that fail to fetch are simply omitted from the result.
func FetchAll() VersionMap {
	vm := make(VersionMap)
	for name, fetcher := range fetchers {
		version, err := fetcher()
		if err != nil {
			// Silently skip agents whose fetch fails.
			continue
		}
		vm[name] = version
	}
	return vm
}

// FetchForAgent fetches versions only for agents that are in the active install list.
// The installs slice comes from agent.ResolveInstalls and includes the agent name.
func FetchForAgent(installNames []string) VersionMap {
	agentSet := make(map[string]bool, len(installNames))
	for _, name := range installNames {
		agentSet[name] = true
	}

	vm := make(VersionMap)
	for _, name := range installNames {
		fetcher, ok := fetchers[name]
		if !ok {
			continue
		}
		version, err := fetcher()
		if err != nil {
			continue
		}
		vm[name] = version
	}
	return vm
}
