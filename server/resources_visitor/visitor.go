package visitor

import (
	"siuu/tunnel/routing"
	"sync"
)

type Visitor interface {
	sync.Locker

	RLock()
	RUnlock()
	Visit() AccessibleResources
}

type AccessibleResources struct {
	Router routing.Router
}
