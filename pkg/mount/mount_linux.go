package mount

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

const (
	msBind  uintptr = 4096   // LINUX MS_BIND
	msRec   uintptr = 16384  // LINUX MS_REC
	msSlave uintptr = 524288 // LINUX MS_SLAVE
)

var (
	// procMountInfo is the path the mount information presented by the proc
	// filesystem for the current process. It is overridden in unit tests to
	// test the parsing.
	procMountInfo = "/proc/self/mountinfo"
)

// Slice indices into a parsed mountinfo record. proc(5) "/proc/[pid]/mountinfo"
// documents the mount point as field 5, followed after field 6 by optional
// fields (such as "shared:N" and "master:N") up to a lone "-".
const (
	mountPointIdx    = 4
	optionalFieldIdx = 6
)

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

// makeRSlave makes mountPoint, and everything beneath it, a slave of the peer
// group it belonged to: it keeps receiving mounts and unmounts from there, and
// stops sending its own back. It changes only the calling mount namespace.
func makeRSlave(mountPoint string) error {
	return unix.Mount("", mountPoint, "", msSlave|msRec, "")
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

func isSharedMount(path string) (bool, error) {
	// mountinfo records canonical paths, and mounting follows symlinks.
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return false, fmt.Errorf("unable to resolve %q: %w", path, err)
	}
	f, err := os.Open(procMountInfo)
	if err != nil {
		return false, fmt.Errorf("unable to open mount info: %w", err)
	}
	defer func() { _ = f.Close() }()
	return isSharedMountInReader(f, resolved)
}

// mountInfo is a parsed mountinfo record.
type mountInfo struct {
	mountPoint string
	fields     []string
}

// isShared reports whether the record's optional fields mark it shared.
func (m mountInfo) isShared() bool {
	for i := optionalFieldIdx; i < len(m.fields) && m.fields[i] != "-"; i++ {
		if strings.HasPrefix(m.fields[i], "shared:") {
			return true
		}
	}
	return false
}

// isMountPointInReader scans mountinfo-formatted records from r and reports
// whether any record's mount point (field 5) equals mountPoint. It returns on
// the first match so the per-call working set is independent of the host's
// total mount count.
func isMountPointInReader(r io.Reader, mountPoint string) (bool, error) {
	found := false
	err := scanMountInfo(r, func(m mountInfo) bool {
		found = m.mountPoint == mountPoint
		return found
	})
	return found, err
}

// hasChildMountsInReader reports whether any record in r is mounted beneath
// mountPoint, i.e. whether mountPoint has children that would make a plain
// unmount of it fail with EBUSY.
func hasChildMountsInReader(r io.Reader, mountPoint string) (bool, error) {
	prefix := strings.TrimSuffix(mountPoint, "/") + "/"
	found := false
	err := scanMountInfo(r, func(m mountInfo) bool {
		found = strings.HasPrefix(m.mountPoint, prefix)
		return found
	})
	return found, err
}

// isSharedMountInReader reports whether the mount that path lives on, as
// recorded in r, is shared. The mount that path lives on is the one with the
// longest mount point containing it; of several at the same mount point, the
// last listed is on top.
func isSharedMountInReader(r io.Reader, path string) (bool, error) {
	path = filepath.Clean(path)
	bestLen, shared := -1, false
	err := scanMountInfo(r, func(m mountInfo) bool {
		if containsPath(m.mountPoint, path) && len(m.mountPoint) >= bestLen {
			bestLen, shared = len(m.mountPoint), m.isShared()
		}
		return false
	})
	if err != nil {
		return false, err
	}
	if bestLen < 0 {
		return false, fmt.Errorf("no mount contains %q", path)
	}
	return shared, nil
}

func containsPath(mountPoint, path string) bool {
	return path == mountPoint || mountPoint == "/" || strings.HasPrefix(path, mountPoint+"/")
}

// scanMountInfo calls fn with each mountinfo-formatted record in r, with the
// mount point unescaped, until fn returns true.
func scanMountInfo(r io.Reader, fn func(mountInfo) bool) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) <= mountPointIdx {
			continue
		}
		if fn(mountInfo{mountPoint: unescapeOctal(fields[mountPointIdx]), fields: fields}) {
			return nil
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to scan mount info: %w", err)
	}
	return nil
}

var reOctal = regexp.MustCompile(`\\([0-7]{3})`)

func unescapeOctal(s string) string {
	return reOctal.ReplaceAllStringFunc(s, func(oct string) string {
		// cannot fail due to regex constraints
		r, _ := strconv.ParseUint(oct[1:], 8, 8)
		return string([]byte{byte(r)})
	})
}
