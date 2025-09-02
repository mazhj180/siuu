package service

import (
	"context"
	"siuu/pkg/logger"
	"siuu/pkg/proxy/route"
	"siuu/pkg/proxy/server"
	"siuu/pkg/tunnel"
	"time"
)

// system proxy server callbacks
func GetCallbacks(router route.Router, log *logger.Logger) *server.Callback {

	return &server.Callback{
		OnError: func(ctx *server.Context, err error) {
			if err != nil {
				if ctx != nil {
					log.Error("<%s> [%s] %s", ctx.SessionId(), ctx.Stage, err)
				} else {
					log.Error("error: %s", err)
				}
			}
		},

		OnAcceptd: func(ctx *server.Context) {
			conn := ctx.Conn()
			addr := conn.RemoteAddr()
			log.Debug("<%s> [access] [%s] is arrival", ctx.SessionId(), addr)
		},

		OnConnected: func(ctx *server.Context) tunnel.Tunnel {

			dstHost := ctx.DstHost
			dstPort := ctx.DstPort

			proxy := router.Route(dstHost)
			if proxy == nil {
				ctx.SelectedRoute = "direct"
				log.Debug("<%s> [routing] [%s] proxy not found and use direct ", ctx.SessionId(), dstHost)
				return nil
			}

			ctx.SelectedRoute = proxy.Name()
			log.Debug("<%s> [routing] [%s] proxy is [%s]", ctx.SessionId(), dstHost, proxy.Name())

			prxCtx, cancel := context.WithTimeout(ctx.Context, 30*time.Second)
			defer cancel()

			agency, err := proxy.Connect(prxCtx, "tcp", dstHost, dstPort)
			if err != nil {
				log.Error("<%s> [proxy] [%s] connect failed", ctx.SessionId(), dstHost)
				return nil
			}

			t, err := tunnel.NewSystemProxyTunnel(nil, ctx.Conn(), agency, ctx.SessionId())
			if err != nil {
				log.Error("<%s> [proxy] [%s] create tunnel failed", ctx.SessionId(), dstHost)
				return nil
			}

			return t
		},

		OnFinished: func(ctx *server.Context) {
			stat := ctx.TunnelStatus
			if stat == nil {
				return
			}

			log.Debug("<%s> [finished] [%s] used by [%s] [up:%d B | %s] [down:%d B | %s] [duration: %d ms]",
				ctx.SessionId(),
				ctx.DstHost,
				ctx.SelectedRoute,
				stat.UpBytes,
				tunnel.FormatSpeed(stat.UpSpeed),
				stat.DownBytes,
				tunnel.FormatSpeed(stat.DownSpeed),
				stat.TotalDuration.Milliseconds(),
			)
		},
	}
}
