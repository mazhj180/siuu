package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"siuu/server/config/constant"
	"siuu/util"
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
	if refresh {
		port := util.GetConfig[int64](constant.ServerPort)
		_, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/route/refresh", port))
		if err != nil {
			_, _ = os.Stdout.WriteString(err.Error())
			os.Exit(1)
		}
		_, _ = os.Stdout.WriteString("refresh\n")
	}
}
