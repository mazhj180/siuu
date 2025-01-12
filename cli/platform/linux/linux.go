package linux

import (
	"fmt"
	"os"
	"os/exec"
	"siuu/server/config/constant"
	"siuu/util"
)

type Cli struct{}

func (c *Cli) Logg(follow, prx, _ bool, number int) {
	dir := util.GetConfig[string](constant.LogDirPath)
	var filePath string
	if prx {
		filePath = dir + "/proxy.log"
	} else {
		filePath = dir + "/system.log"
	}
	if follow {
		if err := exec.Command("tail", "-f -n ", fmt.Sprintf("%d", number), filePath).Run(); err != nil {
			_, _ = os.Stdout.WriteString(err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	}

	if err := exec.Command("tail", "-n ", fmt.Sprintf("%d", number), filePath).Run(); err != nil {
		_, _ = os.Stdout.WriteString(err.Error())
		os.Exit(1)
	}
}

func (c *Cli) ProxyOn() {

}

func (c *Cli) ProxyOff() {

}
