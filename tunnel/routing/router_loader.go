package routing

import "errors"

var NoRouterLoaderErr = errors.New("no router loader")

type Loader func(handler Router) error
