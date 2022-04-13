package pkg

import "github.com/oam-dev/kubevela/references/cli"

// CtrlPlaneArgs defines arguments for ctrl-plane command
type CtrlPlaneArgs struct {
	BindIP      string
	DBEndpoint  string
	IsStart     bool
	Token       string
	Controllers string
	// InstallArgs is parameters passed to vela install command
	InstallArgs cli.InstallArgs
}

// LoadBalancerArgs defines arguments for load balancer command
type LoadBalancerArgs struct {
	Hosts []string
	Configuration string
}
