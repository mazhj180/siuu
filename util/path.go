package util

import (
	"fmt"
	"os"
	"strings"
)

func ExpandHomePath(filepath string) string {
	if strings.HasPrefix(filepath, "~") {
		dir, err := os.UserHomeDir()
		if err != nil {
			panic(fmt.Errorf("cannot get user home dir: %s", err))
		}
		return strings.Replace(filepath, "~", dir, 1)
	}
	return filepath
}
