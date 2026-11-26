# Spec Delta

## Purpose

Makes every named Docker volume asylum creates traceable to the project it belongs to and to what it holds, so tooling and users can decide whether a volume is still needed without decoding its name.

## ADDED Requirements

### Requirement: Volumes are labeled with their project
Every named volume asylum creates SHALL carry the label `asylum.project` whose value is the absolute project directory, exactly as used to derive the container name.

#### Scenario: Cache volume label
- **WHEN** a container is started for project `/home/u/code/app` with the node kit active
- **THEN** the volume `<container>-cache-npm` is created with label `asylum.project=/home/u/code/app`

#### Scenario: Project path containing a comma
- **WHEN** the project directory is `/home/u/code/a,b`
- **THEN** the volume is created with label `asylum.project=/home/u/code/a,b` and the container starts normally

### Requirement: Volumes are labeled with their kind
Every named volume asylum creates SHALL carry the label `asylum.kind` with value `cache` for package cache volumes and `node_modules` for shadowed `node_modules` volumes.

#### Scenario: Cache kind
- **WHEN** the volume `<container>-cache-maven` is created
- **THEN** it carries `asylum.kind=cache`

#### Scenario: node_modules kind
- **WHEN** the volume `<container>-npm-<hash>` is created
- **THEN** it carries `asylum.kind=node_modules`

### Requirement: Unlabeled volumes keep working
Volumes created before labeling existed SHALL continue to be mounted and used unchanged; asylum SHALL NOT recreate them to add labels.

#### Scenario: Existing unlabeled volume
- **WHEN** `<container>-cache-npm` already exists without labels and the container starts
- **THEN** the existing volume is mounted with its contents intact
