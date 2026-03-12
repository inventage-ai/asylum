package assets

import _ "embed"

//go:embed Dockerfile
var Dockerfile []byte

//go:embed entrypoint.sh
var Entrypoint []byte
