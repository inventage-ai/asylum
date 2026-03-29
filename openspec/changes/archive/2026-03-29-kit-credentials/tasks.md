## 1. Kit Credential Types and Config

- [x] 1.1 Add CredentialFunc, CredentialOpts, CredentialMount, and CredentialMode types to `internal/kit/kit.go`
- [x] 1.2 Add `Credentials` field to KitConfig in `internal/config/config.go` with custom YAML unmarshaling (supports string `"auto"`, bool `false`, and string list)
- [x] 1.3 Add helper methods on KitConfig to query credential mode and explicit IDs

## 2. Maven Credential Implementation

- [x] 2.1 Add pom.xml parsing to extract server IDs from repositories, pluginRepositories, distributionManagement, and profiles in `internal/kit/java.go`
- [x] 2.2 Add settings.xml parsing to extract server entries from `~/.m2/settings.xml`
- [x] 2.3 Implement settings.xml filtering: match server entries by ID, generate XML comments for missing IDs
- [x] 2.4 Implement the CredentialFunc on the java/maven sub-kit: auto mode (pom.xml discovery), explicit mode (provided IDs), and generated settings.xml output
- [x] 2.5 Write tests for pom.xml parsing (repositories, plugin repos, distribution management, profiles)
- [x] 2.6 Write tests for settings.xml filtering (matching servers, missing servers, no settings file, no matches)

## 3. Container Launch Integration

- [x] 3.1 Add credential generation and mount logic to `appendVolumes()` in `internal/container/container.go`: call CredentialFunc for active kits, write files to `~/.asylum/projects/<cname>/credentials/`, add bind mounts after cache volumes
- [x] 3.2 Pass credential config (mode and explicit IDs) through RunOpts to the container assembly
- [x] 3.3 Write tests for credential mount integration (ordering after cache volumes, error handling)

## 4. First-Run Onboarding

- [x] 4.1 Add a function to collect credential-capable kits (kits with non-nil CredentialFunc) with their display labels
- [x] 4.2 Replace the Y/n credential prompt in `internal/firstrun/firstrun.go` with a TUI multiselect of credential-capable kits
- [x] 4.3 Write selected kits' `credentials: auto` to `~/.asylum/config.yaml` using yaml.Node manipulation (similar to `SetAgentIsolation`)
- [x] 4.4 Update firstrun tests for the new multiselect flow

## 5. Cleanup

- [x] 5.1 Remove the hardcoded `credentials` list and `detectCredentials` function from `internal/firstrun/firstrun.go`
- [x] 5.2 Add changelog entry under Unreleased
