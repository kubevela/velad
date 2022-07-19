package vela

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"

	"github.com/oam-dev/kubevela/pkg/utils/system"
	"github.com/oam-dev/kubevela/references/cli"
	"github.com/pkg/errors"

	"github.com/oam-dev/velad/pkg/apis"
	"github.com/oam-dev/velad/pkg/cluster"
	"github.com/oam-dev/velad/pkg/resources"
	"github.com/oam-dev/velad/pkg/utils"
	"github.com/oam-dev/velad/version"
)

var (
	info  = utils.Info
	infof = utils.Infof
	h     = cluster.DefaultHandler
)

// PrepareVelaChart copy the vela chart to the local directory
func PrepareVelaChart(ctx *apis.Context) error {
	var (
		err       error
		chartFile string
	)
	charts, err := resources.VelaChart.Open("static/vela/charts/vela-core.tgz")
	if err != nil {
		return err
	}
	format := "vela-core-*.tgz"
	info("Saving and temporary helm chart file:", format)
	if !ctx.DryRun {
		chartFile, err = utils.SaveToTemp(charts, format)
		if err != nil {
			return err
		}
	}

	// open the tar to tmpDir/vela-core
	tmpDir, err := utils.GetTmpDir()
	if err != nil {
		return err
	}
	info("open the tar to tmpDir", tmpDir)
	if !ctx.DryRun {
		// #nosec
		untar := exec.Command("tar", "-xzf", chartFile, "-C", tmpDir)
		err = untar.Run()
		if err != nil {
			return err
		}
	}
	ctx.VelaChartPath = path.Join(tmpDir, "vela-core")
	return nil
}

