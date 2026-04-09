package container

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/inventage-ai/asylum/internal/config"
	"github.com/inventage-ai/asylum/internal/kit"
)

func boolPtr(b bool) *bool { return &b }

// stubAgent satisfies agent.Agent for tests that need a minimal implementation.
type stubAgent struct {
	envVars         map[string]string
	hasSession      bool
	command         []string
	asylumConfigDir string
	nativeConfigDir string
}

func (s stubAgent) Name() string    { return "stub" }
func (s stubAgent) Binary() string  { return "stub-bin" }
func (s stubAgent) NativeConfigDir() string {
	if s.nativeConfigDir != "" {
		return s.nativeConfigDir
	}
	return "~/.stub"
}
func (s stubAgent) ContainerConfigDir() string { return "/home/stub/.stub" }
func (s stubAgent) AsylumConfigDir() string {
	if s.asylumConfigDir != "" {
		return s.asylumConfigDir
	}
	return "~/.asylum/agents/stub"
}
func (s stubAgent) EnvVars() map[string]string            { return s.envVars }
func (s stubAgent) HasSession(_, _ string) bool            { return s.hasSession }
func (s stubAgent) Command(resume bool, extra []string) []string {
	if resume {
		return append([]string{"stub-resume"}, extra...)
	}
	return append([]string{"stub"}, extra...)
}

// claudeStubAgent wraps stubAgent but returns "claude" for Name().
type claudeStubAgent struct{ stubAgent }

