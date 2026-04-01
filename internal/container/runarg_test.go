package container

import (
	"strings"
	"testing"

	"github.com/inventage-ai/asylum/internal/kit"
)

func TestDedupKey(t *testing.T) {
	tests := []struct {
		name    string
		arg     kit.RunArg
		wantCat string
		wantKey string
	}{
		{"port with mapping", kit.RunArg{Flag: "-p", Value: "8080:3000"}, "-p", "3000"},
		{"port without mapping", kit.RunArg{Flag: "-p", Value: "3000"}, "-p", "3000"},
		{"volume", kit.RunArg{Flag: "-v", Value: "/host:/container:ro"}, "-v", "/container"},
		{"volume no options", kit.RunArg{Flag: "-v", Value: "/host:/container"}, "-v", "/container"},
		{"mount", kit.RunArg{Flag: "--mount", Value: "type=volume,src=my-vol,dst=/data"}, "--mount", "/data"},
		{"env var", kit.RunArg{Flag: "-e", Value: "FOO=bar"}, "-e", "FOO"},
		{"boolean flag", kit.RunArg{Flag: "--privileged", Value: ""}, "bool", "--privileged"},
		{"boolean rm", kit.RunArg{Flag: "--rm", Value: ""}, "bool", "--rm"},
		{"cap-add", kit.RunArg{Flag: "--cap-add", Value: "SYS_ADMIN"}, "--cap-add", "SYS_ADMIN"},
		{"single-value name", kit.RunArg{Flag: "--name", Value: "my-ctr"}, "single", "--name"},
		{"single-value hostname", kit.RunArg{Flag: "--hostname", Value: "myhost"}, "single", "--hostname"},
		{"single-value workdir", kit.RunArg{Flag: "-w", Value: "/workspace"}, "single", "-w"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cat, key := dedupKey(tt.arg)
			if cat != tt.wantCat || key != tt.wantKey {
				t.Errorf("dedupKey(%v) = (%q, %q), want (%q, %q)", tt.arg, cat, key, tt.wantCat, tt.wantKey)
			}
		})
	}
}

func TestResolveArgsHigherPriorityWins(t *testing.T) {
	args := []kit.RunArg{
		{Flag: "-p", Value: "8080:3000", Source: "ports kit", Priority: kit.PriorityKit},
		{Flag: "-p", Value: "3000:3000", Source: "user config (ports)", Priority: kit.PriorityConfig},
	}
	resolved, overrides, err := ResolveArgs(args)
	if err != nil {
		t.Fatal(err)
	}
	if len(resolved) != 1 {
		t.Fatalf("expected 1 resolved arg, got %d", len(resolved))
	}
	if resolved[0].Source != "user config (ports)" {
		t.Errorf("expected config to win, got source %q", resolved[0].Source)
	}
	if len(overrides) != 1 {
		t.Fatalf("expected 1 override, got %d", len(overrides))
	}
	if overrides[0].Replaced.Source != "ports kit" {
		t.Errorf("expected ports kit to be replaced, got %q", overrides[0].Replaced.Source)
	}
}

func TestResolveArgsSamePriorityConflict(t *testing.T) {
	args := []kit.RunArg{
		{Flag: "-v", Value: "/a:/workspace:ro", Source: "kit A", Priority: kit.PriorityKit},
		{Flag: "-v", Value: "/b:/workspace:rw", Source: "kit B", Priority: kit.PriorityKit},
	}
	_, _, err := ResolveArgs(args)
	if err == nil {
		t.Fatal("expected conflict error")
	}
	if !strings.Contains(err.Error(), "kit A") || !strings.Contains(err.Error(), "kit B") {
		t.Errorf("error should name both sources: %v", err)
	}
}

func TestResolveArgsSamePrioritySameValueOK(t *testing.T) {
	args := []kit.RunArg{
		{Flag: "--privileged", Value: "", Source: "kit A", Priority: kit.PriorityKit},
		{Flag: "--privileged", Value: "", Source: "kit B", Priority: kit.PriorityKit},
	}
	resolved, _, err := ResolveArgs(args)
	if err != nil {
		t.Fatal(err)
	}
	if len(resolved) != 1 {
		t.Fatalf("expected 1 resolved arg, got %d", len(resolved))
	}
}

func TestResolveArgsNoConflictDifferentKeys(t *testing.T) {
	args := []kit.RunArg{
		{Flag: "-p", Value: "3000:3000", Source: "user config", Priority: kit.PriorityConfig},
		{Flag: "-p", Value: "10000:10000", Source: "ports kit", Priority: kit.PriorityKit},
	}
	resolved, _, err := ResolveArgs(args)
	if err != nil {
		t.Fatal(err)
	}
	if len(resolved) != 2 {
		t.Fatalf("expected 2 resolved args, got %d", len(resolved))
	}
}

func TestResolveArgsDeterministicOrder(t *testing.T) {
	args := []kit.RunArg{
		{Flag: "-e", Value: "Z=1", Source: "core", Priority: kit.PriorityCore},
		{Flag: "-e", Value: "A=2", Source: "docker kit", Priority: kit.PriorityKit},
		{Flag: "-p", Value: "3000:3000", Source: "user config (ports)", Priority: kit.PriorityConfig},
	}
	resolved, _, err := ResolveArgs(args)
	if err != nil {
		t.Fatal(err)
	}
	if len(resolved) != 3 {
		t.Fatalf("expected 3 resolved args, got %d", len(resolved))
	}
	if resolved[0].Priority != kit.PriorityCore {
		t.Errorf("first arg should be core priority, got %d", resolved[0].Priority)
	}
	if resolved[1].Priority != kit.PriorityKit {
		t.Errorf("second arg should be kit priority, got %d", resolved[1].Priority)
	}
	if resolved[2].Priority != kit.PriorityConfig {
		t.Errorf("third arg should be config priority, got %d", resolved[2].Priority)
	}
}

