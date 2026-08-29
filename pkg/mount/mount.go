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
// UnmountDetach must be used to undo it; see that function.
func BindMountRecursiveRW(root, mountPoint string) error {
	return bindMountRecursiveRW(root, mountPoint)
}

// Unmount unmounts a mount
func Unmount(mountPoint string) error {
	return unmount(mountPoint)
}

// UnmountDetach detaches a mount and anything mounted beneath it. It is the
// counterpart of BindMountRecursiveRW: a recursive bind can leave the mount with
// children, and unmounting a mount that has children fails with EBUSY.
func UnmountDetach(mountPoint string) error {
	return unmountDetach(mountPoint)
}

// IsMountPoint returns whether or not the given mount point is valid.
func IsMountPoint(mountPoint string) (bool, error) {
	return isMountPoint(mountPoint)
}
