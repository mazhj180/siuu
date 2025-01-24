package handle

import (
	"net/http"
	"siuu/server/config/constant"
	"siuu/server/handler/routing"
	"siuu/util"
)

func RegisterRouterHandle(mux *http.ServeMux, prefix string) {
	mux.HandleFunc(prefix+"/refresh", refreshRouter)
}

func refreshRouter(w http.ResponseWriter, r *http.Request) {
	routePath := util.GetConfigSlice(constant.RouteConfigPath)
	var routePathC []string
	for _, route := range routePath {
		routePathC = append(routePathC, util.ExpandHomePath(route))
	}
	xdbPath := util.GetConfig[string](constant.RouteXdbPath)
	xdbPath = util.ExpandHomePath(xdbPath)
	if err := routing.Refresh(routePath, xdbPath); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
