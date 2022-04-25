//go:build !linux
// +build !linux

package resources

import (
	"embed"
)

var (
	//go:embed static/k3d/images
	K3dImage embed.FS
)
