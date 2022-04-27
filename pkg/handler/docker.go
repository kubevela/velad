//go:build !linux

package handler

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/docker/go-connections/nat"
	"github.com/oam-dev/kubevela/pkg/utils/system"
	"github.com/oam-dev/velad/pkg/apis"
	. "github.com/oam-dev/velad/pkg/resources"
	"github.com/oam-dev/velad/pkg/utils"
	"github.com/pkg/errors"
	"github.com/rancher/k3d/v5/pkg/client"
	k3dClient "github.com/rancher/k3d/v5/pkg/client"
	config "github.com/rancher/k3d/v5/pkg/config/v1alpha3"
	"github.com/rancher/k3d/v5/pkg/runtimes"
	k3d "github.com/rancher/k3d/v5/pkg/types"
)

var (
	DefaultHandler Handler = &DockerHandler{}
	info                   = utils.Info
	errf                   = utils.Errf
)

type DockerHandler struct {
}

func (d DockerHandler) Install(args apis.InstallArgs) error {
	err := setupK3d(args)
	if err != nil {
		return errors.Wrap(err, "failed to setup k3d")
	}
	info("Successfully setup cluster")
	return nil
}

func (d DockerHandler) Uninstall() error {
	return nil
}

func (d DockerHandler) GenKubeconfig(bindIP string) error {
	return nil
}

func (d DockerHandler) PrintKubeConfig(internal, external bool) {

}

func setupK3d(args apis.InstallArgs) error {
	info("Preparing k3d images...")
	err := PrepareK3sImages()
	if err != nil {
		return errors.Wrap(err, "failed to prepare k3d images")
	}
	info("Successfully prepare k3d images")

	info("Extracting k3d images...")
	err = LoadK3dImages()
	if err != nil {
		return errors.Wrap(err, "failed to extract k3d images")
	}
	info("Successfully extract k3d images")

	info("Creating k3d cluster...")
	cfg := GetClusterRunConfig(args)
	ctx := context.Background()
	runClusterIfNotExist(ctx, cfg)
	info("Successfully create k3d cluster")

	return nil
}

func GetClusterRunConfig(args apis.InstallArgs) config.ClusterConfig {
	cluster := getClusterConfig(args.DBEndpoint, args.Token)
	createOpts := getClusterCreateOpts()
	kubeconfigOpts := getKubeconfigOptions()
	runConfig := config.ClusterConfig{
		Cluster:           cluster,
		ClusterCreateOpts: createOpts,
		KubeconfigOpts:    kubeconfigOpts,
	}
	return runConfig

}
func getClusterCreateOpts() k3d.ClusterCreateOpts {
	clusterCreateOpts := k3d.ClusterCreateOpts{
		GlobalLabels: map[string]string{}, // empty init
		GlobalEnv:    []string{},          // empty init
	}

	// ensure, that we have the default object labels
	for k, v := range k3d.DefaultRuntimeLabels {
		clusterCreateOpts.GlobalLabels[k] = v
	}

	return clusterCreateOpts
}

