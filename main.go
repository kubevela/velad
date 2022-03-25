/*
Copyright 2021 The KubeVela Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"embed"
	"fmt"
	"github.com/oam-dev/kubevela/references/cli"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/oam-dev/kubevela/pkg/utils/common"
	cmdutil "github.com/oam-dev/kubevela/pkg/utils/util"
)

var (
	cArgs              CtrlPlaneArgs
	k3sBinaryLocation  = "/usr/local/bin/k3s"
	k3sImageDir        = "/var/lib/rancher/k3s/agent/images/"
	k3sImageLocation   = "/var/lib/rancher/k3s/agent/images/k3s-airgap-images-amd64.tar.gz"
	kubeConfigLocation = "/etc/rancher/k3s/k3s.yaml"

	info func(a ...interface{})
	errf func(format string, a ...interface{})
)

var (
	//go:embed k3s
	K3sDirectory embed.FS

	//go:embed vela/images
	VelaImages embed.FS
	//go:embed vela/charts
	VelaChart embed.FS
)

// CtrlPlaneArgs defines arguments for ctrl-plane command
type CtrlPlaneArgs struct {
	TLSSan                    string
	DBEndpoint                string
	IsJoin                    bool
	Token                     string
	DisableWorkloadController bool
	// InstallArgs is parameters passed to vela install command
	// e.g. "--detail --version=x.y.z"
	InstallArgs cli.InstallArgs
}

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
		NewUninstallCmd(),
	)
	return cmd
}

// NewInstallCmd create install cmd
func NewInstallCmd(c common.Args, ioStreams cmdutil.IOStreams) *cobra.Command {
	info = ioStreams.Info
	errf = ioStreams.Errorf
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
				err := cleanup()
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

			if !cArgs.IsJoin {
				// Step.3 load vela-core images
				err = LoadVelaImages()
				if err != nil {
					errf("Fail to load vela images: %v\n", err)
				}

				// Step.4 save vela-core chart
				chart, err := PrepareVelaChart()
				// Step.5 install vela-core
				info("Installing vela-core Helm chart...")
				installCmd := cli.NewInstallCommand(c, "1", ioStreams)
				installArgs := TransArgsToString(cArgs.InstallArgs)
				if cArgs.DisableWorkloadController {
					installArgs = append(installArgs, "--set", "podOnly=true", "--set", "image.tag=v1.3.0-alpha.1", "--file", chart)
				}
				installCmd.SetArgs(installArgs)
				err = installCmd.Execute()
				if err != nil {
					errf("Fail to install vela-core in control plane: %v. You can try \"vela install\" later\n", err)
					return
				}
			}

			info("Successfully set up KubeVela control plane, run: export KUBECONFIG=$(vela ctrl-plane kubeconfig) to access it")
		},
	}
	cmd.Flags().BoolVar(&cArgs.IsJoin, "join", false, "If set, vela-core won't be installed again")
	cmd.Flags().StringVar(&cArgs.DBEndpoint, "database-endpoint", "", "Use an external database to store control plane metadata")
	cmd.Flags().StringVar(&cArgs.TLSSan, "bind-ip", "", "Bind additional hostname or IP in the kubeconfig TLS cert")
	cmd.Flags().StringVar(&cArgs.Token, "token", "", "Token for identify the cluster. Can be used to restart the control plane or register other node. If not set, random token will be generated")
	cmd.Flags().BoolVar(&cArgs.DisableWorkloadController, "disable-workload-controller", true, "Disable controllers for Deployment/Job/ReplicaSet/StatefulSet/CronJob/DaemonSet")

	// inherit args from `vela install`
	cmd.Flags().StringArrayVarP(&cArgs.InstallArgs.Values, "set", "", []string{}, "set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	cmd.Flags().StringVarP(&cArgs.InstallArgs.Namespace, "namespace", "n", "vela-system", "namespace scope for installing KubeVela Core")
	cmd.Flags().BoolVarP(&cArgs.InstallArgs.Detail, "detail", "d", true, "show detail log of installation")
	cmd.Flags().BoolVarP(&cArgs.InstallArgs.ReuseValues, "reuse", "r", true, "will re-use the user's last supplied values.")

	return cmd
}

func PrepareVelaChart() (string, error) {
	charts, err := VelaChart.Open("vela/charts/vela-core.tgz")
	if err != nil {
		return "", err
	}
	chartFile, err := SaveToTemp(charts, "vela-core-*.tgz")
	if err != nil {
		return "", err
	}
	// open the tar to /var/charts/vela-core
	untar := exec.Command("tar", "-xzf", chartFile, "-C", "/var")
	err = untar.Run()
	if err != nil {
		return "", err
	}
	return "/var/charts/vela-core", nil
}

func LoadVelaImages() error {
	dir, err := VelaImages.ReadDir("vela/images")
	if err != nil {
		return err
	}
	for _, entry := range dir {
		file, err := VelaImages.Open(path.Join("vela/images", entry.Name()))
		if err != nil {
			return err
		}
		name := strings.Split(entry.Name(), ".")[0]
		imageTar, err := SaveToTemp(file, "vela-image-"+name+"-*.tar")
		if err != nil {
			return err
		}
		importCmd := exec.Command("k3s", "ctr", "images", "import", imageTar)
		output, err := importCmd.CombinedOutput()
		if err != nil {
			return err
		}
		fmt.Print(string(output))
		fmt.Println("Successfully load image: ", imageTar)
	}
	return nil
}

func cleanup() error {
	files, err := filepath.Glob("/var/k3s-setup-*.sh")
	if err != nil {
		return err
	}
	images, err := filepath.Glob("/var/vela-image-*.tar")
	if err != nil {
		return err
	}
	charts, err := filepath.Glob("/var/vela-core-*.tgz")
	if err != nil {
		return err
	}

	files = append(files, images...)
	files = append(files, charts...)
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			return err
		}
	}
	return nil
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

// PrepareK3sImages Write embed images
func PrepareK3sImages() error {
	embedK3sImage, err := K3sDirectory.Open("k3s/k3s-airgap-images-amd64.tar.gz")
	if err != nil {
		return err
	}
	defer CloseQuietly(embedK3sImage)
	err = os.MkdirAll(k3sImageDir, 600)
	if err != nil {
		return err
	}
	/* #nosec */
	bin, err := os.OpenFile(k3sImageLocation, os.O_CREATE|os.O_WRONLY, 0700)
	if err != nil {
		return err
	}
	defer CloseQuietly(bin)
	_, err = io.Copy(bin, embedK3sImage)
	if err != nil {
		return err
	}
	unGzipCmd := exec.Command("gzip", "-f", "-d", k3sImageLocation)
	output, err := unGzipCmd.CombinedOutput()
	fmt.Print(string(output))
	if err != nil {
		return err
	}
	info("Successfully prepare k3s image")
	return nil
}

