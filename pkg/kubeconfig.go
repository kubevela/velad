package pkg

import (
	"os"
)

func PrintKubeConfig(internal, external bool) {
	if internal {
		info(KubeConfigLocation)
	}
	if external {
		info(ExternalKubeConfigLocation)
	}
	info("internal kubeconfig: ", KubeConfigLocation)
	_, err := os.Stat(ExternalKubeConfigLocation)
	if err == nil {
		info("external kubeconfig: ", KubeConfigLocation)
	}
}
