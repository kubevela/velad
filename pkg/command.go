package pkg

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/oam-dev/velad/version"

	"github.com/oam-dev/kubevela/pkg/utils/common"
	cmdutil "github.com/oam-dev/kubevela/pkg/utils/util"
	"github.com/oam-dev/kubevela/references/cli"
	"github.com/spf13/cobra"
)

var (
	cArgs                      InstallArgs
	KubeConfigLocation         = "/etc/rancher/k3s/k3s.yaml"
	ExternalKubeConfigLocation = "/etc/rancher/k3s/k3s-external.yaml"
	VelaLinkPos                = "/usr/local/bin/vela"
)

// NewVeladCommand create velad command
func NewVeladCommand() *cobra.Command {
	ioStreams := cmdutil.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	c := common.Args{
		Schema: common.Scheme,
	}
	cmd := &cobra.Command{
		Use:   "velad",
		Short: "Setup a KubeVela control plane air-gapped",
		Long:  "Setup a KubeVela control plane air-gapped, using K3s and only for Linux now",
	}
	cmd.AddCommand(
		NewInstallCmd(c, ioStreams),
		NewLoadBalancerCmd(),
		NewKubeConfigCmd(),
		NewTokenCmd(),
		NewUninstallCmd(),
		NewVersionCmd(),
	)
	return cmd
}

func NewTokenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token",
		Short: "Print control plane token",
		Long:  "Print control plane token, only works if control plane has been set up",
		Run: func(cmd *cobra.Command, args []string) {
			tokenLoc := "/var/lib/rancher/k3s/server/token"
			_, err := os.Stat(tokenLoc)
			if err == nil {
				file, err := os.ReadFile("/var/lib/rancher/k3s/server/token")
				if err != nil {
					errf("Fail to read token file: %s: %v\n", tokenLoc, err)
					return
				}
				fmt.Println(string(file))
				return
			}
			info("No token found, control plane not set up yet.")
		},
	}
	return cmd
}

// NewInstallCmd create install cmd
func NewInstallCmd(c common.Args, ioStreams cmdutil.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Quickly setup a KubeVela control plane",
		Long:  "Quickly setup a KubeVela control plane, using K3s and only for Linux now",
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			//if runtime.GOOS != "linux" {
			//	info("Launch control plane is not supported now in non-linux OS, exiting")
			//	return
			//}
			defer func() {
				err := Cleanup()
				if err != nil {
					errf("Fail to clean up: %v\n", err)
				}
			}()

			// Step.1 Set up K3s as control plane cluster
			err = SetupK3s(cArgs)
			if err != nil {
				errf("Fail to setup k3s: %v\n", err)
				return
			}
			info("Successfully setup cluster")

			// Step.2 Set KUBECONFIG
			err = os.Setenv("KUBECONFIG", KubeConfigLocation)
			if err != nil {
				errf("Fail to set KUBECONFIG environment var: %v\n", err)
				return
			}

			// Step.3 Install Vela CLI
			LinkToVela()

			// Step.4 load vela-core images
			err = LoadVelaImages()
			if err != nil {
				errf("Fail to load vela images: %v\n", err)
			}

			if !cArgs.ClusterOnly {

				// Step.5 save vela-core chart and velaUX addon
				chart, err := PrepareVelaChart()
				if err != nil {
					errf("Fail to prepare vela chart: %v\n", err)
				}
				err = PrepareVelaUX()
				if err != nil {
					errf("Fail to prepare velaUX: %v\n", err)
				}
				// Step.6 install vela-core
				info("Installing vela-core Helm chart...")
				ioStreams.Out = VeladWriter{os.Stdout}
				installCmd := cli.NewInstallCommand(c, "1", ioStreams)
				installArgs := []string{"--file", chart, "--detail=false", "--version", version.VelaVersion}
				if IfDeployByPod(cArgs.Controllers) {
					installArgs = append(installArgs, "--set", "deployByPod=true")
				}
				userDefinedArgs := TransArgsToString(cArgs.InstallArgs)
				installArgs = append(installArgs, userDefinedArgs...)
				installCmd.SetArgs(installArgs)
				err = installCmd.Execute()
				if err != nil {
					errf("Didn't install vela-core in control plane: %v. You can try \"vela install\" later\n", err)
				}
			}

			// Step.7 Generate external kubeconfig
			if cArgs.BindIP != "" {
				err = GenKubeconfig(cArgs.BindIP)
				if err != nil {
					return
				}
			}
			WarnSaveToken(cArgs.Token)
			info("Successfully install KubeVela control plane! Try: vela components")
		},
	}
	cmd.Flags().BoolVar(&cArgs.ClusterOnly, "cluster-only", false, "If set, start cluster without installing vela-core, typically used when restart a control plane where vela-core has been installed")
	cmd.Flags().StringVar(&cArgs.DBEndpoint, "database-endpoint", "", "Use an external database to store control plane metadata, please ref https://rancher.com/docs/k3s/latest/en/installation/datastore/#datastore-endpoint-format-and-functionality for the format")
	cmd.Flags().StringVar(&cArgs.BindIP, "bind-ip", "", "Bind additional hostname or IP in the kubeconfig TLS cert")
	cmd.Flags().StringVar(&cArgs.Token, "token", "", "Token for identify the cluster. Can be used to restart the control plane or register other node. If not set, random token will be generated")
	cmd.Flags().StringVar(&cArgs.Controllers, "controllers", "*", "A list of controllers to enable, check \"--controllers\" argument for more spec in https://kubernetes.io/docs/reference/command-line-tools-reference/kube-controller-manager/")

	// inherit args from `vela install`
	cmd.Flags().StringArrayVarP(&cArgs.InstallArgs.Values, "set", "", []string{}, "set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	cmd.Flags().StringVarP(&cArgs.InstallArgs.Namespace, "namespace", "n", "vela-system", "namespace scope for installing KubeVela Core")
	cmd.Flags().BoolVarP(&cArgs.InstallArgs.Detail, "detail", "d", true, "show detail log of installation")
	cmd.Flags().BoolVarP(&cArgs.InstallArgs.ReuseValues, "reuse", "r", true, "will re-use the user's last supplied values.")

	return cmd
}

