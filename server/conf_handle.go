package server

import (
	"net/http"
	_ "net/http/pprof"
)

func RegisterConfHandle(prefix string) {
	Srv.Mux.HandleFunc(prefix+"/open", openPprof)
	Srv.Mux.HandleFunc(prefix+"/close", closePprof)
}

func openPprof(_ http.ResponseWriter, _ *http.Request) {
	go Srv.startPprofServer()
}

func closePprof(_ http.ResponseWriter, _ *http.Request) {
	Srv.stopPprofServer()
}
