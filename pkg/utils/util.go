package utils

import (
	"fmt"
	"github.com/oam-dev/kubevela/pkg/utils/system"
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

var (
	Info func(a ...interface{})
	Errf func(format string, a ...interface{})

	// tempFiles will be added while installation and clean up after install
	// Can be added by SaveToTemp or AddToTemp
	tempFiles []string

	velauxDir string
)

func init() {
	Info = func(a ...interface{}) {
		fmt.Println(a...)
	}
	Errf = func(format string, a ...interface{}) {
		fmt.Printf(format, a...)
	}
	dir, err := system.GetVelaHomeDir()
	if err != nil {
		fmt.Println("Failed to vela home dir:", err)
	}
	addonsDir := filepath.Join(dir, "addons")
	velauxDir = filepath.Join(addonsDir, "velaux")
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

func AddToTemp(file string) {
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
			Errf("Fail to get token, please run `cat /var/lib/rancher/k3s/server/token` and save it.\n")
			return
		}
		token = string(_token)
	}
	Info()
	Info("Keep the token below in case of restarting the control plane")
	Info(token)
}

func Cleanup() error {
	for _, f := range tempFiles {
		if err := os.RemoveAll(f); err != nil {
			return err
		}
	}
	return nil
}

func InfoBytes(b []byte) {
	if len(b) != 0 {
		Info(string(b))
	}
}

// VeladWriter will change "vela addon enable" hint and print else as it is.
type VeladWriter struct {
	W io.Writer
}

var _ io.Writer = &VeladWriter{}

func (v VeladWriter) Write(p []byte) (n int, err error) {
	if strings.HasPrefix(string(p), "If you want to enable dashboard, please run \"vela addon enable velaux\"") {
		return v.W.Write([]byte(fmt.Sprintf("If you want to enable dashboard, please run \"vela addon enable %s\"\n", velauxDir)))
	} else {
		return v.W.Write(p)
	}
}
