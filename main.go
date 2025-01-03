package main

import (
	"github.com/kardianos/service"
	"log"
	"os"
	"siuu/server"
)

func main() {

	conf := &service.Config{
		Name:        "siuu",
		DisplayName: "siuu CrossPlatform Daemon",
		Description: "siuu is a CrossPlatform Daemon",
		//Option: map[string]interface{}{
		//	//"UserService": true,
		//	"KeepAlive":   false,
		//	"LogOutput":   "/Users/mazhj/siuu.log",
		//},
	}

	s, err := service.New(&server.Server{}, conf)
	if err != nil {
		panic(err)
	}

	// install/uninstall/start/stop/restart
	if len(os.Args) > 1 {
		cmd := os.Args[1]
		err = service.Control(s, cmd)
		if err != nil {
			log.Fatalf("Failed to %s: %v\n", cmd, err)
		}
		return
	}

	if err = s.Run(); err != nil {
		log.Fatal(err)
	}

}
