package mount

import (
	"fmt"
	"os"
	"strings"
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

// TestIsMountPointInReader_OctalEscape verifies that mount points containing
// whitespace match against their octal-escaped representation in mountinfo
// (e.g. "/mnt/has space" appears as "/mnt/has\040space" in field 5).
func TestIsMountPointInReader_OctalEscape(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		target    string
		wantMatch bool
	}{
		{
			name:      "single space",
			line:      `36 35 0:0 / /mnt/has\040space rw,relatime - tmpfs tmpfs rw`,
			target:    "/mnt/has space",
			wantMatch: true,
		},
		{
			name:      "non-consecutive spaces",
			line:      `36 35 0:0 / /mnt/a\040b\040c rw,relatime - tmpfs tmpfs rw`,
			target:    "/mnt/a b c",
			wantMatch: true,
		},
		{
			name:      "consecutive spaces",
			line:      `36 35 0:0 / /mnt/double\040\040space rw,relatime - tmpfs tmpfs rw`,
			target:    "/mnt/double  space",
			wantMatch: true,
		},
		{
			name:      "raw escaped form does not match",
			line:      `36 35 0:0 / /mnt/has\040space rw,relatime - tmpfs tmpfs rw`,
			target:    `/mnt/has\040space`,
			wantMatch: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := isMountPointInReader(strings.NewReader(tt.line+"\n"), tt.target)
			require.NoError(t, err)
			assert.Equal(t, tt.wantMatch, got)
		})
	}
}

func BenchmarkIsMountPoint(b *testing.B) {
	const total = 10_000
	const target = "/var/lib/kubelet/pods/pod-target/volumes/kubernetes.io~csi/spiffe/mount"

	cases := []struct {
		name      string
		targetIdx int // -1 means no line in the input matches.
	}{
		{"FirstMatch", 0},
		{"LastMatch", total - 1},
		{"NoMatch", -1},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			f, err := os.CreateTemp(b.TempDir(), "mountinfo")
			require.NoError(b, err)
			for i := 0; i < total; i++ {
				mp := fmt.Sprintf(
					"/var/lib/kubelet/pods/pod-%d/volumes/kubernetes.io~csi/spiffe/mount",
					i,
				)
				if i == tc.targetIdx {
					mp = target
				}
				_, err := fmt.Fprintf(f, "%d %d 0:%d / %s rw,relatime - tmpfs tmpfs rw\n",
					1000+i, 999, 100+i, mp)
				require.NoError(b, err)
			}
			require.NoError(b, f.Close())

			orig := procMountInfo
			procMountInfo = f.Name()
			b.Cleanup(func() { procMountInfo = orig })

			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				_, _ = IsMountPoint(target)
			}
		})
	}
}
