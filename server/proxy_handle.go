package server

import (
	"encoding/json"
	"io"
	"net/http"
	ps "siuu/server/config/proxies"
	"siuu/tunnel/proxy"
	"siuu/util/pinyin"
	"strings"
)

func RegisterProxyHandle(prefix string) {
	mux := srv.mux
	mux.HandleFunc(prefix, getProxies)
	mux.HandleFunc(prefix+"/add", addProxy)
	mux.HandleFunc(prefix+"/get", getProxy)
	mux.HandleFunc(prefix+"/set", setDefaultProxy)
	mux.HandleFunc(prefix+"/get-default", getDefaultProxy)
	mux.HandleFunc(prefix+"/delay", testDelay)

}

func addProxy(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var proxies []string
	if err = json.Unmarshal(body, &proxies); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	srv.Lock()
	defer srv.Unlock()
	for _, str := range proxies {
		var prx proxy.Proxy
		if prx, err = ps.ParseProxy(str); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		srv.Router.AddProxies(prx)
	}

	w.WriteHeader(http.StatusOK)
}

func getProxy(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	if !query.Has("prx") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	srv.RLock()
	defer srv.RUnlock()

	router := srv.Router
	prx := router.GetProxy(query.Get("prx"))
	if prx == nil {
		prx = &proxy.DirectProxy{}
	}
	data := prx.String()

	if _, err := w.Write([]byte(data)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func getProxies(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	srv.RLock()
	defer srv.RUnlock()

	router := srv.Router

	router.RLock()
	defer router.RUnlock()

	proxies := router.GetAllProxies()
	var prxStr []string

	for idx := range proxies {
		prxStr = append(prxStr, proxies[idx].String())
	}

	if query.Has("prx") {
		prx := query.Get("prx")
		names := make([]string, len(proxies))
		for i := range proxies {
			names[i] = proxies[i].Name()
		}

		names = pinyin.FuzzyMatch(names, prx)
		prxStr = make([]string, len(names))
		for i := range names {
			prxStr[i] = router.GetProxy(names[i]).String()
		}
	}

	var defau proxy.Proxy
	if defau = router.GetDefaultProxy(); defau == nil {
		defau = &proxy.DirectProxy{}
	}
	direct := &proxy.DirectProxy{}
	prxStr = append([]string{defau.String(), direct.String()}, prxStr...)

	data := "[" + strings.Join(prxStr, ",") + "]"

	if _, err := w.Write([]byte(data)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func setDefaultProxy(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	if !query.Has("proxy") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	srv.Lock()
	defer srv.Unlock()

	router := srv.Router
	router.Lock()
	defer router.Unlock()

	err := router.SetDefaultProxy(router.GetProxy(query.Get("proxy")))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func getDefaultProxy(w http.ResponseWriter, r *http.Request) {
	srv.RLock()
	defer srv.RUnlock()

	router := srv.Router
	router.RLock()
	defer router.RUnlock()

	prx := router.GetDefaultProxy()
	if _, err := w.Write([]byte(prx.Name())); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func testDelay(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	srv.RLock()
	defer srv.RUnlock()

	router := srv.Router
	router.RLock()
	defer router.RUnlock()

	proxies := router.GetAllProxies()

	if query.Has("prx") {
		proxies = proxies[1:]
		prx := query.Get("prx")
		names := make([]string, len(proxies))
		for i := range proxies {
			names[i] = proxies[i].Name()
		}

		names = pinyin.FuzzyMatch(names, prx)
		proxies = make([]proxy.Proxy, len(names))
		for i := range names {
			proxies[i] = router.GetProxy(names[i])
		}
	}

	proxies = append([]proxy.Proxy{router.GetDefaultProxy()}, proxies...)

	res := ps.TestProxyConnection(proxies)
	bytes, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err = w.Write(bytes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}
