package handler

import "github.com/oam-dev/velad/pkg/apis"

// Handler defines the interface for handling the cluster(k3d/k3s) management
type Handler interface {
	Install(args apis.InstallArgs) error
	Uninstall() error
	GenKubeconfig(bindIP string) error
	PrintKubeConfig(internal, external bool)
	SetKubeconfig() error
	LoadImage(image string) error
}
