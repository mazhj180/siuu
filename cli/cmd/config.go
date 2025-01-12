package cmd

import (
	"github.com/spf13/cobra"
	"os"
	"siuu/util"
	"strings"
)

var (
	showConfig bool

	configCmd = &cobra.Command{
		Use:   "config",
		Short: "config of siuu",
		Run:   conf,
	}

	setConfigCmd = &cobra.Command{
		Use:   "set",
		Short: "set config for siuu",
		Run:   setConf,
	}

	getConfigCmd = &cobra.Command{
		Use:   "get",
		Short: "get config from siuu",
		Run:   getConf,
	}
)

func init() {
	configCmd.Flags().BoolVarP(&showConfig, "list", "l", false, "list all config")
	configCmd.AddCommand(setConfigCmd, getConfigCmd)
}

func conf(cmd *cobra.Command, args []string) {
	if showConfig {
		settings := util.GetSettings()

		for _, v := range settings {
			_, _ = os.Stdout.WriteString(v + "\n")
		}
		os.Exit(0)
	}
	_, _ = os.Stdout.WriteString("Usage: siu config [set | get | --list | -l]\n")
}

func setConf(cmd *cobra.Command, args []string) {

	arg := strings.TrimSpace(args[0])
	kv := strings.Split(arg, "=")
	if len(kv) != 2 {
		_, _ = os.Stdout.WriteString("invalid format\n")
		os.Exit(1)
	}
	if err := util.SetConfig(kv[0], kv[1]); err != nil {
		_, _ = os.Stdout.WriteString(err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

func getConf(cmd *cobra.Command, args []string) {
	arg := strings.TrimSpace(args[0])
	v := util.GetConfig[string](arg)
	_, _ = os.Stdout.WriteString(v + "\n")
	os.Exit(0)
}
