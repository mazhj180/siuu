package test

import (
	"fmt"
	"os"
	"path"
	"siuu/util"
	"testing"
)

func TestExpand(t *testing.T) {
	p := "~/evil/gopher/"
	fmt.Println(path.Dir(p))
	p = util.ExpandHomePath(p)
	fmt.Println(path.Dir(p))
	f := path.Join(p, "sds.txt")
	fmt.Println(f)
}

func TestPath(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		return
	}

	fmt.Println(executable)

	filename := ".siu/conf/"
	fmt.Println(path.Dir(filename + "/../"))
	fmt.Println(util.AppRootPath)
}

func TestHome(t *testing.T) {
	fmt.Println(util.GetHomeDir())
}
