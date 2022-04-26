//go:build !linux

package handler

import (
	"github.com/oam-dev/velad/pkg/apis"
)

var (
	DefaultHandler Handler = &DockerHandler{}
)

type DockerHandler struct {
}

func (d DockerHandler) Install(args apis.InstallArgs) error {
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
