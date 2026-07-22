package kit

import "testing"

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

func TestValidBrowserURL(t *testing.T) {
	cases := []struct {
		url  string
		want bool
	}{
		{"http://localhost:7036", true},
		{"https://example.com/path?q=1", true},
		{"http://127.0.0.1:8080", true},
		{"file:///etc/passwd", false},
		{"javascript:alert(1)", false},
		{"ftp://host/file", false},
		{"http://", false},
		{"not a url", false},
		{"", false},
	}
	for _, c := range cases {
		if got := validBrowserURL(c.url); got != c.want {
			t.Errorf("validBrowserURL(%q) = %v, want %v", c.url, got, c.want)
		}
	}
}
