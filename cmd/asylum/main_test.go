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

		// ssh-init subcommand
		{
			name:    "ssh-init",
			args:    []string{"ssh-init"},
			wantSub: "ssh-init",
		},
		{
			name:    "ssh-init extra arg errors",
			args:    []string{"ssh-init", "extra"},
			wantErr: true,
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

		// strict: unknown flags error
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
