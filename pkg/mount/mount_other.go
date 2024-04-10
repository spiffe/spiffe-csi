//go:build !linux
// +build !linux

package mount

import (
	"errors"
)

func bindMountRW(string, string) error {
	return errors.New("unsupported on this platform")
}

func unmount(string) error {
	return errors.New("unsupported on this platform")
}

func isMountPoint(string) (bool, error) {
	return false, errors.New("unsupported on this platform")
}
