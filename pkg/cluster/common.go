package cluster

import (
	"github.com/oam-dev/velad/pkg/apis"
	"github.com/oam-dev/velad/pkg/utils"
)

// GetK3sServerArgs convert install args to ones passed to k3s server
func GetK3sServerArgs(args apis.InstallArgs) []string {
	var serverArgs []string
	if args.DBEndpoint != "" {
		serverArgs = append(serverArgs, "--datastore-endpoint="+args.DBEndpoint)
	}
	if args.BindIP != "" {
		serverArgs = append(serverArgs, "--tls-san="+args.BindIP, "--node-external-ip="+args.BindIP)
	}
	if args.Token != "" {
		serverArgs = append(serverArgs, "--token="+args.Token)
	}
	if args.Controllers != "*" {
		serverArgs = append(serverArgs, "--kube-controller-manager-arg=controllers="+args.Controllers)
		// TODO : deal with coredns/local-path-provisioner/metrics-server Deployment when no deployment controllers
		if !utils.HaveController(args.Controllers, "job") {
			// Traefik use Job to install, which is impossible without Job Controller
			serverArgs = append(serverArgs, "--disable", "traefik")
		}
	}
	return serverArgs
}
