# dropshare-kit Specification

## Purpose
Lets an agent inside the container upload a project file through Dropshare on the macOS host and receive the resulting share URL, so it can use the link directly, for example in a GitHub issue.

## Requirements

### Requirement: Opt-in macOS kit
The system SHALL provide a `dropshare` kit that is inactive unless the user enables it in project or user configuration. When active, the kit SHALL contribute a broker route, a container command named `asylum-dropshare`, a Claude skill delivered through the shared kit skill root, and a rules snippet. The kit SHALL only perform uploads when the host is macOS.

#### Scenario: Inactive without configuration
- **WHEN** a project has no `dropshare` entry in its configuration
- **THEN** the kit contributes no route, no command, no skill, and no rules snippet

#### Scenario: Active when enabled
- **WHEN** a project enables the `dropshare` kit
- **THEN** the broker serves the upload route, the container has `asylum-dropshare` on its PATH, and Claude discovers the `dropshare` skill

#### Scenario: Non-macOS host
- **WHEN** the kit is active and the host is not macOS
- **THEN** an upload request fails with a message saying the kit requires a macOS host, and nothing is started on the host

### Requirement: Upload returns the share URL
`asylum-dropshare <path>` SHALL upload the file at `<path>` through the host's Dropshare app using its default connection and SHALL print the resulting share URL on standard output, followed by a newline, and nothing else. It SHALL exit `0` on success and non-zero on any failure, with the reason on standard error.

#### Scenario: Successful upload
- **WHEN** the agent runs `asylum-dropshare ./shots/login.png` for a file inside the project directory and Dropshare finishes the upload
- **THEN** the command prints the share URL Dropshare reported and exits `0`

#### Scenario: Wrong number of arguments
- **WHEN** the command is run with no argument or more than one
- **THEN** it prints usage to standard error, exits non-zero, and sends no request

### Requirement: Only files visible in the container are uploaded
The system SHALL upload only what the container itself can see at the requested path, resolved inside the container's filesystem, symbolic links included. It SHALL hand Dropshare a copy that the container cannot modify, so the file uploaded is the file that was requested. The requested path SHALL be absolute and SHALL name a regular file. The request SHALL carry no value other than the path, so the container cannot set a display name, a connection, or a callback.

#### Scenario: File outside the project
- **WHEN** the requested path is a regular file visible in the container outside the project directory, for example in another mounted directory or in `/tmp`
- **THEN** the file is uploaded

#### Scenario: Symlink to a host-only path
- **WHEN** the requested path is a symbolic link whose target names a host file that is not mounted into the container, for example a file in `~/Documents`
- **THEN** the link resolves inside the container, and the host file is not uploaded

#### Scenario: File replaced after the request
- **WHEN** the container replaces the requested file, for example with a symbolic link to another file, after the request was made
- **THEN** Dropshare uploads the content the file had when it was requested

#### Scenario: Directory
- **WHEN** the requested path is a directory
- **THEN** the route rejects the request without copying the directory, and the command reports that only regular files can be uploaded

### Requirement: Upload runs in the background
The system SHALL hand the upload to Dropshare without bringing Dropshare to the foreground.

#### Scenario: Terminal keeps focus
- **WHEN** an upload starts while the user works in the terminal
- **THEN** Dropshare does not become the active application

### Requirement: Each callback resolves only its own request
Every upload request SHALL carry a fresh unguessable identifier in its callback URL. A callback SHALL complete only the request whose identifier it carries. Concurrent uploads, from one container or several, SHALL each receive their own share URL.

#### Scenario: Concurrent uploads
- **WHEN** two uploads run at the same time, from the same project or from two projects
- **THEN** each command prints the URL of its own file

#### Scenario: Callback with an unknown identifier
- **WHEN** a callback arrives whose identifier no request is waiting for
- **THEN** no request completes with its URL

### Requirement: Uploads reach Dropshare one at a time
Dropshare drops the callback of an upload that starts while another upload is still transferring. The system SHALL therefore start an upload only when no other upload, from any project, holds Dropshare. An upload SHALL hold Dropshare until its callback arrives, or for at most 10 seconds plus 1 second per MiB of file size counted from when the upload is handed to Dropshare, whichever comes first. A holder that exits for any reason SHALL release Dropshare at once.

