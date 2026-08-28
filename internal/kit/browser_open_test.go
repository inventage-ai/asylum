package kit

import (
	"slices"
	"strings"
	"testing"
)

func TestDetectLoopbackCallback(t *testing.T) {
	cases := []struct {
		name     string
		url      string
		wantPort int
		wantV6   bool
		wantOK   bool
	}{
		{"ipv4 redirect_uri", "https://p/auth?client_id=x&redirect_uri=http://127.0.0.1:54321/cb", 54321, false, true},
		{"localhost", "https://p/auth?redirect_uri=http://localhost:8080", 8080, false, true},
		{"ipv6 loopback", "https://p/auth?redirect_uri=http://[::1]:9000/cb", 9000, true, true},
		{"redirect_url alias", "https://p/auth?redirect_url=http://127.0.0.1:7000", 7000, false, true},
		{"percent-encoded", "https://p/auth?redirect_uri=http%3A%2F%2F127.0.0.1%3A5000%2Fcb", 5000, false, true},
		{"non-loopback host", "https://p/auth?redirect_uri=https://example.com/cb", 0, false, false},
		{"loopback without port", "https://p/auth?redirect_uri=http://localhost/cb", 0, false, false},
		{"no redirect param", "https://p/auth?client_id=x", 0, false, false},
		{"plain url", "http://localhost:7036", 0, false, false},
	}
	for _, c := range cases {
		port, v6, ok := detectLoopbackCallback(c.url)
		if port != c.wantPort || v6 != c.wantV6 || ok != c.wantOK {
			t.Errorf("%s: detectLoopbackCallback(%q) = (%d, %v, %v), want (%d, %v, %v)",
				c.name, c.url, port, v6, ok, c.wantPort, c.wantV6, c.wantOK)
		}
	}
}

func TestAllowedURL(t *testing.T) {
	cases := []struct {
		url   string
		extra []string
		want  bool
	}{
		{"http://localhost:7036", nil, true},
		{"https://example.com/path?q=1", nil, true},
		{"http://127.0.0.1:8080", nil, true},
		{"file:///etc/passwd", nil, false},
		{"javascript:alert(1)", nil, false},
		{"ftp://host/file", nil, false},
		{"http://", nil, false},
		{"not a url", nil, false},
		{"", nil, false},
		{"dropshare5:///upload?path=/tmp/x.png", []string{"dropshare5"}, true},
		{"DropShare5:///upload", []string{"dropshare5"}, true},
		{"dropshare5://host/upload", []string{"dropshare5"}, true},
		{"dropshare5:///upload", nil, false},
		{"dropshare5:///upload", []string{"vscode"}, false},
		{"file:///etc/passwd", []string{"file"}, true},
		{"http://", []string{"dropshare5"}, false},
		{"https://example.com", []string{"dropshare5"}, true},
	}
	for _, c := range cases {
		if got := allowedURL(c.url, c.extra); got != c.want {
			t.Errorf("allowedURL(%q, %v) = %v, want %v", c.url, c.extra, got, c.want)
		}
	}
}

func TestNormalizeSchemes(t *testing.T) {
	cases := []struct {
		name         string
		entries      []string
		wantSchemes  []string
		wantRejected []string
	}{
		{"bare name", []string{"dropshare5"}, []string{"dropshare5"}, nil},
		{"trailing colon", []string{"dropshare5:"}, []string{"dropshare5"}, nil},
		{"scheme slashes", []string{"dropshare5://"}, []string{"dropshare5"}, nil},
		{"empty authority", []string{"dropshare5:///"}, []string{"dropshare5"}, nil},
		{"mixed case", []string{"DropShare5:///"}, []string{"dropshare5"}, nil},
		{"surrounding space", []string{"  vscode  "}, []string{"vscode"}, nil},
		{"punctuated scheme", []string{"x-tool.v2+alt"}, []string{"x-tool.v2+alt"}, nil},
		{"duplicates collapse", []string{"vscode", "VSCode://"}, []string{"vscode"}, nil},
		{"http dropped", []string{"http", "https://", "vscode"}, []string{"vscode"}, nil},
		{"invalid entries rejected", []string{"1foo", "a b", "", "vscode"}, []string{"vscode"}, []string{"1foo", "a b", ""}},
		{"no entries", nil, nil, nil},
	}
	for _, c := range cases {
		schemes, rejected := NormalizeSchemes(c.entries)
		if !slices.Equal(schemes, c.wantSchemes) {
			t.Errorf("%s: schemes = %v, want %v", c.name, schemes, c.wantSchemes)
		}
		if !slices.Equal(rejected, c.wantRejected) {
			t.Errorf("%s: rejected = %v, want %v", c.name, rejected, c.wantRejected)
		}
	}
}

func TestAllowedSchemesFromEnv(t *testing.T) {
	cases := []struct {
		name string
		env  string
		want []string
	}{
		{"unset behaves as before", "", nil},
		{"single scheme", "dropshare5", []string{"dropshare5"}},
		{"comma list", "dropshare5,vscode", []string{"dropshare5", "vscode"}},
		{"malformed value filtered", "dropshare5,1foo,,https", []string{"dropshare5"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Setenv(OpenSchemesEnv, c.env)
			if got := allowedSchemes(); !slices.Equal(got, c.want) {
				t.Errorf("allowedSchemes() = %v, want %v", got, c.want)
			}
		})
	}
}

func TestOpenRulesSnippet(t *testing.T) {
	base := openRulesSnippet(nil)
	if !strings.Contains(base, "http(s) URL") {
		t.Errorf("default snippet does not describe http(s): %q", base)
	}
	if strings.Contains(base, "registered for them on the host") {
		t.Errorf("default snippet names extra schemes: %q", base)
	}
	if got := openRulesSnippet(&SnippetConfig{}); got != base {
		t.Errorf("snippet with empty config = %q, want the default", got)
	}
	withSchemes := openRulesSnippet(&SnippetConfig{Schemes: []string{"DropShare5:///", "vscode"}})
	if !strings.HasPrefix(withSchemes, base) {
		t.Errorf("configured snippet dropped the default text: %q", withSchemes)
	}
	if !strings.Contains(withSchemes, "`dropshare5:`") || !strings.Contains(withSchemes, "`vscode:`") {
		t.Errorf("configured snippet does not name both schemes: %q", withSchemes)
	}
}
