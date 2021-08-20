package version

import (
	_ "embed"
	"fmt"
	"strings"
)

var (
	//go:embed VERSION
	baseVersion string

	// gitTag is set by the linker. If set, it must match baseVersion.
	gitTag string

	// gitCommit is the git commit. Set by the linker.
	gitCommit string

	// gitDirty is whether or not the git repo is dirty. Set by the linker.
	gitDirty string

	// version holds the final calculated version
	version string
)

func init() {
	baseVersion = strings.TrimSpace(baseVersion)
	gitTag = strings.TrimSpace(gitTag)
	gitCommit = strings.TrimSpace(gitCommit)
	version = baseVersion
	switch {
	case gitTag == "":
		// If this isn't a tagged build, then add -dev-<commit>
		// e.g. 0.1.0-dev-50f2eef
		version += "-dev"
		if gitCommit != "" {
			version += "-" + gitCommit
		}
	case gitTag != baseVersion:
		// If this is a tagged build, then the base version must match.
		panic(fmt.Errorf("mismatched version information: base=%q tag=%q", baseVersion, gitTag))
	default:
		version = gitTag
	}

	// If the repo is dirty, append "-dirty"
	if gitDirty != "" {
		version += "-dirty"
	}
}

func Version() string {
	return version
}
