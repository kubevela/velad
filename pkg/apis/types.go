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

type UninstallArgs struct {
	Name string
}

type KubeconfigArgs struct {
	Internal bool
	External bool
	Host     bool
	Name     string
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

type ClusterStatus struct {
	// K3dImages only works for non-linux
	K3dImages
	K3s              K3sStatus
	K3d              K3dStatus
	KubeconfigStatus KubeconfigStatus
}

type K3sStatus struct {
	K3sBinary         bool
	K3sServiceRunning bool
}

type K3dStatus struct {
	Reason       string
	K3dContainer []K3dContainer
}

type K3dContainer struct {
	Name       string
	Running    bool
	VelaStatus string
	Reason     string
}

type KubeconfigStatus struct {
	KubeconfigHostGenerated     bool
	KubeconfigExternalGenerated bool
	// KubeconfigInternalGenerated only works in non-linux
	KubeconfigInternalGenerated bool
}

type K3dImages struct {
	K3s      bool
	K3dTools bool
	K3dProxy bool
	Reason   string
}

type VelaStatus struct {
	VelaUXAddonDirPresent bool
	VelaUXAddonDirPath    string
	VelaCLIInstalled      bool
	VelaCLIPath           string
	Reason                string
}

var (
	K3sKubeConfigLocation         = "/etc/rancher/k3s/k3s.yaml"
	K3sExternalKubeConfigLocation = "/etc/rancher/k3s/k3s-external.yaml"
	VelaLinkPos                   = "/usr/local/bin/vela"
	VelaDDockerNetwork            = "k3d-velad"

	K3dImageK3s   = "rancher/k3s:v1.21.10-k3s1"
	K3dImageTools = "rancher/k3d-tools:5.2.2"
	K3dImageProxy = "rancher/k3d-proxy:5.2.2"

	KubeVelaHelmRelease    = "kubevela"
	StatusVelaNotInstalled = "not installed"
	StatusVelaDeployed     = "deployed"
)
