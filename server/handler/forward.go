package handler

import (
	"siuu/tunnel"
)

func forwardHandle(ctx *context) {
	s := ctx.session

	traffic, err := tunnel.T.In(s)
	if err != nil {
		ctx.err = err
		return
	}

	ctx.up, ctx.upSpeed = traffic.Up, traffic.UpSpeed
	ctx.down, ctx.downSpeed = traffic.Down, traffic.DownSpeed
	ctx.delay = traffic.Delay

	ctx.next()
}
