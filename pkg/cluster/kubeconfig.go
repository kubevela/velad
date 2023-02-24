package cluster

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/oam-dev/velad/pkg/apis"
	"github.com/oam-dev/velad/pkg/utils"
)

// PrintKubeConfig helps print kubeconfig locations
func PrintKubeConfig(args apis.KubeconfigArgs) error {
	switch runtime.GOOS {
	case apis.GoosDarwin, apis.GoosWindows:
		return printKubeConfigDocker(args)
	case apis.GoosLinux:
		return printKubeConfigLinux(args)
	default:
		utils.UnsupportedOS(runtime.GOOS)
	}
	return nil
}

func printKubeConfigLinux(args apis.KubeconfigArgs) error {
	if args.Host {
		info(apis.K3sKubeConfigLocation)
		return nil
	}
	if args.External {
		info(apis.K3sExternalKubeConfigLocation)
		return nil
	}
	if args.LB {
		info(apis.K3sLBKubeconfigLocation)
		return nil
	}
	info("internal kubeconfig: ", apis.K3sKubeConfigLocation)
	_, err := os.Stat(apis.K3sExternalKubeConfigLocation)
	if err == nil {
		info("external kubeconfig: ", apis.K3sExternalKubeConfigLocation)
	}
	_, err = os.Stat(apis.K3sLBKubeconfigLocation)
	if err == nil {
		info("kubeconfig to access from load balancer: ", apis.K3sLBKubeconfigLocation)
	}
	return nil
}

func printKubeConfigDocker(args apis.KubeconfigArgs) error {
	clusterName := "velad-cluster-" + args.Name
	if args.Host {
		info(configPath(clusterName))
		return nil
	}
	if args.Internal {
		info(configPathInternal(clusterName))
		return nil
	}
	if args.External {
		info(configPathExternal(clusterName))
		return nil
	}
	if args.LB {
		info(configPathLB(clusterName))
		return nil
	}
	info("host kubeconfig:", configPath(clusterName), "(For accessing from host machine)")
	info("internal kubeconfig:", configPathInternal(clusterName), "(For \"vela cluster join\")")
	cfgExt := configPathExternal(clusterName)
	_, err := os.Stat(cfgExt)
	if err == nil {
		info("external kubeconfig:", configPathExternal(clusterName), "(For accessing from other machines)")
	}
	return nil
}

func configPath(clusterName string) string {
	return filepath.Join(utils.GetKubeconfigDir(), clusterName)
}

func configPathExternal(clusterName string) string {
	return filepath.Join(utils.GetKubeconfigDir(), fmt.Sprintf("%s-external", clusterName))
}

func configPathLB(clusterName string) string {
	return filepath.Join(utils.GetKubeconfigDir(), fmt.Sprintf("%s-lb", clusterName))
}

func configPathInternal(clusterName string) string {
	return filepath.Join(utils.GetKubeconfigDir(), fmt.Sprintf("%s-internal", clusterName))
}
