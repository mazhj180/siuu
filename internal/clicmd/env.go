package clicmd

import (
	"bytes"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var (
	needs = []string{
		"SERVER_HOST",
		"SERVER_PORT",
		"SERVER_PROXY_SOCKS_PORT",
		"SERVER_PROXY_HTTP_PORT",
	}

	isWrite bool

	envCmd = &cobra.Command{
		Use:   "env [--write/-w] [var...]",
		Short: "Display environment variables",
		Long: `Env displays the environment variables used by siuu.

The env command is used to display the environment variables used by siuu.

The --write/-w flag is used to write environment variables to the .env file.

If a var is provided, it will be written to the .env file.
If no var is provided, all environment variables will be written to the .env file.`,
		Run: func(cmd *cobra.Command, args []string) {

			if isWrite {
				if len(args) != 2 {
					fmt.Fprintf(os.Stdout, "Usage: siuu env --write/-w <var> <value>\n")
					return
				}

				if !slices.Contains(needs, args[0]) {
					return
				}

				buf := &bytes.Buffer{}
				for _, k := range needs {
					buf.WriteString(fmt.Sprintf("%s=%s\n", k, os.Getenv(k)))
				}
				buf.WriteString(fmt.Sprintf("%s=%s\n", args[0], args[1]))

				env, err := godotenv.Parse(buf)
				if err != nil {
					fmt.Fprintf(os.Stdout, "%s\n", err)
					return
				}

				err = godotenv.Write(env, "./.env")
				if err != nil {
					fmt.Fprintf(os.Stdout, "%s\n", err)
					return
				}

				return
			}

			if len(args) == 0 {
				args = needs
			}

			envs := os.Environ()
			for _, env := range envs {
				key := strings.SplitN(env, "=", 2)[0]
				if slices.Contains(args, key) {
					fmt.Fprintf(os.Stdout, "%s\n", env)
				}
			}
		},
	}
)

func init() {

	envCmd.Flags().BoolVarP(&isWrite, "write", "w", false, "write environment variables")

	root.AddCommand(envCmd)
}
