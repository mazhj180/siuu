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
	mux := Srv.Mux
	mux.HandleFunc(prefix, getProxies)
	mux.HandleFunc(prefix+"/add", addProxy)
	mux.HandleFunc(prefix+"/remove", removeProxy)
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

	if err = ps.AddProxies(proxies...); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func removeProxy(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	if !query.Has("proxies") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	params := query.Get("proxies")
	var names []string
	for _, name := range strings.Split(params, ",") {
		names = append(names, name)
	}
	ps.RemoveProxies(names...)
	w.WriteHeader(http.StatusOK)
}

func getProxy(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	if !query.Has("prx") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	data, err := json.Marshal(ps.GetProxy(query.Get("prx")))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err = w.Write(data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func getProxies(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	proxies := ps.GetProxies()

	if query.Has("prx") {
		proxies = proxies[1:]
		prx := query.Get("prx")
		names := make([]string, len(proxies))
		for i := range proxies {
			names[i] = proxies[i].GetName()
		}

		names = pinyin.FuzzyMatch(names, prx)
		proxies = make([]proxy.Proxy, len(names))
		for i := range names {
			proxies[i] = ps.GetProxy(names[i])
		}
		proxies = append([]proxy.Proxy{ps.GetSelectedProxy()}, proxies...)
	}

	data, err := json.Marshal(proxies)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err = w.Write(data); err != nil {
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
	err := ps.SetSelectedProxy(query.Get("proxy"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func getDefaultProxy(w http.ResponseWriter, r *http.Request) {
	prx := ps.GetSelectedProxy()
	if _, err := w.Write([]byte(prx.GetName())); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func testDelay(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	proxies := ps.GetProxies()

	if query.Has("prx") {
		proxies = proxies[1:]
		prx := query.Get("prx")
		names := make([]string, len(proxies))
		for i := range proxies {
			names[i] = proxies[i].GetName()
		}

		names = pinyin.FuzzyMatch(names, prx)
		proxies = make([]proxy.Proxy, len(names))
		for i := range names {
			proxies[i] = ps.GetProxy(names[i])
		}
		proxies = append([]proxy.Proxy{ps.GetSelectedProxy()}, proxies...)
	}

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
