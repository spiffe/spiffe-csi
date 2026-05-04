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
	msBind uintptr = 4096 // LINUX MS_BIND
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

func unmount(mountPoint string) error {
	return unix.Unmount(mountPoint, 0)
}

func isMountPoint(mountPoint string) (bool, error) {
	f, err := os.Open(procMountInfo)
	if err != nil {
		return false, fmt.Errorf("unable to open mount info: %w", err)
	}
	defer func() { _ = f.Close() }()
	return isMountPointInReader(f, mountPoint)
}

// isMountPointInReader scans mountinfo-formatted records from r and reports
// whether any record's mount point (field 5) equals mountPoint. It returns on
// the first match so the per-call working set is independent of the host's
// total mount count.
func isMountPointInReader(r io.Reader, mountPoint string) (bool, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) <= mountPointIdx {
			continue
		}
		if unescapeOctal(fields[mountPointIdx]) == mountPoint {
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
