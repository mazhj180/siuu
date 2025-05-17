package server

import (
	"net/http"
	"siuu/server/config/constant"
	"siuu/server/handler/routing"
	"siuu/util"
)

func RegisterRouterHandle(prefix string) {
	Srv.Mux.HandleFunc(prefix+"/refresh", refreshRouter)
	Srv.Mux.HandleFunc(prefix+"/routes", getRelatedRoutes)
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

func getRelatedRoutes(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	if !query.Has("prx") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	router := routing.R()
	if router == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	related := router.RelatedRoutes(query.Get("prx"))
	if _, err := w.Write([]byte(related)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
