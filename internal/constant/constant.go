package constant

import (
	_ "embed"
)

var (
	//go:embed conf.toml
	Config []byte

	//go:embed route_table.toml
	RouteTable []byte
)