// NewKubeConfigCmd create kubeconfig command for ctrl-plane
func NewKubeConfigCmd() *cobra.Command {
	var (
		internal bool
		external bool
	)
	cmd := &cobra.Command{
		Use:   "kubeconfig",
		Short: "print kubeconfig to access control plane",
		Run: func(cmd *cobra.Command, args []string) {
			PrintKubeConfig(internal, external)
		},
	}
	cmd.Flags().BoolVar(&internal, "internal", false, "Print kubeconfig that can only be used in this machine")
	cmd.Flags().BoolVar(&external, "external", false, "Print kubeconfig that can be used on other machine")
	return cmd
}

// NewUninstallCmd create uninstall command
func NewUninstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "uninstall control plane",
		RunE: func(cmd *cobra.Command, args []string) error {
			// #nosec
			uCmd := exec.Command("/usr/local/bin/k3s-uninstall.sh")
			err := uCmd.Run()
			if err != nil {
				errf("Fail to uninstall k3s: %v\n", err)
			}
			dCmd := exec.Command("rm", VelaLinkPos)
			err = dCmd.Run()
			if err != nil {
				errf("Fail to delete vela symlink: %v\n", err)
			}
			return nil
		},
	}
	return cmd
}

func NewLoadBalancerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "load-balancer",
		Short: "Configure load balancer between nodes set up by VelaD",
		Long:  "Configure load balancer between nodes set up by VelaD",
	}
	cmd.AddCommand(
		NewLBInstallCmd(),
		NewLBUninstallCmd(),
	)
	return cmd
}

func NewLBInstallCmd() *cobra.Command {
	var LBArgs LoadBalancerArgs
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Setup load balancer between nodes set up by VelaD",
		Long:  "Setup load balancer between nodes set up by VelaD",
		Run: func(cmd *cobra.Command, args []string) {
			defer func() {
				err := Cleanup()
				if err != nil {
					errf("Fail to clean up: %v\n", err)
				}
			}()

			if len(LBArgs.Hosts) == 0 {
				errf("Must specify one host at least\n")
				os.Exit(1)
			}
			err := ConfigureNginx(LBArgs)
			if err != nil {
				errf("Fail to setup load balancer (nginx): %v\n", err)
			}
			info("Successfully setup load balancer!")
		},
	}
	cmd.Flags().StringSliceVar(&LBArgs.Hosts, "host", []string{}, "Host IPs of control plane node installed by velad, can be specified multiple or separate value by comma like: IP1,IP2")
	cmd.Flags().StringVarP(&LBArgs.Configuration, "conf", "c", "", "(Optional) Specify the nginx configuration file place, this file will be overwrite")
	return cmd
}

func NewLBUninstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall load balancer",
		Long:  "Uninstall load balancer installed by VelaD",
		Run: func(cmd *cobra.Command, args []string) {
			err := UninstallNginx()
			if err != nil {
				errf("Fail to uninstall load balancer (nginx): %v\n", err)
			}
			err = KillNginx()
			if err != nil {
				errf("Fail to kill nginx process: %v\n", err)
			}
		},
	}
	return cmd
}

func NewVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Prints VelaD build version information",
		Long:  "Prints VelaD build version information.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Core Version: %s\n", version.VelaVersion)
			fmt.Printf("VelaD Version: %s\n", version.VelaDVersion)
		},
	}
	return cmd

}
