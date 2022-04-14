package pkg

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"runtime"
	"text/template"

	"github.com/pkg/errors"
)

func ConfigureNginx(args LoadBalancerArgs) error {
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

func UninstallNginx() error {
	file, err := Nginx.Open("static/nginx/remove_nginx.sh")
	if err != nil {
		return err
	}
	scriptName, err := SaveToTemp(file, "install_nginx-*.sh")
	if err != nil {
		return err
	}
	cmd := exec.Command("/bin/bash", scriptName)
	output, err := cmd.CombinedOutput()
	infoBytes(output)
	if err != nil {
		return err
	}
	return nil
}

func installNginx() error {
	file, err := Nginx.Open("static/nginx/install_nginx.sh")
	if err != nil {
		return err
	}
	scriptName, err := SaveToTemp(file, "install_nginx-*.sh")
	if err != nil {
		return err
	}
	cmd := exec.Command("/bin/bash", scriptName)
	output, err := cmd.CombinedOutput()
	infoBytes(output)
	return err
}

func setNginxConf(args LoadBalancerArgs) (string, error) {
	var conf string
	clause, err := getNginxStreamModClause()
	if err != nil {
		return "", err
	}
	conf += clause
	tmpl, err := template.ParseFS(Nginx, "static/nginx/nginx.conf.tmpl")
	if err != nil {
		return "", errors.Wrap(err, "parse tmpl")
	}
	buf := &bytes.Buffer{}
	err = tmpl.ExecuteTemplate(buf, "nginx.conf.tmpl", args)
	if err != nil {
		return "", errors.Wrap(err, "execute template")
	}
	all, err := ioutil.ReadAll(buf)
	if err != nil {
		return "", errors.Wrap(err, "read template result")
	}
	conf += string(all)
	loc, err := writeNginxConf(conf, args.Configuration)
	if err != nil {
		return "", errors.Wrap(err, "write nginx conf")
	}
	return loc, nil
}

func startNginx(conf string) error {
	info("Starting nginx")
	cmd := exec.Command("nginx", "-s", "quit")
	_ = cmd.Run()
	reloadCmd := exec.Command("nginx", "-c", conf)
	output, err := reloadCmd.CombinedOutput()
	infoBytes(output)
	return err
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
	if runtime.GOOS != "linux" {
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

func KillNginx() error {
	kill := exec.Command("pkill", "-9", "nginx")
	output, err := kill.CombinedOutput()
	infoBytes(output)
	return err
}
