package pkg

import (
	"io"
	"io/fs"
	"io/ioutil"
	"k8s.io/utils/strings/slices"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/oam-dev/kubevela/references/cli"
)

// SaveToTemp helps save an embedded file into a temporary file
func SaveToTemp(file fs.File, format string) (string, error) {
	tempFile, err := ioutil.TempFile("/var", format)
	if err != nil {
		return "", err
	}
	defer CloseQuietly(tempFile)

	_, err = io.Copy(tempFile, file)
	if err != nil {
		return "", err
	}
	return tempFile.Name(), nil
}

// CloseQuietly closes `io.Closer` quietly. Very handy and helpful for code
// quality too.
func CloseQuietly(d io.Closer) {
	_ = d.Close()
}

func IfDeployByPod(controllers string) bool {
	needControllers := []string{"deployment", "job", "replicaset"}
	for _, c := range needControllers {
		if !HaveController(controllers, c) {
			return true
		}
	}
	return false
}

func HaveController(controllers string, c string) bool {
	cs := strings.Split(controllers, ",")
	if slices.Contains(cs, "*") {
		return !slices.Contains(cs, "-"+c)
	} else {
		return slices.Contains(cs, c)
	}
}

func TransArgsToString(args cli.InstallArgs) []string {
	var res []string
	if args.Values != nil {
		res = append(res, "--set="+strings.Join(args.Values, ","))
	}
	if args.Namespace != "" {
		res = append(res, "--namespace="+args.Namespace)
	}
	if !args.Detail {
		res = append(res, "--detail=false")
	}
	if !args.ReuseValues {
		res = append(res, "--reuse=false")
	}
	return res
}

func WarnSaveToken(token string) {
	if token == "" {
		getToken := exec.Command("cat", "/var/lib/rancher/k3s/server/token")
		_token, err := getToken.Output()
		if err != nil {
			errf("Fail to get token, please run `cat /var/lib/rancher/k3s/server/token` and save it.")
			return
		}
		token = string(_token)
	}
	info()
	info("Keep the token below in case of restarting the control plane")
	info(token)
}

func Cleanup() error {
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
