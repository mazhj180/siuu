package main

import (
	"fmt"
	"log"
	"os"
	"siuu/internal/config"
	"siuu/internal/server"
	"siuu/pkg/util/path"
	"siuu/pkg/util/toml"
	"strings"

	"github.com/kardianos/service"
)

var (
	baseDir = path.GetHomeDir() + "/.siuu"

	_ = os.MkdirAll(baseDir+"/conf", 0755)
)

// run as a daemon service usage : ./siuu install &&./siuu start
// run as a normal program usage : ./siuu
// you'd better run it as a daemon service, normal mode for debugging.
func main() {

	// build service config
	conf := &service.Config{
		Name:        "siuu",
		DisplayName: "siuu",
		Description: "siuu is a user-level daemon that will automatically restart to maintain background operation but will not run automatically when the user logs in.",
		Option: map[string]interface{}{
			"UserService":  true,
			"Label":        "siuu",
			"LogDirectory": baseDir + "/log",
			"RunAtLoad":    false, // disable auto start
		},
	}

	// load system config from file
	sysConf := &config.SystemConfig{}
	err := toml.LoadTomlFromFile(baseDir+"/conf/conf.toml", sysConf)
	if err != nil {
		panic(err)
	}

	siuu, err := server.New(sysConf)
	if err != nil {
		panic(err)
	}

	s, err := service.New(siuu, conf)
	if err != nil {
		panic(err)
	}

	// you can start the program without adding any parameters, if you don't want to run as a daemon service .
	// parameters: install/uninstall/start/stop/restart
	if len(os.Args) > 1 {
		cmd := os.Args[1]
		if strings.EqualFold(cmd, "install") {
			_, _ = fmt.Fprintf(os.Stdout, "Warn: after installation, do not arbitrarily move the program location.\n")
			_, _ = fmt.Fprintf(os.Stdout, "are you sure you want to install? (y/n)\n")

			var f string
			_, _ = fmt.Scanln(&f)
			if !strings.EqualFold(f, "y") {
				return
			}

			// siuu.InstallConfig()
		}
		if strings.EqualFold(cmd, "uninstall") {
			_, _ = fmt.Fprintf(os.Stdout, "Warn: after uninstallation, all of your configurations will be cleared.\n")
			_, _ = fmt.Fprintf(os.Stdout, "are you sure you want to uninstall? (y/n)\n")

			var f string
			_, _ = fmt.Scanln(&f)
			if !strings.EqualFold(f, "y") {
				return
			}

			// siuu.UninstallConfig()
		}
		err = service.Control(s, cmd)
		if err != nil {
			log.Fatalf("Failed to %s: %v\n", cmd, err)
		}
		return
	}

	if service.Interactive() {
		_, _ = fmt.Fprintf(os.Stdout, "Suggestion: you'd better run this program in the way of daemon service.\n")
	}

	// start service
	if err = s.Run(); err != nil {
		panic(err)
	}

}
