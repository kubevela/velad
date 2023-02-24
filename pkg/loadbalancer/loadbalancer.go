package loadbalancer

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/pkg/errors"
	g "github.com/tufanbarisyildirim/gonginx"

	"github.com/oam-dev/velad/pkg/apis"
	"github.com/oam-dev/velad/pkg/cluster"
	"github.com/oam-dev/velad/pkg/resources"
	"github.com/oam-dev/velad/pkg/utils"
)

var (
	errf = utils.Errf
	info = utils.Info
)

// ConfigureNginx set nginx config file
func ConfigureNginx(args apis.LoadBalancerArgs) error {
	var err error
	err = checkLBCondition()
	if err != nil {
		return err
	}
	err = installNginx()
	if err != nil {
		return err
	}
	confLocation, err := setNginxConf(args)
	if err != nil {
		return err
	}
	return startNginx(confLocation)
}

// UninstallNginx uninstall nginx using package manager
func UninstallNginx() error {
	file, err := resources.Nginx.Open("static/nginx/remove_nginx.sh")
	if err != nil {
		return err
	}
	scriptName, err := utils.SaveToTemp(file, "install_nginx-*.sh")
	if err != nil {
		return err
	}
	// #nosec
	cmd := exec.Command("/bin/bash", scriptName)
	output, err := cmd.CombinedOutput()
	utils.InfoBytes(output)
	if err != nil {
		return err
	}
	return nil
}

func installNginx() error {
	file, err := resources.Nginx.Open("static/nginx/install_nginx.sh")
	if err != nil {
		return err
	}
	scriptName, err := utils.SaveToTemp(file, "install_nginx-*.sh")
	if err != nil {
		return err
	}
	// #nosec
	cmd := exec.Command("/bin/bash", scriptName)
	output, err := cmd.CombinedOutput()
	utils.InfoBytes(output)
	return err
}

func setNginxConf(args apis.LoadBalancerArgs) (string, error) {
	var conf strings.Builder
	clause, err := getNginxStreamModClause()
	if err != nil {
		return "", err
	}
	conf.WriteString(clause)
	other := getOther(args)
	conf.WriteString(other)
	loc, err := writeNginxConf(conf.String(), args.Configuration)
	if err != nil {
		return "", errors.Wrap(err, "write nginx conf")
	}
	return loc, nil
}

func startNginx(conf string) error {
	info("Starting/Restarting nginx")
	cmd := exec.Command("pkill", "-9", "nginx")
	// pkill will return error if nginx is not running, so we ignore it
	output, _ := cmd.CombinedOutput()
	utils.InfoBytes(output)
	// wait for nginx to stop
	time.Sleep(1 * time.Second)
	// #nosec
	reloadCmd := exec.Command("nginx", "-c", conf)
	output, err := reloadCmd.CombinedOutput()
	utils.InfoBytes(output)
	return errors.Wrap(err, "fail to start nginx")
}

func writeNginxConf(conf string, confLocation string) (string, error) {
	var err error
	loc := confLocation
	if loc == "" {
		loc, err = getNginxDefaultConfLoc()
		if err != nil {
			return "", errors.Wrap(err, "locate default config fail, please try specify with -c")
		}
	}
	// #nosec
	confFile, err := os.OpenFile(loc, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return "", errors.Wrap(err, "open conf file")
	}
	_, err = confFile.WriteString(conf)
	if err != nil {
		return "", err
	}
	return loc, nil
}

func getNginxStreamModClause() (string, error) {
	var modLoc string
	for _, loc := range []string{
		"/usr/lib/nginx/modules/ngx_stream_module.so",
		"/usr/lib64/nginx/modules/ngx_stream_module.so",
	} {
		if _, err := os.Stat(loc); err == nil {
			modLoc = loc
			break
		}
	}
	if modLoc != "" {
		return fmt.Sprintf("load_module %s;\n", modLoc), nil
	}
	return "", errors.New("Nginx stream mod lib not found")
}

