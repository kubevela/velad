package utils

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kyokomi/emoji/v2"
	"github.com/oam-dev/kubevela/pkg/utils/system"
	"github.com/oam-dev/kubevela/references/cli"
	"k8s.io/utils/strings/slices"

	"github.com/oam-dev/velad/pkg/apis"
)

var (
	// Info print message
	Info func(a ...interface{})
	// Infof print message with format
	Infof func(format string, a ...interface{})
	// InfoP print message with padding
	InfoP func(padding int, a ...interface{})
	// Errf print error with format
	Errf func(format string, a ...interface{})

	velauxDir string
)

func init() {
	Info = func(a ...interface{}) {
		fmt.Println(a...)
	}
	Errf = func(format string, a ...interface{}) {
		fmt.Printf(format, a...)
	}
	Infof = func(format string, a ...interface{}) {
		fmt.Printf(format, a...)
	}
	InfoP = func(padding int, a ...interface{}) {
		fmt.Printf("%*s", padding, "")
		fmt.Println(a...)
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
	tmpDir, err := GetTmpDir()
	if err != nil {
		return "", err
	}
	tempFile, err := ioutil.TempFile(tmpDir, format)
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

// IfDeployByPod returns true if the given controllers doesn't contain one of Deployment/Job/ReplicaSet
func IfDeployByPod(controllers string) bool {
	needControllers := []string{"deployment", "job", "replicaset"}
	for _, c := range needControllers {
		if !HaveController(controllers, c) {
			return true
		}
	}
	return false
}

// HaveController returns true if the given controllers contains the given controller
func HaveController(controllers string, c string) bool {
	cs := strings.Split(controllers, ",")
	if slices.Contains(cs, "*") {
		return !slices.Contains(cs, "-"+c)
	}
	return slices.Contains(cs, c)
}

// TransArgsToString converts args to string array, which helps to pass args to vela install command
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

// WarnSaveToken warns user to save token for cluster
func WarnSaveToken(token string, clusterName string) {
	var err error
	if token == "" {
		switch runtime.GOOS {
		case "linux":
			// #nosec
			getToken := exec.Command("cat", "/var/lib/rancher/k3s/server/token")
			_token, err := getToken.Output()
			if err != nil {
				Errf("Fail to get token, please run `cat /var/lib/rancher/k3s/server/token` and save it.\n")
				return
			}
			token = string(_token)
		default:
			token, err = GetTokenFromCluster(context.Background(), clusterName)
			if err != nil {
				Errf("Fail to get token from cluster: %v", err)
			}
		}
	}
	Info()
	Info("Keep the token below if you want to restart the control plane")
	if token != "" {
		Info(token)
	} else {
		Info("[No token found]")
	}
}

// Cleanup removes the temporary directory
func Cleanup() error {
	tmpDir, err := GetTmpDir()
	if err != nil {
		return err
	}
	return os.RemoveAll(tmpDir)
}

// InfoBytes is a helper function to print a byte array
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

// Write implements io.Writer. Change the hint to "vela addon enable velaux" and print it with local dir.
func (v VeladWriter) Write(p []byte) (n int, err error) {
	if strings.HasPrefix(string(p), "If you want to enable dashboard, please run \"vela addon enable velaux\"") {
		return v.W.Write([]byte(fmt.Sprintf("If you want to enable dashboard, please run \"vela addon enable %s\"\n", velauxDir)))
	}
	return v.W.Write(p)
}

// GetTmpDir returns the temporary directory when want to save some files
func GetTmpDir() (string, error) {
	dir, err := system.GetVelaHomeDir()
	if err != nil {
		return "", err
	}
	tmpDir := filepath.Join(dir, "tmp")
	if err := os.MkdirAll(tmpDir, 0700); err != nil {
		return "", err
	}
	return tmpDir, nil
}

// GetDefaultVelaDKubeconfigPath returns the default kubeconfig path for VelaD
func GetDefaultVelaDKubeconfigPath() string {
	var kubeconfigPos string
	switch runtime.GOOS {
	case "darwin":
		kubeconfigPos = filepath.Join(os.Getenv("HOME"), ".kube", "velad-cluster-default")
	case "linux":
		kubeconfigPos = apis.K3sKubeConfigLocation
	case "windows":
		kubeconfigPos = filepath.Join(os.Getenv("USERPROFILE"), ".kube", "velad-cluster-default")
	default:
		UnsupportedOS(runtime.GOOS)
	}
	return kubeconfigPos
}

// GetKubeconfigDir returns the kubeconfig directory.
func GetKubeconfigDir() string {
	var kubeconfigDir string
	switch runtime.GOOS {
	case "darwin", "linux":
		kubeconfigDir = filepath.Join(os.Getenv("HOME"), ".kube")
	case "windows":
		kubeconfigDir = filepath.Join(os.Getenv("USERPROFILE"), ".kube")
	}
	return kubeconfigDir
}

// PrintGuide will print guide for user.
func PrintGuide(args apis.InstallArgs) {
	WarnSaveToken(args.Token, args.Name)
	if !args.ClusterOnly {
		emoji.Println(":rocket: Successfully install KubeVela control plane")
		emoji.Println(":telescope: See available commands with `vela help`")
	} else {
		emoji.Println(":rocket: Successfully install a pure cluster! ")
		emoji.Println(":link: If you have a cluster with KubeVela, Join this as sub-cluster:")
		emoji.Println("    vela cluster join $(velad kubeconfig --name foo --internal)")
		emoji.Println(":key: To access the cluster, set KUBECONFIG:")
		emoji.Printf("    export KUBECONFIG=$(velad kubeconfig --name %s --host)\n", args.Name)
	}
}

// IsVelaCommand judge if app start by vela
func IsVelaCommand(s string) bool {
	sl := strings.Split(s, "/")
	return sl[len(sl)-1] == "vela"
}

// SetDefaultKubeConfigEnv helps set KUBECONFIG to the default location
func SetDefaultKubeConfigEnv() {
	RecommendedConfigPathEnvVar := "KUBECONFIG"
	kubeconfig := os.Getenv(RecommendedConfigPathEnvVar)
	if kubeconfig == "" {
		kubeconfig = GetDefaultVelaDKubeconfigPath()
		_ = os.Setenv(RecommendedConfigPathEnvVar, kubeconfig)
	}
}
