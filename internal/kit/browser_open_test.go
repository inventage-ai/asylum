package kit

import "testing"

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
