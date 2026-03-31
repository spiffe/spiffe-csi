// Package version provides the build version of the binary.
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

// Version returns the build version of the binary.
func Version() string {
	return version
}
