package utils

import (
	"fmt"
	cmdutil "github.com/oam-dev/kubevela/pkg/utils/util"
	"helm.sh/helm/v3/pkg/action"
	"k8s.io/client-go/rest"
	"os"
)

func NewActionConfig(config *rest.Config, showDetail bool) (*action.Configuration, error) {
	cfg := new(action.Configuration)
	restClientGetter := cmdutil.NewRestConfigGetterByConfig(config, "")
	log := func(format string, a ...interface{}) {
		if showDetail {
			fmt.Printf(format+"\n", a...)
		}
	}
	err := cfg.Init(restClientGetter, "", os.Getenv("HELM_DRIBVER"), log)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
