package container

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/inventage-ai/asylum/internal/config"
)

// stubAgent satisfies agent.Agent for tests that need a minimal implementation.
type stubAgent struct {
	envVars    map[string]string
	hasSession bool
	command    []string
}

func (s stubAgent) Name() string                          { return "stub" }
func (s stubAgent) Binary() string                        { return "stub-bin" }
func (s stubAgent) NativeConfigDir() string               { return "~/.stub" }
func (s stubAgent) ContainerConfigDir() string            { return "/home/stub/.stub" }
func (s stubAgent) AsylumConfigDir() string               { return "~/.asylum/agents/stub" }
func (s stubAgent) EnvVars() map[string]string            { return s.envVars }
func (s stubAgent) HasSession(projectPath string) bool    { return s.hasSession }
func (s stubAgent) Command(resume bool, extra []string) []string {
	if resume {
		return append([]string{"stub-resume"}, extra...)
	}
	return append([]string{"stub"}, extra...)
}

func TestCopyDir(t *testing.T) {
	t.Run("copies files and nested directories", func(t *testing.T) {
		src := t.TempDir()
		dst := t.TempDir()

		os.MkdirAll(filepath.Join(src, "sub"), 0755)
		os.WriteFile(filepath.Join(src, "file.txt"), []byte("hello"), 0644)
		os.WriteFile(filepath.Join(src, "sub", "nested.txt"), []byte("world"), 0644)

		if err := copyDir(src, dst); err != nil {
			t.Fatal(err)
		}

		data, err := os.ReadFile(filepath.Join(dst, "file.txt"))
		if err != nil || string(data) != "hello" {
			t.Errorf("file.txt: got %q, err %v", data, err)
		}
		data, err = os.ReadFile(filepath.Join(dst, "sub", "nested.txt"))
		if err != nil || string(data) != "world" {
			t.Errorf("sub/nested.txt: got %q, err %v", data, err)
		}
	})

	t.Run("preserves file permissions", func(t *testing.T) {
		src := t.TempDir()
		dst := t.TempDir()

		os.WriteFile(filepath.Join(src, "exec.sh"), []byte("#!/bin/sh"), 0755)

		if err := copyDir(src, dst); err != nil {
			t.Fatal(err)
		}

		info, err := os.Stat(filepath.Join(dst, "exec.sh"))
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0755 {
			t.Errorf("permissions = %o, want 0755", info.Mode().Perm())
		}
	})

	t.Run("recreates symlinks", func(t *testing.T) {
		src := t.TempDir()
		dst := t.TempDir()

		os.WriteFile(filepath.Join(src, "target.txt"), []byte("data"), 0644)
		os.Symlink("target.txt", filepath.Join(src, "link.txt"))

		if err := copyDir(src, dst); err != nil {
			t.Fatal(err)
		}

		linkTarget, err := os.Readlink(filepath.Join(dst, "link.txt"))
		if err != nil {
			t.Fatalf("Readlink: %v", err)
		}
		if linkTarget != "target.txt" {
			t.Errorf("symlink target = %q, want %q", linkTarget, "target.txt")
		}
	})

	t.Run("propagates error on unreadable source file", func(t *testing.T) {
		if os.Getuid() == 0 {
			t.Skip("root ignores permission bits")
		}
		src := t.TempDir()
		dst := t.TempDir()

		path := filepath.Join(src, "unreadable.txt")
		os.WriteFile(path, []byte("data"), 0000)
		defer os.Chmod(path, 0644)

		if err := copyDir(src, dst); err == nil {
			t.Error("expected error reading unreadable file")
		}
	})
}

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

func TestAppendPorts(t *testing.T) {
	tests := []struct {
		name  string
		ports []string
		want  []string
	}{
		{
			name:  "no ports",
			ports: nil,
			want:  []string{},
		},
		{
			name:  "port without colon expands to host:container",
			ports: []string{"8080"},
			want:  []string{"-p", "8080:8080"},
		},
		{
			name:  "port with colon used as-is",
			ports: []string{"8080:9090"},
			want:  []string{"-p", "8080:9090"},
		},
		{
			name:  "multiple ports mixed",
			ports: []string{"3000", "4000:5000"},
			want:  []string{"-p", "3000:3000", "-p", "4000:5000"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := appendPorts([]string{}, tt.ports)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("appendPorts(%v) = %v, want %v", tt.ports, got, tt.want)
			}
		})
	}
}

