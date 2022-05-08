package cmd

import (
	"runtime"

	"github.com/fatih/color"
	"github.com/oam-dev/velad/pkg/apis"
)

var (
	red            = color.New(color.FgRed).SprintFunc()
	green          = color.New(color.FgGreen).SprintFunc()
	yellow         = color.New(color.FgYellow).SprintFunc()
	k3dImageStatus = map[string]bool{}
	x              = red("✘")
	y              = green("✔")
	ar             = yellow("➤")
)

// PrintClusterStatus helps print cluster status
func PrintClusterStatus(status apis.ClusterStatus) bool {
	switch runtime.GOOS {
	case "linux":
		return printClusterStatusK3s(status)
	default:
		return printClusterStatusK3d(status)
	}

}

func printClusterStatusK3d(status apis.ClusterStatus) bool {
	infoP(0, "K3d images status:")
	if status.K3dImages.Reason != "" {
		info(x, "K3d images:", status.K3dImages.Reason)
		return true // k3d images not ready
	}
	k3dImageStatus[apis.K3dImageK3s] = status.K3dImages.K3s
	k3dImageStatus[apis.K3dImageTools] = status.K3dImages.K3dTools
	k3dImageStatus[apis.K3dImageProxy] = status.K3dImages.K3dProxy
	stop := false
	for i, imageStatus := range k3dImageStatus {
		stop = stop || !imageStatus
		if !imageStatus {
			infoP(1, x, "image", i, "not ready")
		} else {
			infoP(1, y, "image", i, "ready")
		}
	}
	if stop {
		return stop
	}
	infoP(0, "Cluster(K3d) status:")
	if status.K3d.Reason != "" {
		info(x, "K3d:", status.K3d.Reason)
		return true // k3d not ready
	}
	for _, c := range status.K3d.K3dContainer {
		if c.Reason != "" {
			infoP(1, x, "cluster", "["+c.Name+"]", "not ready:", c.Reason)
			stop = true
		} else {
			infoP(1, y, "cluster", "["+c.Name+"]", "ready")
			if c.VelaStatus != apis.StatusVelaDeployed {
				infoP(2, ar, "kubevela status:", c.VelaStatus)
			} else {
				infoP(2, y, "kubevela status:", c.VelaStatus)
			}
		}
	}
	if stop {
		return stop
	}

	return false
}

func printClusterStatusK3s(status apis.ClusterStatus) bool {
	infoP(0, "K3s images status:")
	if status.Reason != "" {
		info(x, "Check K3s status:", status.Reason)
	}
	if status.K3s.K3sBinary {
		infoP(1, y, "k3s binary:", "ready")
	} else {
		infoP(1, x, "k3s binary:", "not ready")
		return true
	}
	if status.K3s.K3sServiceStatus != "" {
		infoP(1, y, "k3s service status:", status.K3s.K3sServiceStatus)
	} else {
		infoP(1, x, "k3s service status:", "not found")
		return true
	}
	return false
}

// PrintVelaStatus helps print kubevela status
func PrintVelaStatus(status apis.VelaStatus) {
	infoP(0, "Vela status:")
	if status.VelaCLIInstalled {
		infoP(1, y, "Vela CLI installed")
		infoP(1, y, "Vela CLI path:", status.VelaCLIPath)
	} else {
		infoP(1, x, "Vela CLI not installed")
	}
	if status.VelaUXAddonDirPresent {
		infoP(1, y, "VelaUX addon dir ready")
		infoP(1, y, "VelaUX addon dir path:", status.VelaUXAddonDirPath)
	} else {
		infoP(1, x, "VelaUX addon dir not ready")
	}
	if status.Reason != "" {
		info(x, "Check status err:", status.Reason)
	}

}
