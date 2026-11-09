//go:build !linux

package mount

import (
	"errors"
)

func bindMountRW(string, string) error {
	return errors.New("unsupported on this platform")
}

func bindMountRecursiveRW(string, string) error {
	return errors.New("unsupported on this platform")
}

func makeRSlave(string) error {
	return errors.New("unsupported on this platform")
}

func unmount(string) error {
	return errors.New("unsupported on this platform")
}

func unmountDetach(string) error {
	return errors.New("unsupported on this platform")
}

func isMountPoint(string) (bool, error) {
	return false, errors.New("unsupported on this platform")
}

func hasChildMounts(string) (bool, error) {
	return false, errors.New("unsupported on this platform")
}

func isSharedMount(string) (bool, error) {
	return false, errors.New("unsupported on this platform")
}

func isSlaveMount(string) (bool, error) {
	return false, errors.New("unsupported on this platform")
}
