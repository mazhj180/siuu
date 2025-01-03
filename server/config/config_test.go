package config

import (
	"fmt"
	"testing"
)

func TestConfig(t *testing.T) {
	port := Get[int64](ServerPort)
	fmt.Println(port)
}
