package main

import (
	"github.com/kardianos/service"
	"log"
	"os"
	"siuu/server"
	"siuu/server/config/constant"
	"siuu/util"
)

// run as a daemon service usage : ./siuu install &&./siuu start
// run as a normal program usage : ./siuu
func main() {

	p := util.GetConfig[string](constant.LogDirPath + "/siuu.log")

	// build service config
	conf := &service.Config{
		Name:        "siuu",
		DisplayName: "siuu",
		Description: "siuu is a user-level daemon that will automatically restart to maintain background operation but will not run automatically when the user logs in.",
		Option: map[string]interface{}{
			"UserService": true,
			"Label":       "siuu",
			"LogOutput":   p,
		},
	}

	s, err := service.New(&server.Server{}, conf)
	if err != nil {
		panic(err)
	}

	// you can start the program without adding any parameters, if you don't want to run as a daemon service .
	// parameters: install/uninstall/start/stop/restart
	if len(os.Args) > 1 {
		cmd := os.Args[1]
		err = service.Control(s, cmd)
		if err != nil {
			log.Fatalf("Failed to %s: %v\n", cmd, err)
		}
		return
	}

	// start service
	if err = s.Run(); err != nil {
		panic(err)
	}

}
