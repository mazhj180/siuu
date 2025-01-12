package handle

import (
	"net/http"
	"siuu/routing"
	"siuu/server/config/constant"
	"siuu/util"
)

func RegisterRouterHandle(mux *http.ServeMux, prefix string) {
	mux.HandleFunc(prefix+"/refresh", refreshRouter)
}

func refreshRouter(w http.ResponseWriter, r *http.Request) {
	routePath := util.GetConfig[string](constant.RouteConfigPath)
	routePath = util.ExpandHomePath(routePath)
	xdbPath := util.GetConfig[string](constant.RouteXdbPath)
	xdbPath = util.ExpandHomePath(xdbPath)
	if err := routing.Refresh(routePath, xdbPath); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
