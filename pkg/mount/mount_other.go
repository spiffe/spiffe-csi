//go:build !linux
// +build !linux

package mount

import (
	"errors"
)

func bindMountRO(src, dst string) error {
	return errors.New("unsupported on this platform")
}

func unmount(path string) error {
	return errors.New("unsupported on this platform")
}
