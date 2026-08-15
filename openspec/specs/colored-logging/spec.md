# colored-logging Specification

## Purpose

Gives asylum one place to write user-facing terminal output, with five levels (info, success, warn, error, build) rendered as colored prefixes. Using it instead of bare `fmt.Println` keeps output consistent and makes a failure visually distinct from progress chatter.

## Requirements

### Requirement: Five log levels with colored prefixes
The log package SHALL provide five functions: Info, Success, Warn, Error, Build. Each SHALL print a colored prefix followed by the formatted message and a newline.

#### Scenario: Info message
- **WHEN** `log.Info("starting %s", "build")` is called
- **THEN** it prints `i starting build` with the `i` in blue to stdout

#### Scenario: Success message
- **WHEN** `log.Success("done")` is called
- **THEN** it prints `ok done` with the `ok` in green to stdout

#### Scenario: Warn message
- **WHEN** `log.Warn("missing file")` is called
- **THEN** it prints `! missing file` with the `!` in yellow to stdout

#### Scenario: Error message
- **WHEN** `log.Error("failed: %v", err)` is called
- **THEN** it prints `x failed: <err>` with the `x` in red to stderr

#### Scenario: Build message
- **WHEN** `log.Build("compiling layer 3")` is called
- **THEN** it prints `# compiling layer 3` with the `#` in cyan to stdout
