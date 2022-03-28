package pkg

import (
	"embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
)

var (
	k3sBinaryLocation = "/usr/local/bin/k3s"
	k3sImageDir       = "/var/lib/rancher/k3s/agent/images/"
	k3sImageLocation  = "/var/lib/rancher/k3s/agent/images/k3s-airgap-images-amd64.tar.gz"

	info func(a ...interface{})
	errf func(format string, a ...interface{})
)

var (
	//go:embed static/k3s
	K3sDirectory embed.FS

	//go:embed static/vela/images
	VelaImages embed.FS
	//go:embed static/vela/charts
	VelaChart embed.FS
)

func init() {
	info = func(a ...interface{}) {
		fmt.Println(a...)
	}
	errf = func(format string, a ...interface{}) {
		fmt.Printf(format, a...)
	}
}

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
	return "/var/charts/vela-core", nil
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

// PrepareK3sImages Write embed images
func PrepareK3sImages() error {
	embedK3sImage, err := K3sDirectory.Open("static/k3s/k3s-airgap-images-amd64.tar.gz")
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

// PrepareK3sScript Write k3s install script to local
func PrepareK3sScript() (string, error) {
	embedScript, err := K3sDirectory.Open("static/k3s/setup.sh")
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
	embedK3sBinary, err := K3sDirectory.Open("static/k3s/k3s")
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


