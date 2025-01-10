package handle

import (
	"net/http"
	"siuu/routing"
	"siuu/server/config"
	"siuu/util"
)

func RegisterRouterHandle(mux *http.ServeMux, prefix string) {
	mux.HandleFunc(prefix+"/refresh", refreshRouter)
}

func refreshRouter(w http.ResponseWriter, r *http.Request) {
	routePath := config.Get[string](config.RouteConfigPath)
	routePath = util.ExpandHomePath(routePath)
	xdbPath := config.Get[string](config.RouteXdbPath)
	xdbPath = util.ExpandHomePath(xdbPath)
	if err := routing.Refresh(routePath, xdbPath); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
