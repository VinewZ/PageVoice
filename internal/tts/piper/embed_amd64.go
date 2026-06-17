//go:build linux && amd64

package piper

import _ "embed"

//go:embed dist/piper_linux_x86_64.tar.gz
var archive []byte
