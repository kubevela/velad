//go:build linux

package resources

import (
	"embed"
)

var (
	//go:embed static/k3s/other
	// K3sDirectory is the directory containing the k3s binary and install script
	K3sDirectory embed.FS
)
