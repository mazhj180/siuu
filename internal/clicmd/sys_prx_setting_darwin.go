//go:build darwin

package clicmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"siuu/api"
	"strings"

	"github.com/spf13/cobra"
)

var (
	isSocks bool
	isHttp  bool

	prxCmd = &cobra.Command{
		Use:   "proxy [--socks/-s] [--http] [on/off]",
		Short: "Set global system network proxy settings",
		Long: `Proxy sets global system network proxy settings for HTTP and SOCKS protocols.

By default, when no specific protocol flags are provided, both HTTP and SOCKS
proxies are configured simultaneously.

The command requires either 'on' or 'off' as a parameter to enable or disable
the proxy settings.

The -s, --socks flag configures only the SOCKS proxy settings. When enabled,
it sets up a SOCKS firewall proxy for the system network interface.

The -h, --http flag configures only the HTTP proxy settings. When enabled,
it sets up both HTTP and HTTPS web proxies for the system network interface.

When proxy is enabled, the system automatically configures bypass domains
for local networks (192.168.0.0/16, 10.0.0.0/8, 172.16.0.0/12, 127.0.0.1,
localhost, *.local, timestamp.apple.com) to ensure local traffic is not
routed through the proxy.`,
		Args: func(cmd *cobra.Command, args []string) error {
			// Skip validation if help is requested or if "help" is passed as argument
			if cmd.Flags().Lookup("help").Changed || (len(args) > 0 && args[0] == "help") {
				return nil
			}
			if len(args) == 0 {
				return fmt.Errorf("missing parameter: must be 'on' or 'off'")
			}
			if args[0] != "on" && args[0] != "off" {
				return fmt.Errorf("invalid parameter: %s (must be 'on' or 'off')", args[0])
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {

			if len(args) > 0 && args[0] == "help" {
				_ = cmd.Help()
				return
			}

			both := !isSocks && !isHttp

			url := fmt.Sprintf("http://%s:%s", os.Getenv("SERVER_HOST"), os.Getenv("SERVER_PORT"))
			resp, err := http.Get(url + "/api/system/cfg")
			if err != nil {
				fmt.Fprintf(os.Stdout, "%s\n", err)
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			var res api.Response[api.SystemConfigInfo]
			if err = json.Unmarshal(body, &res); err != nil {
				fmt.Fprintf(os.Stdout, "%s\n", err)
				return
			}

			if res.Code != 0 {
				fmt.Fprintf(os.Stdout, "%s\n", res.Message)
				return
			}

			prxHost := os.Getenv("SERVER_HOST")
			passDomains := []string{"192.168.0.0/16", "10.0.0.0/8", "172.16.0.0/12", "127.0.0.1", "localhost", "*.local", "timestamp.apple.com"}

			network := "Wi-Fi"
			if err = exec.Command("networksetup", "-setproxybypassdomains", network, strings.Join(passDomains, ",")).Run(); err != nil {
				fmt.Fprintf(os.Stdout, "%s\n", err)
				return
			}

			if isSocks || both {
				port := os.Getenv("SERVER_PROXY_SOCKS_PORT")
				if args[0] == "on" {
					if err = exec.Command("networksetup", "-setsocksfirewallproxy", network, prxHost, fmt.Sprintf("%s", port)).Run(); err != nil {
						fmt.Fprintf(os.Stdout, "%s\n", err)
						return
					}

					if err = exec.Command("networksetup", "-setsocksfirewallproxystate", network, "on").Run(); err != nil {
						fmt.Fprintf(os.Stdout, "%s\n", err)
						return
					}
				} else {
					if err = exec.Command("networksetup", "-setsocksfirewallproxy", network, "", "").Run(); err != nil {
						fmt.Fprintf(os.Stdout, "%s\n", err)
						return
					}

					if err = exec.Command("networksetup", "-setsocksfirewallproxystate", network, "off").Run(); err != nil {
						fmt.Fprintf(os.Stdout, "%s\n", err)
						return
					}
				}

			}
			if isHttp || both {
				port := os.Getenv("SERVER_PROXY_HTTP_PORT")
				if args[0] == "on" {
					if err = exec.Command("networksetup", "-setwebproxy", network, prxHost, fmt.Sprintf("%s", port)).Run(); err != nil {
						fmt.Fprintf(os.Stdout, "%s\n", err)
						return
					}

					if err = exec.Command("networksetup", "-setsecurewebproxy", network, prxHost, fmt.Sprintf("%s", port)).Run(); err != nil {
						fmt.Fprintf(os.Stdout, "%s\n", err)
						return
					}

					if err = exec.Command("networksetup", "-setwebproxystate", network, "on").Run(); err != nil {
						fmt.Fprintf(os.Stdout, "%s\n", err)
						return
					}

					if err = exec.Command("networksetup", "-setsecurewebproxystate", network, "on").Run(); err != nil {
						fmt.Fprintf(os.Stdout, "%s\n", err)
						return
					}

				} else {
					if err = exec.Command("networksetup", "-setwebproxy", network, "", "").Run(); err != nil {
						fmt.Fprintf(os.Stdout, "%s\n", err)
						return
					}

					if err = exec.Command("networksetup", "-setsecurewebproxy", network, "", "").Run(); err != nil {
						fmt.Fprintf(os.Stdout, "%s\n", err)
						return
					}

					if err = exec.Command("networksetup", "-setwebproxystate", network, "off").Run(); err != nil {
						fmt.Fprintf(os.Stdout, "%s\n", err)
						return
					}

					if err = exec.Command("networksetup", "-setsecurewebproxystate", network, "off").Run(); err != nil {
						fmt.Fprintf(os.Stdout, "%s\n", err)
						return
					}
				}
			}
		},
	}
)

func init() {
	prxCmd.Flags().BoolVarP(&isSocks, "socks", "s", false, "set socks proxy")
	prxCmd.Flags().BoolVarP(&isHttp, "http", "", false, "set http proxy")

	root.AddCommand(prxCmd)
}
