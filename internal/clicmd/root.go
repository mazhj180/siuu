package clicmd

import (
	"bufio"
	"bytes"
	"os"

	"github.com/spf13/cobra"
)

var (
	ronaldo     = []byte{}
	showVersion bool

	root = &cobra.Command{
		Use:   "siuu",
		Short: "siuu is a user-level daemon that will automatically restart to maintain background operation but will not run automatically when the user logs in.",
		Long: `siuu is a user-level daemon that will automatically restart to maintain background operation but will not run automatically when the user logs in.

By default siuu displays general information about the daemon status.
If the -v or --version flag is provided, siuu prints version information.

The route command displays and manages routing information including proxies, mappings, and rules.
The proxy command sets global system network proxy settings for HTTP and SOCKS protocols.`,
		Run: func(_ *cobra.Command, _ []string) {
			if showVersion {
				_, _ = os.Stdout.WriteString("siu version: 0.0.1\n")
				os.Exit(0)
			}

			buf := bytes.NewReader(ronaldo)
			scanner := bufio.NewScanner(buf)
			for scanner.Scan() {
				_, _ = os.Stdout.WriteString(scanner.Text() + "\n")
			}
		},
	}
)

func init() {
	root.PersistentFlags().BoolVarP(&showVersion, "version", "v", false, "show version")
}

func Execute() error {
	return root.Execute()
}
