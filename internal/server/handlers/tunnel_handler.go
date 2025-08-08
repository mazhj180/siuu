package handlers

import (
	"net/http"
	"siuu/pkg/logger"
	"siuu/pkg/proxy/server"
)

func GetTunnelHandlers(log *logger.Logger, servers ...server.ProxyServer) map[string]http.HandlerFunc {
	return map[string]http.HandlerFunc{
		"/api/tunnel/clients": func(w http.ResponseWriter, r *http.Request) {

			for _, server := range servers {
				tunnels := server.ActiveTunnels()
				for _, t := range tunnels {
					t.GetStatus()
				}
			}
		},
	}
}
