package agent

import "testing"

func TestClaudeEnvVarsIDEHostOverride(t *testing.T) {
	env := (Claude{}).EnvVars()
	if got := env["CLAUDE_CODE_IDE_HOST_OVERRIDE"]; got != "host.docker.internal" {
		t.Fatalf("CLAUDE_CODE_IDE_HOST_OVERRIDE = %q, want %q", got, "host.docker.internal")
	}
}

func TestClaudeEnvVarsSandboxed(t *testing.T) {
	env := (Claude{}).EnvVars()
	if got := env["CLAUDE_CODE_SANDBOXED"]; got != "1" {
		t.Fatalf("CLAUDE_CODE_SANDBOXED = %q, want %q", got, "1")
	}
}
