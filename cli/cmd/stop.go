package cmd

import (
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path"
	"siuu/server/config"
)

var stopCmd = &cobra.Command{
	Use: "stop",
	Run: stop,
}

func stop(cmd *cobra.Command, args []string) {

	var c *exec.Cmd
	if config.Platform == config.Windows {
		c = exec.Command(path.Join(config.RootPath, "siuu.exe"), "stop")
	} else {
		c = exec.Command(path.Join(config.RootPath, "siuu"), "stop")
	}
	output, err := c.CombinedOutput()
	if err != nil {
		_, _ = os.Stdout.Write(output)
		os.Exit(1)
	}
}
