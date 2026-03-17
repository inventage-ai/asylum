package main

import "testing"

func TestParseArgs_Version(t *testing.T) {
	flags, positional, passthrough := parseArgs([]string{"--version"})

	if !flags.Version {
		t.Error("expected Version flag to be true")
	}
	if len(positional) != 0 {
		t.Errorf("expected no positional args, got %v", positional)
	}
	if len(passthrough) != 0 {
		t.Errorf("expected no passthrough args, got %v", passthrough)
	}
}
