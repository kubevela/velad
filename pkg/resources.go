package pkg

import (
	"embed"
)

var (
	k3sBinaryLocation = "/usr/local/bin/k3s"
	k3sImageDir       = "/var/lib/rancher/k3s/agent/images/"
	k3sImageLocation  = "/var/lib/rancher/k3s/agent/images/k3s-airgap-images-amd64.tar.gz"
)

var (
	//go:embed static/k3s
	K3sDirectory embed.FS

	//go:embed static/vela/images
	VelaImages embed.FS
	//go:embed static/vela/charts
	VelaChart embed.FS

	//go:embed static/nginx
	Nginx embed.FS
)
