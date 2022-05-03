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

var (
	K3sKubeConfigLocation         = "/etc/rancher/k3s/k3s.yaml"
	K3sExternalKubeConfigLocation = "/etc/rancher/k3s/k3s-external.yaml"
	VelaLinkPos                   = "/usr/local/bin/vela"
	VelaDDockerNetwork            = "k3d-velad"
)
