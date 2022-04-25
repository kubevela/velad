package resources

import (
	"embed"
)

var (
	K3sBinaryLocation = "/usr/local/bin/k3s"
	K3sImageDir       = "/var/lib/rancher/k3s/agent/images/"
	K3sImageLocation  = "/var/lib/rancher/k3s/agent/images/k3s-airgap-images-amd64.tar.gz"
)

var (
	//go:embed static/k3s/images
	K3sImage embed.FS

	//go:embed static/vela/images
	VelaImages embed.FS
	//go:embed static/vela/charts
	VelaChart embed.FS

	//go:embed static/nginx
	Nginx embed.FS

	//go:embed static/vela/addons
	VelaAddons embed.FS
)
