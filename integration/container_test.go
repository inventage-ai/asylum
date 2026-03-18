//go:build integration

package integration_test

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProjectDirectoryMount(t *testing.T) {
	ensureBaseImage(t)
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "hello.txt"), []byte("integration-test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Single container: verify both mount content and working directory
	out := dockerRunWithProjectDir(t, tmp, `echo "CONTENT=$(cat hello.txt)" && echo "PWD=$(pwd)"`)
	results := parseKeyValues(out)

	if v := results["CONTENT"]; v != "integration-test" {
		t.Errorf("file content = %q, want \"integration-test\"", v)
	}
	if v := results["PWD"]; v != tmp {
		t.Errorf("pwd = %q, want %q", v, tmp)
	}
}