// LoadVelaImages load vela-core and velaUX images
func LoadVelaImages(ctx *apis.Context) error {
	if runtime.GOOS == apis.GoosDarwin && runtime.GOARCH == "arm64" {
		info("Skip importing vela-core and VelaUX image on darwin-arm64")
		return nil
	}
	var (
		err      error
		imageTgz string
		imageTar string
	)
	dir, err := resources.VelaImages.ReadDir("static/vela/images")
	if err != nil {
		return err
	}
	for _, entry := range dir {
		file, err := resources.VelaImages.Open(path.Join("static/vela/images", entry.Name()))
		if err != nil {
			return err
		}
		name := strings.Split(entry.Name(), ".")[0]
		format := "vela-image-" + name + "-*.tar.gz"
		info("Saving and temporary image file:", format)
		if !ctx.DryRun {
			imageTgz, err = utils.SaveToTemp(file, format)
			if err != nil {
				return err
			}
			// #nosec
			unzipCmd := exec.Command("gzip", "-d", imageTgz)
			output, err := unzipCmd.CombinedOutput()
			utils.InfoBytes(output)
			if err != nil {
				return err
			}
		}

		imageTar = strings.TrimSuffix(imageTgz, ".gz")
		infof("Importing image to cluster using temporary file: %s\n", format)
		if !ctx.DryRun {
			err = h.LoadImage(imageTar)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// InstallVelaCLI install vela CLI to local
func InstallVelaCLI(ctx *apis.Context) error {
	info("Checking and installing vela CLI...")
	_, err := exec.LookPath("vela")
	if err == nil {
		info("vela CLI is already installed, skip")
		return nil
	}

	info("vela CLI is not installed, installing...")
	return installVelaCLI(ctx)
}

// installVelaCLI helps install vela CLI
func installVelaCLI(ctx *apis.Context) error {
	pos := utils.GetCLIInstallPath()
	dest, err := os.Executable()
	if err != nil {
		return err
	}
	info("Installing vela CLI at: ", pos)
	if !ctx.DryRun {
		err = os.Symlink(dest, pos)
		if err != nil {
			return errors.Wrap(err, "Fail to create symlink")
		}
	}
	info("Successfully install vela CLI")
	return nil
}

// PrepareVelaUX place vela-ux chart in ~/.vela/addons/velaux/
func PrepareVelaUX(ctx *apis.Context) error {
	velaAddonDir, err := getVelaAddonDir()
	if err != nil {
		return err
	}
	var (
		output        []byte
		filename      = fmt.Sprintf("velaux-%s.tgz", version.VelaUXVersion)
		velaUXTgzPath = path.Join(velaAddonDir, filename)
		velaUXPath    = path.Join(velaAddonDir, "velaux")
	)

	// extract velaux-vx.y.z.tgz to local
	tar, err := resources.VelaAddons.Open(path.Join("static/vela/addons", filename))
	if err != nil {
		return err
	}
	defer utils.CloseQuietly(tar)

	infof("Copy %s file to %s\n", filename, velaUXTgzPath)
	if !ctx.DryRun {
		// #nosec
		file, err := os.OpenFile(velaUXTgzPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}
		defer utils.CloseQuietly(file)
		_, err = io.Copy(file, tar)
		if err != nil {
			return errors.Wrap(err, "error when copy velaux-vx.y.z.tgz to local")
		}
	}

	infof("Extracting %s to %s\n", velaUXTgzPath, velaUXPath)
	if !ctx.DryRun {
		// #nosec
		err = os.RemoveAll(velaUXPath)
		if err != nil {
			return errors.Wrap(err, "error when remove velaux directory")
		}
		// #nosec
		untar := exec.Command("tar", "-xzf", velaUXTgzPath, "-C", velaAddonDir)
		output, err = untar.CombinedOutput()
		utils.InfoBytes(output)
	}
	return errors.Wrap(err, "error when untar velaux-vx.y.z.tgz")
}

// InstallVelaChart helps install vela-core chart
func InstallVelaChart(ctx *apis.Context, args apis.InstallArgs) error {
	var err error
	info("Installing vela-core Helm chart...")
	ctx.IOStreams.Out = utils.VeladWriter{W: os.Stdout}
	installCmd := cli.NewInstallCommand(ctx.CommonArgs, "1", ctx.IOStreams)
	installArgs := []string{"--file", ctx.VelaChartPath, "--detail=false", "--version", version.VelaVersion}
	if utils.IfDeployByPod(args.Controllers) {
		installArgs = append(installArgs, "--set", "deployByPod=true")
	}
	userDefinedArgs := utils.TransArgsToString(args.InstallArgs)
	installArgs = append(installArgs, userDefinedArgs...)
	installCmd.SetArgs(installArgs)
	infof("Executing \"vela install")
	for _, arg := range installArgs {
		infof("%s ", arg)
	}
	info("\"\n")
	if !ctx.DryRun {
		err = installCmd.Execute()
	}
	return errors.Wrapf(err, "fail to install vela-core helm chart. You can try \"vela install\" later\n")
}

func getVelaAddonDir() (string, error) {
	home, err := system.GetVelaHomeDir()
	if err != nil {
		return "", err
	}
	velaAddonDir := path.Join(home, "addons")
	if _, err := os.Stat(velaAddonDir); err != nil && os.IsNotExist(err) {
		err := os.MkdirAll(velaAddonDir, 0750)
		if err != nil {
			return "", errors.Wrap(err, "error when create vela addon directory")
		}
	}
	return velaAddonDir, nil
}

// GetStatus get kubevela status
func GetStatus() apis.VelaStatus {
	status := apis.VelaStatus{}
	fillVelaCLIStatus(&status)
	fillVelaUXStatus(&status)
	return status
}

func fillVelaCLIStatus(status *apis.VelaStatus) {
	pos := utils.GetCLIInstallPath()
	if _, err := os.Stat(pos); err == nil {
		status.VelaCLIInstalled = true
		status.VelaCLIPath = pos
	}
}

func fillVelaUXStatus(status *apis.VelaStatus) {
	velaAddonDir, err := getVelaAddonDir()
	if err != nil {
		status.VelaUXAddonDirPresent = false
		status.Reason = fmt.Sprintf("failed to get vela addon directory: %v", err)
		return
	}
	velauxDir := path.Join(velaAddonDir, "velaux")
	if _, err := os.Stat(velauxDir); err == nil {
		status.VelaUXAddonDirPresent = true
		status.VelaUXAddonDirPath = velauxDir
	}

}
