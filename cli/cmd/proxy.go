package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"os"
	"siuu/server/config/constant"
	"siuu/util"
	"strings"
)

var (
	showPrx bool

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
	proxyCmd.Flags().BoolVarP(&showPrx, "list", "l", false, "list all of proxies")
}

func proxy(cmd *cobra.Command, args []string) {
	if showPrx {
		port := util.GetConfig[int64](constant.ServerPort)
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
	cli.ProxyOn()
	_, _ = os.Stdout.WriteString("proxy on\n")
}

func turnOff(cmd *cobra.Command, args []string) {
	cli.ProxyOff()
	_, _ = os.Stdout.WriteString("proxy off\n")
}

func setDefaultPrx(cmd *cobra.Command, args []string) {
	name := strings.TrimSpace(args[0])
	port := util.GetConfig[int64](constant.ServerPort)
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