func (claudeStubAgent) Name() string { return "claude" }

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

	t.Run("resolves symlinks to regular files", func(t *testing.T) {
		src := t.TempDir()
		dst := t.TempDir()

		os.WriteFile(filepath.Join(src, "target.txt"), []byte("data"), 0644)
		os.Symlink("target.txt", filepath.Join(src, "link.txt"))

		if err := copyDir(src, dst); err != nil {
			t.Fatal(err)
		}

		// Symlink should be resolved to a regular file copy
		data, err := os.ReadFile(filepath.Join(dst, "link.txt"))
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		if string(data) != "data" {
			t.Errorf("content = %q, want %q", data, "data")
		}
		info, err := os.Lstat(filepath.Join(dst, "link.txt"))
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Type() == fs.ModeSymlink {
			t.Error("expected regular file, got symlink")
		}
	})

	t.Run("skips dangling symlinks", func(t *testing.T) {
		src := t.TempDir()
		dst := t.TempDir()

		os.Symlink("nonexistent", filepath.Join(src, "dangling.txt"))

		if err := copyDir(src, dst); err != nil {
			t.Fatal(err)
		}

		if _, err := os.Lstat(filepath.Join(dst, "dangling.txt")); !os.IsNotExist(err) {
			t.Errorf("expected dangling symlink to be skipped, got err: %v", err)
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

func hasRunArg(args []kit.RunArg, flag, value string) bool {
	for _, a := range args {
		if a.Flag == flag && a.Value == value {
			return true
		}
	}
	return false
}

func TestConfigPortArgs(t *testing.T) {
	tests := []struct {
		name      string
		ports     []string
		wantPairs [][2]string // flag, value pairs
		wantErr   bool
	}{
		{
			name:  "no ports",
			ports: nil,
		},
		{
			name:      "port without colon expands to host:container",
			ports:     []string{"8080"},
			wantPairs: [][2]string{{"-p", "8080:8080"}},
		},
		{
			name:      "port with colon used as-is",
			ports:     []string{"8080:9090"},
			wantPairs: [][2]string{{"-p", "8080:9090"}},
		},
		{
			name:      "multiple ports mixed",
			ports:     []string{"3000", "4000:5000"},
			wantPairs: [][2]string{{"-p", "3000:3000"}, {"-p", "4000:5000"}},
		},
		{
			name:    "non-numeric port rejected",
			ports:   []string{"abc"},
			wantErr: true,
		},
		{
			name:    "port zero rejected",
			ports:   []string{"0"},
			wantErr: true,
		},
		{
			name:    "port above 65535 rejected",
			ports:   []string{"70000"},
			wantErr: true,
		},
		{
			name:    "invalid host in mapping rejected",
			ports:   []string{"abc:8080"},
			wantErr: true,
		},
		{
			name:    "invalid container in mapping rejected",
			ports:   []string{"8080:abc"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := configPortArgs(tt.ports)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(tt.wantPairs) == 0 && len(got) != 0 {
				t.Errorf("expected no args, got %v", got)
			}
			for _, pair := range tt.wantPairs {
				if !hasRunArg(got, pair[0], pair[1]) {
					t.Errorf("expected RunArg{Flag:%q, Value:%q} in %v", pair[0], pair[1], got)
				}
			}
		})
	}
}

func TestCoreEnvVars(t *testing.T) {
	t.Run("always includes required env vars", func(t *testing.T) {
		home := t.TempDir()
		opts := RunOpts{
			Config:     config.Config{},
			Agent:      stubAgent{envVars: map[string]string{}},
			ProjectDir: "/work/myproject",
		}
		got, err := coreEnvVars(home, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for _, want := range []string{
			"COLORTERM=truecolor",
			"TERM=xterm-256color",
			"HOST_PROJECT_DIR=/work/myproject",
		} {
			if !hasRunArg(got, "-e", want) {
				t.Errorf("expected RunArg{Flag:\"-e\", Value:%q} in %v", want, got)
			}
		}
		// HISTFILE is dynamic, just check it's present
		found := false
		for _, a := range got {
			if a.Flag == "-e" && strings.HasPrefix(a.Value, "HISTFILE=") {
				found = true
			}
		}
		if !found {
			t.Error("expected HISTFILE env var")
		}
	})

	t.Run("java version included when set", func(t *testing.T) {
		home := t.TempDir()
		cfg := config.Config{Kits: map[string]*config.KitConfig{"java": {DefaultVersion: "17"}}}
		opts := RunOpts{
			Config:     cfg,
			Agent:      stubAgent{envVars: map[string]string{}},
			ProjectDir: "/work/proj",
		}
		got, err := coreEnvVars(home, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !hasRunArg(got, "-e", "ASYLUM_JAVA_VERSION=17") {
			t.Errorf("expected ASYLUM_JAVA_VERSION=17 in %v", got)
		}
	})

	t.Run("java version omitted when empty", func(t *testing.T) {
		home := t.TempDir()
		opts := RunOpts{
			Config:     config.Config{},
			Agent:      stubAgent{envVars: map[string]string{}},
			ProjectDir: "/work/proj",
		}
		got, err := coreEnvVars(home, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, a := range got {
			if a.Flag == "-e" && strings.HasPrefix(a.Value, "ASYLUM_JAVA_VERSION") {
				t.Errorf("unexpected ASYLUM_JAVA_VERSION in %v", got)
			}
		}
	})

	t.Run("agent env vars included", func(t *testing.T) {
		home := t.TempDir()
		opts := RunOpts{
			Config:     config.Config{},
			Agent:      stubAgent{envVars: map[string]string{"MY_TOKEN": "secret"}},
			ProjectDir: "/work/proj",
		}
		got, err := coreEnvVars(home, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !hasRunArg(got, "-e", "MY_TOKEN=secret") {
			t.Errorf("expected MY_TOKEN=secret in %v", got)
		}
	})
}

func TestExecArgsAllModes(t *testing.T) {
	projectDir := t.TempDir()

	tests := []struct {
		name string
		opts ExecOpts
		want []string
	}{
		{
			name: "shell mode",
			opts: ExecOpts{ContainerName: "test", Mode: ModeShell},
			want: []string{"exec", "-it", "test", "/bin/zsh"},
		},
		{
			name: "admin shell mode",
			opts: ExecOpts{ContainerName: "test", Mode: ModeAdminShell},
			want: []string{"exec", "-it", "-u", "root", "test", "/bin/zsh"},
		},
		{
			name: "command mode passes extra args through",
			opts: ExecOpts{ContainerName: "test", Mode: ModeCommand, ExtraArgs: []string{"ls", "-la"}},
			want: []string{"exec", "-it", "test", "ls", "-la"},
		},
		{
			name: "agent mode with new session (no resume)",
			opts: ExecOpts{
				ContainerName: "test",
				Mode:          ModeAgent,
				NewSession:    true,
				Agent:         stubAgent{hasSession: true},
				ProjectDir:    projectDir,
			},
			want: []string{"exec", "-it", "test", "stub"},
		},
		{
			name: "agent mode resumes when session exists",
			opts: ExecOpts{
				ContainerName: "test",
				Mode:          ModeAgent,
				NewSession:    false,
				Agent:         stubAgent{hasSession: true},
				ProjectDir:    projectDir,
			},
			want: []string{"exec", "-it", "test", "stub-resume"},
		},
		{
			name: "agent mode no resume when session absent",
			opts: ExecOpts{
				ContainerName: "test",
				Mode:          ModeAgent,
				NewSession:    false,
				Agent:         stubAgent{hasSession: false},
				ProjectDir:    projectDir,
			},
			want: []string{"exec", "-it", "test", "stub"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExecArgs(tt.opts)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExecArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecArgsAgentExtraArgs(t *testing.T) {
	dir := t.TempDir()
	opts := ExecOpts{
		ContainerName: "test",
		Mode:          ModeAgent,
		NewSession:    false,
		Agent:         stubAgent{hasSession: false},
		ProjectDir:    dir,
		ExtraArgs:     []string{"fix", "the", "bug"},
	}
	got := ExecArgs(opts)
	want := []string{"exec", "-it", "test", "stub", "fix", "the", "bug"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ExecArgs() = %v, want %v", got, want)
	}
}

func TestContainerName(t *testing.T) {
	name1 := ContainerName("/home/user/projectA")
	name2 := ContainerName("/home/user/projectB")

	if name1 == name2 {
		t.Error("different project dirs should produce different container names")
	}
	if !strings.HasPrefix(name1, "asylum-") {
		t.Errorf("container name %q should start with asylum-", name1)
	}
	if !strings.HasSuffix(name1, "-projecta") {
		t.Errorf("container name %q should end with project suffix", name1)
	}
	// Should be deterministic
	if ContainerName("/home/user/projectA") != name1 {
		t.Error("containerName should be deterministic")
	}
}

func TestOldContainerName(t *testing.T) {
	old := OldContainerName("/home/user/projectA")
	if !strings.HasPrefix(old, "asylum-") {
		t.Errorf("old name %q should start with asylum-", old)
	}
	// Old format should NOT contain the project suffix
	if strings.Contains(old, "projecta") {
		t.Errorf("old name %q should not contain project suffix", old)
	}
	// New name should start with old name
	newName := ContainerName("/home/user/projectA")
	if !strings.HasPrefix(newName, old+"-") {
		t.Errorf("new name %q should start with old name %q plus hyphen", newName, old)
	}
}

func TestSanitizeProject(t *testing.T) {
	tests := []struct {
		projectDir string
		want       string
	}{
		{"/home/user/my-project", "my-project"},
		{"/home/user/MyApp", "myapp"},
		{"/home/user/hello world", "hello-world"},
		{"/home/user/foo@bar!baz", "foo-bar-baz"},
		{"/home/user/---trimmed---", "trimmed"},
		{"/", "project"}, // empty after sanitization
	}
	for _, tt := range tests {
		t.Run(tt.projectDir, func(t *testing.T) {
			got := sanitizeProject(tt.projectDir)
			if got != tt.want {
				t.Errorf("sanitizeProject(%q) = %q, want %q", tt.projectDir, got, tt.want)
			}
		})
	}
}

func TestConfigVolumeArgs(t *testing.T) {
	home := t.TempDir()

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
			args, err := configVolumeArgs(tt.volumes, home)
			if err != nil {
				t.Fatalf("configVolumeArgs: %v", err)
			}

			wantMount := tt.wantHost + ":" + tt.wantCont
			if tt.wantOptions != "" {
				wantMount += ":" + tt.wantOptions
			}

			if !hasRunArg(args, "-v", wantMount) {
				t.Errorf("expected RunArg{Flag:\"-v\", Value:%q} in %v", wantMount, args)
			}
		})
	}
}

func TestResolveGitWorktree(t *testing.T) {
	t.Run("worktree detected", func(t *testing.T) {
		// Simulate: project/.git is a file, worktree dir has commondir
		project := t.TempDir()
		mainRepo := t.TempDir()
		mainGit := filepath.Join(mainRepo, ".git")
		wtDir := filepath.Join(mainGit, "worktrees", "feature")
		os.MkdirAll(wtDir, 0755)
		os.MkdirAll(mainGit, 0755)

		// project/.git file points to worktree dir
		os.WriteFile(filepath.Join(project, ".git"), []byte("gitdir: "+wtDir+"\n"), 0644)
		// worktree dir has commondir pointing to main .git
		os.WriteFile(filepath.Join(wtDir, "commondir"), []byte(mainGit+"\n"), 0644)

		wt, common := resolveGitWorktree(project)
		if wt != wtDir {
			t.Errorf("worktreeDir = %q, want %q", wt, wtDir)
		}
		if common != mainGit {
			t.Errorf("commonDir = %q, want %q", common, mainGit)
		}
	})

	t.Run("regular repo", func(t *testing.T) {
		project := t.TempDir()
		os.MkdirAll(filepath.Join(project, ".git"), 0755)

		wt, common := resolveGitWorktree(project)
		if wt != "" || common != "" {
			t.Errorf("expected empty strings for regular repo, got %q, %q", wt, common)
		}
	})

	t.Run("no git", func(t *testing.T) {
		project := t.TempDir()

		wt, common := resolveGitWorktree(project)
		if wt != "" || common != "" {
			t.Errorf("expected empty strings for no git, got %q, %q", wt, common)
		}
	})

	t.Run("relative gitdir", func(t *testing.T) {
		project := t.TempDir()
		mainGit := filepath.Join(project, "..", "main-repo", ".git")
		wtDir := filepath.Join(mainGit, "worktrees", "feature")
		os.MkdirAll(wtDir, 0755)

		// Use relative path in .git file
		os.WriteFile(filepath.Join(project, ".git"), []byte("gitdir: ../main-repo/.git/worktrees/feature\n"), 0644)
		os.WriteFile(filepath.Join(wtDir, "commondir"), []byte("../..\n"), 0644)

		wt, common := resolveGitWorktree(project)
		if wt == "" {
			t.Fatal("expected worktreeDir, got empty")
		}
		if common == "" {
			t.Fatal("expected commonDir, got empty")
		}
		// Both should resolve to absolute clean paths
		if !filepath.IsAbs(wt) {
			t.Errorf("worktreeDir should be absolute, got %q", wt)
		}
		if !filepath.IsAbs(common) {
			t.Errorf("commonDir should be absolute, got %q", common)
		}
	})
}

func TestFindNodeModulesDirs(t *testing.T) {
	t.Run("returns node_modules for package.json", func(t *testing.T) {
		project := t.TempDir()
		os.WriteFile(filepath.Join(project, "package.json"), []byte("{}"), 0644)

		got := FindNodeModulesDirs(project)
		want := []string{filepath.Join(project, "node_modules")}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("works even when node_modules does not exist", func(t *testing.T) {
		project := t.TempDir()
		os.WriteFile(filepath.Join(project, "package.json"), []byte("{}"), 0644)
		// No node_modules directory created

		got := FindNodeModulesDirs(project)
		want := []string{filepath.Join(project, "node_modules")}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("finds monorepo workspace node_modules", func(t *testing.T) {
		project := t.TempDir()
		os.WriteFile(filepath.Join(project, "package.json"), []byte("{}"), 0644)
		pkgDir := filepath.Join(project, "packages", "app")
		os.MkdirAll(pkgDir, 0755)
		os.WriteFile(filepath.Join(pkgDir, "package.json"), []byte("{}"), 0644)

		got := FindNodeModulesDirs(project)
		want := []string{
			filepath.Join(project, "node_modules"),
			filepath.Join(pkgDir, "node_modules"),
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("no package.json anywhere returns empty", func(t *testing.T) {
		project := t.TempDir()
		os.MkdirAll(filepath.Join(project, "src"), 0755)
		got := FindNodeModulesDirs(project)
		if len(got) != 0 {
			t.Errorf("got %v, want empty", got)
		}
	})

	t.Run("package.json only in subdirectory", func(t *testing.T) {
		project := t.TempDir()
		frontend := filepath.Join(project, "frontend")
		os.MkdirAll(frontend, 0755)
		os.WriteFile(filepath.Join(frontend, "package.json"), []byte("{}"), 0644)

		got := FindNodeModulesDirs(project)
		want := []string{filepath.Join(frontend, "node_modules")}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("skips heavy directories", func(t *testing.T) {
		project := t.TempDir()
		os.WriteFile(filepath.Join(project, "package.json"), []byte("{}"), 0644)
		// package.json inside .venv should not be found
		venvPkg := filepath.Join(project, ".venv", "lib")
		os.MkdirAll(venvPkg, 0755)
		os.WriteFile(filepath.Join(venvPkg, "package.json"), []byte("{}"), 0644)

		got := FindNodeModulesDirs(project)
		// Only root node_modules, not .venv/lib/node_modules
		want := []string{filepath.Join(project, "node_modules")}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("skips .claude worktrees", func(t *testing.T) {
		project := t.TempDir()
		os.WriteFile(filepath.Join(project, "package.json"), []byte("{}"), 0644)
		// package.json inside a Claude worktree should not be found
		wtPkg := filepath.Join(project, ".claude", "worktrees", "feat-x")
		os.MkdirAll(wtPkg, 0755)
		os.WriteFile(filepath.Join(wtPkg, "package.json"), []byte("{}"), 0644)

		got := FindNodeModulesDirs(project)
		want := []string{filepath.Join(project, "node_modules")}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("does not recurse into existing node_modules", func(t *testing.T) {
		project := t.TempDir()
		os.WriteFile(filepath.Join(project, "package.json"), []byte("{}"), 0644)
		// Some npm packages have their own package.json inside node_modules
		nested := filepath.Join(project, "node_modules", "some-pkg")
		os.MkdirAll(nested, 0755)
		os.WriteFile(filepath.Join(nested, "package.json"), []byte("{}"), 0644)

		got := FindNodeModulesDirs(project)
		want := []string{filepath.Join(project, "node_modules")}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}

func TestCoreVolumesCacheNamedVolumes(t *testing.T) {
	home := t.TempDir()
	projectDir := t.TempDir()
	cname := ContainerName(projectDir)

	agentConfigDir := filepath.Join(home, ".asylum", "agents", "stub")
	os.MkdirAll(agentConfigDir, 0755)

	opts := RunOpts{
		Config:     config.Config{},
		Agent:      stubAgent{},
		ProjectDir: projectDir,
		CacheDirs: map[string]string{
			"npm":    "~/.npm",
			"pip":    "~/.cache/pip",
			"maven":  "~/.m2",
			"gradle": "~/.gradle",
		},
	}

	args, err := coreVolumes(home, cname, opts)
	if err != nil {
		t.Fatalf("coreVolumes: %v", err)
	}

	for _, tool := range []string{"gradle", "maven", "npm", "pip"} {
		wantPrefix := "type=volume,src=" + cname + "-cache-" + tool + ",dst="
		found := false
		for _, a := range args {
			if a.Flag == "--mount" && strings.HasPrefix(a.Value, wantPrefix) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected named volume mount for cache %q, not found in %v", tool, args)
		}
	}

	// No bind mount to ~/.asylum/cache/ should exist
	for _, a := range args {
		if a.Flag == "-v" && strings.Contains(a.Value, ".asylum/cache") {
			t.Errorf("unexpected bind mount for cache: %q", a.Value)
		}
	}
}

func TestKitCredentialArgs(t *testing.T) {
	home := t.TempDir()
	projectDir := t.TempDir()
	cname := ContainerName(projectDir)

	agentConfigDir := filepath.Join(home, ".asylum", "agents", "stub")
	os.MkdirAll(agentConfigDir, 0755)

	// Create a kit with a CredentialFunc that returns a mount
	testKit := &kit.Kit{
		Name: "java/maven",
		CredentialFunc: func(opts kit.CredentialOpts) ([]kit.CredentialMount, error) {
			return []kit.CredentialMount{
				{
					Content:     []byte("<settings><servers/></settings>"),
					Destination: "~/.m2/settings.xml",
				},
			}, nil
		},
	}

	creds := &config.Credentials{Auto: true}
	opts := RunOpts{
		Config: config.Config{Kits: map[string]*config.KitConfig{
			"java": {Credentials: creds},
		}},
		Agent:      stubAgent{},
		ProjectDir: projectDir,
		CacheDirs:  map[string]string{"maven": "~/.m2"},
		Kits:       []*kit.Kit{testKit},
	}

	args, err := kitCredentialArgs(home, cname, opts)
	if err != nil {
		t.Fatalf("kitCredentialArgs: %v", err)
	}

	// Find credential mount
	credFound := false
	for _, a := range args {
		if a.Flag == "-v" && strings.Contains(a.Value, "settings.xml") && strings.Contains(a.Value, "credentials") {
			credFound = true
			if !strings.HasSuffix(a.Value, ":ro") {
				t.Errorf("credential mount should be read-only, got: %s", a.Value)
			}
		}
	}
	if !credFound {
		t.Fatal("credential mount not found")
	}

	// Verify credential file was written
	credFile := filepath.Join(home, ".asylum", "projects", cname, "credentials", "settings.xml")
	data, err := os.ReadFile(credFile)
	if err != nil {
		t.Fatalf("credential file not written: %v", err)
	}
	if string(data) != "<settings><servers/></settings>" {
		t.Errorf("unexpected credential content: %q", data)
	}
}

func TestKitCredentialArgsHostPath(t *testing.T) {
	home := t.TempDir()
	projectDir := t.TempDir()
	cname := ContainerName(projectDir)

	agentConfigDir := filepath.Join(home, ".asylum", "agents", "stub")
	os.MkdirAll(agentConfigDir, 0755)

	// Create a host directory to mount
	ghDir := filepath.Join(home, ".config", "gh")
	os.MkdirAll(ghDir, 0755)

	testKit := &kit.Kit{
		Name: "github",
		CredentialFunc: func(opts kit.CredentialOpts) ([]kit.CredentialMount, error) {
			return []kit.CredentialMount{
				{
					HostPath:    ghDir,
					Destination: "~/.config/gh",
				},
			}, nil
		},
	}

	creds := &config.Credentials{Auto: true}
	opts := RunOpts{
		Config: config.Config{Kits: map[string]*config.KitConfig{
			"github": {Credentials: creds},
		}},
		Agent:      stubAgent{},
		ProjectDir: projectDir,
		Kits:       []*kit.Kit{testKit},
	}

	args, err := kitCredentialArgs(home, cname, opts)
	if err != nil {
		t.Fatalf("kitCredentialArgs: %v", err)
	}

	// Verify HostPath is bind-mounted directly (not via staging dir)
	found := false
	for _, a := range args {
		if a.Flag == "-v" && strings.Contains(a.Value, ghDir) {
			found = true
			if !strings.HasSuffix(a.Value, ":ro") {
				t.Errorf("host path mount should be read-only, got: %s", a.Value)
			}
			if strings.Contains(a.Value, "credentials") {
				t.Errorf("host path mount should not go through staging dir, got: %s", a.Value)
			}
		}
	}
	if !found {
		t.Fatal("host path credential mount not found")
	}
}

func TestKitCredentialArgsFuncError(t *testing.T) {
	home := t.TempDir()
	projectDir := t.TempDir()
	cname := ContainerName(projectDir)

	agentConfigDir := filepath.Join(home, ".asylum", "agents", "stub")
	os.MkdirAll(agentConfigDir, 0755)

	testKit := &kit.Kit{
		Name: "java/maven",
		CredentialFunc: func(opts kit.CredentialOpts) ([]kit.CredentialMount, error) {
			return nil, fmt.Errorf("test error")
		},
	}

	creds := &config.Credentials{Auto: true}
	opts := RunOpts{
		Config: config.Config{Kits: map[string]*config.KitConfig{
			"java": {Credentials: creds},
		}},
		Agent:      stubAgent{},
		ProjectDir: projectDir,
		Kits:       []*kit.Kit{testKit},
	}

	// Should not fail — error is logged as warning
	_, err := kitCredentialArgs(home, cname, opts)
	if err != nil {
		t.Fatalf("kitCredentialArgs should not fail on credential error: %v", err)
	}
}

func TestMavenCredentialsEndToEnd(t *testing.T) {
	home := t.TempDir()
	projectDir := t.TempDir()
	t.Setenv("HOME", home)
	cname := ContainerName(projectDir)

	// Create fake ~/.m2/settings.xml with a "progress" server
	m2Dir := filepath.Join(home, ".m2")
	os.MkdirAll(m2Dir, 0755)
	os.WriteFile(filepath.Join(m2Dir, "settings.xml"), []byte(`<?xml version="1.0"?>
<settings>
  <servers>
    <server>
      <id>progress</id>
      <username>user</username>
      <password>pass</password>
    </server>
  </servers>
</settings>`), 0644)

	// Write .asylum.local with explicit credentials (simulating the user's config)
	os.WriteFile(filepath.Join(projectDir, ".asylum.local"), []byte(`kits:
  java:
    credentials:
      - progress
`), 0644)

	cfg, err := config.Load(projectDir, config.CLIFlags{}, "")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.KitCredentialMode("java") != "explicit" {
		t.Fatalf("KitCredentialMode = %q, want explicit", cfg.KitCredentialMode("java"))
	}

	// Use the real java/maven kit
	allKits, err := kit.Resolve(cfg.KitNames(), cfg.DisabledKits())
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	opts := RunOpts{
		Config:     cfg,
		Agent:      stubAgent{},
		ProjectDir: projectDir,
		Kits:       allKits,
	}

	args, err := kitCredentialArgs(home, cname, opts)
	if err != nil {
		t.Fatalf("kitCredentialArgs: %v", err)
	}

	credFound := false
	for _, a := range args {
		if a.Flag == "-v" && strings.Contains(a.Value, "settings.xml") {
			credFound = true
		}
	}
	if !credFound {
		t.Fatalf("no settings.xml credential mount in args: %v", args)
	}

	// Verify only "progress" is in the generated file, not other servers
	credFile := filepath.Join(home, ".asylum", "projects", cname, "credentials", "settings.xml")
	data, err := os.ReadFile(credFile)
	if err != nil {
		t.Fatalf("credential file not written: %v", err)
	}
	if !strings.Contains(string(data), "<id>progress</id>") {
		t.Errorf("expected progress server in credential file, got:\n%s", data)
	}
}

func TestCoreVolumesNodeModulesShadowed(t *testing.T) {
	home := t.TempDir()
	projectDir := t.TempDir()
	cname := ContainerName(projectDir)

	agentConfigDir := filepath.Join(home, ".asylum", "agents", "stub")
	os.MkdirAll(agentConfigDir, 0755)
	os.WriteFile(filepath.Join(projectDir, "package.json"), []byte("{}"), 0644)

	nm := filepath.Join(projectDir, "node_modules")
	os.MkdirAll(nm, 0755)

	opts := RunOpts{
		Config:     config.Config{},
		Agent:      stubAgent{},
		ProjectDir: projectDir,
	}

	args, err := coreVolumes(home, cname, opts)
	if err != nil {
		t.Fatalf("coreVolumes: %v", err)
	}

	// "node_modules" hashes to a fixed value
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte("node_modules")))[:11]
	wantMount := "type=volume,src=" + cname + "-npm-" + hash + ",dst=" + nm
	if !hasRunArg(args, "--mount", wantMount) {
		t.Errorf("expected RunArg{Flag:\"--mount\", Value:%q} in %v", wantMount, args)
	}
}

func TestCoreVolumesNodeModulesDisabled(t *testing.T) {
	home := t.TempDir()
	projectDir := t.TempDir()
	cname := ContainerName(projectDir)

	agentConfigDir := filepath.Join(home, ".asylum", "agents", "stub")
	os.MkdirAll(agentConfigDir, 0755)
	os.WriteFile(filepath.Join(projectDir, "package.json"), []byte("{}"), 0644)
	os.MkdirAll(filepath.Join(projectDir, "node_modules"), 0755)

	opts := RunOpts{
		Config:     config.Config{Kits: map[string]*config.KitConfig{"node": {ShadowNodeModules: boolPtr(false)}}},
		Agent:      stubAgent{},
		ProjectDir: projectDir,
	}

	args, err := coreVolumes(home, cname, opts)
	if err != nil {
		t.Fatalf("coreVolumes: %v", err)
	}

	for _, a := range args {
		if strings.Contains(a.Value, "node_modules") && strings.Contains(a.Value, "type=volume") {
			t.Errorf("node_modules shadow should be disabled, found %q", a.Value)
		}
	}
}

func TestEnsureAgentConfig(t *testing.T) {
	home := t.TempDir()
	a := stubAgent{
		asylumConfigDir: filepath.Join(home, ".asylum", "agents", "test"),
		nativeConfigDir: "",
	}

	// First call: creates directory, returns true (seeded)
	seeded, err := EnsureAgentConfig(home, a)
	if err != nil {
		t.Fatal(err)
	}
	if !seeded {
		t.Error("expected seeded=true on first call")
	}

	// Second call: directory exists, returns false
	seeded, err = EnsureAgentConfig(home, a)
	if err != nil {
		t.Fatal(err)
	}
	if seeded {
		t.Error("expected seeded=false on second call")
	}
}

func TestEnsureAgentConfigSeedsFromNative(t *testing.T) {
	home := t.TempDir()

	// Create a native config dir with a file
	nativeDir := filepath.Join(home, ".test-agent")
	os.MkdirAll(nativeDir, 0755)
	os.WriteFile(filepath.Join(nativeDir, "config.json"), []byte(`{"key":"val"}`), 0644)

	a := stubAgent{
		asylumConfigDir: filepath.Join(home, ".asylum", "agents", "test"),
		nativeConfigDir: nativeDir,
	}

	seeded, err := EnsureAgentConfig(home, a)
	if err != nil {
		t.Fatal(err)
	}
	if !seeded {
		t.Error("expected seeded=true")
	}

	// Verify file was copied
	data, err := os.ReadFile(filepath.Join(home, ".asylum", "agents", "test", "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"key":"val"}` {
		t.Errorf("copied content = %q", string(data))
	}
}

func TestGenerateSandboxRules(t *testing.T) {
	home := t.TempDir()
	cname := "asylum-rules-test"

	kits := []*kit.Kit{
		{Name: "github", Tools: []string{"gh"}},
		{Name: "java", RulesSnippet: "### Java\nJDK 17/21/25 via mise.\n"},
		{Name: "java/maven", Tools: []string{"mvn"}},
		{Name: "node", RulesSnippet: "### Node.js\nLTS via fnm.\n"},
		{Name: "python"}, // no snippet, no tools
	}

	dir, err := generateSandboxRules(home, cname, kits, "1.2.3", nil)
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "asylum-sandbox.md"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	// Version
	if !strings.Contains(content, "v1.2.3") {
		t.Error("missing version")
	}

	// Core content
	if !strings.Contains(content, "Asylum Docker container") {
		t.Error("missing core sandbox identity")
	}
	if !strings.Contains(content, "host.docker.internal") {
		t.Error("missing host connectivity info")
	}

	// Reference link
	if !strings.Contains(content, "asylum-reference.md") {
		t.Error("missing reference doc link")
	}

	// Kit tools
	if !strings.Contains(content, "gh (github)") {
		t.Error("missing gh in kit tools")
	}
	if !strings.Contains(content, "mvn (java/maven)") {
		t.Error("missing mvn in kit tools")
	}

	// Kit snippets with blank line separation
	if !strings.Contains(content, "### Java") {
		t.Error("missing java kit snippet")
	}
	if !strings.Contains(content, "### Node.js") {
		t.Error("missing node kit snippet")
	}
	if !strings.Contains(content, "JDK 17/21/25 via mise.\n\n### Node.js") {
		t.Error("missing blank line between kit snippets")
	}

	// Disabled kits section
	if !strings.Contains(content, "## Disabled Kits") {
		t.Error("missing disabled kits section")
	}
	if !strings.Contains(content, "asylum-reference.md") {
		t.Error("disabled kits section should reference asylum-reference.md")
	}

	// Reference doc written
	refData, err := os.ReadFile(filepath.Join(dir, "asylum-reference.md"))
	if err != nil {
		t.Fatal("reference doc not written")
	}
	if !strings.Contains(string(refData), "Asylum Reference") {
		t.Error("reference doc has unexpected content")
	}
}

func TestGenerateSandboxRules_NoKits(t *testing.T) {
	home := t.TempDir()
	cname := "asylum-rules-nokits"

	dir, err := generateSandboxRules(home, cname, nil, "dev", nil)
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "asylum-sandbox.md"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	if !strings.Contains(content, "Asylum Docker container") {
		t.Error("missing core content")
	}
	if strings.Contains(content, "## Kit Tools") {
		t.Error("should not have Kit Tools section with no kits")
	}
	if strings.Contains(content, "## Active Kits") {
		t.Error("should not have Active Kits section with no kit snippets")
	}
	// All kits should appear as disabled when none are active
	if !strings.Contains(content, "## Disabled Kits") {
		t.Error("should have Disabled Kits section listing available kits")
	}
}

func TestRunArgsSandboxRulesMount(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	projectDir := t.TempDir()
	agentConfigDir := filepath.Join(home, ".asylum", "agents", "claude")
	os.MkdirAll(agentConfigDir, 0755)

	a := stubAgent{
		asylumConfigDir: agentConfigDir,
	}
	opts := RunOpts{
		Config:     config.Config{},
		Agent:      claudeStubAgent{stubAgent: a},
		ImageTag:   "asylum:test",
		ProjectDir: projectDir,
		Kits:       []*kit.Kit{{Name: "java", RulesSnippet: "### Java\n"}},
		Version:    "1.0.0",
	}

	args, _, _, err := RunArgs(opts)
	if err != nil {
		t.Fatal(err)
	}

	// Mount targets use ContainerConfigDir() (/home/stub/.stub for claudeStubAgent)
	rulesTarget := "/home/stub/.stub/rules/asylum-sandbox.md"
	refTarget := "/home/stub/.stub/asylum-reference.md"

	foundRules, foundRef := false, false
	for _, arg := range args {
		if strings.Contains(arg, rulesTarget) && strings.HasSuffix(arg, ":ro") {
			foundRules = true
		}
		if strings.Contains(arg, refTarget) && strings.HasSuffix(arg, ":ro") {
			foundRef = true
		}
	}
	if !foundRules {
		t.Errorf("sandbox rules mount not found in args: %v", args)
	}
	if !foundRef {
		t.Errorf("reference doc mount not found in args: %v", args)
	}

	// Mountpoint files must be pre-created in the host config dir so runc
	// doesn't need to create them through a VirtioFS-backed bind mount.
	if !fileExists(filepath.Join(agentConfigDir, "rules", "asylum-sandbox.md")) {
		t.Error("mountpoint file not pre-created: rules/asylum-sandbox.md")
	}
	if !fileExists(filepath.Join(agentConfigDir, "asylum-reference.md")) {
		t.Error("mountpoint file not pre-created: asylum-reference.md")
	}
}

func TestGenerateSandboxRules_WithPorts(t *testing.T) {
	home := t.TempDir()
	cname := "asylum-ports-test"

	dir, err := generateSandboxRules(home, cname, nil, "dev", []int{10000, 10001, 10002})
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "asylum-sandbox.md"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	if !strings.Contains(content, "## Forwarded Ports") {
		t.Error("missing Forwarded Ports section")
	}
	if !strings.Contains(content, "- 10000") {
		t.Error("missing port 10000")
	}
	if !strings.Contains(content, "- 10002") {
		t.Error("missing port 10002")
	}
	if !strings.Contains(content, "http://localhost:<port>") {
		t.Error("missing localhost instructions")
	}
}

func TestGenerateSandboxRules_WithoutPorts(t *testing.T) {
	home := t.TempDir()
	cname := "asylum-noports-test"

	dir, err := generateSandboxRules(home, cname, nil, "dev", nil)
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "asylum-sandbox.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "Forwarded Ports") {
		t.Error("should not have Forwarded Ports section when no ports allocated")
	}
}


