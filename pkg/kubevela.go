package pkg

import (
	"fmt"
	"github.com/oam-dev/kubevela/pkg/utils/system"
	"github.com/oam-dev/velad/version"
	"github.com/pkg/errors"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
)

func PrepareVelaChart() (string, error) {
	charts, err := VelaChart.Open("static/vela/charts/vela-core.tgz")
	if err != nil {
		return "", err
	}
	chartFile, err := SaveToTemp(charts, "vela-core-*.tgz")
	if err != nil {
		return "", err
	}
	// open the tar to /var/vela-core
	untar := exec.Command("tar", "-xzf", chartFile, "-C", "/var")
	err = untar.Run()
	if err != nil {
		return "", err
	}
	untarResult := "/var/vela-core"
	AddToTemp(untarResult)
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
		imageTar, err := SaveToTemp(file, "vela-image-"+name+"-*.tar")
		if err != nil {
			return err
		}
		importCmd := exec.Command("k3s", "ctr", "images", "import", imageTar)
		output, err := importCmd.CombinedOutput()
		fmt.Print(string(output))
		if err != nil {
			return err
		}
		fmt.Println("Successfully load image: ", imageTar)
	}
	return nil
}

// LinkToVela create soft link to from vela to velad vela
func LinkToVela() {
	_, err := exec.LookPath("vela")
	if err == nil {
		return
	}
	info("Creating symlink to", VelaLinkPos)
	link := exec.Command("ln", "-sf", "velad", VelaLinkPos)
	output, err := link.CombinedOutput()
	infoBytes(output)
	if err != nil {
		errf("Fail to create symlink: %v\n", err)
		return
	}
	info("Successfully install vela CLI at: ", VelaLinkPos)
}

// PrepareVelaUX place vela-ux chart in ~/.vela/addons/velaux/
func PrepareVelaUX() error {
	home, err := system.GetVelaHomeDir()
	if err != nil {
		return err
	}
	velaAddonDir := path.Join(home, ".vela", "addons")
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
	defer CloseQuietly(tar)
	file, err := os.OpenFile(path.Join(velaAddonDir, filename), os.O_RDONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer CloseQuietly(file)
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
	infoBytes(output)
	if err != nil {
		return errors.Wrap(err, "error when untar velaux-vx.y.z.tgz")
	}
	return nil
}
