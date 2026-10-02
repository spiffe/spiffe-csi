package mount

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

const (
	msBind uintptr = 4096  // LINUX MS_BIND
	msRec  uintptr = 16384 // LINUX MS_REC
)

var (
	// procMountInfo is the path the mount information presented by the proc
	// filesystem for the current process. It is overridden in unit tests to
	// test the parsing.
	procMountInfo = "/proc/self/mountinfo"
)

// mountPointIdx is the slice index of the mount point in a parsed mountinfo
// record. proc(5) "/proc/[pid]/mountinfo" documents it as field 5.
const mountPointIdx = 4

func bindMountRW(root, mountPoint string) error {
	return unix.Mount(root, mountPoint, "none", msBind, "")
}

// bindMountRecursiveRW binds root to mountPoint along with anything mounted
// beneath it. A plain bind copies only the directory, so a filesystem already
// mounted under root at the time of the bind is not carried across, and one
// mounted afterwards never appears.
func bindMountRecursiveRW(root, mountPoint string) error {
	return unix.Mount(root, mountPoint, "none", msBind|msRec, "")
}

func unmount(mountPoint string) error {
	return unix.Unmount(mountPoint, 0)
}

// unmountDetach detaches mountPoint and anything mounted beneath it. A plain
// unmount of a mount that has children fails with EBUSY.
func unmountDetach(mountPoint string) error {
	return unix.Unmount(mountPoint, unix.MNT_DETACH)
}

func isMountPoint(mountPoint string) (bool, error) {
	f, err := os.Open(procMountInfo)
	if err != nil {
		return false, fmt.Errorf("unable to open mount info: %w", err)
	}
	defer func() { _ = f.Close() }()
	return isMountPointInReader(f, mountPoint)
}

func hasChildMounts(mountPoint string) (bool, error) {
	f, err := os.Open(procMountInfo)
	if err != nil {
		return false, fmt.Errorf("unable to open mount info: %w", err)
	}
	defer func() { _ = f.Close() }()
	return hasChildMountsInReader(f, mountPoint)
}

// isMountPointInReader scans mountinfo-formatted records from r and reports
// whether any record's mount point (field 5) equals mountPoint. It returns on
// the first match so the per-call working set is independent of the host's
// total mount count.
func isMountPointInReader(r io.Reader, mountPoint string) (bool, error) {
	return anyMountPointInReader(r, func(mp string) bool { return mp == mountPoint })
}

// hasChildMountsInReader reports whether any record in r is mounted beneath
// mountPoint, i.e. whether mountPoint has children that would make a plain
// unmount of it fail with EBUSY.
func hasChildMountsInReader(r io.Reader, mountPoint string) (bool, error) {
	prefix := strings.TrimSuffix(mountPoint, "/") + "/"
	return anyMountPointInReader(r, func(mp string) bool { return strings.HasPrefix(mp, prefix) })
}

// anyMountPointInReader scans mountinfo-formatted records from r and reports
// whether match returns true for any record's unescaped mount point.
func anyMountPointInReader(r io.Reader, match func(string) bool) (bool, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) <= mountPointIdx {
			continue
		}
		if match(unescapeOctal(fields[mountPointIdx])) {
			return true, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("failed to scan mount info: %w", err)
	}
	return false, nil
}

var reOctal = regexp.MustCompile(`\\([0-7]{3})`)

func unescapeOctal(s string) string {
	return reOctal.ReplaceAllStringFunc(s, func(oct string) string {
		// cannot fail due to regex constraints
		r, _ := strconv.ParseUint(oct[1:], 8, 8)
		return string([]byte{byte(r)})
	})
}
