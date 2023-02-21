package loadbalancer

import (
	"context"
	"github.com/oam-dev/velad/pkg/utils"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

func PrintPortAndCmd() error {
	cli, err := utils.GetClient()
	if err != nil {
		return err
	}
	svc := v1.Service{}
	err = cli.Get(context.Background(), client.ObjectKey{
		Namespace: "kube-system",
		Name:      "traefik",
	}, &svc)
	if err != nil {
		return err
	}
	portHTTP := 0
	portHTTPS := 0
	for _, port := range svc.Spec.Ports {
		switch port.Port {
		case 80:
			portHTTP = int(port.NodePort)
		case 443:
			portHTTPS = int(port.NodePort)
		}
	}
	if portHTTP == 0 {
		utils.Errf("http port is not found\n")
	}
	if portHTTPS == 0 {
		utils.Errf("https port is not found\n")
	}
	hosts := []string{}
	for _, i := range svc.Status.LoadBalancer.Ingress {
		// todo(chivalryq) support hostname
		hosts = append(hosts, i.IP)
	}
	utils.Infof("To setup load-balancer, run the following command on node acts as load-balancer:\n")
	utils.Infof("  velad load-balancer install --http-port %d --https-port %d --host=%s\n", portHTTP, portHTTPS, strings.Join(hosts, ","))
	return nil
}
