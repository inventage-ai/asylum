//go:build integration

package integration_test

import (
	"testing"

	"github.com/inventage-ai/asylum/internal/docker"
	"github.com/inventage-ai/asylum/internal/image"
)

func TestBaseImageBuild(t *testing.T) {
	ensureBaseImage(t)

	hash, err := docker.InspectLabel("asylum:latest", "asylum.hash")
	if err != nil {
		t.Fatalf("missing asylum.hash label: %v", err)
	}
	if hash == "" {
		t.Fatal("asylum.hash label is empty")
	}

	version, err := docker.InspectLabel("asylum:latest", "asylum.version")
	if err != nil {
		t.Fatalf("missing asylum.version label: %v", err)
	}
	if version == "" {
		t.Fatal("asylum.version label is empty")
	}
}

func TestBaseImageCaching(t *testing.T) {
	ensureBaseImage(t)

	rebuilt, err := image.EnsureBase(nil, nil, testVersion, false)
	if err != nil {
		t.Fatalf("second EnsureBase failed: %v", err)
	}
	if rebuilt {
		t.Fatal("expected rebuilt=false on second call")
	}
}

func TestProjectImageBuild(t *testing.T) {
	ensureBaseImage(t)

	packages := map[string][]string{"apt": {"tree"}}
	tag, err := image.EnsureProject(nil, packages, "", testVersion, false, false)
	if err != nil {
		t.Fatalf("EnsureProject failed: %v", err)
	}
	if tag == "asylum:latest" {
		t.Fatal("expected project-specific tag, got asylum:latest")
	}
	if len(tag) < len("asylum:proj-") {
		t.Fatalf("unexpected tag format: %s", tag)
	}
	t.Cleanup(func() { docker.RemoveImages(tag) })

	hash, err := docker.InspectLabel(tag, "asylum.packages.hash")
	if err != nil {
		t.Fatalf("missing asylum.packages.hash label: %v", err)
	}
	if hash == "" {
		t.Fatal("asylum.packages.hash label is empty")
	}
}

func TestProjectImageCaching(t *testing.T) {
	ensureBaseImage(t)

	packages := map[string][]string{"apt": {"tree"}}
	tag1, err := image.EnsureProject(nil, packages, "", testVersion, false, false)
	if err != nil {
		t.Fatalf("first EnsureProject failed: %v", err)
	}
	t.Cleanup(func() { docker.RemoveImages(tag1) })

	tag2, err := image.EnsureProject(nil, packages, "", testVersion, false, false)
	if err != nil {
		t.Fatalf("second EnsureProject failed: %v", err)
	}
	if tag1 != tag2 {
		t.Fatalf("tags differ: %s vs %s", tag1, tag2)
	}
}

func TestProjectImageNoPackages(t *testing.T) {
	ensureBaseImage(t)

	tag, err := image.EnsureProject(nil, nil, "", testVersion, false, false)
	if err != nil {
		t.Fatalf("EnsureProject failed: %v", err)
	}
	if tag != "asylum:latest" {
		t.Fatalf("expected asylum:latest, got %s", tag)
	}
}
