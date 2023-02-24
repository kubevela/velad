package cluster

import (
	"strconv"

	"github.com/oam-dev/velad/pkg/apis"
)

// K3sListenPort is the port k3s server listens on
var K3sListenPort = 8443

// LBListenPort is the port load balancer of master nodes listens on, LB will port-forward to k3s server's K3sListenPort
var LBListenPort = 6443

// GetK3sServerArgs convert install args to ones passed to k3s server
func GetK3sServerArgs(args apis.InstallArgs) []string {
	var serverArgs []string
	serverArgs = append(serverArgs, "--https-listen-port="+strconv.Itoa(K3sListenPort))
	if args.DBEndpoint != "" {
		serverArgs = append(serverArgs, "--datastore-endpoint="+args.DBEndpoint)
	}
	if args.BindIP != "" {
		serverArgs = append(serverArgs, "--tls-san="+args.BindIP)
	}
	if args.NodePublicIP != "" {
		serverArgs = append(serverArgs, "--node-external-ip="+args.NodePublicIP)
	}
	if args.Token != "" {
		serverArgs = append(serverArgs, "--token="+args.Token)
	}
	return serverArgs
}
