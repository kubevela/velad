//go:build linux

package cluster

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/oam-dev/velad/pkg/apis"
	"github.com/oam-dev/velad/pkg/resources"
	"github.com/oam-dev/velad/pkg/utils"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	config2 "sigs.k8s.io/controller-runtime/pkg/client/config"
)

var (
	info  = utils.Info
	infof = utils.Infof
	// DefaultHandler is the default handler for k3s cluster
	DefaultHandler Handler = &K3sHandler{}
)

// K3sHandler handle k3s in linux
type K3sHandler struct{}

var _ Handler = &K3sHandler{}

// Install install k3s cluster
func (l K3sHandler) Install(args apis.InstallArgs) error {
	err := SetupK3s(args)
	if err != nil {
		return errors.Wrap(err, "fail to setup k3s")
	}
	info("Successfully setup cluster")
	return nil
}

// Uninstall uninstall k3s cluster
func (l K3sHandler) Uninstall(name string) error {
	info("Uninstall k3s...")
	// #nosec
	uCmd := exec.Command("/usr/local/bin/k3s-uninstall.sh")
	err := uCmd.Run()
	if err != nil {
		return errors.Wrap(err, "Fail to uninstall k3s")
	}
	info("Successfully uninstall k3s")
	info("Uninstall vela CLI...")
	// #nosec
	dCmd := exec.Command("rm", apis.VelaLinkPos)
	err = dCmd.Run()
	if err != nil {
		return errors.Wrap(err, "Fail to delete vela link")
	}
	info("Successfully uninstall vela CLI")
	return nil
}

// SetKubeconfig set kubeconfig for k3s
func (l K3sHandler) SetKubeconfig() error {
	return os.Setenv("KUBECONFIG", apis.K3sKubeConfigLocation)
}

// LoadImage load imageTar to k3s cluster
func (l K3sHandler) LoadImage(imageTar string) error {
	// #nosec
	importCmd := exec.Command("k3s", "ctr", "images", "import", imageTar)
	output, err := importCmd.CombinedOutput()
	utils.InfoBytes(output)
	if err != nil {
		return errors.Wrap(err, "Fail to import image")
	}
	infof("Successfully import image %s\n", imageTar)
	return nil
}

// GetStatus get k3s status
func (l K3sHandler) GetStatus() apis.ClusterStatus {
	var status apis.ClusterStatus
	fillK3sBinStatus(&status)
	fillServiceStatus(&status)
	fillVelaStatus(&status)
	return status
}

func fillK3sBinStatus(status *apis.ClusterStatus) {
	_, err := os.Stat(resources.K3sBinaryLocation)
	if err == nil {
		status.K3s.K3sBinary = true
	} else {
		status.K3s.K3sBinary = false
	}
}

func fillServiceStatus(status *apis.ClusterStatus) {
	if status.K3s.Reason != "" {
		return
	}
	// #nosec
	cmd := exec.Command("systemctl", "check", "k3s")
	out, err := cmd.CombinedOutput()
	status.K3s.K3sServiceStatus = string(out)
	if err != nil {
		extErr := new(exec.ExitError)
		err.Error()
		if ok := errors.As(err, &extErr); !ok {
			status.K3s.Reason = fmt.Sprintf("fail to run systemctl: %v", extErr.Error())
		}
	}
}

func fillVelaStatus(status *apis.ClusterStatus) {
	if status.K3s.Reason != "" {
		return
	}
	err := os.Setenv("KUBECONFIG", apis.K3sKubeConfigLocation)
	if err != nil {
		status.K3s.Reason = fmt.Sprintf("fail to set kubeconfig: %v", err)
		return
	}
	restConfig, err := config2.GetConfig()
	if err != nil {
		status.K3s.Reason = fmt.Sprintf("fail to get config: %v", err)
		return
	}
	cfg, err := utils.NewActionConfig(restConfig, false)
	if err != nil {
		status.K3s.Reason = fmt.Sprintf("Failed to get helm action config: %s", err.Error())
		return
	}
	list := action.NewList(cfg)
	list.SetStateMask()
	releases, err := list.Run()
	if err != nil {
		status.K3s.Reason = fmt.Sprintf("Failed to get helm releases: %s", err.Error())
		return
	}
	for _, release := range releases {
		if release.Name == apis.KubeVelaHelmRelease {
			status.K3s.VelaStatus = release.Info.Status.String()
		}
	}
	if status.K3s.VelaStatus == "" {
		status.K3s.VelaStatus = apis.StatusVelaNotInstalled
	}
}

