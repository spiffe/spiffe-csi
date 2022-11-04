package mount

// BindMountRW performs a read-write bind mount from root to mountPoint
func BindMountRW(root, mountPoint string) error {
	return bindMountRW(root, mountPoint)
}

// Unmount unmounts a mount
func Unmount(mountPoint string) error {
	return unmount(mountPoint)
}

// IsMountPoint returns whether or not the given mount point is valid.
func IsMountPoint(mountPoint string) (bool, error) {
	return isMountPoint(mountPoint)
}
