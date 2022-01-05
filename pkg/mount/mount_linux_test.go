package mount

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	procMountInfo = "testdata/mountinfo"
}

func TestIsMountPoint(t *testing.T) {
	const mountPoint = "/var/lib/kubelet/pods/c3a32fc0-f186-4974-8579-429dea58ec6d/volumes/kubernetes.io~csi/spire-agent-socket/mount"

	ok, err := IsMountPoint(mountPoint)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = IsMountPoint(mountPoint + "other")
	require.NoError(t, err)
	assert.False(t, ok)
}
