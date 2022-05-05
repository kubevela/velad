package apis

import (
	"github.com/pkg/errors"
	"runtime"
)

var newErr = errors.New

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

func (u UninstallArgs) Validate() error {
	if runtime.GOOS == "linux" {
		if u.Name != "default" {
			return newErr("name flag not works in linux")
		}
	}
	return nil
}
