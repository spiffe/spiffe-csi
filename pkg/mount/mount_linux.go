package mount

import "golang.org/x/sys/unix"

const (
	msRdOnly uintptr = 1    // LINUX MS_RDONLY
	msBind   uintptr = 4096 // LINUX MS_BIND
)

func bindMountRO(src, dst string) error {
	return unix.Mount(src, dst, "none", msBind|msRdOnly, "")
}

func unmount(path string) error {
	return unix.Unmount(path, 0)
}
