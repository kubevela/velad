package vela

import (
	"fmt"
	"github.com/oam-dev/kubevela/pkg/utils/system"
	"github.com/oam-dev/velad/pkg/cluster"
	"github.com/pkg/errors"
	"io"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"

	. "github.com/oam-dev/velad/pkg/resources"
	"github.com/oam-dev/velad/pkg/utils"
	"github.com/oam-dev/velad/version"
)

var (
	errf  = utils.Errf
	info  = utils.Info
	infof = utils.Infof
	h     = cluster.DefaultHandler
)

func PrepareVelaChart() (string, error) {
	charts, err := VelaChart.Open("static/vela/charts/vela-core.tgz")
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
	dir, err := VelaImages.ReadDir("static/vela/images")
	if err != nil {
		return err
	}
	for _, entry := range dir {
		file, err := VelaImages.Open(path.Join("static/vela/images", entry.Name()))
		if err != nil {
			return err
		}
		name := strings.Split(entry.Name(), ".")[0]
		imageTgz, err := utils.SaveToTemp(file, "vela-image-"+name+"-*.tar.gz")
		if err != nil {
			return err
		}
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

// LinkToVela create soft link to from vela to velad vela
func LinkToVela() {
	info("Checking and installing vela CLI...")
	_, err := exec.LookPath("vela")
	if err == nil {
		info("vela CLI is already installed, skip")
		return
	}

	info("vela CLI is not installed, installing...")
	pos := getCLIInstallPos()
	err = os.Symlink("velad", pos)
	if err != nil {
		errf("Fail to create symlink: %v\n", err)
		return
	}
	info("Successfully install vela CLI at: ", pos)
	if runtime.GOOS == "windows" {
		info("To add vela to PATH, run:")
		infof("  set PATH=%PATH%;%s", pos)
	}
}

func getCLIInstallPos() string {
	// get vela CLI link position depends on the OS
	switch runtime.GOOS {
	case "linux", "darwin":
		return "/usr/local/bin/vela"
	case "windows":
		dir, _ := system.GetVelaHomeDir()
		binDir := path.Join(dir, "bin")
		_ = os.MkdirAll(binDir, 0755)
		return path.Join(binDir, "vela.exe")
	default:
		utils.UnsupportOS(runtime.GOOS)
	}
	return ""
}

// PrepareVelaUX place vela-ux chart in ~/.vela/addons/velaux/
func PrepareVelaUX() error {
	home, err := system.GetVelaHomeDir()
	if err != nil {
		return err
	}
	velaAddonDir := path.Join(home, "addons")
	if _, err := os.Stat(velaAddonDir); err != nil && os.IsNotExist(err) {
		err := os.MkdirAll(velaAddonDir, 0750)
		if err != nil {
			return errors.Wrap(err, "error when create vela addon directory")
		}
	}
	// extract velaux-vx.y.z.tgz to local
	filename := fmt.Sprintf("velaux-%s.tgz", version.VelaVersion)
	tar, err := VelaAddons.Open(path.Join("static/vela/addons", filename))
	if err != nil {
		return err
	}
	defer utils.CloseQuietly(tar)
	file, err := os.OpenFile(path.Join(velaAddonDir, filename), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
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
	untar := exec.Command("tar", "-xzf", file.Name(), "-C", velaAddonDir)
	output, err := untar.CombinedOutput()
	utils.InfoBytes(output)
	if err != nil {
		return errors.Wrap(err, "error when untar velaux-vx.y.z.tgz")
	}
	return nil
}
