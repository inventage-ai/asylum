package docker

import (
	"maps"
	"slices"
	"testing"
)

func TestCountExecSessions(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{
			"empty",
			"",
			0,
		},
		{
			"only PID 1 and check command",
			"    1     0\n   42     0\n",
			1, // 42 is the check command
		},
		{
			"one other session plus check",
			"    1     0\n   38     0\n   42     0\n",
			2,
		},
		{
			"background tasks ignored",
			"    1     0\n    7     1\n  314     1\n   42     0\n",
			1, // only 42 (check), 7 and 314 have PPID=1
		},
		{
			"two sessions plus background tasks",
			"    1     0\n    7     1\n   38     0\n  115     0\n  314     1\n   42     0\n",
			3, // 38, 115, 42
		},
		{
			"child processes of sessions ignored",
			"    1     0\n   38     0\n  114    38\n   42     0\n",
			2, // 38 and 42; 114 has PPID=38
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countExecSessions(tt.input)
			if got != tt.want {
				t.Errorf("countExecSessions() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestParseNames(t *testing.T) {
	got := parseNames("asylum-8a3fef05032f-asylum\nasylum-b269a7c11f58-proj\n\n")
	want := map[string]bool{"asylum-8a3fef05032f-asylum": true, "asylum-b269a7c11f58-proj": true}
	if !maps.Equal(got, want) {
		t.Errorf("parseNames() = %v, want %v", got, want)
	}
	if got := parseNames(""); len(got) != 0 {
		t.Errorf("parseNames(\"\") = %v, want empty", got)
	}
}

func TestParseVolumeLabels(t *testing.T) {
	out := "asylum-a-app-cache-npm\t/home/u/app\nasylum-b-old-cache-npm\t\nother-asylum-x\t/x\n"
	got := parseVolumeLabels(out, "asylum-")
	want := map[string]string{"asylum-a-app-cache-npm": "/home/u/app", "asylum-b-old-cache-npm": ""}
	if !maps.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestParseContainerMounts(t *testing.T) {
	out := "asylum-a-app\trunning\tasylum:proj-abc\tasylum-a-app-cache-npm,/home/u/app\nasylum-b-old\texited\tasylum:latest\tasylum-b-old-npm-123\nfoo\trunning\tx\ty\n"
	got := parseContainerMounts(out, "asylum-")
	if len(got) != 2 {
		t.Fatalf("expected 2 containers, got %v", got)
	}
	if got[0].Name != "asylum-a-app" || !got[0].Running || got[0].Image != "asylum:proj-abc" || !slices.Equal(got[0].Mounts, []string{"asylum-a-app-cache-npm", "/home/u/app"}) {
		t.Errorf("unexpected first container %+v", got[0])
	}
	if got[1].Name != "asylum-b-old" || got[1].Running {
		t.Errorf("unexpected second container %+v", got[1])
	}
}

func TestParseSize(t *testing.T) {
	tests := map[string]int64{"0B": 0, "512B": 512, "12.3kB": 12300, "8.652MB": 8652000, "1.2GB": 1200000000, "2TB": 2e12, "N/A": 0, "": 0}
	for in, want := range tests {
		if got := parseSize(in); got != want {
			t.Errorf("parseSize(%q) = %d, want %d", in, got, want)
		}
	}
}

func TestParseDiskUsage(t *testing.T) {
	out := `{"Images":[{"Repository":"asylum","Tag":"proj-abc","Size":"5.1GB","UniqueSize":"300MB"},{"Repository":"alpine","Tag":"latest","Size":"8.65MB","UniqueSize":""}],"Volumes":[{"Name":"asylum-a-app-cache-npm","Size":"2MB"}],"Containers":[],"BuildCache":[]}`
	vols, imgs, err := parseDiskUsage([]byte(out))
	if err != nil {
		t.Fatal(err)
	}
	if vols["asylum-a-app-cache-npm"] != 2e6 || imgs["asylum:proj-abc"] != 3e8 || imgs["alpine:latest"] != 8650000 {
		t.Errorf("got volumes %v, images %v", vols, imgs)
	}
}
