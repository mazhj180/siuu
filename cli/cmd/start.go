package cmd

import (
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path"
	"siuu/server/config/constant"
)

var (
	startCmd = &cobra.Command{
		Use:   "start",
		Short: "start the siu",
		Long:  "start the siu program as a daemon",
		Run:   start,
	}
)

func start(cmd *cobra.Command, args []string) {

	var c *exec.Cmd
	if constant.Platform == constant.Windows {
		c = exec.Command(path.Join(constant.RootPath, "siuu.exe"), "start")
	} else {
		c = exec.Command(path.Join(constant.RootPath, "siuu"), "start")
	}
	output, err := c.CombinedOutput()
	if err != nil {
		_, _ = os.Stdout.Write(output)
		os.Exit(1)
	}

}
