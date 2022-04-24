package pkg

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
