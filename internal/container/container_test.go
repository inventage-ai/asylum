package container

import (
	"strings"
	"testing"
)

func TestSafeHostname(t *testing.T) {
	tests := []struct {
		name string
		dir  string
		want string
	}{
		{
			name: "simple name",
			dir:  "/home/user/myproject",
			want: "asylum-myproject",
		},
		{
			name: "underscores become dashes",
			dir:  "/home/user/my_project",
			want: "asylum-my-project",
		},
		{
			name: "uppercase lowercased",
			dir:  "/home/user/MyProject",
			want: "asylum-myproject",
		},
		{
			name: "leading dash stripped",
			dir:  "/home/user/_project",
			want: "asylum-project",
		},
		{
			name: "trailing dash stripped after truncation",
			// base name: 56 a's + hyphen + more: truncation at 56 lands on hyphen
			dir:  "/home/user/" + strings.Repeat("a", 55) + "-extra",
			want: "asylum-" + strings.Repeat("a", 55),
		},
		{
			name: "exact 56-char input not truncated",
			dir:  "/home/user/" + strings.Repeat("a", 56),
			want: "asylum-" + strings.Repeat("a", 56),
		},
		{
			name: "all non-alphanumeric becomes dashes then empty -> project",
			dir:  "/home/user/___",
			want: "asylum-project",
		},
		{
			name: "empty base falls back to project",
			dir:  "/",
			want: "asylum-project",
		},
		{
			name: "result within Docker 63-char limit",
			dir:  "/home/user/" + strings.Repeat("b", 63),
			want: "asylum-" + strings.Repeat("b", 56),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := safeHostname(tt.dir)
			if got != tt.want {
				t.Errorf("safeHostname(%q) = %q, want %q", tt.dir, got, tt.want)
			}
			if len(got) > 63 {
				t.Errorf("hostname too long: %d chars", len(got))
			}
			if strings.HasSuffix(got, "-") {
				t.Errorf("hostname has trailing dash: %q", got)
			}
		})
	}
}
