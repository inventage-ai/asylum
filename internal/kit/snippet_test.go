package kit

import "testing"

func TestSnippetIsActive(t *testing.T) {
	tests := []struct {
		name    string
		snippet string
		want    bool
	}{
		{"active entry", "  node:\n    versions:\n", true},
		{"commented entry", "  # rtk:              # Token-reduction proxy\n", false},
		{"blank lines before content", "\n\n  java:\n", true},
		{"empty", "", false},
		{"active entry with inner hint comment", "  node:\n    # versions:\n", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SnippetIsActive(tt.snippet); got != tt.want {
				t.Errorf("SnippetIsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUncommentSnippetRoundTrip(t *testing.T) {
	active := "  node:\n    default: 22\n"
	commented := CommentSnippet(active)
	if SnippetIsActive(commented) {
		t.Fatalf("CommentSnippet left the snippet active: %q", commented)
	}
	if got := UncommentSnippet(commented); got != active {
		t.Errorf("round trip changed the snippet:\n got %q\nwant %q", got, active)
	}
}

// Commenting is not lossless when the active snippet already contains inner
// comment lines: CommentSnippet skips them so the result has no `# # ` pairs,
// and uncommenting then strips them for real. Pinned because it looks like a
// bug at a glance and the alternative produces worse config files.
func TestCommentSnippetLeavesInnerCommentsAlone(t *testing.T) {
	active := "  node:\n    # versions: 20, 22\n    default: 22\n"
	commented := CommentSnippet(active)
	if want := "  # node:\n    # versions: 20, 22\n    # default: 22\n"; commented != want {
		t.Fatalf("CommentSnippet() = %q, want %q", commented, want)
	}
	if got := UncommentSnippet(commented); got == active {
		t.Error("expected the inner hint comment to be consumed by the round trip")
	}
}

// Every opt-in kit authors its snippet commented out, and both the first-run
// writer and the kit-sync prompt uncomment it when the user opts in. A snippet
// that stays commented after that leaves the kit impossible to enable from the
// prompt, which is silent and looks like the prompt doing nothing.
func TestOptInSnippetsCanBeActivated(t *testing.T) {
	for _, name := range All() {
		k := Get(name)
		if k == nil || k.Tier != TierOptIn || k.ConfigSnippet == "" {
			continue
		}
		if !SnippetIsActive(UncommentSnippet(k.ConfigSnippet)) {
			t.Errorf("kit %q cannot be activated: uncommenting its snippet still yields %q",
				name, UncommentSnippet(k.ConfigSnippet))
		}
	}
}
