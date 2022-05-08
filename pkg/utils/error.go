package utils

// UnsupportedOS will panic with a message about the unsupported OS
func UnsupportedOS(os string) {
	panic("unsupported os: " + os)
}
