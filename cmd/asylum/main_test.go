package main

import (
	"reflect"
	"testing"

	"github.com/inventage-ai/asylum/internal/container"
)

func TestResolveMode(t *testing.T) {
	tests := []struct {
		name     string
		subcmd   string
		admin    bool
		wantMode container.Mode
	}{
		{"default agent", "", false, container.ModeAgent},
		{"shell", "shell", false, container.ModeShell},
		{"shell admin", "shell", true, container.ModeAdminShell},
		{"run", "run", false, container.ModeCommand},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveMode(tt.subcmd, tt.admin); got != tt.wantMode {
				t.Errorf("resolveMode(%q, %v) = %d, want %d", tt.subcmd, tt.admin, got, tt.wantMode)
			}
		})
	}
}

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantFlags cliFlags
		wantSub   string
		wantExtra []string
		wantErr   bool
	}{
		{
			name:    "no args",
			args:    nil,
			wantSub: "",
		},
		{
			name:      "version flag",
			args:      []string{"--version"},
			wantFlags: cliFlags{Version: true},
		},
		{
			name:      "help flag",
			args:      []string{"-h"},
			wantFlags: cliFlags{Help: true},
		},
		{
			name:      "agent long form",
			args:      []string{"-a", "gemini"},
			wantFlags: cliFlags{Agent: "gemini"},
		},
		{
			name:      "agent shorthand",
			args:      []string{"-acodex"},
			wantFlags: cliFlags{Agent: "codex"},
		},
		{
			name:      "port single",
			args:      []string{"-p", "8080"},
			wantFlags: cliFlags{Ports: []string{"8080"}},
		},
		{
			name:      "port repeatable",
			args:      []string{"-p", "3000", "-p", "4000"},
			wantFlags: cliFlags{Ports: []string{"3000", "4000"}},
		},
		{
			name:      "volume",
			args:      []string{"-v", "~/data:/data"},
			wantFlags: cliFlags{Volumes: []string{"~/data:/data"}},
		},
		{
			name:      "java",
			args:      []string{"--java", "21"},
			wantFlags: cliFlags{Java: "21"},
		},
		{
			name:      "new flag",
			args:      []string{"-n"},
			wantFlags: cliFlags{New: true},
		},
		{
			name:      "rebuild flag",
			args:      []string{"--rebuild"},
			wantFlags: cliFlags{Rebuild: true},
		},
		{
			name:      "cleanup flag",
			args:      []string{"--cleanup"},
			wantFlags: cliFlags{Cleanup: true},
		},

		// -e env var
		{
			name:      "env var single",
			args:      []string{"-e", "FOO=bar"},
			wantFlags: cliFlags{Env: map[string]string{"FOO": "bar"}},
		},
		{
			name:      "env var repeatable last wins",
			args:      []string{"-e", "K=1", "-e", "K=2"},
			wantFlags: cliFlags{Env: map[string]string{"K": "2"}},
		},
		{
			name:      "env var multiple keys",
			args:      []string{"-e", "A=1", "-e", "B=2"},
			wantFlags: cliFlags{Env: map[string]string{"A": "1", "B": "2"}},
		},
		{
			name:      "env var empty value allowed",
			args:      []string{"-e", "K="},
			wantFlags: cliFlags{Env: map[string]string{"K": ""}},
		},
		{
			name:    "env var missing equals",
			args:    []string{"-e", "NOEQUALS"},
			wantErr: true,
		},
		{
			name:    "env var empty key",
			args:    []string{"-e", "=value"},
			wantErr: true,
		},

		// version subcommand
		{
			name:    "version command",
			args:    []string{"version"},
			wantSub: "version",
		},
		{
			name:      "version command with short",
			args:      []string{"version", "--short"},
			wantSub:   "version",
			wantFlags: cliFlags{Short: true},
		},
		{
			name:    "version command unknown flag errors",
			args:    []string{"version", "--verbose"},
			wantErr: true,
		},

		// cleanup subcommand
		{
			name:    "cleanup command",
			args:    []string{"cleanup"},
			wantSub: "cleanup",
		},
		{
			name:      "cleanup command with all flag",
			args:      []string{"cleanup", "--all"},
			wantSub:   "cleanup",
			wantFlags: cliFlags{All: true},
		},
		{
			name:    "cleanup command unknown flag errors",
			args:    []string{"cleanup", "--unknown"},
			wantErr: true,
		},
		{
			name:    "cleanup command extra arg errors",
			args:    []string{"cleanup", "extra"},
			wantErr: true,
		},
		{
			name:      "cleanup flag alias with all",
			args:      []string{"--cleanup", "--all"},
			wantFlags: cliFlags{Cleanup: true, All: true},
		},

		// -- separator
		{
			name:      "double dash passes args to agent",
			args:      []string{"--", "fix", "the", "bug"},
			wantExtra: []string{"fix", "the", "bug"},
		},
		{
			name:      "flags before double dash",
			args:      []string{"-a", "gemini", "--", "--verbose"},
			wantFlags: cliFlags{Agent: "gemini"},
			wantExtra: []string{"--verbose"},
		},

		// shell subcommand
		{
			name:    "shell",
			args:    []string{"shell"},
			wantSub: "shell",
		},
		{
			name:      "shell admin",
			args:      []string{"shell", "--admin"},
			wantSub:   "shell",
			wantFlags: cliFlags{Admin: true},
		},
		{
			name:    "shell unknown flag errors",
			args:    []string{"shell", "--verbose"},
			wantErr: true,
		},
		{
			name:      "flags before shell",
			args:      []string{"-p", "8080", "shell"},
			wantSub:   "shell",
			wantFlags: cliFlags{Ports: []string{"8080"}},
		},

		// run subcommand
		{
			name:      "run command",
			args:      []string{"run", "python", "test.py", "-v"},
			wantSub:   "run",
			wantExtra: []string{"python", "test.py", "-v"},
		},
		{
			name:      "run with optional double dash",
			args:      []string{"run", "--", "python", "test.py"},
			wantSub:   "run",
			wantExtra: []string{"python", "test.py"},
		},
		{
			name:      "run with flags before it",
			args:      []string{"-p", "8080", "run", "ls"},
			wantSub:   "run",
			wantFlags: cliFlags{Ports: []string{"8080"}},
			wantExtra: []string{"ls"},
		},
		{
			name:    "run with no command errors",
			args:    []string{"run"},
			wantErr: true,
		},
		{
			name:    "run with only double dash errors",
			args:    []string{"run", "--"},
			wantErr: true,
		},

		// self-update subcommand
		{
			name:    "self-update",
			args:    []string{"self-update"},
			wantSub: "self-update",
		},
		{
			name:      "self-update with dev flag",
			args:      []string{"self-update", "--dev"},
			wantSub:   "self-update",
			wantFlags: cliFlags{Dev: true},
		},
		{
			name:    "self-update unknown flag errors",
			args:    []string{"self-update", "--verbose"},
			wantErr: true,
		},
		{
			name:    "selfupdate alias",
			args:    []string{"selfupdate"},
			wantSub: "self-update",
		},
		{
			name:      "selfupdate alias with dev flag",
			args:      []string{"selfupdate", "--dev"},
			wantSub:   "self-update",
			wantFlags: cliFlags{Dev: true},
		},
		{
			name:      "self-update with version",
			args:      []string{"self-update", "0.4.0"},
			wantSub:   "self-update",
			wantFlags: cliFlags{TargetVersion: "0.4.0"},
		},
		{
			name:      "self-update with v-prefixed version",
			args:      []string{"self-update", "v0.4.0"},
			wantSub:   "self-update",
			wantFlags: cliFlags{TargetVersion: "v0.4.0"},
		},
		{
			name:    "self-update dev and version conflict",
			args:    []string{"self-update", "--dev", "0.4.0"},
			wantErr: true,
		},
		{
			name:    "self-update safe and version conflict",
			args:    []string{"self-update", "--safe", "0.4.0"},
			wantErr: true,
		},
		{
			name:      "selfupdate alias with version",
			args:      []string{"selfupdate", "0.4.0"},
			wantSub:   "self-update",
			wantFlags: cliFlags{TargetVersion: "0.4.0"},
		},

		// --skip-onboarding flag
		{
			name:      "skip onboarding",
			args:      []string{"--skip-onboarding"},
			wantFlags: cliFlags{SkipOnboarding: true},
		},

		// --kits flag
		{
			name:      "kits flag",
			args:      []string{"--kits", "java,python"},
			wantFlags: cliFlags{Kits: &[]string{"java", "python"}},
		},

		// --agents flag
		{
			name:      "agents flag",
			args:      []string{"--agents", "claude,gemini"},
			wantFlags: cliFlags{Agents: &[]string{"claude", "gemini"}},
		},
		{
			name:      "agents single",
			args:      []string{"--agents", "claude"},
			wantFlags: cliFlags{Agents: &[]string{"claude"}},
		},

		// --worktree / -w passthrough
		{
			name:      "worktree with name",
			args:      []string{"--worktree", "feat-x"},
			wantExtra: []string{"--worktree", "feat-x"},
		},
		{
			name:      "worktree without name",
			args:      []string{"--worktree"},
			wantExtra: []string{"--worktree"},
		},
		{
			name:      "worktree short flag with name",
			args:      []string{"-w", "feat-x"},
			wantExtra: []string{"--worktree", "feat-x"},
		},
		{
			name:      "worktree short flag without name",
			args:      []string{"-w"},
			wantExtra: []string{"--worktree"},
		},
		{
			name:      "worktree does not consume next flag as name",
			args:      []string{"--worktree", "-n"},
			wantFlags: cliFlags{New: true},
			wantExtra: []string{"--worktree"},
		},

		// strict: unknown flags error
		{
			name:    "all without cleanup errors",
			args:    []string{"--all"},
			wantErr: true,
		},
		{
			name:    "unknown flag errors",
			args:    []string{"--bogus"},
			wantErr: true,
		},
		{
			name:    "unknown short flag errors",
			args:    []string{"-x"},
			wantErr: true,
		},
		{
			name:    "unexpected positional errors",
			args:    []string{"openspec"},
			wantErr: true,
		},

		// trailing flag with no value
		{
			name:    "trailing -a",
			args:    []string{"-a"},
			wantErr: true,
		},
		{
			name:    "trailing --agent",
			args:    []string{"--agent"},
			wantErr: true,
		},
		{
			name:    "trailing -p",
			args:    []string{"-p"},
			wantErr: true,
		},
		{
			name:    "trailing -v",
			args:    []string{"-v"},
			wantErr: true,
		},
		{
			name:    "trailing --java",
			args:    []string{"--java"},
			wantErr: true,
		},
		{
			name:    "trailing -e",
			args:    []string{"-e"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags, subcmd, extra, err := parseArgs(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if subcmd != tt.wantSub {
				t.Errorf("subcommand = %q, want %q", subcmd, tt.wantSub)
			}
			if !reflect.DeepEqual(extra, tt.wantExtra) {
				t.Errorf("extraArgs = %v, want %v", extra, tt.wantExtra)
			}
			if !reflect.DeepEqual(flags, tt.wantFlags) {
				t.Errorf("flags = %+v, want %+v", flags, tt.wantFlags)
			}
		})
	}
}