// getClusterConfig will get different k3d.Cluster based on ordinal , storage for external storage, token is needed if storage is set
func getClusterConfig(endpoint, token string) k3d.Cluster {
	// All cluster will be created in one docker network
	universalK3dNetwork := k3d.ClusterNetwork{
		Name:     fmt.Sprintf("%s-%s", "k3d", "velad"),
		External: false,
	}

	// api
	kubeAPIExposureOpts := k3d.ExposureOpts{
		Host: k3d.DefaultAPIHost,
	}
	kubeAPIExposureOpts.Port = k3d.DefaultAPIPort
	kubeAPIExposureOpts.Binding = nat.PortBinding{
		HostIP:   k3d.DefaultAPIHost,
		HostPort: "6443",
	}

	// fill cluster config
	clusterName := "velad-cluster-control-plane"
	clusterConfig := k3d.Cluster{
		Name:    clusterName,
		Network: universalK3dNetwork,
		KubeAPI: &kubeAPIExposureOpts,
	}

	// klog.Info("disabling load balancer")

	// nodes
	clusterConfig.Nodes = []*k3d.Node{}

	k3sImageDir, err := getK3sImageDir()
	if err != nil {
		errf("failed to get k3s image dir: %v", err)
	}
	serverNode := k3d.Node{
		Name:       client.GenerateNodeName(clusterConfig.Name, k3d.ServerRole, 0),
		Role:       k3d.ServerRole,
		Image:      "rancher/k3s:latest",
		ServerOpts: k3d.ServerOpts{},
		Volumes:    []string{k3sImageDir + ":/var/lib/rancher/k3s/agent/images/"},
	}

	// use external storage in control plane if set
	serverNode.Args = convertToNodeArgs(endpoint, token)
	clusterConfig.Nodes = append(clusterConfig.Nodes, &serverNode)

	return clusterConfig
}

func getKubeconfigOptions() config.SimpleConfigOptionsKubeconfig {
	opts := config.SimpleConfigOptionsKubeconfig{
		UpdateDefaultKubeconfig: true,
		SwitchCurrentContext:    true,
	}
	return opts
}

func convertToNodeArgs(endpoint, token string) []string {
	var res []string
	res = append(res, "--token="+token)
	if endpoint != "" {
		res = append(res, "--datastore-endpoint="+endpoint)
	}
	return res
}

func runClusterIfNotExist(ctx context.Context, cluster config.ClusterConfig) {
	if _, err := k3dClient.ClusterGet(ctx, runtimes.SelectedRuntime, &cluster.Cluster); err == nil {
		info("Detect an existing cluster: ", cluster.Cluster.Name)
		return
	}
	err := k3dClient.ClusterRun(ctx, runtimes.SelectedRuntime, &cluster)
	if err != nil {
		errf("Fail to create cluster: %s, err: %v", cluster.Cluster.Name, err)
		return
	}
}

func PrepareK3sImages() error {
	embedK3sImage, err := K3sImage.Open("static/k3s/images/k3s-airgap-images-amd64.tar.gz")
	if err != nil {
		return err
	}
	defer utils.CloseQuietly(embedK3sImage)

	// save k3s image.tgz to ~/.vela/velad/k3s/images.tgz
	k3sImagesDir, err := getK3sImageDir()
	k3sImagesPath := filepath.Join(k3sImagesDir, "k3s-airgap-images-amd64.tgz")
	k3sImagesFile, err := os.OpenFile(k3sImagesPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer utils.CloseQuietly(k3sImagesFile)
	if _, err := io.Copy(k3sImagesFile, embedK3sImage); err != nil {
		return err
	}

	/* #nosec */
	info("Successfully prepare k3s image: ", k3sImagesPath)
	return nil
}

func getK3sImageDir() (string, error) {
	dir, err := system.GetVelaHomeDir()
	if err != nil {
		return "", err
	}
	k3sImagesDir := filepath.Join(dir, "k3s")
	if err := os.MkdirAll(k3sImagesDir, 0755); err != nil {
		return "", err
	}
	return k3sImagesDir, nil
}

func LoadK3dImages() error {
	dir, err := K3dImage.ReadDir("static/k3d/images")
	if err != nil {
		return err
	}
	for _, entry := range dir {
		file, err := K3dImage.Open(path.Join("static/k3d/images", entry.Name()))
		if err != nil {
			return err
		}
		name := strings.Split(entry.Name(), ".")[0]
		imageTar, err := utils.SaveToTemp(file, "k3d-image-"+name+"-*.tar")
		if err != nil {
			return err
		}
		importCmd := exec.Command("docker", "image", "load", imageTar)
		output, err := importCmd.CombinedOutput()
		fmt.Print(string(output))
		if err != nil {
			return err
		}
		fmt.Println("Successfully load image: ", imageTar)
	}

	return nil
}
