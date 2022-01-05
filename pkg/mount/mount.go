package mount

// BindMountRO performs a read-only bind mount from root to mountPoint
func BindMountRO(root, mountPoint string) error {
	return bindMountRO(root, mountPoint)
}

// Unmount unmounts a mount
func Unmount(mountPoint string) error {
	return unmount(mountPoint)
}

// IsMountPoint returns whether or not the given mount point is valid.
func IsMountPoint(mountPoint string) (bool, error) {
	return isMountPoint(mountPoint)
}
