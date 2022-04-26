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
}

// LoadBalancerArgs defines arguments for load balancer command
type LoadBalancerArgs struct {
	Hosts         []string
	Configuration string
}

var (
	KubeConfigLocation         = "/etc/rancher/k3s/k3s.yaml"
	ExternalKubeConfigLocation = "/etc/rancher/k3s/k3s-external.yaml"
	VelaLinkPos                = "/usr/local/bin/vela"
)
