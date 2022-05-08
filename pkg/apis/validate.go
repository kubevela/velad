package apis

import (
	"runtime"

	"github.com/pkg/errors"
)

var newErr = errors.New

// Validate validates the `kubeconfig` argument
func (k KubeconfigArgs) Validate() error {
	if runtime.GOOS == "linux" {
		if k.Name != "default" {
			return newErr("name flag not works in linux")
		}
		if k.Internal {
			return newErr("internal flag not work in linux")
		}
	}
	return nil
}

// Validate validates the uninstall arguments
func (u UninstallArgs) Validate() error {
	if runtime.GOOS == "linux" {
		if u.Name != "default" {
			return newErr("name flag not works in linux")
		}
	}
	return nil
}