#### Scenario: Simultaneous uploads take turns
- **WHEN** two uploads are requested at the same moment, from the same project or from two projects
- **THEN** the second is handed to Dropshare only after the first one's callback arrives, and both commands print their own URL

#### Scenario: Upload without a callback
- **WHEN** Dropshare never calls back for an upload of a 300 KiB file
- **THEN** other uploads wait for it at most 10 seconds

#### Scenario: Broker exits while holding Dropshare
- **WHEN** the broker holding Dropshare exits, for example because its container stops
- **THEN** the next waiting upload starts without delay

### Requirement: Callback values are parsed as Dropshare sends them
Dropshare percent-encodes spaces, `#`, `%`, and non-ASCII characters in the `file` value but leaves `&`, `+`, and `=` unencoded. The system SHALL recover the exact uploaded path and share URL from such a callback. The system SHALL accept a share URL only if it is an absolute `https` URL, and SHALL treat anything else as a failed upload.

#### Scenario: Filename with reserved characters
- **WHEN** the uploaded file is named `a & b#1+c=ü%20.txt`
- **THEN** the callback is matched to its request and the command prints the share URL

#### Scenario: Non-https URL in callback
- **WHEN** a callback carries a `url` value that is not an absolute `https` URL
- **THEN** the command fails and prints no URL

### Requirement: Bounded wait
The route SHALL wait at most five minutes in total, counting both time spent waiting for another upload and time spent waiting for the callback. When the wait ends without a callback, the command SHALL exit with a status distinct from other failures and SHALL report that the upload may still finish, in which case Dropshare puts the link on the host clipboard.

#### Scenario: No callback within the limit
- **WHEN** Dropshare does not call back within five minutes
- **THEN** the command exits with the timeout status and explains that the link, if the upload finishes, will be on the user's clipboard

#### Scenario: Dropshare busy for the whole limit
- **WHEN** other uploads keep Dropshare busy for five minutes
- **THEN** the command exits non-zero with status `1`, reports that nothing was uploaded, and Dropshare is never handed the file

#### Scenario: Container session ends mid-wait
- **WHEN** the requesting command is interrupted while the route waits
- **THEN** the route stops waiting and a later callback for that request completes nothing

### Requirement: One global callback applet
The system SHALL keep a single callback applet on the host at `~/.asylum/dropshare/`, registered for the `asylum-dropshare:` scheme, shared by all projects. The system SHALL build and register the applet on the first upload that finds it missing, and SHALL rebuild it when the applet definition shipped with Asylum changes. The applet SHALL show no Dock icon or window when it handles a callback, and SHALL accept only callbacks whose identifier has the expected form.

#### Scenario: First upload
- **WHEN** the first upload runs on a host without the applet
- **THEN** the system builds and registers the applet before starting the upload, and the upload completes

#### Scenario: Applet already current
- **WHEN** an upload runs and the installed applet matches the shipped definition
- **THEN** the system does not rebuild it

#### Scenario: Asylum ships a changed applet
- **WHEN** an upload runs after an Asylum update that changed the applet definition
- **THEN** the system rebuilds and re-registers the applet before starting the upload

#### Scenario: Several projects
- **WHEN** uploads run from several projects over time
- **THEN** LaunchServices holds exactly one registration of Asylum's for the `asylum-dropshare:` scheme

#### Scenario: Malformed identifier
- **WHEN** the applet receives an `asylum-dropshare:` URL whose identifier does not have the expected form
- **THEN** it writes nothing

### Requirement: Callback records do not accumulate
The system SHALL remove a callback record once its request consumes it, and a staged copy once its upload's callback arrives or its upload is rejected. It SHALL remove records and staged copies older than one hour that no request consumed.

#### Scenario: Consumed callback
- **WHEN** a request completes with a callback
- **THEN** that callback's record no longer exists on the host

#### Scenario: Orphaned callback
- **WHEN** a callback arrived for a request that had already timed out and an hour has passed
- **THEN** the next upload removes that record
