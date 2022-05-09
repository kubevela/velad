//go:build !linux

package cluster

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"helm.sh/helm/v3/pkg/action"
	"k8s.io/klog/v2"

	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/oam-dev/kubevela/pkg/utils/system"
	"github.com/oam-dev/velad/pkg/apis"
	"github.com/oam-dev/velad/pkg/resources"
	"github.com/oam-dev/velad/pkg/utils"
	"github.com/pkg/errors"
	k3dClient "github.com/rancher/k3d/v5/pkg/client"
	config "github.com/rancher/k3d/v5/pkg/config/v1alpha3"
	"github.com/rancher/k3d/v5/pkg/runtimes"
	k3d "github.com/rancher/k3d/v5/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	// DefaultHandler is the default handler for k3d cluster
	DefaultHandler Handler = &K3dHandler{
		ctx: context.Background(),
	}
	dockerCli client.APIClient
	info      = utils.Info
	errf      = utils.Errf
)

const (
	// K3dImageTag is image tag of k3d
	K3dImageTag = "v1.21.10-k3s1"
)

func init() {
	var err error
	dockerCli, err = client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
}

// K3dHandler will handle the k3d cluster creation and management
type K3dHandler struct {
	ctx context.Context
	cfg config.ClusterConfig
}

// Install will install a k3d cluster
func (d *K3dHandler) Install(args apis.InstallArgs) error {
	d.cfg = GetClusterRunConfig(args)
	err := setupK3d(d.ctx, d.cfg)
	if err != nil {
		return errors.Wrap(err, "failed to setup k3d")
	}
	info("Successfully setup cluster")
	return nil
}

// Uninstall removes a k3d cluster of certain name
func (d *K3dHandler) Uninstall(name string) error {
	clusterList, err := k3dClient.ClusterList(d.ctx, runtimes.SelectedRuntime)
	if err != nil {
		return errors.Wrap(err, "failed to get cluster list")
	}

	if len(clusterList) == 0 {
		return errors.New("no cluster found")
	}

	var veladCluster *k3d.Cluster

	for _, c := range clusterList {
		if c.Name == fmt.Sprintf("velad-cluster-%s", name) {
			veladCluster = c
		}
	}

	err = k3dClient.ClusterDelete(d.ctx, runtimes.SelectedRuntime, veladCluster, k3d.ClusterDeleteOpts{
		SkipRegistryCheck: false,
	})
	if err != nil {
		return errors.Wrap(err, "Fail to delete cluster")
	}
	// TODO: delete Kubeconfig
	return nil
}

// GenKubeconfig generate three kinds of kubeconfig
// 1. kubeconfig for access from host
// 2. kubeconfig for access from other VelaD cluster
// 3. kubeconfig for access from other machine (if bindIP provided)
func (d *K3dHandler) GenKubeconfig(bindIP string) error {
	var err error
	var cluster = d.cfg.Cluster.Name
	// 1. kubeconfig for access from host
	cfg := configPath(cluster)
	if _, err := k3dClient.KubeconfigGetWrite(context.Background(), runtimes.SelectedRuntime, &d.cfg.Cluster, cfg,
		&k3dClient.WriteKubeConfigOptions{UpdateExisting: true, OverwriteExisting: false, UpdateCurrentContext: true}); err != nil {
		return errors.Wrap(err, "failed to gen kubeconfig")
	}

	// 2. kubeconfig for access from other VelaD cluster
	// Basically we replace the IP with IP inside the docker network
	cfgContent, err := os.ReadFile(cfg)
	if err != nil {
		return errors.Wrap(err, "read kubeconfig")
	}
	cfgIn := configPathInternal(cluster)
	networks, err := dockerCli.NetworkInspect(d.ctx, apis.VelaDDockerNetwork, types.NetworkInspectOptions{})
	if err != nil {
		klog.ErrorS(err, "inspect docker network")
		return err
	}
	var containerIP string
	cs := networks.Containers
	for _, c := range cs {
		if c.Name == fmt.Sprintf("k3d-%s-server-0", d.cfg.Cluster.Name) {
			containerIP = strings.TrimSuffix(c.IPv4Address, "/16")
		}
	}
	kubeConfig := string(cfgContent)
	re := regexp.MustCompile(`0.0.0.0:\d{4}`)
	cfgInContent := re.ReplaceAllString(kubeConfig, fmt.Sprintf("%s:6443", containerIP))
	err = ioutil.WriteFile(cfgIn, []byte(cfgInContent), 0600)
	if err != nil {
		return err
	}

	// 3. kubeconfig for access from other machine
	if bindIP != "" {
		cfgOut := configPathExternal(cluster)
		info("Generating kubeconfig for remote access into ", cfgOut)
		originConf, err := os.ReadFile(cfg)
		if err != nil {
			return err
		}
		newConf := strings.Replace(string(originConf), "0.0.0.0", bindIP, 1)
		err = os.WriteFile(cfgOut, []byte(newConf), 0600)
		if err != nil {
			return err
		}
		info("Successfully generate kubeconfig at ", cfgOut)
	}
	return nil
}

