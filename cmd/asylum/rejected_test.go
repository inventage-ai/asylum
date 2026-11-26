package main

import (
	"testing"

	"github.com/inventage-ai/asylum/internal/config"
)

func TestRejectedHash(t *testing.T) {
	vol := config.Rejection{File: ".asylum", Line: 5, Path: "volumes", Snippet: "volumes:\n  - ~/a:/a"}
	docker := config.Rejection{File: ".asylum", Line: 9, Path: "kits.docker", Snippet: "kits:\n  docker: {}"}
	base := rejectedHash([]config.Rejection{vol, docker})

	if rejectedHash([]config.Rejection{docker, vol}) != base {
		t.Error("hash depends on order")
	}
	moved := vol
	moved.Line = 12
	if rejectedHash([]config.Rejection{moved, docker}) != base {
		t.Error("hash depends on line numbers")
	}
	changed := vol
	changed.Snippet = "volumes:\n  - ~/b:/b"
	if rejectedHash([]config.Rejection{changed, docker}) == base {
		t.Error("hash ignores values")
	}
	extra := config.Rejection{File: ".asylum", Path: "volumes", Snippet: "volumes:\n  - ~/c:/c"}
	if rejectedHash([]config.Rejection{vol, docker, extra}) == base {
		t.Error("hash ignores an added entry")
	}
}
