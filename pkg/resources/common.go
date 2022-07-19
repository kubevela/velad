package resources

import (
	"embed"
)

var (
	// K3sBinaryLocation is where to save k3s binary
	K3sBinaryLocation = "/usr/local/bin/k3s"
	// K3sImageDir is the directory to save the k3s air-gap image
	K3sImageDir = "/var/lib/rancher/k3s/agent/images/"
	// K3sImageLocation is where to save k3s air-gap images
	K3sImageLocation = "/var/lib/rancher/k3s/agent/images/k3s-airgap-images.tar.gz"
)

var (
	//go:embed static/k3s/images
	// K3sImage see static/k3s/images
	K3sImage embed.FS

	//go:embed static/vela/images
	// VelaImages see static/vela/images
	VelaImages embed.FS
	//go:embed static/vela/charts
	// VelaChart see static/vela/charts
	VelaChart embed.FS

	//go:embed static/nginx
	// Nginx see static/nginx/
	Nginx embed.FS

	//go:embed static/vela/addons
	// VelaAddons see static/vela/addons/
	VelaAddons embed.FS
)
