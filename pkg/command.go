package pkg

import (
	"fmt"
	"github.com/oam-dev/kubevela/pkg/utils/common"
	cmdutil "github.com/oam-dev/kubevela/pkg/utils/util"
	"github.com/oam-dev/kubevela/references/cli"
	"github.com/oam-dev/velad/version"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"strings"
)

var (
	cArgs                      CtrlPlaneArgs
	kubeConfigLocation         = "/etc/rancher/k3s/k3s.yaml"
	externalKubeConfigLocation = "/etc/rancher/k3s/k3s-external.yaml"
)

// NewVeladCommand create velad command
func NewVeladCommand(c common.Args, ioStreams cmdutil.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "velad",
		Short: "Setup a KubeVela control plane air-gapped",
		Long:  "Setup a KubeVela control plane air-gapped, using K3s and only for Linux now",
	}
	cmd.AddCommand(
		NewInstallCmd(c, ioStreams),
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
					errf("Fail to clean up install script: %v", err)
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
			err = os.Setenv("KUBECONFIG", kubeConfigLocation)
			if err != nil {
				errf("Fail to set KUBECONFIG environment var: %v\n", err)
				return
			}

			if !cArgs.IsStart {
				// Step.3 load vela-core images
				err = LoadVelaImages()
				if err != nil {
					errf("Fail to load vela images: %v\n", err)
				}

				// Step.4 save vela-core chart
				chart, err := PrepareVelaChart()
				if err != nil {
					errf("Fail to prepare vela chart: %v\n", err)
				}
				// Step.5 install vela-core
				info("Installing vela-core Helm chart...")
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

			// Step.6 Generate external kubeconfig
			err = GenKubeconfig(cArgs.BindIP)
			if err != nil {
				return
			}
			WarnSaveToken(cArgs.Token)
		},
	}
	cmd.Flags().BoolVar(&cArgs.IsStart, "start", false, "If set, start cluster without installing vela-core, typically used when restart a control plane where vela-core has been installed")
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

// GenKubeconfig will generate kubeconfig for remote access.
// This won't modify the origin kubeconfig generated by k3s
func GenKubeconfig(bindIP string) error {
	var err error
	if bindIP != "" {
		info("Generating kubeconfig for remote access into ", externalKubeConfigLocation)
		originConf, err := os.ReadFile(kubeConfigLocation)
		if err != nil {
			return err
		}
		newConf := strings.Replace(string(originConf), "127.0.0.1", bindIP, 1)
		err = os.WriteFile(externalKubeConfigLocation, []byte(newConf), 600)
	}
	internalFlag := ""
	if bindIP == "" {
		internalFlag = " --internal"
	}
	info("Successfully set up KubeVela control plane, run: export KUBECONFIG=$(velad kubeconfig" + internalFlag + ") to access it")
	return err
}

// SetupK3s will set up K3s as control plane.

func SetupK3s(cArgs CtrlPlaneArgs) error {
	info("Preparing cluster setup script...")
	script, err := PrepareK3sScript()
	if err != nil {
		return errors.Wrap(err, "fail to prepare k3s setup script")
	}

	info("Preparing k3s binary...")
	err = PrepareK3sBin()
	if err != nil {
		return errors.Wrap(err, "Fail to prepare k3s binary")
	}

	info("Preparing k3s images")
	err = PrepareK3sImages()
	if err != nil {
		return errors.Wrap(err, "Fail to prepare k3s images")
	}

	info("Setting up cluster...")
	args := []string{script}
	other := composeArgs(cArgs)
	args = append(args, other...)
	/* #nosec */
	cmd := exec.Command("/bin/bash", args...)

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "INSTALL_K3S_SKIP_DOWNLOAD=true")
	output, err := cmd.CombinedOutput()
	fmt.Print(string(output))
	return errors.Wrap(err, "K3s install script failed")
}

// composeArgs convert args from command to ones passed to k3s install script
func composeArgs(args CtrlPlaneArgs) []string {
	var shellArgs []string
	if args.DBEndpoint != "" {
		shellArgs = append(shellArgs, "--datastore-endpoint="+args.DBEndpoint)
	}
	if args.BindIP != "" {
		shellArgs = append(shellArgs, "--tls-san="+args.BindIP)
	}
	if args.Token != "" {
		shellArgs = append(shellArgs, "--token="+args.Token)
	}
	if args.Controllers != "*" {
		shellArgs = append(shellArgs, "--kube-controller-manager-arg=controllers="+args.Controllers)
		// TODO : deal with coredns/local-path-provisioner/metrics-server Deployment when no deployment controllers
		if !HaveController(args.Controllers, "job") {
			// Traefik use Job to install, which is impossible without Job Controller
			shellArgs = append(shellArgs, "--disable", "traefik")
		}
	}
	return shellArgs
}

// NewKubeConfigCmd create kubeconfig command for ctrl-plane
func NewKubeConfigCmd() *cobra.Command {
	var internal bool
	cmd := &cobra.Command{
		Use:   "kubeconfig",
		Short: "print kubeconfig to access control plane",
		Run: func(cmd *cobra.Command, args []string) {
			configP := externalKubeConfigLocation
			if internal {
				configP = kubeConfigLocation
			}
			_, err := os.Stat(configP)
			if err != nil {
				return
			}
			fmt.Println(configP)
		},
	}
	cmd.Flags().BoolVar(&internal, "internal", false, "If set, the kubeconfig printed can be only used in this machine")
	return cmd
}

// NewUninstallCmd create uninstall command
func NewUninstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "uninstall control plane",
		RunE: func(cmd *cobra.Command, args []string) error {
			// #nosec
			uninstallCmd := exec.Command("/usr/local/bin/k3s-uninstall.sh")
			return uninstallCmd.Run()
		},
	}
	return cmd
}

func NewVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Prints velad build version information",
		Long:  "Prints velad build version information.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Core Version: %s", version.VelaVersion)
		},
	}
	return cmd

}
