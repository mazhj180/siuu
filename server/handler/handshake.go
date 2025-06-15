package handler

import (
	"fmt"
	"net"
	"siuu/tunnel/logger"
)

func handshakeHandle(ctx *context) {
	s := ctx.session
	err := s.Handshakes()
	if err != nil {
		logger.PError("<%s> [handshake] [cli: %s] handshakes was failed; err: %s", s.ID(), ctx.remoteAddr, err)
		return
	}
	ctx.handshake = true

	host := s.GetHost()
	port := s.GetPort()
	addr := fmt.Sprintf("%s:%d", host, port)
	ctx.dst = addr
	logger.PDebug("<%s> [handshake] [cli: %s] handshakes successfully [dst: %s] ", s.ID(), ctx.remoteAddr, addr)

	if ip := net.ParseIP(host); ip != nil && (ip.IsLoopback() || ip.IsPrivate()) {
		ctx.skip()
		logger.PDebug("<%s> [handshake] [cli: %s] skip route cause: dst-ip is loopback or private", s.ID(), ctx.remoteAddr)
	} else {
		ctx.next()
	}
}
