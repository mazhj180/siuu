package config

import (
	_ "embed"
)

var (
	//go:embed conf.toml
	Config []byte

	//go:embed proxies.toml
	Proxies []byte // TODO: remove this
)
