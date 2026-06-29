package versions

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// roundTripperFunc is a helper type for mocking HTTP in tests.
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func withMockServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(func() { server.Close() })
	return server
}

func replaceHTTPClient(server *httptest.Server) func() {
	old := httpClient
	serverURL, _ := url.Parse(server.URL)
	httpClient = &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		req.URL.Scheme = serverURL.Scheme
		req.URL.Host = serverURL.Host
		return http.DefaultTransport.RoundTrip(req)
	})}
	return func() { httpClient = old }
}

func TestFetchNpmVersion(t *testing.T) {
	tests := []struct {
		version string
		want    string
	}{
		{"1.2.3", "1.2.3"},
		{"0.8.0-beta.1", "0.8.0-beta.1"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			server := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
				json.NewEncoder(w).Encode(NpmVersion{Version: tt.version})
			})
			defer replaceHTTPClient(server)()

			got, err := fetchNpmVersion("@test/pkg")
			if err != nil {
				t.Fatalf("fetchNpmVersion() returned error: %v", err)
			}
			if got != tt.want {
				t.Errorf("fetchNpmVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFetchNpmVersionError(t *testing.T) {
	server := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	})
	defer replaceHTTPClient(server)()

	_, err := fetchNpmVersion("@test/pkg")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFetchGitHubRelease(t *testing.T) {
	tests := []struct {
		tagName string
		want    string
	}{
		{"v1.0.65", "1.0.65"},
		{"1.0.65", "1.0.65"},
	}
	for _, tt := range tests {
		t.Run(tt.tagName, func(t *testing.T) {
			server := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
				json.NewEncoder(w).Encode(struct{ TagName string `json:"tag_name"` }{TagName: tt.tagName})
			})
			defer replaceHTTPClient(server)()

			got, err := fetchGitHubRelease("test/repo")
			if err != nil {
				t.Fatalf("fetchGitHubRelease() returned error: %v", err)
			}
			if got != tt.want {
				t.Errorf("fetchGitHubRelease() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFetchGitHubTags(t *testing.T) {
	tests := []struct {
		name string
		tags []string
		want string
	}{
		{
			"stable tag first",
			[]string{"v2.1.195", "v2.1.194", "v2.1.193"},
			"2.1.195",
		},
		{
			"skip pre-release tags",
			[]string{"v2.1.0-rc1", "v2.1.0-beta2", "v2.1.0", "v2.0.9"},
			"2.1.0",
		},
		{
			"only pre-release tags available",
			[]string{"v2.1.0-rc2", "v2.1.0-rc1", "v2.1.0-beta1"},
			"",
		},
		{
			"version with +build suffix",
			[]string{"v2.1.0+build123", "v2.0.9"},
			"2.0.9",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tagList := make([]struct{ Name string `json:"name"` }, len(tt.tags))
			for i, tag := range tt.tags {
				tagList[i].Name = tag
			}

			server := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
				json.NewEncoder(w).Encode(tagList)
			})
			defer replaceHTTPClient(server)()

			got, err := fetchGitHubTags("test/repo")
			if tt.want == "" {
				if err == nil {
					t.Fatal("expected error for no stable tag, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("fetchGitHubTags() returned error: %v", err)
			}
			if got != tt.want {
				t.Errorf("fetchGitHubTags() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAgentNames(t *testing.T) {
	names := AgentNames()
	expected := []string{"claude", "codex", "copilot", "gemini", "opencode", "pi"}
	if len(names) != len(expected) {
		t.Fatalf("AgentNames() = %d, want %d", len(names), len(expected))
	}
	// Sort both for comparison since map iteration is not guaranteed
	sorted := make([]string, len(names))
	copy(sorted, names)
	// Simple insertion sort for small arrays
	for i := 1; i < len(sorted); i++ {
		for j := i; j > 0 && sorted[j] < sorted[j-1]; j-- {
			sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
		}
	}
	for i, name := range expected {
		if sorted[i] != name {
			t.Errorf("AgentNames()[%d] = %q, want %q", i, sorted[i], name)
		}
	}
}

func TestFetchForAgent(t *testing.T) {
	tests := []struct {
		name    string
		agents  []string
		handler http.HandlerFunc
		want    map[string]string
	}{
		{
			"npm agent",
			[]string{"gemini"},
			func(w http.ResponseWriter, r *http.Request) {
				json.NewEncoder(w).Encode(NpmVersion{Version: "0.8.0"})
			},
			map[string]string{"gemini": "0.8.0"},
		},
		{
			"github agent",
			[]string{"copilot"},
			func(w http.ResponseWriter, r *http.Request) {
				json.NewEncoder(w).Encode(struct{ TagName string `json:"tag_name"` }{TagName: "v1.0.65"})
			},
			map[string]string{"copilot": "1.0.65"},
		},
		{
			"mixed agents",
			[]string{"gemini", "copilot"},
			func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/@google/gemini-cli/latest" {
					json.NewEncoder(w).Encode(NpmVersion{Version: "0.8.0"})
				} else {
					json.NewEncoder(w).Encode(struct{ TagName string `json:"tag_name"` }{TagName: "v1.0.65"})
				}
			},
			map[string]string{"gemini": "0.8.0", "copilot": "1.0.65"},
		},
		{
			"unknown agent skipped",
			[]string{"gemini", "unknown-agent"},
			func(w http.ResponseWriter, r *http.Request) {
				json.NewEncoder(w).Encode(NpmVersion{Version: "0.8.0"})
			},
			map[string]string{"gemini": "0.8.0"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := withMockServer(t, tt.handler)
			defer replaceHTTPClient(server)()

			vm := FetchForAgent(tt.agents)
			for k, v := range tt.want {
				if vm[k] != v {
					t.Errorf("FetchForAgent() [%q] = %q, want %q", k, vm[k], v)
				}
			}
		})
	}
}
