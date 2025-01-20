package main

import (
	"fmt"
	"github.com/kardianos/service"
	"log"
	"os"
	"siuu/server"
	"siuu/util"
	"strings"
)

// run as a daemon service usage : ./siuu install &&./siuu start
// run as a normal program usage : ./siuu
func main() {

	// build service config
	conf := &service.Config{
		Name:        "siuu",
		DisplayName: "siuu",
		Description: "siuu is a user-level daemon that will automatically restart to maintain background operation but will not run automatically when the user logs in.",
		Option: map[string]interface{}{
			"UserService": true,
			"Label":       "siuu",
			"LogOutput":   util.GetHomeDir() + "./siuu/siuu.log",
		},
	}

	siuu := &server.Server{}
	s, err := service.New(siuu, conf)
	if err != nil {
		panic(err)
	}

	// you can start the program without adding any parameters, if you don't want to run as a daemon service .
	// parameters: install/uninstall/start/stop/restart
	if len(os.Args) > 1 {
		cmd := os.Args[1]
		if strings.EqualFold(cmd, "install") {
			log.Println("InstallConfig() method called.")
			siuu.InstallConfig()
		}
		if strings.EqualFold(cmd, "uninstall") {
			siuu.UninstallConfig()
		}
		err = service.Control(s, cmd)
		if err != nil {
			log.Fatalf("Failed to %s: %v\n", cmd, err)
		}
		return
	}

	if !service.Interactive() {
		_, _ = fmt.Fprintf(os.Stdout, "Suggestion: you'd better run this program in the way of daemon service.\n")
	}

	// start service
	if err = s.Run(); err != nil {
		panic(err)
	}

}
