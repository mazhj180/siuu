package cmd

import (
	"bufio"
	"bytes"
	_ "embed"
	"github.com/spf13/cobra"
	"os"
	"siuu/cli/platform"
	"siuu/cli/platform/darwin"
	"siuu/cli/platform/linux"
	"siuu/cli/platform/win"
	"siuu/server/config/constant"
)

var (
	//go:embed ronaldo.txt
	ronaldo     []byte
	showVersion bool

	RootCmd = &cobra.Command{
		Use:   "siuu",
		Short: "cli for siuu",
		Run:   root,
	}

	cli platform.Client
)

func init() {

	switch constant.Platform {
	case constant.Darwin:
		cli = &darwin.Cli{}
	case constant.Linux:
		cli = &linux.Cli{}
	case constant.Windows:
		cli = &win.Cli{}
	default:
		panic("unknown platform")
	}

	RootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "show version information")
	RootCmd.AddCommand(startCmd, stopCmd, configCmd, proxyCmd, logCmd)
}

func root(cmd *cobra.Command, args []string) {

	if showVersion {
		_, _ = os.Stdout.WriteString("siu version: 0.0.1\n")
		os.Exit(0)
	}

	buf := bytes.NewReader(ronaldo)
	scanner := bufio.NewScanner(buf)
	for scanner.Scan() {
		_, _ = os.Stdout.WriteString(scanner.Text() + "\n")
	}
}
