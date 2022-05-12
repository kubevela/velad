package vela

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/oam-dev/kubevela/pkg/utils/system"
	"github.com/oam-dev/velad/pkg/apis"
	"github.com/oam-dev/velad/pkg/cluster"
	"github.com/pkg/errors"

	"github.com/oam-dev/velad/pkg/resources"
	"github.com/oam-dev/velad/pkg/utils"
	"github.com/oam-dev/velad/version"
)

var (
	errf  = utils.Errf
	info  = utils.Info
	infof = utils.Infof
	h     = cluster.DefaultHandler
)

// PrepareVelaChart copy the vela chart to the local directory
func PrepareVelaChart() (string, error) {
	charts, err := resources.VelaChart.Open("static/vela/charts/vela-core.tgz")
	if err != nil {
		return "", err
	}
	chartFile, err := utils.SaveToTemp(charts, "vela-core-*.tgz")
	if err != nil {
		return "", err
	}
	// open the tar to tmpDir/vela-core
	tmpDir, err := utils.GetTmpDir()
	if err != nil {
		return "", err
	}
	// #nosec
	untar := exec.Command("tar", "-xzf", chartFile, "-C", tmpDir)
	err = untar.Run()
	if err != nil {
		return "", err
	}
	untarResult := path.Join(tmpDir, "vela-core")
	return untarResult, nil
}

// LoadVelaImages load vela-core and velaUX images
func LoadVelaImages() error {
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
		imageTgz, err := utils.SaveToTemp(file, "vela-image-"+name+"-*.tar.gz")
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
		imageTar := strings.TrimSuffix(imageTgz, ".gz")
		err = h.LoadImage(imageTar)
		if err != nil {
			return err
		}
	}
	return nil
}

// InstallVelaCLI install vela CLI to local
func InstallVelaCLI() error {
	info("Checking and installing vela CLI...")
	_, err := exec.LookPath("vela")
	if err == nil {
		info("vela CLI is already installed, skip")
		return nil
	}

	info("vela CLI is not installed, installing...")
	return installVelaCLI()
}

// installVelaCLI helps install vela CLI
func installVelaCLI() error {
	pos := utils.GetCLIInstallPath()
	dest, err := os.Executable()
	if err != nil {
		return err
	}
	err = os.Symlink(dest, pos)
	if err != nil {
		return errors.Wrap(err, "Fail to create symlink")
	}
	info("Successfully install vela CLI at: ", pos)
	return nil
}

// PrepareVelaUX place vela-ux chart in ~/.vela/addons/velaux/
func PrepareVelaUX() error {
	velaAddonDir, err := getVelaAddonDir()
	if err != nil {
		return err
	}
	// extract velaux-vx.y.z.tgz to local
	filename := fmt.Sprintf("velaux-%s.tgz", version.VelaVersion)
	tar, err := resources.VelaAddons.Open(path.Join("static/vela/addons", filename))
	if err != nil {
		return err
	}
	defer utils.CloseQuietly(tar)
	// #nosec
	file, err := os.OpenFile(path.Join(velaAddonDir, filename), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer utils.CloseQuietly(file)
	_, err = io.Copy(file, tar)
	if err != nil {
		return errors.Wrap(err, "error when copy velaux-vx.y.z.tgz to local")
	}
	// extract velaux-vx.y.z.tgz to ~/addons/velaux
	err = os.RemoveAll(path.Join(velaAddonDir, "velaux"))
	if err != nil {
		return errors.Wrap(err, "error when remove velaux directory")
	}
	// #nosec
	untar := exec.Command("tar", "-xzf", file.Name(), "-C", velaAddonDir)
	output, err := untar.CombinedOutput()
	utils.InfoBytes(output)
	if err != nil {
		return errors.Wrap(err, "error when untar velaux-vx.y.z.tgz")
	}
	return nil
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
