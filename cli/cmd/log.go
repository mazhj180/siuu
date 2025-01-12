package cmd

import (
	"github.com/spf13/cobra"
)

var (
	number int
	follow bool
	prx    bool
	sys    bool

	logCmd = &cobra.Command{
		Use:   "log",
		Short: "log for siuu",
		Run:   log,
	}
)

func init() {
	logCmd.Flags().IntVarP(&number, "number", "n", 10, "number of logs")
	logCmd.Flags().BoolVarP(&follow, "follow", "f", false, "realtime log")
	logCmd.Flags().BoolVarP(&prx, "proxy", "p", false, "proxy log")
	logCmd.Flags().BoolVarP(&sys, "system", "s", false, "system log")
}

func log(cmd *cobra.Command, args []string) {
	cli.Logg(follow, prx, sys, number)
}
