package cmd

import (
	"os"
	"runtime"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/oam-dev/velad/pkg/apis"
	lb "github.com/oam-dev/velad/pkg/loadbalancer"
	"github.com/oam-dev/velad/pkg/utils"
)

// NewLoadBalancerCmd return loca-balancer command
func NewLoadBalancerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "load-balancer",
		Short: "Configure load balancer between nodes set up by VelaD",
		Long:  "Configure load balancer between nodes set up by VelaD",
	}
	cmd.AddCommand(
		NewLBInstallCmd(),
		NewLBUninstallCmd(),
		NewLBGetPortCmd(),
	)
	return cmd
}

func NewLBGetPortCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-port",
		Short: "Get port info for load-balancer install command",
		Long:  "Get port info for load-balancer install command",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}
	return cmd
}

// NewLBInstallCmd returns load-balancer install command
func NewLBInstallCmd() *cobra.Command {
	var LBArgs apis.LoadBalancerArgs
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Setup load balancer between nodes set up by VelaD",
		Long:  "Setup load balancer between nodes set up by VelaD",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if runtime.GOOS != apis.GoosLinux {
				return errors.New("Installing load balancer is only supported on linux")
			}
			return nil
		},

		Run: func(cmd *cobra.Command, args []string) {
			defer func() {
				err := utils.Cleanup()
				if err != nil {
					errf("Fail to clean up: %v\n", err)
				}
			}()

			if len(LBArgs.Hosts) == 0 {
				errf("Must specify one host at least\n")
				os.Exit(1)
			}
			err := lb.ConfigureNginx(LBArgs)
			if err != nil {
				errf("Fail to setup load balancer (nginx): %v\n", err)
			}
			info("Successfully setup load balancer!")
		},
	}
	cmd.Flags().StringSliceVar(&LBArgs.Hosts, "host", []string{}, "Host IPs of control plane node installed by velad, can be specified multiple or separate value by comma like: IP1,IP2")
	cmd.Flags().StringVarP(&LBArgs.Configuration, "conf", "c", "", "(Optional) Specify the nginx configuration file place, this file will be overwrite")
	cmd.Flags().IntVar(&LBArgs.PortHTTP, "port-http", 0, "Specify the ingress port for HTTP. See velad load-balancer get-port on master node to get the command ")
	cmd.Flags().IntVar(&LBArgs.PortHTTPS, "port-https", 0, "Specify the ingress port for HTTPS. See velad load-balancer get-port on master node to get the command ")
	return cmd
}

// NewLBUninstallCmd returns a cobra command for uninstalling load balancer
func NewLBUninstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall load balancer",
		Long:  "Uninstall load balancer installed by VelaD",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if runtime.GOOS != apis.GoosLinux {
				return errors.New("Uninstalling load balancer is only supported on linux")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			err := lb.UninstallNginx()
			if err != nil {
				errf("Fail to uninstall load balancer (nginx): %v\n", err)
			}
			err = lb.KillNginx()
			if err != nil {
				errf("Fail to kill nginx process: %v\n", err)
			}
		},
	}
	return cmd
}
