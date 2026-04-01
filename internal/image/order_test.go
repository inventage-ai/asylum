package image

import (
	"slices"
	"strings"
	"testing"

	"github.com/inventage-ai/asylum/assets"
	"github.com/inventage-ai/asylum/internal/kit"
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
				{ID: "agent:claude", Priority: 20},
				{ID: "kit:java", Priority: 10},
			},
			previous: nil,
			want:     []string{"kit:java", "agent:claude", "kit:docker"},
		},
		{
			name: "no changes — preserve previous order",
			sources: []dockerSource{
				{ID: "kit:java", Priority: 10},
				{ID: "kit:docker", Priority: 30},
				{ID: "agent:claude", Priority: 20},
			},
			previous: []string{"kit:java", "agent:claude", "kit:docker"},
			want:     []string{"kit:java", "agent:claude", "kit:docker"},
		},
		{
			name: "single new source appended",
			sources: []dockerSource{
				{ID: "kit:java", Priority: 10},
				{ID: "agent:claude", Priority: 20},
				{ID: "kit:docker", Priority: 30},
			},
			previous: []string{"kit:java", "agent:claude"},
			want:     []string{"kit:java", "agent:claude", "kit:docker"},
		},
		{
			name: "multiple new sources sorted by priority",
			sources: []dockerSource{
				{ID: "kit:java", Priority: 10},
				{ID: "kit:github", Priority: 40},
				{ID: "agent:claude", Priority: 20},
			},
			previous: []string{"kit:java"},
			want:     []string{"kit:java", "agent:claude", "kit:github"},
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
			name: "empty sources",
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

func TestOrderingAgentsBeforeKits(t *testing.T) {
	profiles, err := kit.Resolve(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	agents := allAgentInstalls(t)
	sources := collectSources(profiles, agents)
	orderedIDs := computeSourceOrder(sources, nil)

	lastAgent := -1
	firstKit := -1
	for i, id := range orderedIDs {
		if strings.HasPrefix(id, "agent:") {
			lastAgent = i
		}
		if strings.HasPrefix(id, "kit:") && firstKit == -1 {
			firstKit = i
		}
	}
	if lastAgent >= 0 && firstKit >= 0 && lastAgent >= firstKit {
		t.Errorf("agents should come before kits, got last agent at %d, first kit at %d\norder: %v", lastAgent, firstKit, orderedIDs)
	}
}

func TestOrderingClaudeBeforeOtherAgents(t *testing.T) {
	agents := allAgentInstalls(t)
	profiles, _ := kit.Resolve(nil, nil)
	sources := collectSources(profiles, agents)
	orderedIDs := computeSourceOrder(sources, nil)

	claudeIdx := -1
	for i, id := range orderedIDs {
		if id == "agent:claude" {
			claudeIdx = i
			break
		}
	}
	if claudeIdx < 0 {
		t.Fatal("agent:claude not found in order")
	}
	for i, id := range orderedIDs {
		if strings.HasPrefix(id, "agent:") && id != "agent:claude" && i < claudeIdx {
			t.Errorf("agent:claude (idx %d) should come before %s (idx %d)", claudeIdx, id, i)
		}
	}
}

func TestOrderingNewKitAppendedLast(t *testing.T) {
	profiles, _ := kit.Resolve(nil, nil)
	agents := claudeOnlyInstalls(t)
	sources := collectSources(profiles, agents)

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
	agents := claudeOnlyInstalls(t)
	sources := collectSources(profiles, agents)

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
	if claudeIdx > javaIdx {
		t.Error("claude snippet should appear before java snippet in Dockerfile")
	}
}
