package cmd

import (
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"os"
)

var (
	showRoute bool
	refresh   bool

	routeCmd = &cobra.Command{
		Use:   "route",
		Short: "route for siuu",
		Run:   route,
	}
)

func init() {
	routeCmd.Flags().BoolVarP(&showRoute, "show", "s", false, "show route")
	routeCmd.Flags().BoolVarP(&refresh, "refresh", "r", false, "refresh route")
}

func route(cmd *cobra.Command, args []string) {
	if showRoute {
		resp, err := http.Get("http://127.0.0.1:8080/route")
		if err != nil {
			_, _ = os.Stdout.WriteString(err.Error())
			os.Exit(1)
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			_, _ = os.Stdout.WriteString(err.Error())
			os.Exit(1)
		}

		_, _ = os.Stdout.WriteString(string(data))
	}
	if refresh {
		_, _ = os.Stdout.WriteString("refresh\n")
	}
}
