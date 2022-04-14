package pkg

import (
	"fmt"
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
	// open the tar to /var/charts/vela-core
	untar := exec.Command("tar", "-xzf", chartFile, "-C", "/var")
	err = untar.Run()
	if err != nil {
		return "", err
	}
	untarResult := "/var/vela-core"
	AddTpTemp(untarResult)
	return untarResult, nil
}

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
