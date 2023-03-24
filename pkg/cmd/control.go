package cmd

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/oam-dev/kubevela/pkg/utils/common"
	cmdutil "github.com/oam-dev/kubevela/pkg/utils/util"
	"github.com/pkg/errors"

	"github.com/oam-dev/velad/pkg/apis"
	"github.com/oam-dev/velad/pkg/cluster"
	"github.com/oam-dev/velad/pkg/utils"
	"github.com/oam-dev/velad/pkg/vela"
)

func tokenCmd(ctx context.Context, args apis.TokenArgs) error {
	err := args.Validate()
	if err != nil {
		return err
	}
	switch runtime.GOOS {
	case apis.GoosLinux:
		_, err := os.Stat(apis.K3sTokenPath)
		if err != nil {
			info("No token found, control plane not set up yet.")
		}
		file, err := os.ReadFile("/var/lib/rancher/k3s/server/token")
		if err != nil {
			return errors.Wrapf(err, "fail to read token file: %s", apis.K3sTokenPath)
		}
		fmt.Println(string(file))
		return nil

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
	ctx := &apis.Context{
		DryRun:     args.DryRun,
		CommonArgs: c,
		IOStreams:  ioStreams,
	}
	var err error

	err = args.Validate()
	if err != nil {
		return err
	}

	defer func() {
		if args.DryRun {
			return
		}
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
	err = h.GenKubeconfig(*ctx, args.BindIP)
	if err != nil {
		return errors.Wrap(err, "fail to generate kubeconfig")
	}
	err = h.SetKubeconfig()
	if err != nil {
		return errors.Wrap(err, "fail to set kubeconfig")
	}

	// Step.3 Install Vela CLI
	err = vela.InstallVelaCLI(ctx)
	if err != nil {
		// not return because this is acceptable
		errf("fail to install vela CLI: %v\n", err)
	}

	if !args.ClusterOnly {
		// Step.4 load vela-core images
		err = vela.LoadVelaImages(ctx)
		if err != nil {
			return errors.Wrap(err, "fail to load vela images")
		}

		// Step.5 save vela-core chart and velaUX addon
		err := vela.PrepareVelaChart(ctx)
		if err != nil {
			return errors.Wrap(err, "fail to prepare vela chart")
		}
		err = vela.PrepareVelaUX(ctx)
		if err != nil {
			return errors.Wrap(err, "fail to prepare vela UX")
		}
		// Step.6 install vela-core
		err = vela.InstallVelaChart(ctx, args)
		if err != nil {
			return errors.Wrap(err, "fail to install vela-core chart")
		}
	}

	utils.PrintGuide(ctx, args)
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
		return errors.Wrap(err, "Failed to uninstall KubeVela control plane/worker node")
	}
	info("Successfully uninstall KubeVela control plane/worker node")
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

func joinCmd(args apis.JoinArgs) error {
	if err := args.Validate(); err != nil {
		return err
	}
	return h.Join(args)

}
