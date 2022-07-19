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
	k3dClient "github.com/k3d-io/k3d/v5/pkg/client"
	config "github.com/k3d-io/k3d/v5/pkg/config/v1alpha4"
	"github.com/k3d-io/k3d/v5/pkg/runtimes"
	k3d "github.com/k3d-io/k3d/v5/pkg/types"
	"github.com/oam-dev/kubevela/pkg/utils/system"
	"github.com/oam-dev/velad/pkg/apis"
	"github.com/oam-dev/velad/pkg/resources"
	"github.com/oam-dev/velad/pkg/utils"
	"github.com/pkg/errors"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	// DefaultHandler is the default handler for k3d cluster
	DefaultHandler Handler = &K3dHandler{
		ctx: context.Background(),
	}
	dockerCli client.APIClient
	info      = utils.Info
	infof     = utils.Infof
	errf      = utils.Errf
)

type k3dSetupOptions struct {
	dryRun bool
}

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
	var err error
	d.cfg, err = GetClusterRunConfig(args)
	if err != nil {
		return err
	}
	o := k3dSetupOptions{
		dryRun: args.DryRun,
	}
	err = o.setupK3d(d.ctx, d.cfg)
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
func (d *K3dHandler) GenKubeconfig(ctx apis.Context, bindIP string) error {
	var err error
	var cluster = d.cfg.Cluster.Name
	// 1. kubeconfig for access from host
	cfgHost := configPath(cluster)
	info("Generating host kubeconfig into", cfgHost)
	if !ctx.DryRun {
		if _, err := k3dClient.KubeconfigGetWrite(context.Background(), runtimes.SelectedRuntime, &d.cfg.Cluster, cfgHost,
			&k3dClient.WriteKubeConfigOptions{UpdateExisting: true, OverwriteExisting: false, UpdateCurrentContext: true}); err != nil {
			return errors.Wrap(err, "failed to gen kubeconfig")
		}
	}

	_cfgContent, err := os.ReadFile(cfgHost)
	if err != nil {
		return errors.Wrap(err, "read kubeconfig")
	}

	var (
		hostToReplace string
		kubeConfig    = string(_cfgContent)
	)

	if !ctx.DryRun {
		switch {
		case strings.Contains(kubeConfig, "0.0.0.0"):
			hostToReplace = "0.0.0.0"
		case strings.Contains(kubeConfig, "host.docker.internal"):
			hostToReplace = "host.docker.internal"
		default:
			return errors.Wrap(err, "unrecognized kubeconfig format")
		}
	}

	// Replace host config with loop back address
	if !ctx.DryRun {
		cfgHostContent := strings.ReplaceAll(kubeConfig, hostToReplace, "127.0.0.1")
		err = ioutil.WriteFile(cfgHost, []byte(cfgHostContent), 0600)
		if err != nil {
			errf("Fail to re-write host kubeconfig")
		}
	}

	// 2. kubeconfig for access from other VelaD cluster
	// Basically we replace the IP with IP inside the docker network
	cfgIn := configPathInternal(cluster)
	info("Generating internal kubeconfig into", cfgIn)
	if !ctx.DryRun {
		var containerIP string
		networks, err := dockerCli.NetworkInspect(d.ctx, apis.VelaDDockerNetwork, types.NetworkInspectOptions{})
		if err != nil {
			klog.ErrorS(err, "inspect docker network")
			return err
		}
		cs := networks.Containers
		for _, c := range cs {
			if c.Name == fmt.Sprintf("k3d-%s-server-0", d.cfg.Cluster.Name) {
				containerIP = strings.TrimSuffix(c.IPv4Address, "/16")
			}
		}
		re := regexp.MustCompile(hostToReplace + `:\d{4}`)
		cfgInContent := re.ReplaceAllString(kubeConfig, fmt.Sprintf("%s:6443", containerIP))
		err = ioutil.WriteFile(cfgIn, []byte(cfgInContent), 0600)
		if err != nil {
			errf("Fail to write internal kubeconfig")
		} else {
			info("Successfully generate internal kubeconfig at", cfgIn)
		}
	}

	// 3. kubeconfig for access from other machine
	if bindIP != "" {
		cfgOut := configPathExternal(cluster)
		info("Generating external kubeconfig for remote access into ", cfgOut)
		if !ctx.DryRun {
			cfgOutContent := strings.Replace(kubeConfig, hostToReplace, bindIP, 1)
			err = os.WriteFile(cfgOut, []byte(cfgOutContent), 0600)
			if err != nil {
				return err
			}
		}
		info("Successfully generate external kubeconfig at", cfgOut)
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

func (o k3dSetupOptions) setupK3d(ctx context.Context, clusterConfig config.ClusterConfig) error {
	info("Preparing K3s images...")
	err := o.prepareK3sImages()
	if err != nil {
		return errors.Wrap(err, "failed to prepare k3d images")
	}
	info("Successfully prepare k3d images")

	info("Loading k3d images...")
	err = o.loadK3dImages()
	if err != nil {
		return errors.Wrap(err, "failed to extract k3d images")
	}
	info("Successfully load k3d images")

	info("Creating k3d cluster...")
	if err = o.runClusterIfNotExist(ctx, clusterConfig); err != nil {
		return err
	}
	info("Successfully create k3d cluster")
	return nil
}

// GetClusterRunConfig returns the run-config for the k3d cluster
func GetClusterRunConfig(args apis.InstallArgs) (config.ClusterConfig, error) {
	createOpts := getClusterCreateOpts()
	cluster, err := getClusterConfig(args, createOpts)
	if err != nil {
		return config.ClusterConfig{}, err
	}
	kubeconfigOpts := getKubeconfigOptions()
	runConfig := config.ClusterConfig{
		Cluster:           cluster,
		ClusterCreateOpts: createOpts,
		KubeconfigOpts:    kubeconfigOpts,
	}
	return runConfig, nil
}

func getClusterCreateOpts() k3d.ClusterCreateOpts {
	clusterCreateOpts := k3d.ClusterCreateOpts{
		GlobalLabels: map[string]string{}, // empty init
		GlobalEnv:    []string{},          // empty init
		// Enable LoadBalancer for using Ingress to access services
		DisableLoadBalancer: false,
	}

	// ensure, that we have the default object labels
	for k, v := range k3d.DefaultRuntimeLabels {
		clusterCreateOpts.GlobalLabels[k] = v
	}

	return clusterCreateOpts
}

// getClusterConfig will get different k3d.Cluster based on ordinal , storage for external storage, token is needed if storage is set
func getClusterConfig(args apis.InstallArgs, ops k3d.ClusterCreateOpts) (k3d.Cluster, error) {
	// Cluster will be created in one docker network
	var universalK3dNetwork = k3d.ClusterNetwork{
		Name:     apis.VelaDDockerNetwork,
		External: false,
	}
	kubeAPIExposureOpts := k3d.ExposureOpts{
		Host: k3d.DefaultAPIHost,
	}
	port, err := findAvailablePort(6443)
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
	var nodes []*k3d.Node

	// load-balancer for servers

	clusterConfig.ServerLoadBalancer = prepareLoadbalancer(clusterConfig, ops)
	nodes = append(nodes, clusterConfig.ServerLoadBalancer.Node)

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
	nodes = append(nodes, &serverNode)
	clusterConfig.Nodes = nodes

	clusterConfig.ServerLoadBalancer.Config.Ports[fmt.Sprintf("%s.tcp", k3d.DefaultAPIPort)] = append(clusterConfig.ServerLoadBalancer.Config.Ports[fmt.Sprintf("%s.tcp", k3d.DefaultAPIPort)], serverNode.Name)

	// Other configurations
	portWithFilter, err := getPortWithFilters()
	if err != nil {
		return clusterConfig, errors.Wrap(err, "failed to get http ports")
	}
	err = k3dClient.TransformPorts(context.Background(), runtimes.SelectedRuntime, &clusterConfig, []config.PortWithNodeFilters{portWithFilter})
	if err != nil {
		return clusterConfig, errors.Wrap(err, "failed to transform ports")
	}

	return clusterConfig, nil
}

func getKubeconfigOptions() config.SimpleConfigOptionsKubeconfig {
	// TODO: this not working yet, we are updating kubeconfig manually
	opts := config.SimpleConfigOptionsKubeconfig{
		UpdateDefaultKubeconfig: true,
		SwitchCurrentContext:    true,
	}
	return opts
}

func (o k3dSetupOptions) runClusterIfNotExist(ctx context.Context, cluster config.ClusterConfig) error {
	var err error
	info("Launching k3d cluster:", cluster.Cluster.Name)
	if !o.dryRun {
		if _, err = k3dClient.ClusterGet(ctx, runtimes.SelectedRuntime, &cluster.Cluster); err == nil {
			info("Detect an existing cluster: ", cluster.Cluster.Name)
			return nil
		}
		err = k3dClient.ClusterRun(ctx, runtimes.SelectedRuntime, &cluster)
	}
	return errors.Wrapf(err, "fail to create cluster: %s", cluster.Cluster.Name)
}

// prepareK3sImages extracts k3s images to ~/.vela/velad/k3s/images.tg
func (o k3dSetupOptions) prepareK3sImages() error {
	embedK3sImage, err := resources.K3sImage.Open("static/k3s/images/k3s-airgap-images.tar.gz")
	if err != nil {
		return err
	}
	defer utils.CloseQuietly(embedK3sImage)

	k3sImagesDir, err := getK3sImageDir()
	if err != nil {
		return err
	}
	k3sImagesPath := filepath.Join(k3sImagesDir, "k3s-airgap-images.tgz")
	info("Saving k3s image airgap install tarball to", k3sImagesPath)

	if !o.dryRun {
		// #nosec
		k3sImagesFile, err := os.OpenFile(k3sImagesPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}
		defer utils.CloseQuietly(k3sImagesFile)
		if _, err := io.Copy(k3sImagesFile, embedK3sImage); err != nil {
			return err
		}
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

// loadK3dImages loads local k3d images to docker
func (o k3dSetupOptions) loadK3dImages() error {
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
		var (
			format   = "k3d-image-" + name + "-*.tar.gz"
			imageTgz string
			imageTar string
		)
		if o.dryRun {
			info("Saving and temporary image file:", format)
		} else {
			imageTgz, err = utils.SaveToTemp(file, format)
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
			imageTar = strings.TrimSuffix(imageTgz, ".gz")
		}

		if o.dryRun {
			infof("Importing image to docker using temporary file: %s\n", format)
		} else {
			// #nosec
			importCmd := exec.Command("docker", "image", "load", "-i", imageTar)
			output, err := importCmd.CombinedOutput()
			utils.InfoBytes(output)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// findAvailablePort find available port, start by default
func findAvailablePort(start int) (string, error) {
	for i := start; i < 65535; i++ {
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", i))
		if err != nil {
			continue
		}
		utils.CloseQuietly(listener)
		return strconv.Itoa(i), nil
	}
	return "", errors.New("no available port")
}

func prepareLoadbalancer(cluster k3d.Cluster, opts k3d.ClusterCreateOpts) *k3d.Loadbalancer {
	lb := k3d.NewLoadbalancer()

	labels := map[string]string{}
	if opts.GlobalLabels == nil && len(opts.GlobalLabels) == 0 {
		labels = opts.GlobalLabels
	}

	lb.Node.Name = fmt.Sprintf("%s-%s-serverlb", k3d.DefaultObjectNamePrefix, cluster.Name)
	lb.Node.Image = apis.K3dImageProxy
	lb.Node.Ports = nat.PortMap{
		k3d.DefaultAPIPort: []nat.PortBinding{cluster.KubeAPI.Binding},
	}
	lb.Node.Networks = []string{cluster.Network.Name}

	// fixed the lb image
	lb.Node.RuntimeLabels = labels
	lb.Node.Restart = true

	return lb
}

func getPortWithFilters() (config.PortWithNodeFilters, error) {
	var port config.PortWithNodeFilters
	hostPort, err := findAvailablePort(8090)
	if err != nil {
		return port, err
	}
	port.Port = fmt.Sprintf("%s:80", hostPort)
	port.NodeFilters = []string{"loadbalancer"}
	return port, nil
}
