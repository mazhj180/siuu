package cmd

import (
	"bufio"
	"bytes"
	_ "embed"
	"github.com/spf13/cobra"
	"os"
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
)

func init() {
	RootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "show version information")
	RootCmd.AddCommand(startCmd, stopCmd, configCmd, proxyCmd)
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
