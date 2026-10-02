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

func TestHasChildMounts(t *testing.T) {
	for _, tt := range []struct {
		desc       string
		mountPoint string
		want       bool
	}{
		{desc: "mount with children", mountPoint: "/var/lib/kubelet/pods", want: true},
		{desc: "trailing slash", mountPoint: "/var/lib/kubelet/pods/", want: true},
		{desc: "mount without children", mountPoint: "/var/lib/kubelet/pods/c3a32fc0-f186-4974-8579-429dea58ec6d/volumes/kubernetes.io~csi/spire-agent-socket/mount", want: false},
		{desc: "sibling sharing a name prefix", mountPoint: "/var/lib/kubelet/po", want: false},
	} {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := HasChildMounts(tt.mountPoint)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHasChildMountsInReader_OctalEscape(t *testing.T) {
	const line = `36 35 0:0 / /mnt/has\040space/child rw,relatime - tmpfs tmpfs rw`
	got, err := hasChildMountsInReader(strings.NewReader(line+"\n"), "/mnt/has space")
	require.NoError(t, err)
	assert.True(t, got)
}

func TestIsSharedMountInReader(t *testing.T) {
	for _, tt := range []struct {
		desc string
		path string
		want bool
	}{
		{desc: "private mount point", path: "/spire-agent-socket", want: false},
		{desc: "beneath a private mount", path: "/spire-agent-socket/sub", want: false},
		{desc: "shared and slave mount point", path: "/var/lib/kubelet/pods", want: true},
		{desc: "beneath a shared mount", path: "/var/lib/kubelet/pods/not-a-mount", want: true},
		{desc: "falls back to the root mount", path: "/etc/hosts-dir", want: false},
		{desc: "unclean path", path: "/var/lib/kubelet/pods/../pods/", want: true},
	} {
		t.Run(tt.desc, func(t *testing.T) {
			f, err := os.Open(procMountInfo)
			require.NoError(t, err)
			defer f.Close()

			got, err := isSharedMountInReader(f, tt.path)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsSharedMountInReader_Propagation(t *testing.T) {
	for _, tt := range []struct {
		desc  string
		lines []string
		want  bool
	}{
		{
			desc:  "slave only",
			lines: []string{"1 0 0:1 / /sock rw master:7 - tmpfs t rw"},
			want:  false,
		},
		{
			desc:  "shared",
			lines: []string{"1 0 0:1 / /sock rw shared:7 - tmpfs t rw"},
			want:  true,
		},
		{
			desc:  "shared field after the separator is not propagation",
			lines: []string{"1 0 0:1 / /sock rw - tmpfs shared:7 rw"},
			want:  false,
		},
		{
			desc: "last mount at the same point is on top",
			lines: []string{
				"1 0 0:1 / /sock rw shared:7 - tmpfs t rw",
				"2 1 0:2 / /sock rw master:7 - tmpfs t rw",
			},
			want: false,
		},
		{
			desc: "deeper mount wins over its parent",
			lines: []string{
				"1 0 0:1 / / rw shared:1 - tmpfs t rw",
				"2 1 0:2 / /sock rw - tmpfs t rw",
			},
			want: false,
		},
		{
			desc:  "sibling sharing a name prefix does not contain the path",
			lines: []string{"1 0 0:1 / / rw - tmpfs t rw", "2 1 0:2 / /so rw shared:7 - tmpfs t rw"},
			want:  false,
		},
	} {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := isSharedMountInReader(strings.NewReader(strings.Join(tt.lines, "\n")+"\n"), "/sock")
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsSlaveMountInReader(t *testing.T) {
	for _, tt := range []struct {
		desc  string
		lines []string
		want  bool
	}{
		{
			desc:  "slave",
			lines: []string{"1 0 0:1 / /sock rw master:7 - tmpfs t rw"},
			want:  true,
		},
		{
			desc:  "shared and slave",
			lines: []string{"1 0 0:1 / /sock rw shared:8 master:7 - tmpfs t rw"},
			want:  true,
		},
		{
			desc:  "shared only",
			lines: []string{"1 0 0:1 / /sock rw shared:7 - tmpfs t rw"},
			want:  false,
		},
		{
			desc:  "private",
			lines: []string{"1 0 0:1 / /sock rw - tmpfs t rw"},
			want:  false,
		},
		{
			desc:  "master field after the separator is not propagation",
			lines: []string{"1 0 0:1 / /sock rw - tmpfs master:7 rw"},
			want:  false,
		},
		{
			desc: "the mount on top decides",
			lines: []string{
				"1 0 0:1 / /sock rw master:7 - tmpfs t rw",
				"2 1 0:2 / /sock rw - tmpfs t rw",
			},
			want: false,
		},
	} {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := isSlaveMountInReader(strings.NewReader(strings.Join(tt.lines, "\n")+"\n"), "/sock")
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsSlaveMountInReaderFixture(t *testing.T) {
	f, err := os.Open(procMountInfo)
	require.NoError(t, err)
	defer f.Close()

	// shared:195 master:28 in the fixture.
	got, err := isSlaveMountInReader(f, "/var/lib/kubelet/pods")
	require.NoError(t, err)
	assert.True(t, got)
}

func TestIsSharedMountInReader_NoContainingMount(t *testing.T) {
	_, err := isSharedMountInReader(strings.NewReader("1 0 0:1 / /other rw - tmpfs t rw\n"), "/sock")
	require.ErrorContains(t, err, "no mount contains")
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
