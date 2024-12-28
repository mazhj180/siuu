package tunnel

import (
	"siu/tunnel/proto"
)

var (
	inCh   chan proto.Interface
	outCh  chan proto.Interface
	buffer []proto.Interface
	T      Tunnel
)

func init() {
	initInfiniteCh()
	initTunnel()
	go dispatch()
}

type Tunnel interface {
	In(proto.Interface)
	Out() (proto.Interface, bool)
}

type tunnel struct {
}

func initTunnel() {
	T = new(tunnel)
}

func (t *tunnel) In(v proto.Interface) {
	inCh <- v
}

func (t *tunnel) Out() (proto.Interface, bool) {
	v, ok := <-outCh
	return v, ok
}

func initInfiniteCh() {
	buffer = make([]proto.Interface, 0, 10)
	inCh = make(chan proto.Interface, 10)
	outCh = make(chan proto.Interface, 1)

	go func() {
		for {
			if len(buffer) > 0 {
				select {
				case outCh <- buffer[0]:
					buffer = buffer[1:]
				case v := <-inCh:
					buffer = append(buffer, v)
				}
			} else {
				v := <-inCh
				buffer = append(buffer, v)
			}
		}
	}()
}
