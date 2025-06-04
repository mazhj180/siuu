package server

import (
	"net/http"
	_ "net/http/pprof"
)

func RegisterConfHandle(prefix string) {
	srv.mux.HandleFunc(prefix+"/open", openPprof)
	srv.mux.HandleFunc(prefix+"/close", closePprof)
}

func openPprof(_ http.ResponseWriter, _ *http.Request) {
	go srv.startPprofServer()
}

func closePprof(_ http.ResponseWriter, _ *http.Request) {
	srv.stopPprofServer()
}
