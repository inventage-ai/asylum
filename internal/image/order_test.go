package image

import (
	"slices"
	"strings"
	"testing"

	"github.com/inventage-ai/asylum/assets"
	"github.com/inventage-ai/asylum/internal/agent"
	"github.com/inventage-ai/asylum/internal/kit"
	"github.com/inventage-ai/asylum/internal/versions"
)

func TestComputeSourceOrder(t *testing.T) {
	tests := []struct {
		name     string
		sources  []dockerSource
		previous []string
		want     []string
	}{
		{
			name: "no previous order — sort by priority",
			sources: []dockerSource{
				{ID: "kit:docker", Priority: 30},
				{ID: "kit:node", Priority: 20},
				{ID: "kit:java", Priority: 10},
			},
			previous: nil,
			want:     []string{"kit:java", "kit:node", "kit:docker"},
		},
		{
			name: "no changes — preserve previous order",
			sources: []dockerSource{
				{ID: "kit:java", Priority: 10},
				{ID: "kit:docker", Priority: 30},
				{ID: "kit:node", Priority: 20},
			},
			previous: []string{"kit:java", "kit:node", "kit:docker"},
			want:     []string{"kit:java", "kit:node", "kit:docker"},
		},
		{
			name: "single new source appended",
			sources: []dockerSource{
				{ID: "kit:java", Priority: 10},
				{ID: "kit:node", Priority: 20},
				{ID: "kit:docker", Priority: 30},
			},
			previous: []string{"kit:java", "kit:node"},
			want:     []string{"kit:java", "kit:node", "kit:docker"},
		},
		{
			name: "multiple new sources sorted by priority",
			sources: []dockerSource{
				{ID: "kit:java", Priority: 10},
				{ID: "kit:github", Priority: 40},
				{ID: "kit:node", Priority: 20},
			},
			previous: []string{"kit:java"},
			want:     []string{"kit:java", "kit:node", "kit:github"},
		},
		{
			name: "source removed from middle — re-sort suffix",
			sources: []dockerSource{
				{ID: "A", Priority: 10},
				{ID: "C", Priority: 20},
				{ID: "D", Priority: 40},
			},
			previous: []string{"A", "B", "C", "D"},
			want:     []string{"A", "C", "D"},
		},
		{
			name: "source removed from beginning — re-sort all",
			sources: []dockerSource{
				{ID: "B", Priority: 20},
				{ID: "C", Priority: 30},
			},
			previous: []string{"A", "B", "C"},
			want:     []string{"B", "C"},
		},
		{
			name: "multiple sources removed",
			sources: []dockerSource{
				{ID: "A", Priority: 10},
				{ID: "C", Priority: 20},
				{ID: "E", Priority: 15},
			},
			// B(index 1) and D(index 3) removed → earliest removal at 1
			// prefix = [A], suffix = [C, E] re-sorted → [E(15), C(20)]
			previous: []string{"A", "B", "C", "D", "E"},
			want:     []string{"A", "E", "C"},
		},
		{
			name: "removal and addition combined",
			sources: []dockerSource{
				{ID: "A", Priority: 10},
				{ID: "C", Priority: 20},
				{ID: "D", Priority: 25},
			},
			// B removed at index 1 → prefix = [A], suffix = [C] re-sorted = [C]
			// D is new → appended
			previous: []string{"A", "B", "C"},
			want:     []string{"A", "C", "D"},
		},
		{
			name: "tie-breaking by identifier",
			sources: []dockerSource{
				{ID: "kit:foo", Priority: 50},
				{ID: "kit:bar", Priority: 50},
			},
			previous: nil,
			want:     []string{"kit:bar", "kit:foo"},
		},
		{
			name: "previous order with unknown identifiers — ignored",
			sources: []dockerSource{
				{ID: "A", Priority: 10},
				{ID: "B", Priority: 20},
			},
			previous: []string{"A", "unknown", "B"},
			want:     []string{"A", "B"},
		},
		{
			name:     "empty sources",
			sources:  nil,
			previous: []string{"A", "B"},
			want:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeSourceOrder(tt.sources, tt.previous)
			if !slices.Equal(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOrderingAgentsAfterKits(t *testing.T) {
	profiles, err := kit.Resolve(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	agents := allAgentInstalls(t)
	df, _ := testOrderedDockerfile(profiles, agents)
	s := string(df)

	// Every kit snippet must appear before every agent snippet.
	lastKit := -1
	for _, src := range collectSources(profiles, nil) {
		idx := strings.Index(s, strings.TrimRight(src.Snippet, "\n"))
		if idx < 0 {
			t.Fatalf("kit snippet %s not found in Dockerfile", src.ID)
		}
		if idx > lastKit {
			lastKit = idx
		}
	}

	firstAgent := len(s)
	for _, a := range agents {
		idx := strings.Index(s, strings.TrimRight(a.DockerSnippet, "\n"))
		if idx < 0 {
			t.Fatalf("agent snippet %q not found in Dockerfile", a.Name)
		}
		if idx < firstAgent {
			firstAgent = idx
		}
	}

	if lastKit >= firstAgent {
		t.Errorf("agents should come after all kits: lastKit=%d firstAgent=%d", lastKit, firstAgent)
	}

	// The agent block must sit before the tail.
	tailIdx := strings.Index(s, "init.defaultBranch")
	if tailIdx < 0 {
		t.Fatal("tail marker not found")
	}
	if firstAgent >= tailIdx {
		t.Errorf("agent block (idx %d) should appear before the tail (idx %d)", firstAgent, tailIdx)
	}
}

func TestAgentBlockClaudeFirst(t *testing.T) {
	block := agent.AssembleVersionedAgentSnippets(allAgentInstalls(t), nil)
	claudeIdx := strings.Index(block, "claude.ai/install.sh")
	geminiIdx := strings.Index(block, "gemini-cli")
	if claudeIdx < 0 || geminiIdx < 0 {
		t.Fatal("expected claude and gemini snippets in agent block")
	}
	if claudeIdx > geminiIdx {
		t.Error("claude should appear first in the deterministic agent block")
	}
}

func TestAgentVersionBumpPreservesKitLayers(t *testing.T) {
	profiles, _ := kit.Resolve(nil, nil)
	agents := allAgentInstalls(t)
	sources := collectSources(profiles, nil)
	orderedIDs := computeSourceOrder(sources, nil)
	snippetOf := map[string]string{}
	for _, s := range sources {
		snippetOf[s.ID] = s.Snippet
	}

	// The kit prefix is everything up to (but not including) the tail when no
	// agents are present. It must be a byte-identical prefix of the assembled
	// Dockerfile regardless of the pinned agent versions.
	kitPrefix := strings.TrimSuffix(string(assembleDockerfile(orderedIDs, snippetOf, "", "")), string(assets.DockerfileTail))

	df1 := string(assembleDockerfile(orderedIDs, snippetOf, "", agent.AssembleVersionedAgentSnippets(agents, versions.VersionMap{"claude": "1.0.0"})))
	df2 := string(assembleDockerfile(orderedIDs, snippetOf, "", agent.AssembleVersionedAgentSnippets(agents, versions.VersionMap{"claude": "2.0.0"})))

	if !strings.HasPrefix(df1, kitPrefix) || !strings.HasPrefix(df2, kitPrefix) {
		t.Error("kit prefix (core + kit layers) changed when only an agent version bumped")
	}
}

func TestOrderingNewKitAppendedLast(t *testing.T) {
	profiles, _ := kit.Resolve(nil, nil)
	sources := collectSources(profiles, nil)

	order1 := computeSourceOrder(sources, nil)

	extraSources := append(slices.Clone(sources), dockerSource{
		ID: "kit:newkit", Priority: 1, Snippet: "RUN echo newkit\n",
	})
	order2 := computeSourceOrder(extraSources, order1)

	if order2[len(order2)-1] != "kit:newkit" {
		t.Errorf("new kit should be last, got order: %v", order2)
	}
	for i := 0; i < len(order1); i++ {
		if order2[i] != order1[i] {
			t.Errorf("position %d changed: was %s, now %s", i, order1[i], order2[i])
		}
	}
}

func TestOrderingStateRoundTrip(t *testing.T) {
	profiles, _ := kit.Resolve(nil, nil)
	sources := collectSources(profiles, nil)

	order1 := computeSourceOrder(sources, nil)
	order2 := computeSourceOrder(sources, order1)
	if !slices.Equal(order1, order2) {
		t.Errorf("order should be stable across builds\nfirst:  %v\nsecond: %v", order1, order2)
	}
}

func TestDockerfileSnippetOrder(t *testing.T) {
	profiles, _ := kit.Resolve(nil, nil)
	agents := claudeOnlyInstalls(t)
	df, _ := testOrderedDockerfile(profiles, agents)
	s := string(df)

	if !strings.HasPrefix(s, string(assets.DockerfileCore)) {
		t.Error("Dockerfile should start with core")
	}
	if !strings.HasSuffix(s, string(assets.DockerfileTail)) {
		t.Error("Dockerfile should end with tail")
	}

	claudeIdx := strings.Index(s, "claude.ai/install.sh")
	javaIdx := strings.Index(s, "mise install java")
	if claudeIdx < 0 || javaIdx < 0 {
		t.Fatal("expected both claude and java snippets in Dockerfile")
	}
	if javaIdx > claudeIdx {
		t.Error("java (kit) snippet should appear before claude (agent) snippet in Dockerfile")
	}
}