// composeArgs convert args from command to ones passed to k3s install script
func composeArgs(args CtrlPlaneArgs) []string {
	var shellArgs []string
	if args.DBEndpoint != "" {
		shellArgs = append(shellArgs, "--datastore-endpoint="+args.DBEndpoint)
	}
	if args.TLSSan != "" {
		shellArgs = append(shellArgs, "--tls-san="+args.TLSSan)
	}
	if args.Token != "" {
		shellArgs = append(shellArgs, "--token="+args.Token)
	}
	if args.DisableWorkloadController {
		shellArgs = append(shellArgs, "--kube-controller-manager-arg=controllers=*,-deployment,-job,-replicaset,-daemonset,-statefulset,-cronjob",
			// Traefik use Job to install, which is impossible without Job Controller
			"--disable", "traefik")
	}
	return shellArgs
}

// PrepareK3sScript Write k3s install script to local
func PrepareK3sScript() (string, error) {
	embedScript, err := K3sDirectory.Open("k3s/setup.sh")
	if err != nil {
		return "", err
	}
	scriptName, err := SaveToTemp(embedScript, "k3s-setup-*.sh")
	if err != nil {
		return "", err
	}
	return scriptName, nil
}

// PrepareK3sBin prepare k3s bin
func PrepareK3sBin() error {
	embedK3sBinary, err := K3sDirectory.Open("k3s/k3s")
	if err != nil {
		return err
	}
	defer CloseQuietly(embedK3sBinary)
	/* #nosec */
	bin, err := os.OpenFile(k3sBinaryLocation, os.O_CREATE|os.O_WRONLY, 0700)
	if err != nil {
		return err
	}
	defer CloseQuietly(bin)
	_, err = io.Copy(bin, embedK3sBinary)
	if err != nil {
		return err
	}
	info("Successfully place k3s binary to " + k3sBinaryLocation)
	return nil
}

// NewKubeConfigCmd create kubeconfig command for ctrl-plane
func NewKubeConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kubeconfig",
		Short: "print kubeconfig to access control plane",
		Run: func(cmd *cobra.Command, args []string) {
			_, err := os.Stat(kubeConfigLocation)
			if err != nil {
				return
			}
			fmt.Println(kubeConfigLocation)
		},
	}
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

func main() {
	ioStream := cmdutil.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	commandArgs := common.Args{
		Schema: common.Scheme,
	}
	cmd := NewVeladCommand(commandArgs, ioStream)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
