package cluster

import (
	"github.com/oam-dev/velad/pkg/apis"
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
	return serverArgs
}