// SetKubeconfig set kubeconfig environment of cluster stored in K3dHandler
func (d *K3dHandler) SetKubeconfig() error {
	info("Setting kubeconfig env for VelaD...")
	return os.Setenv("KUBECONFIG", configPath(d.cfg.Cluster.Name))
}

// LoadImage loads image from local path
func (d *K3dHandler) LoadImage(image string) error {
	err := k3dClient.ImageImportIntoClusterMulti(d.ctx, runtimes.SelectedRuntime, []string{image}, &d.cfg.Cluster, k3d.ImageImportOpts{})
	return errors.Wrap(err, "failed to import image")
}

// GetStatus returns the status of the cluster
func (d *K3dHandler) GetStatus() apis.ClusterStatus {
	var status apis.ClusterStatus
	list, err := dockerCli.ImageList(d.ctx, types.ImageListOptions{})

	if err != nil {
		status.K3dImages.Reason = fmt.Sprintf("Failed to get image list: %s", err.Error())
		return status
	}
	for _, image := range list {
		fillK3dImageStatus(image, &status)
	}

	clusters, err := k3dClient.ClusterList(d.ctx, runtimes.SelectedRuntime)
	if err != nil {
		status.K3d.Reason = fmt.Sprintf("Failed to get cluster list: %s", err.Error())
		return status
	}
	status.K3d.K3dContainer = []apis.K3dContainer{}
	for _, cluster := range clusters {
		fillK3dCluster(d.ctx, cluster, &status)
	}
	return status
}

func fillK3dImageStatus(image types.ImageSummary, status *apis.ClusterStatus) {
	if len(image.RepoTags) == 0 {
		return
	}
	for _, tag := range image.RepoTags {
		switch tag {
		case apis.K3dImageK3s:
			status.K3dImages.K3s = true
		case apis.K3dImageTools:
			status.K3dImages.K3dTools = true
		case apis.K3dImageProxy:
			status.K3dImages.K3dProxy = true
		}
	}
}

func fillK3dCluster(ctx context.Context, cluster *k3d.Cluster, status *apis.ClusterStatus) {
	if strings.HasPrefix(cluster.Name, "velad-cluster-") {
		container := apis.K3dContainer{
			Name:    strings.TrimPrefix(cluster.Name, "velad-cluster-"),
			Running: true,
		}

		// get k3d cluster kubeconfig
		kubeconfig, err := k3dClient.KubeconfigGet(ctx, runtimes.SelectedRuntime, cluster)
		if err != nil {
			container.Reason = fmt.Sprintf("Failed to get kubeconfig: %s", err.Error())
		}
		restConfig, err := clientcmd.NewDefaultClientConfig(*kubeconfig, nil).ClientConfig()
		if err != nil {
			container.Reason = fmt.Sprintf("Failed to get rest kubeconfig: %s", err.Error())
		}
		cfg, err := utils.NewActionConfig(restConfig, false)
		if err != nil {
			container.Reason = fmt.Sprintf("Failed to get helm action config: %s", err.Error())
		}
		list := action.NewList(cfg)
		list.SetStateMask()
		releases, err := list.Run()
		if err != nil {
			container.Reason = fmt.Sprintf("Failed to get helm releases: %s", err.Error())
		}
		for _, release := range releases {
			if release.Name == apis.KubeVelaHelmRelease {
				container.VelaStatus = release.Info.Status.String()
			}
		}
		if container.VelaStatus == "" {
			container.VelaStatus = apis.StatusVelaNotInstalled
		}

		status.K3d.K3dContainer = append(status.K3d.K3dContainer, container)
	}
}

func setupK3d(ctx context.Context, clusterConfig config.ClusterConfig) error {
	info("Preparing K3s images...")
	err := PrepareK3sImages()
	if err != nil {
		return errors.Wrap(err, "failed to prepare k3d images")
	}
	info("Successfully prepare k3d images")

	info("Loading k3d images...")
	err = LoadK3dImages()
	if err != nil {
		return errors.Wrap(err, "failed to extract k3d images")
	}
	info("Successfully load k3d images")

	info("Creating k3d cluster...")
	if err = runClusterIfNotExist(ctx, clusterConfig); err != nil {
		return err
	}
	info("Successfully create k3d cluster")
	return nil
}

// GetClusterRunConfig returns the run-config for the k3d cluster
func GetClusterRunConfig(args apis.InstallArgs) config.ClusterConfig {
	cluster := getClusterConfig(args)
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
		// TODO: for now, we disable the LB, but we can enable it later for multi-server scenario
		DisableLoadBalancer: true,
	}

	// ensure, that we have the default object labels
	for k, v := range k3d.DefaultRuntimeLabels {
		clusterCreateOpts.GlobalLabels[k] = v
	}

	return clusterCreateOpts
}

