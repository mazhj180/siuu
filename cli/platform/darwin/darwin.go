package darwin

import (
	"fmt"
	"os"
	"os/exec"
	"siuu/server/config/constant"
	"siuu/util"
	"strings"
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

	cmd := exec.Command("tail", "-n", fmt.Sprintf("%d", number), filePath)

	if follow {
		cmd = exec.Command("tail", "-f", "-n", fmt.Sprintf("%d", number), filePath)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 执行命令
	if err := cmd.Run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error running command: %v\n", err)
		os.Exit(1)
	}

}

func (c *Cli) ProxyOn() {

	prxHost := "127.0.0.1"
	passDomains := []string{"192.168.0.0/16", "10.0.0.0/8", "172.16.0.0/12", "127.0.0.1", "localhost", "*.local", "timestamp.apple.com"}

	httpPort := util.GetConfig[int64](constant.ProxyHttpPort)
	socksPort := util.GetConfig[int64](constant.ProxySocksPort)

	network := "Wi-Fi"
	c0 := exec.Command("networksetup", "-setproxybypassdomains", network, strings.Join(passDomains, ","))
	c1 := exec.Command("networksetup", "-setwebproxy", network, prxHost, fmt.Sprintf("%d", httpPort))
	c2 := exec.Command("networksetup", "-setsecurewebproxy", network, prxHost, fmt.Sprintf("%d", httpPort))
	c3 := exec.Command("networksetup", "-setsocksfirewallproxy", network, prxHost, fmt.Sprintf("%d", socksPort))

	c4 := exec.Command("networksetup", "-setwebproxystate", network, "on")
	c5 := exec.Command("networksetup", "-setsecurewebproxystate", network, "on")
	c6 := exec.Command("networksetup", "-setsocksfirewallproxystate", network, "on")

	if err := c0.Run(); err != nil {
		_, _ = os.Stdout.WriteString(err.Error())
		os.Exit(1)
	}

	if err := c1.Run(); err != nil {
		_, _ = os.Stdout.WriteString(err.Error())
		os.Exit(1)
	}
	if err := c2.Run(); err != nil {
		_, _ = os.Stdout.WriteString(err.Error())
		os.Exit(1)
	}
	if err := c3.Run(); err != nil {
		_, _ = os.Stdout.WriteString(err.Error())
		os.Exit(1)
	}

	if err := c4.Run(); err != nil {
		_, _ = os.Stdout.WriteString(err.Error())
		os.Exit(1)
	}
	if err := c5.Run(); err != nil {
		_, _ = os.Stdout.WriteString(err.Error())
		os.Exit(1)
	}
	if err := c6.Run(); err != nil {
		_, _ = os.Stdout.WriteString(err.Error())
		os.Exit(1)
	}
}

func (c *Cli) ProxyOff() {
	c1 := exec.Command("networksetup", "-setwebproxystate", "Wi-Fi", "off")
	c2 := exec.Command("networksetup", "-setsecurewebproxystate", "Wi-Fi", "off")
	c3 := exec.Command("networksetup", "-setsocksfirewallproxystate", "Wi-Fi", "off")

	if err := c1.Run(); err != nil {
		_, _ = os.Stdout.WriteString(err.Error())
		os.Exit(1)
	}
	if err := c2.Run(); err != nil {
		_, _ = os.Stdout.WriteString(err.Error())
		os.Exit(1)
	}
	if err := c3.Run(); err != nil {
		_, _ = os.Stdout.WriteString(err.Error())
		os.Exit(1)
	}
}
