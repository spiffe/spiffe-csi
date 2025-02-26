package version

import (
	"runtime/debug"
)

var version string

func init() {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		panic("failed to read build information")
	}
	version = bi.Main.Version
}

func Version() string {
	return version
}
