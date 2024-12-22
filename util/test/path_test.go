package test

import (
	"evil-gopher/util"
	"fmt"
	"path"
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
