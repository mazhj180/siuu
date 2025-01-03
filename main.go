package main

import (
	"fmt"
	"github.com/kardianos/service"
	"log"
	"os"
	"siuu/server"
)

func main() {

	fmt.Println("starting siuu ...")

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
		log.Println("service.New error: ", err)
		panic(err)
	}

	// install/uninstall/start/stop/restart
	if len(os.Args) > 1 {

		cmd := os.Args[1]
		fmt.Println(cmd)
		err = service.Control(s, cmd)
		if err != nil {
			log.Fatalf("Failed to %s: %v\n", cmd, err)
		}
		return
	}

	fmt.Printf("Starting service...\n")
	// 阻塞，等待信号
	if err = s.Run(); err != nil {
		log.Fatal(err)
	}
	//_, _ = os.Stdout.WriteString("Usage: siuu [install | uninstall | start | stop | restart]\n")
	//os.Exit(0)

}