// PrepareK3sImages Write embed images
func PrepareK3sImages() error {
	embedK3sImage, err := resources.K3sImage.Open("static/k3s/images/k3s-airgap-images-amd64.tar.gz")
	if err != nil {
		return err
	}
	defer utils.CloseQuietly(embedK3sImage)
	err = os.MkdirAll(resources.K3sImageDir, 0600)
	if err != nil {
		return err
	}
	/* #nosec */
	bin, err := os.OpenFile(resources.K3sImageLocation, os.O_CREATE|os.O_WRONLY, 0700)
	if err != nil {
		return err
	}
	defer utils.CloseQuietly(bin)
	_, err = io.Copy(bin, embedK3sImage)
	if err != nil {
		return err
	}
	// #nosec
	unGzipCmd := exec.Command("gzip", "-f", "-d", resources.K3sImageLocation)
	output, err := unGzipCmd.CombinedOutput()
	fmt.Print(string(output))
	if err != nil {
		return err
	}
	info("Successfully prepare k3s image")
	return nil
}

// PrepareK3sScript Write k3s install script to local
func PrepareK3sScript() (string, error) {
	embedScript, err := resources.K3sDirectory.Open("static/k3s/other/setup.sh")
	if err != nil {
		return "", err
	}
	scriptName, err := utils.SaveToTemp(embedScript, "k3s-setup-*.sh")
	if err != nil {
		return "", err
	}
	return scriptName, nil
}

// PrepareK3sBin prepare k3s bin
func PrepareK3sBin() error {
	embedK3sBinary, err := resources.K3sDirectory.Open("static/k3s/other/k3s")
	if err != nil {
		return err
	}
	defer utils.CloseQuietly(embedK3sBinary)
	/* #nosec */
	bin, err := os.OpenFile(resources.K3sBinaryLocation, os.O_CREATE|os.O_WRONLY, 0700)
	if err != nil {
		return err
	}
	defer utils.CloseQuietly(bin)
	_, err = io.Copy(bin, embedK3sBinary)
	if err != nil {
		return err
	}
	info("Successfully place k3s binary to " + resources.K3sBinaryLocation)
	return nil
}

// SetupK3s will set up K3s as control plane.
func SetupK3s(cArgs apis.InstallArgs) error {
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
	other := GetK3sServerArgs(cArgs)
	args = append(args, other...)
	/* #nosec */
	cmd := exec.Command("/bin/bash", args...)

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "INSTALL_K3S_SKIP_DOWNLOAD=true")
	output, err := cmd.CombinedOutput()
	fmt.Print(string(output))
	return errors.Wrap(err, "K3s install script failed")
}

// GenKubeconfig generate kubeconfig for accessing from other machine
func (l K3sHandler) GenKubeconfig(bindIP string) error {
	if bindIP == "" {
		return nil
	}
	info("Generating kubeconfig for remote access into ", apis.K3sExternalKubeConfigLocation)
	originConf, err := os.ReadFile(apis.K3sKubeConfigLocation)
	if err != nil {
		return err
	}
	newConf := strings.Replace(string(originConf), "127.0.0.1", bindIP, 1)
	err = os.WriteFile(apis.K3sExternalKubeConfigLocation, []byte(newConf), 0600)
	info("Successfully generate kubeconfig at ", apis.K3sExternalKubeConfigLocation)
	return err
}
