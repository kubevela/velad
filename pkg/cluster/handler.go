package cluster

import "github.com/oam-dev/velad/pkg/apis"

// Handler defines the interface for handling the cluster(k3d/k3s) management
type Handler interface {
	Install(args apis.InstallArgs) error
	Uninstall(name string) error
	GenKubeconfig(bindIP string) error
	SetKubeconfig() error
	LoadImage(image string) error
	GetStatus() apis.ClusterStatus
}
