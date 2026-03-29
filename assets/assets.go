package assets

import _ "embed"

//go:embed Dockerfile.core
var DockerfileCore []byte

//go:embed Dockerfile.tail
var DockerfileTail []byte

//go:embed entrypoint.core
var EntrypointCore []byte

//go:embed entrypoint.tail
var EntrypointTail []byte

//go:embed asylum-reference.md
var AsylumReference []byte
