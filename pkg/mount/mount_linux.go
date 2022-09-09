package mount

import (
	"bufio"
	"fmt"
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

func bindMountRW(root, mountPoint string) error {
	return unix.Mount(root, mountPoint, "none", msBind, "")
}

func unmount(mountPoint string) error {
	return unix.Unmount(mountPoint, 0)
}

func isMountPoint(mountPoint string) (bool, error) {
	mounts, err := enumerateMounts()
	if err != nil {
		return false, fmt.Errorf("failed to enumerate bind mounts: %w", err)
	}
	for _, mount := range mounts {
		if mount.MountPoint == mountPoint {
			return true, nil
		}
	}
	return false, nil
}

func enumerateMounts() ([]mountInfo, error) {
	f, err := os.Open(procMountInfo)
	if err != nil {
		return nil, fmt.Errorf("unable to open mount info: %w", err)
	}
	defer f.Close()

	var mounts []mountInfo
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		mount, err := parseMountInfo(scanner.Text())
		if err != nil {
			continue
		}
		mounts = append(mounts, mount)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan mount info: %w", err)
	}
	return mounts, nil
}

type mountInfo struct {
	MountID    string
	ParentID   string
	DevID      string
	Root       string
	MountPoint string
}

func parseMountInfo(line string) (mountInfo, error) {
	const minColumns = 5
	fields := strings.Fields(line)
	if len(fields) < minColumns {
		return mountInfo{}, fmt.Errorf("mount info does not have at least %d columns", minColumns)
	}

	return mountInfo{
		MountID:    fields[0],
		ParentID:   fields[1],
		DevID:      fields[2],
		Root:       unescapeOctal(fields[3]),
		MountPoint: unescapeOctal(fields[4]),
	}, nil
}

var reOctal = regexp.MustCompile(`\\([0-7]{3})`)

func unescapeOctal(s string) string {
	return reOctal.ReplaceAllStringFunc(s, func(oct string) string {
		// cannot fail due to regex constraints
		r, _ := strconv.ParseUint(oct[1:], 8, 64)
		return string(rune(r))
	})
}
