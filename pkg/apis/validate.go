package apis

import (
	"runtime"

	"github.com/pkg/errors"
)

var newErr = errors.New

// Validate validates the `install` argument
func (a InstallArgs) Validate() error {
	if runtime.GOOS == GoosLinux {
		if a.Name != DefaultVelaDClusterName {
			return newErr("name flag not works in linux")
		}
	}
	return nil
}

// Validate validates the `kubeconfig` argument
func (a KubeconfigArgs) Validate() error {
	if runtime.GOOS == GoosLinux {
		if a.Name != DefaultVelaDClusterName {
			return newErr("name flag not works in linux")
		}
		if a.Internal {
			return newErr("internal flag not work in linux")
		}
	}
	return nil
}

// Validate validates the uninstall arguments
func (a UninstallArgs) Validate() error {
	if runtime.GOOS == GoosLinux {
		if a.Name != DefaultVelaDClusterName {
			return newErr("name flag not works in linux")
		}
	}
	return nil
}

// Validate validates the token arguments
func (a TokenArgs) Validate() error {
	if runtime.GOOS == GoosLinux {
		if a.Name != DefaultVelaDClusterName {
			return newErr("name flag not works in linux")
		}
	}
	return nil
}
