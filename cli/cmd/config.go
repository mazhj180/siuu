package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"siuu/server/config"
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
		m := config.GetAll()
		for k, v := range m {
			_, _ = os.Stdout.WriteString(fmt.Sprintf("%s=%s\n", k, v))
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
	if err := config.Set(kv[0], kv[1]); err != nil {
		_, _ = os.Stdout.WriteString(err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

func getConf(cmd *cobra.Command, args []string) {
	arg := strings.TrimSpace(args[0])
	v := config.Get[string](arg)
	_, _ = os.Stdout.WriteString(v + "\n")
	os.Exit(0)
}
