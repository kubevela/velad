package cmd

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/oam-dev/kubevela/pkg/utils/common"
	cmdutil "github.com/oam-dev/kubevela/pkg/utils/util"
	"github.com/oam-dev/kubevela/references/cli"
	"github.com/pkg/errors"

	"github.com/oam-dev/velad/pkg/apis"
	"github.com/oam-dev/velad/pkg/cluster"
	"github.com/oam-dev/velad/pkg/utils"
	"github.com/oam-dev/velad/pkg/vela"
	"github.com/oam-dev/velad/version"
)

func tokenCmd(ctx context.Context, args apis.TokenArgs) error {
	err := args.Validate()
	if err != nil {
		return err
	}
	switch runtime.GOOS {
	case "linux":
		_, err := os.Stat(apis.K3sTokenPath)
		if err == nil {
			file, err := os.ReadFile("/var/lib/rancher/k3s/server/token")
			if err != nil {
				return errors.Wrapf(err, "fail to read token file: %s", apis.K3sTokenPath)
			}
			fmt.Println(string(file))
		}
		info("No token found, control plane not set up yet.")
	default:
		token, err := utils.GetTokenFromCluster(ctx, args.Name)
		if err != nil {
			return err
		}
		info(token)
	}
	return nil
}

func installCmd(c common.Args, ioStreams cmdutil.IOStreams, args apis.InstallArgs) error {
	var err error
	err = args.Validate()
	if err != nil {
		return err
	}

	defer func() {
		err := utils.Cleanup()
		if err != nil {
			errf("Fail to clean up: %v\n", err)
		}
	}()

	// Step.1 Set up K3s as control plane cluster
	err = h.Install(args)
	if err != nil {
		return errors.Wrap(err, "Fail to set up cluster")
	}

	// Step.2 Deal with KUBECONFIG
	err = h.GenKubeconfig(args.BindIP)
	if err != nil {
		return errors.Wrap(err, "fail to generate kubeconfig")
	}
	err = h.SetKubeconfig()
	if err != nil {
		return errors.Wrap(err, "fail to set kubeconfig")
	}

	// Step.3 Install Vela CLI
	err = vela.InstallVelaCLI()
	if err != nil {
		// not return because this is acceptable
		errf("fail to install vela CLI: %v\n", err)
	}

	if !args.ClusterOnly {
		// Step.4 load vela-core images
		err = vela.LoadVelaImages()
		if err != nil {
			return errors.Wrap(err, "fail to load vela images")
		}

		// Step.5 save vela-core chart and velaUX addon
		chart, err := vela.PrepareVelaChart()
		if err != nil {
			return errors.Wrap(err, "fail to prepare vela chart")
		}
		err = vela.PrepareVelaUX()
		if err != nil {
			return errors.Wrap(err, "fail to prepare vela UX")
		}
		// Step.6 install vela-core
		info("Installing vela-core Helm chart...")
		ioStreams.Out = utils.VeladWriter{W: os.Stdout}
		installCmd := cli.NewInstallCommand(c, "1", ioStreams)
		installArgs := []string{"--file", chart, "--detail=false", "--version", version.VelaVersion}
		if utils.IfDeployByPod(args.Controllers) {
			installArgs = append(installArgs, "--set", "deployByPod=true")
		}
		userDefinedArgs := utils.TransArgsToString(args.InstallArgs)
		installArgs = append(installArgs, userDefinedArgs...)
		installCmd.SetArgs(installArgs)
		err = installCmd.Execute()
		if err != nil {
			errf("Didn't install vela-core in control plane: %v. You can try \"vela install\" later\n", err)
		}
	}

	utils.PrintGuide(args)
	return nil
}

func kubeconfigCmd(kArgs apis.KubeconfigArgs) error {
	err := kArgs.Validate()
	if err != nil {
		return errors.Wrap(err, "validate kubeconfig args")
	}
	return cluster.PrintKubeConfig(kArgs)

}

func uninstallCmd(uArgs apis.UninstallArgs) error {
	err := uArgs.Validate()
	if err != nil {
		return err
	}
	err = h.Uninstall(uArgs.Name)
	if err != nil {
		return errors.Wrap(err, "Failed to uninstall KubeVela control plane")
	}
	info("Successfully uninstall KubeVela control plane!")
	return nil
}

func statusCmd() {
	info("Checking cluster status...")
	status := h.GetStatus()
	stop := PrintClusterStatus(status)
	if stop {
		return
	}
	info("Checking KubeVela status...")
	vStatus := vela.GetStatus()
	PrintVelaStatus(vStatus)
}
