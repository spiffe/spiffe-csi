// +build !linux

package main

import "errors"

func bindMountRO(src, dst string) error {
	return errors.New("unsupported on this platform")
}

func unmount(path string) error {
	return errors.New("unsupported on this platform")
}
