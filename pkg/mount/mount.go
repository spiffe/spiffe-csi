package mount

// BindMountRO performs a read-only bind mount from src to dest
func BindMountRO(src, dst string) error {
	return bindMountRO(src, dst)
}

// Unmount unmounts a mount
func Unmount(path string) error {
	return unmount(path)
}
