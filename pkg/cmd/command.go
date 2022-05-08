package cmd

import (
	"fmt"
	"os"

	"github.com/oam-dev/kubevela/pkg/utils/common"
	cmdutil "github.com/oam-dev/kubevela/pkg/utils/util"
	"github.com/oam-dev/velad/pkg/apis"
	"github.com/oam-dev/velad/pkg/cluster"
	"github.com/oam-dev/velad/pkg/utils"
	"github.com/oam-dev/velad/version"
	"github.com/spf13/cobra"
)

var (
	errf  = utils.Errf
	info  = utils.Info
	infoP = utils.InfoP
	h     = cluster.DefaultHandler
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
		NewStatusCmd(),
		NewLoadBalancerCmd(),
		NewKubeConfigCmd(),
		NewTokenCmd(),
		NewUninstallCmd(),
		NewVersionCmd(),
	)
	return cmd
}

// NewTokenCmd create token command
func NewTokenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token",
		Short: "Print control plane token",
		Long:  "Print control plane token, only works if control plane has been set up",
		Run: func(cmd *cobra.Command, args []string) {
			tokenCmd()
		},
	}
	return cmd
}

// NewInstallCmd create install cmd
func NewInstallCmd(c common.Args, ioStreams cmdutil.IOStreams) *cobra.Command {
	iArgs := apis.InstallArgs{}
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Quickly setup a KubeVela control plane",
		Long:  "Quickly setup a KubeVela control plane, using K3s and only for Linux now",
		RunE: func(cmd *cobra.Command, args []string) error {
			return installCmd(c, ioStreams, iArgs)
		},
	}
	cmd.Flags().BoolVar(&iArgs.ClusterOnly, "cluster-only", false, "If set, start cluster without installing vela-core, typically used when restart a control plane where vela-core has been installed")
	cmd.Flags().StringVar(&iArgs.DBEndpoint, "database-endpoint", "", "Use an external database to store control plane metadata, please ref https://rancher.com/docs/k3s/latest/en/installation/datastore/#datastore-endpoint-format-and-functionality for the format")
	cmd.Flags().StringVar(&iArgs.BindIP, "bind-ip", "", "Bind additional hostname or IP in the kubeconfig TLS cert")
	cmd.Flags().StringVar(&iArgs.Token, "token", "", "Token for identify the cluster. Can be used to restart the control plane or register other node. If not set, random token will be generated")
	cmd.Flags().StringVar(&iArgs.Controllers, "controllers", "*", "A list of controllers to enable, check \"--controllers\" argument for more spec in https://kubernetes.io/docs/reference/command-line-tools-reference/kube-controller-manager/")
	cmd.Flags().StringVar(&iArgs.Name, "name", "default", "The name of the cluster. only works when NOT in linux environment")

	// inherit args from `vela install`
	cmd.Flags().StringArrayVarP(&iArgs.InstallArgs.Values, "set", "", []string{}, "set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	cmd.Flags().StringVarP(&iArgs.InstallArgs.Namespace, "namespace", "n", "vela-system", "namespace scope for installing KubeVela Core")
	cmd.Flags().BoolVarP(&iArgs.InstallArgs.Detail, "detail", "d", true, "show detail log of installation")
	cmd.Flags().BoolVarP(&iArgs.InstallArgs.ReuseValues, "reuse", "r", true, "will re-use the user's last supplied values.")

	return cmd
}

// NewStatusCmd create status command
func NewStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show the status of the control plane",
		Long:  "Show the status of the control plane",
		Run: func(cmd *cobra.Command, args []string) {
			statusCmd()
		},
	}
	return cmd
}

// NewKubeConfigCmd create kubeconfig command for ctrl-plane
func NewKubeConfigCmd() *cobra.Command {
	kArgs := apis.KubeconfigArgs{}
	cmd := &cobra.Command{
		Use:   "kubeconfig",
		Short: "print kubeconfig to access control plane",
		RunE: func(cmd *cobra.Command, args []string) error {
			return kubeconfigCmd(kArgs)
		},
	}
	cmd.Flags().StringVarP(&kArgs.Name, "name", "n", "default", "The name of cluster, Only works in macOS/Windows")
	cmd.Flags().BoolVar(&kArgs.Internal, "internal", false, "Print kubeconfig that used in Docker network. Typically used in \"vela cluster join\". Only works in macOS/Windows. ")
	cmd.Flags().BoolVar(&kArgs.External, "external", false, "Print kubeconfig that can be used on other machine")
	cmd.Flags().BoolVar(&kArgs.Host, "host", false, "Print kubeconfig path that can be used in this machine")
	return cmd
}

// NewUninstallCmd create uninstall command
func NewUninstallCmd() *cobra.Command {
	uArgs := apis.UninstallArgs{}
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "uninstall control plane",
		RunE: func(cmd *cobra.Command, args []string) error {
			return uninstallCmd(uArgs)
		},
	}
	cmd.Flags().StringVarP(&uArgs.Name, "name", "n", "default", "The name of the control plane. Only works when NOT in linux environment")
	return cmd
}

// NewVersionCmd create version command
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
