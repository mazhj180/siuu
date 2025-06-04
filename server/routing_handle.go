package server

import (
	"net/http"
	"siuu/server/config/router"
	"siuu/tunnel/routing"
)

func RegisterRouterHandle(prefix string) {
	srv.mux.HandleFunc(prefix+"/refresh", refreshRouter)
}

func refreshRouter(w http.ResponseWriter, _ *http.Request) {

	if !srv.EnabledRule {
		return
	}

	var err error
	var rou routing.Router
	if rou, err = router.NewDefaultRouter(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	srv.Lock()
	defer srv.Unlock()

	srv.Router = rou
	w.WriteHeader(http.StatusOK)
}
