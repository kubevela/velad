package apis

import "github.com/oam-dev/kubevela/references/cli"

// InstallArgs defines arguments for velad install command
type InstallArgs struct {
	BindIP      string
	DBEndpoint  string
	ClusterOnly bool
	Token       string
	Controllers string
	// InstallArgs is parameters passed to vela install command
	InstallArgs cli.InstallArgs
	Name        string
}

// UninstallArgs defines arguments for velad uninstall command
type UninstallArgs struct {
	Name string
}

// KubeconfigArgs defines arguments for velad kubeconfig command
type KubeconfigArgs struct {
	Internal bool
	External bool
	Host     bool
	Name     string
}

type TokenArgs struct {
	Name string
}

// LoadBalancerArgs defines arguments for load balancer command
type LoadBalancerArgs struct {
	Hosts         []string
	Configuration string
}

// ControlPlaneStatus defines the status of control plane
type ControlPlaneStatus struct {
	Clusters []ClusterStatus
	Vela     VelaStatus
}

// ClusterStatus defines the status of cluster, including k3s/k3d
type ClusterStatus struct {
	// K3dImages only works for non-linux
	K3dImages
	K3s K3sStatus
	K3d K3dStatus
}

// K3sStatus defines the status of k3s
type K3sStatus struct {
	K3sBinary        bool
	K3sServiceStatus string
	VelaStatus       string
	Reason           string
}

// K3dStatus defines the status of k3d
type K3dStatus struct {
	Reason       string
	K3dContainer []K3dContainer
}

// K3dContainer defines the status of one k3d cluster
type K3dContainer struct {
	Name       string
	Running    bool
	VelaStatus string
	Reason     string
}

// K3dImages defines the status of k3d images
type K3dImages struct {
	K3s      bool
	K3dTools bool
	K3dProxy bool
	Reason   string
}

// VelaStatus is the status of vela in host machine
type VelaStatus struct {
	VelaUXAddonDirPresent bool
	VelaUXAddonDirPath    string
	VelaCLIInstalled      bool
	VelaCLIPath           string
	Reason                string
}

var (
	K3sTokenLoc = "/var/lib/rancher/k3s/server/token"
	// K3sKubeConfigLocation is default path of k3s kubeconfig
	K3sKubeConfigLocation = "/etc/rancher/k3s/k3s.yaml"
	// K3sExternalKubeConfigLocation is where to generate kubeconfig for external access
	K3sExternalKubeConfigLocation = "/etc/rancher/k3s/k3s-external.yaml"
	// VelaLinkPos is path to save vela symlink in linux/macos
	VelaLinkPos = "/usr/local/bin/vela"
	// VelaDDockerNetwork is docker network for k3d cluster when `velad install`
	// all cluster will be created in this network, so they can communicate with each other
	VelaDDockerNetwork = "k3d-velad"

	// K3dImageK3s is k3s image tag
	K3dImageK3s = "rancher/k3s:v1.21.10-k3s1"
	// K3dImageTools is k3d tools image tag
	K3dImageTools = "rancher/k3d-tools:5.2.2"
	// K3dImageProxy is k3d proxy image tag
	K3dImageProxy = "rancher/k3d-proxy:5.2.2"

	// KubeVelaHelmRelease is helm release name for vela
	KubeVelaHelmRelease = "kubevela"
	// StatusVelaNotInstalled is status for kubevela helm chart not installed
	StatusVelaNotInstalled = "not installed"
	// StatusVelaDeployed is success status for kubevela helm chart deployed
	StatusVelaDeployed = "deployed"
)
