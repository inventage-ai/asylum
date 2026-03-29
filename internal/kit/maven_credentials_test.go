package kit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParsePomServerIDs(t *testing.T) {
	tests := []struct {
		name    string
		pom     string
		want    []string
		wantErr bool
	}{
		{
			name: "repositories",
			pom: `<project>
  <repositories>
    <repository><id>central</id></repository>
    <repository><id>nexus-releases</id></repository>
  </repositories>
</project>`,
			want: []string{"central", "nexus-releases"},
		},
		{
			name: "plugin repositories",
			pom: `<project>
  <pluginRepositories>
    <pluginRepository><id>plugin-repo</id></pluginRepository>
  </pluginRepositories>
</project>`,
			want: []string{"plugin-repo"},
		},
		{
			name: "distribution management",
			pom: `<project>
  <distributionManagement>
    <repository><id>releases</id></repository>
    <snapshotRepository><id>snapshots</id></snapshotRepository>
  </distributionManagement>
</project>`,
			want: []string{"releases", "snapshots"},
		},
		{
			name: "profiles",
			pom: `<project>
  <profiles>
    <profile>
      <repositories>
        <repository><id>profile-repo</id></repository>
      </repositories>
      <pluginRepositories>
        <pluginRepository><id>profile-plugin</id></pluginRepository>
      </pluginRepositories>
      <distributionManagement>
        <repository><id>profile-releases</id></repository>
      </distributionManagement>
    </profile>
  </profiles>
</project>`,
			want: []string{"profile-repo", "profile-plugin", "profile-releases"},
		},
		{
			name: "deduplication",
			pom: `<project>
  <repositories>
    <repository><id>nexus</id></repository>
  </repositories>
  <profiles>
    <profile>
      <repositories>
        <repository><id>nexus</id></repository>
      </repositories>
    </profile>
  </profiles>
</project>`,
			want: []string{"nexus"},
		},
		{
			name: "empty project",
			pom:  `<project></project>`,
			want: nil,
		},
		{
			name: "combined",
			pom: `<project>
  <repositories>
    <repository><id>repo1</id></repository>
  </repositories>
  <pluginRepositories>
    <pluginRepository><id>plugin1</id></pluginRepository>
  </pluginRepositories>
  <distributionManagement>
    <repository><id>dist1</id></repository>
    <snapshotRepository><id>snap1</id></snapshotRepository>
  </distributionManagement>
</project>`,
			want: []string{"repo1", "plugin1", "dist1", "snap1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "pom.xml")
			if err := os.WriteFile(path, []byte(tt.pom), 0644); err != nil {
				t.Fatal(err)
			}

			got, err := parsePomServerIDs(path)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parsePomServerIDs() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("got[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestParsePomServerIDs_NoFile(t *testing.T) {
	_, err := parsePomServerIDs("/nonexistent/pom.xml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestParseSettingsServers(t *testing.T) {
	settings := `<settings>
  <servers>
    <server>
      <id>nexus</id>
      <username>deploy</username>
      <password>secret</password>
    </server>
    <server>
      <id>github</id>
      <username>user</username>
      <password>token123</password>
    </server>
  </servers>
</settings>`

	dir := t.TempDir()
	path := filepath.Join(dir, "settings.xml")
	if err := os.WriteFile(path, []byte(settings), 0644); err != nil {
		t.Fatal(err)
	}

	servers, err := parseSettingsServers(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(servers) != 2 {
		t.Fatalf("got %d servers, want 2", len(servers))
	}
	if servers["nexus"].Username != "deploy" {
		t.Errorf("nexus username = %q, want %q", servers["nexus"].Username, "deploy")
	}
	if servers["github"].Password != "token123" {
		t.Errorf("github password = %q, want %q", servers["github"].Password, "token123")
	}
}

func TestGenerateFilteredSettings(t *testing.T) {
	servers := map[string]mavenServer{
		"nexus":  {ID: "nexus", Username: "deploy", Password: "secret"},
		"github": {ID: "github", Username: "user", Password: "token"},
	}

	t.Run("matching servers", func(t *testing.T) {
		got := string(generateFilteredSettings([]string{"nexus"}, servers))
		if !strings.Contains(got, "<id>nexus</id>") {
			t.Error("expected nexus server in output")
		}
		if strings.Contains(got, "<id>github</id>") {
			t.Error("unexpected github server in output")
		}
		if !strings.Contains(got, "<username>deploy</username>") {
			t.Error("expected username in output")
		}
	})

	t.Run("missing server adds comment", func(t *testing.T) {
		got := string(generateFilteredSettings([]string{"unknown"}, servers))
		if !strings.Contains(got, `<!-- server "unknown" referenced in pom.xml but not found in ~/.m2/settings.xml -->`) {
			t.Errorf("expected comment for missing server, got:\n%s", got)
		}
	})

	t.Run("mixed found and missing", func(t *testing.T) {
		got := string(generateFilteredSettings([]string{"nexus", "missing"}, servers))
		if !strings.Contains(got, "<id>nexus</id>") {
			t.Error("expected nexus server")
		}
		if !strings.Contains(got, `<!-- server "missing"`) {
			t.Error("expected comment for missing server")
		}
	})

	t.Run("empty servers map", func(t *testing.T) {
		got := string(generateFilteredSettings([]string{"any"}, map[string]mavenServer{}))
		if !strings.Contains(got, "<!-- server") {
			t.Error("expected comment for missing server")
		}
	})
}

func TestMavenCredentialFunc(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()

	settingsXML := `<settings>
  <servers>
    <server>
      <id>nexus-releases</id>
      <username>deploy</username>
      <password>s3cret</password>
    </server>
    <server>
      <id>nexus-snapshots</id>
      <username>deploy</username>
      <password>snap-pass</password>
    </server>
  </servers>
</settings>`
	m2Dir := filepath.Join(home, ".m2")
	os.MkdirAll(m2Dir, 0755)
	os.WriteFile(filepath.Join(m2Dir, "settings.xml"), []byte(settingsXML), 0644)

	pomXML := `<project>
  <repositories>
    <repository><id>nexus-releases</id></repository>
  </repositories>
</project>`
	os.WriteFile(filepath.Join(project, "pom.xml"), []byte(pomXML), 0644)

	t.Run("auto mode", func(t *testing.T) {
		mounts, err := mavenCredentialFunc(CredentialOpts{
			ProjectDir: project,
			HomeDir:    home,
			Mode:       CredentialAuto,
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(mounts) != 1 {
			t.Fatalf("got %d mounts, want 1", len(mounts))
		}
		content := string(mounts[0].Content)
		if !strings.Contains(content, "<id>nexus-releases</id>") {
			t.Error("expected nexus-releases in output")
		}
		if strings.Contains(content, "<id>nexus-snapshots</id>") {
			t.Error("unexpected nexus-snapshots in output")
		}
		if mounts[0].Destination != "~/.m2/settings.xml" {
			t.Errorf("destination = %q, want ~/.m2/settings.xml", mounts[0].Destination)
		}
	})

	t.Run("explicit mode", func(t *testing.T) {
		mounts, err := mavenCredentialFunc(CredentialOpts{
			ProjectDir: project,
			HomeDir:    home,
			Mode:       CredentialExplicit,
			Explicit:   []string{"nexus-snapshots"},
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(mounts) != 1 {
			t.Fatalf("got %d mounts, want 1", len(mounts))
		}
		content := string(mounts[0].Content)
		if !strings.Contains(content, "<id>nexus-snapshots</id>") {
			t.Error("expected nexus-snapshots in output")
		}
		if strings.Contains(content, "<id>nexus-releases</id>") {
			t.Error("unexpected nexus-releases in output")
		}
	})

	t.Run("no pom.xml in auto mode", func(t *testing.T) {
		emptyDir := t.TempDir()
		mounts, err := mavenCredentialFunc(CredentialOpts{
			ProjectDir: emptyDir,
			HomeDir:    home,
			Mode:       CredentialAuto,
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(mounts) != 0 {
			t.Errorf("got %d mounts, want 0", len(mounts))
		}
	})

	t.Run("no settings.xml", func(t *testing.T) {
		emptyHome := t.TempDir()
		mounts, err := mavenCredentialFunc(CredentialOpts{
			ProjectDir: project,
			HomeDir:    emptyHome,
			Mode:       CredentialAuto,
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(mounts) != 0 {
			t.Errorf("got %d mounts, want 0", len(mounts))
		}
	})

	t.Run("no matching servers", func(t *testing.T) {
		noMatchProject := t.TempDir()
		os.WriteFile(filepath.Join(noMatchProject, "pom.xml"), []byte(`<project>
  <repositories>
    <repository><id>unknown-repo</id></repository>
  </repositories>
</project>`), 0644)

		mounts, err := mavenCredentialFunc(CredentialOpts{
			ProjectDir: noMatchProject,
			HomeDir:    home,
			Mode:       CredentialAuto,
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(mounts) != 0 {
			t.Errorf("got %d mounts, want 0", len(mounts))
		}
	})

	t.Run("off mode", func(t *testing.T) {
		mounts, err := mavenCredentialFunc(CredentialOpts{
			ProjectDir: project,
			HomeDir:    home,
			Mode:       CredentialOff,
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(mounts) != 0 {
			t.Errorf("got %d mounts, want 0", len(mounts))
		}
	})
}
