package session

import (
	"fmt"
	"net"
	"siuu/tunnel/proto"
	"siuu/tunnel/proxy"
	"sync/atomic"
)

type Session interface {
	fmt.Stringer
	proto.Interface
	ID() string
	Handshakes() error
	SetProxy(proxy proxy.Proxy)
}

type Addr struct {
	IP     net.IP
	Port   uint16
	Domain string
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
