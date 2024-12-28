package tunnel

import (
	"siu/logger"
	"siu/tunnel/proto"
	"siu/tunnel/proxy"
)

func dispatch() {
	for {
		packet, ok := T.Out()
		if !ok {
			break
		}
		go do(packet)
	}
}

func do(p proto.Interface) {
	sid := p.ID()
	host, port, conn := p.GetHost(), p.GetPort(), p.GetConn()

	var isTLS bool
	h, ok := p.(proto.HttpInterface)
	if ok {
		isTLS = h.IsTLS()
	}
	client := &proxy.Client{
		Sid:   sid,
		Conn:  conn,
		Host:  host,
		Port:  port,
		IsTLS: isTLS,
	}
	prx := p.GetProxy()
	err := prx.Act(client)
	if err != nil {
		logger.SError("<%s> [%s] [%s] to [%s:%d] using by [%s]  err: %s",
			sid,
			p.GetProtocol(),
			conn.RemoteAddr().String(),
			host,
			port,
			prx.GetName(),
			err)

		logger.PError("<%s> the req of [%s:%d] was wrong using by [%s] ", sid, host, port, prx.GetName())

	} else {
		logger.PInfo("<%s> send to [%s:%d] using by [%s] ", sid, host, port, prx.GetName())
	}
}
