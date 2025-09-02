package handlers

import (
	"encoding/json"
	"net/http"
	"siuu/api"
	"siuu/internal/config"
	"siuu/pkg/logger"
)

func GetSystemHandlers(conf *config.SystemConfig, log *logger.Logger) map[string]http.HandlerFunc {
	return map[string]http.HandlerFunc{
		"/api/system/cfg": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(api.Response[api.SystemConfigInfo]{
				Code:    0,
				Message: "success",
				Data: api.SystemConfigInfo{
					LogPath:              conf.Log.Path,
					LogLevelSystem:       conf.Log.Level.System,
					LogLevelProxy:        conf.Log.Level.Proxy,
					ServerPort:           conf.Server.Port,
					ServerProxyHttpPort:  conf.Server.Proxy.Http.Port,
					ServerProxySocksPort: conf.Server.Proxy.Socks.Port,
					ServerProxyMode:      conf.Server.Proxy.Mode,
					ServerProxyTables:    conf.Server.Proxy.Tables,
					PprofEnable:          conf.Server.Pprof.Enable,
					PprofPort:            int(conf.Server.Pprof.Port),
				},
			}); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		},
	}
}
