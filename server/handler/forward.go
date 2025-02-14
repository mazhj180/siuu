package handler

import (
	"siuu/tunnel"
)

func forwardHandle(ctx *context) {
	s := ctx.session
	tunnel.T.In(s)
	ctx.next()
}
