package clicmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"siuu/api"

	"github.com/spf13/cobra"
)

var (
	isRules         bool
	isMappings      bool
	isProxies       bool
	isDefaultOutlet bool

	routeCmd = &cobra.Command{
		Use:   "route [--mappings/-m] [--proxies/-p] [--rules/-r] [--default/-d] [proxy_name]",
		Short: "Display and manage routing information including proxies, mappings, and rules",
		Long: `Route displays routing information and manages proxy configurations.

By default, route displays all routing information including default outlet,
proxies, mappings, and rules. You can use flags to display specific sections.

If a proxy_name is provided without other flags, it sets that proxy as the
default outlet.

The -m, --mappings flag displays proxy mappings showing the relationship
between keys and proxy names.

The -p, --proxies flag displays all configured proxies with their details
including name, type, server, port, and traffic type.

The -r, --rules flag displays routing rules showing type, key, and proxy
name associations.

The -d, --default flag displays the current default outlet proxy.`,
		Run: func(_ *cobra.Command, arg []string) {

			all := !isMappings && !isProxies && !isDefaultOutlet && !isRules

			host := os.Getenv("SERVER_HOST")
			port := os.Getenv("SERVER_PORT")
			url := fmt.Sprintf("http://%s:%s", host, port)
			resp, err := http.Get(url + "/api/router/clients")
			if err != nil {
				fmt.Fprintf(os.Stdout, "%s\n", err)
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			var res api.Response[api.RouterInfo]
			if err = json.Unmarshal(body, &res); err != nil {
				return
			}

			if res.Code != 0 {
				fmt.Fprintf(os.Stdout, "%s\n", res.Message)
				return
			}
			if isDefaultOutlet || all {
				if len(arg) > 0 && !isMappings && !isProxies {
					resp, err := http.Get(url + "/api/router/set/default_outlet?proxy_name=" + arg[0])
					if err != nil {
						fmt.Fprintf(os.Stdout, "%s\n", err)
						return
					}
					defer resp.Body.Close()

					body, err := io.ReadAll(resp.Body)
					var res api.Response[api.Data]
					if err = json.Unmarshal(body, &res); err != nil {
						return
					}

					if res.Code != 0 {
						fmt.Fprintf(os.Stdout, "%s\n", res.Message)
						return
					}

					fmt.Fprintf(os.Stdout, "Default Outlet : %s\n", arg[0])
				} else {
					fmt.Fprintf(os.Stdout, "Default Outlet : %s\n", res.Data.DefaultOutlet)
				}
			}

			if isProxies || all {
				fmt.Fprintf(os.Stdout, "Proxies <all of proxies that you configured> : \n")
				for _, proxy := range res.Data.Clients {
					fmt.Fprintf(os.Stdout, "\t%s - %s - %s - %d - %s\n", proxy.Name, proxy.Type, proxy.Server, proxy.Port, proxy.TrafficType)
				}
			}

			if isMappings || all {
				fmt.Fprintf(os.Stdout, "Proxy Mappings <key - proxy name> : \n")
				for mapping, proxyName := range res.Data.Mappings {
					fmt.Fprintf(os.Stdout, "\t%s - %s\n", mapping, proxyName)
				}
			}

			if isRules || all {
				fmt.Fprintf(os.Stdout, "Rules <type - key - proxy name> : \n")
				for _, rule := range res.Data.Rules {
					fmt.Fprintf(os.Stdout, "\t%s - %s - %s\n", rule.Typ, rule.Key, rule.Value)
				}
			}

		},
	}
)

func init() {
	routeCmd.Flags().BoolVarP(&isRules, "rules", "r", false, "show rules")
	routeCmd.Flags().BoolVarP(&isMappings, "mappings", "m", false, "show mappings")
	routeCmd.Flags().BoolVarP(&isProxies, "proxies", "p", false, "show proxies")
	routeCmd.Flags().BoolVarP(&isDefaultOutlet, "default", "d", false, "show default outlet")

	root.AddCommand(routeCmd)
}
