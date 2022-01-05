package version_test

import (
	"os"
	"strings"
	"testing"

	"github.com/spiffe/spiffe-csi/internal/version"
	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	versionData, err := os.ReadFile("VERSION")
	require.NoError(t, err)

	actual := version.Version()
	expectedPrefix := strings.TrimSpace(string(versionData))

	require.True(t, strings.HasPrefix(actual, expectedPrefix), "version %q should have prefix %q", actual, expectedPrefix)
}
