package clicmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"siuu/api"

	"github.com/spf13/cobra"
)

var (
	proxies []string
	testUrl string

	testCmd = &cobra.Command{
		Use:   "test [--proxies/-p] [--testurl/-t] [var ...]",
		Short: "Test the delay of the proxies",
		Long: `Test the delay of the proxies.

The test command is used to test the delay of the proxies.

The -p, --proxies flag is used to test the delay of the proxies.

The -t, --testurl flag is used to test the delay of the proxies.

The var ... are the proxies to test the delay of.`,
		Run: func(cmd *cobra.Command, args []string) {

			host := os.Getenv("SERVER_HOST")
			port := os.Getenv("SERVER_PORT")
			url := fmt.Sprintf("http://%s:%s", host, port)

			body, err := json.Marshal(api.PingProxyParam{
				Proxies: proxies,
				TestUrl: testUrl,
			})
			if err != nil {
				fmt.Fprintf(os.Stdout, "%s\n", err)
				return
			}

			resp, err := http.Post(url+"/api/router/proxy/ping", "application/json", bytes.NewReader(body))
			if err != nil {
				fmt.Fprintf(os.Stdout, "%s\n", err)
				return
			}
			defer resp.Body.Close()

			body, err = io.ReadAll(resp.Body)
			if err != nil {
				fmt.Fprintf(os.Stdout, "%s\n", err)
				return
			}
			var res api.Response[[]api.PingResult]
			if err = json.Unmarshal(body, &res); err != nil {
				fmt.Fprintf(os.Stdout, "%s\n", err)
				return
			}

			if res.Code != 0 {
				fmt.Fprintf(os.Stdout, "%s\n", res.Message)
				return
			}
			fmt.Fprintf(os.Stdout, "Proxies - Delay\n")
			for _, result := range res.Data {
				fmt.Fprintf(os.Stdout, "\t%s - %s\n", result.Name, result.Delay)
			}

		},
	}
)

func init() {
	testCmd.Flags().StringSliceVarP(&proxies, "proxies", "p", nil, "test the delay of the proxies")
	testCmd.Flags().StringVarP(&testUrl, "testurl", "t", "", "test the delay of the proxies")
	root.AddCommand(testCmd)
}
