package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"os"
	"os/exec"
	"siuu/server/config"
	"strings"
)

var (
	show bool

	proxyCmd = &cobra.Command{
		Use:   "proxy [command | flags]",
		Short: "proxy for siuu",
		Run:   proxy,
	}

	onCmd = &cobra.Command{
		Use:   "on",
		Short: "proxy on",
		Run:   turnOn,
	}

	offCmd = &cobra.Command{
		Use:   "off",
		Short: "proxy off",
		Run:   turnOff,
	}

	setDefaultPrxCmd = &cobra.Command{
		Use:   "set",
		Short: "set a default proxy",
		Args:  cobra.ExactArgs(1),
		Run:   setDefaultPrx,
	}
)

func init() {
	proxyCmd.AddCommand(onCmd)
	proxyCmd.AddCommand(offCmd)
	proxyCmd.AddCommand(setDefaultPrxCmd)
	proxyCmd.Flags().BoolVarP(&show, "list", "l", false, "list all of proxies")
}

func proxy(cmd *cobra.Command, args []string) {
	if show {
		port := config.Get[int64](config.ServerPort)
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/prx/get", port))
		if err != nil {
			_, _ = os.Stdout.WriteString(err.Error())
			os.Exit(1)
		}
		if resp.StatusCode != 200 {
			_, _ = os.Stdout.WriteString("fail to get proxies\n")
			os.Exit(1)
		}
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			_, _ = os.Stdout.WriteString(err.Error())
			os.Exit(1)
		}
		var proxies []map[string]any
		if err = json.Unmarshal(data, &proxies); err != nil {
			_, _ = os.Stdout.WriteString(err.Error())
			os.Exit(1)
		}

		for _, prx := range proxies {
			_, _ = os.Stdout.WriteString(fmt.Sprintf("%s\n", prx["Name"]))
		}
	}
}

func turnOn(cmd *cobra.Command, args []string) {

	prxHost := "127.0.0.1"
	passDomains := []string{"192.168.0.0/16", "10.0.0.0/8", "172.16.0.0/12", "127.0.0.1", "localhost", "*.local", "timestamp.apple.com"}

	httpPort := config.Get[int64](config.ProxyHttpPort)
	socksPort := config.Get[int64](config.ProxySocksPort)

	switch config.Platform {
	case config.Darwin:
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

	case config.Windows:

	case config.Linux:
	}
	_, _ = os.Stdout.WriteString("proxy on\n")
}

func turnOff(cmd *cobra.Command, args []string) {
	switch config.Platform {
	case config.Darwin:
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
	case config.Windows:

	case config.Linux:
	}
	_, _ = os.Stdout.WriteString("proxy off\n")
}

func setDefaultPrx(cmd *cobra.Command, args []string) {
	name := strings.TrimSpace(args[0])
	port := config.Get[int64](config.ServerPort)
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/set?proxy=%s", port, name))
	if err != nil {
		_, _ = os.Stdout.WriteString(err.Error())
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		_, _ = os.Stdout.WriteString("failed to set default proxy\n")
		os.Exit(1)
	}
	_, _ = os.Stdout.WriteString(name + "\n")
}
