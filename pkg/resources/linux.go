//go:build linux
// +build linux

package resources

import (
	"embed"
)

var (
	//go:embed static/k3s/other
	K3sDirectory embed.FS
)