func TestAppendEnvVars(t *testing.T) {
	t.Run("always includes required env vars", func(t *testing.T) {
		opts := RunOpts{
			Config:     config.Config{},
			Agent:      stubAgent{envVars: map[string]string{}},
			ProjectDir: "/work/myproject",
		}
		got := appendEnvVars([]string{}, opts)
		joined := strings.Join(got, " ")

		for _, want := range []string{
			"-e ASYLUM_DOCKER=1",
			"-e HISTFILE=/home/claude/.shell_history/zsh_history",
			"-e HOST_PROJECT_DIR=/work/myproject",
		} {
			if !strings.Contains(joined, want) {
				t.Errorf("expected %q in args %v", want, got)
			}
		}
	})

	t.Run("java version included when set", func(t *testing.T) {
		cfg := config.Config{Versions: map[string]string{"java": "17"}}
		opts := RunOpts{
			Config:     cfg,
			Agent:      stubAgent{envVars: map[string]string{}},
			ProjectDir: "/work/proj",
		}
		got := appendEnvVars([]string{}, opts)
		joined := strings.Join(got, " ")
		if !strings.Contains(joined, "-e ASYLUM_JAVA_VERSION=17") {
			t.Errorf("expected ASYLUM_JAVA_VERSION=17 in %v", got)
		}
	})

	t.Run("java version omitted when empty", func(t *testing.T) {
		opts := RunOpts{
			Config:     config.Config{},
			Agent:      stubAgent{envVars: map[string]string{}},
			ProjectDir: "/work/proj",
		}
		got := appendEnvVars([]string{}, opts)
		for _, v := range got {
			if strings.HasPrefix(v, "ASYLUM_JAVA_VERSION") {
				t.Errorf("unexpected ASYLUM_JAVA_VERSION in %v", got)
			}
		}
	})

	t.Run("agent env vars included", func(t *testing.T) {
		opts := RunOpts{
			Config:     config.Config{},
			Agent:      stubAgent{envVars: map[string]string{"MY_TOKEN": "secret"}},
			ProjectDir: "/work/proj",
		}
		got := appendEnvVars([]string{}, opts)
		joined := strings.Join(got, " ")
		if !strings.Contains(joined, "-e MY_TOKEN=secret") {
			t.Errorf("expected MY_TOKEN=secret in %v", got)
		}
	})
}

func TestContainerCommand(t *testing.T) {
	projectDir := t.TempDir()

	tests := []struct {
		name string
		opts RunOpts
		want []string
	}{
		{
			name: "shell mode",
			opts: RunOpts{Mode: ModeShell},
			want: []string{"/bin/zsh"},
		},
		{
			name: "admin shell mode",
			opts: RunOpts{Mode: ModeAdminShell},
			want: []string{"bash", "-c", "echo 'Admin shell - sudo access enabled' && exec /bin/zsh"},
		},
		{
			name: "command mode passes extra args through",
			opts: RunOpts{Mode: ModeCommand, ExtraArgs: []string{"ls", "-la"}},
			want: []string{"ls", "-la"},
		},
		{
			name: "agent mode with new session (no resume)",
			opts: RunOpts{
				Mode:       ModeAgent,
				NewSession: true,
				Agent:      stubAgent{hasSession: true},
				ProjectDir: projectDir,
			},
			want: []string{"stub"},
		},
		{
			name: "agent mode resumes when session exists",
			opts: RunOpts{
				Mode:       ModeAgent,
				NewSession: false,
				Agent:      stubAgent{hasSession: true},
				ProjectDir: projectDir,
			},
			want: []string{"stub-resume"},
		},
		{
			name: "agent mode no resume when session absent",
			opts: RunOpts{
				Mode:       ModeAgent,
				NewSession: false,
				Agent:      stubAgent{hasSession: false},
				ProjectDir: projectDir,
			},
			want: []string{"stub"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containerCommand(tt.opts)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("containerCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContainerCommandAgentExtraArgs(t *testing.T) {
	dir := t.TempDir()
	opts := RunOpts{
		Mode:       ModeAgent,
		NewSession: false,
		Agent:      stubAgent{hasSession: false},
		ProjectDir: dir,
		ExtraArgs:  []string{"fix", "the", "bug"},
	}
	got := containerCommand(opts)
	want := []string{"stub", "fix", "the", "bug"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("containerCommand() = %v, want %v", got, want)
	}
}

func TestContainerName(t *testing.T) {
	name1 := containerName("/home/user/projectA")
	name2 := containerName("/home/user/projectB")

	if name1 == name2 {
		t.Error("different project dirs should produce different container names")
	}
	if !strings.HasPrefix(name1, "asylum-") {
		t.Errorf("container name %q should start with asylum-", name1)
	}
	// Should be deterministic
	if containerName("/home/user/projectA") != name1 {
		t.Error("containerName should be deterministic")
	}
}

func TestAppendVolumesUserVolumes(t *testing.T) {
	home := t.TempDir()
	projectDir := t.TempDir()
	cname := containerName(projectDir)

	agentConfigDir := filepath.Join(home, ".asylum", "agents", "stub")
	if err := os.MkdirAll(agentConfigDir, 0755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		volumes     []string
		wantHost    string
		wantCont    string
		wantOptions string
	}{
		{
			name:     "simple absolute path mounts same on both sides",
			volumes:  []string{"/data"},
			wantHost: "/data",
			wantCont: "/data",
		},
		{
			name:     "host:container volume",
			volumes:  []string{"/src:/dst"},
			wantHost: "/src",
			wantCont: "/dst",
		},
		{
			name:        "host:container:options volume",
			volumes:     []string{"/src:/dst:ro"},
			wantHost:    "/src",
			wantCont:    "/dst",
			wantOptions: "ro",
		},
		{
			name:     "tilde expanded in host path",
			volumes:  []string{"~/data:/data"},
			wantHost: filepath.Join(home, "data"),
			wantCont: "/data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := RunOpts{
				Config:     config.Config{Volumes: tt.volumes},
				Agent:      stubAgent{},
				ProjectDir: projectDir,
			}

			args, err := appendVolumes([]string{}, home, cname, opts)
			if err != nil {
				t.Fatalf("appendVolumes: %v", err)
			}

			wantMount := tt.wantHost + ":" + tt.wantCont
			if tt.wantOptions != "" {
				wantMount += ":" + tt.wantOptions
			}

			found := false
			for i, arg := range args {
				if arg == "-v" && i+1 < len(args) && args[i+1] == wantMount {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected -v %q in args %v", wantMount, args)
			}
		})
	}
}

