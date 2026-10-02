// Package mount provides filesystem mount operations for the CSI driver.
package mount

// BindMountRW performs a read-write bind mount from root to mountPoint
func BindMountRW(root, mountPoint string) error {
	return bindMountRW(root, mountPoint)
}

// BindMountRecursiveRW performs a read-write bind mount from root to mountPoint,
// including anything mounted beneath root. Mounts made under root afterwards
// propagate to mountPoint when root is on a shared mount.
//
// The result can have child mounts, which Unmount refuses; see UnmountDetach.
func BindMountRecursiveRW(root, mountPoint string) error {
	return bindMountRecursiveRW(root, mountPoint)
}

// Unmount unmounts a mount
func Unmount(mountPoint string) error {
	return unmount(mountPoint)
}

// UnmountDetach detaches a mount and anything mounted beneath it. Use it for a
// mount that HasChildMounts reports children for, which Unmount refuses with
// EBUSY.
func UnmountDetach(mountPoint string) error {
	return unmountDetach(mountPoint)
}

// IsMountPoint returns whether or not the given mount point is valid.
func IsMountPoint(mountPoint string) (bool, error) {
	return isMountPoint(mountPoint)
}

// HasChildMounts returns whether anything is mounted beneath the given mount
// point.
func HasChildMounts(mountPoint string) (bool, error) {
	return hasChildMounts(mountPoint)
}