func TestFlattenArgs(t *testing.T) {
	args := []kit.RunArg{
		{Flag: "--rm", Value: ""},
		{Flag: "-p", Value: "3000:3000"},
		{Flag: "-e", Value: "FOO=bar"},
		{Flag: "--privileged", Value: ""},
	}
	flat := FlattenArgs(args)
	want := []string{"--rm", "-p", "3000:3000", "-e", "FOO=bar", "--privileged"}
	if len(flat) != len(want) {
		t.Fatalf("got %v, want %v", flat, want)
	}
	for i := range flat {
		if flat[i] != want[i] {
			t.Errorf("flat[%d] = %q, want %q", i, flat[i], want[i])
		}
	}
}

func TestResolveArgsLowerPriorityOverridden(t *testing.T) {
	args := []kit.RunArg{
		{Flag: "-e", Value: "FOO=high", Source: "config", Priority: kit.PriorityConfig},
		{Flag: "-e", Value: "FOO=low", Source: "core", Priority: kit.PriorityCore},
	}
	resolved, overrides, err := ResolveArgs(args)
	if err != nil {
		t.Fatal(err)
	}
	if len(resolved) != 1 {
		t.Fatalf("expected 1 resolved arg, got %d", len(resolved))
	}
	if resolved[0].Value != "FOO=high" {
		t.Errorf("expected high-priority value, got %q", resolved[0].Value)
	}
	if len(overrides) != 1 {
		t.Fatalf("expected 1 override, got %d", len(overrides))
	}
}

func TestResolveArgsSameSourceOrdering(t *testing.T) {
	args := []kit.RunArg{
		{Flag: "-e", Value: "Z=1", Source: "core", Priority: kit.PriorityCore},
		{Flag: "-e", Value: "A=2", Source: "core", Priority: kit.PriorityCore},
		{Flag: "-p", Value: "3000:3000", Source: "core", Priority: kit.PriorityCore},
	}
	resolved, _, err := ResolveArgs(args)
	if err != nil {
		t.Fatal(err)
	}
	if len(resolved) != 3 {
		t.Fatalf("expected 3 resolved args, got %d", len(resolved))
	}
	// All same source and priority — sorted by dedup category+key
	// -e:A < -e:Z < -p:3000
	if resolved[0].Value != "A=2" {
		t.Errorf("first arg should be A=2, got %q", resolved[0].Value)
	}
	if resolved[1].Value != "Z=1" {
		t.Errorf("second arg should be Z=1, got %q", resolved[1].Value)
	}
	if resolved[2].Value != "3000:3000" {
		t.Errorf("third arg should be 3000:3000, got %q", resolved[2].Value)
	}
}

func TestFormatDebugBasic(t *testing.T) {
	resolved := []kit.RunArg{
		{Flag: "--privileged", Value: "", Source: "docker kit", Priority: kit.PriorityKit},
		{Flag: "-p", Value: "10000:10000", Source: "ports kit", Priority: kit.PriorityKit},
		{Flag: "-e", Value: "FOO=bar", Source: "core", Priority: kit.PriorityCore},
	}
	out := FormatDebug(resolved, nil)

	if !strings.Contains(out, "Docker run arguments:") {
		t.Error("missing header")
	}
	if !strings.Contains(out, "--privileged") || !strings.Contains(out, "docker kit") {
		t.Error("missing --privileged line with source")
	}
	if !strings.Contains(out, "-p 10000:10000") || !strings.Contains(out, "ports kit") {
		t.Error("missing port line with source")
	}
	if !strings.Contains(out, "-e FOO=bar") || !strings.Contains(out, "core") {
		t.Error("missing env line with source")
	}
	// No overrides section when nil
	if strings.Contains(out, "Overrides") {
		t.Error("should not show overrides section when none exist")
	}
}

func TestFormatDebugWithOverrides(t *testing.T) {
	resolved := []kit.RunArg{
		{Flag: "-p", Value: "3000:3000", Source: "user config (ports)", Priority: kit.PriorityConfig},
	}
	overrides := []kit.Override{
		{
			Replaced: kit.RunArg{Flag: "-p", Value: "8080:3000", Source: "ports kit", Priority: kit.PriorityKit},
			Winner:   kit.RunArg{Flag: "-p", Value: "3000:3000", Source: "user config (ports)", Priority: kit.PriorityConfig},
		},
	}
	out := FormatDebug(resolved, overrides)

	if !strings.Contains(out, "Overrides (higher priority won):") {
		t.Error("missing overrides section")
	}
	if !strings.Contains(out, "ports kit") || !strings.Contains(out, "user config (ports)") {
		t.Error("overrides section should name both sources")
	}
}

func TestVolumeAndMountDifferentNamespace(t *testing.T) {
	args := []kit.RunArg{
		{Flag: "-v", Value: "/host:/data:ro", Source: "core", Priority: kit.PriorityCore},
		{Flag: "--mount", Value: "type=volume,src=vol,dst=/data", Source: "kit", Priority: kit.PriorityKit},
	}
	resolved, _, err := ResolveArgs(args)
	if err != nil {
		t.Fatal(err)
	}
	if len(resolved) != 2 {
		t.Fatalf("expected 2 resolved args, got %d", len(resolved))
	}
}