func getOther(args apis.LoadBalancerArgs) string {
	hosts := args.Hosts
	type streamPort struct {
		from int
		to   int
	}
	streamBlockMap := map[string]streamPort{
		"rancher_servers_k3s": {from: cluster.LBListenPort, to: cluster.K3sListenPort},
	}
	if args.PortHTTP != 0 {
		streamBlockMap["ingress_http"] = streamPort{from: 80, to: args.PortHTTP}
	}
	if args.PortHTTPS != 0 {
		streamBlockMap["ingress_https"] = streamPort{from: 443, to: args.PortHTTPS}
	}
	streamBlock := g.Block{
		Directives: []g.IDirective{},
	}
	serversDis := func(port streamPort) []g.IDirective {
		ds := make([]g.IDirective, 0)
		for _, h := range hosts {
			ds = append(ds, &g.Directive{
				Name:       "server",
				Parameters: []string{fmt.Sprintf("%s:%d", h, port.to)},
			})
		}
		return ds
	}
	for name, port := range streamBlockMap {
		sds := serversDis(port)
		upstreamBlock := &g.Directive{
			Name: "upstream",
			Block: &g.Block{
				Directives: func() []g.IDirective {
					return append(sds, &g.Directive{
						Name: "least_conn",
					})
				}(),
			},
			Parameters: []string{name},
		}
		serverBlock := &g.Directive{
			Name: "server",
			Block: &g.Block{
				Directives: []g.IDirective{
					&g.Directive{
						Name:       "listen",
						Parameters: []string{fmt.Sprintf("%d", port.from)},
					},
					&g.Directive{
						Name:       "proxy_pass",
						Parameters: []string{name},
					},
				},
			},
		}
		streamBlock.Directives = append(streamBlock.Directives, upstreamBlock, serverBlock)
	}

	block := g.Block{
		Directives: []g.IDirective{
			&g.Directive{
				Name:       "user",
				Parameters: []string{"nginx"},
			},
			&g.Directive{
				Name:       "worker_processes",
				Parameters: []string{"auto"},
			},
			&g.Directive{
				Name:       "error_log",
				Parameters: []string{"/var/log/nginx/error.log"},
			},
			&g.Directive{
				Name:       "pid",
				Parameters: []string{"/run/nginx.pid"},
			},
			&g.Directive{
				Name: "events",
				Block: &g.Block{
					Directives: []g.IDirective{
						&g.Directive{
							Name:       "worker_connections",
							Parameters: []string{"1024"},
						},
					},
				},
			},
			&g.Directive{
				Name:  "stream",
				Block: &streamBlock,
			},
		},
	}
	cfg := g.Config{
		Block:    &block,
		FilePath: "-",
	}
	return g.DumpConfig(&cfg, &g.Style{Indent: 2})
}

func getNginxDefaultConfLoc() (string, error) {
	cmd := exec.Command("nginx", "-t")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.Wrap(err, "run `nginx -t`")
	}
	// get default configuration file place
	r := regexp.MustCompile("/.*/nginx.conf")
	matchString := r.FindStringSubmatch(string(output))
	if len(matchString) != 0 {
		return matchString[0], nil
	}
	return "", errors.New("default nginx conf not found")
}

func checkLBCondition() error {
	info("Checking system...")
	if runtime.GOOS != apis.GoosLinux {
		errf("Linux is required for Launching load balancer\n")
		return errors.New("not linux")
	}
	info("Checking user...")
	current, err := user.Current()
	if err != nil {
		return err
	}
	if current.Uid != "0" {
		info("root user is required for launching load balancer")
		return errors.New("not root")
	}
	return nil
}

// KillNginx kills nginx process
func KillNginx() error {
	kill := exec.Command("pkill", "-9", "nginx")
	output, err := kill.CombinedOutput()
	utils.InfoBytes(output)
	return err
}