// getClusterConfig will get different k3d.Cluster based on ordinal , storage for external storage, token is needed if storage is set
func getClusterConfig(args apis.InstallArgs) k3d.Cluster {
	// Cluster will be created in one docker network
	var universalK3dNetwork = k3d.ClusterNetwork{
		Name:     apis.VelaDDockerNetwork,
		External: false,
	}
	kubeAPIExposureOpts := k3d.ExposureOpts{
		Host: k3d.DefaultAPIHost,
	}
	port, err := findAvailablePort()
	if err != nil {
		panic(err)
	}
	kubeAPIExposureOpts.Port = k3d.DefaultAPIPort
	kubeAPIExposureOpts.Binding = nat.PortBinding{
		HostIP:   k3d.DefaultAPIHost,
		HostPort: port,
	}

	// fill cluster config
	clusterName := fmt.Sprintf("velad-cluster-%s", args.Name)
	clusterConfig := k3d.Cluster{
		Name:    clusterName,
		Network: universalK3dNetwork,
		KubeAPI: &kubeAPIExposureOpts,
	}

	// nodes
	clusterConfig.Nodes = []*k3d.Node{}

	k3sImageDir, err := getK3sImageDir()
	if err != nil {
		errf("failed to get k3s image dir: %v", err)
	}
	serverNode := k3d.Node{
		Name:       k3dClient.GenerateNodeName(clusterConfig.Name, k3d.ServerRole, 0),
		Role:       k3d.ServerRole,
		Image:      fmt.Sprintf("rancher/k3s:%s", K3dImageTag),
		ServerOpts: k3d.ServerOpts{},
		Volumes:    []string{k3sImageDir + ":/var/lib/rancher/k3s/agent/images/"},
	}

	serverNode.Args = GetK3sServerArgs(args)
	clusterConfig.Nodes = append(clusterConfig.Nodes, &serverNode)

	return clusterConfig
}

func getKubeconfigOptions() config.SimpleConfigOptionsKubeconfig {
	// TODO: this not working yet, we are updating kubeconfig manually
	opts := config.SimpleConfigOptionsKubeconfig{
		UpdateDefaultKubeconfig: true,
		SwitchCurrentContext:    true,
	}
	return opts
}

func runClusterIfNotExist(ctx context.Context, cluster config.ClusterConfig) error {
	if _, err := k3dClient.ClusterGet(ctx, runtimes.SelectedRuntime, &cluster.Cluster); err == nil {
		info("Detect an existing cluster: ", cluster.Cluster.Name)
		return nil
	}
	err := k3dClient.ClusterRun(ctx, runtimes.SelectedRuntime, &cluster)
	return errors.Wrapf(err, "fail to create cluster: %s", cluster.Cluster.Name)
}

// PrepareK3sImages extracts k3s images to ~/.vela/velad/k3s/images.tg
func PrepareK3sImages() error {
	embedK3sImage, err := resources.K3sImage.Open("static/k3s/images/k3s-airgap-images-amd64.tar.gz")
	if err != nil {
		return err
	}
	defer utils.CloseQuietly(embedK3sImage)

	// save k3s image.tgz to ~/.vela/velad/k3s/images.tgz
	k3sImagesDir, err := getK3sImageDir()
	if err != nil {
		return err
	}
	k3sImagesPath := filepath.Join(k3sImagesDir, "k3s-airgap-images-amd64.tgz")
	// #nosec
	k3sImagesFile, err := os.OpenFile(k3sImagesPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
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
	k3sImagesDir := filepath.Join(dir, "velad", "k3s")
	if err := os.MkdirAll(k3sImagesDir, 0700); err != nil {
		return "", err
	}
	return k3sImagesDir, nil
}

// LoadK3dImages loads local k3d images to docker
func LoadK3dImages() error {
	dir, err := resources.K3dImage.ReadDir("static/k3d/images")
	if err != nil {
		return err
	}
	for _, entry := range dir {
		file, err := resources.K3dImage.Open(path.Join("static/k3d/images", entry.Name()))
		if err != nil {
			return err
		}
		name := strings.Split(entry.Name(), ".")[0]
		imageTgz, err := utils.SaveToTemp(file, "k3d-image-"+name+"-*.tar.gz")
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
		// #nosec
		importCmd := exec.Command("docker", "image", "load", "-i", imageTar)
		output, err = importCmd.CombinedOutput()
		utils.InfoBytes(output)
		if err != nil {
			return err
		}
	}

	return nil
}

// find available port, 6443 by default
func findAvailablePort() (string, error) {
	for i := 6443; i < 65535; i++ {
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", i))
		if err != nil {
			continue
		}
		utils.CloseQuietly(listener)
		return strconv.Itoa(i), nil
	}
	return "", errors.New("no available port")
}
