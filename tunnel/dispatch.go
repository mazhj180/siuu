package tunnel

import (
	"siuu/logger"
	"siuu/tunnel/monitor"
	"siuu/tunnel/proto"
	"siuu/tunnel/proxy"
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
		Conn:  monitor.Watch(conn),
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
		m := client.Conn.(monitor.Interface)
		up, upSpeed := m.UpTraffic()
		down, downSpeed := m.DownTraffic()

		var rt proto.TrafficRecorder
		if rt, ok = p.(proto.TrafficRecorder); ok {
			rt.RecordUp(up, upSpeed)
			rt.RecordDown(down, downSpeed)
		}

		logger.PInfo("<%s> send to [%s:%d] using by [%s]  [up:%d B | %.2f KB/s] [down:%d B | %.2f KB/s] ",
			sid,
			host,
			port,
			prx.GetName(),
			up, upSpeed/1024, down, downSpeed/1024)
	}
}
