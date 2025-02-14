package handler

import (
	"fmt"
	"net"
	"siuu/logger"
)

func handshakeHandle(ctx *context) {
	s := ctx.session
	err := s.Handshakes()
	if err != nil {
		logger.SError("<%s> ack proxy resp was failed; err: %s", s.ID(), err)
		return
	}
	logger.SDebug("<%s> client: handshakes successfully with siuu server", s.ID())
	host := s.GetHost()
	port := s.GetPort()
	addr := fmt.Sprintf("%s:%d", host, port)
	logger.SDebug("<%s> dst addr was [%s]", s.ID(), addr)

	if ip := net.ParseIP(host); ip != nil && (ip.IsLoopback() || ip.IsPrivate()) {
		ctx.skip()
	} else {
		ctx.next()
	}
}
