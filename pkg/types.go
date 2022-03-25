package pkg

import "github.com/oam-dev/kubevela/references/cli"

// CtrlPlaneArgs defines arguments for ctrl-plane command
type CtrlPlaneArgs struct {
	BindIP                    string
	DBEndpoint                string
	IsJoin                    bool
	Token                     string
	DisableWorkloadController bool
	// InstallArgs is parameters passed to vela install command
	InstallArgs cli.InstallArgs
}

