package server

import (
	"net/http"
	"siuu/server/config/constant"
	"siuu/server/config/router"
	"siuu/util"
)

func RegisterRouterHandle(prefix string) {
	Srv.Mux.HandleFunc(prefix+"/refresh", refreshRouter)
}

func refreshRouter(w http.ResponseWriter, _ *http.Request) {
	routePath := util.GetConfigSlice(constant.RouteConfigPath)
	var routePathC []string
	for _, route := range routePath {
		routePathC = append(routePathC, util.ExpandHomePath(route))
	}
	xdbPath := util.GetConfig[string](constant.RouteXdbPath)
	xdbPath = util.ExpandHomePath(xdbPath)
	if err := router.RefreshBasicRouter(Srv.Router); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
