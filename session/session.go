package session

import (
	"evil-gopher/proxy"
	"evil-gopher/routing"
	"evil-gopher/tunnel"
	"fmt"
	"sync/atomic"
)

type Session interface {
	fmt.Stringer
	tunnel.Interface
	ID() string
	Handshakes() (*routing.TargetHost, error)
	SetProxy(proxy proxy.Proxy)
}

const maxSid = 0x400

var counter int32

func genSid() string {
	for {
		cur := atomic.LoadInt32(&counter)
		newVal := (cur + 1) % (maxSid + 1)
		if atomic.CompareAndSwapInt32(&counter, cur, newVal) {
			return fmt.Sprintf("sid-%#X", newVal)
		}
	}
}
