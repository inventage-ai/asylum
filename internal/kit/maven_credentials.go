package kit

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// pomProject is the minimal pom.xml structure needed to extract server IDs.
type pomProject struct {
	Repositories           []pomRepository `xml:"repositories>repository"`
	PluginRepositories     []pomRepository `xml:"pluginRepositories>pluginRepository"`
	DistributionRepository *pomRepository  `xml:"distributionManagement>repository"`
	DistributionSnapshot   *pomRepository  `xml:"distributionManagement>snapshotRepository"`
	Profiles               []pomProfile    `xml:"profiles>profile"`
}

type pomProfile struct {
	Repositories           []pomRepository `xml:"repositories>repository"`
	PluginRepositories     []pomRepository `xml:"pluginRepositories>pluginRepository"`
	DistributionRepository *pomRepository  `xml:"distributionManagement>repository"`
	DistributionSnapshot   *pomRepository  `xml:"distributionManagement>snapshotRepository"`
}

type pomRepository struct {
	ID string `xml:"id"`
}

// parsePomServerIDs reads a pom.xml and returns all referenced server IDs.
func parsePomServerIDs(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var pom pomProject
	if err := xml.Unmarshal(data, &pom); err != nil {
		return nil, fmt.Errorf("parse pom.xml: %w", err)
	}

	seen := map[string]bool{}
	var ids []string
	add := func(id string) {
		if id != "" && !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}

	collectRepos := func(repos []pomRepository, plugins []pomRepository, dist, snap *pomRepository) {
		for _, r := range repos {
			add(r.ID)
		}
		for _, r := range plugins {
			add(r.ID)
		}
		if dist != nil {
			add(dist.ID)
		}
		if snap != nil {
			add(snap.ID)
		}
	}

	collectRepos(pom.Repositories, pom.PluginRepositories, pom.DistributionRepository, pom.DistributionSnapshot)
	for _, p := range pom.Profiles {
		collectRepos(p.Repositories, p.PluginRepositories, p.DistributionRepository, p.DistributionSnapshot)
	}

	return ids, nil
}

// mavenSettings is the minimal settings.xml structure for server extraction.
type mavenSettings struct {
	XMLName xml.Name      `xml:"settings"`
	Servers []mavenServer `xml:"servers>server"`
}

type mavenServer struct {
	ID       string `xml:"id"`
	Username string `xml:"username,omitempty"`
	Password string `xml:"password,omitempty"`
	// Preserve any other content via raw inner XML
	Extra []mavenServerExtra `xml:",any"`
}

type mavenServerExtra struct {
	XMLName xml.Name
	Content string `xml:",innerxml"`
}

// parseSettingsServers reads a settings.xml and returns server entries keyed by ID.
func parseSettingsServers(path string) (map[string]mavenServer, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var settings mavenSettings
	if err := xml.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("parse settings.xml: %w", err)
	}

	servers := make(map[string]mavenServer, len(settings.Servers))
	for _, s := range settings.Servers {
		servers[s.ID] = s
	}
	return servers, nil
}

// generateFilteredSettings builds a minimal settings.xml with only the matching servers.
func generateFilteredSettings(requestedIDs []string, servers map[string]mavenServer) []byte {
	var b strings.Builder
	b.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	b.WriteString("<settings>\n")
	b.WriteString("  <servers>\n")

	for _, id := range requestedIDs {
		s, ok := servers[id]
		if !ok {
			fmt.Fprintf(&b, "    <!-- server %q referenced in pom.xml but not found in ~/.m2/settings.xml -->\n", id)
			continue
		}
		b.WriteString("    <server>\n")
		fmt.Fprintf(&b, "      <id>%s</id>\n", xmlEscape(s.ID))
		if s.Username != "" {
			fmt.Fprintf(&b, "      <username>%s</username>\n", xmlEscape(s.Username))
		}
		if s.Password != "" {
			fmt.Fprintf(&b, "      <password>%s</password>\n", xmlEscape(s.Password))
		}
		for _, extra := range s.Extra {
			fmt.Fprintf(&b, "      <%s>%s</%s>\n", extra.XMLName.Local, extra.Content, extra.XMLName.Local)
		}
		b.WriteString("    </server>\n")
	}

	b.WriteString("  </servers>\n")
	b.WriteString("</settings>\n")
	return []byte(b.String())
}

func xmlEscape(s string) string {
	var b strings.Builder
	if err := xml.EscapeText(&b, []byte(s)); err != nil {
		return s
	}
	return b.String()
}

// mavenCredentialFunc is the CredentialFunc for the java/maven sub-kit.
func mavenCredentialFunc(opts CredentialOpts) ([]CredentialMount, error) {
	var requestedIDs []string

	switch opts.Mode {
	case CredentialAuto:
		pomPath := filepath.Join(opts.ProjectDir, "pom.xml")
		ids, err := parsePomServerIDs(pomPath)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, nil
			}
			return nil, fmt.Errorf("read pom.xml: %w", err)
		}
		requestedIDs = ids
	case CredentialExplicit:
		requestedIDs = opts.Explicit
	default:
		return nil, nil
	}

	if len(requestedIDs) == 0 {
		return nil, nil
	}

	settingsPath := filepath.Join(opts.HomeDir, ".m2", "settings.xml")
	servers, err := parseSettingsServers(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read settings.xml: %w", err)
	}

	// Only generate if at least one server matches
	hasMatch := false
	for _, id := range requestedIDs {
		if _, ok := servers[id]; ok {
			hasMatch = true
			break
		}
	}
	if !hasMatch {
		return nil, nil
	}

	content := generateFilteredSettings(requestedIDs, servers)
	return []CredentialMount{
		{
			Content:     content,
			Destination: "~/.m2/settings.xml",
		},
	}, nil
}
