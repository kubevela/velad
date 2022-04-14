package pkg

import (
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"k8s.io/utils/strings/slices"
	"os"
	"os/exec"
	"strings"

	"github.com/oam-dev/kubevela/references/cli"
)

var (
	info func(a ...interface{})
	errf func(format string, a ...interface{})

	// tempFiles will be added while installation and clean up after install
	// Can be added by SaveToTemp or AddTpTemp
	tempFiles []string
)

func init() {
	info = func(a ...interface{}) {
		fmt.Println(a...)
	}
	errf = func(format string, a ...interface{}) {
		fmt.Printf(format, a...)
	}
}

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
	tempFiles = append(tempFiles, tempFile.Name())
	return tempFile.Name(), nil
}

func AddTpTemp(file string) {
	tempFiles = append(tempFiles, file)
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
			errf("Fail to get token, please run `cat /var/lib/rancher/k3s/server/token` and save it.\n")
			return
		}
		token = string(_token)
	}
	info()
	info("Keep the token below in case of restarting the control plane")
	info(token)
}

func Cleanup() error {
	for _, f := range tempFiles {
		if err := os.RemoveAll(f); err != nil {
			return err
		}
	}
	return nil
}

func infoBytes(b []byte) {
	if len(b) != 0 {
		info(string(b))
	}
}
